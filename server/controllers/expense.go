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
	"github.com/xuri/excelize/v2"
)

type ExpenseController struct{}

// List 获取费用记录列表
func (ec *ExpenseController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	projectID := c.Query("project_id")
	projectCode := c.Query("project_code")
	reimbursedPersonName := c.Query("reimbursed_person_name")
	documentNo := c.Query("document_no")
	businessScene := c.Query("business_scene")
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
	if projectCode != "" {
		query = query.Where("project_code = ?", projectCode)
	}
	if reimbursedPersonName != "" {
		query = query.Where("reimbursed_person_name LIKE ?", "%"+reimbursedPersonName+"%")
	}
	if documentNo != "" {
		query = query.Where("document_no LIKE ?", "%"+documentNo+"%")
	}
	if businessScene != "" {
		query = query.Where("business_scene LIKE ?", "%"+businessScene+"%")
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

	// 添加调试信息：输出当前用户ID和费用记录的reimbursed_by
	userID, _ := c.Get("userID")
	fmt.Printf("调试信息 - 当前用户ID: %v, 费用记录ReimbursedBy: %v, 报账人姓名: %s\n",
		userID, expense.ReimbursedBy, expense.ReimbursedPersonName)

	utils.Success(c, expense)
}

type CreateExpenseRequest struct {
	ProjectID            *uint   `json:"project_id"`
	ProjectCode          string  `json:"project_code"`
	DocumentNo           string  `json:"document_no" binding:"required"`
	ExpenseType          string  `json:"expense_type"`
	ReimbursementAmount  float64 `json:"reimbursement_amount" binding:"required"`
	PaymentAmount        float64 `json:"payment_amount" binding:"required"`
	InvoiceAmountExclTax float64 `json:"invoice_amount_excl_tax" binding:"required"`
	InvoiceAmountInclTax float64 `json:"invoice_amount_incl_tax" binding:"required"`
	AllocationAmount     float64 `json:"allocation_amount" binding:"required"`
	Summary              string  `json:"summary" binding:"required"`
	BusinessScene        string  `json:"business_scene" binding:"required"`
	DepartmentName       string  `json:"department_name"`
	DocumentStatus       string  `json:"document_status"`
}

// Create 创建费用记录
func (ec *ExpenseController) Create(c *gin.Context) {
	var req CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	userID, _ := c.Get("userID")
	db := config.GetDB()

	expense := models.Expense{
		ProjectID:            req.ProjectID,
		ProjectCode:          req.ProjectCode,
		DocumentNo:           req.DocumentNo,
		ExpenseType:          req.ExpenseType,
		ReimbursementAmount:  req.ReimbursementAmount,
		PaymentAmount:        req.PaymentAmount,
		InvoiceAmountExclTax: req.InvoiceAmountExclTax,
		InvoiceAmountInclTax: req.InvoiceAmountInclTax,
		AllocationAmount:     req.AllocationAmount,
		Summary:              req.Summary,
		BusinessScene:        req.BusinessScene,
		DepartmentName:       req.DepartmentName,
		DocumentStatus:       req.DocumentStatus,
		ReimbursedBy:         userID.(uint),
		CreatedBy:            userID.(uint),
		IsClassified:         req.ProjectID != nil,
	}

	if err := db.Create(&expense).Error; err != nil {
		utils.ServerError(c, "创建失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "create", "expense", "expense", expense.ID, expense.DocumentNo, "创建费用记录", "success")

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

	expense.ProjectID = req.ProjectID
	expense.ProjectCode = req.ProjectCode
	expense.ExpenseType = req.ExpenseType
	expense.ReimbursementAmount = req.ReimbursementAmount
	expense.PaymentAmount = req.PaymentAmount
	expense.InvoiceAmountExclTax = req.InvoiceAmountExclTax
	expense.InvoiceAmountInclTax = req.InvoiceAmountInclTax
	expense.AllocationAmount = req.AllocationAmount
	expense.Summary = req.Summary
	expense.BusinessScene = req.BusinessScene
	expense.DepartmentName = req.DepartmentName
	expense.DocumentStatus = req.DocumentStatus
	expense.IsClassified = req.ProjectID != nil

	db.Save(&expense)

	// 记录日志
	middleware.LogOperation(c, "update", "expense", "expense", expense.ID, expense.DocumentNo, "更新费用记录", "success")

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
	middleware.LogOperation(c, "delete", "expense", "expense", expense.ID, expense.DocumentNo, "删除费用记录", "success")

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
	db.Select("id, name, innovation_code, labor_cost, direct_cost, outsourcing_cost, other_cost").Find(&projects)

	type ExpenseStat struct {
		ProjectID    uint    `json:"project_id"`
		ExpenseType  string  `json:"expense_type"`
		TotalInclTax float64 `json:"total_incl_tax"`
		TotalExclTax float64 `json:"total_excl_tax"`
	}
	var expenseStats []ExpenseStat
	db.Model(&models.Expense{}).
		Select("project_id, expense_type, SUM(reimbursement_amount) as total_incl_tax, SUM(allocation_amount) as total_excl_tax").
		Where("project_id IS NOT NULL AND is_classified = ?", true).
		Group("project_id, expense_type").
		Scan(&expenseStats)

	// 构建统计结果
	type ProjectComparison struct {
		ProjectID                uint    `json:"project_id"`
		ProjectName              string  `json:"project_name"`
		InnovationCode           string  `json:"innovation_code"`
		LaborBudget              float64 `json:"labor_budget"`
		LaborActualInclTax       float64 `json:"labor_actual_incl_tax"`
		LaborActualExclTax       float64 `json:"labor_actual_excl_tax"`
		DirectBudget             float64 `json:"direct_budget"`
		DirectActualInclTax      float64 `json:"direct_actual_incl_tax"`
		DirectActualExclTax      float64 `json:"direct_actual_excl_tax"`
		OutsourcingBudget        float64 `json:"outsourcing_budget"`
		OutsourcingActualInclTax float64 `json:"outsourcing_actual_incl_tax"`
		OutsourcingActualExclTax float64 `json:"outsourcing_actual_excl_tax"`
		OtherBudget              float64 `json:"other_budget"`
		OtherActualInclTax       float64 `json:"other_actual_incl_tax"`
		OtherActualExclTax       float64 `json:"other_actual_excl_tax"`
		TotalBudget              float64 `json:"total_budget"`
		TotalActualInclTax       float64 `json:"total_actual_incl_tax"`
		TotalActualExclTax       float64 `json:"total_actual_excl_tax"`
	}

	var result []ProjectComparison
	for _, project := range projects {
		comp := ProjectComparison{
			ProjectID:         project.ID,
			ProjectName:       project.Name,
			InnovationCode:    project.InnovationCode,
			LaborBudget:       project.LaborCost,
			DirectBudget:      project.DirectCost,
			OutsourcingBudget: project.OutsourcingCost,
			OtherBudget:       project.OtherCost,
		}

		// 计算实际费用（按类型分组）
		for _, stat := range expenseStats {
			if stat.ProjectID == project.ID {
				switch stat.ExpenseType {
				case "labor":
					comp.LaborActualInclTax = stat.TotalInclTax
					comp.LaborActualExclTax = stat.TotalExclTax
				case "direct":
					comp.DirectActualInclTax = stat.TotalInclTax
					comp.DirectActualExclTax = stat.TotalExclTax
				case "outsourcing":
					comp.OutsourcingActualInclTax = stat.TotalInclTax
					comp.OutsourcingActualExclTax = stat.TotalExclTax
				case "other":
					comp.OtherActualInclTax = stat.TotalInclTax
					comp.OtherActualExclTax = stat.TotalExclTax
				}
			}
		}

		// 计算总计
		comp.TotalBudget = comp.LaborBudget + comp.DirectBudget + comp.OutsourcingBudget + comp.OtherBudget
		comp.TotalActualInclTax = comp.LaborActualInclTax + comp.DirectActualInclTax + comp.OutsourcingActualInclTax + comp.OtherActualInclTax
		comp.TotalActualExclTax = comp.LaborActualExclTax + comp.DirectActualExclTax + comp.OutsourcingActualExclTax + comp.OtherActualExclTax

		result = append(result, comp)
	}

	utils.Success(c, result)
}

// GetNonProjectExpenseStats 获取非研发项目费用统计（按业务场景分组）
func (ec *ExpenseController) GetNonProjectExpenseStats(c *gin.Context) {
	db := config.GetDB()

	type BusinessSceneStat struct {
		BusinessScene string  `json:"business_scene"`
		TotalInclTax  float64 `json:"total_incl_tax"`
		TotalExclTax  float64 `json:"total_excl_tax"`
	}

	var stats []BusinessSceneStat
	db.Model(&models.Expense{}).
		Select("business_scene, SUM(reimbursement_amount) as total_incl_tax, SUM(allocation_amount) as total_excl_tax").
		Where("project_id IS NULL AND (expense_type IS NULL OR expense_type = '')").
		Group("business_scene").
		Order("total_incl_tax DESC").
		Scan(&stats)

	// 计算总计
	var grandTotalInclTax float64
	var grandTotalExclTax float64
	for _, stat := range stats {
		grandTotalInclTax += stat.TotalInclTax
		grandTotalExclTax += stat.TotalExclTax
	}

	// 返回结果
	result := map[string]interface{}{
		"data":                 stats,
		"grand_total_incl_tax": grandTotalInclTax,
		"grand_total_excl_tax": grandTotalExclTax,
	}

	utils.Success(c, result)
}

// ImportExpenses 从 Excel 导入费用记录
func (ec *ExpenseController) ImportExpenses(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请上传Excel文件")
		return
	}

	// 保存临时文件
	uploadDir := filepath.Join(config.UploadPath, "temp")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ServerError(c, "创建目录失败")
		return
	}

	filePath := filepath.Join(uploadDir, fmt.Sprintf("expense_import_%d.xlsx", time.Now().Unix()))
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.ServerError(c, "保存文件失败")
		return
	}
	defer os.Remove(filePath) // 处理完后删除临时文件

	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		utils.ServerError(c, "打开Excel文件失败")
		return
	}
	defer f.Close()

	// 读取第一个工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		utils.BadRequest(c, "Excel文件为空")
		return
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		utils.ServerError(c, "读取数据失败")
		return
	}

	if len(rows) < 2 {
		utils.BadRequest(c, "Excel文件没有数据")
		return
	}

	userID, _ := c.Get("userID")
	db := config.GetDB()

	// 预先查询所有用户，构建姓名到ID的映射
	var users []models.User
	db.Select("id, name").Find(&users)
	userNameMap := make(map[string]uint)
	for _, user := range users {
		if user.Name != "" {
			userNameMap[user.Name] = user.ID
			fmt.Printf("用户映射 - 姓名: '%s', ID: %d\n", user.Name, user.ID)
		}
	}

	var expenses []models.Expense
	var successCount, errorCount int
	var errorMessages []string

	// 从第2行开始读取数据（第1行是表头）
	for i, row := range rows {
		if i == 0 {
			continue // 跳过表头
		}

		// 如果整行都是空的，跳过（不计入任何统计）
		// 1. 检查行是否为空或列数为0
		if len(row) == 0 {
			continue
		}

		// 2. 先检查单据编号（必填项，第4列，索引3）
		// 如果单据编号为空，直接跳过此行（视为空行）
		documentNo := getString(row, 3)
		if documentNo == "" {
			continue // 单据编号为空视为空行，直接跳过
		}

		// 解析数据（使用安全的getString和getFloat，防止索引越界）
		// 根据报账人姓名查找对应的用户ID
		reimbursedPersonName := getString(row, 8) // 报账人姓名
		reimbursedByID := userID.(uint)           // 默认为导入者
		if reimbursedPersonName != "" {
			if uid, exists := userNameMap[reimbursedPersonName]; exists {
				reimbursedByID = uid // 如果找到匹配的用户，使用该用户ID
				fmt.Printf("导入调试 - 报账人: '%s', 匹配到用户ID: %d\n", reimbursedPersonName, uid)
			} else {
				fmt.Printf("导入警告 - 报账人: '%s', 未找到匹配的用户，使用导入者ID: %d\n", reimbursedPersonName, userID.(uint))
			}
		}

		expense := models.Expense{
			ProjectCode:          getString(row, 0),             // 创新项目编码
			UnitCode:             getString(row, 1),             // 单位编号
			UnitName:             getString(row, 2),             // 单位名称
			DocumentNo:           documentNo,                    // 单据编号
			BusinessScene:        getString(row, 4),             // 业务场景
			CrossIndustryCode:    getString(row, 5),             // 跨行业业务编码
			Summary:              getString(row, 6),             // 摘要
			DepartmentName:       getString(row, 7),             // 部门名称
			ReimbursedPersonName: reimbursedPersonName,          // 报账人姓名
			DocumentStatus:       getString(row, 9),             // 单据状态
			FrozenStatus:         getString(row, 10),            // 冻结状态
			ReimbursementAmount:  getFloat(row, 11),             // 报账金额
			PaymentAmount:        getFloat(row, 12),             // 支付金额
			WriteOffAmount:       getFloat(row, 13),             // 核销金额
			InvoiceAmountExclTax: getFloat(row, 14),             // 发票不含税金额
			InvoiceAmountInclTax: getFloat(row, 15),             // 发票含税金额
			CurrentProcess:       getString(row, 16),            // 当前处理环节
			CurrentProcessor:     getString(row, 17),            // 当前处理人
			PhysicalStatus:       getString(row, 18),            // 实物状态
			PhysicalLocation:     getString(row, 19),            // 实物位置
			DocumentType:         getString(row, 20),            // 单据类型
			DocumentTypeName:     getString(row, 21),            // 单据类型名称
			SupplierCode:         getString(row, 22),            // 供应商编号
			SupplierName:         getString(row, 23),            // 供应商名称
			CreateDocTime:        parseTime(getString(row, 24)), // 制单时间
			SubmitTime:           parseTime(getString(row, 25)), // 提交时间
			InternalCode:         getString(row, 26),            // 内码
			SharedProcess:        getString(row, 27),            // 共享处理环节
			SharedProcessor:      getString(row, 28),            // 共享处理人
			PaymentAccount:       getString(row, 29),            // 付款账号
			ReimbursedBy:         reimbursedByID,                // 根据报账人姓名匹配的用户ID
			CreatedBy:            userID.(uint),                 // 导入者ID
			IsClassified:         false,                         // 默认未归类
		}

		expenses = append(expenses, expense)
	}

	// 打印调试信息
	fmt.Printf("Excel总行数: %d, 解析出的有效记录数: %d, 错误数: %d\n", len(rows), len(expenses), errorCount)

	// 批量插入数据库（先去重）
	if len(expenses) > 0 {
		// 先查询数据库中已存在的单据编号
		documentNos := make([]string, len(expenses))
		for i, exp := range expenses {
			documentNos[i] = exp.DocumentNo
		}

		// 查找已存在的记录
		var existingExpenses []models.Expense
		db.Where("`document_no` IN ?", documentNos).Find(&existingExpenses)

		// 构建已存在记录的map
		existingMap := make(map[string]uint)
		for _, exp := range existingExpenses {
			existingMap[exp.DocumentNo] = exp.ID
		}

		// 分类处理：新增和更新
		var toCreate []models.Expense
		var toUpdate []models.Expense
		var updateCount int

		for _, expense := range expenses {
			if existingID, exists := existingMap[expense.DocumentNo]; exists {
				// 已存在，更新
				expense.ID = existingID
				toUpdate = append(toUpdate, expense)
			} else {
				// 不存在，新增
				toCreate = append(toCreate, expense)
			}
		}

		// 执行新增
		if len(toCreate) > 0 {
			if err := db.Create(&toCreate).Error; err != nil {
				fmt.Printf("新增记录错误: %v\n", err)
				errorMessages = append(errorMessages, fmt.Sprintf("部分记录新增失败: %v", err))
				errorCount += len(toCreate)
			}
		}

		// 执行更新
		for _, expense := range toUpdate {
			if err := db.Model(&models.Expense{}).Where("id = ?", expense.ID).Updates(map[string]interface{}{
				"project_code":            expense.ProjectCode,
				"unit_code":               expense.UnitCode,
				"unit_name":               expense.UnitName,
				"business_scene":          expense.BusinessScene,
				"cross_industry_code":     expense.CrossIndustryCode,
				"summary":                 expense.Summary,
				"department_name":         expense.DepartmentName,
				"reimbursed_person_name":  expense.ReimbursedPersonName,
				"document_status":         expense.DocumentStatus,
				"frozen_status":           expense.FrozenStatus,
				"reimbursement_amount":    expense.ReimbursementAmount,
				"payment_amount":          expense.PaymentAmount,
				"write_off_amount":        expense.WriteOffAmount,
				"invoice_amount_excl_tax": expense.InvoiceAmountExclTax,
				"invoice_amount_incl_tax": expense.InvoiceAmountInclTax,
				"current_process":         expense.CurrentProcess,
				"current_processor":       expense.CurrentProcessor,
				"physical_status":         expense.PhysicalStatus,
				"physical_location":       expense.PhysicalLocation,
				"document_type":           expense.DocumentType,
				"document_type_name":      expense.DocumentTypeName,
				"supplier_code":           expense.SupplierCode,
				"supplier_name":           expense.SupplierName,
				"create_doc_time":         expense.CreateDocTime,
				"submit_time":             expense.SubmitTime,
				"internal_code":           expense.InternalCode,
				"shared_process":          expense.SharedProcess,
				"shared_processor":        expense.SharedProcessor,
				"payment_account":         expense.PaymentAccount,
				"reimbursed_by":           expense.ReimbursedBy,
			}).Error; err != nil {
				fmt.Printf("更新记录失败(单据号:%s): %v\n", expense.DocumentNo, err)
				errorMessages = append(errorMessages, fmt.Sprintf("单据编号: %s", expense.DocumentNo))
				errorCount++
			} else {
				updateCount++
			}
		}

		// 重新计算成功数量
		successCount = len(toCreate) + updateCount

		// 记录日志
		middleware.LogOperation(c, "import", "expense", "expense", 0,
			fmt.Sprintf("新增%d条,更新%d条,失败%d条", len(toCreate), updateCount, errorCount),
			"导入费用记录", "success")

		utils.Success(c, gin.H{
			"success_count": successCount,
			"create_count":  len(toCreate),
			"update_count":  updateCount,
			"error_count":   errorCount,
			"errors":        errorMessages,
			"message":       fmt.Sprintf("导入完成：新增%d条，更新%d条", len(toCreate), updateCount),
		})
		return
	}
}

// DeleteAll 一键删除所有费用记录（仅管理员）
func (ec *ExpenseController) DeleteAll(c *gin.Context) {
	// 权限检查：只有管理员可以一键删除
	roleCode, _ := c.Get("roleCode")
	if roleCode != config.RoleAdmin {
		utils.Forbidden(c, "只有管理员可以执行一键删除操作")
		return
	}

	db := config.GetDB()

	// 查询所有记录数量
	var count int64
	db.Model(&models.Expense{}).Count(&count)

	if count == 0 {
		utils.BadRequest(c, "没有需要删除的记录")
		return
	}

	// 删除所有凭证文件
	var expenses []models.Expense
	db.Select("voucher_path").Find(&expenses)
	for _, expense := range expenses {
		if expense.VoucherPath != "" {
			filePaths := strings.Split(expense.VoucherPath, ",")
			for _, path := range filePaths {
				os.Remove(path)
			}
		}
	}

	// 执行删除（软删除）
	result := db.Where("1 = 1").Delete(&models.Expense{})
	if result.Error != nil {
		utils.ServerError(c, fmt.Sprintf("删除失败: %v", result.Error))
		return
	}

	// 记录日志
	middleware.LogOperation(c, "delete_all", "expense", "expense", 0,
		fmt.Sprintf("删除了%d条费用记录", count),
		"一键删除所有费用记录", "success")

	utils.SuccessWithMessage(c, fmt.Sprintf("成功删除%d条记录", count), nil)
}

// 辅助函数
func getString(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func getFloat(row []string, index int) float64 {
	if index >= len(row) {
		return 0
	}
	// 去除空格和千分位逗号
	valStr := strings.TrimSpace(row[index])
	valStr = strings.ReplaceAll(valStr, ",", "") // 移除逗号
	valStr = strings.ReplaceAll(valStr, " ", "") // 移除空格

	if valStr == "" {
		return 0
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		fmt.Printf("解析数字失败[%s]: %v\n", row[index], err)
		return 0
	}
	return val
}

func parseTime(timeStr string) *time.Time {
	if timeStr == "" {
		return nil
	}
	// 尝试多种日期格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return &t
		}
	}
	return nil
}
