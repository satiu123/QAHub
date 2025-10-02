package main

import (
	"context"
	"log"
	"os"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/qa-service/internal/handler"
	"qahub/qa-service/internal/service"
	"qahub/qa-service/internal/store"
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
	defer db.Close()
	// 3. 依赖注入：初始化 store, service, handler
	qaStore := store.NewQAStore(db)
	qaService := service.NewQAService(qaStore, config.Conf.Kafka)
	qaGrpcHandler := handler.NewQAGrpcServer(qaService)

	// 启动 gRPC 服务器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := qaGrpcHandler.Run(ctx, config.Conf); err != nil {
		log.Fatalf("failed to run gRPC server: %v", err)
	}
}
