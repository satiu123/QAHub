package handler

import (
	"context"
	"log"
	pb "qahub/api/proto/notification"
	"qahub/notification-service/internal/service"
	"qahub/pkg/pagination"

	"google.golang.org/grpc"
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

// RegisterServer 注册 gRPC 服务器
func (s *NotificationGrpcServer) RegisterServer(grpcServer *grpc.Server) {
	pb.RegisterNotificationServiceServer(grpcServer, s)
}
