# 📋 Domain-MAX 操作手册

本手册详细介绍 Domain-MAX 域名管理系统的功能使用方法，包括用户端和管理员端的完整操作指南。

> 📖 **文档导航**：[项目概述](./README.md) → [部署指南](./DEPLOYMENT.md) → **操作手册**

---

## 📑 目录导航

- [🚀 快速入门](#-快速入门)
- [👤 用户功能](#-用户功能)
  - [账户管理](#账户管理)
  - [DNS 记录管理](#dns记录管理)
  - [个人设置](#个人设置)
- [🛡️ 管理员功能](#️-管理员功能)
  - [系统概览](#系统概览)
  - [用户管理](#用户管理)
  - [域名资源管理](#域名资源管理)
  - [DNS 服务商配置](#dns服务商配置)
  - [SMTP 邮件配置](#smtp邮件配置)
- [🔧 高级功能](#-高级功能)
- [❓ 常见问题](#-常见问题)
- [🆘 故障排除](#-故障排除)

---

## 🚀 快速入门

### 系统访问

**部署完成后的访问地址**：

| 功能模块     | 访问地址                           | 说明                     |
| ------------ | ---------------------------------- | ------------------------ |
| **用户门户** | `http://localhost:8080`            | 普通用户的主要操作界面   |
| **管理后台** | `http://localhost:8080/admin`      | 管理员专用的系统管理界面 |
| **API 接口** | `http://localhost:8080/api`        | RESTful API 接口         |
| **健康检查** | `http://localhost:8080/api/health` | 系统状态监控端点         |

> 💡 **部署提醒**：如果系统尚未部署，请先参考 [部署指南](./DEPLOYMENT.md) 完成环境搭建。

### 默认账户信息

**管理员账户**：

- **邮箱**：`admin@example.com`
- **密码**：`admin123`
- **权限**：系统管理员，拥有所有功能权限

> ⚠️ **重要安全提醒**：首次登录后请立即修改默认密码！

### 基本使用流程

```mermaid
graph LR
    A[访问系统] --> B{已有账户?}
    B -->|否| C[注册账户]
    B -->|是| D[登录系统]
    C --> E[邮箱验证]
    E --> D
    D --> F[选择功能]
    F --> G[DNS记录管理]
    F --> H[账户设置]
    F --> I[管理后台]
```

---

## 👤 用户功能

### 账户管理

#### 用户注册

1. **访问注册页面**

   - 点击登录页面的"立即注册"链接
   - 或直接访问：`http://localhost:8080/register`

2. **填写注册信息**

   ```
   📧 邮箱地址：your-email@example.com
   🔒 登录密码：至少8位，包含字母和数字
   🔒 确认密码：再次输入相同密码
   👤 用户昵称：可选，用于显示
   ```

3. **邮箱验证**

   - 系统自动发送验证邮件
   - 检查邮箱（包括垃圾邮件）
   - 点击验证链接激活账户

4. **注册要求**
   - 邮箱必须有效且未被注册
   - 密码强度：8 位以上，包含字母数字
   - 验证链接 24 小时内有效

#### 用户登录

1. **登录方式**

   - 邮箱 + 密码登录
   - 支持"记住我"功能（7 天免登录）
   - 安全的 JWT Token 认证

2. **登录过程**

   ```
   📧 邮箱：输入注册时的邮箱地址
   🔒 密码：输入账户密码
   ✅ 验证：系统自动验证身份
   🔐 授权：获取访问令牌并跳转到仪表盘
   ```

3. **安全特性**
   - **速率限制**：5 次/分钟，防止暴力破解
   - **CSRF 保护**：防跨站请求伪造
   - **会话管理**：安全的会话控制
   - **自动登出**：长时间无操作自动退出

#### 密码管理

**忘记密码重置**：

1. **发起重置**

   - 访问登录页面
   - 点击"忘记密码"
   - 输入注册邮箱

2. **邮件处理**

   - 检查邮箱中的重置链接
   - 重置链接 1 小时内有效
   - 点击链接跳转到重置页面

3. **设置新密码**
   - 输入新密码（8 位以上）
   - 确认新密码
   - 提交完成重置

**修改密码**：

1. 登录后进入"个人设置"
2. 在密码设置区域输入新密码
3. 确认后保存更改

---

### DNS 记录管理

这是系统的核心功能模块，用户可以管理自己的 DNS 解析记录。

#### 查看 DNS 记录

1. **访问记录管理页面**

   - 登录后点击"DNS 记录"菜单
   - 或访问：`http://localhost:8080/dns-records`

2. **记录列表显示**
   | 列名 | 说明 | 示例 |
   |------|------|------|
   | **完整域名** | 子域名.主域名 | `www.example.com` |
   | **记录类型** | DNS 记录类型 | `A`, `CNAME`, `TXT`, `MX` |
   | **记录值** | 解析目标值 | `192.168.1.1`, `target.domain.com` |
   | **TTL** | 生存时间（秒） | `600`（10 分钟） |
   | **创建时间** | 记录创建时间 | `2024-01-15 14:30:25` |

#### 创建 DNS 记录

1. **点击"添加记录"按钮**

2. **填写记录信息**

   ```
   🌐 主域名：从可用域名列表中选择
   📝 子域名：输入子域名（如：www、blog、api）
   📋 记录类型：选择DNS记录类型
   📍 记录值：输入解析目标
   ⏰ TTL：设置缓存时间（可选，默认600秒）
   ```

3. **记录类型详解**

   **A 记录**：

   - **用途**：将域名解析到 IPv4 地址
   - **示例**：
     ```
     子域名：www
     记录值：192.168.1.100
     结果：www.example.com → 192.168.1.100
     ```

   **CNAME 记录**：

   - **用途**：将域名解析到另一个域名
   - **示例**：
     ```
     子域名：blog
     记录值：myblog.github.io
     结果：blog.example.com → myblog.github.io
     ```

   **TXT 记录**：

   - **用途**：存储文本信息，常用于域名验证
   - **示例**：
     ```
     子域名：@
     记录值：google-site-verification=abc123
     用途：Google搜索控制台验证
     ```

   **MX 记录**：

   - **用途**：指定邮件服务器
   - **示例**：
     ```
     子域名：@
     记录值：10 mail.example.com
     结果：example.com的邮件由mail.example.com处理
     ```

4. **验证和提交**
   - 系统自动验证记录格式
   - 检查记录是否重复
   - 验证通过后同步到 DNS 服务商

#### 编辑 DNS 记录

1. **在记录列表中点击"编辑"按钮**
2. **修改需要变更的字段**
   - 子域名、记录类型一般不建议修改
   - 常修改：记录值、TTL
3. **保存更改**
   - 系统验证修改内容
   - 自动同步到 DNS 服务商
   - 记录修改历史（审计功能）

#### 删除 DNS 记录

1. **确认删除操作**

   - 点击记录旁的"删除"按钮
   - 系统弹出确认对话框
   - 确认后执行删除

2. **删除后果**
   - DNS 记录立即失效
   - 域名解析将停止工作
   - 删除操作不可撤销

> ⚠️ **删除警告**：删除 DNS 记录会立即影响域名解析，请确认无误后操作

---

### 个人设置

#### 账户信息管理

1. **访问个人设置**

   - 点击右上角用户头像
   - 选择"个人设置"菜单

2. **可修改的信息**

   ```
   📧 邮箱地址：用于登录和接收通知
   👤 用户昵称：显示名称
   🔒 登录密码：账户安全密码
   ```

3. **邮箱修改流程**
   - 输入新邮箱地址
   - 系统发送验证邮件到新邮箱
   - 点击验证链接确认更改

#### 安全设置

**会话管理**：

- 查看当前登录会话
- 远程登出其他设备
- 设置会话超时时间

**操作日志**：

- 查看账户操作历史
- 包括：登录记录、DNS 记录变更
- 提供安全审计信息

---

## 🛡️ 管理员功能

管理员拥有系统的完整管理权限，可以管理用户、域名资源、系统配置等。

### 系统概览

#### 管理仪表盘

1. **访问管理后台**

   - 使用管理员账户登录
   - 自动跳转到管理仪表盘
   - 或访问：`http://localhost:8080/admin`

2. **关键指标概览**

   | 指标项         | 说明                 | 监控要点     |
   | -------------- | -------------------- | ------------ |
   | **总用户数**   | 系统注册用户总数     | 用户增长趋势 |
   | **活跃用户**   | 近 30 天有操作的用户 | 系统活跃度   |
   | **总域名数**   | 系统可用域名数量     | 资源配置情况 |
   | **活跃域名**   | 有 DNS 记录的域名    | 资源使用率   |
   | **DNS 记录数** | 所有用户创建的记录   | 系统负载指标 |
   | **DNS 服务商** | 配置的服务商数量     | 服务可用性   |

3. **系统状态监控**
   - 服务运行状态
   - 数据库连接状态
   - DNS 服务商连接状态
   - 邮件服务状态

---

### 用户管理

#### 用户列表管理

1. **用户列表功能**

   - **搜索用户**：按邮箱、昵称搜索
   - **分页浏览**：支持大量用户的分页显示
   - **状态筛选**：按激活状态筛选用户

2. **用户信息字段**
   ```
   🆔 用户ID：系统唯一标识
   📧 邮箱地址：登录凭证和联系方式
   👤 用户昵称：显示名称
   ✅ 激活状态：已激活/未激活/已禁用
   🛡️ 权限级别：普通用户/管理员
   📅 注册时间：账户创建时间
   📊 DNS配额：允许创建的最大DNS记录数
   ```

#### 用户账户操作

**编辑用户信息**：

1. **基本信息修改**

   ```
   📧 邮箱地址：修改用户登录邮箱
   👤 昵称：更新显示名称
   🔒 密码重置：为用户重置登录密码
   ```

2. **权限设置**

   ```
   ✅ 账户状态：
     - 激活：用户可正常使用系统
     - 禁用：暂时禁止用户登录
     - 封禁：永久禁止用户使用

   🛡️ 管理员权限：
     - 启用：用户获得管理员权限
     - 禁用：撤销管理员权限
   ```

3. **配额管理**
   ```
   📊 DNS记录配额：
     - 默认配额：10条记录
     - 可调整范围：1-1000条
     - VIP用户可设置更高配额
   ```

**删除用户账户**：

1. **删除前确认**

   - 系统显示用户的所有 DNS 记录
   - 确认删除后果（记录将一并删除）
   - 输入确认文本进行二次确认

2. **删除后处理**
   - 用户账户标记为删除
   - 相关 DNS 记录自动清理
   - 从 DNS 服务商删除对应记录

> ⚠️ **删除警告**：用户删除操作不可撤销，请谨慎操作

---

### 域名资源管理

域名管理功能允许管理员配置用户可以使用的域名资源。

#### 查看域名列表

1. **访问域名管理**

   - 进入管理后台
   - 点击"域名管理"菜单

2. **域名信息展示**
   | 字段 | 说明 | 示例 |
   |------|------|------|
   | **域名名称** | 完整域名 | `example.com` |
   | **域名类型** | 域名分类 | `二级域名` |
   | **状态** | 启用/禁用 | `已启用` |
   | **DNS 记录数** | 该域名下的记录数 | `25条记录` |
   | **创建时间** | 添加时间 | `2024-01-15` |

#### 添加新域名

1. **添加域名资源**

   ```
   🌐 域名名称：example.com
   📝 域名描述：示例网站域名
   ✅ 启用状态：立即启用供用户使用
   ```

2. **域名验证**

   - 系统自动验证域名格式
   - 检查域名是否已在 DNS 服务商中存在
   - 验证管理权限

3. **同步到 DNS 服务商**
   - 自动从 DNS 服务商同步域名
   - 验证域名控制权
   - 获取现有 DNS 记录

#### 域名同步功能

**手动同步**：

1. 点击"同步域名"按钮
2. 系统从 DNS 服务商获取最新域名列表
3. 自动添加新域名，更新状态

**自动同步**：

- 系统每 24 小时自动同步一次
- 发现新域名时自动添加
- 检测域名状态变更

---

### DNS 服务商配置

系统支持多个 DNS 服务商，目前完整支持 DNSPod。

#### 支持的 DNS 服务商

| 服务商            | API 版本 | 配置类型   | 状态        |
| ----------------- | -------- | ---------- | ----------- |
| **腾讯云 DNSPod** | 传统 API | Token 认证 | ✅ 完整支持 |
| **腾讯云 DNSPod** | API v3.0 | 密钥对认证 | ✅ 完整支持 |
| **阿里云 DNS**    | API v3.0 | 密钥对认证 | 🔄 开发中   |
| **Cloudflare**    | API v4.0 | Token 认证 | 📋 计划中   |

#### DNSPod 传统 API 配置

1. **获取 API Token**

   - 访问：[DNSPod 控制台](https://console.dnspod.cn/account/token)
   - 创建新的 API Token
   - 格式：`用户ID,Token字符串`

2. **添加服务商配置**

   ```
   📛 配置名称：DNSPod主账号
   🔧 服务商类型：dnspod
   🔑 配置信息：
   {
     "token": "123456,your_token_here"
   }
   📝 配置描述：主要DNSPod账号配置
   ```

3. **测试连接**
   - 系统自动验证 Token 有效性
   - 测试 API 连接状态
   - 获取账号下的域名列表

#### DNSPod API v3 配置

1. **获取云 API 密钥**

   - 访问：[腾讯云控制台](https://console.cloud.tencent.com/cam/capi)
   - 创建子用户（推荐）或使用主账号
   - 授予 DNSPod 相关权限

2. **添加 v3 服务商配置**

   ```
   📛 配置名称：腾讯云DNSPod
   🔧 服务商类型：dnspod_v3
   🔑 配置信息：
   {
     "secret_id": "AKIDxxxxxxxxxxxxx",
     "secret_key": "xxxxxxxxxxxxxxxxx",
     "region": "ap-guangzhou"
   }
   ```

3. **权限要求**
   - DNSPod 域名解析权限
   - 建议创建专用子用户
   - 最小权限原则

#### 服务商管理操作

**激活/禁用服务商**：

- 激活：启用该服务商进行 DNS 操作
- 禁用：停止使用但保留配置

**设置默认服务商**：

- 新 DNS 记录将优先使用默认服务商
- 可设置多个服务商作为备用

**删除服务商配置**：

- 确认无 DNS 记录正在使用该服务商
- 删除配置信息
- 清理相关外部记录 ID

---

### SMTP 邮件配置

管理员可以配置多个 SMTP 服务器，用于发送系统邮件。

#### 支持的邮件服务

| 服务商       | SMTP 地址          | 端口    | 加密方式 | 备注               |
| ------------ | ------------------ | ------- | -------- | ------------------ |
| **Gmail**    | smtp.gmail.com     | 587     | TLS      | 需要应用专用密码   |
| **QQ 邮箱**  | smtp.qq.com        | 587/465 | TLS/SSL  | 需要开启 SMTP 服务 |
| **163 邮箱** | smtp.163.com       | 994     | SSL      | 需要开启客户端授权 |
| **企业邮箱** | smtp.exmail.qq.com | 587     | TLS      | 腾讯企业邮箱       |
| **自建邮服** | 自定义             | 自定义  | TLS/SSL  | 支持自建邮件服务器 |

#### 添加 SMTP 配置

1. **基础配置**

   ```
   📛 配置名称：Gmail主邮箱
   🏢 SMTP服务器：smtp.gmail.com
   📡 端口号：587
   👤 用户名：your-email@gmail.com
   🔒 密码：your-app-password
   ```

2. **发件人设置**

   ```
   📧 发件人邮箱：noreply@yourdomain.com
   👤 发件人名称：Domain-MAX系统
   🔒 使用TLS加密：是（推荐）
   ```

3. **高级配置**
   ```
   📝 配置描述：主要邮件发送账户
   ✅ 启用状态：立即启用
   ⭐ 设为默认：作为默认SMTP配置
   ```

#### SMTP 配置测试

1. **测试邮件发送**

   ```
   📧 测试邮箱：输入测试邮件接收地址
   📤 发送测试：系统发送测试邮件
   ✅ 验证结果：检查邮件是否成功送达
   ```

2. **常见测试问题**
   - **认证失败**：检查用户名密码
   - **连接超时**：检查 SMTP 服务器和端口
   - **TLS 错误**：检查加密设置
   - **被拒绝**：检查邮箱服务商的 SMTP 设置

#### 邮件模板管理

**系统邮件模板**：

1. **注册验证邮件**

   ```
   主题：[Domain-MAX] 激活您的账户
   内容：包含验证链接，24小时有效
   ```

2. **密码重置邮件**

   ```
   主题：[Domain-MAX] 密码重置请求
   内容：包含重置链接，1小时有效
   ```

3. **系统通知邮件**
   ```
   主题：[Domain-MAX] 重要通知
   内容：系统维护、安全警告等
   ```

---

## 🔧 高级功能

### 批量 DNS 操作

#### 批量导入 DNS 记录

1. **准备 CSV 文件**

   ```csv
   subdomain,type,value,ttl,domain
   www,A,192.168.1.1,600,example.com
   mail,MX,"10 mail.example.com",3600,example.com
   ftp,CNAME,files.example.com,600,example.com
   ```

2. **导入流程**

   - 上传 CSV 文件
   - 系统预览导入内容
   - 验证所有记录格式
   - 确认后批量创建

3. **错误处理**
   - 显示验证失败的记录
   - 允许修正后重新导入
   - 记录导入操作日志

#### 批量导出 DNS 记录

1. **选择导出范围**

   - 全部记录
   - 按域名筛选
   - 按用户筛选
   - 按时间范围

2. **导出格式选择**
   - CSV 格式：便于 Excel 处理
   - JSON 格式：便于程序处理
   - Zone 文件格式：便于 DNS 服务器

### API 接口使用

#### 获取 API 访问令牌

1. **生成个人访问令牌**

   - 进入"个人设置" → "API 设置"
   - 点击"生成新令牌"
   - 设置令牌权限范围
   - 复制并安全保存令牌

2. **API 使用示例**

   ```bash
   # 获取DNS记录
   curl -X GET "http://localhost:8080/api/dns-records" \
     -H "Authorization: Bearer your_token_here"

   # 创建DNS记录
   curl -X POST "http://localhost:8080/api/dns-records" \
     -H "Authorization: Bearer your_token_here" \
     -H "Content-Type: application/json" \
     -d '{
       "domain_id": 1,
       "subdomain": "api",
       "type": "A",
       "value": "192.168.1.2",
       "ttl": 600
     }'
   ```

#### API 接口文档

**用户接口**：

| 接口                    | 方法   | 功能              | 参数                  |
| ----------------------- | ------ | ----------------- | --------------------- |
| `/api/dns-records`      | GET    | 获取 DNS 记录列表 | -                     |
| `/api/dns-records`      | POST   | 创建 DNS 记录     | JSON Body             |
| `/api/dns-records/{id}` | PUT    | 更新 DNS 记录     | Record ID + JSON Body |
| `/api/dns-records/{id}` | DELETE | 删除 DNS 记录     | Record ID             |
| `/api/domains`          | GET    | 获取可用域名      | -                     |
| `/api/profile`          | GET    | 获取用户信息      | -                     |
| `/api/profile`          | PUT    | 更新用户信息      | JSON Body             |

**管理员接口**：

| 接口                       | 方法     | 功能           | 权限要求 |
| -------------------------- | -------- | -------------- | -------- |
| `/api/admin/stats`         | GET      | 获取系统统计   | 管理员   |
| `/api/admin/users`         | GET      | 获取用户列表   | 管理员   |
| `/api/admin/users/{id}`    | PUT      | 更新用户信息   | 管理员   |
| `/api/admin/domains`       | GET/POST | 域名管理       | 管理员   |
| `/api/admin/dns-providers` | GET/POST | DNS 服务商管理 | 管理员   |
| `/api/admin/smtp-configs`  | GET/POST | SMTP 配置管理  | 管理员   |

---

## ❓ 常见问题

### 账户相关

**Q：注册后没有收到验证邮件怎么办？**

A：请按以下步骤排查：

1. **检查垃圾邮件**：验证邮件可能被误判为垃圾邮件
2. **检查邮箱地址**：确认注册时邮箱输入正确
3. **重新发送**：在登录页面点击"重发验证邮件"
4. **联系管理员**：如果仍未收到，请联系系统管理员检查 SMTP 配置

**Q：忘记密码怎么办？**

A：密码重置流程：

1. **访问重置页面**：在登录页面点击"忘记密码"
2. **输入邮箱**：输入注册时的邮箱地址
3. **检查邮件**：重置链接将发送到邮箱
4. **设置新密码**：点击链接设置新密码
5. **链接有效期**：重置链接 1 小时内有效

**Q：如何提高账户安全性？**

A：安全建议：

- **强密码**：使用 12 位以上包含大小写字母、数字、特殊字符的密码
- **定期更换**：建议每 90 天更换一次密码
- **安全邮箱**：使用安全的邮箱账户
- **及时登出**：使用公共设备后及时退出登录

### DNS 记录相关

**Q：DNS 记录创建后多久生效？**

A：生效时间说明：

- **系统内立即生效**：记录在系统中立即可见
- **DNS 传播时间**：根据 TTL 设置，通常 10 分钟-24 小时
- **加速生效方法**：
  - 设置较小的 TTL 值（如 300 秒）
  - 清除本地 DNS 缓存：`ipconfig /flushdns`（Windows）或 `sudo dscacheutil -flushcache`（macOS）

**Q：支持哪些 DNS 记录类型？**

A：当前支持的记录类型：

| 记录类型  | 用途          | 格式要求               | 示例                                |
| --------- | ------------- | ---------------------- | ----------------------------------- |
| **A**     | IPv4 地址解析 | 有效的 IPv4 地址       | `192.168.1.1`                       |
| **CNAME** | 域名别名      | 有效的域名             | `target.example.com`                |
| **TXT**   | 文本记录      | 任意文本（255 字符内） | `"google-site-verification=abc123"` |
| **MX**    | 邮件服务器    | 优先级 + 域名          | `10 mail.example.com`               |

**Q：为什么 DNS 记录创建失败？**

A：常见失败原因：

1. **格式错误**：记录值格式不符合要求
2. **重复记录**：相同子域名和类型的记录已存在
3. **权限不足**：DNS 服务商权限配置问题
4. **配额限制**：已达到 DNS 记录配额上限
5. **服务商错误**：DNS 服务商 API 临时故障

**解决方法**：

- 检查记录值格式是否正确
- 确认是否存在重复记录
- 联系管理员检查配额设置
- 查看详细错误信息进行针对性处理

### 系统管理相关

**Q：如何备份系统数据？**

A：数据备份方案（详细步骤参考 [部署指南 - 维护与运维](./DEPLOYMENT.md#️-维护与运维)）：

1. **数据库备份**

   ```bash
   # 手动备份
   make backup

   # 自动定期备份
   crontab -e
   0 2 * * * cd /opt/domain-manager && make backup
   ```

2. **配置文件备份**

   ```bash
   # 备份配置
   cp .env /opt/backups/env_$(date +%Y%m%d).backup

   # 备份DNS服务商配置
   docker-compose exec app pg_dump -U postgres -t dns_providers domain_manager > dns_providers_backup.sql
   ```

**Q：如何升级系统版本？**

A：升级流程（完整升级指南参考 [部署指南 - 维护与运维](./DEPLOYMENT.md#️-维护与运维)）：

1. **备份当前系统**

   ```bash
   # 完整备份
   make backup-all
   ```

2. **下载新版本**

   ```bash
   git fetch origin
   git checkout v1.2.0  # 替换为目标版本
   ```

3. **执行升级**

   ```bash
   # Docker方式升级
   docker-compose build --no-cache
   docker-compose up -d

   # 源码方式升级
   make build
   sudo systemctl restart domain-manager
   ```

4. **验证升级**

   ```bash
   # 检查服务状态
   curl http://localhost:8080/api/health

   # 检查版本信息
   curl http://localhost:8080/api/version
   ```

> 📖 **详细升级指南**：完整的版本升级流程和注意事项请参考 [部署指南 - 维护与运维](./DEPLOYMENT.md#️-维护与运维)

---

## 🆘 故障排除

### 系统故障诊断

#### 服务无法访问

**症状**：无法访问 Web 界面或 API

**诊断步骤**：

1. **检查服务状态**

   ```bash
   # Docker部署
   docker-compose ps
   docker-compose logs app

   # 源码部署
   sudo systemctl status domain-manager
   journalctl -u domain-manager -f
   ```

2. **检查端口占用**

   ```bash
   # 检查8080端口
   netstat -tulpn | grep :8080
   lsof -i :8080
   ```

3. **检查网络连接**

   ```bash
   # 本地连接测试
   curl -i http://localhost:8080/api/health

   # 外部访问测试
   telnet your-domain.com 8080
   ```

#### 数据库连接问题

**症状**：应用启动失败，提示数据库连接错误

**解决方案**：

1. **检查数据库服务**

   ```bash
   # Docker方式
   docker-compose exec db psql -U postgres -c "SELECT version();"

   # 系统服务方式
   sudo systemctl status postgresql
   sudo -u postgres psql -c "SELECT version();"
   ```

2. **验证连接参数**

   ```bash
   # 检查配置
   cat .env | grep DB_

   # 测试连接
   psql "postgresql://postgres:password@localhost:5432/domain_manager"
   ```

3. **重置数据库连接**

   ```bash
   # 重启数据库
   docker-compose restart db

   # 重建连接池
   docker-compose restart app
   ```

#### DNS 同步问题

**症状**：DNS 记录创建成功但不生效

**诊断流程**：

1. **检查 DNS 服务商状态**

   ```bash
   # 测试API连接
   curl -X POST "https://dnsapi.cn/Domain.List" \
     -d "login_token=$DNSPOD_TOKEN&format=json"
   ```

2. **验证记录同步**

   ```bash
   # 查询DNS记录
   nslookup subdomain.yourdomain.com
   dig subdomain.yourdomain.com
   ```

3. **强制重新同步**
   - 进入管理后台 → 域名管理
   - 点击"同步域名"按钮
   - 检查同步日志和错误信息

### 性能优化建议

#### 系统性能监控

```bash
# 1. 系统资源监控
htop
free -h
df -h

# 2. 应用性能监控
docker stats domain-max_app_1

# 3. 数据库性能
docker-compose exec db psql -U postgres -c "
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;"
```

#### 优化建议

**数据库优化**：

- 定期执行 `VACUUM ANALYZE`
- 监控慢查询日志
- 适当增加连接池大小

**应用优化**：

- 启用 Gzip 压缩
- 配置静态资源缓存
- 使用 Redis 缓存热点数据

**系统优化**：

- 配置日志轮转
- 监控磁盘空间使用
- 定期清理临时文件

---

## 🔐 安全最佳实践

### 系统安全配置

1. **强制 HTTPS**

   ```nginx
   # Nginx配置强制HTTPS
   server {
       listen 80;
       server_name yourdomain.com;
       return 301 https://$server_name$request_uri;
   }
   ```

2. **安全响应头**

   ```nginx
   add_header X-Content-Type-Options nosniff;
   add_header X-Frame-Options DENY;
   add_header X-XSS-Protection "1; mode=block";
   add_header Strict-Transport-Security "max-age=31536000";
   ```

3. **防火墙配置**
   ```bash
   # 基础防火墙规则
   sudo ufw enable
   sudo ufw allow ssh
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw deny 8080/tcp  # 仅内网访问应用端口
   ```

### 定期安全维护

1. **密钥轮转**（每 90 天）

   ```bash
   # 生成新的JWT密钥
   openssl rand -base64 64 | tr -d "=+/" | cut -c1-64

   # 更新配置并重启
   docker-compose restart app
   ```

2. **用户审计**（每月）

   - 审查管理员权限分配
   - 清理无效用户账户
   - 检查异常登录记录

3. **系统更新**（及时）
   - 关注安全更新公告
   - 及时升级系统版本
   - 更新基础镜像版本

---

## 📞 获取支持

### 支持渠道

| 问题类型     | 联系方式                                                                   | 响应时间  | 说明             |
| ------------ | -------------------------------------------------------------------------- | --------- | ---------------- |
| **Bug 报告** | [GitHub Issues](https://github.com/Domain-MAX/Domain-MAX/issues)           | 24 小时内 | 软件缺陷和错误   |
| **功能建议** | [GitHub Discussions](https://github.com/Domain-MAX/Domain-MAX/discussions) | 48 小时内 | 新功能需求讨论   |
| **实时交流** | [Discord 社区](https://discord.gg/n4AdZGwy5K)                              | 即时响应  | 技术交流和讨论   |
| **使用咨询** | 查看项目文档                                                               | -         | 查阅详细使用说明 |

### 问题反馈模板

**使用问题反馈**：

```markdown
**问题描述：**
[详细描述遇到的问题]

**操作步骤：**

1. [步骤 1]
2. [步骤 2]
3. [步骤 3]

**预期结果：**
[描述期望的正确行为]

**实际结果：**
[描述实际发生的情况]

**环境信息：**

- 浏览器：Chrome 118.0
- 操作系统：Windows 11
- 系统版本：v1.0.0

**截图/日志：**
[如有必要，提供截图或错误日志]
```

**功能建议模板**：

```markdown
**功能概述：**
[简短描述建议的功能]

**使用场景：**
[描述什么情况下需要此功能]

**预期效果：**
[描述功能实现后的预期效果]

**实现建议：**
[如果有技术实现建议，可以描述]

**优先级：**
[高/中/低] 并说明原因
```

---

## 📈 进阶使用技巧

### 高效 DNS 管理

1. **记录命名规范**

   ```
   🌐 服务类型命名：
   - www：网站主入口
   - api：API服务端点
   - mail：邮件服务
   - ftp：文件传输服务

   🏢 环境标识：
   - dev：开发环境
   - staging：测试环境
   - prod：生产环境
   ```

2. **TTL 设置策略**

   ```
   ⚡ 快速更新场景（TTL: 60-300秒）：
   - 应用发布期间
   - A/B测试环境
   - 故障切换场景

   🔄 正常业务场景（TTL: 600-3600秒）：
   - 稳定的业务服务
   - 日常运营环境

   🐌 长期稳定场景（TTL: 3600-86400秒）：
   - MX邮件记录
   - NS记录
   - 很少变更的服务
   ```

### 批量操作技巧

1. **CSV 导入格式**

   ```csv
   subdomain,type,value,ttl,comment
   www,A,192.168.1.10,600,"主站点"
   api,A,192.168.1.11,300,"API服务"
   mail,MX,"10 mail.domain.com",3600,"邮件服务"
   ```

2. **导入前验证**
   - 使用 Excel 或文本编辑器检查格式
   - 确保所有必填字段完整
   - 验证 IP 地址和域名格式

### API 自动化

**自动化脚本示例**：

```python
#!/usr/bin/env python3
import requests
import json

# API配置
API_BASE = "http://localhost:8080/api"
TOKEN = "your_api_token_here"

headers = {
    "Authorization": f"Bearer {TOKEN}",
    "Content-Type": "application/json"
}

def create_dns_record(subdomain, record_type, value, domain_id=1, ttl=600):
    """创建DNS记录"""
    data = {
        "domain_id": domain_id,
        "subdomain": subdomain,
        "type": record_type,
        "value": value,
        "ttl": ttl
    }

    response = requests.post(f"{API_BASE}/dns-records",
                           headers=headers,
                           json=data)

    if response.status_code == 201:
        print(f"✅ 记录创建成功: {subdomain}.domain.com")
        return response.json()
    else:
        print(f"❌ 记录创建失败: {response.json()}")
        return None

# 批量创建示例
records = [
    {"subdomain": "www", "type": "A", "value": "192.168.1.10"},
    {"subdomain": "api", "type": "A", "value": "192.168.1.11"},
    {"subdomain": "mail", "type": "MX", "value": "10 mail.domain.com"},
]

for record in records:
    create_dns_record(**record)
```

---

## 🔄 系统维护指南

### 日常维护任务

#### 每日检查

- [ ] **系统健康状态**：检查服务是否正常运行
- [ ] **错误日志**：查看是否有异常错误
- [ ] **磁盘空间**：确保有足够的存储空间
- [ ] **备份状态**：验证自动备份是否正常执行

#### 每周维护

- [ ] **性能监控**：检查系统响应时间和资源使用
- [ ] **安全审计**：审查用户活动和权限变更
- [ ] **数据清理**：清理过期的验证令牌和临时文件
- [ ] **更新检查**：检查是否有新版本或安全补丁

#### 每月维护

- [ ] **用户统计**：分析用户增长和活跃度
- [ ] **容量规划**：评估系统资源需求
- [ ] **安全评估**：审查系统配置和访问控制
- [ ] **备份测试**：验证备份文件的完整性和恢复能力

### 监控指标

#### 关键指标

| 指标类别     | 监控指标           | 正常范围 | 告警阈值 |
| ------------ | ------------------ | -------- | -------- |
| **系统性能** | CPU 使用率         | < 60%    | > 80%    |
|              | 内存使用率         | < 70%    | > 85%    |
|              | 磁盘使用率         | < 80%    | > 90%    |
| **应用性能** | 响应时间           | < 500ms  | > 2000ms |
|              | 错误率             | < 1%     | > 5%     |
|              | 并发连接数         | < 100    | > 500    |
| **业务指标** | 每日活跃用户       | -        | 显著下降 |
|              | DNS 记录创建成功率 | > 99%    | < 95%    |
|              | 邮件发送成功率     | > 98%    | < 90%    |

#### 告警配置

```yaml
# prometheus告警规则示例
groups:
  - name: domain-manager
    rules:
      - alert: HighMemoryUsage
        expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "内存使用率过高"

      - alert: APIResponseSlow
        expr: http_request_duration_seconds{job="domain-manager"} > 2
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "API响应时间过长"
```

---

## 🎓 最佳实践指南

### 用户使用最佳实践

1. **DNS 记录规划**

   - **命名规范**：使用有意义的子域名
   - **环境分离**：开发、测试、生产环境使用不同子域名
   - **备份策略**：重要记录做好备份

2. **安全使用**
   - **定期修改密码**：建议每 3 个月更换
   - **监控异常活动**：关注登录地点和时间
   - **权限最小化**：仅申请必要的 DNS 记录配额

### 管理员管理最佳实践

1. **用户管理**

   - **权限审核**：定期审核管理员权限
   - **配额合理分配**：根据用户需求合理设置 DNS 配额
   - **及时处理**：快速响应用户问题和请求

2. **系统维护**

   - **定期备份**：确保数据安全
   - **监控告警**：设置关键指标监控
   - **版本更新**：及时更新到稳定版本

3. **DNS 服务商管理**
   - **多服务商冗余**：配置多个 DNS 服务商避免单点故障
   - **权限最小化**：DNS 服务商 API 权限仅授予必要范围
   - **定期测试**：验证 DNS 服务商连接状态

---

<div align="center">

## 🎉 开始使用 Domain-MAX！

现在您已经掌握了 Domain-MAX 的完整使用方法。

**遇到问题？** 查看 [部署指南](./DEPLOYMENT.md) 解决技术问题  
**想要贡献？** 查看 [贡献指南](./CONTRIBUTING.md) 参与项目开发  
**需要支持？** 加入 [Discord 社区](https://discord.gg/n4AdZGwy5K) 获取实时帮助

[🏠 返回首页](./README.md) | [🚀 部署指南](./DEPLOYMENT.md) | [🤝 参与贡献](./CONTRIBUTING.md)

</div>
