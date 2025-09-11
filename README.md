# Domain MAX - 二级域名分发管理系统

[![Go Version](https://img.shields.io/github/go-mod/go-version/Domain-MAX/Domain-MAX)](https://golang.org/)
[![License](https://img.shields.io/github/license/Domain-MAX/Domain-MAX)](LICENSE)

**Domain MAX** 是一个现代、高效的二级域名分发管理系统。采用前后端分离架构，基于 Go 和 React 构建，通过单一 Docker 镜像交付，为用户和管理员提供了强大而直观的域名及 DNS 记录管理功能。

---

## ✨ 已实现功能

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

## 🔧 DNS 服务商支持

目前已集成以下 DNS 服务商：

- **✅ DNSPod (腾讯云)**: 支持传统 API 和 API v3.0 两个版本
- **📋 计划中**: 阿里云 DNS、Cloudflare、华为云 DNS 等

支持的 DNS 记录类型：**A**、**CNAME**、**TXT**、**MX**

## 🛠️ 技术栈

- **后端**: Go, Gin, GORM
- **前端**: React, TypeScript, Vite, Ant Design, Zustand
- **数据库**: PostgreSQL, MySQL
- **部署**: Docker, Docker Compose (多阶段构建)
- **认证**: JWT

## 🚀 快速上手 (Docker)

在 **3 分钟** 内启动您的域名分发系统。

> **项目状态**：✅ 核心功能已完成开发，可用于生产环境部署。

### 环境要求

- [Docker](https://docs.docker.com/get-docker/) (v20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)

### 部署步骤

1.  **克隆项目**

    ```bash
    git clone https://github.com/Domain-MAX/Domain-MAX.git
    cd Domain-MAX
    ```

2.  **配置环境变量**

    复制环境变量模板文件。

    ```bash
    cp env.example .env
    ```

    然后，编辑 `.env` 文件，**至少修改以下两项**:

    ```dotenv
    # 数据库密码 - 务必修改为一个强密码
    DB_PASSWORD=your_secure_password_here

    # JWT 密钥 - 务必修改为一个32位以上的随机字符串
    JWT_SECRET=your_jwt_secret_key_change_this_in_production
    ```

3.  **一键启动**

    ```bash
    docker-compose up -d
    ```

    服务将在后台启动并运行。

4.  **访问系统**

    - **用户端**: `http://localhost:8080`
    - **管理后台**: `http://localhost:8080/admin`

    > **默认管理员账户**
    >
    > - **邮箱**: `admin@example.com`
    > - **密码**: `admin123`
    >
    > ⚠️ **首次登录后，请务必在管理后台修改默认密码！**

## 📊 开发完成度

| 模块            | 状态    | 说明                                     |
| --------------- | ------- | ---------------------------------------- |
| 🔐 用户认证     | ✅ 100% | 注册、登录、邮箱验证、密码重置           |
| 📝 DNS 记录管理 | ✅ 100% | 支持 A/CNAME/TXT/MX 记录的完整 CRUD 操作 |
| 👥 用户管理     | ✅ 100% | 管理员可对用户进行全面管理               |
| 🌐 域名管理     | ✅ 100% | 主域名的添加、删除、同步功能             |
| ⚙️ DNS 服务商   | ✅ 90%  | DNSPod 完整支持，其他服务商待扩展        |
| 📧 邮件服务     | ✅ 100% | SMTP 邮件发送，开发模式控制台输出        |
| 🎨 用户界面     | ✅ 100% | 响应式设计，完整的前端交互               |
| 🛡️ 管理界面     | ✅ 100% | 管理员专用的后台管理系统                 |
| 🐳 Docker 部署  | ✅ 100% | 多阶段构建，一键部署                     |
| 📚 文档         | ✅ 95%  | 详细的部署和使用文档                     |

**总体完成度：97%** - 核心功能完整，可用于生产环境

## 📖 详细部署与运维

对于更高级的部署场景，如 **生产环境配置**、**数据备份**，或希望 **在本地直接运行源码进行开发**，请参阅我们为您准备的详细文档：

- **[部署与运维指南 (DEPLOYMENT.md)](./DEPLOYMENT.md)**

## 🗺️ 开发路线图

### 即将实现 (v1.1)

- [ ] 更多 DNS 服务商支持 (阿里云 DNS、Cloudflare)
- [ ] 更多 DNS 记录类型 (AAAA、SRV、NS)
- [ ] 批量操作功能
- [ ] API 访问令牌管理

### 计划中 (v1.2+)

- [ ] 多级权限管理
- [ ] DNS 记录模板功能
- [ ] 监控告警系统
- [ ] 操作审计日志

## 🤝 参与贡献

我们欢迎任何形式的贡献！无论是提交 Issue、发起 Pull Request，还是改进文档，都是对项目的巨大支持。

**当前贡献需求**：

- 新增 DNS 服务商适配器
- 前端 UI/UX 优化
- 文档翻译和改进
- 测试用例编写

## 📄 开源许可

本项目基于 [MIT License](LICENSE) 开源。

```
MIT License

Copyright (c) 2025 Domain-MAX

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
