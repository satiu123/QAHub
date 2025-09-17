# 部署步骤

本项目包含多个服务。以下是部署 `user-service` 的步骤。

## 1. 先决条件

- 确保您已经安装了 Go (建议版本 1.25+)。
- 确保您的系统上正在运行 MySQL 数据库。
- 安装数据库迁移工具 `golang-migrate/migrate`:
  ```bash
  go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

## 2. 数据库设置

1.  **创建数据库:**
    在 MySQL 中，创建一个名为 `qahub` 的数据库。
    ```sql
    CREATE DATABASE IF NOT EXISTS qahub;
    ```

2.  **运行数据库迁移:**
    在项目根目录下，执行以下命令来创建 `users` 表：
    ```bash
    migrate -database "mysql://root:12345678@tcp(localhost:3306)/qahub?charset=utf8mb4&parseTime=True&loc=Local" -path scripts/migrations/user up
    ```
    > **注意:** 请根据您的实际情况修改命令中的数据库连接字符串。

## 3. 运行服务

1.  **启动 user-service:**
    ```bash
    go run ./cmd/user-service/main.go
    ```

2.  服务将会在 `8080` 端口上启动。
