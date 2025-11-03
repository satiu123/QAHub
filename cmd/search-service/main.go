package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"qahub/search-service/internal/handler"
	"qahub/search-service/internal/service"

	"qahub/pkg/clients"
	"qahub/pkg/config"
	"qahub/pkg/health"
	"qahub/pkg/interceptor"
	logpkg "qahub/pkg/log"
	"qahub/pkg/messaging"
	"qahub/pkg/server"
	"qahub/pkg/util"
	"qahub/search-service/internal/store"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Starting Search Service...")

	// 初始化配置
	if err := config.Init("configs"); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化日志
	logpkg.InitLogger(&config.Conf.Log)
	logger := slog.Default()

	serviceName := "search.SearchService"
	logger.Info("搜索服务启动中...",
		slog.String("service", "search-service"),
		slog.String("grpc_port", config.Conf.Services.SearchService.GrpcPort),
	)

	// 初始化依赖并注入
	qaServiceAddr := config.Conf.Services.Gateway.QaServiceEndpoint
	logger.Info("初始化 Elasticsearch 存储...",
		slog.String("qa_service_addr", qaServiceAddr),
	)
	esStore, err := store.NewEsStore(config.Conf.Elasticsearch, qaServiceAddr)
	if err != nil {
		logger.Error("初始化 Elasticsearch store 失败",
			slog.String("error", err.Error()),
		)
		log.Fatalf("初始化 Elasticsearch store 失败: %v", err)
	}
	defer util.Cleanup("Elasticsearch client", esStore.Close)
	logger.Info("Elasticsearch 连接成功")

	logger.Info("初始化 Kafka 消费者...")
	consumer := messaging.NewKafkaConsumer(config.Conf.Kafka, service.TopicQuestions, service.GroupID, nil)
	searchService := service.NewSearchService(esStore)

	// 注册事件处理器
	consumer.SetHandlers(searchService.RegisterHandlers())
	logger.Info("Kafka 消费者初始化成功")

	searchHandler := handler.NewSearchServer(searchService)

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

	// 创建并运行 gRPC 服务器
	logger.Info("初始化 gRPC 服务器...",
		slog.String("service_name", serviceName),
	)
	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.LogUnaryServerInterceptor(),
			interceptor.AuthUnaryServerInterceptor(userClient, config.Conf.Services.SearchService.PublicMethods...),
		),
	}
	grpcSrv := server.NewGrpcServer(serviceName, config.Conf.Services.SearchService.GrpcPort, serverOpts...)

	// 设置健康检查
	healthUpdater := grpcSrv.HealthServer()
	health.SetHealthChecks(
		healthUpdater,
		serviceName,
		consumer, esStore)

	// 启动后台任务
	logger.Info("启动 Kafka 消费者后台任务...")
	go consumer.Start(context.Background())

	logger.Info("搜索服务准备就绪，开始监听请求",
		slog.String("grpc_port", config.Conf.Services.SearchService.GrpcPort),
	)

	grpcSrv.Run(func(s *grpc.Server) {
		searchHandler.RegisterServer(s)
	})

	defer util.Cleanup("user-service gRPC client", userClient.Close)
}
