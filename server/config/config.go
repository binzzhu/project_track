package config

import (
	"os"
	"time"
)

// 系统配置
var (
	// JWT配置
	JWTSecret     = []byte(getEnv("JWT_SECRET", "project-flow-secret-key-2024"))
	JWTExpireTime = time.Hour * 24 // Token有效期24小时

	// 服务器配置
	ServerPort = getEnv("SERVER_PORT", ":8080")

	// 文件上传配置
	UploadPath  = getEnv("UPLOAD_PATH", "./uploads")
	MaxFileSize = int64(100 * 1024 * 1024) // 100MB

	// 数据库配置
	DBPath = getEnv("DB_PATH", "./data/project_flow.db")

	// 密码复杂度要求
	MinPasswordLength = 8
)

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 项目阶段常量（固定阶段）
const (
	PhaseInitiation = "initiation" // 立项（固定）
	PhaseBidding    = "bidding"    // 招标（固定）
	PhaseContract   = "contract"   // 合同签订（固定）
	PhaseAcceptance = "acceptance" // 验收（固定）
	PhaseClosing    = "closing"    // 结项（固定）
)

// 阶段状态
const (
	StatusNotStarted = "not_started" // 未开始
	StatusInProgress = "in_progress" // 进行中
	StatusCompleted  = "completed"   // 已完成
	StatusRejected   = "rejected"    // 被驳回
)

// 任务状态
const (
	TaskNotStarted = "not_started" // 未开始
	TaskInProgress = "in_progress" // 进行中
	TaskCompleted  = "completed"   // 已完成
	TaskRejected   = "rejected"    // 被驳回
)

// 角色类型（组织级）
const (
	RoleAdmin       = "admin"        // 系统管理员
	RoleDeptManager = "dept_manager" // 部门经理
	RoleTeamLeader  = "team_leader"  // 组长
	RoleTeamMember  = "team_member"  // 组员
)

// 部门类型
const (
	DeptBMS      = "BMS研发部"  // BMS研发部
	DeptPACK     = "PACK研发部" // PACK研发部
	DeptGeneral  = "综合部"     // 综合部
	DeptExternal = "外部单位"    // 外部单位
)

// 允许的部门列表
var AllowedDepartments = []string{
	DeptBMS,
	DeptPACK,
	DeptGeneral,
	DeptExternal,
}
