package db

import (
	"douyin-mall/configs"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMySQLClient 创建MySQL数据库连接
func NewMySQLClient(config *configs.MySQLConfig) (*gorm.DB, error) {
	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	// 创建数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect to mysql failed: %v", err)
	}

	return db, nil
}
