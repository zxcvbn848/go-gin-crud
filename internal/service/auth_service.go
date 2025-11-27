package service

import (
	"errors"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var accessKey = []byte("ACCESS_SECRET")   // 建議改成環境變數
var refreshKey = []byte("REFRESH_SECRET") // 建議改成環境變數

type AuthService interface {
	Register(email, password string) error
	Login(req dto.LoginRequest) (string, string, error)
	Refresh(refreshToken string) (string, error)
	GetUserProfile(userID uint) (*models.User, error)
}

type authService struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository
}

func NewAuthService(userRepo repository.UserRepository, authRepo repository.AuthRepository) AuthService {
	return &authService{
		userRepo: userRepo,
		authRepo: authRepo,
	}
}

func (s *authService) Register(email, password string) error {
	// 檢查 Email 是否已存在
	_, err := s.userRepo.FindByEmail(email)
	if err == nil {
		return errors.New("email 已存在")
	}

	// 加密密碼
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}

	// 創建使用者
	user := &models.User{
		Email:    email,
		Password: string(hashed),
	}

	return s.userRepo.Create(user)
}

func (s *authService) Login(req dto.LoginRequest) (string, string, error) {
	// 查找使用者
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return "", "", errors.New("使用者不存在")
	}

	// 檢查密碼
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", "", errors.New("密碼錯誤")
	}

	// Access Token: 15 分鐘
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString(accessKey)
	if err != nil {
		return "", "", err
	}

	// Refresh Token: 7 天
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}).SignedString(refreshKey)
	if err != nil {
		return "", "", err
	}

	// 存 DB
	if err := s.authRepo.SaveRefreshToken(&models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *authService) Refresh(refreshToken string) (string, error) {
	// 檢查 DB 是否存在
	saved, err := s.authRepo.FindRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("refresh token 無效")
	}

	// 檢查是否過期
	if saved.ExpiresAt.Before(time.Now()) {
		return "", errors.New("refresh token 已過期")
	}

	// 驗證 JWT
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return refreshKey, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("refresh token 無效")
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	// 產生新的 Access Token
	newAccessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString(accessKey)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *authService) GetUserProfile(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}
