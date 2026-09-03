package repository

import (
	"context"
	"time"

	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"

	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *models.Product) error
	FindByID(id uint) (*models.Product, error)
	Update(product *models.Product) error
	Delete(id uint) error
	FindAllWithPagination(page, pageSize int, search string) ([]models.Product, int64, error)

	// 報表用
	CountAll(ctx context.Context) (int64, error)
	CountDailyCreated(ctx context.Context, from, to time.Time) ([]DailyCount, error)
	SumStockValue(ctx context.Context) (float64, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository() ProductRepository {
	return &productRepository{
		db: database.DB,
	}
}

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) FindByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Update(product *models.Product) error {
	return r.db.Model(product).Updates(product).Error
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

func (r *productRepository) FindAllWithPagination(page, pageSize int, search string) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{})

	// 搜尋功能（搜尋 name 或 description）
	query = WhereLikeOr(query, search, "name", "description")

	// 計算總數
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分頁查詢
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// CountAll 未刪除的商品總數
func (r *productRepository) CountAll(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).Count(&n).Error
	return n, err
}

// CountDailyCreated 區間內每日新增的商品數
func (r *productRepository) CountDailyCreated(ctx context.Context, from, to time.Time) ([]DailyCount, error) {
	var rows []DailyCount
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Select("DATE(created_at) AS date, COUNT(*) AS count").
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("DATE(created_at)").
		Order("date").
		Scan(&rows).Error
	return rows, err
}

// SumStockValue 庫存總值（price × stock）。
//
// COALESCE 是必要的：表為空時 SUM 回傳 NULL，掃進 float64 會失敗。
func (r *productRepository) SumStockValue(ctx context.Context) (float64, error) {
	var v float64
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Select("COALESCE(SUM(price * stock), 0)").
		Scan(&v).Error
	return v, err
}
