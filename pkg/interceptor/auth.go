package interceptor

import (
	"context"
	"qahub/pkg/auth"
	"qahub/pkg/clients"
	"slices"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthUnaryServerInterceptor 创建一个 gRPC 服务端拦截器，用于通过 user-service 验证 token
// 这个拦截器将 token 从 metadata 中提取出来，并调用 user-service 进行验证。
func AuthUnaryServerInterceptor(userClient *clients.UserServiceClient, publicMethods ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
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

// AuthStreamServerInterceptor 创建一个 gRPC 流服务端拦截器，用于通过 user-service 验证 token
// 这个拦截器将 token 从 metadata 中提取出来，并调用 user-service 进行验证。
func AuthStreamServerInterceptor(userClient *clients.UserServiceClient, publicMethods ...string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 检查是否在白名单中
		if slices.Contains(publicMethods, info.FullMethod) {
			// 白名单路径，跳过认证
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Errorf(codes.Unauthenticated, "缺少认证信息 (metadata)")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return status.Errorf(codes.Unauthenticated, "请求未包含授权标头")
		}

		authHeader := authHeaders[0]
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // 如果没有 "Bearer " 前缀，TrimPrefix 不会改变字符串
			return status.Errorf(codes.Unauthenticated, "授权标头格式不正确，需要 'Bearer ' 前缀")
		}

		// 调用 user-service 的 ValidateToken RPC
		validateResp, err := userClient.ValidateToken(ss.Context(), tokenString)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "token 验证失败: %v", err)
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
		newCtx := auth.WithIdentity(ss.Context(), identity)

		// 使用新的上下文创建一个包装的 ServerStream
		wrappedSS := newWrappedServerStream(newCtx, ss)

		// 继续处理请求
		return handler(srv, wrappedSS)
	}
}
