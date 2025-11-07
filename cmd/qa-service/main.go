package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"qahub/pkg/clients"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/health"
	"qahub/pkg/interceptor"
	logpkg "qahub/pkg/log"
	"qahub/pkg/messaging"
	"qahub/pkg/server"
	"qahub/pkg/util"
	"qahub/qa-service/internal/handler"
	"qahub/qa-service/internal/service"
	"qahub/qa-service/internal/store"

	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatalf("配置加载失败: %v", err)
		os.Exit(1)
	}

	// 初始化日志
	logpkg.InitLogger(&config.Conf.Log)
	logger := slog.Default()

	serviceName := "qa.QAService"
	logger.Info("问答服务启动中...",
		slog.String("service", "qa-service"),
		slog.String("grpc_port", config.Conf.Services.QAService.GrpcPort),
	)

	// 初始化数据库连接
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("初始化 MySQL 连接...")
	db, err := database.NewMySQLConnection(ctx, config.Conf.MySQL)
	if err != nil {
		logger.Error("MySQL 连接失败",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer util.Cleanup("MySQL connection", db.Close)
	logger.Info("MySQL 连接成功")

	// 初始化 Kafka 生产者
	logger.Info("初始化 Kafka 生产者...")
	kafkaProducer := messaging.NewKafkaProducer(config.Conf.Kafka)
	defer util.Cleanup("Kafka producer", kafkaProducer.Close)
	logger.Info("Kafka 生产者初始化成功")

	// 依赖注入：初始化 store, service, handler
	qaStore := store.NewQAStore(db)
	qaService := service.NewQAService(qaStore, kafkaProducer, &config.Conf)
	qaHandler := handler.NewQAGrpcServer(qaService)

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
			interceptor.AuthUnaryServerInterceptor(userClient, config.Conf.Services.QAService.PublicMethods...),
		),
	}
	grpcSrv := server.NewGrpcServer(serviceName, config.Conf.Services.QAService.GrpcPort, serverOpts...)

	// 设置健康检查
	healthUpdater := grpcSrv.HealthServer()
	health.SetHealthChecks(healthUpdater, serviceName,
		kafkaProducer, qaStore)

	logger.Info("问答服务准备就绪，开始监听请求",
		slog.String("grpc_port", config.Conf.Services.QAService.GrpcPort),
	)

	grpcSrv.Run(func(s *grpc.Server) {
		qaHandler.RegisterServer(s)
	})

	defer util.Cleanup("user-service gRPC client", userClient.Close)
}
