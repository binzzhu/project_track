package controllers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"project-flow/config"
	"project-flow/middleware"
	"project-flow/models"
	"project-flow/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type DocumentController struct{}

// List 获取文档列表
func (dc *DocumentController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	projectID := c.Query("project_id")
	phaseID := c.Query("phase_id")
	taskID := c.Query("task_id")
	status := c.Query("status")

	db := config.GetDB()
	var docs []models.Document
	var total int64

	query := db.Model(&models.Document{}).Preload("Uploader").Preload("Phase")

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if phaseID != "" {
		query = query.Where("phase_id = ?", phaseID)
	}
	if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&docs)

	utils.SuccessPage(c, docs, total, page, pageSize)
}

// Upload 上传文档（只有项目创建者或子负责人可上传）
func (dc *DocumentController) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > config.MaxFileSize {
		utils.BadRequest(c, "文件大小超过限制(100MB)")
		return
	}

	projectIDStr := c.PostForm("project_id")
	phaseIDStr := c.PostForm("phase_id")
	taskIDStr := c.PostForm("task_id")

	if projectIDStr == "" {
		utils.BadRequest(c, "项目参数错误")
		return
	}

	projectIDUint64, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "项目参数错误")
		return
	}

	var phaseIDUint64 uint64
	if phaseIDStr != "" {
		if phaseIDUint64, err = strconv.ParseUint(phaseIDStr, 10, 32); err != nil {
			utils.BadRequest(c, "阶段参数错误")
			return
		}
	}

	var taskIDUint64 uint64
	if taskIDStr != "" {
		if taskIDUint64, err = strconv.ParseUint(taskIDStr, 10, 32); err != nil {
			utils.BadRequest(c, "任务参数错误")
			return
		}
	}

	docName := c.PostForm("doc_name")
	docType := c.PostForm("doc_type")
	remark := c.PostForm("remark")

	if docName == "" {
		docName = header.Filename
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		utils.ServerError(c, "用户信息异常")
		return
	}

	roleCodeValue, _ := c.Get("roleCode")
	roleCode, _ := roleCodeValue.(string)

	db := config.GetDB()

	// 检查上传权限：管理员始终可以上传
	if roleCode != config.RoleAdmin && projectIDUint64 > 0 {
		var project models.Project
		if err := db.First(&project, projectIDUint64).Error; err != nil {
			utils.NotFound(c, "项目不存在")
			return
		}

		// 检查是否是项目经理
		isProjectManager := project.ManagerID == userID

		// 检查是否是创建者
		isCreator := project.CreatedBy == userID

		// 检查是否是子负责人（在project_members表中的用户即为子负责人）
		isSubManager := false
		var member models.ProjectMember
		if db.Where("project_id = ? AND user_id = ?", uint(projectIDUint64), userID).First(&member).RowsAffected > 0 {
			isSubManager = true
		}

		// 如果是任务交付件，检查是否是任务负责人
		isTaskAssignee := false
		if taskIDUint64 > 0 {
			var task models.Task
			if db.First(&task, taskIDUint64).Error == nil {
				// 任务未完成时，任务负责人可以上传
				if task.Status != config.TaskCompleted && task.AssigneeID == userID {
					isTaskAssignee = true
				}
				// 任务已完成时，只有项目经理可以上传
				if task.Status == config.TaskCompleted && !isProjectManager {
					utils.Forbidden(c, "任务完成后，只有项目经理可以上传交付件")
					return
				}
			}
		}

		if !isCreator && !isSubManager && !isTaskAssignee && !isProjectManager {
			utils.Forbidden(c, "只有项目创建者、项目经理、子负责人或任务负责人才能上传资料")
			return
		}
	}

	// 创建上传目录
	uploadDir := filepath.Join(config.UploadPath, "documents", time.Now().Format("200601"))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ServerError(c, "创建目录失败")
		return
	}

	// 生成文件名
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), userID, ext)
	filePath := filepath.Join(uploadDir, filename)

	// 保存文件
	out, err := os.Create(filePath)
	if err != nil {
		utils.ServerError(c, "保存文件失败")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		utils.ServerError(c, "保存文件失败")
		return
	}

	doc := models.Document{
		ProjectID:  uint(projectIDUint64),
		PhaseID:    uint(phaseIDUint64),
		TaskID:     uint(taskIDUint64),
		DocName:    docName,
		DocType:    docType,
		FilePath:   filePath,
		FileSize:   header.Size,
		MimeType:   header.Header.Get("Content-Type"),
		Version:    "1.0",
		Status:     "pending",
		UploadedBy: userID,
		Remark:     remark,
	}

	if err := db.Create(&doc).Error; err != nil {
		utils.ServerError(c, "保存文档信息失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "upload", "document", "document", doc.ID, doc.DocName, "上传文档: "+doc.DocName, "success")

	utils.SuccessWithMessage(c, "上传成功", doc)
}

// Get 获取文档详情
func (dc *DocumentController) Get(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var doc models.Document
	if err := db.Preload("Uploader").First(&doc, id).Error; err != nil {
		utils.NotFound(c, "文档不存在")
		return
	}

	utils.Success(c, doc)
}

// Download 下载文档
func (dc *DocumentController) Download(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var doc models.Document
	if err := db.First(&doc, id).Error; err != nil {
		utils.NotFound(c, "文档不存在")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		utils.NotFound(c, "文件不存在")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", doc.DocName))
	c.Header("Content-Type", doc.MimeType)
	c.File(doc.FilePath)
}

// UpdateDocRequest 更新文档请求
type UpdateDocRequest struct {
	DocName string `json:"doc_name"`
	DocType string `json:"doc_type"`
	Status  string `json:"status"`
	Remark  string `json:"remark"`
}

// Update 更新文档信息
func (dc *DocumentController) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	db := config.GetDB()
	var doc models.Document
	if err := db.First(&doc, id).Error; err != nil {
		utils.NotFound(c, "文档不存在")
		return
	}

	updates := make(map[string]interface{})
	if req.DocName != "" {
		updates["doc_name"] = req.DocName
	}
	if req.DocType != "" {
		updates["doc_type"] = req.DocType
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	db.Model(&doc).Updates(updates)

	// 记录日志
	middleware.LogOperation(c, "update", "document", "document", doc.ID, doc.DocName, "更新文档: "+doc.DocName, "success")

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// Delete 删除文档
func (dc *DocumentController) Delete(c *gin.Context) {
	id := c.Param("id")

	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	db := config.GetDB()
	var doc models.Document
	if err := db.Preload("Task").First(&doc, id).Error; err != nil {
		utils.NotFound(c, "文档不存在")
		return
	}

	// 权限检查：
	// 1. 如果是任务交付件，且任务已完成，只有项目经理有权限删除
	// 2. 如果是任务交付件，且任务未完成，只有任务负责人有权限删除
	// 3. 其他情况下，管理员或上传者可以删除
	if doc.TaskID > 0 {
		// 获取任务信息
		var task models.Task
		if err := db.Preload("Project").First(&task, doc.TaskID).Error; err == nil {
			// 如果任务已完成，只有项目经理有权限删除
			if task.Status == config.TaskCompleted {
				if roleCode != config.RoleAdmin && task.Project.ManagerID != userID.(uint) {
					utils.Forbidden(c, "任务完成后，只有项目经理有权限删除交付件")
					return
				}
			} else {
				// 任务未完成，只有任务负责人可以删除
				if roleCode != config.RoleAdmin && task.AssigneeID != userID.(uint) {
					utils.Forbidden(c, "只有任务负责人才能删除交付件")
					return
				}
			}
		}
	} else {
		// 非任务交付件，管理员或上传者可以删除
		if roleCode != config.RoleAdmin && doc.UploadedBy != userID.(uint) {
			utils.Forbidden(c, "只有管理员或上传者才能删除文档")
			return
		}
	}

	// 删除文件
	if doc.FilePath != "" {
		os.Remove(doc.FilePath)
	}

	if err := db.Delete(&doc).Error; err != nil {
		utils.ServerError(c, "删除失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "delete", "document", "document", doc.ID, doc.DocName, "删除文档: "+doc.DocName, "success")

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// Archive 归档文档
func (dc *DocumentController) Archive(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var doc models.Document
	if err := db.First(&doc, id).Error; err != nil {
		utils.NotFound(c, "文档不存在")
		return
	}

	db.Model(&doc).Update("status", "archived")

	// 记录日志
	middleware.LogOperation(c, "archive", "document", "document", doc.ID, doc.DocName, "归档文档: "+doc.DocName, "success")

	utils.SuccessWithMessage(c, "归档成功", nil)
}
