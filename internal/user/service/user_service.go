package service

import (
	"context"
	"errors"
	"time"

	"qahub/internal/user/dto"
	"qahub/internal/user/model"
	"qahub/internal/user/store"
	"qahub/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, username, email, bio, password string) (*dto.UserResponse, error)
	Login(ctx context.Context, username, password string) (string, error)
	Logout(ctx context.Context, tokenString string, claims jwt.MapClaims) error
	GetUserProfile(ctx context.Context, userID int64) (*dto.UserResponse, error)
	UpdateUserProfile(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, userID int64) error
}

type userService struct {
	userStore store.UserStore
}

func NewUserService(store store.UserStore) UserService {
	return &userService{userStore: store}
}

func (s *userService) Register(ctx context.Context, username, email, bio, password string) (*dto.UserResponse, error) {
	// 检查邮箱是否已存在
	if existingUser, _ := s.userStore.GetUserByEmail(email); existingUser != nil {
		return nil, errors.New("该邮箱已被注册")
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
		Bio:      bio,
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
		Bio:      bio,
	}

	return response, nil
}

func (s *userService) Login(ctx context.Context, username, password string) (string, error) {
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
		"exp":      time.Now().Add(time.Hour * time.Duration(config.Conf.Services.UserService.TokenExpireHours)).Unix(), // token于72小时后过期
		"iat":      time.Now().Unix(),                                                                                   // token的签发时间
	}

	var jwtSecret = []byte(config.Conf.Services.UserService.JWTSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Logout 将 token 加入黑名单
func (s *userService) Logout(ctx context.Context, tokenString string, claims jwt.MapClaims) error {
	blacklister, ok := s.userStore.(store.TokenBlacklister)
	if !ok {
		// 如果当前的 userStore 没有实现黑名单功能，则静默返回
		// 这意味着登出操作在不支持黑名单的存储后端上无效，但不会报错
		return nil
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid token expiration")
	}

	// 计算剩余的过期时间
	remaining := time.Until(time.Unix(int64(exp), 0))
	if remaining <= 0 {
		return nil // Token 已过期，无需操作
	}

	return blacklister.AddToBlacklist(tokenString, remaining)
}

func (s *userService) GetUserProfile(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	user, err := s.userStore.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	// 转换为DTO
	response := &dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Bio:      user.Bio,
	}

	return response, nil
}

func (s *userService) UpdateUserProfile(ctx context.Context, user *model.User) error {
	return s.userStore.UpdateUser(user)
}

func (s *userService) DeleteUser(ctx context.Context, userID int64) error {
	return s.userStore.DeleteUser(userID)
}
