# Domain MAX - 部署与运维指南

本文档为 **Domain MAX** 系统提供了全面的部署、配置、运维及故障排查指导。

## 目录

- [环境准备](#-环境准备)
- [快速部署 (Docker Compose)](#-快速部署-docker-compose)
- [从源码构建与部署](#-从源码构建与部署)
- [生产环境最佳实践](#-生产环境最佳实践)
  - [使用 Nginx 进行反向代理](#1-使用-nginx-进行反向代理)
  - [配置 HTTPS](#2-配置-https)
  - [安全加固](#3-安全加固)
- [数据备份与恢复](#-数据备份与恢复)
- [系统监控与日志](#-系统监控与日志)
- [故障排查](#-故障排查)

---

## 📋 环境准备

在开始部署之前，请确保您的服务器满足以下条件：

- **操作系统**: 推荐使用主流 Linux 发行版 (如 Ubuntu 20.04+, CentOS 8+)。
- **硬件**: 至少 2GB RAM 和 10GB 磁盘空间。
- **软件**:
  - [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
  - [Docker](https://docs.docker.com/engine/install/) (v20.10+)
  - [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)

**Docker 与 Docker Compose 安装 (以 Ubuntu 为例):**

```bash
# 更新系统包
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg

# 添加 Docker 的官方 GPG 密钥
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# 设置 Docker 仓库
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装 Docker
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 启动并设置开机自启
sudo systemctl enable --now docker
```

## 🚀 快速部署 (Docker Compose)

这是最推荐的部署方式，适用于绝大多数场景。

1.  **克隆项目代码**

    ```bash
    git clone https://github.com/Domain-MAX/Domain-MAX.git
    cd Domain-MAX
    ```

2.  **创建并配置 `.env` 文件**

    ```bash
    cp env.example .env
    nano .env
    ```

    在编辑器中，**务必修改** `DB_PASSWORD` 和 `JWT_SECRET` 的值，并根据需要配置 `SMTP` 相关参数用于邮件发送。

3.  **启动服务**

    ```bash
    docker-compose up -d
    ```

    该命令会在后台构建并启动应用容器和数据库容器。

4.  **验证部署**
    - 访问 `http://<your-server-ip>:8080` 查看系统主页。
    - 默认管理员账户: `admin@example.com` / `admin123`

## 🏗️ 从源码构建与部署

如果您希望自行构建或对代码进行二次开发，可以按照以下步骤操作。

### 1. 构建前端

```bash
cd frontend
npm install
npm run build
```

构建产物将生成在 `frontend/dist` 目录下。

### 2. 构建后端

```bash
# 确保 Go 版本 >= 1.21
go mod tidy
go build -o domain-max-server main.go
```

这将生成一个名为 `domain-max-server` 的二进制可执行文件。

### 3. 运行

1.  **准备配置文件**: 将 `.env` 文件放置在 `domain-max-server` 同级目录下。
2.  **准备静态文件**: 将 `frontend/dist` 目录整体复制到 `domain-max-server` 同级目录下。
3.  **启动数据库**: 您需要自行准备一个 PostgreSQL 或 MySQL 数据库，并在 `.env` 中配置正确的连接信息。
4.  **启动服务**:
    ```bash
    ./domain-max-server
    ```

## 🛡️ 生产环境最佳实践

### 1. 使用 Nginx 进行反向代理

在生产环境中，强烈建议使用 Nginx 作为反向代理。这可以帮助您轻松实现 HTTPS、负载均衡和静态资源缓存。

**Nginx 配置示例 (`/etc/nginx/sites-available/domain-max.conf`):**

```nginx
server {
    listen 80;
    server_name your.domain.com; # 替换为您的域名

    # 将所有 HTTP 请求重定向到 HTTPS
    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name your.domain.com; # 替换为您的域名

    # SSL 证书配置 (请替换为您的证书路径)
    ssl_certificate /path/to/your/fullchain.pem;
    ssl_certificate_key /path/to/your/privkey.pem;

    # 推荐的 SSL 安全配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers off;
    ssl_ciphers "EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH";

    # 安全 Headers
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;

    location / {
        proxy_pass http://127.0.0.1:8080; # 代理到在本机运行的应用
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**启用配置:**

```bash
sudo ln -s /etc/nginx/sites-available/domain-max.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

### 2. 配置 HTTPS

推荐使用 [Let's Encrypt](https://letsencrypt.org/) 和 `certbot` 免费获取和自动续订 SSL 证书。

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your.domain.com
```

### 3. 安全加固

- **数据库**: 在 `.env` 中为数据库设置强密码。在生产环境中，不建议将数据库端口 `5432` 暴露到公网，`docker-compose.yml` 默认配置已遵循此实践。
- **防火墙 (UFW)**: 只开放必要的端口。
  ```bash
  sudo ufw allow ssh     # 22端口
  sudo ufw allow http    # 80端口
  sudo ufw allow https   # 443端口
  sudo ufw enable
  ```
- **定期更新**: 定期拉取最新的代码和基础镜像，并重新构建部署，以获取安全更新。
  ```bash
  git pull
  docker-compose pull
  docker-compose up -d --build
  ```

## 💾 数据备份与恢复

### 备份

使用 `docker-compose exec` 命令可以轻松备份 PostgreSQL 数据库。

```bash
# 创建一个存放备份的目录
mkdir -p backups

# 执行备份命令
docker-compose exec -T db pg_dump -U postgres domain_manager | gzip > backups/backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

建议使用 `cron` 设置定时任务，实现自动化备份。

### 恢复

```bash
# 停止应用服务以避免数据写入
docker-compose stop app

# 将备份文件恢复到数据库容器
gunzip < backups/your_backup_file.sql.gz | docker-compose exec -T db psql -U postgres -d domain_manager

# 重启应用服务
docker-compose start app
```

## 📊 系统监控与日志

### 查看日志

```bash
# 查看应用和数据库的实时日志
docker-compose logs -f

# 只查看应用服务的日志
docker-compose logs -f app
```

### 健康检查

系统提供了一个健康检查端点，可以用于监控服务的可用性。

- **URL**: `/api/health`
- **命令**: `curl http://localhost:8080/api/health`
- **成功响应**: `{"status":"ok","message":"服务运行正常"}`

## 🆘 故障排查

- **容器未启动**:
  - 运行 `docker-compose logs app` 查看应用日志，排查错误原因。
  - 检查 `.env` 文件中的配置项是否正确，特别是数据库密码。
- **数据库连接失败**:
  - 运行 `docker-compose logs db` 查看数据库日志。
  - 确保 `app` 容器和 `db` 容器在同一个 Docker 网络中。
- **Nginx 502 Bad Gateway**:
  - 检查应用服务是否正常运行 `docker-compose ps`。
  - 确认 Nginx 配置中的 `proxy_pass` 地址 (`127.0.0.1:8080`) 是否正确。
