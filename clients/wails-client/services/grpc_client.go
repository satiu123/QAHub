package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	ntpb "qahub/api/proto/notification"
	qapb "qahub/api/proto/qa"
	searchpb "qahub/api/proto/search"
	userpb "qahub/api/proto/user"
)

// GRPCClient 封装 gRPC 服务客户端
type GRPCClient struct {
	userConn         *grpc.ClientConn
	qaConn           *grpc.ClientConn
	searchConn       *grpc.ClientConn
	notificationConn *grpc.ClientConn

	UserClient         userpb.UserServiceClient
	QAClient           qapb.QAServiceClient
	SearchClient       searchpb.SearchServiceClient
	NotificationClient ntpb.NotificationServiceClient

	// 存储当前用户的 token 和信息
	token    string
	userID   int64
	username string
}

// NewGRPCClient 创建新的 gRPC 客户端连接
func NewGRPCClient(userAddr, qaAddr, searchAddr, notificationAddr string) (*GRPCClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	client := &GRPCClient{}
	// 为健康检查设置一个超时
	checkTimeout := 5 * time.Second

	// 使用一个 map 来定义所有需要连接的目标，代码更具扩展性
	targets := []struct {
		name string
		addr string
		conn **grpc.ClientConn // 使用指向连接字段的指针，以便在循环中赋值
	}{
		{"User", userAddr, &client.userConn},
		{"QA", qaAddr, &client.qaConn},
		{"Search", searchAddr, &client.searchConn},
		{"Notification", notificationAddr, &client.notificationConn},
	}

	// 循环创建连接
	for _, target := range targets {
		log.Printf("Connecting to %s Service at %s...", target.name, target.addr)
		conn, err := grpc.NewClient(target.addr, opts...)
		if err != nil {
			return nil, fmt.Errorf("gRPC dial setup failed for %s service: %w", target.name, err)
		}
		*target.conn = conn

		// 检查服务健康状况
		// TODO:后端服务需要实现健康检查接口，当前都未实现，会导致连接失败
		ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
		defer cancel()
		if err := checkServiceHealth(ctx, conn, target.name); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("health check failed for %s service: %w", target.name, err)
		}
		log.Printf("Connected to %s Service successfully.", target.name)
	}

	// 初始化所有 gRPC 客户端
	client.UserClient = userpb.NewUserServiceClient(client.userConn)
	client.QAClient = qapb.NewQAServiceClient(client.qaConn)
	client.SearchClient = searchpb.NewSearchServiceClient(client.searchConn)
	client.NotificationClient = ntpb.NewNotificationServiceClient(client.notificationConn)

	log.Println("✅ All gRPC clients connected successfully.")
	return client, nil
}

func checkServiceHealth(ctx context.Context, conn *grpc.ClientConn, serviceName string) error {
	//创建一个健康检查客户端
	healthClient := grpc_health_v1.NewHealthClient(conn)

	//发送健康检查请求
	req := &grpc_health_v1.HealthCheckRequest{
		Service: "", // 空字符串表示检查整个服务的健康状况
	}
	resp, err := healthClient.Check(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to check health of %s service: %w", serviceName, err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("%s service is not healthy: %s", serviceName, resp.Status.String())
	}

	log.Printf("%s service is healthy.", serviceName)
	return nil
}

// Close 关闭所有连接
func (c *GRPCClient) Close() error {
	var errs []string

	// 定义一个辅助结构，方便循环处理
	conns := []struct {
		name string
		conn *grpc.ClientConn
	}{
		{"User", c.userConn},
		{"QA", c.qaConn},
		{"Search", c.searchConn},
		{"Notification", c.notificationConn},
	}

	for _, item := range conns {
		if item.conn != nil {
			// 收集错误信息，以便向上层返回一个总的错误状态
			err := item.conn.Close()
			if err != nil {
				msg := fmt.Sprintf("failed to close %s gRPC client: %v", item.name, err)
				log.Printf("ERROR: %s", msg)
				errs = append(errs, msg)
			}
		}
	}

	// 如果在关闭过程中收集到了任何错误，就返回一个聚合的错误
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
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
