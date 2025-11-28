package service

import (
	"errors"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/repository"
	"math"
)

type PostService interface {
	CreatePost(authorID uint, req dto.CreatePostRequest) (*dto.PostResponse, error)
	GetPostByID(id uint) (*dto.PostResponse, error)
	UpdatePost(id uint, authorID uint, role string, req dto.UpdatePostRequest) (*dto.PostResponse, error)
	DeletePost(id uint, authorID uint, role string) error
	GetPostsWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error)
}

type postService struct {
	postRepo repository.PostRepository
}

func NewPostService(postRepo repository.PostRepository) PostService {
	return &postService{
		postRepo: postRepo,
	}
}

// toPostResponse 將 models.Post 轉換為 dto.PostResponse
func toPostResponse(post *models.Post) *dto.PostResponse {
	response := &dto.PostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		AuthorID:  post.AuthorID,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	// 如果有 Author 關聯，轉換為 UserResponse
	if post.Author.ID != 0 {
		response.Author = &dto.UserResponse{
			ID:        post.Author.ID,
			Email:     post.Author.Email,
			Role:      post.Author.Role,
			CreatedAt: post.Author.CreatedAt,
			UpdatedAt: post.Author.UpdatedAt,
		}
	}

	return response
}

func (s *postService) CreatePost(authorID uint, req dto.CreatePostRequest) (*dto.PostResponse, error) {
	post := &models.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID,
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	// 重新載入以獲取關聯的 Author
	postWithAuthor, err := s.postRepo.FindByID(post.ID)
	if err != nil {
		return nil, err
	}

	return toPostResponse(postWithAuthor), nil
}

func (s *postService) GetPostByID(id uint) (*dto.PostResponse, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return toPostResponse(post), nil
}

func (s *postService) UpdatePost(id uint, authorID uint, role string, req dto.UpdatePostRequest) (*dto.PostResponse, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 檢查是否為作者本人或管理員
	if role != "admin" && post.AuthorID != authorID {
		return nil, errors.New("無權限修改此文章")
	}

	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	// 重新載入以獲取關聯的 Author
	updatedPost, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return toPostResponse(updatedPost), nil
}

func (s *postService) DeletePost(id uint, authorID uint, role string) error {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 檢查是否為作者本人或管理員
	if role != "admin" && post.AuthorID != authorID {
		return errors.New("無權限刪除此文章")
	}

	return s.postRepo.Delete(id)
}

func (s *postService) GetPostsWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error) {
	// 設定預設值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	posts, total, err := s.postRepo.FindAllWithPagination(req.Page, req.PageSize, req.Search)
	if err != nil {
		return nil, err
	}

	// 轉換為 PostResponse
	postResponses := make([]dto.PostResponse, len(posts))
	for i, post := range posts {
		postResponses[i] = *toPostResponse(&post)
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dto.PaginationResponse{
		Data:       postResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

