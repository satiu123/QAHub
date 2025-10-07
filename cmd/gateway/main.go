package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	qapb "qahub/api/proto/qa"
	searchpb "qahub/api/proto/search"
	userpb "qahub/api/proto/user"
	"qahub/pkg/config"
)

func main() {
	if err := config.Init("configs"); err != nil {
		os.Exit(1)
	}
	// 读取配置
	gatewayConfig := config.Conf.Services.Gateway
	gatewayPort := gatewayConfig.Port

	userServiceEndpoint := gatewayConfig.UserServiceEndpoint
	qaServiceEndpoint := gatewayConfig.QaServiceEndpoint
	searchServiceEndpoint := gatewayConfig.SearchServiceEndpoint

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 自定义JSONPb编解码器
	jpb := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	// 创建 gRPC-Gateway mux
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jpb),
	)

	// 配置 gRPC 连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// 注册 User Service
	log.Printf("Connecting to User Service at %s", userServiceEndpoint)
	err := userpb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, userServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register user service handler: %v", err)
	}
	// 注册 QA Service
	log.Printf("Connecting to QA Service at %s", qaServiceEndpoint)
	err = qapb.RegisterQAServiceHandlerFromEndpoint(ctx, mux, qaServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register QA service handler: %v", err)
	}
	// 注册 Search Service
	log.Printf("Connecting to Search Service at %s", searchServiceEndpoint)
	err = searchpb.RegisterSearchServiceHandlerFromEndpoint(ctx, mux, searchServiceEndpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register search service handler: %v", err)
	}

	// 添加 CORS 支持
	handler := corsMiddleware(mux)

	// 启动 HTTP 服务器
	serverAddr := ":" + gatewayPort
	log.Printf("🚀 gRPC-Gateway listening on %s", serverAddr)
	log.Printf("📡 Proxying to User Service at %s", userServiceEndpoint)
	log.Printf("📡 Proxying to QA Service at %s", qaServiceEndpoint)
	log.Printf("📝 Example: curl http://localhost:%s/api/v1/auth/login", gatewayPort)

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
