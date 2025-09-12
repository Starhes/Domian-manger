package database

import (
	authmodels "domain-max/pkg/auth/models"
	dnsmodels "domain-max/pkg/dns/models"
	emailmodels "domain-max/pkg/email/models"
	"log"

	"gorm.io/gorm"
)

// Migrate 执行数据库迁移
func Migrate(db *gorm.DB) error {
	log.Println("开始数据库迁移...")
	
	// 用户相关表
	if err := db.AutoMigrate(
		&authmodels.User{},
		&authmodels.EmailVerification{},
		&authmodels.PasswordReset{},
	); err != nil {
		return err
	}
	
	// DNS相关表
	if err := db.AutoMigrate(
		&dnsmodels.Domain{},
		&dnsmodels.DNSRecord{},
		&dnsmodels.DNSProvider{},
	); err != nil {
		return err
	}
	
	// 邮件相关表
	if err := db.AutoMigrate(
		&emailmodels.SMTPConfig{},
	); err != nil {
		return err
	}
	
	log.Println("数据库迁移完成")
	return nil
}