# 项目结构说明

本文档详细说明了域名管理系统的项目结构和各个文件的作用。

## 📁 根目录结构

```
domain-manager/
├── 📁 frontend/                 # 前端项目目录
├── 📁 internal/                 # Go后端源码目录
├── 📄 main.go                  # Go应用入口文件
├── 📄 go.mod                   # Go模块定义文件
├── 📄 go.sum                   # Go依赖版本锁定文件
├── 📄 Dockerfile               # 多阶段Docker构建文件
├── 📄 docker-compose.yml       # Docker Compose配置文件
├── 📄 init.sql                 # 数据库初始化脚本
├── 📄 env.example              # 环境变量配置模板
├── 📄 .dockerignore            # Docker构建忽略文件
├── 📄 start.sh                 # 一键启动脚本
├── 📄 Makefile                 # 构建和管理命令
├── 📄 api-test.http            # API接口测试文件
├── 📄 README.md                # 项目说明文档
├── 📄 DEPLOYMENT.md            # 部署指南文档
└── 📄 PROJECT_STRUCTURE.md     # 本文件
```

## 🗂️ 后端结构 (internal/)

```
internal/
├── 📁 api/                     # API处理器层
│   ├── 📄 routes.go           # 路由配置
│   ├── 📄 auth.go             # 认证相关API
│   ├── 📄 dns.go              # DNS记录管理API
│   └── 📄 admin.go            # 管理员API
├── 📁 config/                  # 配置管理
│   └── 📄 config.go           # 配置结构和加载
├── 📁 database/                # 数据库层
│   └── 📄 database.go         # 数据库连接和迁移
├── 📁 middleware/              # 中间件
│   ├── 📄 cors.go             # CORS跨域处理
│   └── 📄 auth.go             # JWT认证中间件
├── 📁 models/                  # 数据模型
│   └── 📄 models.go           # 数据库模型和请求/响应结构
├── 📁 providers/               # DNS服务商接口
│   ├── 📄 interface.go        # DNS服务商接口定义
│   └── 📄 dnspod.go           # DNSPod服务商实现
└── 📁 services/                # 业务逻辑层
    ├── 📄 auth.go             # 认证业务逻辑
    ├── 📄 dns.go              # DNS记录业务逻辑
    └── 📄 admin.go            # 管理员业务逻辑
```

## 🎨 前端结构 (frontend/)

```
frontend/
├── 📁 public/                  # 静态资源
├── 📁 src/                     # 源码目录
│   ├── 📁 components/          # 公共组件
│   │   ├── 📄 Layout.tsx      # 用户端布局组件
│   │   └── 📄 AdminLayout.tsx # 管理端布局组件
│   ├── 📁 pages/               # 页面组件
│   │   ├── 📄 Login.tsx       # 登录页面
│   │   ├── 📄 Register.tsx    # 注册页面
│   │   ├── 📄 Dashboard.tsx   # 用户仪表盘
│   │   ├── 📄 DNSRecords.tsx  # DNS记录管理页面
│   │   ├── 📄 Profile.tsx     # 用户资料页面
│   │   └── 📁 admin/           # 管理员页面
│   │       ├── 📄 Dashboard.tsx   # 管理仪表盘
│   │       ├── 📄 Users.tsx       # 用户管理
│   │       ├── 📄 Domains.tsx     # 域名管理
│   │       └── 📄 Providers.tsx   # DNS服务商管理
│   ├── 📁 stores/              # 状态管理
│   │   └── 📄 authStore.ts    # 认证状态管理
│   ├── 📁 utils/               # 工具函数
│   │   └── 📄 api.ts          # API请求封装
│   ├── 📄 App.tsx             # 应用根组件
│   ├── 📄 main.tsx            # 应用入口
│   └── 📄 index.css           # 全局样式
├── 📄 package.json             # npm依赖配置
├── 📄 tsconfig.json            # TypeScript配置
├── 📄 vite.config.ts           # Vite构建配置
├── 📄 index.html               # HTML模板
└── 📄 .eslintrc.cjs            # ESLint配置
```

## 🔧 关键文件说明

### 后端关键文件

#### `main.go`

- 应用程序入口点
- 初始化配置、数据库连接
- 设置路由和中间件
- 嵌入前端静态文件

#### `internal/config/config.go`

- 环境变量配置管理
- 数据库、JWT、邮件、DNS 服务商配置

#### `internal/models/models.go`

- 数据库模型定义 (User, Domain, DNSRecord 等)
- API 请求/响应结构体定义

#### `internal/providers/dnspod.go`

- DNSPod API 集成实现
- DNS 记录的增删改查操作

### 前端关键文件

#### `src/App.tsx`

- React 应用根组件
- 路由配置和权限控制

#### `src/stores/authStore.ts`

- 用户认证状态管理
- 登录、登出、JWT token 管理

#### `src/utils/api.ts`

- HTTP 请求封装
- 请求/响应拦截器
- 错误处理

### 配置文件

#### `Dockerfile`

- 多阶段构建配置
- 前端构建 → 后端构建 → 最终镜像

#### `docker-compose.yml`

- 应用服务和数据库服务配置
- 网络和数据卷配置
- 环境变量映射

#### `init.sql`

- 数据库初始化脚本
- 默认管理员账户创建
- 示例数据插入

## 🚀 工作流程

### 1. 开发流程

```
1. 修改代码
2. 本地测试 (go run main.go 或 npm run dev)
3. 构建测试 (docker-compose build)
4. 提交代码
```

### 2. 部署流程

```
1. 拉取最新代码
2. 配置环境变量 (.env文件)
3. 执行部署 (docker-compose up -d)
4. 验证服务状态
```

### 3. 数据流程

```
前端页面 → API请求 → 路由处理 → 中间件验证 → 业务服务 → 数据库操作
                                                ↓
DNS服务商API ← 服务提供商接口 ← DNS业务逻辑 ←
```

## 📝 代码规范

### Go 代码规范

- 遵循 Go 官方代码风格
- 使用 `gofmt` 格式化代码
- 错误处理使用标准模式
- 包名使用小写字母

### TypeScript 代码规范

- 使用 ESLint + Prettier
- 组件名使用 PascalCase
- 文件名使用 camelCase
- 接口定义使用 interface

### 数据库规范

- 表名使用复数形式 (users, domains)
- 字段名使用下划线命名 (created_at)
- 外键使用 `表名_id` 格式
- 软删除使用 `deleted_at` 字段

## 🔐 安全考虑

### 后端安全

- JWT token 验证
- 密码 bcrypt 加密
- SQL 注入防护 (GORM ORM)
- CORS 跨域控制

### 前端安全

- XSS 防护 (React 自动转义)
- CSRF 防护 (JWT token)
- 路由权限控制
- API 请求拦截

### 部署安全

- Docker 非 root 用户运行
- 环境变量敏感信息隔离
- 数据库访问限制
- HTTPS 传输加密

---

通过这个项目结构说明，您可以快速了解整个系统的组织方式和各个组件的职责。如需了解具体实现细节，请查看相应的源码文件。
