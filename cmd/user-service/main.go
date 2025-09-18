package main

import (
	"log"
	"qahub/internal/user/handler"
	"qahub/internal/user/service"
	"qahub/internal/user/store"
	"qahub/pkg/config"
	"qahub/pkg/database"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(handler *handler.UserHandler, port string) {
	router := gin.Default()
	api := router.Group("/api")
	userGroup := api.Group("/users")
	{
		userGroup.POST("/register", handler.Register)
		userGroup.POST("/login", handler.Login)
		userGroup.GET("/:id", handler.GetProfile)
	}

	router.Run(":" + port)
}

func main() {
	if err := config.Init("configs"); err != nil {
		log.Fatal("配置加载失败:", err)
	}
	db, err := database.NewMySQLConnection(config.Conf.MySQL)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	userStore := store.NewMySQLUserStore(db)
	userService := service.NewUserService(userStore)
	UserHandler := handler.NewUserHandler(userService)
	RegisterRoutes(UserHandler, config.Conf.Server.Port)

}
