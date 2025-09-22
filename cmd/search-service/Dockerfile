# ---- Build Stage ----
# 使用官方的 Go 镜像作为编译环境
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 编译 search-service 应用
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/search-service /app/cmd/search-service/main.go

# ---- Release Stage ----
# 使用一个非常小的基础镜像作为运行环境
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从 builder 阶段复制编译好的二进制文件和配置文件
COPY --from=builder /app/search-service .
COPY configs/config.docker.yaml /app/configs/config.yaml

# 暴露服务端口 (在容器内部)
# 这个端口应该与 nginx.conf 中为 search-service 配置的端口一致 (8083)
EXPOSE 8083

# 容器启动时运行的命令
ENTRYPOINT ["/app/search-service"]
