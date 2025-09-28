package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"qahub/internal/user/store"
	"qahub/pkg/auth"
	"qahub/pkg/config"

	"github.com/gin-gonic/gin"
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
		identity, err := auth.ParseToken(tokenString, []byte(config.Conf.Services.UserService.JWTSecret))
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "无效的token: %v", err)
		}

		newCtx := auth.WithIdentity(ctx, identity)
		return handler(newCtx, req)
	}
}

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

// AuthMiddleware 创建一个Gin中间件，用于JWT身份验证（强制）
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

		identity, err := auth.ParseToken(tokenString, []byte(config.Conf.Services.UserService.JWTSecret))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

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
