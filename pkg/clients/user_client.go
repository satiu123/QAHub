package clients

import (
	"context"
	pb "qahub/api/proto/user"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// UserServiceClient 是 UserService 的 gRPC 客户端封装
type UserServiceClient struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

// NewUserGrpcServer 创建一个新的 gRPC 服务端处理器
func NewUserServiceClient(address string) (*UserServiceClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, err
	}
	client := pb.NewUserServiceClient(conn)
	return &UserServiceClient{
		conn:   conn,
		client: client,
	}, nil
}

// ValidateToken 验证 JWT token
func (c *UserServiceClient) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	// 设置请求超时时间
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		JwtToken: token,
	})
}

// GetUserProfile 获取用户信息
func (c *UserServiceClient) GetUserProfile(ctx context.Context, userID int64) (*pb.GetUserProfileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return c.client.GetUserProfile(ctx, &pb.GetUserProfileRequest{
		UserId: userID,
	})
}

// Close 关闭 gRPC 连接
func (c *UserServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
