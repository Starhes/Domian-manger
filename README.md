# Domain MAX - 域名与DNS管理平台

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![Node Version](https://img.shields.io/badge/Node-18+-green.svg)](https://nodejs.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

一个现代化的全栈域名与DNS管理系统，采用Go后端 + React前端架构，提供完整的用户管理、DNS记录管理、邮件服务和管理后台功能。

## ✨ 特性

### 🔐 用户系统
- 用户注册/登录，支持邮箱验证
- 密码重置功能
- JWT认证 + 刷新令牌机制
- CSRF保护和速率限制
- 管理员权限控制

### 🌐 DNS管理
- 支持多种DNS记录类型（A/AAAA/CNAME/TXT/MX/NS/PTR/SRV/CAA）
- 批量操作和导入导出
- 实时验证和安全检查
- 多DNS服务商支持（DNSPod等）

### 📧 邮件服务
- 灵活的SMTP配置管理
- 密码AES加密存储
- 邮件模板系统
- 发送状态监控

### 🛡️ 安全特性
- 输入验证和SQL注入防护
- 密码强度检查
- 配置安全验证
- 生产环境安全检查

### 🚀 部署方式
- Docker一键部署
- 多阶段构建优化
- 健康检查和监控
- 环境配置管理

## 📁 项目结构

```
domain-max/
├── cmd/server/              # 应用入口
│   └── main.go
├── pkg/                     # 核心包
│   ├── auth/               # 认证模块
│   │   └── models/
│   ├── dns/                # DNS管理模块
│   │   ├── models/
│   │   └── providers/
│   ├── email/              # 邮件模块
│   │   └── models/
│   ├── admin/              # 管理模块
│   ├── database/           # 数据库模块
│   ├── config/             # 配置模块
│   ├── middleware/         # 中间件
│   └── utils/              # 工具函数
├── web/                    # 前端应用
│   ├── src/
│   │   ├── components/     # 组件
│   │   ├── pages/          # 页面
│   │   ├── stores/         # 状态管理
│   │   ├── utils/          # 工具函数
│   │   └── types/          # 类型定义
│   ├── public/             # 静态资源
│   └── dist/               # 构建输出
├── configs/                # 配置文件
│   ├── env.example         # 环境变量示例
│   └── init.sql            # 数据库初始化
├── deployments/            # 部署配置
│   ├── Dockerfile          # Docker构建文件
│   └── docker-compose.yml  # 容器编排
├── scripts/                # 构建脚本
│   └── build.sh            # 构建脚本
└── docs/                   # 项目文档
```

## 🚀 快速开始

### 环境要求

- Go 1.23+
- Node.js 18+
- PostgreSQL 12+ 或 MySQL 8.0+
- Docker & Docker Compose（可选）

### 本地开发

1. **克隆项目**
```bash
git clone <repository-url>
cd domain-max
```

2. **配置环境变量**
```bash
cp configs/env.example .env
# 编辑 .env 文件，设置必要的配置项
```

3. **启动数据库**
```bash
# 使用Docker启动PostgreSQL
docker run -d \
  --name domain-max-db \
  -e POSTGRES_DB=domain_manager \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=your_password \
  -p 5432:5432 \
  postgres:15-alpine
```

4. **构建并运行**
```bash
# 构建前端和后端
./scripts/build.sh

# 运行应用
./domain-max
```

5. **访问应用**
- 前端界面：http://localhost:8080
- API文档：http://localhost:8080/api/health
- 默认管理员：admin@example.com / admin123

### Docker部署

1. **准备环境变量**
```bash
cp configs/env.example .env
# 编辑 .env 文件
```

2. **一键启动**
```bash
cd deployments
docker-compose up -d --build
```

## 🔧 配置说明

### 必需配置

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `DB_PASSWORD` | 数据库密码 | `your_secure_password` |
| `JWT_SECRET` | JWT签名密钥 | `64位随机字符串` |
| `ENCRYPTION_KEY` | AES加密密钥 | `32字节十六进制字符串` |

### 可选配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `PORT` | 服务端口 | `8080` |
| `ENVIRONMENT` | 运行环境 | `development` |
| `BASE_URL` | 系统基础URL | 自动检测 |
| `SMTP_*` | 邮件服务配置 | 可在后台配置 |

## 🛠️ 开发指南

### 代码规范

- **Go代码**：遵循Go官方代码规范，使用`gofmt`格式化
- **TypeScript代码**：使用ESLint + Prettier，遵循Airbnb规范
- **命名规范**：统一使用小写字母+连字符格式（kebab-case）
- **提交规范**：使用Conventional Commits格式

### 目录规范

- `pkg/`：按功能域组织，每个模块独立
- `web/src/`：前端代码按类型和功能组织
- `configs/`：所有配置文件统一存放
- `deployments/`：部署相关文件
- `docs/`：项目文档

### 安全最佳实践

1. **密码安全**：使用bcrypt哈希，强制密码复杂度
2. **数据验证**：所有输入进行严格验证
3. **权限控制**：基于角色的访问控制
4. **加密存储**：敏感数据AES加密存储
5. **安全头**：设置适当的HTTP安全头

## 📚 API文档

### 认证接口

- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/refresh` - 刷新令牌
- `POST /api/auth/logout` - 用户登出

### DNS管理接口

- `GET /api/dns/records` - 获取DNS记录列表
- `POST /api/dns/records` - 创建DNS记录
- `PUT /api/dns/records/:id` - 更新DNS记录
- `DELETE /api/dns/records/:id` - 删除DNS记录

### 管理接口

- `GET /api/admin/users` - 用户管理
- `GET /api/admin/domains` - 域名管理
- `GET /api/admin/providers` - DNS服务商管理
- `GET /api/admin/smtp-configs` - SMTP配置管理

## 🔍 监控和日志

### 健康检查

- 端点：`GET /api/health`
- 检查项：数据库连接、服务状态

### 日志级别

- **开发环境**：INFO级别，详细SQL日志
- **生产环境**：ERROR级别，关键错误日志

### 性能监控

- 请求响应时间统计
- 数据库连接池监控
- 内存使用情况跟踪

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'feat: add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

- 📖 [项目文档](docs/)
- 🐛 [问题反馈](../../issues)
- 💬 [讨论区](../../discussions)

## 🎯 路线图

- [ ] 支持更多DNS服务商
- [ ] 添加DNS记录模板功能
- [ ] 实现API限流和配额管理
- [ ] 添加操作审计日志
- [ ] 支持多语言国际化
- [ ] 移动端适配优化

---

**Domain MAX** - 让域名管理更简单、更安全、更高效！