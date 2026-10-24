package database

import (
	"fmt"
	"math/rand/v2"
	"time"

	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/logger"
)

// BulkSpec 量產的筆數設定。零值代表該表不動。
type BulkSpec struct {
	Users    int
	Posts    int
	Products int
	Books    int
	// SpreadDays 把 created_at 平均散布在過去幾天內。
	//
	// 報表是按日期分組的，所有資料擠在同一天的話 GROUP BY 只會產生一列，
	// 量得出來的東西跟正式環境完全不同。
	SpreadDays int
	// BatchSize 每次 INSERT 的筆數。太大會撞 MySQL 的 max_allowed_packet，
	// 太小則來回次數過多。
	BatchSize int
}

// DefaultBulkSpec 供 cmd/seed 使用的預設量。
//
// ponytail: 這個量級（55 萬筆）是「在本機 MySQL 上跑得出可觀察的慢查詢」的
// 下限估計，不是實測值。實際要多少取決於機器與 buffer pool，觀察不到差異
// 就往上加 Posts。
var DefaultBulkSpec = BulkSpec{
	Users:      50_000,
	Posts:      400_000,
	Products:   50_000,
	Books:      50_000,
	SpreadDays: 365,
	BatchSize:  2_000,
}

// randomTimeWithin 回傳過去 days 天內的隨機時間點
func randomTimeWithin(days int) time.Time {
	if days <= 0 {
		return time.Now()
	}
	offset := time.Duration(rand.N(int64(days) * int64(24*time.Hour)))
	return time.Now().Add(-offset)
}

// SeedBulk 量產測試資料。
//
// 目的是讓報表 API 有足夠的資料量能觀察到真實的查詢成本 —— 沒有量的話
// 全表掃描和索引掃描的差異量不出來，後續的 Bulkhead 也沒有保護對象。
//
// 只新增，不刪除既有資料；email 與名稱都帶隨機後綴避免 unique 衝突。
func (s *Seeder) SeedBulk(spec BulkSpec) error {
	if spec.BatchSize <= 0 {
		spec.BatchSize = 1_000
	}

	// 每次執行帶一個唯一前綴，重複跑不會撞 users.email 的 unique index
	runID := time.Now().UnixNano()

	if spec.Users > 0 {
		if err := s.seedUsers(spec, runID); err != nil {
			return err
		}
	}
	if spec.Products > 0 {
		if err := s.seedProducts(spec, runID); err != nil {
			return err
		}
	}
	if spec.Books > 0 {
		if err := s.seedBooks(spec, runID); err != nil {
			return err
		}
	}
	// Posts 放最後：它需要既有的 user id 當 author
	if spec.Posts > 0 {
		if err := s.seedPosts(spec, runID); err != nil {
			return err
		}
	}

	logger.Log.Info("量產完成")
	return nil
}

func (s *Seeder) seedUsers(spec BulkSpec, runID int64) error {
	logger.Log.Infof("量產 users: %d 筆", spec.Users)
	rows := make([]models.User, 0, spec.Users)
	for i := 0; i < spec.Users; i++ {
		at := randomTimeWithin(spec.SpreadDays)
		role := "user"
		if i%50 == 0 {
			role = "admin"
		}
		rows = append(rows, models.User{
			Email: fmt.Sprintf("seed%d_%d@example.com", runID, i),
			// 固定的假雜湊值 —— 這些帳號不用於登入，不需要真的跑 bcrypt
			// （50,000 次 bcrypt cost 12 要跑好幾分鐘）
			Password:  "$2a$12$seedonlynotusableforloginXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
			Role:      role,
			CreatedAt: at,
			UpdatedAt: at,
		})
	}
	return s.insertBatches("users", rows, spec.BatchSize)
}

func (s *Seeder) seedProducts(spec BulkSpec, runID int64) error {
	logger.Log.Infof("量產 products: %d 筆", spec.Products)
	rows := make([]models.Product, 0, spec.Products)
	for i := 0; i < spec.Products; i++ {
		at := randomTimeWithin(spec.SpreadDays)
		rows = append(rows, models.Product{
			Name:        fmt.Sprintf("seed-product-%d-%d", runID, i),
			Description: "seeded for report benchmarking",
			Price:       float64(rand.N(100_000)) / 100,
			Stock:       rand.N(1_000),
			CreatedAt:   at,
			UpdatedAt:   at,
		})
	}
	return s.insertBatches("products", rows, spec.BatchSize)
}

func (s *Seeder) seedBooks(spec BulkSpec, runID int64) error {
	logger.Log.Infof("量產 books: %d 筆", spec.Books)
	rows := make([]models.Book, 0, spec.Books)
	for i := 0; i < spec.Books; i++ {
		at := randomTimeWithin(spec.SpreadDays)
		rows = append(rows, models.Book{
			Title:     fmt.Sprintf("seed-book-%d-%d", runID, i),
			Author:    fmt.Sprintf("seed-author-%d", rand.N(500)),
			CreatedAt: at,
			UpdatedAt: at,
		})
	}
	return s.insertBatches("books", rows, spec.BatchSize)
}

// seedPosts 需要既有的 user id 當 author_id。
//
// 只撈 id 而不是整個 user，因為 400,000 筆 post 只需要一個 id 池，
// 把 50,000 個完整的 user 讀進記憶體沒有意義。
func (s *Seeder) seedPosts(spec BulkSpec, runID int64) error {
	var authorIDs []uint
	if err := s.db.Model(&models.User{}).Pluck("id", &authorIDs).Error; err != nil {
		return fmt.Errorf("撈取 author id 失敗: %w", err)
	}
	if len(authorIDs) == 0 {
		return fmt.Errorf("users 表沒有資料，無法量產 posts")
	}

	logger.Log.Infof("量產 posts: %d 筆（作者池 %d 個）", spec.Posts, len(authorIDs))
	rows := make([]models.Post, 0, spec.Posts)
	for i := 0; i < spec.Posts; i++ {
		at := randomTimeWithin(spec.SpreadDays)
		rows = append(rows, models.Post{
			Title: fmt.Sprintf("seed-post-%d-%d", runID, i),
			// 內容刻意短 —— 報表不查 content，塞長文只是讓資料檔變大、
			// 拖慢 seed 本身，對查詢成本沒有幫助
			Content:   "seeded for report benchmarking",
			AuthorID:  authorIDs[rand.N(len(authorIDs))],
			CreatedAt: at,
			UpdatedAt: at,
		})
	}
	return s.insertBatches("posts", rows, spec.BatchSize)
}

// insertBatches 分批寫入並回報耗時
func (s *Seeder) insertBatches(table string, rows interface{}, batchSize int) error {
	start := time.Now()
	if err := s.db.CreateInBatches(rows, batchSize).Error; err != nil {
		logger.Log.WithError(err).Errorf("量產 %s 失敗", table)
		return err
	}
	logger.Log.Infof("量產 %s 完成，耗時 %s", table, time.Since(start).Round(time.Millisecond))
	return nil
}
