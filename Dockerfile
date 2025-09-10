# 多阶段构建Dockerfile

# 第一阶段：构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端依赖文件
COPY frontend/package.json ./

# 安装前端依赖（包括构建工具）
RUN npm install

# 复制前端源码
COPY frontend/ ./

# 构建前端（显示详细输出）
RUN npm run build --verbose

# 第二阶段：构建后端
FROM golang:1.21-alpine AS backend-builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 复制Go模块文件
COPY go.mod ./

# 如果go.sum存在则复制，否则跳过
COPY go.su[m] ./

# 下载依赖并生成go.sum
RUN go mod download

# 复制后端源码
COPY . .

# 从第一阶段复制前端构建产物
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# 构建后端应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main .

# 第三阶段：最终镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 创建非root用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/main .

# 设置文件权限
RUN chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# 启动应用
CMD ["./main"]
