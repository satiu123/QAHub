package main

import (
	"context"
	"log"
	"os/signal"
	"qahub/notification-service/internal/handler"
	"qahub/notification-service/internal/service"
	"qahub/notification-service/internal/store"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"syscall"
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

	// 3.初始化store, streamHub, service
	ntStore := store.NewMongoNotificationStore(client.Database(config.Conf.MongoDB.Database))
	streamHub := service.NewStreamHub()
	go streamHub.Run()
	ntService := service.NewNotificationService(ntStore, streamHub, config.Conf.Kafka)
	defer ntService.Close()

	// 4.启动Kafka消费者
	go ntService.StartConsumer(ctx)

	// 启动 gRPC 服务器
	grpcServer := handler.NewNotificationGrpcServer(ntService)

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := grpcServer.Run(stopCtx, config.Conf); err != nil {
		log.Fatalf("Failed to run gRPC server: %v", err)
	}

	<-stopCtx.Done()

	grpcServer.Stop()
	log.Println("Notification service shut down gracefully")
}
