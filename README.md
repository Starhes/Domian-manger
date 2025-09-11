# Domain MAX - 二级域名分发管理系统

[![Go Version](https://img.shields.io/github/go-mod/go-version/Domain-MAX/Domain-MAX)](https://golang.org/)
[![License](https://img.shields.io/github/license/Domain-MAX/Domain-MAX)](LICENSE)
[![Docker Build](https://github.com/Domain-MAX/Domain-MAX/actions/workflows/docker-build.yml/badge.svg)](https://github.com/Domain-MAX/Domain-MAX/actions/workflows/docker-build.yml)

**Domain MAX** 是一个现代、高效的二级域名分发管理系统。采用前后端分离架构，基于 Go 和 React 构建，通过单一 Docker 镜像交付，为用户和管理员提供了强大而直观的域名及 DNS 记录管理功能。

![系统截图](https://your-image-url.com/screenshot.png) <!-- 请替换为您的系统截图URL -->

---

## ✨ 核心功能

<table width="100%">
  <tr>
    <td width="50%" valign="top">
      <h4>👤 用户端</h4>
      <ul>
        <li><b>自助注册激活</b>：支持邮箱注册、邮件激活及密码重置。</li>
        <li><b>DNS记录管理</b>：自助对 A, CNAME, TXT, MX 等记录进行增删改查。</li>
        <li><b>多域名支持</b>：可在管理员开放的多个主域名下创建子域名。</li>
        <li><b>实时同步</b>：所有DNS记录变更将实时同步至DNS服务商。</li>
      </ul>
    </td>
    <td width="50%" valign="top">
      <h4>🛡️ 管理端</h4>
      <ul>
        <li><b>多维度仪表盘</b>：可视化展示用户、域名、记录等核心数据。</li>
        <li><b>用户管理</b>：支持对用户进行查看、搜索、编辑和禁用等操作。</li>
        <li><b>域名管理</b>：统一管理可供分发的主域名列表。</li>
        <li><b>多服务商支持</b>：可配置并切换多个DNS服务商的API凭证。</li>
      </ul>
    </td>
  </tr>
</table>

## 🛠️ 技术栈

- **后端**: Go, Gin, GORM
- **前端**: React, TypeScript, Vite, Ant Design, Zustand
- **数据库**: PostgreSQL, MySQL
- **部署**: Docker, Docker Compose (多阶段构建)
- **认证**: JWT

## 🚀 快速上手 (Docker)

在 **3 分钟** 内启动您的域名分发系统。

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

## 📖 详细部署与运维

对于更高级的部署场景，如 **生产环境配置**、**数据备份**，或希望 **在本地直接运行源码进行开发**，请参阅我们为您准备的详细文档：

- **[部署与运维指南 (DEPLOYMENT.md)](./DEPLOYMENT.md)**

## 🤝 参与贡献

我们欢迎任何形式的贡献！无论是提交 Issue、发起 Pull Request，还是改进文档，都是对项目的巨大支持。

请在贡献前阅读我们的 **[贡献指南](CONTRIBUTING.md)**。

## 📄 开源许可

本项目基于 [MIT License](LICENSE) 开源。
