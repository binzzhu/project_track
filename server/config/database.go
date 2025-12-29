package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase() {
	// 确保数据库目录存在
	dbDir := filepath.Dir(DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatal("创建数据库目录失败:", err)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	log.Printf("数据库连接成功: %s", DBPath)
}

func GetDB() *gorm.DB {
	return DB
}
