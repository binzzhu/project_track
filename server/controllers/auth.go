package controllers

import (
	"project-flow/config"
	"project-flow/middleware"
	"project-flow/models"
	"project-flow/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// Login 用户登录
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请输入用户名和密码")
		return
	}

	db := config.GetDB()
	var user models.User
	if err := db.Preload("Role").Where("username = ?", req.Username).First(&user).Error; err != nil {
		utils.Error(c, 401, "用户名或密码错误")
		return
	}

	// 检查账号是否被锁定
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		utils.Error(c, 403, "账号已被锁定，请稍后再试")
		return
	}

	// 检查账号状态
	if user.Status != 1 {
		utils.Error(c, 403, "账号已被禁用")
		return
	}

	// 验证密码
	if !utils.CheckPassword(user.Password, req.Password) {
		// 记录失败次数
		user.FailedLogins++
		if user.FailedLogins >= 5 {
			lockTime := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockTime
		}
		db.Save(&user)
		utils.Error(c, 401, "用户名或密码错误")
		return
	}

	// 重置失败次数
	user.FailedLogins = 0
	user.LockedUntil = nil
	db.Save(&user)

	// 生成Token
	roleCode := ""
	if user.Role != nil {
		roleCode = user.Role.Code
	}
	token, err := utils.GenerateToken(user.ID, user.Username, roleCode)
	if err != nil {
		utils.ServerError(c, "生成Token失败")
		return
	}

	// 记录登录日志
	go func() {
		log := models.OperationLog{
			UserID:      user.ID,
			Action:      "login",
			Module:      "auth",
			Description: user.Name + " 登录系统",
			Result:      "success",
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			CreatedAt:   time.Now(),
		}
		db.Create(&log)
	}()

	utils.Success(c, LoginResponse{
		Token: token,
		User:  user,
	})
}

// GetCurrentUser 获取当前用户信息
func (ac *AuthController) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("userID")

	db := config.GetDB()
	var user models.User
	if err := db.Preload("Role").First(&user, userID).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword 修改密码
func (ac *AuthController) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请输入完整信息")
		return
	}

	// 验证新密码复杂度
	if !utils.ValidatePassword(req.NewPassword) {
		utils.BadRequest(c, "密码至少8位，需包含大小写字母、数字和特殊符号")
		return
	}

	userID, _ := c.Get("userID")
	db := config.GetDB()
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	// 验证旧密码
	if !utils.CheckPassword(user.Password, req.OldPassword) {
		utils.Error(c, 400, "原密码错误")
		return
	}

	// 更新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.ServerError(c, "密码加密失败")
		return
	}

	user.Password = hashedPassword
	db.Save(&user)

	// 记录日志
	middleware.LogOperation(c, "change_password", "auth", "user", user.ID, user.Name, "修改密码", "success")

	utils.SuccessWithMessage(c, "密码修改成功", nil)
}

// Logout 用户退出
func (ac *AuthController) Logout(c *gin.Context) {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	// 记录退出日志
	go func() {
		log := models.OperationLog{
			UserID:      userID.(uint),
			Action:      "logout",
			Module:      "auth",
			Description: username.(string) + " 退出系统",
			Result:      "success",
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			CreatedAt:   time.Now(),
		}
		config.GetDB().Create(&log)
	}()

	utils.SuccessWithMessage(c, "退出成功", nil)
}
