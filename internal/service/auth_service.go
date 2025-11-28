package service

import (
	"errors"
	"go-gin-crud/internal/config"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCurrentPassword = errors.New("目前密碼不正確")

type AuthService interface {
	Register(email, password string) error
	Login(req dto.LoginRequest) (string, string, error)
	Refresh(refreshToken string) (string, error)
	Logout(accessToken string) error
	IsTokenBlacklisted(token string) (bool, error)
	GetUserProfile(userID uint) (*models.User, error)
	ChangePassword(userID uint, req dto.ChangePasswordRequest) error
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
		Role:     "user",
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
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString(config.AccessSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token: 7 天
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}).SignedString(config.RefreshSecret)
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
		return config.RefreshSecret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("refresh token 無效")
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	// 取得使用者以取得最新角色
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", errors.New("找不到使用者")
	}

	// 產生新的 Access Token
	newAccessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString(config.AccessSecret)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *authService) Logout(accessToken string) error {
	// 解析 token 獲取過期時間
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return config.AccessSecret, nil
	})
	if err != nil {
		return errors.New("token 無效")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("token claims 無效")
	}

	// 獲取 user_id
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return errors.New("token user_id 無效")
	}
	userID := uint(userIDFloat)

	// 獲取過期時間
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("token 過期時間無效")
	}

	expiresAt := time.Unix(int64(exp), 0)

	// 將 token 加入黑名單
	blacklistToken := &models.BlacklistToken{
		Token:     accessToken,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.authRepo.SaveBlacklistToken(blacklistToken); err != nil {
		return err
	}

	// 刪除該用戶的所有 refresh token
	return s.authRepo.DeleteRefreshTokensByUserID(userID)
}

func (s *authService) IsTokenBlacklisted(token string) (bool, error) {
	return s.authRepo.IsTokenBlacklisted(token)
}

func (s *authService) GetUserProfile(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *authService) ChangePassword(userID uint, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidCurrentPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)
	if err != nil {
		return err
	}

	user.Password = string(hashed)
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}
