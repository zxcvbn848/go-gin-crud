package database

import (
	"time"

	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/logger"

	"gorm.io/gorm"
)

// Seeder 提供資料庫種子資料功能
type Seeder struct {
	db *gorm.DB
}

// NewSeeder 創建新的 Seeder 實例
func NewSeeder() *Seeder {
	return &Seeder{
		db: DB,
	}
}

// updateTimestampField 更新指定表的時間戳欄位（只更新 NULL 的欄位）
func (s *Seeder) updateTimestampField(model interface{}, tableName, fieldName string, now time.Time) error {
	result := s.db.Model(model).
		Where(fieldName+" IS NULL").
		Update(fieldName, now)
	if result.Error != nil {
		logger.Log.WithError(result.Error).Errorf("更新 %s.%s 失敗", tableName, fieldName)
		return result.Error
	}
	if result.RowsAffected > 0 {
		logger.Log.Infof("更新 %s.%s: %d 筆記錄", tableName, fieldName, result.RowsAffected)
	}
	return nil
}

// SeedTimestamps 為所有現有記錄設定 created_at 和 updated_at
// 只更新 NULL 的欄位，有值的欄位保持不變
func (s *Seeder) SeedTimestamps() error {
	now := time.Now()
	logger.Log.Info("開始更新既有資料的時間戳（只更新 NULL 欄位）...")

	// 更新 Users 表
	if err := s.updateTimestampField(&models.User{}, "Users", "created_at", now); err != nil {
		return err
	}
	if err := s.updateTimestampField(&models.User{}, "Users", "updated_at", now); err != nil {
		return err
	}

	// 更新 Books 表
	if err := s.updateTimestampField(&models.Book{}, "Books", "created_at", now); err != nil {
		return err
	}
	if err := s.updateTimestampField(&models.Book{}, "Books", "updated_at", now); err != nil {
		return err
	}

	// 更新 Products 表
	if err := s.updateTimestampField(&models.Product{}, "Products", "created_at", now); err != nil {
		return err
	}
	if err := s.updateTimestampField(&models.Product{}, "Products", "updated_at", now); err != nil {
		return err
	}

	// 更新 Posts 表（只更新 updated_at，created_at 通常已有值）
	if err := s.updateTimestampField(&models.Post{}, "Posts", "updated_at", now); err != nil {
		return err
	}

	// 更新 RefreshTokens 表
	if err := s.updateTimestampField(&models.RefreshToken{}, "RefreshTokens", "created_at", now); err != nil {
		return err
	}
	if err := s.updateTimestampField(&models.RefreshToken{}, "RefreshTokens", "updated_at", now); err != nil {
		return err
	}

	// 更新 BlacklistTokens 表（只更新 updated_at）
	if err := s.updateTimestampField(&models.BlacklistToken{}, "BlacklistTokens", "updated_at", now); err != nil {
		return err
	}

	logger.Log.Info("時間戳更新完成")
	return nil
}
