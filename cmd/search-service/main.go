package main

import (
	"context"
	"fmt"
	"log"

	"qahub/search-service/internal/handler"
	"qahub/search-service/internal/service"

	"qahub/pkg/clients"
	"qahub/pkg/config"
	"qahub/pkg/middleware"
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

	// 初始化依赖并注入
	qaServiceAddr := config.Conf.Services.Gateway.QaServiceEndpoint
	esStore, err := store.New(config.Conf.Elasticsearch, qaServiceAddr)
	if err != nil {
		log.Fatalf("初始化 Elasticsearch store 失败: %v", err)
	}
	defer util.Cleanup("Elasticsearch client", esStore.Close)

	searchService := service.New(esStore, config.Conf.Kafka)
	go searchService.StartConsumer(context.Background())

	searchHandler := handler.NewSearchServer(searchService)

	// 初始化 user-service 的客户端连接
	userClient, err := clients.NewUserServiceClient(config.Conf.Services.Gateway.UserServiceEndpoint)
	if err != nil {
		log.Fatalf("无法连接到 user-service: %v", err)
	}

	// 创建并运行 gRPC 服务器
	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.GrpcAuthInterceptor(userClient, config.Conf.Services.SearchService.PublicMethods...)),
	}
	grpcSrv := server.NewGrpcServer("search.SearchService", config.Conf.Services.SearchService.GrpcPort, serverOpts...)
	grpcSrv.Run(func(s *grpc.Server) {
		searchHandler.RegisterServer(s)
	})

	defer util.Cleanup("user-service gRPC client", userClient.Close)
}
