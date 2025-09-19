package main

import (
	"fmt"
	"log"
	"net"

	pb "qahub/api/proto/user"
	"qahub/internal/user/handler"
	"qahub/internal/user/service"
	"qahub/internal/user/store"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/middleware"
	"qahub/pkg/redis"

	"github.com/gin-gonic/gin"

	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatal("配置加载失败:", err)
	}
	// 创建数据库连接
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()
	// 创建 Redis 客户端
	redisClient, err := redis.NewClient(config.Conf.Redis)
	if err != nil {
		log.Fatal("Redis 连接失败:", err)
	}
	defer redisClient.Close()

	userStore := store.NewMySQLUserStore(db)
	cacheStore := store.NewUserCacheStore(redisClient, userStore)
	userService := service.NewUserService(cacheStore)

	// --- 启动 gRPC 服务器 ---
	go func() {
		grpcPort := config.Conf.Services.UserService.GrpcPort
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
		if err != nil {
			log.Fatalf("无法监听 gRPC 端口 %s: %v", grpcPort, err)
		}

		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(middleware.GrpcAuthInterceptor()),
		)

		userGrpcServer := handler.NewUserGrpcServer(userService)
		pb.RegisterUserServiceServer(grpcServer, userGrpcServer)

		log.Printf("gRPC 用户服务正在监听端口: %s", grpcPort)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("启动 gRPC 服务失败: %v", err)
		}
	}()

	// --- 启动 Gin HTTP 服务器 ---
	gin.SetMode(config.Conf.Server.Mode)
	router := gin.Default()

	// 添加 CORS 中间件
	router.Use(middleware.CORSMiddleware())

	userHandler := handler.NewUserHandler(userService)

	// 定义路由
	api := router.Group("/api/v1")
	{
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)
	}
	// 受保护的路由
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", userHandler.GetProfile)
		protected.PUT("/profile", userHandler.UpdateProfile)
	}

	httpPort := config.Conf.Services.UserService.HttpPort
	log.Printf("HTTP 用户服务正在监听端口: %s", httpPort)
	if err := router.Run(fmt.Sprintf(":%s", httpPort)); err != nil {
		log.Fatalf("启动 HTTP 服务失败: %v", err)
	}
}
