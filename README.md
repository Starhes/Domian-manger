# 域名管理系统

一个现代化的二级域名分发管理系统，采用前后端分离架构，支持用户自助管理DNS记录，管理员统一管理域名资源和DNS服务商配置。

## 🚀 功能特性

### 用户端功能
- **用户注册登录**: 支持邮箱注册，邮件激活，忘记密码重置
- **DNS记录管理**: 支持 A、CNAME、TXT、MX 记录的增删改查
- **多域名支持**: 用户可在多个可用域名下创建子域名
- **实时同步**: DNS记录变更实时同步到DNS服务商

### 管理端功能
- **用户管理**: 查看、搜索、编辑、禁用用户账户
- **域名管理**: 添加、删除可供分发的主域名
- **DNS服务商管理**: 配置多个DNS服务商API凭证
- **系统统计**: 查看用户、域名、记录等统计信息

### 技术特性
- **单一容器部署**: 前后端打包为单一Docker镜像
- **前后端分离**: React + Go Gin 架构
- **JWT认证**: 无状态用户认证
- **数据库支持**: PostgreSQL / MySQL
- **API对接**: 首要支持DNSPod，架构可扩展

## 🛠️ 技术栈

### 后端
- **语言**: Go 1.21+
- **框架**: Gin
- **数据库**: PostgreSQL 15 / MySQL 8.0+
- **ORM**: GORM
- **认证**: JWT + bcrypt

### 前端
- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **UI库**: Ant Design
- **状态管理**: Zustand
- **HTTP客户端**: Axios

### 部署
- **容器化**: Docker + Docker Compose
- **多阶段构建**: 前端构建 → 后端构建 → 最终镜像
- **数据库**: PostgreSQL 15 Alpine

## 📦 快速开始

### 环境要求
- Docker 20.10+
- Docker Compose 2.0+

### 一键部署

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd domain-manager
   ```

2. **配置环境变量**
   ```bash
   cp env.example .env
   # 编辑 .env 文件，配置数据库密码、JWT密钥等
   ```

3. **启动服务**
   ```bash
   docker-compose up -d
   ```

4. **访问系统**
   - 用户端: http://localhost:8080
   - 管理后台: http://localhost:8080/admin
   - API文档: http://localhost:8080/api/health

### 默认管理员账户
- 邮箱: `admin@example.com`
- 密码: `admin123`

⚠️ **生产环境请立即修改默认密码！**

## ⚙️ 配置说明

### 环境变量配置

| 变量名 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `PORT` | 服务端口 | `8080` | 否 |
| `ENVIRONMENT` | 运行环境 | `development` | 否 |
| `DB_HOST` | 数据库主机 | `localhost` | 是 |
| `DB_PORT` | 数据库端口 | `5432` | 是 |
| `DB_USER` | 数据库用户名 | `postgres` | 是 |
| `DB_PASSWORD` | 数据库密码 | - | 是 |
| `DB_NAME` | 数据库名称 | `domain_manager` | 是 |
| `DB_TYPE` | 数据库类型 | `postgres` | 是 |
| `JWT_SECRET` | JWT密钥 | - | 是 |

### 邮件配置 (可选)
| 变量名 | 说明 | 示例 |
|--------|------|------|
| `SMTP_HOST` | SMTP服务器 | `smtp.gmail.com` |
| `SMTP_PORT` | SMTP端口 | `587` |
| `SMTP_USER` | 邮箱用户名 | `your@gmail.com` |
| `SMTP_PASSWORD` | 邮箱密码/应用密码 | `app_password` |
| `SMTP_FROM` | 发件人地址 | `noreply@yourdomain.com` |

### DNS服务商配置

系统支持通过管理后台配置DNS服务商，也可以通过环境变量预配置：

#### DNSPod配置
```bash
DNSPOD_TOKEN=your_dnspod_token_here
```

获取DNSPod Token:
1. 登录 [DNSPod控制台](https://console.dnspod.cn/)
2. 进入 "API密钥" 页面
3. 创建密钥，格式: `ID,Token`

## 🏗️ 开发指南

### 本地开发环境

1. **后端开发**
   ```bash
   # 安装Go依赖
   go mod tidy
   
   # 启动开发服务器
   go run main.go
   ```

2. **前端开发**
   ```bash
   cd frontend
   
   # 安装依赖
   npm install
   
   # 启动开发服务器
   npm run dev
   ```

### 项目结构
```
domain-manager/
├── main.go                 # 应用入口
├── go.mod                  # Go模块定义
├── Dockerfile              # 多阶段构建文件
├── docker-compose.yml      # 容器编排配置
├── init.sql               # 数据库初始化脚本
├── internal/              # 后端源码
│   ├── api/               # API处理器
│   ├── config/            # 配置管理
│   ├── database/          # 数据库连接
│   ├── middleware/        # 中间件
│   ├── models/            # 数据模型
│   ├── providers/         # DNS服务商接口
│   └── services/          # 业务逻辑
└── frontend/              # 前端源码
    ├── src/
    │   ├── components/    # React组件
    │   ├── pages/         # 页面组件
    │   ├── stores/        # 状态管理
    │   └── utils/         # 工具函数
    ├── package.json
    └── vite.config.ts
```

### API接口文档

#### 认证接口
- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录
- `GET /api/verify-email/:token` - 邮箱验证
- `POST /api/forgot-password` - 忘记密码
- `POST /api/reset-password` - 重置密码

#### DNS记录接口
- `GET /api/dns-records` - 获取用户DNS记录
- `POST /api/dns-records` - 创建DNS记录
- `PUT /api/dns-records/:id` - 更新DNS记录
- `DELETE /api/dns-records/:id` - 删除DNS记录

#### 管理员接口
- `GET /api/admin/users` - 获取用户列表
- `PUT /api/admin/users/:id` - 更新用户
- `GET /api/admin/domains` - 获取域名列表
- `POST /api/admin/domains` - 添加域名
- `GET /api/admin/dns-providers` - 获取DNS服务商列表

## 🔧 运维指南

### 数据备份
```bash
# 备份数据库
docker-compose exec db pg_dump -U postgres domain_manager > backup.sql

# 恢复数据库
docker-compose exec -T db psql -U postgres domain_manager < backup.sql
```

### 日志查看
```bash
# 查看应用日志
docker-compose logs -f app

# 查看数据库日志
docker-compose logs -f db
```

### 更新部署
```bash
# 重新构建并启动
docker-compose up -d --build

# 仅重启应用服务
docker-compose restart app
```

### 性能监控
- 应用健康检查: `GET /api/health`
- 容器资源监控: `docker stats`
- 数据库连接监控: 查看应用日志

## 🛡️ 安全建议

### 生产环境配置
1. **修改默认密码**: 立即修改管理员默认密码
2. **强化JWT密钥**: 使用复杂的JWT_SECRET
3. **数据库安全**: 使用强密码，限制网络访问
4. **HTTPS配置**: 配置反向代理启用HTTPS
5. **防火墙设置**: 只开放必要端口

### 定期维护
- 定期更新依赖包和基础镜像
- 监控系统资源使用情况
- 定期备份数据库数据
- 检查DNS服务商API配额使用情况

## 🤝 贡献指南

欢迎提交Issue和Pull Request来改进项目！

### 开发规范
- Go代码遵循gofmt格式化标准
- 前端代码使用ESLint + Prettier
- 提交信息使用约定式提交格式
- 添加必要的单元测试

## 📄 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 📞 支持

如果您在使用过程中遇到问题：

1. 查看本文档的常见问题部分
2. 搜索已有的 [Issues](../../issues)
3. 创建新的 Issue 描述问题
4. 联系开发团队

---

**🎉 感谢使用域名管理系统！**
