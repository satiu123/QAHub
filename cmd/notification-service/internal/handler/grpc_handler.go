package handler

import (
	"context"
	"log"
	"net"
	pb "qahub/api/proto/notification"
	"qahub/notification-service/internal/service"
	"qahub/pkg/clients"
	"qahub/pkg/config"
	"qahub/pkg/middleware"
	"qahub/pkg/pagination"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NotificationGrpcServer struct {
	pb.UnimplementedNotificationServiceServer
	notificationService service.NotificationService
}

func NewNotificationGrpcServer(notificationService service.NotificationService) *NotificationGrpcServer {
	return &NotificationGrpcServer{
		notificationService: notificationService,
	}
}

func (s *NotificationGrpcServer) GetNotifications(ctx context.Context, req *pb.GetNotificationsRequest) (*pb.GetNotificationsResponse, error) {
	limit, offset := pagination.LimitOffsetFromRequest(req)
	notifications, err := s.notificationService.GetNotifications(ctx, req.UserId, limit, offset)
	if err != nil {
		return nil, err
	}

	var pbNotifications []*pb.Notification
	for _, n := range notifications {
		pbNotifications = append(pbNotifications, &pb.Notification{
			Id:          n.ID.Hex(),
			RecipientId: n.RecipientID,
			SenderId:    n.SenderID,
			SenderName:  n.SenderName,
			Type:        n.Type,
			Content:     n.Content,
			IsRead:      n.IsRead,
			CreatedAt:   timestamppb.New(n.CreatedAt),
			TargetUrl:   n.TargetURL,
		})
	}

	return &pb.GetNotificationsResponse{
		Notifications: pbNotifications,
	}, nil
}

func (s *NotificationGrpcServer) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	ModifiedCount, err := s.notificationService.MarkNotificationsAsRead(ctx, req.GetUserId(), req.GetNotificationIds(), req.GetMarkAll())
	if err != nil {
		return nil, err
	}
	return &pb.MarkAsReadResponse{
		ModifiedCount: ModifiedCount,
	}, nil
}

func (s *NotificationGrpcServer) DeleteNotification(ctx context.Context, req *pb.DeleteNotificationRequest) (*pb.DeleteNotificationResponse, error) {
	err := s.notificationService.DeleteNotification(ctx, req.GetUserId(), req.GetNotificationId())
	if err != nil {
		return nil, err
	}
	return &pb.DeleteNotificationResponse{
		NotificationId: req.GetNotificationId(),
	}, nil
}

func (s *NotificationGrpcServer) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	count, err := s.notificationService.GetUnreadCount(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &pb.GetUnreadCountResponse{
		UnreadCount: count,
	}, nil
}

// SubscribeNotifications 实现服务端流式推送实时通知
func (s *NotificationGrpcServer) SubscribeNotifications(req *pb.SubscribeNotificationsRequest, stream pb.NotificationService_SubscribeNotificationsServer) error {
	userID := req.GetUserId()
	log.Printf("User %d subscribing to notifications stream", userID)

	// 创建流客户端
	streamClient := &service.StreamClient{
		UserID: userID,
		Stream: stream,
		Done:   make(chan struct{}),
	}

	// 注册到 StreamHub
	streamHub := s.notificationService.GetStreamHub()
	streamHub.Register(streamClient)
	defer streamHub.Unregister(streamClient)

	// 保持连接直到客户端断开或上下文取消
	select {
	case <-stream.Context().Done():
		log.Printf("User %d stream context done: %v", userID, stream.Context().Err())
		return stream.Context().Err()
	case <-streamClient.Done:
		log.Printf("User %d stream client closed", userID)
		return nil
	}
}

func (s *NotificationGrpcServer) Run(ctx context.Context, config config.Config) error {
	serverAddr := ":" + config.Services.NotificationService.GrpcPort
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
	pb.RegisterNotificationServiceServer(server, s)

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
