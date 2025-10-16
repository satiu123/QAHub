package main

import (
	"context"
	"fmt"
	"log"
	"qahub/pkg/util"

	"changeme/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	grpcClient *services.GRPCClient

	// 服务实例
	UserService         *services.UserService
	QAService           *services.QAService
	SearchService       *services.SearchService
	NotificationService *services.NotificationService
	NotificationStream  *services.NotificationStream

	// 前端通知回调
	onNotificationReceived func(notification *services.Notification)
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
	// 注意:确保服务在对应端口运行
	client, err := services.NewGRPCClient(
		"localhost:50051", // user-service
		"localhost:50052", // qa-service
		"localhost:50053", // search-service
		"localhost:50054", // notification-service
	)
	if err != nil {
		log.Printf("⚠️  Failed to create gRPC client: %v", err)
		log.Println("请确保服务在以下端口运行:")
		log.Println("  - User Service: localhost:50051")
		log.Println("  - QA Service: localhost:50052")
		log.Println("  - Search Service: localhost:50053")
		log.Println("  - Notification Service: localhost:50054")
		log.Println("⚠️  应用将以离线模式启动")

		// 显示错误对话框通知用户
		_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "服务连接失败",
			Message: "无法连接到后端服务。\n请确保所有服务正在运行,或稍后重试。",
		})
		return
	}

	a.grpcClient = client
	a.UserService = services.NewUserService(client)
	a.QAService = services.NewQAService(client)
	a.SearchService = services.NewSearchService(client)
	a.NotificationService = services.NewNotificationService(client)

	// 初始化通知流
	a.NotificationStream = services.NewNotificationStream(client)

	// 添加通知处理器:收到通知时发送到前端
	a.NotificationStream.AddHandler(func(notification *services.Notification) {
		log.Printf("📨 Received notification in app: %s", notification.Content)
		// 发送事件到前端
		runtime.EventsEmit(a.ctx, "notification:received", notification)
	})

	log.Println("✅ QAHub Wails Client started successfully")
	// _, _ = a.Login("saocong", "12345678") // 自动登录测试账号
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	if a.NotificationStream != nil {
		a.NotificationStream.Stop()
	}
	if a.grpcClient != nil {
		util.Cleanup("gRPC client", a.grpcClient.Close)
	}
}

// ===== 系统相关方法 =====

// IsServiceConnected 检查服务是否已连接
func (a *App) IsServiceConnected() bool {
	return a.grpcClient != nil
}

// GetServiceStatus 获取服务连接状态
func (a *App) GetServiceStatus() map[string]interface{} {
	return map[string]interface{}{
		"connected": a.grpcClient != nil,
		"message": func() string {
			if a.grpcClient != nil {
				return "已连接到后端服务"
			}
			return "未连接到后端服务,请确保服务正在运行"
		}(),
	}
}

// ===== 用户相关方法 (供前端调用) =====

// Login 用户登录
func (a *App) Login(username, password string) (*services.LoginResponse, error) {
	if a.UserService == nil {
		return nil, fmt.Errorf("服务未连接,请先启动后端服务")
	}
	return a.UserService.Login(a.ctx, services.LoginRequest{
		Username: username,
		Password: password,
	})
}

// Register 用户注册
func (a *App) Register(username, email, password string) (*services.RegisterResponse, error) {
	if a.UserService == nil {
		return nil, fmt.Errorf("服务未连接,请先启动后端服务")
	}
	return a.UserService.Register(a.ctx, services.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	})
}

// Logout 用户登出
func (a *App) Logout() {
	if a.UserService != nil {
		a.UserService.Logout()
	}
}

// GetCurrentUser 获取当前登录用户信息
func (a *App) GetCurrentUser() (*services.UserProfile, error) {
	if a.UserService == nil {
		return nil, fmt.Errorf("服务未连接,请先启动后端服务")
	}
	return a.UserService.GetCurrentUser(a.ctx)
}

// IsLoggedIn 检查是否已登录
func (a *App) IsLoggedIn() bool {
	if a.UserService == nil {
		return false
	}
	return a.UserService.IsLoggedIn()
}

// GetUsername 获取当前用户名
func (a *App) GetUsername() string {
	if a.UserService == nil {
		return ""
	}
	return a.UserService.GetUsername()
}

// ===== 问答相关方法 (供前端调用) =====

// ListQuestions 获取问题列表
func (a *App) ListQuestions(page, pageSize int32) ([]services.Question, error) {
	if a.QAService == nil {
		return nil, fmt.Errorf("服务未连接,请先启动后端服务")
	}
	questions, _, err := a.QAService.ListQuestions(a.ctx, page, pageSize)
	return questions, err
}

// GetQuestion 获取问题详情
func (a *App) GetQuestion(id int64) (*services.Question, error) {
	if a.QAService == nil {
		return nil, fmt.Errorf("服务未连接,请先启动后端服务")
	}
	return a.QAService.GetQuestion(a.ctx, id)
}

// CreateQuestion 创建问题
func (a *App) CreateQuestion(title, content string) (*services.Question, error) {
	if a.QAService == nil {
		return nil, fmt.Errorf("服务未连接,请先启动后端服务")
	}
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

// ===== 通知相关方法 (供前端调用) =====

// NotificationListResult 通知列表结果
type NotificationListResult struct {
	Notifications []services.Notification `json:"notifications"`
	Total         int64                   `json:"total"`
	UnreadCount   int64                   `json:"unread_count"`
}

// GetNotifications 获取通知列表
func (a *App) GetNotifications(page int32, pageSize int64, unreadOnly bool) (*NotificationListResult, error) {
	notifications, total, unreadCount, err := a.NotificationService.GetNotifications(a.ctx, page, pageSize, unreadOnly)
	if err != nil {
		return nil, err
	}
	return &NotificationListResult{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
	}, nil
}

// GetUnreadCount 获取未读通知数量
func (a *App) GetUnreadCount() (int64, error) {
	return a.NotificationService.GetUnreadCount(a.ctx)
}

// MarkAsRead 标记通知为已读
func (a *App) MarkAsRead(notificationIDs []string, markAll bool) (int64, error) {
	return a.NotificationService.MarkAsRead(a.ctx, notificationIDs, markAll)
}

// DeleteNotification 删除通知
func (a *App) DeleteNotification(notificationID string) error {
	return a.NotificationService.DeleteNotification(a.ctx, notificationID)
}

// StartNotificationStream 启动通知流连接
func (a *App) StartNotificationStream() error {
	if a.NotificationStream == nil {
		return fmt.Errorf("notification stream not initialized")
	}
	return a.NotificationStream.Start()
}

// StopNotificationStream 停止通知流连接
func (a *App) StopNotificationStream() {
	if a.NotificationStream != nil {
		a.NotificationStream.Stop()
	}
}

// IsNotificationStreamConnected 检查通知流是否已连接
func (a *App) IsNotificationStreamConnected() bool {
	if a.NotificationStream == nil {
		return false
	}
	return a.NotificationStream.IsConnected()
}
