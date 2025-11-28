package config

import (
	"os"

	"go-gin-crud/internal/logger"

	"github.com/joho/godotenv"
)

var (
	AccessSecret  []byte
	RefreshSecret []byte
)

// Load 載入環境變數並初始化配置
func Load() {
	// 載入 .env 檔案（如果存在）
	_ = godotenv.Load()

	// 從環境變數讀取 JWT Secrets
	accessSecret := getEnv("ACCESS_SECRET", "ACCESS_SECRET")
	refreshSecret := getEnv("REFRESH_SECRET", "REFRESH_SECRET")

	// 檢查是否使用預設值（安全性警告）
	if accessSecret == "ACCESS_SECRET" {
		logger.Log.Warn("使用預設的 ACCESS_SECRET，建議在 .env 檔案中設定自訂值")
	}
	if refreshSecret == "REFRESH_SECRET" {
		logger.Log.Warn("使用預設的 REFRESH_SECRET，建議在 .env 檔案中設定自訂值")
	}

	AccessSecret = []byte(accessSecret)
	RefreshSecret = []byte(refreshSecret)

	logger.Log.Info("JWT 配置載入完成")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

