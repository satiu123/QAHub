package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Starting Search Service...")

	r := gin.Default()

	// TODO: 从配置中读取服务地址和端口
	port := "8083" // 临时端口

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	// TODO: 添加搜索 API 路由
	// r.GET("/search", searchHandler.Search)

	fmt.Printf("Search Service is running on port %s\n", port)
	if err := r.Run(":8083"); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
