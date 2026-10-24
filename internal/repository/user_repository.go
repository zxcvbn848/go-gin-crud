package repository

import (
	"context"
	"time"

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

	// 報表用。收 ctx 是為了讓請求層逾時能真的中斷慢查詢 ——
	// 其餘方法沿用既有簽章，沒有一次改完是刻意的：報表是唯一會慢到
	// 需要被中斷的查詢，其他方法要改應該有各自的理由。
	CountAll(ctx context.Context) (int64, error)
	CountDailyCreated(ctx context.Context, from, to time.Time) ([]DailyCount, error)
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

// CountAll 未刪除的使用者總數
func (r *userRepository) CountAll(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&n).Error
	return n, err
}

// CountDailyCreated 區間內每日新增的使用者數
func (r *userRepository) CountDailyCreated(ctx context.Context, from, to time.Time) ([]DailyCount, error) {
	var rows []DailyCount
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Select("DATE(created_at) AS date, COUNT(*) AS count").
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("DATE(created_at)").
		Order("date").
		Scan(&rows).Error
	return rows, err
}
