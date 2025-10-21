package main

import (
	"context"
	"fmt"
	"log"

	"qahub/search-service/internal/handler"
	"qahub/search-service/internal/service"

	"qahub/pkg/clients"
	"qahub/pkg/config"
	"qahub/pkg/messaging"
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

	serviceName := "search.SearchService"
	// 初始化依赖并注入
	qaServiceAddr := config.Conf.Services.Gateway.QaServiceEndpoint
	esStore, err := store.NewEsStore(config.Conf.Elasticsearch, qaServiceAddr)
	if err != nil {
		log.Fatalf("初始化 Elasticsearch store 失败: %v", err)
	}
	defer util.Cleanup("Elasticsearch client", esStore.Close)
	consumer := messaging.NewKafkaConsumer(config.Conf.Kafka, service.TopicQuestions, service.GroupID, nil)
	searchService := service.NewSearchService(esStore)

	// 注册事件处理器
	consumer.SetHandlers(searchService.RegisterHandlers())

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
	grpcSrv := server.NewGrpcServer(serviceName, config.Conf.Services.SearchService.GrpcPort, serverOpts...)

	// 设置健康检查
	healthUpdater := grpcSrv.HealthServer()
	util.SetHealthChecks(
		healthUpdater,
		serviceName,
		consumer, esStore)

	// 启动后台任务
	go consumer.Start(context.Background())
	grpcSrv.Run(func(s *grpc.Server) {
		searchHandler.RegisterServer(s)
	})

	defer util.Cleanup("user-service gRPC client", userClient.Close)
}
