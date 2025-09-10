# 部署指南

本文档详细介绍了域名管理系统的部署方法和配置选项。

## 📋 部署前准备

### 系统要求
- **操作系统**: Linux (推荐 Ubuntu 20.04+, CentOS 8+)
- **内存**: 最低 2GB，推荐 4GB+
- **存储**: 最低 10GB 可用空间
- **网络**: 需要访问外网进行DNS API调用

### 软件依赖
- Docker 20.10+
- Docker Compose 2.0+
- Git (用于克隆代码)

### 安装Docker (Ubuntu)
```bash
# 更新包索引
sudo apt update

# 安装依赖
sudo apt install apt-transport-https ca-certificates curl gnupg lsb-release

# 添加Docker官方GPG密钥
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# 添加Docker仓库
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装Docker
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 添加用户到docker组
sudo usermod -aG docker $USER
```

## 🚀 标准部署

### 1. 获取源码
```bash
git clone <repository-url>
cd domain-manager
```

### 2. 环境配置
```bash
# 复制环境变量模板
cp env.example .env

# 编辑配置文件
nano .env
```

### 3. 关键配置项
```bash
# 数据库配置 - 必须修改
DB_PASSWORD=your_very_secure_password_here

# JWT密钥 - 必须修改
JWT_SECRET=your_jwt_secret_key_at_least_32_characters_long

# 邮件配置 - 可选但推荐
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=noreply@yourdomain.com
```

### 4. 启动服务
```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 5. 验证部署
```bash
# 检查健康状态
curl http://localhost:8080/api/health

# 预期响应
{"status":"ok","message":"服务运行正常"}
```

## 🔧 生产环境部署

### 1. 反向代理配置 (Nginx)

创建 Nginx 配置文件 `/etc/nginx/sites-available/domain-manager`:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL证书配置
    ssl_certificate /path/to/your/certificate.crt;
    ssl_certificate_key /path/to/your/private.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # 代理到应用
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 静态文件缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        proxy_pass http://127.0.0.1:8080;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

启用配置:
```bash
sudo ln -s /etc/nginx/sites-available/domain-manager /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 2. 防火墙配置
```bash
# 允许HTTP和HTTPS
sudo ufw allow 80
sudo ufw allow 443

# 如果需要直接访问应用端口 (不推荐生产环境)
sudo ufw allow 8080

# 启用防火墙
sudo ufw enable
```

### 3. 生产环境Docker Compose

创建 `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "127.0.0.1:8080:8080"  # 只绑定到本地
    environment:
      - PORT=8080
      - ENVIRONMENT=production
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=domain_manager
      - DB_TYPE=postgres
      - JWT_SECRET=${JWT_SECRET}
      - SMTP_HOST=${SMTP_HOST}
      - SMTP_PORT=${SMTP_PORT}
      - SMTP_USER=${SMTP_USER}
      - SMTP_PASSWORD=${SMTP_PASSWORD}
      - SMTP_FROM=${SMTP_FROM}
    depends_on:
      - db
    restart: unless-stopped
    networks:
      - domain-manager-network
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=domain_manager
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql:ro
      - ./backups:/backups  # 备份目录
    restart: unless-stopped
    networks:
      - domain-manager-network
    deploy:
      resources:
        limits:
          memory: 1G
        reservations:
          memory: 512M

networks:
  domain-manager-network:
    driver: bridge

volumes:
  postgres_data:
    driver: local
```

启动生产环境:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## 🔐 安全加固

### 1. 系统安全
```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装fail2ban
sudo apt install fail2ban

# 配置SSH (如果使用)
sudo nano /etc/ssh/sshd_config
# 设置: PermitRootLogin no, PasswordAuthentication no

# 重启SSH服务
sudo systemctl restart sshd
```

### 2. Docker安全
```bash
# 限制Docker daemon访问
sudo chmod 660 /var/run/docker.sock

# 使用非root用户运行容器 (已在Dockerfile中配置)

# 定期更新镜像
docker-compose pull
docker-compose up -d
```

### 3. 数据库安全
```bash
# 进入数据库容器
docker-compose exec db psql -U postgres domain_manager

# 创建应用专用数据库用户
CREATE USER app_user WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE domain_manager TO app_user;
GRANT USAGE ON SCHEMA public TO app_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO app_user;
```

## 📊 监控和日志

### 1. 日志管理
```bash
# 配置日志轮转
sudo nano /etc/docker/daemon.json
```

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

```bash
sudo systemctl restart docker
```

### 2. 监控脚本

创建 `monitor.sh`:
```bash
#!/bin/bash

# 检查服务状态
check_service() {
    if docker-compose ps | grep -q "Up"; then
        echo "✅ 服务运行正常"
    else
        echo "❌ 服务异常"
        docker-compose ps
    fi
}

# 检查健康状态
check_health() {
    response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health)
    if [ "$response" = "200" ]; then
        echo "✅ 应用健康检查通过"
    else
        echo "❌ 应用健康检查失败: $response"
    fi
}

# 检查磁盘空间
check_disk() {
    usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ "$usage" -lt 80 ]; then
        echo "✅ 磁盘空间充足: ${usage}%"
    else
        echo "⚠️  磁盘空间不足: ${usage}%"
    fi
}

echo "=== 系统监控报告 $(date) ==="
check_service
check_health
check_disk
echo "================================"
```

设置定时检查:
```bash
chmod +x monitor.sh

# 添加到crontab (每5分钟检查一次)
echo "*/5 * * * * /path/to/monitor.sh >> /var/log/domain-manager-monitor.log 2>&1" | crontab -
```

## 💾 备份和恢复

### 1. 自动备份脚本

创建 `backup.sh`:
```bash
#!/bin/bash

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
DB_CONTAINER="domain-manager_db_1"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 数据库备份
docker exec $DB_CONTAINER pg_dump -U postgres domain_manager | gzip > $BACKUP_DIR/db_backup_$DATE.sql.gz

# 保留最近7天的备份
find $BACKUP_DIR -name "db_backup_*.sql.gz" -mtime +7 -delete

echo "备份完成: db_backup_$DATE.sql.gz"
```

设置每日备份:
```bash
chmod +x backup.sh
echo "0 2 * * * /path/to/backup.sh" | crontab -
```

### 2. 恢复数据
```bash
# 停止应用服务
docker-compose stop app

# 恢复数据库
gunzip -c /backups/db_backup_YYYYMMDD_HHMMSS.sql.gz | docker-compose exec -T db psql -U postgres domain_manager

# 重启服务
docker-compose up -d
```

## 🔄 更新部署

### 1. 应用更新
```bash
# 拉取最新代码
git pull origin main

# 重新构建并部署
docker-compose up -d --build

# 清理无用镜像
docker image prune -f
```

### 2. 零停机更新 (使用多实例)

创建 `docker-compose.ha.yml` 支持多实例:
```yaml
version: '3.8'

services:
  app1:
    # ... 配置同上
    ports:
      - "8081:8080"
  
  app2:
    # ... 配置同上  
    ports:
      - "8082:8080"
      
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - app1
      - app2
```

## 🆘 故障排查

### 常见问题

1. **容器启动失败**
   ```bash
   # 查看详细日志
   docker-compose logs app
   
   # 检查配置文件
   docker-compose config
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker-compose exec db pg_isready -U postgres
   
   # 查看数据库日志
   docker-compose logs db
   ```

3. **DNS API调用失败**
   ```bash
   # 检查网络连接
   docker-compose exec app ping dnsapi.cn
   
   # 验证API凭证
   # 登录管理后台检查DNS服务商配置
   ```

4. **内存不足**
   ```bash
   # 查看容器资源使用
   docker stats
   
   # 增加swap空间
   sudo fallocate -l 2G /swapfile
   sudo chmod 600 /swapfile
   sudo mkswap /swapfile
   sudo swapon /swapfile
   ```

### 紧急恢复

如果系统完全不可用:
```bash
# 1. 停止所有服务
docker-compose down

# 2. 备份当前数据
docker run --rm -v domain-manager_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/emergency_backup.tar.gz /data

# 3. 重新部署
docker-compose up -d --force-recreate

# 4. 如需恢复数据
docker run --rm -v domain-manager_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/emergency_backup.tar.gz -C /
```

---

通过以上配置，您就可以在生产环境中安全、稳定地运行域名管理系统了。如有任何问题，请参考故障排查部分或联系技术支持。
