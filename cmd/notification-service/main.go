package main

import (
	"context"
	"log"
	"qahub/notification-service/internal/handler"
	"qahub/notification-service/internal/service"
	"qahub/notification-service/internal/store"
	"qahub/pkg/clients"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/messaging"
	"qahub/pkg/middleware"
	"qahub/pkg/server"
	"qahub/pkg/util"

	"google.golang.org/grpc"
)

func main() {
	// 1. 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2.连接数据库
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := database.NewMongoConection(ctx, config.Conf.MongoDB)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer util.Cleanup("MongoDB client", func() error { return client.Disconnect(ctx) })

	// 3.初始化store, streamHub, service, handler
	ntStore := store.NewMongoNotificationStore(client.Database(config.Conf.MongoDB.Database))
	streamHub := service.NewStreamHub()
	go streamHub.Run()

	consumer := messaging.NewKafkaConsumer(config.Conf.Kafka, service.TopicNotifications, service.GroupID, nil)
	ntService := service.NewNotificationService(ntStore, streamHub, consumer)
	defer util.Cleanup("Notification service", ntService.Close)
	ntHandler := handler.NewNotificationGrpcServer(ntService)
	// 启动Kafka消费者
	go ntService.StartConsumer(ctx)

	// 初始化 user-service 的客户端连接
	userClient, err := clients.NewUserServiceClient(config.Conf.Services.Gateway.UserServiceEndpoint)
	if err != nil {
		log.Fatalf("无法连接到 user-service: %v", err)
	}
	// 启动 gRPC 服务器
	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.GrpcAuthInterceptor(userClient, config.Conf.Services.NotificationService.PublicMethods...)),
	}
	grpcSrv := server.NewGrpcServer("notification.NotificationService", config.Conf.Services.NotificationService.GrpcPort, serverOpts...)

	// 设置健康检查
	healthUpdater := grpcSrv.HealthServer()
	util.SetHealthChecks(healthUpdater, "notification.NotificationService", consumer, ntStore)
	grpcSrv.Run(func(s *grpc.Server) {
		ntHandler.RegisterServer(s)
	})

	defer util.Cleanup("user-service gRPC client", userClient.Close)
}
