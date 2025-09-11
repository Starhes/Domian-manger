package models

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 用户模型
type User struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Email          string         `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Password       string         `json:"-" gorm:"not null;size:255"` // bcrypt哈希后的密码
	Nickname       string         `json:"nickname" gorm:"size:100"`   // 用户昵称
	Avatar         string         `json:"avatar" gorm:"size:500"`     // 头像URL
	IsActive       bool           `json:"is_active" gorm:"default:false;index"`
	IsAdmin        bool           `json:"is_admin" gorm:"default:false;index"`
	LastLoginAt    *time.Time     `json:"last_login_at"`                        // 最后登录时间
	LoginCount     int            `json:"login_count" gorm:"default:0"`         // 登录次数
	DNSRecordQuota int            `json:"dns_record_quota" gorm:"default:10"`   // DNS记录配额
	Status         string         `json:"status" gorm:"default:normal;size:20"` // 用户状态：normal, suspended, banned
	CreatedAt      time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	DNSRecords []DNSRecord `json:"dns_records,omitempty" gorm:"foreignKey:UserID"`
}

// 域名模型
type Domain struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"` // 主域名，如 example.com
	DomainType  string         `json:"domain_type"`                      // 域名类型：二级域名、三级域名等
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	Description string         `json:"description"` // 域名描述
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	DNSRecords []DNSRecord `json:"dns_records,omitempty" gorm:"foreignKey:DomainID"`
}

// DNS记录模型
type DNSRecord struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"not null;index"`                           // 添加索引提升查询性能
	DomainID   uint           `json:"domain_id" gorm:"not null;index"`                         // 添加索引
	Subdomain  string         `json:"subdomain" gorm:"not null;size:63"`                       // 子域名，限制长度
	Type       string         `json:"type" gorm:"not null;size:10;index"`                      // 记录类型，添加索引
	Value      string         `json:"value" gorm:"not null;size:500"`                          // 记录值，限制长度
	TTL        int            `json:"ttl" gorm:"default:600;check:ttl >= 1 AND ttl <= 604800"` // TTL范围检查
	Priority   int            `json:"priority" gorm:"default:0"`                               // MX记录优先级
	ExternalID string         `json:"external_id" gorm:"size:100"`                             // DNS服务商记录ID
	Status     string         `json:"status" gorm:"default:active;size:20"`                    // 记录状态
	Comment    string         `json:"comment" gorm:"size:255"`                                 // 记录备注
	CreatedAt  time.Time      `json:"created_at" gorm:"index"`                                 // 添加时间索引
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Domain Domain `json:"domain,omitempty" gorm:"foreignKey:DomainID;constraint:OnDelete:CASCADE"`
}

// DNS服务商模型
type DNSProvider struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100;uniqueIndex"` // 服务商名称唯一
	Type        string         `json:"type" gorm:"not null;size:50;index"`        // 服务商类型索引
	Config      string         `json:"config" gorm:"type:text"`                   // JSON格式配置
	IsActive    bool           `json:"is_active" gorm:"default:false;index"`      // 默认禁用，添加索引
	Description string         `json:"description" gorm:"size:500"`               // 服务商描述
	SortOrder   int            `json:"sort_order" gorm:"default:0"`               // 排序字段
	LastTestAt  *time.Time     `json:"last_test_at"`                              // 最后测试时间
	TestResult  string         `json:"test_result" gorm:"size:1000"`              // 测试结果
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// 邮箱验证模型
type EmailVerification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// 密码重置模型
type PasswordReset struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// SMTP配置模型
type SMTPConfig struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100"`                    // 配置名称
	Host        string         `json:"host" gorm:"not null;size:255"`                    // SMTP服务器地址
	Port        int            `json:"port" gorm:"not null;default:587"`                 // SMTP端口
	Username    string         `json:"username" gorm:"not null;size:255"`                // 用户名
	Password    string         `json:"-" gorm:"not null;size:255"`                       // 密码，不返回给前端
	FromEmail   string         `json:"from_email" gorm:"not null;size:255"`              // 发件人邮箱
	FromName    string         `json:"from_name" gorm:"size:100"`                        // 发件人名称
	IsActive    bool           `json:"is_active" gorm:"default:false;index"`             // 是否启用
	IsDefault   bool           `json:"is_default" gorm:"default:false"`                  // 是否为默认配置
	UseTLS      bool           `json:"use_tls" gorm:"default:true"`                      // 是否使用TLS
	Description string         `json:"description" gorm:"size:500"`                      // 配置描述
	LastTestAt  *time.Time     `json:"last_test_at"`                                     // 最后测试时间
	TestResult  string         `json:"test_result" gorm:"size:1000"`                     // 测试结果
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// 请求和响应结构体

// 用户注册请求
type RegisterRequest struct {
	Email           string `json:"email" binding:"required,email,max=255"`
	Password        string `json:"password" binding:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	Nickname        string `json:"nickname" binding:"max=100"`
}

// 用户登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// 登录响应
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User         User   `json:"user"`
}

// DNS记录创建请求
type CreateDNSRecordRequest struct {
	DomainID  uint   `json:"domain_id" binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=A CNAME TXT MX"`
	Value     string `json:"value" binding:"required"`
	TTL       int    `json:"ttl"`
}

// DNS记录更新请求
type UpdateDNSRecordRequest struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type" binding:"oneof=A CNAME TXT MX"`
	Value     string `json:"value"`
	TTL       int    `json:"ttl"`
}

// 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// 重置密码请求
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// SMTP配置创建请求
type CreateSMTPConfigRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Host        string `json:"host" binding:"required,max=255"`
	Port        int    `json:"port" binding:"required,min=1,max=65535"`
	Username    string `json:"username" binding:"required,max=255"`
	Password    string `json:"password" binding:"required,max=255"`
	FromEmail   string `json:"from_email" binding:"required,email,max=255"`
	FromName    string `json:"from_name" binding:"max=100"`
	UseTLS      bool   `json:"use_tls"`
	Description string `json:"description" binding:"max=500"`
}

// SMTP配置更新请求
type UpdateSMTPConfigRequest struct {
	Name        string `json:"name" binding:"max=100"`
	Host        string `json:"host" binding:"max=255"`
	Port        int    `json:"port" binding:"min=1,max=65535"`
	Username    string `json:"username" binding:"max=255"`
	Password    string `json:"password" binding:"max=255"` // 可选，为空则不更新密码
	FromEmail   string `json:"from_email" binding:"email,max=255"`
	FromName    string `json:"from_name" binding:"max=100"`
	UseTLS      *bool  `json:"use_tls"` // 使用指针以区分false和未设置
	Description string `json:"description" binding:"max=500"`
}

// SMTP配置测试请求
type TestSMTPConfigRequest struct {
	ToEmail string `json:"to_email" binding:"required,email"`
}

// SMTP配置响应（脱敏版本）
type SMTPConfigResponse struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Host        string     `json:"host"`
	Port        int        `json:"port"`
	Username    string     `json:"username"`
	FromEmail   string     `json:"from_email"`
	FromName    string     `json:"from_name"`
	IsActive    bool       `json:"is_active"`
	IsDefault   bool       `json:"is_default"`
	UseTLS      bool       `json:"use_tls"`
	Description string     `json:"description"`
	LastTestAt  *time.Time `json:"last_test_at"`
	TestResult  string     `json:"test_result"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ========================= 验证方法 =========================

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码长度至少8位")
	}
	
	if len(password) > 100 {
		return errors.New("密码长度不能超过100位")
	}
	
	// 检查是否包含字母
	hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, password)
	if !hasLetter {
		return errors.New("密码必须包含字母")
	}
	
	// 检查是否包含数字
	hasNumber, _ := regexp.MatchString(`[0-9]`, password)
	if !hasNumber {
		return errors.New("密码必须包含数字")
	}
	
	// 检查是否包含特殊字符（可选，但推荐）
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]`, password)
	
	// 如果密码长度小于12，则必须包含特殊字符
	if len(password) < 12 && !hasSpecial {
		return errors.New("密码长度小于12位时必须包含特殊字符")
	}
	
	return nil
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) error {
	if len(email) == 0 {
		return errors.New("邮箱不能为空")
	}
	
	if len(email) > 255 {
		return errors.New("邮箱长度不能超过255位")
	}
	
	// 基本的邮箱格式检查
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("邮箱格式不正确")
	}
	
	// 检查是否包含常见的不安全字符
	unsafeChars := []string{"<", ">", "\"", "'", "&", ";", "|", "`", "$"}
	for _, char := range unsafeChars {
		if strings.Contains(email, char) {
			return errors.New("邮箱包含不安全字符")
		}
	}
	
	return nil
}

// ValidateNickname 验证用户昵称
func ValidateNickname(nickname string) error {
	if len(nickname) > 100 {
		return errors.New("昵称长度不能超过100位")
	}
	
	if len(nickname) > 0 {
		// 检查是否包含不安全字符
		unsafeChars := []string{"<", ">", "\"", "'", "&", ";", "|", "`", "$", "\\"}
		for _, char := range unsafeChars {
			if strings.Contains(nickname, char) {
				return errors.New("昵称包含不安全字符")
			}
		}
		
		// 去除首尾空格
		nickname = strings.TrimSpace(nickname)
		if len(nickname) == 0 {
			return errors.New("昵称不能只包含空格")
		}
	}
	
	return nil
}

// Validate 验证注册请求
func (req *RegisterRequest) Validate() error {
	// 验证邮箱
	if err := ValidateEmail(req.Email); err != nil {
		return err
	}
	
	// 验证密码
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	
	// 验证确认密码
	if req.Password != req.ConfirmPassword {
		return errors.New("两次输入的密码不一致")
	}
	
	// 验证昵称
	if err := ValidateNickname(req.Nickname); err != nil {
		return err
	}
	
	return nil
}

// ========================= DNS记录验证方法 =========================

// ValidateDNSRecord 验证DNS记录的有效性
func (r *DNSRecord) ValidateDNSRecord() error {
	// 验证子域名格式
	if err := validateSubdomain(r.Subdomain); err != nil {
		return fmt.Errorf("子域名格式错误: %v", err)
	}
	
	// 验证记录类型
	if err := validateDNSType(r.Type); err != nil {
		return fmt.Errorf("DNS记录类型错误: %v", err)
	}
	
	// 根据记录类型验证记录值
	if err := validateDNSValue(r.Type, r.Value); err != nil {
		return fmt.Errorf("DNS记录值错误: %v", err)
	}
	
	// 验证TTL值
	if err := validateTTL(r.TTL); err != nil {
		return fmt.Errorf("TTL值错误: %v", err)
	}
	
	// 如果是MX记录，验证优先级
	if r.Type == "MX" && (r.Priority < 0 || r.Priority > 65535) {
		return fmt.Errorf("MX记录优先级必须在0-65535之间")
	}
	
	return nil
}

// validateSubdomain 验证子域名格式
func validateSubdomain(subdomain string) error {
	if len(subdomain) == 0 {
		return errors.New("子域名不能为空")
	}
	
	if len(subdomain) > 63 {
		return errors.New("子域名长度不能超过63个字符")
	}
	
	// 检查是否包含非法字符
	validSubdomainPattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?$`)
	if !validSubdomainPattern.MatchString(subdomain) {
		return errors.New("子域名只能包含字母、数字和连字符，且不能以连字符开头或结尾")
	}
	
	// 检查是否包含不安全的保留字
	reservedNames := []string{"www", "mail", "ftp", "admin", "administrator", "root", "api", "dns", "ns1", "ns2"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(subdomain, reserved) {
			return fmt.Errorf("子域名 '%s' 是保留名称，请使用其他名称", subdomain)
		}
	}
	
	return nil
}

// validateDNSType 验证DNS记录类型
func validateDNSType(dnsType string) error {
	validTypes := map[string]bool{
		"A":     true,
		"AAAA":  true,
		"CNAME": true,
		"MX":    true,
		"TXT":   true,
		"NS":    true,
		"PTR":   true,
		"SRV":   true,
	}
	
	if !validTypes[strings.ToUpper(dnsType)] {
		return fmt.Errorf("不支持的DNS记录类型: %s", dnsType)
	}
	
	return nil
}

// validateDNSValue 根据DNS记录类型验证记录值
func validateDNSValue(dnsType, value string) error {
	if len(value) == 0 {
		return errors.New("DNS记录值不能为空")
	}
	
	if len(value) > 500 {
		return errors.New("DNS记录值长度不能超过500个字符")
	}
	
	switch strings.ToUpper(dnsType) {
	case "A":
		return validateIPv4Address(value)
	case "AAAA":
		return validateIPv6Address(value)
	case "CNAME":
		return validateDomainName(value)
	case "MX":
		return validateMXRecord(value)
	case "TXT":
		return validateTXTRecord(value)
	case "NS":
		return validateDomainName(value)
	case "PTR":
		return validateDomainName(value)
	case "SRV":
		return validateSRVRecord(value)
	default:
		// 对于其他类型，只进行基本的长度检查
		return nil
	}
}

// validateIPv4Address 验证IPv4地址
func validateIPv4Address(ip string) error {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return errors.New("IPv4地址格式不正确")
	}
	
	for _, part := range parts {
		if num, err := strconv.Atoi(part); err != nil || num < 0 || num > 255 {
			return errors.New("IPv4地址格式不正确")
		}
	}
	
	// 检查私有网络地址（安全考虑）
	if isPrivateIP(ip) {
		return errors.New("不允许使用私有网络地址")
	}
	
	return nil
}

// validateIPv6Address 验证IPv6地址
func validateIPv6Address(ip string) error {
	// 简单的IPv6格式检查
	if strings.Count(ip, ":") < 2 || strings.Count(ip, ":") > 7 {
		return errors.New("IPv6地址格式不正确")
	}
	
	// 检查是否包含非法字符
	validIPv6Pattern := regexp.MustCompile(`^[0-9a-fA-F:]+$`)
	if !validIPv6Pattern.MatchString(ip) {
		return errors.New("IPv6地址包含非法字符")
	}
	
	return nil
}

// validateDomainName 验证域名格式
func validateDomainName(domain string) error {
	if len(domain) == 0 {
		return errors.New("域名不能为空")
	}
	
	if len(domain) > 253 {
		return errors.New("域名长度不能超过253个字符")
	}
	
	// 检查域名格式
	domainPattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-\.]*[a-zA-Z0-9])?$`)
	if !domainPattern.MatchString(domain) {
		return errors.New("域名格式不正确")
	}
	
	// 检查是否以点结尾（完全限定域名）
	if strings.HasSuffix(domain, ".") {
		domain = domain[:len(domain)-1] // 移除尾部的点进行验证
	}
	
	// 验证每个标签
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return errors.New("域名标签长度必须在1-63个字符之间")
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return errors.New("域名标签不能以连字符开头或结尾")
		}
	}
	
	return nil
}

// validateMXRecord 验证MX记录值
func validateMXRecord(value string) error {
	// MX记录值格式：优先级 邮件服务器域名
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return errors.New("MX记录格式应为：优先级 邮件服务器域名")
	}
	
	// 验证优先级
	priority, err := strconv.Atoi(parts[0])
	if err != nil || priority < 0 || priority > 65535 {
		return errors.New("MX记录优先级必须是0-65535之间的数字")
	}
	
	// 验证邮件服务器域名
	return validateDomainName(parts[1])
}

// validateTXTRecord 验证TXT记录值
func validateTXTRecord(value string) error {
	// TXT记录可以包含任意文本，但需要检查长度和一些安全问题
	if len(value) > 255 {
		return errors.New("单个TXT记录长度不能超过255个字符")
	}
	
	// 检查是否包含危险的脚本标签
	dangerousPatterns := []string{"<script", "javascript:", "vbscript:", "onload=", "onerror="}
	lowerValue := strings.ToLower(value)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerValue, pattern) {
			return errors.New("TXT记录不能包含脚本内容")
		}
	}
	
	return nil
}

// validateSRVRecord 验证SRV记录值  
func validateSRVRecord(value string) error {
	// SRV记录格式：优先级 权重 端口 目标域名
	parts := strings.Fields(value)
	if len(parts) != 4 {
		return errors.New("SRV记录格式应为：优先级 权重 端口 目标域名")
	}
	
	// 验证优先级
	if priority, err := strconv.Atoi(parts[0]); err != nil || priority < 0 || priority > 65535 {
		return errors.New("SRV记录优先级必须是0-65535之间的数字")
	}
	
	// 验证权重
	if weight, err := strconv.Atoi(parts[1]); err != nil || weight < 0 || weight > 65535 {
		return errors.New("SRV记录权重必须是0-65535之间的数字")
	}
	
	// 验证端口
	if port, err := strconv.Atoi(parts[2]); err != nil || port < 1 || port > 65535 {
		return errors.New("SRV记录端口必须是1-65535之间的数字")
	}
	
	// 验证目标域名
	return validateDomainName(parts[3])
}

// validateTTL 验证TTL值
func validateTTL(ttl int) error {
	if ttl < 1 {
		return errors.New("TTL值不能小于1秒")
	}
	
	if ttl > 604800 { // 7天
		return errors.New("TTL值不能超过604800秒（7天）")
	}
	
	return nil
}

// isPrivateIP 检查是否是私有IP地址
func isPrivateIP(ip string) bool {
	privateRanges := []string{
		"10.0.0.0/8",     // 10.0.0.0 - 10.255.255.255
		"172.16.0.0/12",  // 172.16.0.0 - 172.31.255.255
		"192.168.0.0/16", // 192.168.0.0 - 192.168.255.255
		"127.0.0.0/8",    // 127.0.0.0 - 127.255.255.255 (localhost)
		"169.254.0.0/16", // 169.254.0.0 - 169.254.255.255 (link-local)
	}
	
	for _, cidr := range privateRanges {
		if isIPInCIDR(ip, cidr) {
			return true
		}
	}
	
	return false
}

// isIPInCIDR 检查IP是否在CIDR范围内
func isIPInCIDR(ip, cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}
	
	return network.Contains(ipAddr)
}
