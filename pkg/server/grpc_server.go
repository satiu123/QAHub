package server

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// GrpcServer 封装了 gRPC 服务器的通用逻辑
type GrpcServer struct {
	grpcSrv     *grpc.Server
	healthSrv   *health.Server
	serviceName string
	port        string
}

// NewGrpcServer 创建一个新的通用 gRPC 服务器实例
// opts 参数允许每个服务传入自己独有的拦截器等 gRPC 选项
func NewGrpcServer(serviceName, port string, opts ...grpc.ServerOption) *GrpcServer {
	grpcSrv := grpc.NewServer(opts...)
	healthSrv := health.NewServer()

	// 自动注册健康检查和反射服务
	healthv1.RegisterHealthServer(grpcSrv, healthSrv)
	reflection.Register(grpcSrv)

	log.Println("gRPC Health Check and Reflection services have been registered.")

	return &GrpcServer{
		grpcSrv:     grpcSrv,
		healthSrv:   healthSrv,
		serviceName: serviceName,
		port:        port,
	}
}

// Run 启动服务器，并处理优雅关闭
// registerBusinessServer 是一个回调函数，用于注册每个微服务自己的业务 handler
func (s *GrpcServer) Run(registerBusinessServer func(srv *grpc.Server)) {
	// 1. 注册该微服务的具体业务
	registerBusinessServer(s.grpcSrv)
	log.Printf("Business service '%s' has been registered.", s.serviceName)

	// 2. 设置初始健康状态
	s.healthSrv.SetServingStatus(s.serviceName, healthv1.HealthCheckResponse_SERVING)
	log.Printf("Service '%s' is healthy and serving.", s.serviceName)

	// 3. 启动监听
	serverAddr := ":" + s.port
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", s.port, err)
	}

	go func() {
		log.Printf("gRPC server is listening on %v", lis.Addr())
		if err := s.grpcSrv.Serve(lis); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// 4. 等待关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 5. 执行优雅关闭
	s.healthSrv.SetServingStatus(s.serviceName, healthv1.HealthCheckResponse_NOT_SERVING)

	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.grpcSrv.GracefulStop()
	log.Printf("Server '%s' shut down gracefully.", s.serviceName)
}
