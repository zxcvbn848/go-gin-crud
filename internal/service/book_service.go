package service

import (
	"context"
	"go-gin-crud/internal/cache"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/logger"
	"go-gin-crud/internal/repository"
	"math"
)

type BookService interface {
	CreateBook(req dto.CreateBookRequest) (*dto.BookResponse, error)
	GetBookByID(id uint) (*dto.BookResponse, error)
	UpdateBook(id uint, req dto.UpdateBookRequest) (*dto.BookResponse, error)
	DeleteBook(id uint) error
	GetBooksWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error)
}

type bookService struct {
	bookRepo  repository.BookRepository
	bookCache cache.BookCache
}

func NewBookService(bookRepo repository.BookRepository, bookCache cache.BookCache) BookService {
	return &bookService{
		bookRepo:  bookRepo,
		bookCache: bookCache,
	}
}

func (s *bookService) CreateBook(req dto.CreateBookRequest) (*dto.BookResponse, error) {
	book := &models.Book{
		Title:  req.Title,
		Author: req.Author,
	}

	if err := s.bookRepo.Create(book); err != nil {
		return nil, err
	}

	resp := &dto.BookResponse{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		CreatedAt: book.CreatedAt,
		UpdatedAt: book.UpdatedAt,
	}
	if s.bookCache != nil {
		_ = s.bookCache.SetBook(context.Background(), book.ID, resp)
		logger.Log.WithField("book_id", book.ID).Info("Book 快取已寫入（Create）")
	}
	return resp, nil
}

func (s *bookService) GetBookByID(id uint) (*dto.BookResponse, error) {
	ctx := context.Background()
	if s.bookCache != nil {
		if cached, err := s.bookCache.GetBook(ctx, id); err == nil && cached != nil {
			logger.Log.WithField("book_id", id).Info("Book 從 Redis 快取取得")
			return cached, nil
		}
	}

	book, err := s.bookRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	resp := &dto.BookResponse{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		CreatedAt: book.CreatedAt,
		UpdatedAt: book.UpdatedAt,
	}
	if s.bookCache != nil {
		_ = s.bookCache.SetBook(ctx, id, resp)
		logger.Log.WithField("book_id", id).Info("Book 從 DB 取得並寫入快取")
	} else {
		logger.Log.WithField("book_id", id).Info("Book 從 DB 取得（未啟用快取）")
	}
	return resp, nil
}

func (s *bookService) UpdateBook(id uint, req dto.UpdateBookRequest) (*dto.BookResponse, error) {
	book, err := s.bookRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		book.Title = req.Title
	}
	if req.Author != "" {
		book.Author = req.Author
	}

	if err := s.bookRepo.Update(book); err != nil {
		return nil, err
	}

	resp := &dto.BookResponse{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		CreatedAt: book.CreatedAt,
		UpdatedAt: book.UpdatedAt,
	}
	if s.bookCache != nil {
		_ = s.bookCache.SetBook(context.Background(), book.ID, resp)
		logger.Log.WithField("book_id", book.ID).Info("Book 快取已更新（Update）")
	}
	return resp, nil
}

func (s *bookService) DeleteBook(id uint) error {
	if err := s.bookRepo.Delete(id); err != nil {
		return err
	}
	if s.bookCache != nil {
		_ = s.bookCache.DeleteBook(context.Background(), id)
		logger.Log.WithField("book_id", id).Info("Book 快取已刪除（Delete）")
	}
	return nil
}

func (s *bookService) GetBooksWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error) {
	// 設定預設值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	books, total, err := s.bookRepo.FindAllWithPagination(req.Page, req.PageSize, req.Search)
	if err != nil {
		return nil, err
	}

	// 轉換為 BookResponse
	bookResponses := make([]dto.BookResponse, len(books))
	for i, book := range books {
		bookResponses[i] = dto.BookResponse{
			ID:        book.ID,
			Title:     book.Title,
			Author:    book.Author,
			CreatedAt: book.CreatedAt,
			UpdatedAt: book.UpdatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dto.PaginationResponse{
		Data:       bookResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}
