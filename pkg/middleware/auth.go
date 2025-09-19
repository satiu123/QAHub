package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"qahub/internal/user/store"
	"qahub/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GrpcAuthInterceptor 创建一个 gRPC 一元拦截器，用于JWT身份验证
func GrpcAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		// 定义需要保护的方法列表
		protectedMethods := map[string]bool{
			"/user.UserService/UpdateUserProfile": true,
			"/user.UserService/DeleteUser":        true,
		}

		// 如果当前方法不需要保护，则直接跳过
		if !protectedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "缺少认证信息")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "请求未包含授权标头")
		}

		parts := strings.Split(authHeaders[0], " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, status.Errorf(codes.Unauthenticated, "授权标头格式不正确")
		}

		tokenString := parts[1]
		var jwtSecret = []byte(config.Conf.Services.UserService.JWTSecret)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "无效的token: %v", err)
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if userID, ok := claims["user_id"].(float64); ok {
				// 将用户ID注入到新的context中
				newCtx := context.WithValue(ctx, "userID", int64(userID))
				return handler(newCtx, req)
			} else {
				return nil, status.Errorf(codes.Unauthenticated, "token中缺少用户信息")
			}
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "无效的token claims")
		}
	}
}

// AuthMiddleware 创建一个Gin中间件，用于JWT身份验证
func AuthMiddleware(userStore store.UserStore) gin.HandlerFunc {
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

		// 检查 token 是否在黑名单中
		if blacklister, ok := userStore.(store.TokenBlacklister); ok {
			isBlacklisted, err := blacklister.IsBlacklisted(tokenString)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
				c.Abort()
				return
			}
			if isBlacklisted {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
				c.Abort()
				return
			}
		}

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
