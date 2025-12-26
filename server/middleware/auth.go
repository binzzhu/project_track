package middleware

import (
	"project-flow/config"
	"project-flow/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 优先从 Header 获取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Bearer token格式
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		// 如果 Header 没有，从 Query 参数获取（用于文件下载等场景）
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			utils.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			utils.Unauthorized(c, "Token无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roleCode", claims.RoleCode)
		c.Next()
	}
}

// RoleMiddleware 角色权限中间件
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode, exists := c.Get("roleCode")
		if !exists {
			utils.Unauthorized(c, "未获取到用户角色信息")
			c.Abort()
			return
		}

		// 管理员拥有所有权限
		if roleCode == config.RoleAdmin {
			c.Next()
			return
		}

		// 检查角色是否在允许列表中
		allowed := false
		for _, role := range allowedRoles {
			if roleCode == role {
				allowed = true
				break
			}
		}

		if !allowed {
			utils.Forbidden(c, "没有权限执行此操作")
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
