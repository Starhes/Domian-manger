## 使用与运维（OPERATIONS）

文档导航： [README](README.md) | [DEPLOYMENT](DEPLOYMENT.md) | [OPERATIONS](OPERATIONS.md)

本文面向使用者与管理员，介绍日常操作、API 摘要、管理员工作流与故障排查。结合 [README.md](README.md) 与 [DEPLOYMENT.md](DEPLOYMENT.md) 一起阅读更高效。

### 1. 首次登录与基础流程

1. 管理员账号：系统初始化插入 `admin@example.com / admin123`，请立即登录并修改密码。
2. 普通用户注册：
   - 前台注册成功后，系统发送验证邮件；未配置 SMTP 时，开发模式会在控制台打印验证链接
   - 点击验证后即可登录
3. 登录后：后端会设置认证 Cookie，并返回 `csrf_token`，前端将其放入 `X-CSRF-Token` 请求头。

相关代码位置：

- 登录接口：`/api/login`（见 `internal/api/handlers.go::Login`）
- 邮件发送：`internal/services/email.go`（支持 DB/ENV 两种 SMTP 来源）

### 2. 用户常用操作

- 资料查看：`GET /api/profile`
- 资料更新：`PUT /api/profile`（可更新邮箱、密码；更换邮箱需重新验证）
- 忘记密码：`POST /api/forgot-password`，然后 `POST /api/reset-password`

前端相关：`frontend/src/stores/authStore.ts`、`frontend/src/utils/api.ts`。

### 3. DNS 记录管理（用户侧）

- 查询我的记录：`GET /api/dns-records`
- 创建记录：`POST /api/dns-records`
- 更新记录：`PUT /api/dns-records/:id`
- 删除记录：`DELETE /api/dns-records/:id`

请求体参见模型 `models.CreateDNSRecordRequest` 与 `models.UpdateDNSRecordRequest`（包括 `domain_id/subdomain/type/value/ttl`）。系统会校验配额与格式并调用当前启用的 DNS 服务商。

### 4. 管理后台功能

入口：登录管理员账户后进入后台菜单（前端 `pages/admin/*`）。对应后端路由位于 `internal/api/handlers.go` 的 `/api/admin/*`。

#### 4.1 用户管理

- 列表/分页/搜索：`GET /api/admin/users?page=&pageSize=&search=`
- 查看：`GET /api/admin/users/:id`
- 更新：`PUT /api/admin/users/:id`（可设置 `email/password/is_active/is_admin`）
- 删除：`DELETE /api/admin/users/:id`（管理员账号不可删除）

#### 4.2 域名管理

- 列表：`GET /api/admin/domains`
- 创建：`POST /api/admin/domains`
- 更新：`PUT /api/admin/domains/:id`
- 删除：`DELETE /api/admin/domains/:id`（需先清空关联记录）
- 同步外部域名：`POST /api/admin/domains/sync`

#### 4.3 DNS 服务商管理

- 列表：`GET /api/admin/dns-providers`
- 创建：`POST /api/admin/dns-providers`
- 更新：`PUT /api/admin/dns-providers/:id`
- 删除：`DELETE /api/admin/dns-providers/:id`

配置字段中 `config` 为 JSON 字符串，请按所选服务商要求填写，例如 DNSPod：`{"token":"<your_dnspod_token>"}`。

#### 4.4 SMTP 配置管理

- 列表：`GET /api/admin/smtp-configs`
- 查看：`GET /api/admin/smtp-configs/:id`
- 创建：`POST /api/admin/smtp-configs`
- 更新：`PUT /api/admin/smtp-configs/:id`
- 删除：`DELETE /api/admin/smtp-configs/:id`（默认配置不可删）
- 激活：`POST /api/admin/smtp-configs/:id/activate`
- 设为默认：`POST /api/admin/smtp-configs/:id/set-default`
- 测试发送：`POST /api/admin/smtp-configs/:id/test`（body: `{"to":"test@xx.com"}`）

密码将使用 `ENCRYPTION_KEY` 进行 AES 加密存储。激活后邮件服务优先使用 DB 中的激活配置。

#### 4.5 系统统计

- `GET /api/admin/stats` 返回用户、域名、记录、服务商数量等聚合数据。

### 5. API 调用要点

- 所有修改类接口需携带 `X-CSRF-Token`（登录返回）。
- 认证通过 Cookie 完成；前端基于 Axios 已封装错误处理与重定向逻辑。
- 后端返回统一包裹格式时，前端会自动解包 `data` 字段。

### 6. 运行维护

- 日志查看（Docker）：`docker compose logs -f app | cat`
- 数据库备份：备份 Compose 卷 `postgres_data` 或使用外部 RDS 方案
- 配置变更：修改 `.env` 或在后台更新 SMTP/DNS 服务商配置后重启应用

### 7. 故障排查

- 登录 401/403：
  - 检查 Cookie 是否被浏览器阻止
  - 确认反向代理设置 `X-Forwarded-Proto/Host` 与 `BASE_URL` 一致
  - 前端请求是否包含 `X-CSRF-Token`
- 邮件发送失败：
  - 若为开发环境且未配 SMTP，查看控制台打印的“验证/重置链接”
  - 若为生产，检查 SMTP 主机/端口/发件人、凭据是否正确，测试接口是否成功
- DNS 操作失败：
  - 确认已启用有效的 DNS 服务商配置，凭据正确
  - 检查 TTL、记录值格式、配额限制
- 数据库迁移/连接失败：
  - 检查 `DB_*` 配置、网络连通性与权限
  - 查看容器/应用日志中的详细错误

### 8. 安全建议（摘要）

- 强密钥与最小权限：定期轮换 `JWT_SECRET/ENCRYPTION_KEY`，限制 DB 暴露
- 强认证与审核：限制管理员数量，定期审查活跃 SMTP/DNS 配置
- 传输安全：HTTPS、HSTS、正确的反向代理头

更多部署与安全细节见 [DEPLOYMENT.md](DEPLOYMENT.md)。
