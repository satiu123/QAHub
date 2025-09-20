package main

import (
	"log"

	"qahub/internal/qa/handler"
	"qahub/internal/qa/service"
	"qahub/internal/qa/store"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/middleware"

	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 1. 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	cfg := config.Conf.Services.QAService

	// 2. 初始化数据库连接
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. 依赖注入：初始化 store, service, handler
	qaStore := store.NewQAStore(db)
	qaService := service.NewQAService(qaStore)
	qaHandler := handler.NewQAHandler(qaService)

	// 4. 初始化 Gin 引擎
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORSMiddleware())

	// 5. 注册路由
	// 创建一个新的顶层分组来匹配 Nginx 添加的前缀
	protectedQA := router.Group("/_protected_qa")
	protectedQA.Use(middleware.NginxAuthMiddleware()) // 添加nginx认证中间件
	{
		apiV1 := protectedQA.Group("/api/v1")
		{
			// 问题
			apiV1.POST("/questions", qaHandler.CreateQuestion)
			apiV1.PUT("/questions/:question_id", qaHandler.UpdateQuestion)
			apiV1.DELETE("/questions/:question_id", qaHandler.DeleteQuestion)

			// 回答
			apiV1.POST("/questions/:question_id/answers", qaHandler.CreateAnswer)
			apiV1.PUT("/answers/:answer_id", qaHandler.UpdateAnswer)
			apiV1.DELETE("/answers/:answer_id", qaHandler.DeleteAnswer)

			// 评论
			apiV1.POST("/answers/:answer_id/comments", qaHandler.CreateComment)
			apiV1.PUT("/comments/:comment_id", qaHandler.UpdateComment)
			apiV1.DELETE("/comments/:comment_id", qaHandler.DeleteComment)

		}
	}
	// 公共 API，无需认证
	publicApiV1 := router.Group("/api/v1")
	{
		publicApiV1.GET("/questions", qaHandler.ListQuestions)
		publicApiV1.GET("/questions/:question_id", qaHandler.GetQuestion)
		publicApiV1.GET("/questions/:question_id/answers", qaHandler.ListAnswers)
		publicApiV1.GET("/answers/:answer_id/comments", qaHandler.ListComments)
	}

	// 6. 启动服务器
	serverAddr := ":" + cfg.HttpPort
	log.Printf("QA service starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
