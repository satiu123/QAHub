package main

import (
	"log"
	"os"
	"qahub/pkg/clients"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/middleware"
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
		os.Exit(1)
	}

	// 初始化数据库连接
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		os.Exit(1)
	}
	defer util.Cleanup("MySQL connection", db.Close)
	// 依赖注入：初始化 store, service, handler
	qaStore := store.NewQAStore(db)
	qaService := service.NewQAService(qaStore, config.Conf.Kafka)
	qaHandler := handler.NewQAGrpcServer(qaService)
	// 初始化 user-service 的客户端连接
	userClient, err := clients.NewUserServiceClient(config.Conf.Services.Gateway.UserServiceEndpoint)
	if err != nil {
		log.Fatalf("无法连接到 user-service: %v", err)
	}
	// 启动 gRPC 服务器
	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.GrpcAuthInterceptor(userClient, config.Conf.Services.QAService.PublicMethods...)),
	}
	grpcSrv := server.NewGrpcServer("qa.QAService", config.Conf.Services.QAService.GrpcPort, serverOpts...)
	grpcSrv.Run(func(s *grpc.Server) {
		qaHandler.RegisterServer(s)
	})

	defer util.Cleanup("user-service gRPC client", userClient.Close)
}
