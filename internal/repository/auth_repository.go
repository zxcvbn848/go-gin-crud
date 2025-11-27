package repository

import (
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	SaveRefreshToken(rt *models.RefreshToken) error
	FindRefreshToken(token string) (*models.RefreshToken, error)
	DeleteRefreshToken(token string) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository() AuthRepository {
	return &authRepository{
		db: database.DB,
	}
}

func (r *authRepository) SaveRefreshToken(rt *models.RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *authRepository) FindRefreshToken(token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *authRepository) DeleteRefreshToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.RefreshToken{}).Error
}
