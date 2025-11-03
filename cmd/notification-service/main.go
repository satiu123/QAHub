package main

import (
	"context"
	"log"
	"log/slog"
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
	logger := slog.Default()

	serviceName := "notification.NotificationService"
	logger.Info("通知服务启动中...",
		slog.String("service", "notification-service"),
		slog.String("grpc_port", config.Conf.Services.NotificationService.GrpcPort),
	)

	// 2.连接数据库
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("初始化 MongoDB 连接...")
	client, err := database.NewMongoConection(ctx, config.Conf.MongoDB)
	if err != nil {
		logger.Error("MongoDB 连接失败",
			slog.String("error", err.Error()),
		)
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer util.Cleanup("MongoDB client", func() error { return client.Disconnect(ctx) })
	logger.Info("MongoDB 连接成功")

	// 3.初始化store, streamHub, service, handler
	ntStore := store.NewMongoNotificationStore(client.Database(config.Conf.MongoDB.Database))
	streamHub := service.NewStreamHub()

	logger.Info("初始化 Kafka 消费者...")
	consumer := messaging.NewKafkaConsumer(config.Conf.Kafka, service.TopicNotifications, service.GroupID, nil)
	ntService := service.NewNotificationService(ntStore, streamHub)
	ntHandler := handler.NewNotificationGrpcServer(ntService)

	// 注册事件处理器
	consumer.SetHandlers(ntService.RegisterHandlers())
	logger.Info("Kafka 消费者初始化成功")

	// 初始化 user-service 的客户端连接
	logger.Info("连接到 user-service...",
		slog.String("endpoint", config.Conf.Services.Gateway.UserServiceEndpoint),
	)
	userClient, err := clients.NewUserServiceClient(config.Conf.Services.Gateway.UserServiceEndpoint)
	if err != nil {
		logger.Error("连接 user-service 失败",
			slog.String("error", err.Error()),
		)
		log.Fatalf("无法连接到 user-service: %v", err)
	}
	logger.Info("user-service 连接成功")

	// 启动 gRPC 服务器
	logger.Info("初始化 gRPC 服务器...",
		slog.String("service_name", serviceName),
	)
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
	logger.Info("启动后台任务：StreamHub 和 Kafka 消费者...")
	go streamHub.Run()
	go consumer.Start(context.Background())

	logger.Info("通知服务准备就绪，开始监听请求",
		slog.String("grpc_port", config.Conf.Services.NotificationService.GrpcPort),
	)

	// 启动 gRPC 服务
	grpcSrv.Run(func(s *grpc.Server) {
		ntHandler.RegisterServer(s)
	})

	defer util.Cleanup("user-service gRPC client", userClient.Close)
}
