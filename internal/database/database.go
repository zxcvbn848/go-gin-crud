package database

import (
	"log"
	"os"
	"time"

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
		log.Fatal("❌ 無法連線到資料庫: ", err)
	}

	log.Println("✅ 成功連上資料庫")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
