package service

import (
	"errors"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("secret_key_123") // 建議改成環境變數

type AuthService interface {
	Register(email, password string) error
	Login(email, password string) (string, error)
	GetUserProfile(userID uint) (*models.User, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) Register(email, password string) error {
	// 檢查 Email 是否已存在
	_, err := s.userRepo.FindByEmail(email)
	if err == nil {
		return errors.New("Email 已存在")
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

func (s *authService) Login(email, password string) (string, error) {
	// 查找使用者
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("帳號或密碼錯誤")
	}

	// 檢查密碼
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("帳號或密碼錯誤")
	}

	// 生成 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *authService) GetUserProfile(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

