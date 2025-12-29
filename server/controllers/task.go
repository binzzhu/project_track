package controllers

import (
	"project-flow/config"
	"project-flow/middleware"
	"project-flow/models"
	"project-flow/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetTaskStatistics 获取任务统计
func (tc *TaskController) GetTaskStatistics(c *gin.Context) {
	db := config.GetDB()

	var total int64
	var notStarted int64
	var inProgress int64
	var completed int64
	var rejected int64

	// 所有人都查看全部任务的统计情况，不做权限过滤
	getBaseQuery := func() *gorm.DB {
		return db.Model(&models.Task{})
	}

	// 总数
	getBaseQuery().Count(&total)

	// 分别统计各状态（每次都创建新的查询对象）
	getBaseQuery().Where("status = ?", config.TaskNotStarted).Count(&notStarted)
	getBaseQuery().Where("status = ?", config.TaskInProgress).Count(&inProgress)
	getBaseQuery().Where("status = ?", config.TaskCompleted).Count(&completed)
	getBaseQuery().Where("status = ?", config.TaskRejected).Count(&rejected)

	stats := map[string]int64{
		"total":       total,
		"not_started": notStarted,
		"in_progress": inProgress,
		"completed":   completed,
		"rejected":    rejected,
	}

	utils.Success(c, stats)
}

type TaskController struct{}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	ProjectID    uint   `json:"project_id" binding:"required"`
	PhaseID      uint   `json:"phase_id"`
	TaskName     string `json:"task_name" binding:"required"`
	Description  string `json:"description"`
	TaskType     string `json:"task_type"`
	AssigneeID   uint   `json:"assignee_id"`
	AssigneeType string `json:"assignee_type"`
	Deadline     string `json:"deadline"`
	Priority     int    `json:"priority"`
	Deliverables string `json:"deliverables"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	TaskName     string `json:"task_name"`
	Description  string `json:"description"`
	TaskType     string `json:"task_type"`
	AssigneeID   uint   `json:"assignee_id"`
	AssigneeType string `json:"assignee_type"`
	Deadline     string `json:"deadline"`
	Priority     int    `json:"priority"`
	Deliverables string `json:"deliverables"`
	Status       string `json:"status"`
}

// List 获取任务列表
func (tc *TaskController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	projectID := c.Query("project_id")
	phaseID := c.Query("phase_id")
	status := c.Query("status")
	assigneeID := c.Query("assignee_id")
	keyword := c.Query("keyword")

	db := config.GetDB()

	var tasks []models.Task
	var total int64

	query := db.Model(&models.Task{}).Preload("Project").Preload("Phase").Preload("Assignee")

	// 所有角色都可以查看所有任务

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if phaseID != "" {
		query = query.Where("phase_id = ?", phaseID)
	}
	if status != "" {
		query = query.Where("tasks.status = ?", status)
	}
	if assigneeID != "" {
		query = query.Where("assignee_id = ?", assigneeID)
	}
	if keyword != "" {
		query = query.Where("task_name LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("tasks.id DESC").Find(&tasks)

	utils.SuccessPage(c, tasks, total, page, pageSize)
}

// Create 创建任务
func (tc *TaskController) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写任务名称")
		return
	}

	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	db := config.GetDB()

	// 检查项目是否存在
	var project models.Project
	if err := db.First(&project, req.ProjectID).Error; err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	// 检查权限：只有项目负责人或管理员可创建任务
	if project.ManagerID != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只有项目负责人才能创建任务")
		return
	}

	task := models.Task{
		ProjectID:    req.ProjectID,
		PhaseID:      req.PhaseID,
		TaskName:     req.TaskName,
		Description:  req.Description,
		TaskType:     req.TaskType,
		AssigneeID:   req.AssigneeID,
		AssigneeType: req.AssigneeType,
		Priority:     req.Priority,
		Deliverables: req.Deliverables,
		Status:       config.TaskNotStarted,
		CreatedBy:    userID.(uint),
	}

	if req.Deadline != "" {
		t, _ := time.Parse("2006-01-02", req.Deadline)
		task.Deadline = &t
	}

	if err := db.Create(&task).Error; err != nil {
		utils.ServerError(c, "创建任务失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "create", "task", "task", task.ID, task.TaskName, "创建任务: "+task.TaskName, "success")

	utils.SuccessWithMessage(c, "创建成功", task)
}

// BatchCreateRequest 批量创建任务请求
type BatchCreateRequest struct {
	Tasks []CreateTaskRequest `json:"tasks" binding:"required"`
}

// BatchCreate 批量创建任务
func (tc *TaskController) BatchCreate(c *gin.Context) {
	var req BatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	db := config.GetDB()

	// 检查第一个任务的项目权限（假设批量创建都在同一个项目下）
	if len(req.Tasks) > 0 {
		var project models.Project
		if err := db.First(&project, req.Tasks[0].ProjectID).Error; err != nil {
			utils.NotFound(c, "项目不存在")
			return
		}

		// 检查权限：只有项目负责人或管理员可创建任务
		if project.ManagerID != userID.(uint) && roleCode != config.RoleAdmin {
			utils.Forbidden(c, "只有项目负责人才能创建任务")
			return
		}
	}

	var createdTasks []models.Task
	for _, t := range req.Tasks {
		task := models.Task{
			ProjectID:    t.ProjectID,
			PhaseID:      t.PhaseID,
			TaskName:     t.TaskName,
			Description:  t.Description,
			TaskType:     t.TaskType,
			AssigneeID:   t.AssigneeID,
			AssigneeType: t.AssigneeType,
			Priority:     t.Priority,
			Deliverables: t.Deliverables,
			Status:       config.TaskNotStarted,
			CreatedBy:    userID.(uint),
		}
		if t.Deadline != "" {
			deadline, _ := time.Parse("2006-01-02", t.Deadline)
			task.Deadline = &deadline
		}
		db.Create(&task)
		createdTasks = append(createdTasks, task)
	}

	utils.SuccessWithMessage(c, "批量创建成功", createdTasks)
}

// Get 获取任务详情
func (tc *TaskController) Get(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var task models.Task
	if err := db.Preload("Project").Preload("Phase").Preload("Assignee").Preload("Documents").First(&task, id).Error; err != nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	utils.Success(c, task)
}

// Update 更新任务
func (tc *TaskController) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	db := config.GetDB()
	var task models.Task
	if err := db.First(&task, id).Error; err != nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	updates := make(map[string]interface{})
	if req.TaskName != "" {
		updates["task_name"] = req.TaskName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.TaskType != "" {
		updates["task_type"] = req.TaskType
	}
	if req.AssigneeID != 0 {
		updates["assignee_id"] = req.AssigneeID
	}
	if req.AssigneeType != "" {
		updates["assignee_type"] = req.AssigneeType
	}
	if req.Priority != 0 {
		updates["priority"] = req.Priority
	}
	if req.Deliverables != "" {
		updates["deliverables"] = req.Deliverables
	}
	if req.Status != "" {
		updates["status"] = req.Status
		if req.Status == config.TaskCompleted {
			now := time.Now()
			updates["completed_at"] = now
		}
	}
	if req.Deadline != "" {
		t, _ := time.Parse("2006-01-02", req.Deadline)
		updates["deadline"] = t
	}

	if err := db.Model(&task).Updates(updates).Error; err != nil {
		utils.ServerError(c, "更新失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "update", "task", "task", task.ID, task.TaskName, "更新任务: "+task.TaskName, "success")

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// Delete 删除任务（只有项目负责人或管理员可删除）
func (tc *TaskController) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")

	db := config.GetDB()
	var task models.Task
	if err := db.First(&task, id).Error; err != nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	// 检查权限：管理员或项目负责人可删除
	if roleCode != config.RoleAdmin {
		var project models.Project
		if err := db.First(&project, task.ProjectID).Error; err != nil {
			utils.NotFound(c, "项目不存在")
			return
		}
		if project.ManagerID != userID.(uint) {
			utils.Forbidden(c, "只有项目负责人才能删除任务")
			return
		}
	}

	if err := db.Delete(&task).Error; err != nil {
		utils.ServerError(c, "删除失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "delete", "task", "task", task.ID, task.TaskName, "删除任务: "+task.TaskName, "success")

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// UpdateStatusRequest 更新任务状态请求
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateStatus 更新任务状态
func (tc *TaskController) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请选择状态")
		return
	}

	userID, _ := c.Get("userID")
	db := config.GetDB()
	var task models.Task
	if err := db.First(&task, id).Error; err != nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	// 权限检查：只有任务负责人有权限更改状态
	if task.AssigneeID != userID.(uint) {
		utils.Forbidden(c, "只有任务负责人才能更改任务状态")
		return
	}

	updates := map[string]interface{}{"status": req.Status}
	if req.Status == config.TaskCompleted {
		now := time.Now()
		updates["completed_at"] = now
	}

	db.Model(&task).Updates(updates)

	// 记录日志
	middleware.LogOperation(c, "update_status", "task", "task", task.ID, task.TaskName, "更新任务状态: "+req.Status, "success")

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// ReviewRequest 审核请求
type ReviewRequest struct {
	Status  string `json:"status" binding:"required"` // approved/rejected
	Comment string `json:"comment"`
}

// ReviewTask 审核任务交付件
func (tc *TaskController) ReviewTask(c *gin.Context) {
	id := c.Param("id")

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请选择审核结果")
		return
	}

	userID, _ := c.Get("userID")
	db := config.GetDB()

	var task models.Task
	if err := db.First(&task, id).Error; err != nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"review_status":  req.Status,
		"review_comment": req.Comment,
		"reviewed_by":    userID,
		"reviewed_at":    now,
	}

	if req.Status == "approved" {
		updates["status"] = config.TaskCompleted
		updates["completed_at"] = now
	} else if req.Status == "rejected" {
		updates["status"] = config.TaskRejected
	}

	db.Model(&task).Updates(updates)

	// 记录日志
	action := "审核通过"
	if req.Status == "rejected" {
		action = "审核驳回"
	}
	middleware.LogOperation(c, "review", "task", "task", task.ID, task.TaskName, action+": "+task.TaskName, "success")

	utils.SuccessWithMessage(c, "审核完成", nil)
}

// GetMyTasks 获取我的任务
func (tc *TaskController) GetMyTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	userID, _ := c.Get("userID")
	db := config.GetDB()

	var tasks []models.Task
	var total int64

	query := db.Model(&models.Task{}).Where("assignee_id = ?", userID).
		Preload("Project").Preload("Phase")

	if status != "" {
		// 支持多状态查询（逗号分隔）
		statuses := strings.Split(status, ",")
		query = query.Where("status IN ?", statuses)
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("deadline ASC, priority ASC").Find(&tasks)

	utils.SuccessPage(c, tasks, total, page, pageSize)
}
