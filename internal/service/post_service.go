package service

import (
	"errors"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/repository"
	"math"
	"time"
)

type PostService interface {
	CreatePost(authorID uint, req dto.CreatePostRequest) (*models.Post, error)
	GetPostByID(id uint) (*models.Post, error)
	UpdatePost(id uint, authorID uint, req dto.UpdatePostRequest) (*models.Post, error)
	DeletePost(id uint, authorID uint) error
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

func (s *postService) CreatePost(authorID uint, req dto.CreatePostRequest) (*models.Post, error) {
	post := &models.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID,
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	// 重新載入以獲取關聯的 Author
	return s.postRepo.FindByID(post.ID)
}

func (s *postService) GetPostByID(id uint) (*models.Post, error) {
	return s.postRepo.FindByID(id)
}

func (s *postService) UpdatePost(id uint, authorID uint, req dto.UpdatePostRequest) (*models.Post, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 檢查是否為作者本人
	if post.AuthorID != authorID {
		return nil, errors.New("無權限修改此文章")
	}

	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}

	post.UpdatedAt = time.Now()

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	return s.postRepo.FindByID(id)
}

func (s *postService) DeletePost(id uint, authorID uint) error {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return err
	}

	// 檢查是否為作者本人
	if post.AuthorID != authorID {
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

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dto.PaginationResponse{
		Data:       posts,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

