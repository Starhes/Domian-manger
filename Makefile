# Makefile for Domain Manager

.PHONY: help build run dev clean test docker-build docker-run docker-clean

# 默认目标
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  dev          - Run in development mode"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-clean - Clean Docker resources"

# 构建应用
build:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build
	@echo "Building backend..."
	go mod tidy
	go build -o main .

# 运行应用
run:
	@echo "Starting application..."
	./main

# 开发模式
dev:
	@echo "Starting in development mode..."
	go run main.go

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	rm -f main
	rm -rf frontend/dist
	rm -rf frontend/node_modules

# 运行测试
test:
	@echo "Running Go tests..."
	go test ./...
	@echo "Running frontend tests..."
	cd frontend && npm test

# Docker构建
docker-build:
	@echo "Building Docker image..."
	docker-compose build

# Docker运行
docker-run:
	@echo "Starting with Docker Compose..."
	docker-compose up -d

# Docker清理
docker-clean:
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f

# 生产部署
deploy:
	@echo "Deploying to production..."
	docker-compose -f docker-compose.yml up -d --build

# 查看日志
logs:
	docker-compose logs -f

# 备份数据库
backup:
	@echo "Creating database backup..."
	mkdir -p backups
	docker-compose exec db pg_dump -U postgres domain_manager | gzip > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql.gz

# 恢复数据库
restore:
	@echo "Please specify backup file: make restore-from BACKUP=backup_file.sql.gz"

restore-from:
	@if [ -z "$(BACKUP)" ]; then echo "Please specify BACKUP file"; exit 1; fi
	@echo "Restoring from $(BACKUP)..."
	gunzip -c backups/$(BACKUP) | docker-compose exec -T db psql -U postgres domain_manager
