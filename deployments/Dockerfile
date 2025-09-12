# 多阶段构建 Dockerfile

# 第一阶段：构建前端
FROM node:18-alpine AS web-builder

WORKDIR /app/web

# 复制前端依赖文件
COPY web/package*.json ./
RUN npm ci --only=production

# 复制前端源码并构建
COPY web/ ./
RUN npm run build

# 第二阶段：构建后端
FROM golang:1.23-alpine AS server-builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/

# 从第一阶段复制构建好的前端文件
COPY --from=web-builder /app/web/dist ./web/dist

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o domain-max ./cmd/server

# 第三阶段：运行时镜像
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates wget tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=server-builder /app/domain-max .

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 更改文件所有者
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# 启动应用
CMD ["./domain-max"]