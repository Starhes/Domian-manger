package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"domain-manager/internal/utils"
)

func main() {
	fmt.Println("=== 域名管理系统安全配置生成器 ===")
	fmt.Println()

	// 检查是否已存在.env文件
	envPath := ".env"
	if _, err := os.Stat(envPath); err == nil {
		fmt.Print("检测到已存在的.env文件，是否覆盖？(y/N): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.ToLower(scanner.Text()) != "y" {
			fmt.Println("配置生成已取消")
			return
		}
	}

	// 生成配置
	config, err := generateConfig()
	if err != nil {
		fmt.Printf("生成配置失败: %v\n", err)
		os.Exit(1)
	}

	// 写入.env文件
	if err := writeEnvFile(envPath, config); err != nil {
		fmt.Printf("写入配置文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 配置文件生成成功: .env")
	fmt.Println()
	fmt.Println("🔐 安全提醒:")
	fmt.Println("1. 请妥善保管生成的密钥，不要提交到代码仓库")
	fmt.Println("2. 生产环境请使用更强的密码和随机密钥")
	fmt.Println("3. 定期更新密钥以提高安全性")
	fmt.Println()
	fmt.Println("🚀 下一步:")
	fmt.Println("1. 检查并修改.env文件中的配置")
	fmt.Println("2. 运行 'make build' 构建应用")
	fmt.Println("3. 运行 'make run' 启动服务")
}

// ConfigData 配置数据
type ConfigData struct {
	Port          string
	Environment   string
	BaseURL       string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBType        string
	JWTSecret     string
	EncryptionKey string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPassword  string
	SMTPFrom      string
	DNSPodToken   string
}

// generateConfig 生成配置
func generateConfig() (*ConfigData, error) {
	scanner := bufio.NewScanner(os.Stdin)

	config := &ConfigData{}

	// 基本配置
	fmt.Print("请输入服务器端口 (默认: 8080): ")
	scanner.Scan()
	config.Port = getOrDefault(scanner.Text(), "8080")

	fmt.Print("请选择环境 (development/production，默认: development): ")
	scanner.Scan()
	config.Environment = getOrDefault(scanner.Text(), "development")

	if config.Environment == "production" {
		fmt.Print("请输入基础URL (如: https://yourdomain.com): ")
		scanner.Scan()
		config.BaseURL = strings.TrimSpace(scanner.Text())
	}

	// 数据库配置
	fmt.Println("\n--- 数据库配置 ---")
	fmt.Print("数据库主机 (默认: localhost): ")
	scanner.Scan()
	config.DBHost = getOrDefault(scanner.Text(), "localhost")

	fmt.Print("数据库端口 (默认: 5432): ")
	scanner.Scan()
	config.DBPort = getOrDefault(scanner.Text(), "5432")

	fmt.Print("数据库用户名 (默认: postgres): ")
	scanner.Scan()
	config.DBUser = getOrDefault(scanner.Text(), "postgres")

	fmt.Print("数据库名称 (默认: domain_manager): ")
	scanner.Scan()
	config.DBName = getOrDefault(scanner.Text(), "domain_manager")

	fmt.Print("数据库类型 (postgres/mysql，默认: postgres): ")
	scanner.Scan()
	config.DBType = getOrDefault(scanner.Text(), "postgres")

	// 生成安全密钥
	fmt.Println("\n--- 安全密钥生成 ---")
	fmt.Print("是否自动生成安全密钥？(Y/n): ")
	scanner.Scan()
	autoGenerate := getOrDefault(scanner.Text(), "Y")

	if strings.ToLower(autoGenerate) == "y" {
		// 自动生成密码
		password, err := utils.GenerateSecurePassword(16)
		if err != nil {
			return nil, fmt.Errorf("生成数据库密码失败: %v", err)
		}
		config.DBPassword = password

		// 生成JWT密钥
		jwtLength := 64
		if config.Environment == "production" {
			jwtLength = 128
		}
		jwtSecret, err := utils.GenerateJWTSecret(jwtLength)
		if err != nil {
			return nil, fmt.Errorf("生成JWT密钥失败: %v", err)
		}
		config.JWTSecret = jwtSecret

		// 生成加密密钥
		encryptionKey, err := utils.GenerateEncryptionKey()
		if err != nil {
			return nil, fmt.Errorf("生成加密密钥失败: %v", err)
		}
		config.EncryptionKey = encryptionKey

		fmt.Println("✅ 安全密钥生成完成")
	} else {
		// 手动输入
		fmt.Print("请输入数据库密码: ")
		scanner.Scan()
		config.DBPassword = strings.TrimSpace(scanner.Text())

		fmt.Print("请输入JWT密钥 (至少64个字符): ")
		scanner.Scan()
		config.JWTSecret = strings.TrimSpace(scanner.Text())

		fmt.Print("请输入加密密钥 (64个十六进制字符): ")
		scanner.Scan()
		config.EncryptionKey = strings.TrimSpace(scanner.Text())
	}

	// SMTP配置（可选）
	fmt.Println("\n--- SMTP邮件配置（可选）---")
	fmt.Print("是否配置SMTP邮件服务？(y/N): ")
	scanner.Scan()
	configureSMTP := strings.ToLower(scanner.Text())

	if configureSMTP == "y" {
		fmt.Print("SMTP服务器地址 (默认: smtp.gmail.com): ")
		scanner.Scan()
		config.SMTPHost = getOrDefault(scanner.Text(), "smtp.gmail.com")

		fmt.Print("SMTP端口 (默认: 587): ")
		scanner.Scan()
		config.SMTPPort = getOrDefault(scanner.Text(), "587")

		fmt.Print("SMTP用户名: ")
		scanner.Scan()
		config.SMTPUser = strings.TrimSpace(scanner.Text())

		fmt.Print("SMTP密码: ")
		scanner.Scan()
		config.SMTPPassword = strings.TrimSpace(scanner.Text())

		fmt.Print("发件人邮箱: ")
		scanner.Scan()
		config.SMTPFrom = strings.TrimSpace(scanner.Text())
	}

	// DNSPod配置（可选）
	fmt.Println("\n--- DNSPod配置（可选）---")
	fmt.Print("是否配置DNSPod Token？(y/N): ")
	scanner.Scan()
	configureDNSPod := strings.ToLower(scanner.Text())

	if configureDNSPod == "y" {
		fmt.Print("DNSPod Token: ")
		scanner.Scan()
		config.DNSPodToken = strings.TrimSpace(scanner.Text())
	}

	return config, nil
}

// writeEnvFile 写入.env文件
func writeEnvFile(path string, config *ConfigData) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入配置头部注释
	file.WriteString("# 域名管理系统配置文件\n")
	file.WriteString("# 🔐 请妥善保管此文件，不要提交到代码仓库\n")
	file.WriteString("# 生成时间: " + getCurrentTime() + "\n\n")

	// 服务器配置
	file.WriteString("# 服务器配置\n")
	file.WriteString(fmt.Sprintf("PORT=%s\n", config.Port))
	file.WriteString(fmt.Sprintf("ENVIRONMENT=%s\n", config.Environment))
	if config.BaseURL != "" {
		file.WriteString(fmt.Sprintf("BASE_URL=%s\n", config.BaseURL))
	}
	file.WriteString("\n")

	// 数据库配置
	file.WriteString("# 数据库配置\n")
	file.WriteString(fmt.Sprintf("DB_HOST=%s\n", config.DBHost))
	file.WriteString(fmt.Sprintf("DB_PORT=%s\n", config.DBPort))
	file.WriteString(fmt.Sprintf("DB_USER=%s\n", config.DBUser))
	file.WriteString(fmt.Sprintf("DB_PASSWORD=%s\n", config.DBPassword))
	file.WriteString(fmt.Sprintf("DB_NAME=%s\n", config.DBName))
	file.WriteString(fmt.Sprintf("DB_TYPE=%s\n", config.DBType))
	file.WriteString("\n")

	// 安全配置
	file.WriteString("# 安全配置\n")
	file.WriteString(fmt.Sprintf("JWT_SECRET=%s\n", config.JWTSecret))
	file.WriteString(fmt.Sprintf("ENCRYPTION_KEY=%s\n", config.EncryptionKey))
	file.WriteString("\n")

	// SMTP配置
	if config.SMTPHost != "" || config.SMTPUser != "" {
		file.WriteString("# SMTP邮件配置\n")
		if config.SMTPHost != "" {
			file.WriteString(fmt.Sprintf("SMTP_HOST=%s\n", config.SMTPHost))
		}
		if config.SMTPPort != "" {
			file.WriteString(fmt.Sprintf("SMTP_PORT=%s\n", config.SMTPPort))
		}
		if config.SMTPUser != "" {
			file.WriteString(fmt.Sprintf("SMTP_USER=%s\n", config.SMTPUser))
		}
		if config.SMTPPassword != "" {
			file.WriteString(fmt.Sprintf("SMTP_PASSWORD=%s\n", config.SMTPPassword))
		}
		if config.SMTPFrom != "" {
			file.WriteString(fmt.Sprintf("SMTP_FROM=%s\n", config.SMTPFrom))
		}
		file.WriteString("\n")
	}

	// DNSPod配置
	if config.DNSPodToken != "" {
		file.WriteString("# DNSPod配置\n")
		file.WriteString(fmt.Sprintf("DNSPOD_TOKEN=%s\n", config.DNSPodToken))
		file.WriteString("\n")
	}

	return nil
}

// getOrDefault 获取值或默认值
func getOrDefault(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// getCurrentTime 获取当前时间
func getCurrentTime() string {
	// 简化的时间格式，避免依赖time包
	return "generated by config tool"
}
