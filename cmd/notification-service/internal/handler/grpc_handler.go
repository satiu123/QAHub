package handler

import (
	"context"
	"log/slog"

	pb "qahub/api/proto/notification"
	"qahub/notification-service/internal/service"
	pkglog "qahub/pkg/log"
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
	logger := pkglog.FromContext(ctx)

	logger.Info("获取通知请求",
		slog.Int64("user_id", req.UserId),
	)

	limit, offset := pagination.LimitOffsetFromRequest(req)
	notifications, err := s.notificationService.GetNotifications(ctx, req.UserId, limit, offset)
	if err != nil {
		logger.Error("获取通知失败",
			slog.Int64("user_id", req.UserId),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("获取通知成功",
		slog.Int64("user_id", req.UserId),
		slog.Int("count", len(notifications)),
	)

	var pbNotifications []*pb.Notification
	for _, n := range notifications {
		pbNotifications = append(pbNotifications, &pb.Notification{
			Id:          n.ID.Hex(),
			RecipientId: n.RecipientID,
			SenderId:    n.SenderID,
			SenderName:  n.SenderName,
			Type:        pb.NotificationType(n.Type),
			Content:     n.Content,
			Status:      n.Status.ToProto(),
			CreatedAt:   timestamppb.New(n.CreatedAt),
			TargetUrl:   n.TargetURL,
		})
	}

	return &pb.GetNotificationsResponse{
		Notifications: pbNotifications,
	}, nil
}

func (s *NotificationGrpcServer) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	logger := pkglog.FromContext(ctx)

	logger.Info("标记通知为已读请求",
		slog.Int64("user_id", req.GetUserId()),
		slog.Bool("mark_all", req.GetMarkAll()),
		slog.Int("notification_count", len(req.GetNotificationIds())),
	)

	ModifiedCount, err := s.notificationService.MarkNotificationsAsRead(ctx, req.GetUserId(), req.GetNotificationIds(), req.GetMarkAll())
	if err != nil {
		logger.Error("标记通知为已读失败",
			slog.Int64("user_id", req.GetUserId()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("标记通知为已读成功",
		slog.Int64("user_id", req.GetUserId()),
		slog.Int64("modified_count", ModifiedCount),
	)

	return &pb.MarkAsReadResponse{
		ModifiedCount: ModifiedCount,
	}, nil
}

func (s *NotificationGrpcServer) DeleteNotification(ctx context.Context, req *pb.DeleteNotificationRequest) (*pb.DeleteNotificationResponse, error) {
	logger := pkglog.FromContext(ctx)

	logger.Info("删除通知请求",
		slog.Int64("user_id", req.GetUserId()),
		slog.String("notification_id", req.GetNotificationId()),
	)

	err := s.notificationService.DeleteNotification(ctx, req.GetUserId(), req.GetNotificationId())
	if err != nil {
		logger.Error("删除通知失败",
			slog.Int64("user_id", req.GetUserId()),
			slog.String("notification_id", req.GetNotificationId()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("删除通知成功",
		slog.Int64("user_id", req.GetUserId()),
		slog.String("notification_id", req.GetNotificationId()),
	)

	return &pb.DeleteNotificationResponse{
		NotificationId: req.GetNotificationId(),
	}, nil
}

func (s *NotificationGrpcServer) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	logger := pkglog.FromContext(ctx)

	logger.Info("获取未读通知数请求",
		slog.Int64("user_id", req.GetUserId()),
	)

	count, err := s.notificationService.GetUnreadCount(ctx, req.GetUserId())
	if err != nil {
		logger.Error("获取未读通知数失败",
			slog.Int64("user_id", req.GetUserId()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	logger.Info("获取未读通知数成功",
		slog.Int64("user_id", req.GetUserId()),
		slog.Int64("unread_count", count),
	)

	return &pb.GetUnreadCountResponse{
		UnreadCount: count,
	}, nil
}

// SubscribeNotifications 实现服务端流式推送实时通知
func (s *NotificationGrpcServer) SubscribeNotifications(req *pb.SubscribeNotificationsRequest, stream pb.NotificationService_SubscribeNotificationsServer) error {
	logger := pkglog.FromContext(stream.Context())
	userID := req.GetUserId()

	logger.Info("用户订阅通知流",
		slog.Int64("user_id", userID),
	)

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
		logger.Info("通知流上下文完成",
			slog.Int64("user_id", userID),
			slog.String("error", stream.Context().Err().Error()),
		)
		return stream.Context().Err()
	case <-streamClient.Done:
		logger.Info("通知流客户端关闭",
			slog.Int64("user_id", userID),
		)
		return nil
	}
}

// RegisterServer 注册 gRPC 服务器
func (s *NotificationGrpcServer) RegisterServer(grpcServer *grpc.Server) {
	pb.RegisterNotificationServiceServer(grpcServer, s)
}
