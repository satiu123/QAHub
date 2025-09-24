package main

import (
	"net"
	"os"
	"time"

	"qahub/internal/qa/handler"
	"qahub/internal/qa/service"
	"qahub/internal/qa/store"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 1. 加载配置
	if err := config.Init("configs"); err != nil {
		os.Exit(1)
	}
	cfg := config.Conf.Services.QAService

	// 2. 初始化数据库连接
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()

	// 3. 初始化 Kafka Writer
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(config.Conf.Kafka.Brokers...),
		Topic:    config.Conf.Kafka.Topics.QAEvents,
		Balancer: &kafka.LeastBytes{},
		// 在 Docker 环境中，可能需要自定义 dialer 来确保连接成功
		Transport: &kafka.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
		},
	}
	defer kafkaWriter.Close()

	// 4. 依赖注入：初始化 store, service, handler
	qaStore := store.NewQAStore(db)
	qaService := service.NewQAService(qaStore, kafkaWriter)
	qaHandler := handler.NewQAHandler(qaService)

	// 5. 初始化 Gin 引擎
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORSMiddleware())

	// 6. 注册路由
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

			// 点赞
			apiV1.POST("/answers/:answer_id/upvote", qaHandler.UpvoteAnswer)
			apiV1.POST("/answers/:answer_id/downvote", qaHandler.DownvoteAnswer)

			// 评论
			apiV1.POST("/answers/:answer_id/comments", qaHandler.CreateComment)
			apiV1.PUT("/comments/:comment_id", qaHandler.UpdateComment)
			apiV1.DELETE("/comments/:comment_id", qaHandler.DeleteComment)

		}
	}
	// 公共 API，无需认证
	publicApiV1 := router.Group("/api/v1")
	publicApiV1.Use(middleware.OptionalAuthMiddleware()) // 可选认证中间件
	{
		// 获取问题列表的统一接口，通过查询参数区分不同场景
		// 例如: /questions?user_id=123, /questions?author=me
		publicApiV1.GET("/questions", qaHandler.ListQuestions)
		publicApiV1.GET("/questions/:question_id", qaHandler.GetQuestion)
		publicApiV1.GET("/questions/:question_id/answers", qaHandler.ListAnswers)
		publicApiV1.GET("/answers/:answer_id/comments", qaHandler.ListComments)
	}

	// 7. 启动服务器
	serverAddr := ":" + cfg.HttpPort
	if err := router.Run(serverAddr); err != nil {
		os.Exit(1)
	}
}
