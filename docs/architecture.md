# Domain MAX 架构设计

## 系统架构概览

Domain MAX 采用现代化的微服务架构设计，前后端分离，模块化开发，确保系统的可扩展性、可维护性和安全性。

## 技术栈

### 后端技术栈
- **语言**：Go 1.23+
- **Web框架**：Gin
- **ORM**：GORM
- **数据库**：PostgreSQL / MySQL
- **认证**：JWT + 刷新令牌
- **加密**：AES-256 + bcrypt
- **容器化**：Docker + Docker Compose

### 前端技术栈
- **语言**：TypeScript
- **框架**：React 18
- **构建工具**：Vite
- **UI库**：Ant Design
- **路由**：React Router
- **状态管理**：Zustand
- **HTTP客户端**：Axios

## 系统架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Browser   │    │  Mobile App     │    │  API Client     │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Load Balancer  │
                    │   (Nginx/ALB)   │
                    └─────────┬───────┘
                              │
                 ┌─────────────────────┐
                 │   Domain MAX App    │
                 │  (Go + Embedded     │
                 │   React Frontend)   │
                 └─────────┬───────────┘
                           │
              ┌─────────────────────────┐
              │     Database Layer      │
              │  PostgreSQL / MySQL     │
              └─────────────────────────┘
```

## 模块架构

### 后端模块设计

```
pkg/
├── auth/           # 认证授权模块
│   ├── models/     # 用户模型
│   ├── service/    # 认证服务
│   └── handlers/   # 认证处理器
├── dns/            # DNS管理模块
│   ├── models/     # DNS模型
│   ├── service/    # DNS服务
│   ├── providers/  # DNS服务商
│   └── handlers/   # DNS处理器
├── email/          # 邮件服务模块
│   ├── models/     # 邮件模型
│   ├── service/    # 邮件服务
│   └── templates/  # 邮件模板
├── admin/          # 管理模块
│   ├── service/    # 管理服务
│   └── handlers/   # 管理处理器
├── database/       # 数据库模块
│   ├── connection/ # 数据库连接
│   └── migration/  # 数据库迁移
├── config/         # 配置模块
├── middleware/     # 中间件
│   ├── auth/       # 认证中间件
│   ├── cors/       # CORS中间件
│   └── rate-limit/ # 限流中间件
└── utils/          # 工具函数
```

### 前端模块设计

```
web/src/
├── components/     # 通用组件
│   ├── layout/     # 布局组件
│   ├── forms/      # 表单组件
│   └── common/     # 通用组件
├── pages/          # 页面组件
│   ├── auth/       # 认证页面
│   ├── dns/        # DNS管理页面
│   └── admin/      # 管理页面
├── stores/         # 状态管理
│   ├── auth-store/ # 认证状态
│   └── dns-store/  # DNS状态
├── utils/          # 工具函数
│   ├── api/        # API客户端
│   ├── auth/       # 认证工具
│   └── validation/ # 验证工具
└── types/          # 类型定义
```

## 数据流架构

### 请求处理流程

```
Client Request
      ↓
  Load Balancer
      ↓
   Middleware Stack
   ├── CORS
   ├── Rate Limiting
   ├── Authentication
   └── Logging
      ↓
   Route Handler
      ↓
  Business Logic
      ↓
   Data Access Layer
      ↓
    Database
      ↓
   Response
```

### 认证流程

```
1. 用户登录
   ├── 验证邮箱密码
   ├── 生成JWT Token
   ├── 生成Refresh Token
   └── 返回Token

2. API请求
   ├── 提取JWT Token
   ├── 验证Token有效性
   ├── 提取用户信息
   └── 继续处理请求

3. Token刷新
   ├── 验证Refresh Token
   ├── 生成新JWT Token
   └── 返回新Token
```

## 安全架构

### 多层安全防护

1. **网络层安全**
   - HTTPS强制加密
   - 防火墙规则
   - DDoS防护

2. **应用层安全**
   - JWT认证
   - CSRF保护
   - 输入验证
   - SQL注入防护

3. **数据层安全**
   - 数据库访问控制
   - 敏感数据加密
   - 备份加密

### 权限控制模型

```
用户 (User)
├── 普通用户权限
│   ├── 查看自己的DNS记录
│   ├── 管理自己的DNS记录
│   └── 修改个人信息
└── 管理员权限
    ├── 用户管理
    ├── 域名管理
    ├── DNS服务商管理
    └── 系统配置管理
```

## 性能架构

### 缓存策略

1. **应用层缓存**
   - JWT Token缓存
   - 用户会话缓存
   - DNS记录缓存

2. **数据库优化**
   - 索引优化
   - 查询优化
   - 连接池管理

3. **前端优化**
   - 代码分割
   - 懒加载
   - 静态资源缓存

### 扩展性设计

1. **水平扩展**
   - 无状态应用设计
   - 负载均衡支持
   - 数据库读写分离

2. **垂直扩展**
   - 资源监控
   - 性能调优
   - 容量规划

## 部署架构

### 容器化部署

```
Docker Compose Stack
├── Application Container
│   ├── Go Binary
│   ├── Embedded Frontend
│   └── Health Check
├── Database Container
│   ├── PostgreSQL
│   ├── Data Volume
│   └── Backup Script
└── Reverse Proxy (Optional)
    ├── Nginx
    ├── SSL Termination
    └── Load Balancing
```

### 环境隔离

1. **开发环境**
   - 本地开发
   - 热重载
   - 详细日志

2. **测试环境**
   - 自动化测试
   - 性能测试
   - 安全测试

3. **生产环境**
   - 高可用部署
   - 监控告警
   - 备份恢复

## 监控架构

### 应用监控

1. **健康检查**
   - 服务状态检查
   - 数据库连接检查
   - 依赖服务检查

2. **性能监控**
   - 响应时间统计
   - 吞吐量监控
   - 错误率统计

3. **业务监控**
   - 用户行为分析
   - 功能使用统计
   - 异常操作监控

### 日志架构

```
Application Logs
├── Access Logs
├── Error Logs
├── Security Logs
└── Business Logs
      ↓
  Log Aggregation
      ↓
   Log Analysis
      ↓
  Alerting System
```

## 数据架构

### 数据模型设计

1. **用户数据**
   - 用户基本信息
   - 认证信息
   - 权限信息

2. **DNS数据**
   - 域名信息
   - DNS记录
   - 服务商配置

3. **系统数据**
   - 配置信息
   - 日志信息
   - 统计信息

### 数据一致性

1. **事务管理**
   - ACID特性保证
   - 分布式事务处理
   - 数据完整性约束

2. **数据同步**
   - 主从同步
   - 数据备份
   - 灾难恢复

## 总结

Domain MAX的架构设计遵循以下原则：

1. **模块化**：清晰的模块划分，低耦合高内聚
2. **安全性**：多层安全防护，数据加密保护
3. **可扩展性**：支持水平和垂直扩展
4. **可维护性**：标准化的代码结构和文档
5. **高性能**：优化的数据访问和缓存策略
6. **可观测性**：完善的监控和日志系统

这种架构设计确保了系统的稳定性、安全性和可扩展性，为用户提供可靠的域名和DNS管理服务。