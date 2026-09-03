package repository

import (
	"context"
	"time"

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

	// 報表用
	CountAll(ctx context.Context) (int64, error)
	CountDailyCreated(ctx context.Context, from, to time.Time) ([]DailyCount, error)
	TopAuthors(ctx context.Context, limit int) ([]AuthorCount, error)
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
	query = WhereLikeOr(query, search, "title", "content")

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

// CountAll 未刪除的文章總數
func (r *postRepository) CountAll(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Post{}).Count(&n).Error
	return n, err
}

// CountDailyCreated 區間內每日新增的文章數
func (r *postRepository) CountDailyCreated(ctx context.Context, from, to time.Time) ([]DailyCount, error) {
	var rows []DailyCount
	err := r.db.WithContext(ctx).Model(&models.Post{}).
		Select("DATE(created_at) AS date, COUNT(*) AS count").
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("DATE(created_at)").
		Order("date").
		Scan(&rows).Error
	return rows, err
}

// TopAuthors 發文數排行。
//
// JOIN users 是為了帶出 email —— 若只回 author_id，service 層就得再查一次
// users，那才是真正的 N+1。
//
// ponytail: LIMIT 幫不上忙，要先算完所有作者的計數才知道前幾名是誰。
// 這支是目前最慢的查詢（見 docs/REPORT_API_BENCHMARK.md），也是後續
// Bulkhead 的保護對象。
func (r *postRepository) TopAuthors(ctx context.Context, limit int) ([]AuthorCount, error) {
	var rows []AuthorCount
	err := r.db.WithContext(ctx).Model(&models.Post{}).
		Select("posts.author_id AS author_id, users.email AS email, COUNT(posts.id) AS count").
		Joins("JOIN users ON users.id = posts.author_id").
		Group("posts.author_id, users.email").
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}
