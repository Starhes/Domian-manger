package constants

import "time"

// 时间常量
const (
	// JWT令牌有效期
	JWTTokenExpiration = 24 * time.Hour // 24小时
	
	// 刷新令牌有效期
	RefreshTokenExpiration = 30 * 24 * time.Hour // 30天
	
	// 邮箱验证令牌有效期
	EmailVerificationExpiration = 24 * time.Hour // 24小时
	
	// 密码重置令牌有效期
	PasswordResetExpiration = 1 * time.Hour // 1小时
	
	// CSRF令牌有效期
	CSRFTokenExpiration = 24 * time.Hour // 24小时
	
	// 令牌清理间隔
	TokenCleanupInterval = 1 * time.Hour // 每小时清理一次
	
	// 数据库连接超时
	DBConnectionTimeout = 30 * time.Second
	
	// HTTP请求超时
	HTTPRequestTimeout = 10 * time.Second
)

// 数据库常量
const (
	// 默认分页大小
	DefaultPageSize = 10
	
	// 最大分页大小
	MaxPageSize = 100
	
	// 最大页码
	MaxPageNumber = 10000
	
	// 默认DNS记录配额
	DefaultDNSRecordQuota = 10
	
	// 最大DNS记录配额
	MaxDNSRecordQuota = 1000
)

// 验证常量
const (
	// 最小密码长度
	MinPasswordLength = 8
	
	// 最大密码长度
	MaxPasswordLength = 128
	
	// 最小JWT密钥长度
	MinJWTSecretLength = 32
	
	// 生产环境最小JWT密钥长度
	MinJWTSecretLengthProduction = 64
	
	// 最大邮箱长度
	MaxEmailLength = 255
	
	// 最大昵称长度
	MaxNicknameLength = 100
	
	// 最大搜索查询长度
	MaxSearchQueryLength = 100
	
	// 最大DNS记录值长度
	MaxDNSRecordValueLength = 500
	
	// 最大域名长度
	MaxDomainNameLength = 253
	
	// 最大子域名长度
	MaxSubdomainLength = 63
)

// HTTP状态常量
const (
	// 请求头名称
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
	HeaderUserAgent     = "User-Agent"
	HeaderXForwardedFor = "X-Forwarded-For"
	HeaderXRealIP       = "X-Real-IP"
	HeaderXCSRFToken    = "X-CSRF-Token"
	HeaderXDevelopment  = "X-Development"
	
	// Content-Type值
	ContentTypeJSON = "application/json"
	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeHTML = "text/html"
)

// 错误消息常量
const (
	// 通用错误
	ErrMsgInternalError     = "服务器内部错误，请稍后再试"
	ErrMsgInvalidRequest    = "请求参数无效"
	ErrMsgUnauthorized      = "未授权访问"
	ErrMsgForbidden         = "权限不足"
	ErrMsgNotFound          = "资源不存在"
	ErrMsgConflict          = "资源冲突"
	ErrMsgTooManyRequests   = "请求过于频繁，请稍后再试"
	
	// 认证错误
	ErrMsgInvalidCredentials = "邮箱或密码错误"
	ErrMsgAccountNotActive   = "账户未激活，请检查邮箱验证链接"
	ErrMsgAccountSuspended   = "账户已被暂停，请联系管理员"
	ErrMsgAccountBanned      = "账户已被封禁，请联系管理员"
	ErrMsgTokenInvalid       = "令牌无效或已过期"
	ErrMsgTokenRevoked       = "令牌已被撤销"
	ErrMsgCSRFTokenMissing   = "CSRF令牌缺失"
	ErrMsgCSRFTokenInvalid   = "CSRF令牌无效"
	
	// 验证错误
	ErrMsgEmailInvalid      = "邮箱格式不正确"
	ErrMsgEmailExists       = "该邮箱已被注册"
	ErrMsgPasswordWeak      = "密码强度不够"
	ErrMsgPasswordMismatch  = "两次输入的密码不一致"
	ErrMsgNicknameInvalid   = "昵称格式不正确"
	ErrMsgUserIDInvalid     = "用户ID格式不正确"
	ErrMsgPageInvalid       = "页码格式不正确"
	ErrMsgSearchInvalid     = "搜索关键词包含不安全字符"
	
	// DNS相关错误
	ErrMsgDNSRecordInvalid    = "DNS记录格式不正确"
	ErrMsgDNSRecordExists     = "DNS记录已存在"
	ErrMsgDNSRecordQuotaExceeded = "DNS记录数量已达到配额上限"
	ErrMsgDomainInvalid       = "域名格式不正确"
	ErrMsgSubdomainInvalid    = "子域名格式不正确"
	ErrMsgIPAddressInvalid    = "IP地址格式不正确"
	ErrMsgPrivateIPNotAllowed = "不允许使用私有网络IP地址"
	
	// 成功消息
	MsgLoginSuccess           = "登录成功"
	MsgLogoutSuccess          = "登出成功"
	MsgRegisterSuccess        = "注册成功，请检查邮箱激活账户"
	MsgEmailVerifySuccess     = "邮箱验证成功，账户已激活"
	MsgPasswordResetSuccess   = "密码重置成功"
	MsgProfileUpdateSuccess   = "资料更新成功"
	MsgDNSRecordCreateSuccess = "DNS记录创建成功"
	MsgDNSRecordUpdateSuccess = "DNS记录更新成功"
	MsgDNSRecordDeleteSuccess = "DNS记录删除成功"
)

// 业务常量
const (
	// 用户状态
	UserStatusNormal    = "normal"
	UserStatusSuspended = "suspended"
	UserStatusBanned    = "banned"
	
	// DNS记录状态
	DNSRecordStatusActive   = "active"
	DNSRecordStatusInactive = "inactive"
	DNSRecordStatusPending  = "pending"
	
	// DNS记录类型
	DNSTypeA     = "A"
	DNSTypeAAAA  = "AAAA"
	DNSTypeCNAME = "CNAME"
	DNSTypeMX    = "MX"
	DNSTypeTXT   = "TXT"
	DNSTypeNS    = "NS"
	DNSTypePTR   = "PTR"
	DNSTypeSRV   = "SRV"
	
	// 环境类型
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTesting     = "testing"
	
	// 数据库类型
	DBTypePostgres = "postgres"
	DBTypeMySQL    = "mysql"
)

// 正则表达式常量
const (
	// 邮箱格式验证
	RegexEmail = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	
	// 子域名格式验证
	RegexSubdomain = `^[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?$`
	
	// 域名格式验证
	RegexDomain = `^[a-zA-Z0-9]([a-zA-Z0-9\-\.]*[a-zA-Z0-9])?$`
	
	// IPv4地址验证
	RegexIPv4 = `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	
	// IPv6地址验证（简化版）
	RegexIPv6 = `^[0-9a-fA-F:]+$`
	
	// 十六进制字符串验证
	RegexHex = `^[0-9a-fA-F]+$`
	
	// 用户ID验证（纯数字）
	RegexUserID = `^[0-9]+$`
)

// 配置默认值
const (
	// 服务器默认配置
	DefaultPort        = "8080"
	DefaultEnvironment = EnvDevelopment
	
	// 数据库默认配置
	DefaultDBHost = "localhost"
	DefaultDBPort = "5432"
	DefaultDBUser = "postgres"
	DefaultDBName = "domain_manager"
	DefaultDBType = DBTypePostgres
	
	// SMTP默认配置
	DefaultSMTPHost = "smtp.gmail.com"
	DefaultSMTPPort = 587
	DefaultSMTPFrom = "noreply@example.com"
	
	// 安全默认配置
	DefaultBcryptCost = 12 // bcrypt默认成本
)

// 速率限制常量
const (
	// 登录速率限制：每分钟最多5次尝试
	LoginRateLimit = 5
	LoginRateWindow = 1 * time.Minute
	
	// 注册速率限制：每小时最多3次注册
	RegisterRateLimit = 3
	RegisterRateWindow = 1 * time.Hour
	
	// API通用速率限制：每分钟最多100次请求
	APIRateLimit = 100
	APIRateWindow = 1 * time.Minute
	
	// DNS记录操作限制：每分钟最多10次
	DNSOperationRateLimit = 10
	DNSOperationRateWindow = 1 * time.Minute
)

// 文件大小常量
const (
	MaxUploadSize = 10 << 20 // 10MB
	MaxJSONSize   = 1 << 20  // 1MB
	MaxLogSize    = 100 << 20 // 100MB
)

// 缓存常量
const (
	CacheKeyUserSession     = "user_session:"
	CacheKeyLoginAttempts   = "login_attempts:"
	CacheKeyRegisterAttempts = "register_attempts:"
	CacheKeyAPIRateLimit    = "api_rate_limit:"
	CacheKeyDNSOperations   = "dns_operations:"
	
	// 缓存过期时间
	CacheExpiryUserSession = 30 * time.Minute
	CacheExpiryRateLimit   = 1 * time.Hour
)

// 日志常量
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)
