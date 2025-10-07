package handler

import (
	"context"
	"log"
	"net"
	pb "qahub/api/proto/search"
	"qahub/pkg/clients"
	"qahub/pkg/config"
	"qahub/pkg/middleware"
	"qahub/search-service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type SearchGrpcServer struct {
	pb.UnimplementedSearchServiceServer
	service service.SearchService
}

func NewSearchServer(s service.SearchService) *SearchGrpcServer {
	return &SearchGrpcServer{service: s}
}

func (h *SearchGrpcServer) SearchQuestions(ctx context.Context, req *pb.SearchQuestionsRequest) (*pb.SearchQuestionsResponse, error) {
	results, err := h.service.SearchQuestions(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	// 将结果转换为 gRPC 响应格式
	resp := &pb.SearchQuestionsResponse{
		Questions: make([]*pb.Question, len(results)),
	}
	for i, q := range results {
		resp.Questions[i] = &pb.Question{
			Id:       q.ID,
			Title:    q.Title,
			Content:  q.Content,
			AuthorId: q.AuthorID,
		}
	}

	return resp, nil
}

func (s *SearchGrpcServer) Run(ctx context.Context, config config.Config) error {
	serverAddr := ":" + config.Services.SearchService.GrpcPort
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("无法监听 gRPC 端口: %v", err)
	}
	// 初始化 user-service 的客户端连接
	userClient, err := clients.NewUserServiceClient(config.Services.Gateway.UserServiceEndpoint)
	if err != nil {
		log.Fatalf("无法连接到 user-service: %v", err)
	}
	// 创建 gRPC 服务器实例，注册服务，并启动监听
	server := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GrpcAuthInterceptor(userClient, config.Services.QAService.PublicMethods...)),
	)
	pb.RegisterSearchServiceServer(server, s)

	// 注册 reflection 服务，使 grpcurl 等工具可以动态发现服务
	reflection.Register(server)

	log.Printf("gRPC 服务正在监听: %v", lis.Addr())
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("启动 gRPC 服务失败: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("正在关闭服务...")
	server.GracefulStop()
	log.Println("服务已关闭")
	return nil
}
