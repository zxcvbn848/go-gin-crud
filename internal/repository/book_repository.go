package repository

import (
	"context"

	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"

	"gorm.io/gorm"
)

type BookRepository interface {
	Create(book *models.Book) error
	FindByID(id uint) (*models.Book, error)
	Update(book *models.Book) error
	Delete(id uint) error
	FindAllWithPagination(page, pageSize int, search string) ([]models.Book, int64, error)

	// 報表用
	CountAll(ctx context.Context) (int64, error)
}

type bookRepository struct {
	db *gorm.DB
}

func NewBookRepository() BookRepository {
	return &bookRepository{
		db: database.DB,
	}
}

func (r *bookRepository) Create(book *models.Book) error {
	return r.db.Create(book).Error
}

func (r *bookRepository) FindByID(id uint) (*models.Book, error) {
	var book models.Book
	err := r.db.First(&book, id).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *bookRepository) Update(book *models.Book) error {
	return r.db.Model(book).Updates(book).Error
}

func (r *bookRepository) Delete(id uint) error {
	return r.db.Delete(&models.Book{}, id).Error
}

func (r *bookRepository) FindAllWithPagination(page, pageSize int, search string) ([]models.Book, int64, error) {
	var books []models.Book
	var total int64

	query := r.db.Model(&models.Book{})

	// 搜尋功能
	query = WhereLikeOr(query, search, "title", "author")

	// 計算總數
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分頁查詢
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&books).Error; err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

// CountAll 未刪除的書籍總數
func (r *bookRepository) CountAll(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Book{}).Count(&n).Error
	return n, err
}
