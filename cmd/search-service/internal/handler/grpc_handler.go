package handler

import (
	"context"
	pb "qahub/api/proto/search"
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
	err := h.service.IndexAllQuestions(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.IndexAllQuestionsResponse{
		Message:      "成功索引所有问题",
		IndexedCount: 0, // TODO: 返回实际索引的问题数量
	}, nil
}

func (h *SearchGrpcServer) DeleteIndexAllQuestions(ctx context.Context, req *pb.DeleteIndexAllQuestionsRequest) (*pb.DeleteIndexAllQuestionsResponse, error) {
	err := h.service.DeleteIndexAllQuestions(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteIndexAllQuestionsResponse{
		Message: "成功删除所有问题索引",
	}, nil
}

func (h *SearchGrpcServer) RegisterServer(grpcServer *grpc.Server) {
	pb.RegisterSearchServiceServer(grpcServer, h)
}
