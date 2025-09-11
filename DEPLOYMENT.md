# 🚀 Domain MAX 部署指南

> **全面的部署与运维文档** - 从零基础到生产环境的完整指导

本文档提供了 **Domain MAX 二级域名分发管理系统** 的完整部署方案，包括快速入门、生产部署、运维管理等各个环节。

## 📋 目录导航

- [🎯 快速开始](#-快速开始) - 5 分钟快速体验
- [🏗️ 部署方式选择](#️-部署方式选择) - 选择适合你的部署方案
- [🐳 Docker 部署（推荐）](#-docker-部署推荐) - 生产环境首选
- [💻 源码部署](#-源码部署) - 开发和定制化需求
- [🏢 生产环境配置](#-生产环境配置) - 安全加固和性能优化
- [🛠️ 运维管理](#️-运维管理) - 监控、备份、升级
- [🆘 故障排查](#-故障排查) - 常见问题解决
- [📚 相关文档](#-相关文档) - 其他重要文档链接

---

## 🎯 快速开始

**只需 3 步，5 分钟内启动系统！**

### 前提条件

- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [Docker Compose](https://docs.docker.com/compose/install/) 2.0+

### 一键部署

```bash
# 1. 克隆项目
git clone https://github.com/Domain-MAX/Domain-MAX.git
cd Domain-MAX

# 2. 生成安全配置（自动生成强密码和密钥）
go run scripts/generate-config.go

# 3. 启动服务
docker-compose up -d
```

### 访问系统

- **用户门户**: http://localhost:8080
- **管理后台**: http://localhost:8080/admin
- **默认管理员**: `admin@example.com` / `admin123`

⚠️ **首次登录后请立即修改管理员密码！**

---

## 🏗️ 部署方式选择

根据您的需求选择合适的部署方案：

| 部署方式                            | 适用场景           | 难度     | 推荐指数   |
| ----------------------------------- | ------------------ | -------- | ---------- |
| [Docker Compose](#-docker-部署推荐) | 生产环境、快速部署 | ⭐⭐     | ⭐⭐⭐⭐⭐ |
| [源码部署](#-源码部署)              | 开发环境、定制需求 | ⭐⭐⭐   | ⭐⭐⭐     |
| [Kubernetes](#kubernetes-部署)      | 大规模集群部署     | ⭐⭐⭐⭐ | ⭐⭐⭐⭐   |

---

## 🐳 Docker 部署（推荐）

### 环境准备

**系统要求**:

- 操作系统: Linux (Ubuntu 20.04+/CentOS 8+) 推荐
- 内存: 最低 2GB，推荐 4GB+
- 存储: 最低 10GB，推荐 50GB+
- CPU: 最低 1 核，推荐 2 核+

**安装 Docker (Ubuntu 示例)**:

```bash
# 快速安装脚本
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
sudo systemctl enable --now docker

# 安装 Docker Compose
sudo apt-get install docker-compose-plugin
```

### 配置部署

1. **获取项目代码**

   ```bash
   git clone https://github.com/Domain-MAX/Domain-MAX.git
   cd Domain-MAX
   ```

2. **生成安全配置**

   ```bash
   # 自动生成包含强密码和随机密钥的 .env 文件
   go run scripts/generate-config.go

   # 或手动配置（高级用户）
   cp env.example .env
   nano .env  # 编辑配置文件
   ```

3. **关键配置项说明**

   ```bash
   # === 必须配置项 ===
   DB_PASSWORD=<16位强密码>        # 数据库密码
   JWT_SECRET=<64位随机字符串>      # JWT 签名密钥
   ENCRYPTION_KEY=<32字节hex>      # AES 加密密钥

   # === 生产环境额外配置 ===
   ENVIRONMENT=production          # 环境标识
   BASE_URL=https://your-domain.com # 系统访问域名

   # === 邮件服务配置（可选）===
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=your-email@gmail.com
   SMTP_PASSWORD=your-app-password
   SMTP_FROM=noreply@your-domain.com
   ```

4. **启动服务**

   ```bash
   # 后台启动
   docker-compose up -d

   # 查看启动日志
   docker-compose logs -f

   # 检查服务状态
   docker-compose ps
   ```

### 验证部署

```bash
# 检查服务健康状态
curl http://localhost:8080/api/health

# 预期响应
{"status":"ok","message":"服务运行正常"}
```

### 首次配置

1. 登录管理后台: `http://your-domain:8080/admin`
2. 使用默认账户: `admin@example.com` / `admin123`
3. **立即修改管理员密码**
4. 配置 DNS 服务商（目前支持 DNSPod）
5. 添加主域名资源
6. 系统即可正常使用

---

## 💻 源码部署

适用于开发环境或需要定制化的场景。

### 环境要求

- **Go**: 1.21+ ([安装指南](https://golang.org/doc/install))
- **Node.js**: 18+ ([安装指南](https://nodejs.org/))
- **数据库**: PostgreSQL 13+ 或 MySQL 8.0+

### 部署步骤

1. **准备数据库**

   ```sql
   -- PostgreSQL
   CREATE DATABASE domain_manager;

   -- MySQL
   CREATE DATABASE domain_manager CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

2. **获取代码并配置**

   ```bash
   git clone https://github.com/Domain-MAX/Domain-MAX.git
   cd Domain-MAX

   # 生成配置文件
   go run scripts/generate-config.go

   # 根据实际环境调整 .env
   nano .env
   ```

3. **构建前端**

   ```bash
   cd frontend
   npm install
   npm run build
   cd ..
   ```

4. **启动后端**

   ```bash
   # 安装依赖
   go mod tidy

   # 开发模式
   go run main.go

   # 或构建后运行
   go build -o domain-max main.go
   ./domain-max
   ```

### 开发环境

开发时需要前后端分离运行：

```bash
# 终端1: 启动后端 (8080端口)
go run main.go

# 终端2: 启动前端开发服务器 (5173端口)
cd frontend
npm run dev
```

访问 http://localhost:5173 即可开发调试。

---

## 🏢 生产环境配置

### 1. 反向代理配置

**使用 Nginx（推荐）**:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL 证书配置
    ssl_certificate /path/to/fullchain.pem;
    ssl_certificate_key /path/to/privkey.pem;

    # 安全头设置
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";

    # 反向代理
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 支持（如需要）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

**自动获取 SSL 证书**:

```bash
# 安装 certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d your-domain.com

# 自动续期（添加到 crontab）
0 12 * * * /usr/bin/certbot renew --quiet
```

### 2. 安全加固

**防火墙配置**:

```bash
# UFW 配置示例
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 'Nginx Full'
sudo ufw enable
```

**Docker 安全**:

```yaml
# docker-compose.override.yml 生产环境配置
version: "3.8"

services:
  app:
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: "1.0"

  db:
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    ports: [] # 移除端口暴露，仅内网访问
```

### 3. 性能优化

**数据库优化**:

```sql
-- PostgreSQL 性能配置建议
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
SELECT pg_reload_conf();
```

**应用层优化**:

```bash
# .env 生产环境配置
ENVIRONMENT=production
GIN_MODE=release

# 数据库连接池
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5m
```

---

## 🛠️ 运维管理

### 1. 日常运维命令

```bash
# 查看服务状态
docker-compose ps

# 查看实时日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 更新服务
git pull
docker-compose pull
docker-compose up -d --build

# 进入容器调试
docker-compose exec app sh
docker-compose exec db psql -U postgres domain_manager
```

### 2. 数据备份

**自动备份脚本**:

```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/path/to/backups"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# 数据库备份
docker-compose exec -T db pg_dump -U postgres domain_manager | gzip > "$BACKUP_DIR/db_backup_$DATE.sql.gz"

# 配置文件备份
cp .env "$BACKUP_DIR/env_backup_$DATE"

# 清理旧备份（保留7天）
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "备份完成: $DATE"
```

**设置定时备份**:

```bash
# 添加到 crontab
crontab -e

# 每天凌晨2点备份
0 2 * * * /path/to/backup.sh >> /var/log/domain-max-backup.log 2>&1
```

**恢复数据**:

```bash
# 停止应用
docker-compose stop app

# 恢复数据库
gunzip -c backup_file.sql.gz | docker-compose exec -T db psql -U postgres domain_manager

# 重启服务
docker-compose start app
```

### 3. 监控告警

**健康检查监控**:

```bash
#!/bin/bash
# health_check.sh
HEALTH_URL="http://localhost:8080/api/health"
WEBHOOK_URL="your_alert_webhook_url"

if ! curl -s --max-time 10 $HEALTH_URL | grep -q '"status":"ok"'; then
    # 发送告警通知
    curl -X POST $WEBHOOK_URL -H 'Content-Type: application/json' \
         -d '{"text": "Domain-MAX 服务异常，请检查！"}'
    exit 1
fi

echo "健康检查正常"
```

**系统资源监控**:

```bash
# 查看资源使用情况
docker stats

# 查看磁盘使用
df -h
docker system df

# 查看日志大小
du -sh /var/lib/docker/containers/*/
```

### 4. 版本升级

**安全升级流程**:

```bash
# 1. 备份数据
./backup.sh

# 2. 拉取新版本
git fetch --tags
git checkout v1.1.0  # 替换为目标版本

# 3. 检查配置变更
diff .env env.example

# 4. 构建并测试
docker-compose build
docker-compose -f docker-compose.test.yml up

# 5. 生产环境部署
docker-compose up -d --build

# 6. 验证升级结果
curl http://localhost:8080/api/health
```

---

## 🆘 故障排查

### 常见问题速查

| 问题症状         | 可能原因               | 解决方法                       |
| ---------------- | ---------------------- | ------------------------------ |
| 应用无法启动     | 配置文件错误           | 检查 `.env` 文件，运行配置验证 |
| 数据库连接失败   | 数据库未启动或配置错误 | 检查数据库容器状态和连接参数   |
| 前端页面空白     | 构建失败或静态资源问题 | 重新构建前端，检查构建日志     |
| DNS 记录同步失败 | API 凭证错误           | 检查 DNS 服务商配置            |
| 邮件发送失败     | SMTP 配置错误          | 验证 SMTP 服务器设置           |

### 详细诊断步骤

**1. 应用启动问题**

```bash
# 检查容器状态
docker-compose ps

# 查看启动日志
docker-compose logs app

# 检查配置文件
go run scripts/generate-config.go --validate
```

**2. 数据库问题**

```bash
# 检查数据库容器
docker-compose logs db

# 手动连接测试
docker-compose exec db psql -U postgres domain_manager

# 检查数据库连接
docker-compose exec app wget -qO- http://localhost:8080/api/health
```

**3. 网络问题**

```bash
# 检查端口占用
netstat -tlnp | grep 8080

# 检查防火墙
sudo ufw status

# 测试内部连接
docker-compose exec app ping db
```

**4. 性能问题**

```bash
# 查看资源使用
docker stats

# 检查数据库性能
docker-compose exec db pg_stat_activity

# 分析慢查询
docker-compose exec db pg_stat_statements
```

### 紧急恢复程序

```bash
# 快速回滚
git checkout HEAD~1
docker-compose down
docker-compose up -d --build

# 从备份恢复
docker-compose down
docker volume rm domain-max_postgres_data
gunzip -c latest_backup.sql.gz | docker-compose exec -T db psql -U postgres domain_manager
docker-compose up -d
```

---

## 📚 相关文档

- **[📖 用户操作手册](./OPERATIONS.md)** - 详细的功能使用指南
- **[🔒 安全升级指南](./SECURITY-UPGRADES.md)** - 安全特性和升级说明
- **[🛠️ 开发指南](./DEVELOPMENT.md)** - 开发环境搭建和代码贡献
- **[🐛 问题反馈](https://github.com/Domain-MAX/Domain-MAX/issues)** - Bug 报告和功能请求

---

## 🤝 获得帮助

- **文档问题**: 查看 [相关文档](#-相关文档) 获取更详细信息
- **技术问题**: 提交 [Issue](https://github.com/Domain-MAX/Domain-MAX/issues)
- **功能建议**: 参与 [Discussions](https://github.com/Domain-MAX/Domain-MAX/discussions)

---

**🎉 完成部署后，您就拥有了一个功能完整、安全可靠的域名管理系统！**
