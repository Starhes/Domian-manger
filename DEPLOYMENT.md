# 🚀 Domain-MAX 部署指南

本文档提供 Domain-MAX 域名管理系统的完整部署方案，包括 Docker 快速部署、源码构建、生产环境配置和运维管理。

> 📖 **文档导航**：[项目概述](./README.md) → **部署指南** → [操作手册](./OPERATIONS.md)

---

## 📋 目录导航

- [⚡ 快速开始](#-快速开始)
- [🐳 Docker 部署（推荐）](#-docker部署推荐)
- [🔧 源码部署](#-源码部署)
- [⚙️ 环境配置](#️-环境配置)
- [🏭 生产环境部署](#-生产环境部署)
- [🔍 健康检查与监控](#-健康检查与监控)
- [🛠️ 维护与运维](#️-维护与运维)
- [❌ 故障排除](#-故障排除)

---

## ⚡ 快速开始

### 最快 3 分钟部署体验版

```bash
# 1. 克隆项目
git clone https://github.com/Domain-MAX/Domain-MAX.git
cd Domain-MAX

# 2. 生成配置（自动生成安全密钥）
go run scripts/generate_config.go

# 3. 一键启动
docker-compose up -d

# 4. 查看启动状态
docker-compose logs -f
```

**访问地址**：

- 🌐 **用户门户**：http://localhost:8080
- 🛡️ **管理后台**：http://localhost:8080/admin
- 📊 **健康检查**：http://localhost:8080/api/health

**默认管理员账户**：

- 邮箱：`admin@example.com`
- 密码：`admin123`

> ⚠️ **安全提醒**：首次登录后请立即修改默认密码！操作步骤请参考 [操作手册 - 账户管理](./OPERATIONS.md#账户管理)

**🎉 部署完成！** 继续阅读 [操作手册](./OPERATIONS.md) 了解系统使用方法。

---

## 🐳 Docker 部署（推荐）

### 系统要求

| 组件               | 最低版本 | 推荐版本 | 说明         |
| ------------------ | -------- | -------- | ------------ |
| **Docker**         | 20.10+   | 24.0+    | 容器运行时   |
| **Docker Compose** | 1.29+    | 2.20+    | 容器编排工具 |
| **内存**           | 1GB      | 2GB+     | 系统运行内存 |
| **磁盘**           | 5GB      | 20GB+    | 数据存储空间 |
| **CPU**            | 1 核     | 2 核+    | 处理器要求   |

### 步骤 1：环境准备

```bash
# 安装Docker（Ubuntu/Debian）
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 验证安装
docker --version
docker-compose --version
```

### 步骤 2：项目配置

```bash
# 1. 克隆项目
git clone https://github.com/Domain-MAX/Domain-MAX.git
cd Domain-MAX

# 2. 生成安全配置（推荐）
go run scripts/generate_config.go
# 或手动复制配置模板
# cp env.example .env

# 3. 编辑配置文件
nano .env  # 或使用其他编辑器
```

### 步骤 3：启动服务

```bash
# 开发/测试环境
docker-compose up -d

# 生产环境（带健康检查和自动重启）
docker-compose -f docker-compose.yml up -d

# 查看启动日志
docker-compose logs -f

# 检查服务状态
docker-compose ps
```

### 步骤 4：验证部署

```bash
# 健康检查
curl http://localhost:8080/api/health

# 预期返回：
# {"status":"ok","message":"服务运行正常"}

# 检查前端页面
curl -s http://localhost:8080 | grep -o "<title>.*</title>"
```

---

## 🔧 源码部署

适用于开发环境、自定义部署或高度定制化需求。

### 环境要求

| 组件           | 版本要求 | 安装方式                            |
| -------------- | -------- | ----------------------------------- |
| **Go**         | 1.23+    | [官方下载](https://golang.org/dl/)  |
| **Node.js**    | 18.0+    | [官方下载](https://nodejs.org/)     |
| **PostgreSQL** | 13+      | [官方下载](https://postgresql.org/) |
| **或 MySQL**   | 8.0+     | [官方下载](https://mysql.com/)      |

### 后端部署

```bash
# 1. 准备Go环境
go version  # 确认版本 >= 1.23

# 2. 下载依赖
cd Domain-MAX
go mod download
go mod tidy

# 3. 配置环境变量
cp env.example .env
# 编辑.env文件，设置必要配置

# 4. 构建后端
go build -o domain-manager main.go

# 5. 运行后端服务
./domain-manager
# 或开发模式：go run main.go
```

### 前端部署

```bash
# 1. 安装Node.js依赖
cd frontend
npm install

# 2. 开发模式运行
npm run dev  # 启动开发服务器（端口5173）

# 3. 生产构建
npm run build  # 构建产物输出到 dist/

# 4. 验证构建
npm run preview  # 预览生产构建
```

### 数据库配置

#### PostgreSQL 设置

```bash
# 1. 安装PostgreSQL
sudo apt update && sudo apt install postgresql postgresql-contrib

# 2. 创建数据库和用户
sudo -u postgres psql
```

```sql
-- 在PostgreSQL中执行
CREATE DATABASE domain_manager;
CREATE USER domain_user WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE domain_manager TO domain_user;
ALTER USER domain_user CREATEDB;
\q
```

#### MySQL 设置

```bash
# 1. 安装MySQL
sudo apt update && sudo apt install mysql-server

# 2. 安全初始化
sudo mysql_secure_installation
```

```sql
-- 在MySQL中执行
   CREATE DATABASE domain_manager CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'domain_user'@'localhost' IDENTIFIED BY 'your_secure_password';
GRANT ALL PRIVILEGES ON domain_manager.* TO 'domain_user'@'localhost';
FLUSH PRIVILEGES;
```

---

## ⚙️ 环境配置

### 配置文件详解

**`.env` 配置文件结构：**

```bash
# ==================== 服务器配置 ====================
PORT=8080                    # 服务端口
ENVIRONMENT=production       # 环境模式: development/production
BASE_URL=https://domain.com  # 系统访问URL（生产环境必须设置）

# ==================== 数据库配置 ====================
DB_TYPE=postgres            # 数据库类型: postgres/mysql
DB_HOST=localhost           # 数据库主机
DB_PORT=5432               # 数据库端口（MySQL用3306）
DB_NAME=domain_manager     # 数据库名称
DB_USER=domain_user        # 数据库用户
DB_PASSWORD=               # ⚠️ 必须设置强密码

# ==================== 安全配置 ====================
JWT_SECRET=                # ⚠️ 必须设置（64位+随机字符）
ENCRYPTION_KEY=            # ⚠️ 必须设置（64个十六进制字符）

# ==================== 邮件配置 ====================
SMTP_HOST=smtp.gmail.com   # SMTP服务器
SMTP_PORT=587              # SMTP端口
SMTP_USER=                 # 邮箱账号
SMTP_PASSWORD=             # 邮箱密码/应用密码
SMTP_FROM=noreply@domain.com  # 发件人地址

# ==================== DNS服务商配置 ====================
DNSPOD_TOKEN=              # DNSPod Token（格式：ID,Token）
```

### 自动配置生成

**使用配置生成器**（推荐）：

```bash
# 交互式配置生成
go run scripts/generate_config.go

# 自动生成模式（用于CI/CD）
go run scripts/generate_config.go --auto

# 验证配置文件
go run scripts/generate_config.go --validate
```

**手动配置**：

```bash
# 1. 复制配置模板
cp env.example .env

# 2. 生成随机密钥
# JWT密钥（64位字符）
openssl rand -base64 64 | tr -d "=+/" | cut -c1-64

# AES加密密钥（64个十六进制字符）
openssl rand -hex 32

# 数据库密码（16位随机密码）
openssl rand -base64 16 | tr -d "=+/"
```

---

## 🏭 生产环境部署

### 生产级 Docker 部署

**生产环境 docker-compose.yml 示例：**

```yaml
version: "3.8"

services:
  # 应用服务
  app:
    image: domain-max:latest
    restart: always
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - BASE_URL=https://yourdomain.com
      - DB_HOST=db
      # 从外部文件加载敏感配置
    env_file:
      - .env.production
    depends_on:
      - db
      - redis
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      replicas: 2 # 高可用部署
      resources:
        limits:
          memory: 512M
          cpus: "0.5"

  # 数据库服务
  db:
    image: postgres:15-alpine
    restart: always
    environment:
      POSTGRES_DB: domain_manager
      POSTGRES_USER: postgres
    env_file:
      - .env.production
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/
      - ./backups:/backups
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis缓存（可选）
  redis:
    image: redis:7-alpine
    restart: always
    networks:
      - app-network
    volumes:
      - redis_data:/data

  # 反向代理
  nginx:
    image: nginx:alpine
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    networks:
      - app-network

networks:
  app-network:
    driver: overlay # Swarm模式网络

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local
```

### Nginx 反向代理配置

创建 `nginx.conf`：

```nginx
events {
    worker_connections 1024;
}

http {
    upstream domain_manager {
        server app:8080;
    }

    # HTTPS重定向
server {
    listen 80;
        server_name yourdomain.com;
        return 301 https://$server_name$request_uri;
}

    # 主服务
server {
    listen 443 ssl http2;
        server_name yourdomain.com;

        # SSL配置
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;

        # 安全头
    add_header X-Content-Type-Options nosniff;
        add_header X-Frame-Options DENY;
    add_header X-XSS-Protection "1; mode=block";
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";

        # 代理配置
    location / {
            proxy_pass http://domain_manager;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

            # 超时配置
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }

        # 静态资源优化
        location /static/ {
            proxy_pass http://domain_manager;
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }
}
```

### 生产环境启动

```bash
# 1. 生产环境配置
cp env.example .env.production
# 编辑生产配置

# 2. 构建镜像
docker-compose build --no-cache

# 3. 启动服务
docker-compose -f docker-compose.yml up -d

# 4. 验证部署
make verify-deployment
```

---

## 🔧 源码部署

### 系统依赖安装

#### Ubuntu/Debian

```bash
# 1. 更新系统
sudo apt update && sudo apt upgrade -y

# 2. 安装Go环境
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc
source ~/.bashrc

# 3. 安装Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 4. 安装PostgreSQL
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### CentOS/RHEL

```bash
# 1. 安装Go环境
sudo yum install -y wget
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bash_profile
source ~/.bash_profile

# 2. 安装Node.js
curl -sL https://rpm.nodesource.com/setup_18.x | sudo bash -
sudo yum install -y nodejs

# 3. 安装PostgreSQL
sudo yum install -y postgresql-server postgresql-contrib
sudo postgresql-setup initdb
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### 构建部署

```bash
# 1. 克隆项目
git clone https://github.com/Domain-MAX/Domain-MAX.git
cd Domain-MAX

# 2. 后端构建
go mod download
go build -o domain-manager main.go

# 3. 前端构建
cd frontend
npm ci --only=production
npm run build
cd ..

# 4. 配置环境变量
cp env.example .env
# 编辑配置文件

# 5. 数据库初始化
psql -U postgres -d domain_manager -f init.sql

# 6. 启动服务
./domain-manager
```

### 系统服务配置

创建 systemd 服务文件 `/etc/systemd/system/domain-manager.service`：

```ini
[Unit]
Description=Domain-MAX Domain Management Service
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=domain-manager
Group=domain-manager
WorkingDirectory=/opt/domain-manager
Environment=GIN_MODE=release
EnvironmentFile=/opt/domain-manager/.env
ExecStart=/opt/domain-manager/domain-manager
ExecReload=/bin/kill -s HUP $MAINPID
KillMode=mixed
Restart=always
RestartSec=5

# 安全设置
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/opt/domain-manager/logs

[Install]
WantedBy=multi-user.target
```

```bash
# 启用服务
sudo systemctl daemon-reload
sudo systemctl enable domain-manager
sudo systemctl start domain-manager

# 检查状态
sudo systemctl status domain-manager
```

---

## ⚙️ 环境配置

### 核心配置项

#### 🔒 安全配置（必须设置）

| 配置项           | 要求              | 生成方法                                | 说明           |
| ---------------- | ----------------- | --------------------------------------- | -------------- |
| `DB_PASSWORD`    | 12 位+强密码      | `openssl rand -base64 16`               | 数据库连接密码 |
| `JWT_SECRET`     | 64 位+随机字符    | `openssl rand -base64 64 \| cut -c1-64` | JWT 签名密钥   |
| `ENCRYPTION_KEY` | 64 个十六进制字符 | `openssl rand -hex 32`                  | AES 加密密钥   |

#### 📧 邮件配置（可选但推荐）

```bash
# Gmail示例配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password  # 使用应用专用密码
SMTP_FROM=noreply@yourdomain.com

# QQ邮箱示例
SMTP_HOST=smtp.qq.com
SMTP_PORT=587
SMTP_USER=your-email@qq.com
SMTP_PASSWORD=your-authorization-code

# 企业邮箱示例
SMTP_HOST=smtp.exmail.qq.com
SMTP_PORT=465
SMTP_USER=admin@yourdomain.com
SMTP_PASSWORD=your-password
```

#### 🌐 DNS 服务商配置

**DNSPod 传统 API**：

```bash
# 获取Token：https://console.dnspod.cn/account/token
DNSPOD_TOKEN=123456,your_token_here
```

**腾讯云 DNSPod API v3**：

```json
{
  "secret_id": "AKIDxxxxxxxxxxxxxxx",
  "secret_key": "xxxxxxxxxxxxxxx",
  "region": "ap-guangzhou"
}
```

### 配置验证

```bash
# 1. 使用配置验证工具
go run scripts/generate_config.go --validate

# 2. 测试数据库连接
docker-compose exec app go run -c "
package main
import \"domain-manager/internal/database\"
import \"domain-manager/internal/config\"
func main() {
    cfg := config.Load()
    _, err := database.Connect(cfg)
    if err != nil {
        panic(err)
    }
    println(\"数据库连接成功\")
}
"

# 3. 测试邮件配置
curl -X POST http://localhost:8080/api/admin/smtp-configs/1/test \
  -H "Content-Type: application/json" \
  -d '{"to_email":"test@example.com"}'
```

---

## 🏭 生产环境部署

### 生产环境清单

#### 部署前检查

- [ ] **安全配置**：所有密钥已设置为强随机值
- [ ] **数据库**：已配置备份策略和访问控制
- [ ] **HTTPS**：已配置 SSL 证书和强制 HTTPS
- [ ] **防火墙**：已配置必要端口的访问控制
- [ ] **监控**：已配置日志收集和告警机制
- [ ] **备份**：已设置自动备份策略

#### 生产环境优化

```yaml
# docker-compose.production.yml
version: "3.8"

services:
  app:
    image: domain-max:latest
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
        order: start-first
      restart_policy:
        condition: any
        delay: 5s
        max_attempts: 3
    environment:
      - GIN_MODE=release
      - ENVIRONMENT=production
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "3"

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_SHARED_PRELOAD_LIBRARIES=pg_stat_statements
    command: |
      postgres 
      -c max_connections=200
      -c shared_buffers=256MB
      -c effective_cache_size=1GB
      -c log_statement=all

  # 添加监控服务
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
```

### 高可用部署

#### Docker Swarm 集群

```bash
# 1. 初始化Swarm集群
docker swarm init

# 2. 部署服务栈
docker stack deploy -c docker-compose.production.yml domain-max

# 3. 查看服务状态
docker service ls
docker stack ps domain-max
```

#### 负载均衡配置

```nginx
# nginx upstream配置
upstream domain_manager_backend {
    least_conn;
    server app1:8080 max_fails=3 fail_timeout=30s;
    server app2:8080 max_fails=3 fail_timeout=30s;
    server app3:8080 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    # SSL和安全配置...

    location / {
        proxy_pass http://domain_manager_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";

        # 健康检查
        proxy_next_upstream error timeout http_500 http_502 http_503 http_504;
        proxy_next_upstream_tries 3;
        proxy_next_upstream_timeout 10s;
    }
}
```

---

## 🔍 健康检查与监控

### 内置健康检查

```bash
# 基础健康检查
curl http://localhost:8080/api/health
# 返回: {"status":"ok","message":"服务运行正常"}

# 数据库连接检查
curl http://localhost:8080/api/health/db
# 返回数据库连接状态

# DNS服务商连接检查
curl http://localhost:8080/api/health/providers
# 返回DNS服务商连接状态
```

### 日志管理

```bash
# Docker日志查看
docker-compose logs -f app
docker-compose logs -f db

# 应用日志位置
tail -f logs/application.log
tail -f logs/error.log
tail -f logs/security.log

# 日志轮转配置（/etc/logrotate.d/domain-manager）
/opt/domain-manager/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    copytruncate
    notifempty
}
```

### 性能监控

```bash
# 系统资源监控
docker stats

# 应用性能指标
curl http://localhost:8080/api/metrics

# 数据库性能
docker-compose exec db psql -U postgres -c "
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY total_time DESC LIMIT 10;"
```

---

## 🛠️ 维护与运维

### 日常维护命令

```bash
# 查看服务状态
make status

# 查看实时日志
make logs

# 重启服务
make restart

# 更新应用
make update

# 备份数据
make backup

# 恢复数据
make restore BACKUP=backup_20250911_120000.sql.gz
```

### 更新部署

```bash
# 1. 拉取最新代码
git pull origin main

# 2. 构建新镜像
docker-compose build --no-cache

# 3. 滚动更新（零停机）
docker-compose up -d --force-recreate --no-deps app

# 4. 验证更新
curl http://localhost:8080/api/health
```

### 数据备份策略

#### 自动备份脚本

创建 `scripts/backup.sh`：

```bash
#!/bin/bash

BACKUP_DIR="/opt/backups/domain-manager"
DATE=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/backup_$DATE.sql.gz"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 数据库备份
docker-compose exec -T db pg_dump -U postgres domain_manager | gzip > $BACKUP_FILE

# 配置文件备份
cp .env $BACKUP_DIR/.env_$DATE

# 清理旧备份（保留30天）
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete
find $BACKUP_DIR -name ".env_*" -mtime +30 -delete

echo "备份完成: $BACKUP_FILE"
```

#### 定时备份配置

```bash
# 添加到crontab
crontab -e

# 每日凌晨2点自动备份
0 2 * * * /opt/domain-manager/scripts/backup.sh >> /var/log/domain-manager-backup.log 2>&1
```

---

## ❌ 故障排除

### 常见问题与解决方案

#### 🔥 服务无法启动

**问题症状**：

```bash
docker-compose logs app
# 输出：数据库连接失败: dial tcp 127.0.0.1:5432: connect: connection refused
```

**解决方案**：

```bash
# 1. 检查数据库服务状态
docker-compose ps db

# 2. 检查网络连接
docker-compose exec app ping db

# 3. 验证数据库配置
docker-compose exec app env | grep DB_

# 4. 重启数据库服务
docker-compose restart db

# 5. 检查数据库日志
docker-compose logs db
```

#### 🔐 JWT 密钥错误

**问题症状**：

```
配置验证失败: JWT_SECRET 不能为空
```

**解决方案**：

```bash
# 1. 生成新的JWT密钥
openssl rand -base64 64 | tr -d "=+/" | cut -c1-64

# 2. 更新.env文件
echo "JWT_SECRET=your_generated_secret_here" >> .env

# 3. 重启应用
docker-compose restart app
```

#### 📧 邮件发送失败

**问题症状**：

```
SMTP连接失败: dial tcp smtp.gmail.com:587: i/o timeout
```

**解决方案**：

```bash
# 1. 检查SMTP配置
curl -X POST http://localhost:8080/api/admin/smtp-configs/1/test \
  -H "Content-Type: application/json" \
  -d '{"to_email":"test@example.com"}'

# 2. 验证网络连接
docker-compose exec app ping smtp.gmail.com

# 3. 检查防火墙设置
sudo ufw status
# 确保出站端口587开放

# 4. 测试手动SMTP连接
telnet smtp.gmail.com 587
```

#### 🌐 DNS 同步失败

**问题症状**：

```
DNSPod API错误 [-1]: 登录失败，请检查Token是否正确
```

**解决方案**：

```bash
# 1. 验证DNSPod Token格式
echo $DNSPOD_TOKEN
# 格式应为：123456,your_token_here

# 2. 测试API连接
curl -X POST "https://dnsapi.cn/Domain.List" \
  -d "login_token=$DNSPOD_TOKEN&format=json"

# 3. 检查DNSPod控制台
# 访问：https://console.dnspod.cn/account/token
# 验证Token状态和权限
```

### 性能优化

#### 数据库优化

```sql
-- PostgreSQL性能优化
-- 1. 添加索引
CREATE INDEX CONCURRENTLY idx_dns_records_user_domain ON dns_records(user_id, domain_id);
CREATE INDEX CONCURRENTLY idx_dns_records_subdomain ON dns_records(subdomain);

-- 2. 分析查询性能
EXPLAIN ANALYZE SELECT * FROM dns_records WHERE user_id = 1;

-- 3. 更新统计信息
ANALYZE dns_records;
```

#### 应用优化

```bash
# 1. Go应用性能分析
go tool pprof http://localhost:8080/debug/pprof/profile

# 2. 内存使用分析
go tool pprof http://localhost:8080/debug/pprof/heap

# 3. 并发性能测试
ab -n 1000 -c 10 http://localhost:8080/api/health
```

### 安全强化

```bash
# 1. 更新系统补丁
sudo apt update && sudo apt upgrade -y

# 2. 配置防火墙
sudo ufw enable
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp  # SSH

# 3. SSL证书配置（Let's Encrypt）
sudo apt install certbot
sudo certbot certonly --standalone -d yourdomain.com

# 4. 设置定期证书更新
echo "0 2 * * * /usr/bin/certbot renew --quiet" | sudo crontab -
```

### 灾难恢复

#### 完整系统恢复流程

```bash
# 1. 准备新服务器环境
# 按照本文档进行基础环境配置

# 2. 恢复配置文件
scp backup-server:/opt/backups/domain-manager/.env_20250911_120000 ./.env

# 3. 恢复数据库
gunzip -c backup_20250911_120000.sql.gz | docker-compose exec -T db psql -U postgres domain_manager

# 4. 启动服务
docker-compose up -d

# 5. 验证恢复
curl http://localhost:8080/api/health
```

---

## 📞 技术支持

### 支持渠道

- **📋 问题反馈**：[GitHub Issues](https://github.com/Domain-MAX/Domain-MAX/issues)
- **💬 功能讨论**：[GitHub Discussions](https://github.com/Domain-MAX/Domain-MAX/discussions)
- **💬 实时交流**：[Discord 社区](https://discord.gg/n4AdZGwy5K) - 快速获得技术支持
- **📚 项目文档**：查看项目 README 和相关文档

### 问题报告模板

提交问题时，请提供以下信息：

```markdown
**环境信息：**

- 操作系统：Ubuntu 20.04 LTS
- Docker 版本：24.0.5
- 部署方式：Docker Compose

**问题描述：**
[详细描述遇到的问题]

**复现步骤：**

1. [步骤 1]
2. [步骤 2]
3. [步骤 3]

**预期行为：**
[描述期望的正确行为]

**实际行为：**
[描述实际发生的情况]

**错误日志：**
```

[粘贴相关错误日志]

```

**相关配置：**
[贴出相关配置信息，注意隐藏敏感信息]
```

---

## 🎯 下一步

部署完成后，建议按以下顺序进行系统配置：

### 🔐 首要安全配置

1. **修改默认密码**：[操作手册 - 账户管理](./OPERATIONS.md#账户管理)
2. **配置 HTTPS**：[本文档 - 生产环境部署](#-生产环境部署)
3. **安全强化**：[本文档 - 安全强化](#安全强化)

### ⚙️ 系统基础配置

4. **邮件服务配置**：[操作手册 - SMTP 邮件配置](./OPERATIONS.md#smtp邮件配置)
5. **DNS 服务商配置**：[操作手册 - DNS 服务商配置](./OPERATIONS.md#dns服务商配置)
6. **域名资源配置**：[操作手册 - 域名资源管理](./OPERATIONS.md#域名资源管理)

### 👥 用户和权限管理

7. **创建普通用户**：[操作手册 - 用户管理](./OPERATIONS.md#用户管理)
8. **设置 DNS 配额**：[操作手册 - 用户管理](./OPERATIONS.md#用户管理)
9. **权限分配**：[操作手册 - 高级功能](./OPERATIONS.md#-高级功能)

### 📊 监控和维护

10. **设置监控**：[本文档 - 健康检查与监控](#-健康检查与监控)
11. **配置备份**：[本文档 - 维护与运维](#️-维护与运维)
12. **制定维护计划**：[操作手册 - 系统维护指南](./OPERATIONS.md#-系统维护指南)

---

<div align="center">

**🚀 现在开始使用 Domain-MAX！**

[返回项目首页](./README.md) | [查看操作手册](./OPERATIONS.md) | [参与贡献](./CONTRIBUTING.md)

</div>
