package controllers

import (
	"fmt"
	"project-flow/config"
	"project-flow/middleware"
	"project-flow/models"
	"project-flow/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ProjectController struct{}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name            string  `json:"name" binding:"required"`
	ProjectNo       string  `json:"project_no"`
	ProjectType     string  `json:"project_type" binding:"required"`     // 成本性/资本性
	ManagerID       uint    `json:"manager_id" binding:"required"`       // 项目负责人
	ContractNo      string  `json:"contract_no"`                         // 合同编号（非必填）
	BudgetCode      string  `json:"budget_code"`                         // 预算编码（非必填）
	InnovationCode  string  `json:"innovation_code"`                     // 创新项目编码（非必填）
	InitiationDate  string  `json:"initiation_date" binding:"required"`  // 立项日期
	ClosingDate     string  `json:"closing_date" binding:"required"`     // 结项日期
	LaborCost       float64 `json:"labor_cost" binding:"required"`       // 人工费用
	DirectCost      float64 `json:"direct_cost" binding:"required"`      // 直接投入费用
	OutsourcingCost float64 `json:"outsourcing_cost" binding:"required"` // 委托研发费用
	OtherCost       float64 `json:"other_cost" binding:"required"`       // 其他费用
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name            string  `json:"name"`
	ProjectType     string  `json:"project_type"`
	ManagerID       uint    `json:"manager_id"`
	ContractNo      string  `json:"contract_no"`
	BudgetCode      string  `json:"budget_code"`
	InnovationCode  string  `json:"innovation_code"`
	InitiationDate  string  `json:"initiation_date"`
	ClosingDate     string  `json:"closing_date"`
	LaborCost       float64 `json:"labor_cost"`
	DirectCost      float64 `json:"direct_cost"`
	OutsourcingCost float64 `json:"outsourcing_cost"`
	OtherCost       float64 `json:"other_cost"`
	CurrentPhase    string  `json:"current_phase"`
	Status          string  `json:"status"`
}

// List 获取项目列表（所有用户可查看所有项目）
func (pc *ProjectController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	status := c.Query("status")
	phase := c.Query("phase")
	managerID := c.Query("manager_id")

	db := config.GetDB()

	var projects []models.Project
	var total int64

	query := db.Model(&models.Project{}).Preload("Manager").Preload("SubManager").Preload("Creator")

	// 所有用户可查看所有项目，不再根据角色过滤

	if keyword != "" {
		query = query.Where("name LIKE ? OR project_no LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if phase != "" {
		query = query.Where("current_phase = ?", phase)
	}
	if managerID != "" {
		query = query.Where("manager_id = ?", managerID)
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&projects)

	utils.SuccessPage(c, projects, total, page, pageSize)
}

// Create 创建项目
func (pc *ProjectController) Create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写完整信息")
		return
	}

	// 验证项目类型
	if req.ProjectType != "成本性" && req.ProjectType != "资本性" {
		utils.BadRequest(c, "项目类型必须是成本性或资本性")
		return
	}

	userID, _ := c.Get("userID")
	db := config.GetDB()

	// 生成项目编号
	projectNo := req.ProjectNo
	if projectNo == "" {
		projectNo = fmt.Sprintf("PRJ%s%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
	}

	// 检查项目编号是否存在
	var existing models.Project
	if db.Where("project_no = ?", projectNo).First(&existing).RowsAffected > 0 {
		utils.Error(c, 400, "项目编号已存在")
		return
	}

	project := models.Project{
		ProjectNo:       projectNo,
		Name:            req.Name,
		ProjectType:     req.ProjectType,
		ManagerID:       req.ManagerID,
		ContractNo:      req.ContractNo,
		BudgetCode:      req.BudgetCode,
		InnovationCode:  req.InnovationCode,
		LaborCost:       req.LaborCost,
		DirectCost:      req.DirectCost,
		OutsourcingCost: req.OutsourcingCost,
		OtherCost:       req.OtherCost,
		CurrentPhase:    config.PhaseInitiation,
		Status:          config.StatusInProgress,
		CreatedBy:       userID.(uint),
	}

	// 解析日期
	if req.InitiationDate != "" {
		t, _ := time.Parse("2006-01-02", req.InitiationDate)
		project.InitiationDate = &t
	}
	if req.ClosingDate != "" {
		t, _ := time.Parse("2006-01-02", req.ClosingDate)
		project.ClosingDate = &t
	}

	if err := db.Create(&project).Error; err != nil {
		utils.ServerError(c, "创建项目失败")
		return
	}

	// 创建项目阶段（5个固定阶段：立项、招标、合同签订 + 验收、结项，中间由用户自定义）
	phases := []models.ProjectPhase{
		{ProjectID: project.ID, PhaseName: config.PhaseInitiation, PhaseOrder: 1, IsFixed: true, Status: config.StatusInProgress},
		{ProjectID: project.ID, PhaseName: config.PhaseBidding, PhaseOrder: 2, IsFixed: true, Status: config.StatusNotStarted},
		{ProjectID: project.ID, PhaseName: config.PhaseContract, PhaseOrder: 3, IsFixed: true, Status: config.StatusNotStarted},
		// 中间阶段由项目总工自定义，顺序从4开始
		{ProjectID: project.ID, PhaseName: config.PhaseAcceptance, PhaseOrder: 100, IsFixed: true, Status: config.StatusNotStarted},
		{ProjectID: project.ID, PhaseName: config.PhaseClosing, PhaseOrder: 101, IsFixed: true, Status: config.StatusNotStarted},
	}
	db.Create(&phases)

	// 添加项目负责人为成员
	if req.ManagerID != 0 {
		member := models.ProjectMember{
			ProjectID: project.ID,
			UserID:    req.ManagerID,
			RoleType:  "manager", // 项目负责人
			JoinDate:  time.Now(),
		}
		db.Create(&member)
	}

	// 记录日志
	middleware.LogOperation(c, "create", "project", "project", project.ID, project.Name, "创建项目: "+project.Name, "success")

	utils.SuccessWithMessage(c, "创建成功", project)
}

// Get 获取项目详情
func (pc *ProjectController) Get(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var project models.Project
	if err := db.Preload("Manager").Preload("SubManager").Preload("Creator").Preload("Phases").First(&project, id).Error; err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	utils.Success(c, project)
}

// Update 更新项目（只有创建者或管理员可修改）
func (pc *ProjectController) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	db := config.GetDB()
	var project models.Project
	if err := db.First(&project, id).Error; err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	// 检查权限：只有项目负责人或管理员可修改
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if project.ManagerID != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只有项目负责人才能修改项目信息")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ProjectType != "" {
		updates["project_type"] = req.ProjectType
	}
	if req.ManagerID != 0 {
		updates["manager_id"] = req.ManagerID
	}
	if req.ContractNo != "" {
		updates["contract_no"] = req.ContractNo
	}
	if req.BudgetCode != "" {
		updates["budget_code"] = req.BudgetCode
	}
	if req.InnovationCode != "" {
		updates["innovation_code"] = req.InnovationCode
	}
	if req.LaborCost != 0 {
		updates["labor_cost"] = req.LaborCost
	}
	if req.DirectCost != 0 {
		updates["direct_cost"] = req.DirectCost
	}
	if req.OutsourcingCost != 0 {
		updates["outsourcing_cost"] = req.OutsourcingCost
	}
	if req.OtherCost != 0 {
		updates["other_cost"] = req.OtherCost
	}
	if req.CurrentPhase != "" {
		updates["current_phase"] = req.CurrentPhase
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.InitiationDate != "" {
		t, _ := time.Parse("2006-01-02", req.InitiationDate)
		updates["initiation_date"] = t
	}
	if req.ClosingDate != "" {
		t, _ := time.Parse("2006-01-02", req.ClosingDate)
		updates["closing_date"] = t
	}

	if err := db.Model(&project).Updates(updates).Error; err != nil {
		utils.ServerError(c, "更新失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "update", "project", "project", project.ID, project.Name, "更新项目: "+project.Name, "success")

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// Delete 删除项目
func (pc *ProjectController) Delete(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var project models.Project
	if err := db.First(&project, id).Error; err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	// 检查权限：只有项目负责人或管理员可删除
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if project.ManagerID != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只有项目负责人才能删除项目")
		return
	}

	// 软删除项目
	if err := db.Delete(&project).Error; err != nil {
		utils.ServerError(c, "删除失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "delete", "project", "project", project.ID, project.Name, "删除项目: "+project.Name, "success")

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// GetPhases 获取项目阶段列表
func (pc *ProjectController) GetPhases(c *gin.Context) {
	projectID := c.Param("id")

	db := config.GetDB()
	var phases []models.ProjectPhase
	db.Where("project_id = ?", projectID).Order("phase_order").Find(&phases)

	utils.Success(c, phases)
}

// UpdatePhaseRequest 更新阶段请求
type UpdatePhaseRequest struct {
	Status string `json:"status"`
	Remark string `json:"remark"`
}

// UpdatePhase 更新项目阶段状态（只有创建者或管理员可操作）
func (pc *ProjectController) UpdatePhase(c *gin.Context) {
	projectID := c.Param("id")
	phaseID := c.Param("phaseId")

	var req UpdatePhaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	db := config.GetDB()

	// 检查项目并验证权限
	var project models.Project
	if err := db.First(&project, projectID).Error; err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if project.ManagerID != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只有项目负责人才能修改项目阶段")
		return
	}

	var phase models.ProjectPhase
	if err := db.Where("id = ? AND project_id = ?", phaseID, projectID).First(&phase).Error; err != nil {
		utils.NotFound(c, "阶段不存在")
		return
	}

	updates := make(map[string]interface{})
	if req.Status != "" {
		updates["status"] = req.Status
		if req.Status == config.StatusCompleted {
			now := time.Now()
			updates["completed_at"] = now
		}
		if req.Status == config.StatusInProgress && phase.StartDate == nil {
			now := time.Now()
			updates["start_date"] = now
		}
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	db.Model(&phase).Updates(updates)

	// 如果当前阶段完成，自动开启下一阶段
	if req.Status == config.StatusCompleted {
		var nextPhase models.ProjectPhase
		if db.Where("project_id = ? AND phase_order = ?", projectID, phase.PhaseOrder+1).First(&nextPhase).RowsAffected > 0 {
			now := time.Now()
			db.Model(&nextPhase).Updates(map[string]interface{}{
				"status":     config.StatusInProgress,
				"start_date": now,
			})
			// 更新项目当前阶段
			db.Model(&models.Project{}).Where("id = ?", projectID).Update("current_phase", nextPhase.PhaseName)
		} else {
			// 所有阶段完成，项目结项
			db.Model(&models.Project{}).Where("id = ?", projectID).Update("status", config.StatusCompleted)
		}
	}

	// 记录日志
	middleware.LogOperation(c, "update_phase", "project", "phase", phase.ID, phase.PhaseName, "更新阶段状态: "+phase.PhaseName, "success")

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	RoleType string `json:"role_type" binding:"required"`
}

// AddMember 添加项目成员（只有创建者或管理员可操作）
func (pc *ProjectController) AddMember(c *gin.Context) {
	projectID := c.Param("id")

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请选择成员")
		return
	}

	db := config.GetDB()

	// 检查项目是否存在
	var project models.Project
	if err := db.First(&project, projectID).Error; err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	// 检查权限：只有项目负责人或管理员可添加成员
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if project.ManagerID != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只有项目负责人才能添加项目成员")
		return
	}

	pid, _ := strconv.ParseUint(projectID, 10, 32)

	// 检查是否已是成员
	var existing models.ProjectMember
	if db.Where("project_id = ? AND user_id = ?", projectID, req.UserID).First(&existing).RowsAffected > 0 {
		utils.Error(c, 400, "该用户已是项目成员")
		return
	}

	member := models.ProjectMember{
		ProjectID: uint(pid),
		UserID:    req.UserID,
		RoleType:  req.RoleType,
		JoinDate:  time.Now(),
	}

	if err := db.Create(&member).Error; err != nil {
		utils.ServerError(c, "添加失败")
		return
	}

	// 记录日志
	var user models.User
	db.First(&user, req.UserID)
	middleware.LogOperation(c, "add_member", "project", "project_member", member.ID, user.Name, "添加项目成员: "+user.Name, "success")

	utils.SuccessWithMessage(c, "添加成功", nil)
}

// GetMembers 获取项目成员列表
func (pc *ProjectController) GetMembers(c *gin.Context) {
	projectID := c.Param("id")

	db := config.GetDB()
	var members []models.ProjectMember
	db.Where("project_id = ?", projectID).Preload("User").Find(&members)

	utils.Success(c, members)
}

// RemoveMember 移除项目成员
func (pc *ProjectController) RemoveMember(c *gin.Context) {
	projectID := c.Param("id")
	memberID := c.Param("memberId")

	db := config.GetDB()

	// 检查项目并验证权限
	var project models.Project
	if err := db.First(&project, projectID).Error; err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	// 检查权限：只有项目负责人或管理员可移除成员
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if project.ManagerID != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只有项目负责人才能移除项目成员")
		return
	}

	var member models.ProjectMember
	if err := db.Where("id = ? AND project_id = ?", memberID, projectID).First(&member).Error; err != nil {
		utils.NotFound(c, "成员不存在")
		return
	}

	// 获取用户信息用于日志
	var user models.User
	db.First(&user, member.UserID)

	if err := db.Delete(&member).Error; err != nil {
		utils.ServerError(c, "移除失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "remove_member", "project", "project_member", member.ID, user.Name, "移除项目成员: "+user.Name, "success")

	utils.SuccessWithMessage(c, "移除成功", nil)
}

// GetStatistics 获取项目统计信息
func (pc *ProjectController) GetStatistics(c *gin.Context) {
	db := config.GetDB()

	var totalProjects int64
	var inProgressProjects int64
	var completedProjects int64
	var notStartedProjects int64

	db.Model(&models.Project{}).Count(&totalProjects)
	db.Model(&models.Project{}).Where("status = ?", config.StatusInProgress).Count(&inProgressProjects)
	db.Model(&models.Project{}).Where("status = ?", config.StatusCompleted).Count(&completedProjects)
	db.Model(&models.Project{}).Where("status = ?", config.StatusNotStarted).Count(&notStartedProjects)

	// 各阶段项目数量
	type PhaseCount struct {
		Phase string `json:"phase"`
		Count int64  `json:"count"`
	}
	var phaseCounts []PhaseCount
	db.Model(&models.Project{}).Select("current_phase as phase, count(*) as count").
		Group("current_phase").Scan(&phaseCounts)

	utils.Success(c, gin.H{
		"total":        totalProjects,
		"in_progress":  inProgressProjects,
		"completed":    completedProjects,
		"not_started":  notStartedProjects,
		"phase_counts": phaseCounts,
	})
}

// AddPhaseRequest 添加阶段请求
type AddPhaseRequest struct {
	PhaseName string `json:"phase_name" binding:"required"`
}

// AddPhase 添加自定义阶段（只有创建者或管理员可操作）
func (pc *ProjectController) AddPhase(c *gin.Context) {
	projectID := c.Param("id")

	var req AddPhaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请输入阶段名称")
		return
	}

	db := config.GetDB()

	// 检查项目并验证权限
	var project models.Project
	if err := db.First(&project, projectID).Error; err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if project.ManagerID != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只有项目负责人才能添加项目阶段")
		return
	}

	// 获取当前最大的自定义阶段顺序（圈定范围：大于等于4且小于100）
	var maxOrder int
	db.Model(&models.ProjectPhase{}).Where("project_id = ? AND phase_order >= 4 AND phase_order < 100", projectID).
		Select("COALESCE(MAX(phase_order), 3)").Scan(&maxOrder)

	newPhase := models.ProjectPhase{
		ProjectID:  project.ID,
		PhaseName:  req.PhaseName,
		PhaseOrder: maxOrder + 1,
		IsFixed:    false,
		Status:     config.StatusNotStarted,
	}

	if err := db.Create(&newPhase).Error; err != nil {
		utils.ServerError(c, "添加阶段失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "create", "project", "phase", newPhase.ID, newPhase.PhaseName, "添加项目阶段: "+newPhase.PhaseName, "success")

	utils.SuccessWithMessage(c, "添加成功", newPhase)
}

// DeletePhase 删除自定义阶段（固定阶段不可删除）
func (pc *ProjectController) DeletePhase(c *gin.Context) {
	projectID := c.Param("id")
	phaseID := c.Param("phaseId")

	db := config.GetDB()

	// 检查项目并验证权限
	var project models.Project
	if err := db.First(&project, projectID).Error; err != nil {
		utils.NotFound(c, "项目不存在")
		return
	}

	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if project.ManagerID != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只有项目负责人才能删除项目阶段")
		return
	}

	var phase models.ProjectPhase
	if err := db.Where("id = ? AND project_id = ?", phaseID, projectID).First(&phase).Error; err != nil {
		utils.NotFound(c, "阶段不存在")
		return
	}

	// 固定阶段不可删除
	if phase.IsFixed {
		utils.Error(c, 400, "固定阶段不可删除")
		return
	}

	// 检查阶段下是否有任务
	var taskCount int64
	db.Model(&models.Task{}).Where("phase_id = ?", phaseID).Count(&taskCount)
	if taskCount > 0 {
		utils.Error(c, 400, "该阶段下存在任务，无法删除")
		return
	}

	phaseName := phase.PhaseName
	if err := db.Delete(&phase).Error; err != nil {
		utils.ServerError(c, "删除失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "delete", "project", "phase", phase.ID, phaseName, "删除项目阶段: "+phaseName, "success")

	utils.SuccessWithMessage(c, "删除成功", nil)
}
