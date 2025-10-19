package handler

import (
	"context"
	"log"
	pb "qahub/api/proto/qa"
	"qahub/pkg/auth"
	"qahub/pkg/pagination"
	"qahub/qa-service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// QAGrpcServer 实现了 pb.QAServiceServer 接口，处理 gRPC 请求
type QAGrpcServer struct {
	pb.UnimplementedQAServiceServer
	qaService service.QAService
}

func NewQAGrpcServer(svc service.QAService) *QAGrpcServer {
	return &QAGrpcServer{
		qaService: svc,
	}
}

func (s *QAGrpcServer) CreateQuestion(ctx context.Context, req *pb.CreateQuestionRequest) (*pb.QuestionResponse, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	question, err := s.qaService.CreateQuestion(ctx, req.Title, req.Content, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &pb.QuestionResponse{
		Id:        question.ID,
		Title:     question.Title,
		Content:   question.Content,
		UserId:    question.UserID,
		CreatedAt: timestamppb.New(question.CreatedAt),
		UpdatedAt: timestamppb.New(question.UpdatedAt),
	}, nil
}

func (s *QAGrpcServer) GetQuestion(ctx context.Context, req *pb.GetQuestionRequest) (*pb.QuestionResponse, error) {
	question, err := s.qaService.GetQuestion(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.QuestionResponse{
		Id:          question.ID,
		Title:       question.Title,
		Content:     question.Content,
		UserId:      question.UserID,
		CreatedAt:   timestamppb.New(question.CreatedAt),
		UpdatedAt:   timestamppb.New(question.UpdatedAt),
		AuthorName:  question.AuthorName,
		AnswerCount: question.AnswerCount,
	}, nil
}

func (s *QAGrpcServer) ListQuestions(ctx context.Context, req *pb.ListQuestionsRequest) (*pb.ListQuestionsResponse, error) {
	page, pageSize := pagination.NormalizePageAndSize(req)
	questions, count, err := s.qaService.ListQuestions(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	log.Println("Fetched questions:", questions)
	var pbQuestions []*pb.QuestionResponse
	for _, q := range questions {
		pbQuestions = append(pbQuestions, &pb.QuestionResponse{
			Id:          q.ID,
			Title:       q.Title,
			Content:     q.Content,
			UserId:      q.UserID,
			CreatedAt:   timestamppb.New(q.CreatedAt),
			UpdatedAt:   timestamppb.New(q.UpdatedAt),
			AuthorName:  q.AuthorName,
			AnswerCount: q.AnswerCount,
		})
	}
	return &pb.ListQuestionsResponse{
		Questions:  pbQuestions,
		TotalCount: count,
	}, nil
}

func (s *QAGrpcServer) UpdateQuestion(ctx context.Context, req *pb.UpdateQuestionRequest) (*pb.QuestionResponse, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	question, err := s.qaService.UpdateQuestion(ctx, req.Id, req.Title, req.Content, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &pb.QuestionResponse{
		Id:        question.ID,
		Title:     question.Title,
		Content:   question.Content,
		UserId:    question.UserID,
		CreatedAt: timestamppb.New(question.CreatedAt),
		UpdatedAt: timestamppb.New(question.UpdatedAt),
	}, nil
}

func (s *QAGrpcServer) DeleteQuestion(ctx context.Context, req *pb.DeleteQuestionRequest) (*emptypb.Empty, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	err := s.qaService.DeleteQuestion(ctx, req.Id, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *QAGrpcServer) CreateAnswer(ctx context.Context, req *pb.CreateAnswerRequest) (*pb.AnswerResponse, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	answer, err := s.qaService.CreateAnswer(ctx, req.QuestionId, req.Content, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &pb.AnswerResponse{
		Id:          answer.ID,
		QuestionId:  answer.QuestionID,
		Content:     answer.Content,
		UserId:      answer.UserID,
		UpvoteCount: int32(answer.UpvoteCount),
		CreatedAt:   timestamppb.New(answer.CreatedAt),
		UpdatedAt:   timestamppb.New(answer.UpdatedAt),
	}, nil
}

func (s *QAGrpcServer) ListAnswers(ctx context.Context, req *pb.ListAnswersRequest) (*pb.ListAnswersResponse, error) {
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	page, pageSize := pagination.NormalizePageAndSize(req)
	answers, count, err := s.qaService.ListAnswers(ctx, req.QuestionId, page, pageSize, identity.UserID)
	if err != nil {
		return nil, err
	}
	var pbAnswers []*pb.AnswerResponse
	for _, a := range answers {
		pbAnswers = append(pbAnswers, &pb.AnswerResponse{
			Id:              a.ID,
			QuestionId:      a.QuestionID,
			Content:         a.Content,
			UserId:          a.UserID,
			UpvoteCount:     int32(a.UpvoteCount),
			CreatedAt:       timestamppb.New(a.CreatedAt),
			UpdatedAt:       timestamppb.New(a.UpdatedAt),
			Username:        a.Username,
			IsUpvotedByUser: a.IsUpvotedByUser,
		})
	}
	return &pb.ListAnswersResponse{
		Answers:    pbAnswers,
		TotalCount: count,
	}, nil
}

func (s *QAGrpcServer) UpdateAnswer(ctx context.Context, req *pb.UpdateAnswerRequest) (*pb.AnswerResponse, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	answer, err := s.qaService.UpdateAnswer(ctx, req.Id, req.Content, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &pb.AnswerResponse{
		Id:          answer.ID,
		QuestionId:  answer.QuestionID,
		Content:     answer.Content,
		UserId:      answer.UserID,
		UpvoteCount: int32(answer.UpvoteCount),
		CreatedAt:   timestamppb.New(answer.CreatedAt),
		UpdatedAt:   timestamppb.New(answer.UpdatedAt),
	}, nil
}

func (s *QAGrpcServer) DeleteAnswer(ctx context.Context, req *pb.DeleteAnswerRequest) (*emptypb.Empty, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	err := s.qaService.DeleteAnswer(ctx, req.Id, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *QAGrpcServer) UpvoteAnswer(ctx context.Context, req *pb.UpvoteAnswerRequest) (*emptypb.Empty, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	err := s.qaService.UpvoteAnswer(ctx, req.AnswerId, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *QAGrpcServer) DownvoteAnswer(ctx context.Context, req *pb.DownvoteAnswerRequest) (*emptypb.Empty, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	err := s.qaService.DownvoteAnswer(ctx, req.AnswerId, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *QAGrpcServer) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CommentResponse, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	comment, err := s.qaService.CreateComment(ctx, req.AnswerId, req.Content, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &pb.CommentResponse{
		Id:        comment.ID,
		AnswerId:  comment.AnswerID,
		Content:   comment.Content,
		UserId:    comment.UserID,
		CreatedAt: timestamppb.New(comment.CreatedAt),
		UpdatedAt: timestamppb.New(comment.UpdatedAt),
	}, nil
}

func (s *QAGrpcServer) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	_ = identity // 目前未使用身份信息，但保留以备将来使用
	page, pageSize := pagination.NormalizePageAndSize(req)
	comments, count, err := s.qaService.ListComments(ctx, req.AnswerId, page, pageSize)
	if err != nil {
		return nil, err
	}
	var pbComments []*pb.CommentResponse
	for _, c := range comments {
		pbComments = append(pbComments, &pb.CommentResponse{
			Id:        c.ID,
			AnswerId:  c.AnswerID,
			Content:   c.Content,
			UserId:    c.UserID,
			CreatedAt: timestamppb.New(c.CreatedAt),
			UpdatedAt: timestamppb.New(c.UpdatedAt),
			Username:  c.Username,
		})
	}
	return &pb.ListCommentsResponse{
		Comments:   pbComments,
		TotalCount: count,
	}, nil
}

func (s *QAGrpcServer) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.CommentResponse, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	comment, err := s.qaService.UpdateComment(ctx, req.Id, req.Content, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &pb.CommentResponse{
		Id:        comment.ID,
		AnswerId:  comment.AnswerID,
		Content:   comment.Content,
		UserId:    comment.UserID,
		CreatedAt: timestamppb.New(comment.CreatedAt),
		UpdatedAt: timestamppb.New(comment.UpdatedAt),
	}, nil
}

func (s *QAGrpcServer) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*emptypb.Empty, error) {
	//从context中获取用户信息
	identity, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}
	err := s.qaService.DeleteComment(ctx, req.Id, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *QAGrpcServer) RegisterServer(grpcServer *grpc.Server) {
	pb.RegisterQAServiceServer(grpcServer, s)
}
