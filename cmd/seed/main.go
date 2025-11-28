package main

import (
	"log"

	"go-gin-crud/internal/database"
)

func main() {
	// 連線資料庫
	database.Connect()

	// 創建 Seeder
	seeder := database.NewSeeder()

	// 只更新 NULL 的時間戳，有值的不更新
	log.Println("📝 執行模式: 只更新 NULL 的時間戳（有值的不更新）")
	if err := seeder.SeedTimestamps(); err != nil {
		log.Fatalf("❌ Seeder 執行失敗: %v", err)
	}

	log.Println("🎉 Seeder 執行完成")
}
