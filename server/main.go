package main

import (
	"log"
	"os"
	"project-flow/config"
	"project-flow/middleware"
	"project-flow/models"
	"project-flow/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	config.InitDatabase()

	// 自动迁移数据库表
	models.AutoMigrate()

	// 初始化默认数据
	models.InitDefaultData()

	// 确保上传目录存在
	os.MkdirAll(config.UploadPath, 0755)

	// 创建Gin实例
	r := gin.Default()

	// 跨域中间件
	r.Use(middleware.CORSMiddleware())

	// 静态文件服务
	r.Static("/uploads", config.UploadPath)

	// 设置路由
	routes.SetupRoutes(r)

	// 启动服务器
	log.Printf("服务器启动在 http://localhost%s", config.ServerPort)
	if err := r.Run(config.ServerPort); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
