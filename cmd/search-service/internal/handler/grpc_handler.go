package handler

import (
	"context"
	"log/slog"

	pb "qahub/api/proto/search"
	pkglog "qahub/pkg/log"
	"qahub/search-service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SearchGrpcServer struct {
	pb.UnimplementedSearchServiceServer
	service service.SearchService
}

func NewSearchServer(s service.SearchService) *SearchGrpcServer {
	return &SearchGrpcServer{service: s}
}

func (h *SearchGrpcServer) SearchQuestions(ctx context.Context, req *pb.SearchQuestionsRequest) (*pb.SearchQuestionsResponse, error) {
	logger := pkglog.FromContext(ctx)

	logger.Info("搜索问题请求",
		slog.String("query", req.Query),
	)

	results, err := h.service.SearchQuestions(ctx, req.Query)
	if err != nil {
		logger.Error("搜索问题失败",
			slog.String("query", req.Query),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("搜索问题成功",
		slog.String("query", req.Query),
		slog.Int("result_count", len(results)),
	)

	// 将结果转换为 gRPC 响应格式
	resp := &pb.SearchQuestionsResponse{
		Questions: make([]*pb.Question, len(results)),
	}
	for i, q := range results {
		resp.Questions[i] = &pb.Question{
			Id:         q.ID,
			Title:      q.Title,
			Content:    q.Content,
			AuthorId:   q.AuthorID,
			AuthorName: q.AuthorName,
			CreatedAt:  timestamppb.New(q.CreatedAt),
			UpdatedAt:  timestamppb.New(q.UpdatedAt),
		}
	}

	return resp, nil
}

func (h *SearchGrpcServer) IndexAllQuestions(ctx context.Context, req *pb.IndexAllQuestionsRequest) (*pb.IndexAllQuestionsResponse, error) {
	logger := pkglog.FromContext(ctx)

	logger.Info("索引所有问题请求")

	err := h.service.IndexAllQuestions(ctx)
	if err != nil {
		logger.Error("索引所有问题失败",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("索引所有问题成功")

	return &pb.IndexAllQuestionsResponse{
		Message:      "成功索引所有问题",
		IndexedCount: 0, // TODO: 返回实际索引的问题数量
	}, nil
}

func (h *SearchGrpcServer) DeleteIndexAllQuestions(ctx context.Context, req *pb.DeleteIndexAllQuestionsRequest) (*pb.DeleteIndexAllQuestionsResponse, error) {
	logger := pkglog.FromContext(ctx)

	logger.Info("删除所有问题索引请求")

	err := h.service.DeleteIndexAllQuestions(ctx)
	if err != nil {
		logger.Error("删除所有问题索引失败",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("删除所有问题索引成功")

	return &pb.DeleteIndexAllQuestionsResponse{
		Message: "成功删除所有问题索引",
	}, nil
}

func (h *SearchGrpcServer) RegisterServer(grpcServer *grpc.Server) {
	pb.RegisterSearchServiceServer(grpcServer, h)
}
