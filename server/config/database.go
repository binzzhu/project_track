package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		DBUser,
		DBPassword,
		DBHost,
		DBPort,
		DBName,
	)

	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			log.Printf("数据库连接成功: %s:%s/%s", DBHost, DBPort, DBName)
			return
		}
		log.Printf("数据库连接失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		log.Fatal("数据库连接失败（已达最大重试次数）:", err)
	}
	log.Printf("数据库连接成功: %s:%s/%s", DBHost, DBPort, DBName)
}

func GetDB() *gorm.DB {
	return DB
}
