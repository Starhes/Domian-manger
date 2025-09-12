package database

import (
	"domain-max/pkg/config"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect 连接数据库
func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	
	switch cfg.DBType {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		dialector = mysql.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.DBType)
	}
	
	// 配置GORM日志级别
	logLevel := logger.Info
	if cfg.Environment == "production" {
		logLevel = logger.Error
	}
	
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}
	
	// 获取底层sql.DB以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接池失败: %v", err)
	}
	
	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	
	log.Println("数据库连接成功")
	return db, nil
}