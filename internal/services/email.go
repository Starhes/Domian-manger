package services

import (
	"crypto/tls"
	"domain-manager/internal/config"
	"domain-manager/internal/models"
	"domain-manager/internal/utils"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EmailService struct {
	cfg    *config.Config
	db     *gorm.DB
	crypto *utils.CryptoService
}

func NewEmailService(cfg *config.Config) *EmailService {
	// 初始化加密服务
	crypto, err := utils.NewCryptoService(cfg.EncryptionKey[:32])
	if err != nil {
		crypto = nil // 如果初始化失败，设为nil，后续会检查
	}
	
	return &EmailService{
		cfg:    cfg,
		crypto: crypto,
	}
}

func NewEmailServiceWithDB(cfg *config.Config, db *gorm.DB) *EmailService {
	// 初始化加密服务
	crypto, err := utils.NewCryptoService(cfg.EncryptionKey[:32])
	if err != nil {
		crypto = nil // 如果初始化失败，设为nil，后续会检查
	}
	
	return &EmailService{
		cfg:    cfg,
		db:     db,
		crypto: crypto,
	}
}

// SendVerificationEmail 发送邮箱验证邮件
func (s *EmailService) SendVerificationEmail(email, token string) error {
	return s.SendVerificationEmailWithContext(nil, email, token)
}

// SendVerificationEmailWithContext 发送邮箱验证邮件（支持HTTP上下文）
func (s *EmailService) SendVerificationEmailWithContext(c *gin.Context, email, token string) error {
	baseURL := s.getBaseURL(c)
	
	if !s.isConfigured() {
		// 开发环境下，如果没有配置邮件服务，打印到控制台
		fmt.Printf("📧 邮箱验证链接（开发模式）: %s/api/verify-email/%s\n", baseURL, token)
		fmt.Printf("📧 用户邮箱: %s\n", email)
		return nil
	}

	subject := "激活您的账户 - 域名管理系统"
	body := s.buildVerificationEmailBodyWithURL(email, token, baseURL)

	return s.sendEmail(email, subject, body)
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *EmailService) SendPasswordResetEmail(email, token string) error {
	return s.SendPasswordResetEmailWithContext(nil, email, token)
}

// SendPasswordResetEmailWithContext 发送密码重置邮件（支持HTTP上下文）
func (s *EmailService) SendPasswordResetEmailWithContext(c *gin.Context, email, token string) error {
	baseURL := s.getBaseURL(c)
	
	if !s.isConfigured() {
		// 开发环境下，如果没有配置邮件服务，打印到控制台
		fmt.Printf("🔐 密码重置链接（开发模式）: %s/reset-password?token=%s\n", baseURL, token)
		fmt.Printf("📧 用户邮箱: %s\n", email)
		return nil
	}

	subject := "重置您的密码 - 域名管理系统"
	body := s.buildPasswordResetEmailBodyWithURL(email, token, baseURL)

	return s.sendEmail(email, subject, body)
}

// isConfigured 检查邮件服务是否配置完成
func (s *EmailService) isConfigured() bool {
	// 优先检查数据库配置
	if s.db != nil {
		if config := s.getActiveSMTPConfig(); config != nil {
			return true
		}
	}
	
	// 回退到环境变量配置
	return s.cfg.SMTPHost != "" &&
		s.cfg.SMTPUser != "" &&
		s.cfg.SMTPPassword != "" &&
		s.cfg.SMTPFrom != ""
}

// getActiveSMTPConfig 获取激活的SMTP配置
func (s *EmailService) getActiveSMTPConfig() *models.SMTPConfig {
	if s.db == nil {
		return nil
	}
	
	var config models.SMTPConfig
	if err := s.db.Where("is_active = ?", true).First(&config).Error; err != nil {
		return nil
	}
	
	return &config
}

// decryptPassword 解密SMTP密码
func (s *EmailService) decryptPassword(encryptedPassword string) (string, error) {
	if s.crypto == nil {
		return "", fmt.Errorf("加密服务未初始化")
	}
	
	decryptedPassword, err := s.crypto.Decrypt(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("密码解密失败: %v", err)
	}
	
	return decryptedPassword, nil
}

// getBaseURL 获取基础URL，优先级：配置文件 > HTTP请求头 > 默认值
func (s *EmailService) getBaseURL(c *gin.Context) string {
	// 如果配置中已设置BASE_URL，直接使用
	if s.cfg.BaseURL != "" && !strings.Contains(s.cfg.BaseURL, "localhost") {
		return s.cfg.BaseURL
	}
	
	// 尝试从HTTP请求头获取域名信息
	if c != nil {
		// 检查X-Forwarded-Proto和X-Forwarded-Host（反向代理）
		proto := c.GetHeader("X-Forwarded-Proto")
		host := c.GetHeader("X-Forwarded-Host")
		
		if proto == "" {
			proto = "http"
			if c.Request.TLS != nil {
				proto = "https"
			}
		}
		
		if host == "" {
			host = c.GetHeader("Host")
		}
		
		if host != "" {
			return fmt.Sprintf("%s://%s", proto, host)
		}
	}
	
	// 回退到配置中的BaseURL
	return s.cfg.BaseURL
}

// sendEmail 发送邮件的核心功能
func (s *EmailService) sendEmail(to, subject, body string) error {
	// 获取SMTP配置（数据库优先，环境变量次之）
	var host, user, password, from string
	var port int
	var useTLS bool
	
	if dbConfig := s.getActiveSMTPConfig(); dbConfig != nil {
		// 使用数据库配置
		host = dbConfig.Host
		port = dbConfig.Port
		user = dbConfig.Username
		from = dbConfig.FromEmail
		useTLS = dbConfig.UseTLS
		
		// 解密密码（注意：实际应用中需要实现真正的解密）
		decryptedPassword, err := s.decryptPassword(dbConfig.Password)
		if err != nil {
			return fmt.Errorf("密码解密失败: %v", err)
		}
		password = decryptedPassword
	} else {
		// 回退到环境变量配置
		host = s.cfg.SMTPHost
		port = s.cfg.SMTPPort
		user = s.cfg.SMTPUser
		password = s.cfg.SMTPPassword
		from = s.cfg.SMTPFrom
		useTLS = (port == 587) // 默认587端口使用TLS
	}

	// 构建邮件内容
	message := s.buildEmailMessage(to, subject, body)

	// 设置认证
	auth := smtp.PlainAuth("", user, password, host)

	// SMTP服务器地址
	addr := fmt.Sprintf("%s:%d", host, port)

	// 如果需要TLS
	if useTLS || port == 587 {
		return s.sendEmailWithTLS(addr, auth, from, []string{to}, []byte(message), host)
	}

	// 标准SMTP发送
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
}

// sendEmailWithTLS 使用TLS发送邮件（适用于Gmail等）
func (s *EmailService) sendEmailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	// 创建客户端
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %v", err)
	}
	defer client.Close()

	// 启动TLS
	if err = client.StartTLS(&tls.Config{ServerName: host}); err != nil {
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
	return s.buildVerificationEmailBodyWithURL(email, token, s.cfg.BaseURL)
}

// buildVerificationEmailBodyWithURL 使用指定URL构建邮箱验证邮件内容
func (s *EmailService) buildVerificationEmailBodyWithURL(email, token, baseURL string) string {
	verifyURL := fmt.Sprintf("%s/api/verify-email/%s", baseURL, token)

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
	return s.buildPasswordResetEmailBodyWithURL(email, token, s.cfg.BaseURL)
}

// buildPasswordResetEmailBodyWithURL 使用指定URL构建密码重置邮件内容
func (s *EmailService) buildPasswordResetEmailBodyWithURL(email, token, baseURL string) string {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

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
