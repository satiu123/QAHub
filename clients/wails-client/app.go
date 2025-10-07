package main

import (
	"context"
	"log"

	"changeme/services"
)

// App struct
type App struct {
	ctx        context.Context
	grpcClient *services.GRPCClient

	// 用户服务
	UserService *services.UserService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化 gRPC 客户端 (连接到本地 user-service)
	// 注意：确保 user-service 在 localhost:50051 运行
	client, err := services.NewGRPCClient("localhost:50051")
	if err != nil {
		log.Printf("⚠️  Failed to create gRPC client: %v", err)
		log.Println("请确保 user-service 在 localhost:50051 运行")
		// 不要 fatal，允许应用启动，只是功能不可用
	}

	a.grpcClient = client
	a.UserService = services.NewUserService(client)

	log.Println("✅ QAHub Wails Client started successfully")
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	if a.grpcClient != nil {
		a.grpcClient.Close()
	}
}

// ===== 用户相关方法 (供前端调用) =====

// Login 用户登录
func (a *App) Login(username, password string) (*services.LoginResponse, error) {
	return a.UserService.Login(a.ctx, services.LoginRequest{
		Username: username,
		Password: password,
	})
}

// Register 用户注册
func (a *App) Register(username, email, password string) (*services.RegisterResponse, error) {
	return a.UserService.Register(a.ctx, services.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	})
}

// Logout 用户登出
func (a *App) Logout() {
	a.UserService.Logout()
}

// GetCurrentUser 获取当前登录用户信息
func (a *App) GetCurrentUser() (*services.UserProfile, error) {
	return a.UserService.GetCurrentUser(a.ctx)
}

// IsLoggedIn 检查是否已登录
func (a *App) IsLoggedIn() bool {
	return a.UserService.IsLoggedIn()
}

// GetUsername 获取当前用户名
func (a *App) GetUsername() string {
	return a.UserService.GetUsername()
}

