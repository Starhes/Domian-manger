# Domain MAX 部署指南

本文档详细介绍了Domain MAX的各种部署方式，包括本地开发、Docker部署和生产环境部署。

## 📋 部署前准备

### 系统要求

**最低配置**
- CPU: 1核心
- 内存: 1GB RAM
- 存储: 10GB可用空间
- 操作系统: Linux/macOS/Windows

**推荐配置**
- CPU: 2核心以上
- 内存: 2GB RAM以上
- 存储: 20GB可用空间
- 操作系统: Ubuntu 20.04+ / CentOS 8+ / macOS 12+

### 依赖软件

**必需软件**
- Go 1.23+
- Node.js 18+
- PostgreSQL 12+ 或 MySQL 8.0+

**可选软件**
- Docker 20.10+
- Docker Compose 2.0+
- Nginx (反向代理)

## 🚀 快速部署 (Docker)

### 1. 下载项目

```bash
git clone <repository-url>
cd domain-max
```

### 2. 配置环境变量

```bash
cp configs/env.example .env
```

编辑 `.env` 文件，设置必要的配置：

```bash
# 数据库密码 (必需)
DB_PASSWORD=your_secure_password_here

# JWT密钥 (必需，生产环境建议64位以上)
JWT_SECRET=your_jwt_secret_key_here_at_least_64_characters_long

# 加密密钥 (必需，32字节十六进制)
ENCRYPTION_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef

# 可选：邮件配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=noreply@yourdomain.com
```

### 3. 启动服务

```bash
cd deployments
docker-compose up -d --build
```

### 4. 验证部署

```bash
# 检查服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app

# 健康检查
curl http://localhost:8080/api/health
```

### 5. 访问应用

- 应用地址: http://localhost:8080
- 默认管理员: admin@example.com / admin123

**⚠️ 重要：首次登录后请立即修改默认密码！**

## 🛠️ 本地开发部署

### 1. 环境准备

```bash
# 安装Go (如果未安装)
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 安装Node.js (如果未安装)
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 验证安装
go version
node --version
npm --version
```

### 2. 数据库准备

**PostgreSQL**
```bash
# 安装PostgreSQL
sudo apt-get install postgresql postgresql-contrib

# 创建数据库和用户
sudo -u postgres psql
CREATE DATABASE domain_manager;
CREATE USER domain_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE domain_manager TO domain_user;
\q
```

**MySQL**
```bash
# 安装MySQL
sudo apt-get install mysql-server

# 创建数据库和用户
sudo mysql
CREATE DATABASE domain_manager;
CREATE USER 'domain_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON domain_manager.* TO 'domain_user'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 3. 项目配置

```bash
# 克隆项目
git clone <repository-url>
cd domain-max

# 配置环境变量
cp configs/env.example .env
# 编辑 .env 文件

# 安装依赖
go mod tidy
cd web && npm install && cd ..
```

### 4. 构建和运行

```bash
# 使用构建脚本
./scripts/build.sh

# 或者手动构建
cd web && npm run build && cd ..
go build -o domain-max ./cmd/server

# 运行应用
./domain-max
```

## 🏭 生产环境部署

### 1. 服务器准备

```bash
# 更新系统
sudo apt-get update && sudo apt-get upgrade -y

# 安装必要软件
sudo apt-get install -y curl wget git nginx certbot python3-certbot-nginx

# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 2. 安全配置

```bash
# 配置防火墙
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable

# 创建应用用户
sudo useradd -m -s /bin/bash domain-max
sudo usermod -aG docker domain-max
```

### 3. 部署应用

```bash
# 切换到应用用户
sudo su - domain-max

# 克隆项目
git clone <repository-url>
cd domain-max

# 配置生产环境变量
cp configs/env.example .env
```

编辑生产环境配置：

```bash
# 生产环境配置
ENVIRONMENT=production
BASE_URL=https://yourdomain.com

# 强密码配置
DB_PASSWORD=<strong-random-password>
JWT_SECRET=<64-character-random-string>
ENCRYPTION_KEY=<32-byte-hex-string>

# 邮件配置
SMTP_HOST=smtp.yourdomain.com
SMTP_PORT=587
SMTP_USER=noreply@yourdomain.com
SMTP_PASSWORD=<smtp-password>
SMTP_FROM=noreply@yourdomain.com
```

```bash
# 启动服务
cd deployments
docker-compose up -d --build
```

### 4. 配置反向代理

创建Nginx配置文件：

```bash
sudo nano /etc/nginx/sites-available/domain-max
```

```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    
    # 重定向到HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;
    
    # SSL配置
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    
    # SSL安全配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # 安全头
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    # 代理到应用
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # 静态文件缓存
    location /static/ {
        proxy_pass http://localhost:8080;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

启用配置：

```bash
sudo ln -s /etc/nginx/sites-available/domain-max /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 5. 配置SSL证书

```bash
# 获取Let's Encrypt证书
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# 设置自动续期
sudo crontab -e
# 添加以下行
0 12 * * * /usr/bin/certbot renew --quiet
```

### 6. 配置监控

创建systemd服务文件：

```bash
sudo nano /etc/systemd/system/domain-max.service
```

```ini
[Unit]
Description=Domain MAX Application
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/home/domain-max/domain-max/deployments
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
User=domain-max
Group=domain-max

[Install]
WantedBy=multi-user.target
```

启用服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable domain-max
sudo systemctl start domain-max
```

## 📊 监控和维护

### 健康检查

```bash
# 检查应用状态
curl -f http://localhost:8080/api/health || echo "Service is down"

# 检查容器状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

### 备份策略

**数据库备份**
```bash
# PostgreSQL备份
docker-compose exec db pg_dump -U postgres domain_manager > backup_$(date +%Y%m%d_%H%M%S).sql

# MySQL备份
docker-compose exec db mysqldump -u root -p domain_manager > backup_$(date +%Y%m%d_%H%M%S).sql
```

**配置备份**
```bash
# 备份配置文件
tar -czf config_backup_$(date +%Y%m%d_%H%M%S).tar.gz .env configs/
```

### 更新部署

```bash
# 拉取最新代码
git pull origin main

# 重新构建和部署
docker-compose down
docker-compose up -d --build

# 验证更新
curl http://localhost:8080/api/health
```

### 性能优化

**数据库优化**
```sql
-- PostgreSQL优化
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
SELECT pg_reload_conf();
```

**应用优化**
```bash
# 调整Docker资源限制
# 在docker-compose.yml中添加：
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

## 🔧 故障排除

### 常见问题

**1. 数据库连接失败**
```bash
# 检查数据库状态
docker-compose logs db

# 检查网络连接
docker-compose exec app ping db
```

**2. 前端资源加载失败**
```bash
# 检查构建输出
ls -la web/dist/

# 重新构建前端
cd web && npm run build
```

**3. SSL证书问题**
```bash
# 检查证书状态
sudo certbot certificates

# 手动续期
sudo certbot renew --dry-run
```

### 日志分析

```bash
# 应用日志
docker-compose logs -f app

# 数据库日志
docker-compose logs -f db

# Nginx日志
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### 性能调试

```bash
# 检查资源使用
docker stats

# 检查数据库性能
docker-compose exec db psql -U postgres -d domain_manager -c "SELECT * FROM pg_stat_activity;"
```

## 📚 参考资料

- [Docker官方文档](https://docs.docker.com/)
- [PostgreSQL文档](https://www.postgresql.org/docs/)
- [Nginx配置指南](https://nginx.org/en/docs/)
- [Let's Encrypt文档](https://letsencrypt.org/docs/)

---

如有部署问题，请查看[故障排除指南](troubleshooting.md)或提交[Issue](../../issues)。