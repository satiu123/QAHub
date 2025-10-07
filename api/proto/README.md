# Protobuf 代码生成指南

本文档记录了如何使用 `protoc` 编译器从 `.proto` 文件生成 Go 语言的 gRPC 和 Protobuf 代码。

## 生成命令

在项目根目录下运行以下命令来生成代码。

### 通用命令模板

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/<service_name>/<service_name>.proto
```

**参数说明:**

*   `--go_out=.`: 指定生成的 Protobuf 消息代码（`.pb.go` 文件）的输出目录。 `.` 表示当前目录。
*   `--go_opt=paths=source_relative`: 确保生成的代码与 `.proto` 文件保持相同的相对目录结构。
*   `--go-grpc_out=.`: 指定生成的 gRPC 服务代码（`_grpc.pb.go` 文件）的输出目录。
*   `--go-grpc_opt=paths=source_relative`: 同样，确保 gRPC 代码与 `.proto` 文件保持相同的相对目录结构。
*   `api/proto/<service_name>/<service_name>.proto`: 需要编译的 `.proto` 文件的路径。

### 示例

#### 1. 生成用户服务 (User Service) 代码

```bash
protoc -I . \
       -I third_party/googleapis \
       --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
       --grpc-gateway_opt=generate_unbound_methods=true \
       api/proto/user/user.proto
```

#### 2. 生成问答服务 (QA Service) 代码

当 `api/proto/qa/qa.proto` 文件创建后，使用以下命令生成代码：

```bash
protoc -I . \
       -I third_party/googleapis \
       --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
       --grpc-gateway_opt=generate_unbound_methods=true \
       api/proto/qa/qa.proto
```

### 3. 生成搜索服务 (Search Service) 代码

```bash
protoc -I . \
       -I third_party/googleapis \
       --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
       --grpc-gateway_opt=generate_unbound_methods=true \
       api/proto/search/search.proto
```
