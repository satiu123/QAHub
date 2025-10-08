package clients

import (
	"context"
	"fmt"

	pb "qahub/api/proto/qa"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// QAServiceClient 是 QAService 的 gRPC 客户端封装
type QAServiceClient struct {
	conn   *grpc.ClientConn
	client pb.QAServiceClient
}

// NewQAServiceClient 创建一个新的 QA Service 客户端
func NewQAServiceClient(address string) (*QAServiceClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("连接到 QA Service 失败: %w", err)
	}

	client := pb.NewQAServiceClient(conn)
	return &QAServiceClient{
		conn:   conn,
		client: client,
	}, nil
}

// ListQuestions 获取所有问题列表
func (c *QAServiceClient) ListQuestions(ctx context.Context, page, pageSize int32) (*pb.ListQuestionsResponse, error) {
	return c.client.ListQuestions(ctx, &pb.ListQuestionsRequest{
		Page:     page,
		PageSize: pageSize,
	})
}

// Close 关闭客户端连接
func (c *QAServiceClient) Close() error {
	return c.conn.Close()
}
