## 部署指南（DEPLOYMENT）

文档导航： [README](README.md) | [DEPLOYMENT](DEPLOYMENT.md) | [OPERATIONS](OPERATIONS.md)

本文档覆盖本地开发、Docker 一体化、生产部署、安全加固与常见问题。配合 [README.md](README.md) 与 [OPERATIONS.md](OPERATIONS.md) 联动使用。

### 1. 先决条件

- Go 1.21+（本地开发）
- Node 18+（本地前端开发）
- Docker 24+ 与 Docker Compose（推荐部署方式）
- Postgres 15（如使用外部数据库）

### 2. 配置环境变量

复制示例并填写必填项：

```bash
cp env.example .env
# 填写：DB_PASSWORD / JWT_SECRET / ENCRYPTION_KEY
# 可选：BASE_URL、SMTP_*、DNSPOD_TOKEN
```

关键校验（生产环境）：

- `JWT_SECRET` 建议 ≥64 位高强度字符串
- `ENCRYPTION_KEY` 必须为 32 字节的 hex（`openssl rand -hex 32`）
- `BASE_URL` 必须是对外可访问的域名（禁止 localhost）

更多校验在 `internal/config/config.go` 中已经强制执行（生产环境含安全词检查）。

### 3. 本地开发

选项 A：前后端分别起（前端热更新）

```bash
# 起数据库（也可用本地安装的 Postgres）
docker compose up -d db

# 前端
cd frontend && npm ci && npm run dev

# 后端（另开终端）
go mod tidy
go run .
```

访问 `http://localhost:5173`（前端开发服务器）或 `http://localhost:8080`（后端内置静态页，需先构建前端）。

选项 B：只起后端并使用后端静态资源

```bash
cd frontend && npm ci && npm run build
cd .. && go run .
```

### 4. Docker 一体化运行（推荐）

```bash
docker compose up -d --build
```

- 多阶段 `Dockerfile` 会先构建前端再编译后端，最终产物体积小、启动快
- 应用监听 `:8080`，健康检查端点：`/api/health`
- `docker-compose.yml` 中 `app` 服务会读取 `.env` 的关键变量

升级/重建：

```bash
docker compose pull
docker compose build --no-cache app
docker compose up -d
```

日志与排障：

```bash
docker compose logs -f app | cat
docker compose logs -f db | cat
```

### 5. 生产部署参考

#### 5.1 外部数据库

在 `docker-compose.yml` 中将 `db` 服务移除或注释，改为配置外部 Postgres 地址：

```yaml
environment:
  - DB_HOST=<your-rds-host>
  - DB_PORT=5432
  - DB_USER=postgres
  - DB_PASSWORD=${DB_PASSWORD}
  - DB_NAME=domain_manager
  - DB_TYPE=postgres
```

#### 5.2 反向代理（Nginx 示例）

确保已配置 `BASE_URL=https://your-domain.com`，并开放 8080 给内网。

```nginx
server {
  listen 80;
  server_name your-domain.com;
  location / {
    proxy_pass http://app:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Real-IP $remote_addr;
  }
}
```

生产建议使用 HTTPS 与 HTTP/2，或直接用云厂商的负载均衡/证书。

#### 5.3 邮件服务

两种方式：

- 环境变量：设置 `SMTP_HOST/PORT/USER/PASSWORD/SMTP_FROM`
- 管理后台：在“SMTP 配置”模块新增配置（密码将使用 `ENCRYPTION_KEY` 加密存储），然后“激活/设为默认/测试”

#### 5.4 DNS 服务商

- 示例为 `DNSPod`：在“DNS 服务商管理”添加类型与 JSON 配置（token 等）
- 后续用户创建记录时将调用该服务商 API，同步域名可用“域名同步”进行

### 6. 数据初始化

容器首次启动会通过 `init.sql` 插入：

- 管理员：`admin@example.com / admin123`（请立即修改密码）
- 示例域名：`example.com`, `test.com`
- 示例 DNSPod 配置（默认禁用）

### 7. 安全加固清单（生产必读）

- 强制设置：`DB_PASSWORD`、`JWT_SECRET`、`ENCRYPTION_KEY`、`BASE_URL`
- 检查后端环境：`ENVIRONMENT=production`
- 使用强随机 JWT/加密密钥，避免包含 `test/demo/example/default/localhost` 等字样
- 使用 HTTPS/反代，传递 `X-Forwarded-*` 请求头
- 关闭数据库对外暴露端口，仅内网访问
- 修改默认管理员密码，限制管理员数量
- 配置 CSRF 相关 Cookie 属性与同站策略（中间件已默认处理）
- 开启日志采集与告警，限制登录/注册/DNS 操作速率（已内置）

### 8. 备份与升级

- 数据在 Postgres 卷 `postgres_data`，请按需做周期快照
- 升级镜像时先备份 DB，再滚动更新 `app` 服务

### 9. 常见问题

- 无法发送邮件：确认已配置 SMTP 或在控制台查看开发模式下的“验证/重置链接”打印
- 登录 401：确认 Cookie 域与 `BASE_URL`、反向代理头一致，并携带 `X-CSRF-Token`
- DNS 无法创建：检查“DNS 服务商管理”是否已启用并填入正确凭据

更多使用操作细节见 [OPERATIONS.md](OPERATIONS.md)。
