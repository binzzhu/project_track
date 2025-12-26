package controllers

import (
	"project-flow/config"
	"project-flow/models"
	"project-flow/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LogController struct{}

// List 获取操作日志列表
func (lc *LogController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userIDParam := c.Query("user_id")
	action := c.Query("action")
	module := c.Query("module")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	targetType := c.Query("target_type")
	keyword := c.Query("keyword")

	db := config.GetDB()
	roleCode, _ := c.Get("roleCode")
	currentUserID, _ := c.Get("userID")

	var logs []models.OperationLog
	var total int64

	query := db.Model(&models.OperationLog{}).Preload("User")

	// 管理员和部门经理可以查看所有日志，其他角色只看自己的
	if roleCode != config.RoleAdmin && roleCode != config.RoleDeptManager {
		query = query.Where("user_id = ?", currentUserID)
	}

	if userIDParam != "" {
		query = query.Where("user_id = ?", userIDParam)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}
	if keyword != "" {
		query = query.Where("description LIKE ? OR target_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&logs)

	utils.SuccessPage(c, logs, total, page, pageSize)
}

// GetActions 获取操作类型列表
func (lc *LogController) GetActions(c *gin.Context) {
	actions := []map[string]string{
		{"value": "login", "label": "登录"},
		{"value": "logout", "label": "退出"},
		{"value": "create", "label": "创建"},
		{"value": "update", "label": "更新"},
		{"value": "delete", "label": "删除"},
		{"value": "upload", "label": "上传"},
		{"value": "download", "label": "下载"},
		{"value": "review", "label": "审核"},
		{"value": "archive", "label": "归档"},
		{"value": "add_member", "label": "添加成员"},
		{"value": "remove_member", "label": "移除成员"},
		{"value": "update_phase", "label": "更新阶段"},
		{"value": "update_status", "label": "更新状态"},
		{"value": "change_password", "label": "修改密码"},
		{"value": "reset_password", "label": "重置密码"},
		{"value": "new_version", "label": "上传新版本"},
	}
	utils.Success(c, actions)
}

// GetModules 获取模块列表
func (lc *LogController) GetModules(c *gin.Context) {
	modules := []map[string]string{
		{"value": "auth", "label": "认证"},
		{"value": "user", "label": "用户管理"},
		{"value": "project", "label": "项目管理"},
		{"value": "task", "label": "任务管理"},
		{"value": "document", "label": "文档管理"},
		{"value": "knowledge", "label": "知识库"},
		{"value": "contract", "label": "合同管理"},
	}
	utils.Success(c, modules)
}

// GetStatistics 获取日志统计
func (lc *LogController) GetStatistics(c *gin.Context) {
	db := config.GetDB()

	// 今日操作数
	var todayCount int64
	db.Model(&models.OperationLog{}).Where("DATE(created_at) = DATE('now')").Count(&todayCount)

	// 本周操作数
	var weekCount int64
	db.Model(&models.OperationLog{}).Where("created_at >= DATE('now', '-7 days')").Count(&weekCount)

	// 按模块统计
	type ModuleCount struct {
		Module string `json:"module"`
		Count  int64  `json:"count"`
	}
	var moduleCounts []ModuleCount
	db.Model(&models.OperationLog{}).Select("module, count(*) as count").
		Group("module").Scan(&moduleCounts)

	// 按操作类型统计
	type ActionCount struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	var actionCounts []ActionCount
	db.Model(&models.OperationLog{}).Select("action, count(*) as count").
		Group("action").Order("count DESC").Limit(10).Scan(&actionCounts)

	utils.Success(c, gin.H{
		"today_count":   todayCount,
		"week_count":    weekCount,
		"module_counts": moduleCounts,
		"action_counts": actionCounts,
	})
}
