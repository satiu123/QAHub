package services

import (
	"context"
	"fmt"
	"log"

	userpb "wails-client/api/proto/user"

	"google.golang.org/protobuf/types/known/emptypb"
)

// UserService 用户服务的业务逻辑层
type UserService struct {
	client *GRPCClient
}

// NewUserService 创建用户服务实例
func NewUserService(client *GRPCClient) *UserService {
	return &UserService{client: client}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Success bool   `json:"success"`
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

// UserProfile 用户信息
type UserProfile struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// 调用 gRPC 登录
	resp, err := s.client.UserClient.Login(ctx, &userpb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: fmt.Sprintf("登录失败: %v", err),
		}, nil
	}

	// 保存 token (username 暂时从请求中获取，后续可以从 token 解析)
	s.client.SetAuth(resp.Token, 0, req.Username)

	return &LoginResponse{
		Success: true,
		Token:   resp.Token,
		Message: "登录成功",
	}, nil
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// 调用 gRPC 注册
	resp, err := s.client.UserClient.Register(ctx, &userpb.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return &RegisterResponse{
			Success: false,
			Message: fmt.Sprintf("注册失败: %v", err),
		}, nil
	}

	return &RegisterResponse{
		Success: true,
		UserID:  resp.User.Id,
		Message: "注册成功，请登录",
	}, nil
}

// GetProfile 获取用户信息
func (s *UserService) GetProfile(ctx context.Context, userID int64) (*UserProfile, error) {
	authCtx := s.client.NewAuthContext(ctx)
	resp, err := s.client.UserClient.GetUserProfile(authCtx, &userpb.GetUserProfileRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	log.Println("Fetched user profile:", resp.User.CreatedAt.AsTime().Format("2006-01-02 15:04:05"))
	return &UserProfile{
		UserID:    resp.User.Id,
		Username:  resp.User.Username,
		Email:     resp.User.Email,
		Bio:       resp.User.Bio,
		CreatedAt: resp.User.CreatedAt.AsTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// GetCurrentUser 获取当前登录用户信息
func (s *UserService) GetCurrentUser(ctx context.Context) (*UserProfile, error) {
	if !s.client.IsAuthenticated() {
		return nil, fmt.Errorf("用户未登录")
	}

	// 如果有存储的 userID，直接获取
	if s.client.GetUserID() > 0 {
		return s.GetProfile(ctx, s.client.GetUserID())
	}

	// 否则需要先验证 token 获取 userID
	authCtx := s.client.NewAuthContext(ctx)
	validateResp, err := s.client.UserClient.ValidateToken(authCtx, &userpb.ValidateTokenRequest{
		JwtToken: s.client.GetToken(),
	})
	if err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	// 更新 userID
	s.client.SetAuth(s.client.GetToken(), validateResp.UserId, validateResp.Username)

	return s.GetProfile(ctx, validateResp.UserId)
}

// Logout 用户登出
func (s *UserService) Logout() {
	if !s.client.IsAuthenticated() {
		log.Println("用户未登录，无需登出")
		return
	}

	authCtx := s.client.NewAuthContext(context.Background())
	_, err := s.client.UserClient.Logout(authCtx, &emptypb.Empty{})
	if err != nil {
		log.Printf("登出失败: %v", err)
	} else {
		log.Println("登出成功")
	}
	s.client.ClearAuth()
}

// IsLoggedIn 检查是否已登录
func (s *UserService) IsLoggedIn() bool {
	return s.client.IsAuthenticated()
}

// GetUsername 获取当前用户名
func (s *UserService) GetUsername() string {
	return s.client.GetUsername()
}

func (s *UserService) SetAuth() {
	s.client.SetAuth(
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjAyNDIxMzIsImlhdCI6MTc2MDE1NTczMiwidXNlcl9pZCI6MywidXNlcm5hbWUiOiJzYW9jb25nIn0.Q5VnlVrhoshVFblwr-Nht4708o5TCek5EiasMEV2tHk",
		3,
		"saocong",
	)
}
