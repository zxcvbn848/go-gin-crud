package database

import (
	"os"
	"time"

	"go-gin-crud/internal/logger"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	// 載入 .env 檔案（如果存在）
	// 忽略錯誤，因為 .env 檔案是可選的
	_ = godotenv.Load()

	// 從環境變數取得資料庫連線資訊，如果沒有則使用預設值
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3307")
	dbUser := getEnv("DB_USER", "gogin")
	dbPassword := getEnv("DB_PASSWORD", "a3935522")
	dbName := getEnv("DB_NAME", "goGinCRUD")

	dsn := dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})

	if err != nil {
		logger.Log.WithError(err).Fatal("無法連線到資料庫")
	}

	// 配置連接池參數
	// GORM 底層使用 Go 標準庫的 database/sql 包，該包內建連接池功能
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Log.WithError(err).Fatal("無法獲取底層 sql.DB 對象")
	}

	// 設置連接池參數
	// SetMaxIdleConns: 設置最大空閒連接數（默認值：2）
	// 空閒連接是指當前未使用的連接，保留這些連接可以避免頻繁建立/關閉連接
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns: 設置最大打開連接數（默認值：0，表示無限制）
	// 這是連接池的最大容量，超過此數量的請求會等待
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime: 設置連接的最大存活時間（默認值：0，表示永不過期）
	// 超過此時間的連接會被關閉並重新建立，避免使用過期的連接
	sqlDB.SetConnMaxLifetime(time.Hour)

	// SetConnMaxIdleTime: 設置連接的最大空閒時間（默認值：0，表示永不過期）
	// 空閒連接超過此時間會被關閉
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	logger.Log.Info("成功連上資料庫，連接池已配置")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
