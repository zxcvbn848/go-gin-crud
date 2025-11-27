package repository

import (
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"

	"gorm.io/gorm"
)

type PostRepository interface {
	Create(post *models.Post) error
	FindByID(id uint) (*models.Post, error)
	Update(post *models.Post) error
	Delete(id uint) error
	FindAllWithPagination(page, pageSize int, search string) ([]models.Post, int64, error)
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository() PostRepository {
	return &postRepository{
		db: database.DB,
	}
}

func (r *postRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *postRepository) FindByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("Author").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) Update(post *models.Post) error {
	return r.db.Model(post).Updates(post).Error
}

func (r *postRepository) Delete(id uint) error {
	return r.db.Delete(&models.Post{}, id).Error
}

func (r *postRepository) FindAllWithPagination(page, pageSize int, search string) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.Model(&models.Post{}).Preload("Author")

	// 搜尋功能（搜尋 title 或 content）
	if search != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 計算總數
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分頁查詢（按創建時間倒序）
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

