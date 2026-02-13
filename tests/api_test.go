package tests

import (
	"bytes"
	"encoding/json"
	"go-gin-crud/internal/config"
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/routes"
	"go-gin-crud/internal/service"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB
var testRouter *gin.Engine
var testAuthService service.AuthService

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 在測試中隱藏 GORM 日誌
	})
	if err != nil {
		panic("無法連線測試資料庫: " + err.Error())
	}

	// 自動遷移
	db.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Product{},
		&models.Post{},
		&models.RefreshToken{},
		&models.BlacklistToken{},
	)

	return db
}

func setupTestRouter() *gin.Engine {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 設置測試環境變數
	os.Setenv("ACCESS_SECRET", "test_access_secret_key_for_testing_only")
	os.Setenv("REFRESH_SECRET", "test_refresh_secret_key_for_testing_only")

	// 載入配置
	config.Load()

	// 設置測試資料庫
	testDB = setupTestDB()
	database.DB = testDB

	// 創建路由器
	r := gin.New()

	// 註冊路由
	routes.RegisterHealthRoutes(r)
	testAuthService = routes.RegisterAuthRoutes(r)
	routes.RegisterBookRoutes(r, testAuthService, nil)
	routes.RegisterUserRoutes(r, testAuthService, nil)
	routes.RegisterProductRoutes(r, testAuthService, nil)
	routes.RegisterPostRoutes(r, testAuthService, nil)
	routes.RegisterCounterRoutes(r)
	routes.RegisterAccountRoutes(r)
	routes.RegisterTaskExecutorRoutes(r)

	testRouter = r

	return r
}

func makeRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
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
	testRouter.ServeHTTP(w, req)

	return w
}

func registerTestUser(t *testing.T, email, password string) string {
	req := map[string]string{
		"email":    email,
		"password": password,
	}
	w := makeRequest("POST", "/register", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	// 在登入前，先刪除該用戶的舊 refresh tokens（避免 UNIQUE 約束錯誤）
	var user models.User
	testDB.Where("email = ?", email).First(&user)
	if user.ID != 0 {
		testDB.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{})
	}

	// 登入獲取 token
	w = makeRequest("POST", "/login", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["access_token"]
}

func registerTestAdmin(t *testing.T) string {
	// 使用 service 創建管理員用戶
	userRepo := repository.NewUserRepository()
	authRepo := repository.NewAuthRepository()
	authService := service.NewAuthService(userRepo, authRepo)
	
	// 創建測試管理員
	err := authService.Register("admin@test.com", "password123")
	if err != nil && err.Error() != "email 已存在" {
		t.Fatal("無法創建測試管理員:", err)
	}

	// 更新角色為 admin
	var user models.User
	testDB.Where("email = ?", "admin@test.com").First(&user)
	user.Role = "admin"
	testDB.Save(&user)

	// 在登入前，先刪除該用戶的舊 refresh tokens（避免 UNIQUE 約束錯誤）
	testDB.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{})

	// 登入
	loginReq := dto.LoginRequest{
		Email:    "admin@test.com",
		Password: "password123",
	}
	accessToken, _, err := authService.Login(loginReq)
	if err != nil {
		t.Fatal("無法登入測試管理員:", err)
	}

	return accessToken
}

func TestMain(m *testing.M) {
	// 設置測試環境
	setupTestRouter()
	
	// 執行測試
	code := m.Run()
	
	// 清理
	if testDB != nil {
		sqlDB, _ := testDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
	
	os.Exit(code)
}

// setupTestDBForTest 為每個測試設置獨立的資料庫
func setupTestDBForTest() {
	testDB = setupTestDB()
	database.DB = testDB
}

