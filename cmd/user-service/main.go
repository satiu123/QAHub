// qahub/user-service/main.go
package main

import (
	"log"
	"log/slog"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/health"
	"qahub/pkg/interceptor"
	logpkg "qahub/pkg/log"
	"qahub/pkg/redis"
	"qahub/pkg/server"
	"qahub/pkg/util"
	"qahub/user-service/internal/handler"
	"qahub/user-service/internal/service"
	"qahub/user-service/internal/store"

	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	cfg := config.Conf.Services.UserService

	// 初始化日志
	logpkg.InitLogger(&config.Conf.Log)
	logger := slog.Default()

	logger.Info("用户服务启动中...",
		slog.String("service", "user-service"),
		slog.String("grpc_port", cfg.GrpcPort),
	)

	// 初始化业务依赖 (DB, Redis, Store, Service, Handler)
	serviceName := "user.UserService"

	logger.Info("初始化数据库连接...")
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		logger.Error("数据库连接失败",
			slog.String("error", err.Error()),
		)
		log.Fatalf("Database connection failed: %v", err)
	}
	defer util.Cleanup("MySQL connection", db.Close)
	logger.Info("数据库连接成功")

	logger.Info("初始化 Redis 连接...")
	redisClient, err := redis.NewClient(config.Conf.Redis)
	if err != nil {
		logger.Error("Redis 连接失败",
			slog.String("error", err.Error()),
		)
		log.Fatalf("Redis connection failed: %v", err)
	}
	defer util.Cleanup("Redis client", redisClient.Close)
	logger.Info("Redis 连接成功")

	userStore := store.NewUserCacheStore(redisClient, store.NewMySQLUserStore(db))
	userService := service.NewUserService(userStore)
	userHandler := handler.NewUserGrpcServer(userService)

	logger.Info("初始化 gRPC 服务器...",
		slog.String("service_name", serviceName),
	)

	// 创建服务器
	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.LogUnaryServerInterceptor(),
			userService.AuthUnaryServerInterceptor(config.Conf.Services.UserService.PublicMethods...),
		),
	}

	// 创建通用服务器实例
	grpcSrv := server.NewGrpcServer(serviceName, cfg.GrpcPort, serverOpts...)

	// 设置健康检查
	healthUpdater := grpcSrv.HealthServer()
	health.SetHealthChecks(
		healthUpdater,
		serviceName,
		userStore)

	logger.Info("用户服务准备就绪，开始监听请求",
		slog.String("grpc_port", cfg.GrpcPort),
	)

	// 运行服务器，并传入业务注册的逻辑
	grpcSrv.Run(func(s *grpc.Server) {
		userHandler.RegisterServer(s)
	})
}
