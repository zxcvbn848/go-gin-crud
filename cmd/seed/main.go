package main

import (
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/logger"
)

func main() {
	// 連線資料庫
	database.Connect()

	// 創建 Seeder
	seeder := database.NewSeeder()

	// 只更新 NULL 的時間戳，有值的不更新
	logger.Log.Info("執行模式: 只更新 NULL 的時間戳（有值的不更新）")
	if err := seeder.SeedTimestamps(); err != nil {
		logger.Log.WithError(err).Fatal("Seeder 執行失敗")
	}

	logger.Log.Info("Seeder 執行完成")
}
