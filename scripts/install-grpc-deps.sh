#!/bin/bash

echo "📦 安装 gRPC 和 gRPC-Gateway 依赖..."

# 安装 protoc 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# 添加必要的依赖到 go.mod
cd /home/satiu/code/QAHub/api

go get google.golang.org/grpc@latest
go get google.golang.org/protobuf/reflect/protoreflect@latest
go get google.golang.org/protobuf/runtime/protoimpl@latest
go get github.com/grpc-ecosystem/grpc-gateway/v2/runtime@latest
go get github.com/grpc-ecosystem/grpc-gateway/v2/utilities@latest
go get google.golang.org/genproto/googleapis/api/annotations@latest

# 下载依赖
go mod download
go mod tidy

echo "✅ 依赖安装完成！"