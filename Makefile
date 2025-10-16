.PHONY: help test test-coverage lint fmt clean build docker-build docker-up docker-down

# 默认目标
help:
	@echo "QAHub 项目 Makefile"
	@echo ""
	@echo "可用命令:"
	@echo "  make test              - 运行所有单元测试"
	@echo "  make test-coverage     - 运行测试并生成覆盖率报告"
	@echo "  make lint              - 运行代码检查"
	@echo "  make fmt               - 格式化代码"
	@echo "  make clean             - 清理构建产物"
	@echo "  make build             - 构建所有服务"
	@echo "  make docker-build      - 构建 Docker 镜像"
	@echo "  make docker-up         - 启动 Docker Compose 服务"
	@echo "  make docker-down       - 停止 Docker Compose 服务"

# 运行所有测试
test:
	@echo "运行单元测试..."
	@go test -v work

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	@go test work
	@echo ""
	@echo "=== pkg 覆盖率 ==="
	@go tool cover -func=pkg-coverage.out | tail -1
	@echo ""
	@echo "=== user-service 覆盖率 ==="
	@go tool cover -func=user-coverage.out | tail -1
	@echo ""
	@echo "=== qa-service 覆盖率 ==="
	@go tool cover -func=qa-coverage.out | tail -1
	@echo ""
	@echo "生成 HTML 报告: go tool cover -html=<coverage-file>.out"

# 运行代码检查
lint:
	@echo "运行代码检查..."
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint: https://golangci-lint.run/docs/welcome/install/" && exit 1)
	@exit_code=0; \
	echo ""; \
	echo "=== 检查 api ==="; \
	cd api && golangci-lint run --path-prefix=api || exit_code=1; \
	cd ..; \
	echo ""; \
	echo "=== 检查 pkg ==="; \
	cd pkg && golangci-lint run --path-prefix=pkg || exit_code=1; \
	cd ..; \
	echo ""; \
	echo "=== 检查 user-service ==="; \
	cd cmd/user-service && golangci-lint run --path-prefix=cmd/user-service || exit_code=1; \
	cd ../..; \
	echo ""; \
	echo "=== 检查 qa-service ==="; \
	cd cmd/qa-service && golangci-lint run --path-prefix=cmd/qa-service || exit_code=1; \
	cd ../..; \
	echo ""; \
	echo "=== 检查 search-service ==="; \
	cd cmd/search-service && golangci-lint run --path-prefix=cmd/search-service || exit_code=1; \
	cd ../..; \
	echo ""; \
	echo "=== 检查 notification-service ==="; \
	cd cmd/notification-service && golangci-lint run --path-prefix=cmd/notification-service || exit_code=1; \
	cd ../..; \
	echo ""; \
	echo "=== 检查 wails-client ==="; \
	cd clients/wails-client && golangci-lint run --path-prefix=clients/wails-client || exit_code=1; \
	cd ../..; \
	echo ""; \
	if [ "$$exit_code" != "0" ]; then \
		echo "❌ 代码检查发现问题"; \
		exit 1; \
	else \
		echo "✓ 所有模块代码检查完成！"; \
	fi

# 格式化代码
fmt:
	@echo "格式化代码..."
	@gofmt -s -w .
	@goimports -w .

# 清理构建产物
clean:
	@echo "清理构建产物..."
	@rm -f *-coverage.out
	@rm -rf cmd/*/bin
	@go clean -cache -testcache -modcache

# 构建所有服务
build:
	@echo "构建 user-service..."
	@cd cmd/user-service && go build -o bin/user-service .
	@echo "构建 qa-service..."
	@cd cmd/qa-service && go build -o bin/qa-service .
	@echo "构建 search-service..."
	@cd cmd/search-service && go build -o bin/search-service .
	@echo "构建 notification-service..."
	@cd cmd/notification-service && go build -o bin/notification-service .
	@echo "构建完成！"

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像..."
	@docker-compose build

# 启动服务
docker-up:
	@echo "启动 Docker Compose 服务..."
	@docker-compose up -d

# 停止服务
docker-down:
	@echo "停止 Docker Compose 服务..."
	@docker-compose down

# 查看服务日志
docker-logs:
	@docker-compose logs -f

# 生成 mock 文件
generate-mocks:
	@echo "生成 mock 文件..."
	@cd cmd/user-service/internal/service && go generate
	@cd cmd/qa-service/internal/service && go generate
	@echo "Mock 文件生成完成！"

# 安装依赖
install-deps:
	@echo "安装项目依赖..."
	@go mod download
	@cd cmd/user-service && go mod download
	@cd cmd/qa-service && go mod download
	@cd cmd/notification-service && go mod download
	@cd cmd/search-service && go mod download
	@echo "依赖安装完成！"

# 安装开发工具
install-tools:
	@echo "安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install go.uber.org/mock/mockgen@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "开发工具安装完成！"
