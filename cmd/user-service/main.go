package main

import (
	"context"
	"log"
	"os/signal"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/redis"
	"qahub/user-service/internal/handler"
	"qahub/user-service/internal/service"
	"qahub/user-service/internal/store"
	"syscall"
)

func main() {
	// 1. 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 2. 初始化数据库和 Redis
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	redisClient, err := redis.NewClient(config.Conf.Redis)
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	defer redisClient.Close()
	// 3. 依赖注入
	userStoreWithBlacklist := store.NewUserCacheStore(redisClient, store.NewMySQLUserStore(db))
	userService := service.NewUserService(userStoreWithBlacklist)
	userGrpcHandler := handler.NewUserGrpcServer(userService)

	// 4. 启动 gRPC 服务器
	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := userGrpcHandler.Run(stopCtx, config.Conf.Services.UserService); err != nil {
		log.Fatalf("Failed to run gRPC server: %v", err)
	}

	<-stopCtx.Done()

	userGrpcHandler.Stop()
	log.Println("User service shut down gracefully")
}
