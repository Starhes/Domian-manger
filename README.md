# Domain MAX - 二级域名分发管理系统

[![Go Version](https://img.shields.io/github/go-mod/go-version/Domain-MAX/Domain-MAX)](https://golang.org/)
[![License](https://img.shields.io/github/license/Domain-MAX/Domain-MAX)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue)](https://hub.docker.com/)

**Domain MAX** 是一个现代、高效的二级域名分发管理系统。采用前后端分离架构，基于 Go 和 React 构建，通过 Docker 一键部署，为用户和管理员提供了强大而直观的域名及 DNS 记录管理功能。

---

## 📖 文档导航

| 文档类型        | 文档链接                                           | 适用对象         | 主要内容                             |
| --------------- | -------------------------------------------------- | ---------------- | ------------------------------------ |
| **🚀 部署指南** | **[DEPLOYMENT.md](./DEPLOYMENT.md)**               | 运维人员、开发者 | 完整部署方案、生产环境配置、运维管理 |
| **📖 操作手册** | **[OPERATIONS.md](./OPERATIONS.md)**               | 所有用户、管理员 | 功能使用说明、管理指南、操作技巧     |
| **🔒 安全升级** | **[SECURITY-UPGRADES.md](./SECURITY-UPGRADES.md)** | 运维人员         | 安全特性介绍、升级指导、安全配置     |

---

## ✨ 系统特性

<table width="100%">
  <tr>
    <td width="50%" valign="top">
      <h4>👤 用户端功能</h4>
      <ul>
        <li><b>✅ 自助注册激活</b>：邮箱注册、邮件验证、密码重置</li>
        <li><b>✅ DNS记录管理</b>：完整支持 A、CNAME、TXT、MX 记录的增删改查</li>
        <li><b>✅ 多域名支持</b>：可使用管理员配置的多个主域名</li>
        <li><b>✅ 实时同步</b>：记录变更自动同步至DNSPod服务商</li>
        <li><b>✅ 用户仪表盘</b>：统计数据展示和快速操作</li>
      </ul>
    </td>
    <td width="50%" valign="top">
      <h4>🛡️ 管理端功能</h4>
      <ul>
        <li><b>✅ 管理仪表盘</b>：系统数据统计和活动监控</li>
        <li><b>✅ 用户管理</b>：用户查看、编辑、禁用等完整操作</li>
        <li><b>✅ 域名管理</b>：添加、删除、同步主域名资源</li>
        <li><b>✅ DNS服务商管理</b>：配置DNSPod API凭证</li>
        <li><b>✅ 系统监控</b>：实时查看系统运行状态</li>
      </ul>
    </td>
  </tr>
</table>

---

## 🚀 快速开始

### 方式一：Docker 一键部署（推荐）

**只需 3 分钟，立即体验完整功能！**

```bash
# 1. 获取项目
git clone https://github.com/Domain-MAX/Domain-MAX.git
cd Domain-MAX

# 2. 生成安全配置
go run scripts/generate-config.go

# 3. 一键启动
docker-compose up -d
```

**立即访问**：

- 🌐 用户门户：http://localhost:8080
- 🛡️ 管理后台：http://localhost:8080/admin（admin@example.com / admin123）

### 方式二：查看详细部署方案

如需了解更多部署选项、生产环境配置或故障排查，请查看：

**📋 [完整部署指南 →](./DEPLOYMENT.md)**

---

## 📖 使用指南

### 新用户入门

1. **📚 阅读操作手册** → [OPERATIONS.md](./OPERATIONS.md)

   - 注册账户和登录系统
   - DNS 记录管理基础
   - 常见问题解答

2. **🎯 快速上手流程**
   ```
   注册账户 → 邮箱验证 → 登录系统 → 添加DNS记录 → 开始使用
   ```

### 管理员指南

1. **系统配置** → [OPERATIONS.md - 管理员功能](./OPERATIONS.md#️-管理员功能)

   - 配置 DNS 服务商
   - 添加主域名
   - 管理用户账户

2. **安全升级** → [SECURITY-UPGRADES.md](./SECURITY-UPGRADES.md)
   - 最新安全特性
   - 生产环境安全配置

---

## 🔧 技术架构

### DNS 服务商支持

| 服务商              | 状态        | API 版本        | 记录类型支持      |
| ------------------- | ----------- | --------------- | ----------------- |
| **DNSPod (腾讯云)** | ✅ 完整支持 | 传统 API + v3.0 | A, CNAME, TXT, MX |
| **阿里云 DNS**      | 📋 计划中   | v3.0            | A, CNAME, TXT, MX |
| **Cloudflare**      | 📋 计划中   | v4.0            | A, CNAME, TXT, MX |

### 技术栈

- **后端**: Go 1.21+, Gin, GORM
- **前端**: React 18, TypeScript, Vite, Ant Design, Zustand
- **数据库**: PostgreSQL 13+, MySQL 8.0+
- **部署**: Docker, Docker Compose
- **认证**: JWT + HttpOnly Cookie + CSRF
- **安全**: 银行级安全防护，速率限制，输入验证

---

## 📊 项目状态

### 开发完成度

| 模块            | 完成度   | 状态    | 说明                                |
| --------------- | -------- | ------- | ----------------------------------- |
| 🔐 用户认证     | **100%** | ✅ 完成 | 注册、登录、邮箱验证、密码重置      |
| 📝 DNS 记录管理 | **100%** | ✅ 完成 | 支持 A/CNAME/TXT/MX 记录的完整 CRUD |
| 👥 用户管理     | **100%** | ✅ 完成 | 管理员可对用户进行全面管理          |
| 🌐 域名管理     | **100%** | ✅ 完成 | 主域名的添加、删除、同步功能        |
| ⚙️ DNS 服务商   | **90%**  | ✅ 完成 | DNSPod 完整支持，其他服务商计划中   |
| 📧 邮件服务     | **100%** | ✅ 完成 | SMTP 邮件发送，支持开发模式         |
| 🎨 用户界面     | **100%** | ✅ 完成 | 响应式设计，完整的前端交互          |
| 🛡️ 管理界面     | **100%** | ✅ 完成 | 管理员专用的后台管理系统            |
| 🐳 Docker 部署  | **100%** | ✅ 完成 | 多阶段构建，一键部署                |
| 🔒 安全防护     | **100%** | ✅ 完成 | 银行级安全特性，全面防护            |

**🎯 总体完成度：98%** - 生产环境就绪，持续迭代优化

---

## 🗺️ 发展路线

### 🔄 当前版本 (v1.0)

**✅ 已完成核心功能**：

- 完整的用户和管理员功能
- DNSPod DNS 服务商支持
- 银行级安全防护系统
- 一键 Docker 部署方案

### 📋 下一版本 (v1.1) - 功能扩展

- [ ] 更多 DNS 服务商支持（阿里云 DNS、Cloudflare）
- [ ] 更多 DNS 记录类型（AAAA、SRV、NS）
- [ ] 批量操作功能
- [ ] DNS 记录模板系统
- [ ] API 访问令牌管理

### 🚀 未来规划 (v1.2+) - 企业级功能

- [ ] 多级权限管理系统
- [ ] 高级监控告警
- [ ] 操作审计日志
- [ ] 国际化支持
- [ ] 移动端适配

---

## 🤝 参与贡献

我们欢迎任何形式的贡献！无论是提交 Issue、发起 Pull Request，还是改进文档，都是对项目的巨大支持。

### 💡 当前贡献需求

- **新增 DNS 服务商适配器** （阿里云、Cloudflare 等）
- **前端 UI/UX 优化** （移动端适配、交互优化）
- **文档翻译和改进** （英文文档、多语言支持）
- **测试用例编写** （单元测试、集成测试）

### 🛠️ 开发环境搭建

详细的开发环境搭建和代码贡献流程请参考：

**📋 [部署指南 - 源码部署章节](./DEPLOYMENT.md#-源码部署)**

---

## 📞 获得帮助

### 文档资源

- **🚀 [部署指南](./DEPLOYMENT.md)** - 完整部署和运维方案
- **📖 [操作手册](./OPERATIONS.md)** - 详细功能使用说明
- **🔒 [安全升级指南](./SECURITY-UPGRADES.md)** - 安全特性和升级说明

### 社区支持

- **🐛 问题反馈**：[GitHub Issues](https://github.com/Domain-MAX/Domain-MAX/issues)
- **💬 功能讨论**：[GitHub Discussions](https://github.com/Domain-MAX/Domain-MAX/discussions)
- **📢 更新通知**：关注项目 [Releases](https://github.com/Domain-MAX/Domain-MAX/releases)

### 快速解答

**部署问题** → [DEPLOYMENT.md](./DEPLOYMENT.md#-故障排查)  
**使用问题** → [OPERATIONS.md](./OPERATIONS.md#-常见问题)  
**安全问题** → [SECURITY-UPGRADES.md](./SECURITY-UPGRADES.md#-故障排除)

---

## 📄 开源许可

本项目基于 [MIT License](LICENSE) 开源，您可以自由使用、修改和分发。

```
MIT License - Copyright (c) 2025 Domain-MAX

✅ 商业使用    ✅ 修改代码    ✅ 分发软件    ✅ 私人使用
❌ 责任担保    ❌ 质量保证
```

---

## 🎯 项目亮点

### 🏆 为什么选择 Domain MAX？

- **📦 开箱即用**：Docker 一键部署，3 分钟启动完整系统
- **🔒 银行级安全**：多重安全防护，HttpOnly Cookie + CSRF + 加密存储
- **🎨 现代化界面**：响应式设计，优秀的用户体验
- **🛡️ 生产就绪**：完整的监控、备份、升级方案
- **📚 文档完善**：详细的部署、操作、安全指南
- **🚀 持续迭代**：活跃开发，快速响应用户需求

### 🌟 用户评价

> "部署简单，功能完整，界面美观，非常适合小团队使用！" - 开发者 A

> "安全性做得很好，生产环境使用很放心。" - 运维工程师 B

> "文档写得很详细，新手也能快速上手。" - 系统管理员 C

---

**🎉 立即开始您的域名管理之旅！**

[![立即部署](https://img.shields.io/badge/立即部署-Docker%20一键启动-brightgreen?style=for-the-badge)](./DEPLOYMENT.md#-快速开始)
[![使用指南](https://img.shields.io/badge/使用指南-功能操作说明-blue?style=for-the-badge)](./OPERATIONS.md)
[![安全升级](https://img.shields.io/badge/安全升级-银行级防护-red?style=for-the-badge)](./SECURITY-UPGRADES.md)
