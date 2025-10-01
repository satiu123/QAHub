package main

import (
	"context"
	"log"
	"qahub/notification-service/internal/handler"
	"qahub/notification-service/internal/service"
	"qahub/notification-service/internal/store"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	cfg := config.Conf.Services.NotificationService

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

	// 5. 初始化 gin handler 和 router
	ntHandler := handler.NewHandler(ntService)
	router := gin.Default()

	// 6. 设置路由
	apiGroup := router.Group("/api/v1")
	apiGroup.Use(gin.Logger(), middleware.NginxAuthMiddleware(), middleware.CORSMiddleware()) // 使用 Nginx 传递的用户信息进行认证

	// WebSocket 路由
	apiGroup.GET("/ws", ntHandler.WsHandler)

	// 通知管理路由
	notificationGroup := apiGroup.Group("/notifications")
	{
		notificationGroup.GET("", ntHandler.GetNotifications)
		notificationGroup.POST("/read", ntHandler.MarkAsRead) // 也可以用PUT
		notificationGroup.DELETE("/:id", ntHandler.DeleteNotification)
	}

	// 7. 启动服务器
	serveAddr := ":" + cfg.HttpPort
	log.Printf("通知服务正在端口 %s 上运行", serveAddr)
	if err := router.Run(serveAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
