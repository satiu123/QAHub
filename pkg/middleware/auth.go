package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"qahub/pkg/auth"
	"qahub/pkg/config"

	"github.com/gin-gonic/gin"
)

// NginxAuthMiddleware 创建一个Gin中间件，用于从Nginx传递的头部读取用户ID
// 这个中间件用于处理通过nginx auth_request验证后的请求
func NginxAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从nginx传递的X-User-ID头部获取用户ID
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		// 将字符串转换为int64
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无效的用户ID格式"})
			c.Abort()
			return
		}

		identity := auth.Identity{UserID: userID, Username: c.GetHeader("X-User-Name")}
		auth.InjectIntoGin(c, identity)
		c.Next()
	}
}

// OptionalAuthMiddleware 创建一个Gin中间件，用于可选的JWT身份验证
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next() // 没有token，直接继续
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next() // token格式不正确，直接继续
			return
		}

		tokenString := parts[1]
		var jwtSecret = []byte(config.Conf.Services.UserService.JWTSecret)

		identity, err := auth.ParseToken(tokenString, jwtSecret)
		if err == nil {
			auth.InjectIntoGin(c, identity)
		}
		// 无论token是否有效，都继续处理请求
		c.Next()
	}
}
