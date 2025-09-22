package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"qahub/internal/search/handler"
	"qahub/internal/search/service"
	"qahub/internal/search/store"
	"qahub/pkg/config"
)

func main() {
	fmt.Println("Starting Search Service...")

	// 1. 初始化配置
	// Dockerfile 中将配置文件复制到了 /app/configs/ 目录下
	if err := config.Init("configs"); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 2. 初始化 Elasticsearch Store
	esStore, err := store.New(config.Conf.Elasticsearch)
	if err != nil {
		log.Fatalf("初始化 Elasticsearch store 失败: %v", err)
	}

	// 3. 初始化 Service，并启动 Kafka 消费者
	searchService := service.New(esStore, config.Conf.Kafka)
	go searchService.StartConsumer(context.Background())

	// 4. 初始化 Handler
	searchHandler := handler.New(searchService)

	r := gin.Default()

	// 从配置中读取服务地址和端口
	port := config.Conf.Services.SearchService.HttpPort
	if port == "" {
		port = "8083" // 如果配置为空，提供一个默认值
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	// 注册搜索 API 路由
	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/search", searchHandler.Search)
	}

	fmt.Printf("Search Service is running on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run Search Service: %v", err)
	}
}
