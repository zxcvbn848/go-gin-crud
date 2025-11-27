package service

import (
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/repository"
	"math"
)

type BookService interface {
	CreateBook(req dto.CreateBookRequest) (*models.Book, error)
	GetBookByID(id uint) (*models.Book, error)
	UpdateBook(id uint, req dto.UpdateBookRequest) (*models.Book, error)
	DeleteBook(id uint) error
	GetBooksWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error)
}

type bookService struct {
	bookRepo repository.BookRepository
}

func NewBookService(bookRepo repository.BookRepository) BookService {
	return &bookService{
		bookRepo: bookRepo,
	}
}

func (s *bookService) CreateBook(req dto.CreateBookRequest) (*models.Book, error) {
	book := &models.Book{
		Title:  req.Title,
		Author: req.Author,
	}

	if err := s.bookRepo.Create(book); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *bookService) GetBookByID(id uint) (*models.Book, error) {
	return s.bookRepo.FindByID(id)
}

func (s *bookService) UpdateBook(id uint, req dto.UpdateBookRequest) (*models.Book, error) {
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

	return book, nil
}

func (s *bookService) DeleteBook(id uint) error {
	return s.bookRepo.Delete(id)
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

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dto.PaginationResponse{
		Data:       books,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

