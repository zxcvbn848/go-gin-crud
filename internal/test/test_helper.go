package test

import (
	"bytes"
	"encoding/json"
	"go-gin-crud/internal/config"
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/routes"
	"go-gin-crud/internal/service"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	TestDB     *gorm.DB
	TestRouter *gin.Engine
	TestAuth   service.AuthService
)

// SetupTestDB 設置測試資料庫（使用 SQLite 記憶體資料庫）
func SetupTestDB() *gorm.DB {
	var err error
	TestDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("無法連線測試資料庫: " + err.Error())
	}

	// 自動遷移
	if err := TestDB.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Product{},
		&models.Post{},
		&models.RefreshToken{},
		&models.BlacklistToken{},
	); err != nil {
		panic("測試資料庫遷移失敗: " + err.Error())
	}

	return TestDB
}

// SetupTestRouter 設置測試路由器
func SetupTestRouter() *gin.Engine {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 載入配置
	config.Load()

	// 設置測試資料庫
	originalDB := database.DB
	database.DB = SetupTestDB()

	// 創建路由器
	r := gin.New()

	// 註冊路由
	TestAuth = routes.RegisterAuthRoutes(r)
	routes.RegisterBookRoutes(r, TestAuth, nil)
	routes.RegisterUserRoutes(r, TestAuth, nil)
	routes.RegisterProductRoutes(r, TestAuth, nil)
	routes.RegisterPostRoutes(r, TestAuth, nil)

	TestRouter = r

	// 恢復原始資料庫（測試結束後）
	defer func() {
		database.DB = originalDB
	}()

	return r
}

// CleanupTestDB 清理測試資料庫
func CleanupTestDB() {
	if TestDB != nil {
		sqlDB, _ := TestDB.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}

// MakeRequest 發送 HTTP 請求的輔助函數
func MakeRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	return w
}

// RegisterTestUser 註冊測試用戶並返回 token
func RegisterTestUser(email, password string) (string, string, error) {
	// 創建用戶
	user := &models.User{
		Email:    email,
		Password: password, // 注意：實際應該加密，這裡簡化
		Role:     "user",
	}
	TestDB.Create(user)

	// 登入獲取 token
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}
	w := MakeRequest("POST", "/login", loginReq, "")

	var response map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &response)

	return response["access_token"], response["refresh_token"], nil
}

// RegisterTestAdmin 註冊測試管理員並返回 token
func RegisterTestAdmin(email, password string) (string, string, error) {
	// 創建管理員
	user := &models.User{
		Email:    email,
		Password: password, // 注意：實際應該加密，這裡簡化
		Role:     "admin",
	}
	TestDB.Create(user)

	// 登入獲取 token
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}
	w := MakeRequest("POST", "/login", loginReq, "")

	var response map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &response)

	return response["access_token"], response["refresh_token"], nil
}

func init() {
	// 設置測試環境變數
	_ = os.Setenv("ACCESS_SECRET", "test_access_secret_key_for_testing_only")
	_ = os.Setenv("REFRESH_SECRET", "test_refresh_secret_key_for_testing_only")
}
