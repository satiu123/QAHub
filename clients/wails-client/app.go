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

	// 服务实例
	UserService   *services.UserService
	QAService     *services.QAService
	SearchService *services.SearchService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化 gRPC 客户端 (连接到本地服务)
	// 注意：确保服务在对应端口运行
	client, err := services.NewGRPCClient(
		"localhost:50051", // user-service
		"localhost:50052", // qa-service
		"localhost:50053", // search-service
	)
	if err != nil {
		log.Printf("⚠️  Failed to create gRPC client: %v", err)
		log.Println("请确保服务在以下端口运行:")
		log.Println("  - User Service: localhost:50051")
		log.Println("  - QA Service: localhost:50052")
		log.Println("  - Search Service: localhost:50053")
		// 不要 fatal，允许应用启动，只是功能不可用
	}

	a.grpcClient = client
	a.UserService = services.NewUserService(client)
	a.QAService = services.NewQAService(client)
	a.SearchService = services.NewSearchService(client)

	log.Println("✅ QAHub Wails Client started successfully")
	_, _ = a.Login("saocong", "12345678") // 自动登录测试账号
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

// ===== 问答相关方法 (供前端调用) =====

// ListQuestions 获取问题列表
func (a *App) ListQuestions(page, pageSize int32) ([]services.Question, error) {
	questions, _, err := a.QAService.ListQuestions(a.ctx, page, pageSize)
	return questions, err
}

// GetQuestion 获取问题详情
func (a *App) GetQuestion(id int64) (*services.Question, error) {
	return a.QAService.GetQuestion(a.ctx, id)
}

// CreateQuestion 创建问题
func (a *App) CreateQuestion(title, content string) (*services.Question, error) {
	return a.QAService.CreateQuestion(a.ctx, title, content)
}

// UpdateQuestion 更新问题
func (a *App) UpdateQuestion(id int64, title, content string) (*services.Question, error) {
	return a.QAService.UpdateQuestion(a.ctx, id, title, content)
}

// DeleteQuestion 删除问题
func (a *App) DeleteQuestion(id int64) error {
	return a.QAService.DeleteQuestion(a.ctx, id)
}

// ListAnswers 获取回答列表
func (a *App) ListAnswers(questionID int64, page, pageSize int32) ([]services.Answer, error) {
	answers, _, err := a.QAService.ListAnswers(a.ctx, questionID, page, pageSize)
	return answers, err
}

// CreateAnswer 创建回答
func (a *App) CreateAnswer(questionID int64, content string) (*services.Answer, error) {
	return a.QAService.CreateAnswer(a.ctx, questionID, content)
}

// UpdateAnswer 更新回答
func (a *App) UpdateAnswer(id int64, content string) (*services.Answer, error) {
	return a.QAService.UpdateAnswer(a.ctx, id, content)
}

// DeleteAnswer 删除回答
func (a *App) DeleteAnswer(id int64) error {
	return a.QAService.DeleteAnswer(a.ctx, id)
}

// UpvoteAnswer 点赞回答
func (a *App) UpvoteAnswer(answerID int64) error {
	return a.QAService.UpvoteAnswer(a.ctx, answerID)
}

// DownvoteAnswer 取消点赞
func (a *App) DownvoteAnswer(answerID int64) error {
	return a.QAService.DownvoteAnswer(a.ctx, answerID)
}

// ListComments 获取评论列表
func (a *App) ListComments(answerID int64, page, pageSize int32) ([]services.Comment, error) {
	comments, _, err := a.QAService.ListComments(a.ctx, answerID, page, pageSize)
	return comments, err
}

// CreateComment 创建评论
func (a *App) CreateComment(answerID int64, content string) (*services.Comment, error) {
	return a.QAService.CreateComment(a.ctx, answerID, content)
}

// UpdateComment 更新评论
func (a *App) UpdateComment(id int64, content string) (*services.Comment, error) {
	return a.QAService.UpdateComment(a.ctx, id, content)
}

// DeleteComment 删除评论
func (a *App) DeleteComment(id int64) error {
	return a.QAService.DeleteComment(a.ctx, id)
}

// ===== 搜索相关方法 (供前端调用) =====

// SearchQuestions 搜索问题
func (a *App) SearchQuestions(query string, limit, offset int32) ([]services.SearchResult, error) {
	return a.SearchService.SearchQuestions(a.ctx, query, limit, offset)
}

// IndexAllQuestions 索引所有问题（仅用于测试/管理）
func (a *App) IndexAllQuestions() (string, error) {
	return a.SearchService.IndexAllQuestions(a.ctx)
}

// DeleteIndexAllQuestions 删除所有问题索引（仅用于测试/管理）
func (a *App) DeleteIndexAllQuestions() (string, error) {
	return a.SearchService.DeleteIndexAllQuestions(a.ctx)
}
