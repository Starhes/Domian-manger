package database

import (
	"domain-manager/internal/config"
	"domain-manager/internal/models"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dsn string
	var dialector gorm.Dialector

	switch cfg.DBType {
	case "postgres":
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		dialector = mysql.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.DBType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Domain{},
		&models.DNSRecord{},
		&models.DNSProvider{},
		&models.EmailVerification{},
		&models.PasswordReset{},
		&models.SMTPConfig{},
	)
}
