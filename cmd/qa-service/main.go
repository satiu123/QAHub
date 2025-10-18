package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/util"
	"qahub/qa-service/internal/handler"
	"qahub/qa-service/internal/service"
	"qahub/qa-service/internal/store"
	"syscall"
)

func main() {
	// 1. 加载配置
	if err := config.Init("configs"); err != nil {
		os.Exit(1)
	}

	// 2. 初始化数据库连接
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		os.Exit(1)
	}
	defer util.Cleanup("MySQL connection", db.Close)
	// 3. 依赖注入：初始化 store, service, handler
	qaStore := store.NewQAStore(db)
	qaService := service.NewQAService(qaStore, config.Conf.Kafka)

	// 启动 gRPC 服务器
	grpcServer := handler.NewQAGrpcServer(qaService)

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := grpcServer.Run(stopCtx, config.Conf); err != nil {
		log.Fatalf("Failed to run gRPC server: %v", err)
	}

	<-stopCtx.Done()

	grpcServer.Stop()
	log.Println("QA service shut down gracefully")
}
