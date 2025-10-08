package main

import (
	"context"
	"fmt"
	"log"

	"qahub/search-service/internal/handler"
	"qahub/search-service/internal/service"

	"qahub/pkg/config"
	"qahub/search-service/internal/store"
)

func main() {
	fmt.Println("Starting Search Service...")

	// 1. 初始化配置
	// Dockerfile 中将配置文件复制到了 /app/configs/ 目录下
	if err := config.Init("configs"); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 2. 初始化 Elasticsearch Store，传入 QA Service 地址
	qaServiceAddr := config.Conf.Services.Gateway.QaServiceEndpoint
	esStore, err := store.New(config.Conf.Elasticsearch, qaServiceAddr)
	if err != nil {
		log.Fatalf("初始化 Elasticsearch store 失败: %v", err)
	}
	defer esStore.Close()

	// 3. 初始化 Service，并启动 Kafka 消费者
	searchService := service.New(esStore, config.Conf.Kafka)
	go searchService.StartConsumer(context.Background())

	// 4.初始化 Grpc
	grpcServer := handler.NewSearchServer(searchService)

	if err := grpcServer.Run(context.Background(), config.Conf); err != nil {
		log.Fatalf("启动 gRPC 服务器失败: %v", err)
	}
}
