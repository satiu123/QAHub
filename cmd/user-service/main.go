package main

import (
	"qahub/internal/user/handler"
	"qahub/internal/user/service"
	"qahub/internal/user/store"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterRoutes(handler *handler.UserHandler) {
	router := gin.Default()
	api := router.Group("/api")
	userGroup := api.Group("/users")
	{
		userGroup.POST("/register", handler.Register)
		userGroup.POST("/login", handler.Login)
		userGroup.GET("/:id", handler.GetProfile)
	}

	router.Run(":8080")
}

func main() {
	dsn := "root:12345678@tcp(localhost:3306)/qahub?charset=utf8mb4&parseTime=True&loc=Local"
	db := sqlx.MustConnect("mysql", dsn)
	defer db.Close()

	userStore := store.NewMySQLUserStore(db)
	userService := service.NewUserService(userStore)
	UserHandler := handler.NewUserHandler(userService)
	RegisterRoutes(UserHandler)

}
