package services

import (
	"crypto/tls"
	"domain-manager/internal/config"
	"fmt"
	"net/smtp"
	"strings"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

// SendVerificationEmail 发送邮箱验证邮件
func (s *EmailService) SendVerificationEmail(email, token string) error {
	if !s.isConfigured() {
		// 开发环境下，如果没有配置邮件服务，打印到控制台
		fmt.Printf("📧 邮箱验证链接（开发模式）: http://localhost:8080/api/verify-email/%s\n", token)
		fmt.Printf("📧 用户邮箱: %s\n", email)
		return nil
	}

	subject := "激活您的账户 - 域名管理系统"
	body := s.buildVerificationEmailBody(email, token)

	return s.sendEmail(email, subject, body)
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *EmailService) SendPasswordResetEmail(email, token string) error {
	if !s.isConfigured() {
		// 开发环境下，如果没有配置邮件服务，打印到控制台
		fmt.Printf("🔐 密码重置链接（开发模式）: http://localhost:8080/reset-password?token=%s\n", token)
		fmt.Printf("📧 用户邮箱: %s\n", email)
		return nil
	}

	subject := "重置您的密码 - 域名管理系统"
	body := s.buildPasswordResetEmailBody(email, token)

	return s.sendEmail(email, subject, body)
}

// isConfigured 检查邮件服务是否配置完成
func (s *EmailService) isConfigured() bool {
	return s.cfg.SMTPHost != "" &&
		s.cfg.SMTPUser != "" &&
		s.cfg.SMTPPassword != "" &&
		s.cfg.SMTPFrom != ""
}

// sendEmail 发送邮件的核心功能
func (s *EmailService) sendEmail(to, subject, body string) error {
	// 构建邮件内容
	message := s.buildEmailMessage(to, subject, body)

	// 设置认证
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	// SMTP服务器地址
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	// 如果是Gmail或其他需要TLS的服务
	if s.cfg.SMTPPort == 587 {
		return s.sendEmailWithTLS(addr, auth, s.cfg.SMTPFrom, []string{to}, []byte(message))
	}

	// 标准SMTP发送
	return smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{to}, []byte(message))
}

// sendEmailWithTLS 使用TLS发送邮件（适用于Gmail等）
func (s *EmailService) sendEmailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// 创建客户端
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %v", err)
	}
	defer client.Close()

	// 启动TLS
	if err = client.StartTLS(&tls.Config{ServerName: s.cfg.SMTPHost}); err != nil {
		return fmt.Errorf("启动TLS失败: %v", err)
	}

	// 认证
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	// 设置发件人
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	// 设置收件人
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("设置收件人失败: %v", err)
		}
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取邮件写入器失败: %v", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("关闭邮件写入器失败: %v", err)
	}

	return client.Quit()
}

// buildEmailMessage 构建标准邮件格式
func (s *EmailService) buildEmailMessage(to, subject, body string) string {
	headers := make(map[string]string)
	headers["From"] = s.cfg.SMTPFrom
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	return msg.String()
}

// buildVerificationEmailBody 构建邮箱验证邮件内容
func (s *EmailService) buildVerificationEmailBody(email, token string) string {
	verifyURL := fmt.Sprintf("http://localhost:8080/api/verify-email/%s", token)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>激活您的账户</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #1890ff; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; background: #1890ff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { margin-top: 30px; font-size: 12px; color: #666; text-align: center; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 域名管理系统</h1>
    </div>
    <div class="content">
        <h2>欢迎加入我们！</h2>
        <p>您好，</p>
        <p>感谢您注册域名管理系统。为了确保账户安全，请点击下面的按钮激活您的账户：</p>
        
        <p style="text-align: center;">
            <a href="%s" class="button">🔗 激活账户</a>
        </p>
        
        <p>如果按钮无法点击，请复制以下链接到浏览器地址栏：</p>
        <p style="word-break: break-all; background: #e6f7ff; padding: 10px; border-radius: 3px;">
            <code>%s</code>
        </p>
        
        <p><strong>注意：</strong></p>
        <ul>
            <li>此链接将在24小时后过期</li>
            <li>如果您没有注册账户，请忽略此邮件</li>
            <li>请勿将此链接分享给他人</li>
        </ul>
        
        <p>如有任何问题，请联系我们的技术支持。</p>
        
        <p>祝您使用愉快！<br>域名管理系统团队</p>
    </div>
    <div class="footer">
        <p>此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>`, verifyURL, verifyURL)
}

// buildPasswordResetEmailBody 构建密码重置邮件内容
func (s *EmailService) buildPasswordResetEmailBody(email, token string) string {
	resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>重置您的密码</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #ff4d4f; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; background: #ff4d4f; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { margin-top: 30px; font-size: 12px; color: #666; text-align: center; }
        .warning { background: #fff2f0; border-left: 4px solid #ff4d4f; padding: 15px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔐 密码重置</h1>
    </div>
    <div class="content">
        <h2>重置您的密码</h2>
        <p>您好，</p>
        <p>我们收到了重置您账户密码的请求。请点击下面的按钮来设置新密码：</p>
        
        <p style="text-align: center;">
            <a href="%s" class="button">🔑 重置密码</a>
        </p>
        
        <p>如果按钮无法点击，请复制以下链接到浏览器地址栏：</p>
        <p style="word-break: break-all; background: #e6f7ff; padding: 10px; border-radius: 3px;">
            <code>%s</code>
        </p>
        
        <div class="warning">
            <p><strong>⚠️ 安全提醒：</strong></p>
            <ul>
                <li>此链接将在1小时后过期</li>
                <li>如果您没有请求重置密码，请忽略此邮件</li>
                <li>重置密码后，所有设备将需要重新登录</li>
                <li>请勿将此链接分享给他人</li>
            </ul>
        </div>
        
        <p>如果您需要帮助或怀疑账户被盗用，请立即联系我们的安全团队。</p>
        
        <p>保持安全！<br>域名管理系统团队</p>
    </div>
    <div class="footer">
        <p>此邮件由系统自动发送，请勿回复。</p>
        <p>如果您频繁收到此类邮件，可能是有人在尝试访问您的账户。</p>
    </div>
</body>
</html>`, resetURL, resetURL)
}
