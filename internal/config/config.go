package config

import (
	"os"
	"strconv"
	"time"

	"go-gin-crud/internal/logger"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// DefaultBcryptCost 為 production 預設的 bcrypt 成本（OWASP 建議範圍，平衡安全與效能）
const DefaultBcryptCost = 12

// DefaultRequestTimeout 請求層逾時預設值
const DefaultRequestTimeout = 10 * time.Second

var (
	AccessSecret  []byte
	RefreshSecret []byte
	RedisAddr     string // 例: "localhost:6379"，空字串表示不啟用 Redis
	BcryptCost    int    // bcrypt 雜湊成本，可用 BCRYPT_COST 環境變數覆寫（測試環境建議設低以加速）

	// RequestTimeout 請求層逾時，可用 REQUEST_TIMEOUT 覆寫（time.ParseDuration 格式）
	RequestTimeout time.Duration
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
	RedisAddr = getEnv("REDIS_ADDR", "")

	if RedisAddr != "" {
		logger.Log.Info("Redis 位址已設定，將啟用 Book 快取")
	}

	// bcrypt 成本：預設 DefaultBcryptCost，可由 BCRYPT_COST 覆寫並限制在合法範圍
	BcryptCost = DefaultBcryptCost
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		if cost, err := strconv.Atoi(v); err == nil && cost >= bcrypt.MinCost && cost <= bcrypt.MaxCost {
			BcryptCost = cost
		} else {
			logger.Log.Warnf("BCRYPT_COST 值無效 (%s)，使用預設值 %d", v, DefaultBcryptCost)
		}
	}

	// 請求逾時：預設 DefaultRequestTimeout，可由 REQUEST_TIMEOUT 覆寫（例 "5s"、"1m"）
	RequestTimeout = DefaultRequestTimeout
	if v := os.Getenv("REQUEST_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			RequestTimeout = d
		} else {
			logger.Log.Warnf("REQUEST_TIMEOUT 值無效 (%s)，使用預設值 %s", v, DefaultRequestTimeout)
		}
	}

	logger.Log.Info("JWT 配置載入完成")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
