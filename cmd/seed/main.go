package main

import (
	"flag"

	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/logger"
)

func main() {
	bulk := flag.Bool("bulk", false, "量產測試資料（供報表 API 效能觀察用）")
	users := flag.Int("users", database.DefaultBulkSpec.Users, "量產 users 筆數")
	posts := flag.Int("posts", database.DefaultBulkSpec.Posts, "量產 posts 筆數")
	products := flag.Int("products", database.DefaultBulkSpec.Products, "量產 products 筆數")
	books := flag.Int("books", database.DefaultBulkSpec.Books, "量產 books 筆數")
	spreadDays := flag.Int("spread-days", database.DefaultBulkSpec.SpreadDays, "created_at 散布在過去幾天內")
	batchSize := flag.Int("batch-size", database.DefaultBulkSpec.BatchSize, "每批 INSERT 筆數")
	flag.Parse()

	// 連線資料庫
	database.Connect()

	// 創建 Seeder
	seeder := database.NewSeeder()

	if *bulk {
		logger.Log.Info("執行模式: 量產測試資料")

		// 量產通常是對一個乾淨的 bench schema 跑（例如 DB_NAME=goGinCRUD_bench），
		// 那裡還沒有表，所以先建結構。既有的表 AutoMigrate 不會動。
		if err := database.DB.AutoMigrate(
			&models.User{}, &models.Post{}, &models.Product{}, &models.Book{},
		); err != nil {
			logger.Log.WithError(err).Fatal("資料庫遷移失敗")
		}

		spec := database.BulkSpec{
			Users:      *users,
			Posts:      *posts,
			Products:   *products,
			Books:      *books,
			SpreadDays: *spreadDays,
			BatchSize:  *batchSize,
		}
		if err := seeder.SeedBulk(spec); err != nil {
			logger.Log.WithError(err).Fatal("量產失敗")
		}
		logger.Log.Info("Seeder 執行完成")
		return
	}

	// 只更新 NULL 的時間戳，有值的不更新
	logger.Log.Info("執行模式: 只更新 NULL 的時間戳（有值的不更新）")
	if err := seeder.SeedTimestamps(); err != nil {
		logger.Log.WithError(err).Fatal("Seeder 執行失敗")
	}

	logger.Log.Info("Seeder 執行完成")
}
