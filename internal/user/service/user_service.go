package service

import (
	"errors"
	"time"

	"qahub/internal/user/dto"
	"qahub/internal/user/model"
	"qahub/internal/user/store"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// IMPORTANT: This secret key should be loaded from a secure configuration, not hardcoded.
var jwtSecret = []byte("your-super-secret-key-change-it")

type UserService interface {
	Register(username, email, password string) (*dto.UserResponse, error)
	Login(username, password string) (string, error)
	GetUserProfile(userID int64) (*dto.UserResponse, error)
	UpdateUserProfile(user *model.User) error // Note: Input might also become a DTO
	DeleteUser(userID int64) error
}

type userService struct {
	userStore store.UserStore
}

func NewUserService(store store.UserStore) UserService {
	return &userService{userStore: store}
}

func (s *userService) Register(username, email, password string) (*dto.UserResponse, error) {
	// 检查用户名是否已存在
	if existingUser, _ := s.userStore.GetUserByEmail(email); existingUser != nil {
		return nil, errors.New("username already exists")
	}
	// 验证密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &model.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	newID, err := s.userStore.CreateUser(newUser)
	if err != nil {
		return nil, err
	}

	// 转换为DTO
	response := &dto.UserResponse{
		ID:       newID,
		Username: username,
		Email:    email,
	}

	return response, nil
}

func (s *userService) Login(username, password string) (string, error) {
	user, err := s.userStore.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // token于72小时后过期
		"iat":      time.Now().Unix(),                     // token的签发时间
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userService) GetUserProfile(userID int64) (*dto.UserResponse, error) {
	user, err := s.userStore.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	// 转换为DTO
	response := &dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	return response, nil
}

func (s *userService) UpdateUserProfile(user *model.User) error {
	return s.userStore.UpdateUser(user)
}

func (s *userService) DeleteUser(userID int64) error {
	return s.userStore.DeleteUser(userID)
}
