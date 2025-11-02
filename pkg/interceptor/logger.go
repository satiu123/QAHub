package interceptor

import (
	"context"
	"log/slog"
	"qahub/pkg/log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// wrappedServerStream 是对 grpc.ServerStream 的包装，允许我们修改上下文
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// newWrappedServerStream 创建一个包装后的 ServerStream
func newWrappedServerStream(ctx context.Context, ss grpc.ServerStream) *wrappedServerStream {
	return &wrappedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}

// Context 返回包装后的上下文
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// LogUnaryServerInterceptor 创建一个 gRPC 服务端拦截器，用于记录每个请求的日志
func LogUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 获取请求 ID
		var requestID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ids := md.Get("x-request-id")
			if len(ids) > 0 {
				requestID = ids[0]
			}
		}

		// 如果请求 ID 不存在，则生成一个新的请求 ID
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 创建带有上下文的 logger
		reqLogger := slog.Default().With(
			"x-request-id", requestID,
			"grpc.method", info.FullMethod,
		)

		// 将 logger 存入上下文
		newCtx := log.WithContext(ctx, reqLogger)

		startTime := time.Now()
		// 调用下一个处理器
		resp, err = handler(newCtx, req)

		// 记录请求完成的日志
		duration := time.Since(startTime)
		code := status.Code(err)
		logAddrs := []any{
			slog.String("grpc.code", code.String()),
			// slog.Duration("duration", duration), // 流持续时间，不直观
			slog.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		}

		if err != nil {
			reqLogger.Error("gRPC request completed with error", slog.Any("error", err), slog.Group("details", logAddrs...))
		} else {
			reqLogger.Info("gRPC request completed", slog.Group("details", logAddrs...))
		}
		return resp, err

	}
}

// LogStreamServerInterceptor 创建一个 gRPC 流服务端拦截器，用于记录每个流请求的日志
func LogStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 获取请求 ID
		var requestID string
		if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
			ids := md.Get("x-request-id")
			if len(ids) > 0 {
				requestID = ids[0]
			}
		}
		// 如果请求 ID 不存在，则生成一个新的请求 ID
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 创建带有上下文的 logger
		streamLogger := slog.Default().With(
			"x-request-id", requestID,
			"grpc.method", info.FullMethod,
			"grpc.is_stream", true,
		)

		// 将 logger 存入上下文
		newCtx := log.WithContext(ss.Context(), streamLogger)
		// 包装 ServerStream 以使用新的上下文和日志记录
		wrappedSS := newWrappedServerStream(newCtx, ss)

		startTime := time.Now()
		streamLogger.Debug("gRPC stream started")

		// 调用下一个处理器
		err := handler(srv, wrappedSS)

		duration := time.Since(startTime)
		code := status.Code(err)

		logAddrs := []any{
			slog.String("grpc.code", code.String()),
			// slog.Duration("duration", duration),                              // 流持续时间，不直观
			slog.Float64("duration_ms", float64(duration.Nanoseconds())/1e6), // 以毫秒为单位的持续时间
		}

		if err != nil {
			streamLogger.Error("gRPC stream completed with error", slog.Any("error", err), slog.Group("details", logAddrs...))
		} else {
			streamLogger.Info("gRPC stream completed", slog.Group("details", logAddrs...))
		}

		return err

	}
}

func generateRequestID() string {
	reqID, err := uuid.NewUUID()
	if err != nil {
		return ""
	}
	return reqID.String()
}
