package main

import (
	"log"
	"qahub/internal/user/handler"
	"qahub/internal/user/service"
	"qahub/internal/user/store"
	"qahub/pkg/config"
	"qahub/pkg/database"
	"qahub/pkg/middleware" // 导入中间件包
	"qahub/pkg/redis"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(handler *handler.UserHandler, port string) {
	router := gin.Default()
	api := router.Group("/api")
	userGroup := api.Group("/users")
	{
		// 公开路由
		userGroup.POST("/register", handler.Register)
		userGroup.POST("/login", handler.Login)
		userGroup.GET("/:id", handler.GetProfile) // 查看任何人信息是公开的

		// 需要认证的路由
		authRequired := userGroup.Group("/")
		authRequired.Use(middleware.AuthMiddleware())
		{
			authRequired.PUT("/:id", handler.UpdateProfile) // 更新用户信息
			authRequired.DELETE("/:id", handler.DeleteUser) // 删除用户
		}
	}

	router.Run(":" + port)
}

func main() {
	// 加载配置
	if err := config.Init("configs"); err != nil {
		log.Fatal("配置加载失败:", err)
	}
	// 创建数据库连接
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()
	// 创建 Redis 客户端
	redisClient, err := redis.NewClient(config.Conf.Redis)
	if err != nil {
		log.Fatal("Redis 连接失败:", err)
	}
	defer redisClient.Close()

	userStore := store.NewMySQLUserStore(db)
	cacheStore := store.NewUserCacheStore(redisClient, userStore)
	userService := service.NewUserService(cacheStore)
	UserHandler := handler.NewUserHandler(userService)
	RegisterRoutes(UserHandler, config.Conf.Server.Port)

}
