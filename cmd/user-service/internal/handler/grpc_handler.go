package handler

import (
	"context"
	"log"
	"net"

	pb "qahub/api/proto/user"
	"qahub/pkg/auth"
	"qahub/pkg/config"
	"qahub/user-service/internal/model"
	"qahub/user-service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// UserGrpcServer 实现了 user_grpc.pb.go 中定义的 UserServiceServer 接口
type UserGrpcServer struct {
	pb.UnimplementedUserServiceServer // 必须嵌入，以实现向前兼容
	userService                       service.UserService
}

// NewUserGrpcServer 创建一个新的 gRPC 服务端处理器
func NewUserGrpcServer(svc service.UserService) *UserGrpcServer {
	return &UserGrpcServer{
		userService: svc,
	}
}

func (s *UserGrpcServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userResponse, err := s.userService.Register(ctx, req.Username, req.Email, req.Bio, req.Password)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		User: &pb.User{
			Id:       userResponse.ID,
			Username: userResponse.Username,
			Email:    userResponse.Email,
			Bio:      userResponse.Bio,
		},
	}, nil
}

func (s *UserGrpcServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := s.userService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{Token: token}, nil
}

func (s *UserGrpcServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	identity, err := s.userService.ValidateToken(ctx, req.JwtToken)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}
	return &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   identity.UserID,
		Username: identity.Username,
	}, nil
}

func (s *UserGrpcServer) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	userResponse, err := s.userService.GetUserProfile(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserProfileResponse{
		User: &pb.User{
			Id:       userResponse.ID,
			Username: userResponse.Username,
			Email:    userResponse.Email,
			Bio:      userResponse.Bio,
		},
	}, nil
}

func (s *UserGrpcServer) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*emptypb.Empty, error) {
	// 从 context 中获取认证用户ID
	identity, ok := auth.FromContext(ctx)
	if !ok || identity.UserID == 0 {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}

	// 权限校验：确保用户只能更新自己的信息
	if identity.UserID != req.UserId {
		return nil, status.Errorf(codes.PermissionDenied, "没有权限执行此操作")
	}

	updateModel := &model.User{
		ID:       req.UserId,
		Username: req.Username,
		Email:    req.Email,
		Bio:      req.Bio,
	}

	err := s.userService.UpdateUserProfile(ctx, updateModel)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *UserGrpcServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	// 从 context 中获取认证用户ID
	identity, ok := auth.FromContext(ctx)
	if !ok || identity.UserID == 0 {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}

	// 权限校验：确保用户只能删除自己的账户
	if identity.UserID != req.UserId {
		return nil, status.Errorf(codes.PermissionDenied, "没有权限执行此操作")
	}

	err := s.userService.DeleteUser(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *UserGrpcServer) Run(ctx context.Context, config config.UserService) error {
	serverAddr := ":" + config.GrpcPort
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalln("failed to listen:", err)
	}
	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, s)

	// 注册 reflection 服务，使 grpcurl 等工具可以动态发现服务
	reflection.Register(server)

	log.Printf("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gRPC server...")
	server.GracefulStop()
	log.Println("gRPC server stopped.")
	return nil
}
