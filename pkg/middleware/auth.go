package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"qahub/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// var jwtSecret = []byte(config.Conf.Services.UserService.JWTSecret) // Moved into AuthMiddleware

// AuthMiddleware 创建一个Gin中间件，用于JWT身份验证
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请求未包含授权标头"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "授权标头格式不正确"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		var jwtSecret = []byte(config.Conf.Services.UserService.JWTSecret)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			// 确保token的签名方法是我们期望的
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 从claims中获取用户ID
			if userID, ok := claims["user_id"].(float64); ok {
				// 将用户ID存储在上下文中，以便后续的处理函数使用
				c.Set("userID", int64(userID))
				c.Next()
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token中缺少用户信息"})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token claims"})
			c.Abort()
			return
		}
	}
}
