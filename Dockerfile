# ---- Build Stage ----
FROM golang:1.25-alpine AS builder
ARG SERVICE_NAME

WORKDIR /workspace

# 设置 Go Module Proxy 为国内镜像，解决下载超时问题
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 依赖缓存
COPY go.work go.work
COPY go.work.sum go.work.sum
COPY cmd/user-service/go.mod cmd/user-service/go.sum ./cmd/user-service/
COPY cmd/qa-service/go.mod cmd/qa-service/go.sum ./cmd/qa-service/
COPY cmd/search-service/go.mod cmd/search-service/go.sum ./cmd/search-service/
COPY cmd/notification-service/go.mod cmd/notification-service/go.sum ./cmd/notification-service/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY api/go.mod api/go.sum ./api/

# 同步工作区并下载所有模块的依赖
RUN --mount=type=cache,target=/go/pkg/mod go work sync

# ---- 编译步骤 ----
# 复制所有源码
COPY . .

# 编译指定的服务
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-trimpath=${PWD}" -ldflags="-s -w" -o /workspace/bin/${SERVICE_NAME} ./cmd/${SERVICE_NAME}/main.go
# ---- Runtime Stage ----
FROM alpine:latest
ARG SERVICE_NAME
ARG PORT
ENV SERVICE_NAME=${SERVICE_NAME}
WORKDIR /app
COPY --from=builder /workspace/bin/${SERVICE_NAME} /app/${SERVICE_NAME}
COPY configs/config.docker.yaml /app/configs/config.yaml
EXPOSE ${PORT}
ENTRYPOINT ["/bin/sh", "-c", "/app/${SERVICE_NAME}"]