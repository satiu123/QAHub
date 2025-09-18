# 问答服务（QA Service）开发指南

本文档提供了构建问答服务的详细步骤，旨在与现有项目结构保持一致。

---

### 第 1 步：定义数据模型和数据库表结构

首先，我们需要为“问题（Question）”和“回答（Answer）”设计数据库表。

1.  **设计表结构:**
    *   `questions` 表: `id`, `user_id` (提问者), `title` (标题), `content` (内容), `created_at`, `updated_at`。
    *   `answers` 表: `id`, `question_id` (关联问题), `user_id` (回答者), `content` (内容), `upvote_count` (点赞数), `created_at`, `updated_at`。

2.  **创建数据库迁移文件:**
    在 `scripts/migrations/` 目录下，为问答服务创建一个新的子目录 `qa`。然后创建迁移文件，就像用户服务那样：
    *   `scripts/migrations/qa/000001_create_questions_table.up.sql`
    *   `scripts/migrations/qa/000001_create_questions_table.down.sql`
    *   `scripts/migrations/qa/000002_create_answers_table.up.sql`
    *   `scripts/migrations/qa/000002_create_answers_table.down.sql`

---

### 第 2 步：创建 gRPC 和 Protobuf 定义

为了让其他微服务（如未来的搜索服务）能与问答服务通信，我们需要定义 gRPC 接口。

1.  **编写 `.proto` 文件:**
    在 `api/proto/qa/` 目录下创建一个 `qa.proto` 文件。定义服务和消息体：
    *   `service QaService { ... }`
    *   定义 `CreateQuestion`, `GetQuestion`, `CreateAnswer` 等 RPC 方法。
    *   定义 `Question`, `Answer` 等消息（Message）结构。

2.  **生成 Go 代码:**
    使用 `protoc` 工具根据 `qa.proto` 文件生成 Go 代码（`.pb.go` 和 `_grpc.pb.go` 文件）。

---

### 第 3 步：实现核心业务逻辑

这是编码的主要部分。在 `internal/` 目录下创建 `qa` 目录，并参照 `internal/user` 的结构创建子目录和文件。

1.  **`internal/qa/model/`**:
    *   创建 `qa_model.go` 文件，定义与数据库表对应的 `Question` 和 `Answer` Go 结构体。

2.  **`internal/qa/store/`**:
    *   创建 `qa_store.go` 文件，定义数据访问层的接口（例如 `CreateQuestion`, `GetAnswerByID` 等）。
    *   实现这个接口，编写与数据库交互的实际代码。

3.  **`internal/qa/service/`**:
    *   创建 `qa_service.go` 文件，编写核心业务逻辑。它会调用 `store` 层来操作数据，并处理各种业务规则（例如，检查问题是否存在才能回答）。

4.  **`internal/qa/handler/`**:
    *   创建 `grpc_handler.go`：实现由 Protobuf 生成的 gRPC 服务接口。这个处理程序会调用 `service` 层来完成具体工作。
    *   创建 `http_handler.go`（可选，如果需要对外提供 REST API）：创建 Gin 框架的 HTTP 处理函数，同样调用 `service` 层。

---

### 第 4 步：构建服务入口

1.  **编辑 `cmd/qa-service/main.go`**:
    *   这是问答服务的启动文件。
    *   在这里，您需要：
        *   加载配置。
        *   初始化数据库连接。
        *   初始化 `store`, `service`, `handler`。
        *   创建并注册 gRPC 服务器。
        *   （可选）创建并注册 Gin HTTP 服务器。
        *   启动服务并监听端口。

---

### 第 5 步：更新项目配置

1.  **编辑 `configs/config.yaml`**:
    *   为 `qa-service` 添加新的配置节，包括数据库连接信息、服务端口等。

2.  **编辑 `docker-compose.yml`**:
    *   为 `qa-service` 添加一个新的服务定义，以便能通过 Docker 运行它。

---

### 总结

1.  **数据库先行**：创建迁移文件。
2.  **接口定义**：编写 `.proto` 文件并生成代码。
3.  **由内而外编码**：依次实现 `model` -> `store` -> `service` -> `handler`。
4.  **整合启动**：编写 `main.go` 将所有部分连接起来。
5.  **配置和部署**：更新 `config.yaml` 和 `docker-compose.yml`。
