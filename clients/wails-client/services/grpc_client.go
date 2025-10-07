package services

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	// 动态导入 proto (通过 go.work 工作区)
	// 注意：需要在项目根目录运行 wails dev

	qapb "qahub/api/proto/qa"
	userpb "qahub/api/proto/user"
)

// GRPCClient 封装 gRPC 服务客户端
type GRPCClient struct {
	userConn *grpc.ClientConn
	qaConn   *grpc.ClientConn

	UserClient userpb.UserServiceClient
	QAClient   qapb.QAServiceClient

	// 存储当前用户的 token 和信息
	token    string
	userID   int64
	username string
}

// NewGRPCClient 创建新的 gRPC 客户端连接
func NewGRPCClient(userAddr, qaAddr string) (*GRPCClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// 连接 User Service
	userConn, err := grpc.NewClient(userAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	// 连接 QA Service
	qaConn, err := grpc.NewClient(qaAddr, opts...)
	if err != nil {
		userConn.Close()
		return nil, fmt.Errorf("failed to connect to qa service: %w", err)
	}

	client := &GRPCClient{
		userConn:   userConn,
		qaConn:     qaConn,
		UserClient: userpb.NewUserServiceClient(userConn),
		QAClient:   qapb.NewQAServiceClient(qaConn),
	}

	log.Println("✅ gRPC clients connected successfully")
	log.Printf("  - User Service: %s", userAddr)
	log.Printf("  - QA Service: %s", qaAddr)
	return client, nil
}

// Close 关闭所有连接
func (c *GRPCClient) Close() error {
	if c.userConn != nil {
		c.userConn.Close()
	}
	if c.qaConn != nil {
		c.qaConn.Close()
	}
	return nil
}

// SetAuth 设置认证信息
func (c *GRPCClient) SetAuth(token string, userID int64, username string) {
	c.token = token
	c.userID = userID
	c.username = username
}

// ClearAuth 清除认证信息
func (c *GRPCClient) ClearAuth() {
	c.token = ""
	c.userID = 0
	c.username = ""
}

// GetToken 获取当前 token
func (c *GRPCClient) GetToken() string {
	return c.token
}

// GetUserID 获取用户 ID
func (c *GRPCClient) GetUserID() int64 {
	return c.userID
}

// GetUsername 获取用户名
func (c *GRPCClient) GetUsername() string {
	return c.username
}

// IsAuthenticated 检查是否已登录
func (c *GRPCClient) IsAuthenticated() bool {
	return c.token != ""
}

// NewAuthContext 创建带 token 的 context
func (c *GRPCClient) NewAuthContext(ctx context.Context) context.Context {
	if c.token != "" {
		md := metadata.Pairs("authorization", "Bearer "+c.token)
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

func init() {
	// 确保导入路径正确
	_ = filepath.Join
}
