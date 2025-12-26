package middleware

import (
	"project-flow/config"
	"project-flow/models"
	"time"

	"github.com/gin-gonic/gin"
)

// OperationLogMiddleware 操作日志中间件
func OperationLogMiddleware(action, module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 记录操作日志
		userID, exists := c.Get("userID")
		if !exists {
			return
		}

		log := models.OperationLog{
			UserID:    userID.(uint),
			Action:    action,
			Module:    module,
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Result:    "success",
			CreatedAt: time.Now(),
		}

		// 异步写入日志
		go func() {
			config.GetDB().Create(&log)
		}()
	}
}

// LogOperation 记录操作日志
func LogOperation(c *gin.Context, action, module, targetType string, targetID uint, targetName, description, result string) {
	userID, exists := c.Get("userID")
	if !exists {
		return
	}

	log := models.OperationLog{
		UserID:      userID.(uint),
		Action:      action,
		Module:      module,
		TargetType:  targetType,
		TargetID:    targetID,
		TargetName:  targetName,
		Description: description,
		Result:      result,
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		CreatedAt:   time.Now(),
	}

	config.GetDB().Create(&log)
}
