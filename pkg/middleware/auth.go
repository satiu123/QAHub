package middleware

import (
	"context"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"qahub/pkg/auth"
	"qahub/pkg/clients"
	"qahub/pkg/config"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GrpcAuthInterceptor 创建一个 gRPC 服务端拦截器，用于通过 user-service 验证 token
// 这个拦截器将 token 从 metadata 中提取出来，并调用 user-service 进行验证。
func GrpcAuthInterceptor(userClient *clients.UserServiceClient, publicMethods ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log.Println("gRPC method:", info.FullMethod)
		// 检查是否在白名单中
		if slices.Contains(publicMethods, info.FullMethod) {
			// 白名单路径，跳过认证
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "缺少认证信息 (metadata)")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "请求未包含授权标头")
		}

		authHeader := authHeaders[0]
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // 如果没有 "Bearer " 前缀，TrimPrefix 不会改变字符串
			return nil, status.Errorf(codes.Unauthenticated, "授权标头格式不正确，需要 'Bearer ' 前缀")
		}

		// 调用 user-service 的 ValidateToken RPC
		validateResp, err := userClient.ValidateToken(ctx, tokenString)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token 验证失败: %v", err)
		}

		// 将 structpb.Value map 转换为 jwt.MapClaims
		claims := make(map[string]any)
		for k, v := range validateResp.Claims {
			claims[k] = v.AsInterface()
		}

		// 验证成功，将用户信息注入到 context 中
		identity := auth.Identity{
			UserID:   validateResp.UserId,
			Username: validateResp.Username,
			Claims:   claims,
		}
		newCtx := auth.WithIdentity(ctx, identity)

		// 继续处理请求
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
