package repository

import (
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	SaveRefreshToken(rt *models.RefreshToken) error
	FindRefreshToken(token string) (*models.RefreshToken, error)
	DeleteRefreshTokensByUserID(userID uint) error
	SaveBlacklistToken(bt *models.BlacklistToken) error
	IsTokenBlacklisted(token string) (bool, error)
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

func (r *authRepository) DeleteRefreshTokensByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}

func (r *authRepository) SaveBlacklistToken(bt *models.BlacklistToken) error {
	return r.db.Create(bt).Error
}

func (r *authRepository) IsTokenBlacklisted(token string) (bool, error) {
	var count int64
	err := r.db.Model(&models.BlacklistToken{}).Where("token = ?", token).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
