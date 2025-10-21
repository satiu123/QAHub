package handler

import (
	"context"

	pb "qahub/api/proto/user"
	"qahub/pkg/auth"
	"qahub/user-service/internal/dto"
	"qahub/user-service/internal/model"
	"qahub/user-service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	userResponse, err := s.userService.Register(ctx, dto.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Bio:      req.Bio,
		Password: req.Password,
	})
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

func (s *UserGrpcServer) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	// 从 context 中获取认证用户ID
	identity, ok := auth.FromContext(ctx)
	if !ok || identity.UserID == 0 {
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}

	err := s.userService.Logout(ctx, identity.Token, identity.Claims)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *UserGrpcServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	identity, err := s.userService.ValidateToken(ctx, req.JwtToken)
	if err != nil {
		return &pb.ValidateTokenResponse{}, nil
	}

	// 将 jwt.MapClaims 转换为 map[string]string
	claimsMap := make(map[string]*structpb.Value)
	if identity.Claims != nil {
		for key, value := range identity.Claims {
			claimsMap[key], _ = structpb.NewValue(value)
		}
	}

	return &pb.ValidateTokenResponse{
		UserId:   identity.UserID,
		Username: identity.Username,
		Claims:   claimsMap,
	}, nil
}

func (s *UserGrpcServer) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	userResponse, err := s.userService.GetUserProfile(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserProfileResponse{
		User: &pb.User{
			Id:        userResponse.ID,
			Username:  userResponse.Username,
			Email:     userResponse.Email,
			Bio:       userResponse.Bio,
			CreatedAt: timestamppb.New(userResponse.CreatedAt),
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

// RegisterServer 将此 handler 注册到给定的 gRPC 服务器上
func (s *UserGrpcServer) RegisterServer(grpcServer *grpc.Server) {
	pb.RegisterUserServiceServer(grpcServer, s)
}
