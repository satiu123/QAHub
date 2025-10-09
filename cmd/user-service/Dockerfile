# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# 预热 go.work 及各模块的 go.mod/go.sum，最大化缓存命中
COPY go.work go.work
COPY go.work.sum go.work.sum
COPY cmd/user-service/go.mod cmd/user-service/go.sum ./cmd/user-service/
COPY cmd/qa-service/go.mod cmd/qa-service/go.sum ./cmd/qa-service/
COPY cmd/search-service/go.mod cmd/search-service/go.sum ./cmd/search-service/
COPY cmd/notification-service/go.mod cmd/notification-service/go.sum ./cmd/notification-service/
COPY cmd/gateway/go.mod cmd/gateway/go.sum ./cmd/gateway/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY api/go.mod api/go.sum ./api/
COPY clients/wails-client/go.mod clients/wails-client/go.sum ./clients/wails-client/

RUN --mount=type=cache,target=/go/pkg/mod \
    go work sync

# 下载用户服务依赖
WORKDIR /workspace/cmd/user-service
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 复制剩余源码
WORKDIR /workspace
COPY . .

# 编译用户服务
WORKDIR /workspace/cmd/user-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /workspace/bin/user-service .

# ---- Runtime Stage ----
FROM alpine:latest

WORKDIR /app

COPY --from=builder /workspace/bin/user-service /app/user-service
COPY --from=builder /workspace/configs/config.docker.yaml /app/configs/config.yaml

# 暴露 gRPC 端口
EXPOSE 50051

ENTRYPOINT ["/app/user-service"]
