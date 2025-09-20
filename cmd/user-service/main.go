package main

import (
	"log"

	"qahub/internal/user/handler"
	"qahub/internal/user/service"
	"qahub/internal/user/store"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/middleware"
	"qahub/pkg/redis"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 1. 加载配置
	if err := config.Init("/app/configs"); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 2. 初始化数据库和 Redis
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.NewClient(config.Conf.Redis)
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	defer redisClient.Close()

	// 3. 依赖注入
	// 注意：这里我们同时需要一个支持黑名单的 UserStore
	userStoreWithBlacklist := store.NewUserCacheStore(redisClient, store.NewMySQLUserStore(db))
	userService := service.NewUserService(userStoreWithBlacklist)
	userHandler := handler.NewUserHandler(userService)

	// 4. 初始化 Gin 引擎
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORSMiddleware())

	// 5. 注册路由
	apiV1 := router.Group("/api/v1")
	{
		// 公开路由
		apiV1.POST("/users/register", userHandler.Register)
		apiV1.POST("/users/login", userHandler.Login)

		apiV1.GET("/auth/validate", userHandler.ValidateToken)

		apiV1.POST("/users/logout", userHandler.Logout)
		// 注意：GET /users/me 是一个潜在的扩展点，用于专门获取当前登录用户的配置文件
		apiV1.GET("/users/:id", userHandler.GetProfile)
		apiV1.PUT("/users/:id", userHandler.UpdateProfile)
		apiV1.DELETE("/users/:id", userHandler.DeleteUser)
	}

	// 6. 启动服务器
	serverAddr := ":" + config.Conf.Services.UserService.HttpPort
	log.Printf("User service starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
