package controllers

import (
	"fmt"
	"io"
	"log"
	"net/url"
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

type KnowledgeController struct{}

// List 获取知识库列表
func (kc *KnowledgeController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	categoryID := c.Query("category_id")
	status := c.Query("status")
	uploadedBy := c.Query("uploaded_by") // 新增：上传人搜索

	db := config.GetDB()
	var items []models.KnowledgeBase
	var total int64

	query := db.Model(&models.KnowledgeBase{}).Preload("Category").Preload("Uploader")

	if keyword != "" {
		query = query.Where("title LIKE ? OR keywords LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if uploadedBy != "" {
		query = query.Where("uploaded_by = ?", uploadedBy)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status = ?", "published")
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&items)

	utils.SuccessPage(c, items, total, page, pageSize)
}

// Upload 上传知识库资料
func (kc *KnowledgeController) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	if header.Size > config.MaxFileSize {
		utils.BadRequest(c, "文件大小超过限制(100MB)")
		return
	}

	title := c.PostForm("title")

	categoryStr := c.PostForm("category_id")
	if categoryStr == "" {
		utils.BadRequest(c, "请选择资料分类")
		return
	}
	categoryIDUint64, err := strconv.ParseUint(categoryStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "分类参数错误")
		return
	}
	categoryID := uint(categoryIDUint64)

	keywords := c.PostForm("keywords")
	description := c.PostForm("description")
	status := c.DefaultPostForm("status", "published")

	if title == "" {
		title = header.Filename
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "请先登录")
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		log.Printf("知识库上传用户信息类型异常: %#v\n", userIDValue)
		utils.ServerError(c, "用户信息异常")
		return
	}

	db := config.GetDB()

	var category models.KBCategory
	if err := db.First(&category, categoryID).Error; err != nil {
		log.Printf("知识库上传分类不存在，categoryID=%d, err=%v\n", categoryID, err)
		utils.BadRequest(c, "所选分类不存在")
		return
	}

	// 创建上传目录
	uploadDir := filepath.Join(config.UploadPath, "knowledge", time.Now().Format("200601"))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("知识库上传创建目录失败，path=%s, err=%v\n", uploadDir, err)
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
		log.Printf("知识库上传保存文件失败，file=%s, err=%v\n", filePath, err)
		utils.ServerError(c, "保存文件失败")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		log.Printf("知识库上传写入文件失败，file=%s, err=%v\n", filePath, err)
		utils.ServerError(c, "保存文件失败")
		return
	}

	kb := models.KnowledgeBase{
		Title:       title,
		CategoryID:  categoryID,
		Keywords:    keywords,
		Description: description,
		FilePath:    filePath,
		FileSize:    header.Size,
		MimeType:    header.Header.Get("Content-Type"),
		Version:     "1.0",
		Status:      status,
		UploadedBy:  userID,
	}

	if err := db.Create(&kb).Error; err != nil {
		log.Printf("知识库上传保存信息失败，kb=%+v, err=%v\n", kb, err)
		utils.ServerError(c, "保存信息失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "upload", "knowledge", "knowledge", kb.ID, kb.Title, "上传知识库资料: "+kb.Title, "success")

	utils.SuccessWithMessage(c, "上传成功", kb)
}

// Get 获取详情
func (kc *KnowledgeController) Get(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var kb models.KnowledgeBase
	if err := db.Preload("Category").Preload("Uploader").First(&kb, id).Error; err != nil {
		utils.NotFound(c, "资料不存在")
		return
	}

	// 增加查看次数
	db.Model(&kb).Update("view_count", kb.ViewCount+1)

	utils.Success(c, kb)
}

// Download 下载资料
func (kc *KnowledgeController) Download(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var kb models.KnowledgeBase
	if err := db.First(&kb, id).Error; err != nil {
		utils.NotFound(c, "资料不存在")
		return
	}

	// 检查文件是否存在，并记录详细错误
	if _, err := os.Stat(kb.FilePath); os.IsNotExist(err) {
		// 记录错误日志
		fmt.Printf("文件不存在: %s\n", kb.FilePath)
		utils.NotFound(c, fmt.Sprintf("文件不存在: %s", kb.FilePath))
		return
	}

	// 增加下载次数
	db.Model(&kb).Update("download_count", kb.DownloadCount+1)

	// 设置响应头，使用URL编码处理中文文件名
	filename := url.QueryEscape(kb.Title)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", kb.MimeType)
	c.File(kb.FilePath)
}

// UpdateKBRequest 更新请求
type UpdateKBRequest struct {
	Title       string `json:"title"`
	CategoryID  uint   `json:"category_id"`
	Keywords    string `json:"keywords"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// Update 更新资料信息
func (kc *KnowledgeController) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	db := config.GetDB()
	var kb models.KnowledgeBase
	if err := db.First(&kb, id).Error; err != nil {
		utils.NotFound(c, "资料不存在")
		return
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.CategoryID != 0 {
		updates["category_id"] = req.CategoryID
	}
	if req.Keywords != "" {
		updates["keywords"] = req.Keywords
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	db.Model(&kb).Updates(updates)

	// 记录日志
	middleware.LogOperation(c, "update", "knowledge", "knowledge", kb.ID, kb.Title, "更新知识库资料: "+kb.Title, "success")

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// Delete 删除资料
func (kc *KnowledgeController) Delete(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var kb models.KnowledgeBase
	if err := db.First(&kb, id).Error; err != nil {
		utils.NotFound(c, "资料不存在")
		return
	}

	// 权限检查：管理员和部门经理可以删除所有资料，其他用户只能删除自己上传的资料
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	isAdmin := roleCode == config.RoleAdmin || roleCode == config.RoleDeptManager
	isOwner := kb.UploadedBy == userID.(uint)

	if !isAdmin && !isOwner {
		utils.Forbidden(c, "只能删除自己上传的资料")
		return
	}

	// 删除文件
	if kb.FilePath != "" {
		os.Remove(kb.FilePath)
	}

	if err := db.Delete(&kb).Error; err != nil {
		utils.ServerError(c, "删除失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "delete", "knowledge", "knowledge", kb.ID, kb.Title, "删除知识库资料: "+kb.Title, "success")

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// NewVersion 上传新版本
func (kc *KnowledgeController) NewVersion(c *gin.Context) {
	id := c.Param("id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	changeNote := c.PostForm("change_note")

	userID, _ := c.Get("userID")
	db := config.GetDB()

	var kb models.KnowledgeBase
	if err := db.First(&kb, id).Error; err != nil {
		utils.NotFound(c, "资料不存在")
		return
	}

	// 保存旧版本记录
	version := models.KBVersion{
		KnowledgeID: kb.ID,
		Version:     kb.Version,
		FilePath:    kb.FilePath,
		UploadedBy:  kb.UploadedBy,
		CreatedAt:   kb.UpdatedAt,
	}
	db.Create(&version)

	// 上传新文件
	uploadDir := filepath.Join(config.UploadPath, "knowledge", time.Now().Format("200601"))
	os.MkdirAll(uploadDir, 0755)

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), userID, ext)
	filePath := filepath.Join(uploadDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		utils.ServerError(c, "保存文件失败")
		return
	}
	defer out.Close()

	io.Copy(out, file)

	// 更新版本号
	newVersion := incrementVersion(kb.Version)

	db.Model(&kb).Updates(map[string]interface{}{
		"file_path": filePath,
		"file_size": header.Size,
		"mime_type": header.Header.Get("Content-Type"),
		"version":   newVersion,
	})

	// 记录版本变更说明
	if changeNote != "" {
		versionRecord := models.KBVersion{
			KnowledgeID: kb.ID,
			Version:     newVersion,
			FilePath:    filePath,
			ChangeNote:  changeNote,
			UploadedBy:  userID.(uint),
			CreatedAt:   time.Now(),
		}
		db.Create(&versionRecord)
	}

	// 记录日志
	middleware.LogOperation(c, "new_version", "knowledge", "knowledge", kb.ID, kb.Title, "上传新版本: "+newVersion, "success")

	utils.SuccessWithMessage(c, "新版本上传成功", gin.H{"version": newVersion})
}

// GetVersions 获取版本历史
func (kc *KnowledgeController) GetVersions(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var versions []models.KBVersion
	db.Where("knowledge_id = ?", id).Order("created_at DESC").Find(&versions)

	utils.Success(c, versions)
}

// GetCategories 获取分类列表
func (kc *KnowledgeController) GetCategories(c *gin.Context) {
	db := config.GetDB()
	var categories []models.KBCategory
	db.Order("sort_order").Find(&categories)
	utils.Success(c, categories)
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	ParentID    uint   `json:"parent_id"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// CreateCategory 创建分类
func (kc *KnowledgeController) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写分类名称")
		return
	}

	db := config.GetDB()
	category := models.KBCategory{
		Name:        req.Name,
		ParentID:    req.ParentID,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	if err := db.Create(&category).Error; err != nil {
		utils.ServerError(c, "创建失败")
		return
	}

	utils.SuccessWithMessage(c, "创建成功", category)
}

// UpdateCategory 更新分类
func (kc *KnowledgeController) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	db := config.GetDB()
	var category models.KBCategory
	if err := db.First(&category, id).Error; err != nil {
		utils.NotFound(c, "分类不存在")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	updates["parent_id"] = req.ParentID
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["sort_order"] = req.SortOrder

	db.Model(&category).Updates(updates)

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteCategory 删除分类
func (kc *KnowledgeController) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var category models.KBCategory
	if err := db.First(&category, id).Error; err != nil {
		utils.NotFound(c, "分类不存在")
		return
	}

	// 检查是否有资料使用该分类
	var count int64
	db.Model(&models.KnowledgeBase{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "该分类下有资料，无法删除")
		return
	}

	db.Delete(&category)

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// GetHotItems 获取热门资料
func (kc *KnowledgeController) GetHotItems(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	db := config.GetDB()
	var items []models.KnowledgeBase
	db.Where("status = ?", "published").
		Order("view_count DESC, download_count DESC").
		Limit(limit).Find(&items)

	utils.Success(c, items)
}

// incrementVersion 递增版本号
func incrementVersion(version string) string {
	var major, minor int
	fmt.Sscanf(version, "%d.%d", &major, &minor)
	minor++
	if minor >= 10 {
		major++
		minor = 0
	}
	return fmt.Sprintf("%d.%d", major, minor)
}
