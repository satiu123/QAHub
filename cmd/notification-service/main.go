package main

import (
	"context"
	"log"
	"qahub/notification-service/internal/handler"
	"qahub/notification-service/internal/service"
	"qahub/notification-service/internal/store"
	"qahub/pkg/config"
	"qahub/pkg/database"
)

func main() {
	// 1. 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2.连接数据库
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := database.NewMongoConection(ctx, config.Conf.MongoDB)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer client.Disconnect(ctx)

	// 3.初始化store, hub, service
	ntStore := store.NewMongoNotificationStore(client.Database(config.Conf.MongoDB.Database))
	hub := service.NewHub()
	go hub.Run()
	ntService := service.NewNotificationService(ntStore, hub, config.Conf.Kafka)
	defer ntService.Close()

	// 4.启动Kafka消费者
	go ntService.StartConsumer(ctx)

	//5. 初始化GrpcServer
	ntServer := handler.NewNotificationGrpcServer(ntService)

	// 启动 gRPC 服务器
	gctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := ntServer.Run(gctx, config.Conf); err != nil {
		log.Fatalf("failed to run gRPC server: %v", err)
	}
}
