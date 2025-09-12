## Domian-MAX（域名与 DNS 管理平台）

文档导航： [README](README.md) | [DEPLOYMENT](DEPLOYMENT.md) | [OPERATIONS](OPERATIONS.md)

一个开箱即用的全栈域名与 DNS 管理系统：后端基于 Go + Gin + Gorm，前端基于 React + Vite + Ant Design，内置用户注册登录、邮箱验证、密码找回、DNS 记录管理、DNS 服务商管理（示例：DNSPod），以及 SMTP 配置加密存储与测试。支持 Docker 一键构建与运行。

- 详细部署指南请见 [DEPLOYMENT.md](DEPLOYMENT.md)
- 使用与运维指南请见 [OPERATIONS.md](OPERATIONS.md)

### 特性

- 用户体系：注册/登录、邮箱验证、密码重置、CSRF 与速率限制
- DNS 管理：按用户的 A/AAAA/CNAME/TXT/MX 等记录的创建、更新、删除
- 管理后台：用户、域名、DNS 服务商、SMTP 配置、系统统计
- 邮件服务：支持从环境变量或数据库配置加载，SMTP 密码使用 AES 加密
- 安全基线：JWT+刷新令牌、Token 撤销、CSRF、防爆破、输入校验
- 交付方式：多阶段 Dockerfile，`docker-compose` 一键起服务

### 目录结构

```text
internal/            # 服务端代码
  api/               # Gin 路由与处理器
  services/          # 业务逻辑（auth/dns/admin/email/...）
  models/            # 数据模型
  middleware/        # 认证、权限、速率限制、CORS、CSRF 等
  database/          # 连接与迁移
  config/            # 配置加载与校验
frontend/            # 前端（React + Vite + AntD）
Dockerfile           # 多阶段构建（前端+后端）
docker-compose.yml   # 应用与 Postgres
init.sql             # 管理员与示例数据初始化
env.example          # 环境变量示例
```

### 快速开始（本地开发）

1. 准备环境变量

```bash
cp env.example .env
# 设置 DB_PASSWORD / JWT_SECRET / ENCRYPTION_KEY 等必填项
```

2. 启动 Postgres（可用 docker 或本地服务）。若使用 docker：

```bash
docker compose up -d db
```

3. 启动前端（可选，若走后端内置静态资源可跳过）：

```bash
cd frontend && npm ci && npm run dev
```

4. 启动后端：

```bash
go mod tidy
go run .
```

访问：`http://localhost:8080`（后端也会在生产构建后内置前端静态页），健康检查：`/api/health`。

更多部署方式（含 Docker 一体化构建）请阅读 [DEPLOYMENT.md](DEPLOYMENT.md)。

### 必要环境变量

请参考根目录 `env.example`。至少需要：

- DB_PASSWORD：数据库密码
- JWT_SECRET：JWT 签名密钥（生产建议 ≥64 位高强度）
- ENCRYPTION_KEY：AES 密钥（hex 32 bytes，用于加密 SMTP 密码）

可选：

- BASE_URL：对外访问地址（用于生成邮件链接，生产必须为域名）
- SMTP_HOST/PORT/USER/PASSWORD/SMTP_FROM：邮件服务（也可在后台配置）
- DNSPOD_TOKEN：若用 DNSPod，可先在后台录入正式配置

### 构建与运行（Docker 一体化）

```bash
docker compose up -d --build
```

构建会先打包前端并嵌入后端二进制中，容器启动后暴露 `:8080`。

更多生产部署、安全清单、反向代理示例见 [DEPLOYMENT.md](DEPLOYMENT.md)。

### 使用向导

- 首次启动后，数据库会通过 `init.sql` 初始化一个管理员：`admin@example.com / admin123`（请立刻修改）。
- 前台注册用户需邮箱验证；未配置 SMTP 时，系统会在控制台打印验证/重置链接用于开发调试。
- 管理端提供用户/域名/DNS/SMTP 等管理能力，详见 [OPERATIONS.md](OPERATIONS.md)。

### 测试文件清理

项目包含一些开发和测试辅助文件，生产部署前建议清理：

```bash
# 预览将要清理的测试文件
bash scripts/cleanup_test_files.sh --preview

# 备份重要配置后清理
bash scripts/cleanup_test_files.sh --backup
bash scripts/cleanup_test_files.sh

# 或直接强制清理
bash scripts/cleanup_test_files.sh --force
```

详细说明见 [docs/TEST_FILES_MANAGEMENT.md](docs/TEST_FILES_MANAGEMENT.md)。

### 开源许可

见 `LICENSE`。
