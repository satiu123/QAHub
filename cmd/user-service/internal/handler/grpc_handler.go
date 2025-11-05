package handler

import (
	"context"
	"log/slog"

	pb "qahub/api/proto/user"
	"qahub/pkg/auth"
	"qahub/pkg/log"
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
	logger := log.FromContext(ctx)

	logger.Info("用户注册请求",
		slog.String("username", req.Username),
		slog.String("email", req.Email),
	)

	userResponse, err := s.userService.Register(ctx, dto.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Bio:      req.Bio,
		Password: req.Password,
	})
	if err != nil {
		logger.Error("用户注册失败",
			slog.String("username", req.Username),
			slog.String("email", req.Email),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("用户注册成功",
		slog.Int64("user_id", userResponse.ID),
		slog.String("username", userResponse.Username),
	)

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
	logger := log.FromContext(ctx)

	logger.Info("用户登录请求",
		slog.String("username", req.Username),
	)

	token, err := s.userService.Login(ctx, req.Username, req.Password)
	if err != nil {
		logger.Warn("用户登录失败",
			slog.String("username", req.Username),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("用户登录成功",
		slog.String("username", req.Username),
	)

	return &pb.LoginResponse{Token: token}, nil
}

func (s *UserGrpcServer) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx)

	// 从 context 中获取认证用户ID
	identity, ok := auth.FromContext(ctx)
	if !ok || identity.UserID == 0 {
		logger.Error("登出失败：无法从context获取用户信息")
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}

	logger.Info("用户登出请求",
		slog.Int64("user_id", identity.UserID),
		slog.String("username", identity.Username),
	)

	err := s.userService.Logout(ctx, identity.Token, identity.Claims)
	if err != nil {
		logger.Error("用户登出失败",
			slog.Int64("user_id", identity.UserID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("用户登出成功",
		slog.Int64("user_id", identity.UserID),
	)

	return &emptypb.Empty{}, nil
}

func (s *UserGrpcServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	logger := log.FromContext(ctx)

	logger.Debug("Token 验证请求")

	identity, err := s.userService.ValidateToken(ctx, req.JwtToken)
	if err != nil {
		logger.Warn("Token 验证失败",
			slog.String("error", err.Error()),
		)
		return &pb.ValidateTokenResponse{}, nil
	}

	logger.Debug("Token 验证成功",
		slog.Int64("user_id", identity.UserID),
		slog.String("username", identity.Username),
	)

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
	logger := log.FromContext(ctx)

	logger.Info("获取用户资料请求",
		slog.Int64("user_id", req.UserId),
	)

	userResponse, err := s.userService.GetUserProfile(ctx, req.UserId)
	if err != nil {
		logger.Error("获取用户资料失败",
			slog.Int64("user_id", req.UserId),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("获取用户资料成功",
		slog.Int64("user_id", req.UserId),
		slog.String("username", userResponse.Username),
	)

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
	// 从 context 中获取 logger
	logger := log.FromContext(ctx)

	// 记录请求开始，包含关键输入参数
	logger.Info("开始更新用户资料",
		slog.Int64("target_user_id", req.UserId),
		slog.String("new_username", req.Username),
		slog.String("new_email", req.Email),
	)

	// 从 context 中获取认证用户ID
	identity, ok := auth.FromContext(ctx)
	if !ok || identity.UserID == 0 {
		logger.Error("无法从context获取用户信息",
			slog.Int64("target_user_id", req.UserId),
		)
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}

	// 权限校验：确保用户只能更新自己的信息
	if identity.UserID != req.UserId {
		logger.Warn("权限校验失败：用户尝试更新他人资料",
			slog.Int64("authenticated_user_id", identity.UserID),
			slog.Int64("target_user_id", req.UserId),
		)
		return nil, status.Errorf(codes.PermissionDenied, "没有权限执行此操作")
	}

	updateModel := &model.User{
		ID:       req.UserId,
		Username: req.Username,
		Email:    req.Email,
		Bio:      req.Bio,
	}

	// 调用 service 层更新用户资料
	err := s.userService.UpdateUserProfile(ctx, updateModel)
	if err != nil {
		logger.Error("更新用户资料失败",
			slog.Int64("user_id", req.UserId),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// 记录成功日志
	logger.Info("用户资料更新成功",
		slog.Int64("user_id", req.UserId),
		slog.String("username", req.Username),
	)

	return &emptypb.Empty{}, nil
}

func (s *UserGrpcServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx)

	logger.Info("删除用户请求",
		slog.Int64("target_user_id", req.UserId),
	)

	// 从 context 中获取认证用户ID
	identity, ok := auth.FromContext(ctx)
	if !ok || identity.UserID == 0 {
		logger.Error("删除用户失败：无法从context获取用户信息",
			slog.Int64("target_user_id", req.UserId),
		)
		return nil, status.Errorf(codes.Internal, "无法从context获取用户信息")
	}

	// 权限校验：确保用户只能删除自己的账户
	if identity.UserID != req.UserId {
		logger.Warn("权限校验失败：用户尝试删除他人账户",
			slog.Int64("authenticated_user_id", identity.UserID),
			slog.Int64("target_user_id", req.UserId),
		)
		return nil, status.Errorf(codes.PermissionDenied, "没有权限执行此操作")
	}

	err := s.userService.DeleteUser(ctx, req.UserId)
	if err != nil {
		logger.Error("删除用户失败",
			slog.Int64("user_id", req.UserId),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("删除用户成功",
		slog.Int64("user_id", req.UserId),
	)

	return &emptypb.Empty{}, nil
}

// RegisterServer 将此 handler 注册到给定的 gRPC 服务器上
func (s *UserGrpcServer) RegisterServer(grpcServer *grpc.Server) {
	pb.RegisterUserServiceServer(grpcServer, s)
}
