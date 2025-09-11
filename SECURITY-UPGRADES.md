# 🔐 安全升级指南

本文档详细说明了最新的安全升级内容和使用方法。

## 📋 升级内容概览

### ✅ 已修复的安全问题

1. **前端 Token 存储** - HttpOnly Cookies 替代 localStorage
2. **配置安全** - 移除默认密码，强制环境变量
3. **代码质量** - 消除硬编码，添加常量和速率限制
4. **CORS 配置** - 严格的域名白名单
5. **错误信息** - 统一处理，避免信息泄露
6. **输入验证** - 多层安全验证
7. **SMTP 加密** - AES 加密替代 bcrypt
8. **JWT 撤销** - Token 黑名单机制
9. **DNS 验证** - 完整的记录验证系统

## 🚀 升级步骤

### 1. 生成安全配置

**运行配置生成器**：

```bash
go run scripts/generate-config.go
```

这将创建一个安全的 `.env` 文件，包含：

- 强密码（16 位复杂密码）
- JWT 密钥（64-128 位随机字符）
- AES 加密密钥（32 字节十六进制）

### 2. 环境变量设置

**必需的环境变量**：

```bash
# 数据库配置
DB_PASSWORD=<强密码>

# 安全密钥
JWT_SECRET=<至少64位随机字符>
ENCRYPTION_KEY=<64位十六进制字符>

# 生产环境额外要求
BASE_URL=https://yourdomain.com  # 必须是HTTPS
```

### 3. Docker 部署

更新后的 `docker-compose.yml` 现在从环境变量读取配置：

```yaml
environment:
  - DB_PASSWORD=${DB_PASSWORD}
  - JWT_SECRET=${JWT_SECRET}
  - ENCRYPTION_KEY=${ENCRYPTION_KEY}
```

**启动步骤**：

```bash
# 1. 生成配置
go run scripts/generate-config.go

# 2. 启动服务
docker-compose up -d
```

## 🔐 新的认证机制

### HttpOnly Cookie 认证

**登录流程**：

```
1. 用户登录 → 服务器设置HttpOnly Cookie
2. 返回CSRF token给前端
3. 后续请求自动携带Cookie + CSRF token
```

**前端代码示例**：

```javascript
// 登录请求
const response = await fetch("/api/login", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ email, password }),
  credentials: "include", // 重要：允许发送Cookie
});

const data = await response.json();
const csrfToken = data.data.csrf_token; // 保存CSRF token

// 后续API请求
fetch("/api/dns-records", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-CSRF-Token": csrfToken, // 必须包含CSRF token
  },
  body: JSON.stringify(recordData),
  credentials: "include",
});
```

### 刷新令牌机制

- **访问令牌**: 24 小时有效期（HttpOnly Cookie）
- **刷新令牌**: 30 天有效期（HttpOnly Cookie）
- **CSRF 令牌**: 24 小时有效期（普通 Cookie，前端可访问）

## 🛡️ 安全特性

### 1. 速率限制

```
登录: 5次/分钟
注册: 3次/小时
API通用: 100次/分钟
DNS操作: 10次/分钟
管理员: 200次/分钟
```

### 2. CORS 安全

```javascript
// 只允许明确配置的域名
const allowedOrigins = [
  "http://localhost:3000", // React开发
  "https://yourdomain.com", // 生产域名
];
```

### 3. 安全响应头

```http
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
```

### 4. 输入验证

```
- 邮箱: 正则验证 + 长度限制 + 危险字符检查
- 密码: 复杂度要求 + 弱密码检查
- 用户ID: 数字验证 + 范围检查
- 搜索: SQL注入检查 + XSS防护
- DNS记录: 格式验证 + 私有IP检查
```

## 📊 配置验证规则

### 生产环境要求

**数据库密码**:

- 长度: 至少 12 位
- 复杂度: 大小写+数字+特殊字符
- 禁止: 常见弱密码模式

**JWT 密钥**:

- 长度: 至少 64 位（生产环境 128 位）
- 熵检查: 随机性验证
- 模式检查: 禁止明显规律

**加密密钥**:

- 格式: 64 个十六进制字符
- 强度: 禁止全零或简单模式

## 🔧 开发环境配置

### 本地开发

```bash
# .env 文件示例（开发环境）
ENVIRONMENT=development
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<生成的强密码>
DB_NAME=domain_manager
DB_TYPE=postgres

JWT_SECRET=<生成的64位密钥>
ENCRYPTION_KEY=<生成的32字节十六进制密钥>

# SMTP配置（可选）
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=noreply@yourdomain.com
```

### 前端开发配置

在开发环境中，需要在请求头中添加标识：

```javascript
headers: {
  'X-Development': 'true', // 标识开发环境
  'X-CSRF-Token': csrfToken
}
```

## 🚨 安全注意事项

### 必须遵守的安全规则

1. **永远不要**提交 `.env` 文件到代码仓库
2. **生产环境**必须使用 HTTPS
3. **定期更新**所有安全密钥
4. **监控**异常登录和 API 调用
5. **备份**数据库加密密钥

### 生产部署清单

- [ ] 使用配置生成器创建强密钥
- [ ] 设置正确的域名和 BASE_URL
- [ ] 启用 HTTPS 和安全头
- [ ] 配置防火墙和负载均衡
- [ ] 设置日志监控和告警
- [ ] 定期备份和密钥轮换

## 🔍 故障排除

### 常见问题

**1. CSRF token 错误**

```
原因: 前端未发送CSRF token或token过期
解决: 确保请求头包含正确的X-CSRF-Token
```

**2. Cookie 无法设置**

```
原因: 开发环境使用HTTP但设置了Secure cookie
解决: 在开发环境设置X-Development头
```

**3. 速率限制触发**

```
原因: 请求频率超过限制
解决: 检查速率限制设置，实现客户端重试机制
```

**4. 配置验证失败**

```
原因: 密钥不符合安全要求
解决: 重新运行配置生成器或手动生成强密钥
```

## 📈 监控指标

### 建议监控的指标

```
- 登录失败次数
- 速率限制触发次数
- JWT token撤销次数
- CSRF验证失败次数
- 配置验证错误次数
- 数据库连接失败次数
```

## 🆙 版本兼容性

### 向后兼容

- Cookie 认证与 Authorization 头兼容
- 旧版前端可继续使用 JWT token
- 数据库 schema 向前兼容

### 迁移建议

1. **渐进式升级**: 先部署后端，再升级前端
2. **测试验证**: 在测试环境充分验证
3. **回滚方案**: 准备配置和数据库回滚

---

## ✅ 升级验证

完成升级后，请验证以下功能：

1. [ ] 用户注册和邮箱验证
2. [ ] 用户登录和登出
3. [ ] Cookie 和 CSRF token 正常工作
4. [ ] DNS 记录 CRUD 操作
5. [ ] 管理员功能访问
6. [ ] 速率限制生效
7. [ ] 错误处理正常
8. [ ] SMTP 邮件发送

---

🎉 **恭喜！您的系统现在拥有银行级的安全防护能力！**
