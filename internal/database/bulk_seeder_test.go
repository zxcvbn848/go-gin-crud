package database

import (
	"testing"
	"time"

	"go-gin-crud/internal/database/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// newTestSeeder 用 sqlite :memory: 建一個 Seeder，不碰真實 DB
func newTestSeeder(t *testing.T) *Seeder {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(
		&models.User{}, &models.Post{}, &models.Product{}, &models.Book{},
	))

	return &Seeder{db: db}
}

// smallSpec 測試用的小量設定。批次刻意小於總筆數，才會真的走到分批。
func smallSpec() BulkSpec {
	return BulkSpec{
		Users:      20,
		Posts:      50,
		Products:   10,
		Books:      10,
		SpreadDays: 30,
		BatchSize:  7,
	}
}

// TestSeedBulkSpreadsCreatedAt created_at 要散布在多個日期上。
//
// 報表是按日期分組的，全部擠在同一天的話 GROUP BY 只產生一列，
// 量出來的東西跟正式環境完全不同。
func TestSeedBulkSpreadsCreatedAt(t *testing.T) {
	s := newTestSeeder(t)
	spec := smallSpec()
	require.NoError(t, s.SeedBulk(spec))

	var distinctDays int64
	require.NoError(t, s.db.Model(&models.Post{}).
		Distinct("date(created_at)").Count(&distinctDays).Error)

	assert.Greater(t, distinctDays, int64(1), "created_at 應散布在多個日期，而非集中同一天")

	// 也不該散到未來或超出視窗。
	//
	// 用排序取單筆而不是 min()/max() —— sqlite 會把聚合結果回成字串，
	// 掃進 time.Time 會失敗，而 MySQL 不會。讀 model 欄位在兩邊都通。
	var earliest, latest models.Post
	require.NoError(t, s.db.Order("created_at asc").First(&earliest).Error)
	require.NoError(t, s.db.Order("created_at desc").First(&latest).Error)

	assert.False(t, latest.CreatedAt.After(time.Now().Add(time.Minute)), "不該有未來的時間")
	assert.False(t, earliest.CreatedAt.Before(time.Now().AddDate(0, 0, -spec.SpreadDays-1)), "不該超出 SpreadDays 視窗")
}
