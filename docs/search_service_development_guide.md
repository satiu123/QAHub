# 搜索服务 (Search Service) 开发指南

## 概述

搜索服务是智能问答社区的核心组件之一，负责为用户提供对问题和回答内容的高效全文检索功能。它独立于其他服务，通过消息队列与核心业务服务（问答服务）进行解耦，保证了系统的高可用性和可扩展性。

## 核心架构

- **数据同步:** 服务通过消费 Kafka 中的事件来获取最新的数据。问答服务在创建或更新问题/回答后，会向 Kafka 发送消息。
- **数据索引:** 搜索服务接收到消息后，会将数据处理并索引到 Elasticsearch 中，以便进行快速检索。
- **查询接口:** 服务通过 RESTful API (使用 Gin 框架) 对外提供搜索能力。

## 技术栈

- **语言:** Go
- **API框架:** Gin
- **搜索引擎:** Elasticsearch
- **消息队列:** Kafka

## 数据流

1.  **生产端 (`qa-service`):**
    - 用户创建/更新问题或回答。
    - `qa-service` 将操作和数据封装成消息（如 `question_created`）。
    - 消息被发送到 Kafka 的 `qa_events` 主题。

2.  **消费端 (`search-service`):**
    - `search-service` 订阅 `qa_events` 主题。
    - 接收到消息后，解析其内容。
    - 将解析后的数据转换为 Elasticsearch 文档格式。
    - 调用 Elasticsearch 客户端，将文档索引。

3.  **查询端 (用户):**
    - 用户通过前端 UI 发起搜索请求。
    - API 网关将请求路由到 `search-service` 的 `/api/search` 端点。
    - `search-service` 查询 Elasticsearch 并返回结果。

## Elasticsearch 索引设计

- **索引名称:** `qahub`
- **文档结构示例 (Question):**
  ```json
  {
    "question_id": "string",
    "title": "text",
    "content": "text",
    "tags": ["keyword"],
    "author_id": "string",
    "created_at": "date"
  }
  ```

## 开发步骤

1.  **环境配置:** 在 `compose.yml` 中添加 Elasticsearch 和 Kafka 服务。
2.  **项目结构:**
    - 创建 `cmd/search-service/main.go`。
    - 创建 `internal/search` 目录存放业务逻辑。
3.  **修改 `qa-service`:** 在 `Create/Update` 方法中添加 Kafka 生产逻辑。
4.  **实现 `search-service`:**
    - 实现 Kafka 消费者来监听和处理消息。
    - 实现 Elasticsearch 客户端来索引数据。
    - 使用 Gin 创建 `/search` API 端点来处理用户查询。
5.  **测试:** 进行端到端的集成测试，验证从创建问题到能够搜索到的完整流程。
