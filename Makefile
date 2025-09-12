# Domain MAX Makefile

.PHONY: help build clean test lint dev docker-build docker-up docker-down install deps

# Default target
help: ## Show this help message
	@echo "Domain MAX - 域名与DNS管理平台"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 安装依赖
install: deps ## Install all dependencies
	@echo "📦 安装依赖..."
	go mod tidy
	cd web && npm ci

deps: ## Download Go dependencies
	@echo "📦 下载Go依赖..."
	go mod download

# 开发相关
dev: ## Start development server
	@echo "🚀 启动开发服务器..."
	@echo "前端: http://localhost:5173"
	@echo "后端: http://localhost:8080"
	@echo ""
	@echo "请在另一个终端运行: cd web && npm run dev"
	go run ./cmd/server

dev-web: ## Start frontend development server
	@echo "🌐 启动前端开发服务器..."
	cd web && npm run dev

# 构建相关
build: build-web build-server ## Build both frontend and backend

build-web: ## Build frontend
	@echo "🏗️  构建前端..."
	cd web && npm run build

build-server: ## Build backend
	@echo "🏗️  构建后端..."
	CGO_ENABLED=0 go build -ldflags="-w -s" -o domain-max ./cmd/server

build-linux: ## Build for Linux
	@echo "🏗️  构建Linux版本..."
	cd web && npm run build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o domain-max-linux ./cmd/server

build-windows: ## Build for Windows
	@echo "🏗️  构建Windows版本..."
	cd web && npm run build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o domain-max.exe ./cmd/server

build-all: build-linux build-windows ## Build for all platforms

# 测试相关
test: ## Run all tests
	@echo "🧪 运行测试..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "🧪 运行测试并生成覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

test-web: ## Run frontend tests
	@echo "🧪 运行前端测试..."
	cd web && npm test

# 代码质量
lint: ## Run linters
	@echo "🔍 代码检查..."
	golangci-lint run
	cd web && npm run lint

fmt: ## Format code
	@echo "🎨 格式化代码..."
	go fmt ./...
	cd web && npm run lint --fix

# 清理相关
clean: ## Clean build artifacts
	@echo "🧹 清理构建产物..."
	rm -f domain-max domain-max.exe domain-max-linux
	rm -rf web/dist web/node_modules
	rm -f coverage.out coverage.html

clean-all: clean ## Clean everything including caches
	@echo "🧹 深度清理..."
	go clean -cache -modcache
	cd web && npm cache clean --force

# Docker相关
docker-build: ## Build Docker image
	@echo "🐳 构建Docker镜像..."
	docker build -f deployments/Dockerfile -t domain-max:latest .

docker-up: ## Start services with Docker Compose
	@echo "🐳 启动Docker服务..."
	cd deployments && docker-compose up -d

docker-down: ## Stop Docker services
	@echo "🐳 停止Docker服务..."
	cd deployments && docker-compose down

docker-logs: ## Show Docker logs
	@echo "📋 查看Docker日志..."
	cd deployments && docker-compose logs -f

docker-rebuild: docker-down docker-build docker-up ## Rebuild and restart Docker services

# 数据库相关
db-migrate: ## Run database migrations
	@echo "🗄️  执行数据库迁移..."
	go run ./cmd/server --migrate-only

db-seed: ## Seed database with sample data
	@echo "🌱 填充示例数据..."
	psql -h localhost -U postgres -d domain_manager -f configs/init.sql

# 部署相关
deploy-staging: ## Deploy to staging environment
	@echo "🚀 部署到测试环境..."
	./scripts/deploy.sh staging

deploy-production: ## Deploy to production environment
	@echo "🚀 部署到生产环境..."
	./scripts/deploy.sh production

# 安全检查
security-check: ## Run security checks
	@echo "🔒 安全检查..."
	gosec ./...
	cd web && npm audit

# 性能测试
benchmark: ## Run benchmarks
	@echo "⚡ 性能测试..."
	go test -bench=. -benchmem ./...

# 生成文档
docs: ## Generate documentation
	@echo "📚 生成文档..."
	godoc -http=:6060 &
	@echo "文档服务: http://localhost:6060"

# 版本管理
version: ## Show version information
	@echo "Domain MAX 版本信息:"
	@echo "Go版本: $(shell go version)"
	@echo "Node版本: $(shell node --version)"
	@echo "Git提交: $(shell git rev-parse --short HEAD)"
	@echo "构建时间: $(shell date)"

# 健康检查
health-check: ## Check application health
	@echo "🏥 健康检查..."
	@curl -f http://localhost:8080/api/health || echo "❌ 服务不可用"

# 备份
backup: ## Backup configuration and data
	@echo "💾 备份配置和数据..."
	./scripts/backup.sh

# 监控
monitor: ## Show system monitoring
	@echo "📊 系统监控..."
	@echo "CPU使用率:"
	@top -l 1 | grep "CPU usage" || echo "无法获取CPU信息"
	@echo ""
	@echo "内存使用:"
	@free -h || echo "无法获取内存信息"
	@echo ""
	@echo "磁盘使用:"
	@df -h || echo "无法获取磁盘信息"

# 快速启动
quick-start: install build ## Quick start (install deps and build)
	@echo "🎉 快速启动完成！"
	@echo "运行: make dev 启动开发服务器"
	@echo "或者: ./domain-max 启动应用"