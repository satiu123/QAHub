# ---- Build Stage ----
# 使用官方的 Go 镜像作为编译环境
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件并下载依赖
# 这样可以利用 Docker 的层缓存机制，只有在依赖变化时才重新下载
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 编译 user-service 应用
# -o 指定输出文件名，CGO_ENABLED=0 禁用 CGO 以便静态链接
# -ldflags "-s -w" 去除调试信息，减小文件体积
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/user-service /app/cmd/user-service/main.go

FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从 builder 阶段复制编译好的二进制文件和配置文件
COPY --from=builder /app/user-service .
COPY configs/config.docker.yaml /app/configs/config.yaml

# 暴露服务端口 (在容器内部)
# 这个端口应该与 config.yaml 中为 user-service 配置的端口一致
EXPOSE 8081

# 容器启动时运行的命令
ENTRYPOINT ["/app/user-service"]
