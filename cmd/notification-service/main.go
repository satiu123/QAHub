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
	"qahub/pkg/health"
	"qahub/pkg/interceptor"
	logpkg "qahub/pkg/log"
	"qahub/pkg/messaging"
	"qahub/pkg/server"
	"qahub/pkg/util"

	"google.golang.org/grpc"
)

func main() {
	// 1. 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logpkg.InitLogger(&config.Conf.Log)

	serviceName := "notification.NotificationService"
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

	consumer := messaging.NewKafkaConsumer(config.Conf.Kafka, service.TopicNotifications, service.GroupID, nil)
	ntService := service.NewNotificationService(ntStore, streamHub)
	ntHandler := handler.NewNotificationGrpcServer(ntService)

	// 注册事件处理器
	consumer.SetHandlers(ntService.RegisterHandlers())

	// 初始化 user-service 的客户端连接
	userClient, err := clients.NewUserServiceClient(config.Conf.Services.Gateway.UserServiceEndpoint)
	if err != nil {
		log.Fatalf("无法连接到 user-service: %v", err)
	}
	// 启动 gRPC 服务器
	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.LogUnaryServerInterceptor(),
			interceptor.AuthUnaryServerInterceptor(userClient, config.Conf.Services.NotificationService.PublicMethods...),
		),
		grpc.ChainStreamInterceptor(
			interceptor.LogStreamServerInterceptor(),
			interceptor.AuthStreamServerInterceptor(userClient, config.Conf.Services.NotificationService.PublicMethods...),
		),
	}
	grpcSrv := server.NewGrpcServer(serviceName, config.Conf.Services.NotificationService.GrpcPort, serverOpts...)

	// 设置健康检查
	healthUpdater := grpcSrv.HealthServer()
	health.SetHealthChecks(
		healthUpdater,
		serviceName,
		consumer, ntStore)

	// 启动后台任务
	go streamHub.Run()
	go consumer.Start(context.Background())

	// 启动 gRPC 服务
	grpcSrv.Run(func(s *grpc.Server) {
		ntHandler.RegisterServer(s)
	})

	defer util.Cleanup("user-service gRPC client", userClient.Close)
}
