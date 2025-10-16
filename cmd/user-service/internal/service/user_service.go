//go:generate mockgen -source=../store/user_store.go -destination=user_store_mock.go -package=service
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"qahub/pkg/auth"
	"qahub/pkg/config"
	"qahub/user-service/internal/dto"
	"qahub/user-service/internal/model"
	"qahub/user-service/internal/store"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type UserService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, username, password string) (string, error)
	Logout(ctx context.Context, tokenString string, claims jwt.MapClaims) error
	ValidateToken(ctx context.Context, tokenString string) (auth.Identity, error)
	AuthInterceptor(publicMethods ...string) grpc.UnaryServerInterceptor
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

func (s *userService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	// 检查邮箱是否已存在
	if existingUser, _ := s.userStore.GetUserByEmail(ctx, req.Email); existingUser != nil {
		return nil, errors.New("该邮箱已被注册")
	}

	// 验证密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 使用 ToUser 方法转换
	newUser := req.ToUser(string(hashedPassword))

	newID, err := s.userStore.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	// 设置新创建的ID
	newUser.ID = newID

	// 使用 NewUserResponse 方法转换
	return dto.NewUserResponse(newUser), nil
}

func (s *userService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userStore.GetUserByUsername(ctx, username)
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
		log.Println("userStore does not support token blacklisting")
		return nil
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid token expiration")
	}

	remaining := time.Until(time.Unix(int64(exp), 0))
	if remaining <= 0 {
		return nil
	}

	return blacklister.AddToBlacklist(ctx, tokenString, remaining)
}

func (s *userService) ValidateToken(ctx context.Context, tokenString string) (auth.Identity, error) {
	blacklister, hasBlacklist := s.userStore.(store.TokenBlacklister)
	if hasBlacklist {
		isBlacklisted, err := blacklister.IsBlacklisted(ctx, tokenString)
		if err != nil {
			return auth.Identity{}, fmt.Errorf("failed to check blacklist: %w", err)
		}
		if isBlacklisted {
			return auth.Identity{}, errors.New("token is blacklisted")
		}
	}
	identity, err := auth.ParseToken(tokenString, []byte(config.Conf.Services.UserService.JWTSecret))
	if err != nil {
		return auth.Identity{}, fmt.Errorf("token parsing error: %w", err)
	}

	return identity, nil
}

func (s *userService) GetUserProfile(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	user, err := s.userStore.GetUserByID(ctx, userID)
	log.Println("userService GetUserProfile user:", user)
	if err != nil {
		return nil, err
	}

	// 使用 NewUserResponse 方法转换
	return dto.NewUserResponse(user), nil
}

func (s *userService) UpdateUserProfile(ctx context.Context, user *model.User) error {
	return s.userStore.UpdateUser(ctx, user)
}

func (s *userService) DeleteUser(ctx context.Context, userID int64) error {
	return s.userStore.DeleteUser(ctx, userID)
}

func (s *userService) AuthInterceptor(publicMethods ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log.Println("AuthInterceptor called for method:", info.FullMethod)
		// 检查是否在白名单中
		if slices.Contains(publicMethods, info.FullMethod) {
			// 白名单路径，跳过认证
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "缺少认证信息 (metadata)")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "请求未包含授权标头")
		}

		authHeader := authHeaders[0]
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // 如果没有 "Bearer " 前缀，TrimPrefix 不会改变字符串
			return nil, status.Errorf(codes.Unauthenticated, "授权标头格式不正确，需要 'Bearer ' 前缀")
		}

		// 调用 user-service 的 ValidateToken 方法验证 token
		validateResp, err := s.ValidateToken(ctx, tokenString)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token 验证失败: %v", err)
		}

		// 验证成功，将用户信息注入到 context 中
		newCtx := auth.WithIdentity(ctx, validateResp)

		// 继续处理请求
		return handler(newCtx, req)
	}
}
