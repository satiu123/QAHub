package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userpb "qahub/api/proto/user"
)

var (
	// gRPC 服务地址
	userServiceEndpoint = flag.String("user-service-endpoint", "localhost:50051", "User service endpoint")

	// Gateway HTTP 服务端口
	gatewayPort = flag.String("gateway-port", "8080", "Gateway HTTP port")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 创建 gRPC-Gateway mux
	mux := runtime.NewServeMux()

	// 配置 gRPC 连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// 注册 User Service
	log.Printf("Connecting to User Service at %s", *userServiceEndpoint)
	err := userpb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, *userServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register user service handler: %v", err)
	}

	// 添加 CORS 支持
	handler := corsMiddleware(mux)

	// 启动 HTTP 服务器
	serverAddr := ":" + *gatewayPort
	log.Printf("🚀 gRPC-Gateway listening on %s", serverAddr)
	log.Printf("📡 Proxying to User Service at %s", *userServiceEndpoint)
	log.Printf("📝 Example: curl http://localhost:%s/api/v1/auth/login", *gatewayPort)

	if err := http.ListenAndServe(serverAddr, handler); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// CORS 中间件
func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
