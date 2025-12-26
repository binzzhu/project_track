package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"project-flow/config"
	"project-flow/middleware"
	"project-flow/models"
	"project-flow/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ExpenseController struct{}

// List 获取费用记录列表
func (ec *ExpenseController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	projectID := c.Query("project_id")
	expenseType := c.Query("expense_type")
	reimbursedBy := c.Query("reimbursed_by")

	db := config.GetDB()
	var expenses []models.Expense
	var total int64

	query := db.Model(&models.Expense{}).
		Preload("Project").
		Preload("ReimbursedUser")

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if expenseType != "" {
		query = query.Where("expense_type = ?", expenseType)
	}
	if reimbursedBy != "" {
		query = query.Where("reimbursed_by = ?", reimbursedBy)
	}

	// 所有用户都可以查看所有费用记录，不做权限过滤

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&expenses)

	utils.SuccessPage(c, expenses, total, page, pageSize)
}

// Get 获取费用记录详情
func (ec *ExpenseController) Get(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var expense models.Expense
	if err := db.Preload("Project").
		Preload("ReimbursedUser").
		First(&expense, id).Error; err != nil {
		utils.NotFound(c, "费用记录不存在")
		return
	}

	utils.Success(c, expense)
}

// CreateExpenseRequest 创建费用记录请求
type CreateExpenseRequest struct {
	ProjectID   uint    `json:"project_id" binding:"required"`
	ExpenseType string  `json:"expense_type" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	ExpenseDate string  `json:"expense_date" binding:"required"`
	Description string  `json:"description"`
	Remark      string  `json:"remark"`
}

// Create 创建费用记录
func (ec *ExpenseController) Create(c *gin.Context) {
	var req CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	// 验证费用类型
	validTypes := []string{"labor", "direct", "outsourcing", "other"}
	isValid := false
	for _, t := range validTypes {
		if req.ExpenseType == t {
			isValid = true
			break
		}
	}
	if !isValid {
		utils.BadRequest(c, "无效的费用类型")
		return
	}

	// 解析日期
	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		utils.BadRequest(c, "日期格式错误")
		return
	}

	userID, _ := c.Get("userID")
	db := config.GetDB()

	expense := models.Expense{
		ProjectID:    req.ProjectID,
		ExpenseType:  req.ExpenseType,
		Amount:       req.Amount,
		ExpenseDate:  &expenseDate,
		Description:  req.Description,
		ReimbursedBy: userID.(uint),
		Remark:       req.Remark,
	}

	if err := db.Create(&expense).Error; err != nil {
		utils.ServerError(c, "创建失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "create", "expense", "expense", expense.ID, fmt.Sprintf("%.2f", expense.Amount), "创建费用记录", "success")

	utils.SuccessWithMessage(c, "创建成功", expense)
}

// Update 更新费用记录
func (ec *ExpenseController) Update(c *gin.Context) {
	id := c.Param("id")

	var req CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	db := config.GetDB()
	var expense models.Expense
	if err := db.First(&expense, id).Error; err != nil {
		utils.NotFound(c, "费用记录不存在")
		return
	}

	// 权限检查：只有报账人本人或系统管理员可以修改
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if expense.ReimbursedBy != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只能修改自己的费用记录")
		return
	}

	// 解析日期
	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		utils.BadRequest(c, "日期格式错误")
		return
	}

	expense.ProjectID = req.ProjectID
	expense.ExpenseType = req.ExpenseType
	expense.Amount = req.Amount
	expense.ExpenseDate = &expenseDate
	expense.Description = req.Description
	expense.Remark = req.Remark

	db.Save(&expense)

	// 记录日志
	middleware.LogOperation(c, "update", "expense", "expense", expense.ID, fmt.Sprintf("%.2f", expense.Amount), "更新费用记录", "success")

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// Delete 删除费用记录
func (ec *ExpenseController) Delete(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var expense models.Expense
	if err := db.First(&expense, id).Error; err != nil {
		utils.NotFound(c, "费用记录不存在")
		return
	}

	// 权限检查：只有报账人本人或系统管理员可以删除
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if expense.ReimbursedBy != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只能删除自己的费用记录")
		return
	}

	// 删除凭证文件
	if expense.VoucherPath != "" {
		filePaths := strings.Split(expense.VoucherPath, ",")
		for _, path := range filePaths {
			os.Remove(path)
		}
	}

	db.Delete(&expense)

	// 记录日志
	middleware.LogOperation(c, "delete", "expense", "expense", expense.ID, fmt.Sprintf("%.2f", expense.Amount), "删除费用记录", "success")

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// UploadVoucher 上传凭证文件
func (ec *ExpenseController) UploadVoucher(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var expense models.Expense
	if err := db.First(&expense, id).Error; err != nil {
		utils.NotFound(c, "费用记录不存在")
		return
	}

	// 权限检查：只有报账人本人或系统管理员可以上传凭证
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if expense.ReimbursedBy != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只能上传自己的费用凭证")
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.BadRequest(c, "请选择文件")
		return
	}

	// 创建上传目录
	uploadDir := filepath.Join(config.UploadPath, "expense", time.Now().Format("200601"))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ServerError(c, "创建目录失败")
		return
	}

	var filePaths []string
	for _, file := range files {
		// 保留原始文件名，但添加时间戳前缀以避免重名
		originalName := file.Filename
		// 使用时间戳作为前缀，保留原始文件名
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), originalName)
		filePath := filepath.Join(uploadDir, filename)

		// 保存文件
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			utils.ServerError(c, "保存文件失败")
			return
		}

		filePaths = append(filePaths, filePath)
	}

	// 更新凭证路径
	if expense.VoucherPath != "" {
		filePaths = append([]string{expense.VoucherPath}, filePaths...)
	}
	expense.VoucherPath = strings.Join(filePaths, ",")
	db.Save(&expense)

	// 记录日志
	middleware.LogOperation(c, "upload_voucher", "expense", "expense", expense.ID, fmt.Sprintf("%.2f", expense.Amount), "上传费用凭证", "success")

	utils.SuccessWithMessage(c, "上传成功", nil)
}

// DownloadVoucher 下载凭证文件
func (ec *ExpenseController) DownloadVoucher(c *gin.Context) {
	id := c.Param("id")
	fileIndex, _ := strconv.Atoi(c.Query("index"))

	db := config.GetDB()
	var expense models.Expense
	if err := db.First(&expense, id).Error; err != nil {
		utils.NotFound(c, "费用记录不存在")
		return
	}

	if expense.VoucherPath == "" {
		utils.NotFound(c, "没有凭证文件")
		return
	}

	filePaths := strings.Split(expense.VoucherPath, ",")
	if fileIndex < 0 || fileIndex >= len(filePaths) {
		utils.BadRequest(c, "文件索引错误")
		return
	}

	filePath := filePaths[fileIndex]
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.NotFound(c, "文件不存在")
		return
	}

	c.File(filePath)
}

// DeleteVoucher 删除凭证文件
func (ec *ExpenseController) DeleteVoucher(c *gin.Context) {
	id := c.Param("id")
	fileIndex, _ := strconv.Atoi(c.Query("index"))

	db := config.GetDB()
	var expense models.Expense
	if err := db.First(&expense, id).Error; err != nil {
		utils.NotFound(c, "费用记录不存在")
		return
	}

	// 权限检查：只有报账人本人或系统管理员可以删除凭证
	userID, _ := c.Get("userID")
	roleCode, _ := c.Get("roleCode")
	if expense.ReimbursedBy != userID.(uint) && roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只能删除自己的费用凭证")
		return
	}

	if expense.VoucherPath == "" {
		utils.NotFound(c, "没有凭证文件")
		return
	}

	filePaths := strings.Split(expense.VoucherPath, ",")
	if fileIndex < 0 || fileIndex >= len(filePaths) {
		utils.BadRequest(c, "文件索引错误")
		return
	}

	// 删除物理文件
	filePath := filePaths[fileIndex]
	if err := os.Remove(filePath); err != nil {
		// 文件可能已经不存在，不影响数据库更新
		fmt.Printf("删除文件失败: %v\n", err)
	}

	// 从数组中移除该路径
	filePaths = append(filePaths[:fileIndex], filePaths[fileIndex+1:]...)
	expense.VoucherPath = strings.Join(filePaths, ",")
	db.Save(&expense)

	// 记录日志
	middleware.LogOperation(c, "delete_voucher", "expense", "expense", expense.ID, fmt.Sprintf("%.2f", expense.Amount), "删除费用凭证", "success")

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// GetStatistics 获取费用统计
func (ec *ExpenseController) GetStatistics(c *gin.Context) {
	projectID := c.Query("project_id")

	db := config.GetDB()
	query := db.Model(&models.Expense{})

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	// 按费用类型统计
	var stats []struct {
		ExpenseType string  `json:"expense_type"`
		Total       float64 `json:"total"`
	}

	query.Select("expense_type, SUM(amount) as total").
		Group("expense_type").
		Scan(&stats)

	utils.Success(c, stats)
}

// GetProjectComparison 获取项目费用对比统计
func (ec *ExpenseController) GetProjectComparison(c *gin.Context) {
	db := config.GetDB()

	// 获取所有项目及其预算
	var projects []models.Project
	db.Select("id, name, labor_cost, direct_cost, outsourcing_cost, other_cost").Find(&projects)

	// 获取每个项目的实际费用
	type ExpenseStat struct {
		ProjectID   uint    `json:"project_id"`
		ExpenseType string  `json:"expense_type"`
		Total       float64 `json:"total"`
	}
	var expenseStats []ExpenseStat
	db.Model(&models.Expense{}).
		Select("project_id, expense_type, SUM(amount) as total").
		Group("project_id, expense_type").
		Scan(&expenseStats)

	// 构建统计结果
	type ProjectComparison struct {
		ProjectID         uint    `json:"project_id"`
		ProjectName       string  `json:"project_name"`
		LaborBudget       float64 `json:"labor_budget"`
		LaborActual       float64 `json:"labor_actual"`
		LaborRate         float64 `json:"labor_rate"`
		DirectBudget      float64 `json:"direct_budget"`
		DirectActual      float64 `json:"direct_actual"`
		DirectRate        float64 `json:"direct_rate"`
		OutsourcingBudget float64 `json:"outsourcing_budget"`
		OutsourcingActual float64 `json:"outsourcing_actual"`
		OutsourcingRate   float64 `json:"outsourcing_rate"`
		OtherBudget       float64 `json:"other_budget"`
		OtherActual       float64 `json:"other_actual"`
		OtherRate         float64 `json:"other_rate"`
		TotalBudget       float64 `json:"total_budget"`
		TotalActual       float64 `json:"total_actual"`
		TotalRate         float64 `json:"total_rate"`
	}

	var result []ProjectComparison
	for _, project := range projects {
		comp := ProjectComparison{
			ProjectID:         project.ID,
			ProjectName:       project.Name,
			LaborBudget:       project.LaborCost,
			DirectBudget:      project.DirectCost,
			OutsourcingBudget: project.OutsourcingCost,
			OtherBudget:       project.OtherCost,
		}

		// 计算实际费用
		for _, stat := range expenseStats {
			if stat.ProjectID == project.ID {
				switch stat.ExpenseType {
				case "labor":
					comp.LaborActual = stat.Total
				case "direct":
					comp.DirectActual = stat.Total
				case "outsourcing":
					comp.OutsourcingActual = stat.Total
				case "other":
					comp.OtherActual = stat.Total
				}
			}
		}

		// 计算执行率
		if comp.LaborBudget > 0 {
			comp.LaborRate = comp.LaborActual / comp.LaborBudget * 100
		}
		if comp.DirectBudget > 0 {
			comp.DirectRate = comp.DirectActual / comp.DirectBudget * 100
		}
		if comp.OutsourcingBudget > 0 {
			comp.OutsourcingRate = comp.OutsourcingActual / comp.OutsourcingBudget * 100
		}
		if comp.OtherBudget > 0 {
			comp.OtherRate = comp.OtherActual / comp.OtherBudget * 100
		}

		// 计算总计
		comp.TotalBudget = comp.LaborBudget + comp.DirectBudget + comp.OutsourcingBudget + comp.OtherBudget
		comp.TotalActual = comp.LaborActual + comp.DirectActual + comp.OutsourcingActual + comp.OtherActual
		if comp.TotalBudget > 0 {
			comp.TotalRate = comp.TotalActual / comp.TotalBudget * 100
		}

		result = append(result, comp)
	}

	utils.Success(c, result)
}
