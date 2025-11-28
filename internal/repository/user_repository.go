package repository

import (
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	FindAllWithPagination(page, pageSize int, search string) ([]models.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.DB,
	}
}

func (r *userRepository) Create(user *models.User) error {
	// GORM 會自動處理 CreatedAt 和 UpdatedAt（透過 autoCreateTime 和 autoUpdateTime 標籤）
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Model(user).Updates(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) FindAllWithPagination(page, pageSize int, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// 搜尋功能（只搜尋 email）
	query = WhereLike(query, "email", search)

	// 計算總數
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分頁查詢
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
