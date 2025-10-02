# 多阶段构建
FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# 复制 go.work 文件
COPY go.work go.work
COPY go.work.sum go.work.sum

# 复制所有 go.mod 和 go.sum（go.work 中的所有模块）
COPY cmd/gateway/go.mod cmd/gateway/go.sum ./cmd/gateway/
COPY cmd/user-service/go.mod cmd/user-service/go.sum ./cmd/user-service/
COPY cmd/qa-service/go.mod cmd/qa-service/go.sum ./cmd/qa-service/
COPY cmd/search-service/go.mod cmd/search-service/go.sum ./cmd/search-service/
COPY cmd/notification-service/go.mod cmd/notification-service/go.sum ./cmd/notification-service/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY api/go.mod api/go.sum ./api/

# 下载依赖
RUN --mount=type=cache,target=/go/pkg/mod \
    go work sync

WORKDIR /workspace/cmd/gateway
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 复制源代码
WORKDIR /workspace
COPY . .

# 编译
WORKDIR /workspace/cmd/gateway
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /workspace/bin/gateway

# 最终镜像
FROM alpine:latest

WORKDIR /app

# 复制二进制文件
COPY --from=builder /workspace/bin/gateway /app/gateway

# 暴露端口
EXPOSE 8080

CMD ["/app/gateway"]
