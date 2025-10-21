// qahub/user-service/main.go
package main

import (
	"log"
	"qahub/pkg/config"
	"qahub/pkg/database"
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

	// 初始化业务依赖 (DB, Redis, Store, Service, Handler)
	serviceName := "user.UserService"

	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer util.Cleanup("MySQL connection", db.Close)

	redisClient, err := redis.NewClient(config.Conf.Redis)
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	defer util.Cleanup("Redis client", redisClient.Close)

	userStore := store.NewUserCacheStore(redisClient, store.NewMySQLUserStore(db))
	userService := service.NewUserService(userStore)
	userHandler := handler.NewUserGrpcServer(userService)

	// 创建服务器
	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(userService.AuthInterceptor(cfg.PublicMethods...)),
	}

	// 创建通用服务器实例
	grpcSrv := server.NewGrpcServer(serviceName, cfg.GrpcPort, serverOpts...)

	// 设置健康检查
	healthUpdater := grpcSrv.HealthServer()
	util.SetHealthChecks(
		healthUpdater,
		serviceName,
		userStore)
	// 运行服务器，并传入业务注册的逻辑
	grpcSrv.Run(func(s *grpc.Server) {
		userHandler.RegisterServer(s)
	})
}
