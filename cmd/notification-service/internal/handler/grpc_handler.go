package handler

import (
	pb "qahub/api/proto/notification"
	"qahub/notification-service/internal/service"
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
