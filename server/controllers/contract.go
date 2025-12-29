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

type ContractController struct{}

// CreateContractRequest 创建合同请求
type CreateContractRequest struct {
	ProjectID     uint    `json:"project_id" binding:"required"`
	ContractNo    string  `json:"contract_no"`
	ContractName  string  `json:"contract_name" binding:"required"`
	PartyA        string  `json:"party_a"`
	PartyB        string  `json:"party_b"`
	Amount        float64 `json:"amount"`
	SignDate      string  `json:"sign_date"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
	PaymentMethod string  `json:"payment_method"`
}

// List 获取合同列表
func (cc *ContractController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	projectID := c.Query("project_id")
	keyword := c.Query("keyword")

	db := config.GetDB()

	var contracts []models.Contract
	var total int64

	query := db.Model(&models.Contract{})

	// 所有角色都可以查看所有合同

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if keyword != "" {
		query = query.Where("contract_name LIKE ? OR contract_no LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&contracts)

	utils.SuccessPage(c, contracts, total, page, pageSize)
}

// Create 创建合同
func (cc *ContractController) Create(c *gin.Context) {
	var req CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写合同名称")
		return
	}

	userID, _ := c.Get("userID")
	db := config.GetDB()

	// 生成合同编号
	contractNo := req.ContractNo
	if contractNo == "" {
		contractNo = fmt.Sprintf("CON%s%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
	}

	contract := models.Contract{
		ProjectID:     req.ProjectID,
		ContractNo:    contractNo,
		ContractName:  req.ContractName,
		PartyA:        req.PartyA,
		PartyB:        req.PartyB,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Status:        "draft",
		CreatedBy:     userID.(uint),
	}

	if req.SignDate != "" {
		t, _ := time.Parse("2006-01-02", req.SignDate)
		contract.SignDate = &t
	}
	if req.StartDate != "" {
		t, _ := time.Parse("2006-01-02", req.StartDate)
		contract.StartDate = &t
	}
	if req.EndDate != "" {
		t, _ := time.Parse("2006-01-02", req.EndDate)
		contract.EndDate = &t
	}

	if err := db.Create(&contract).Error; err != nil {
		utils.ServerError(c, "创建合同失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "create", "contract", "contract", contract.ID, contract.ContractName, "创建合同: "+contract.ContractName, "success")

	utils.SuccessWithMessage(c, "创建成功", contract)
}

// Get 获取合同详情
func (cc *ContractController) Get(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var contract models.Contract
	if err := db.First(&contract, id).Error; err != nil {
		utils.NotFound(c, "合同不存在")
		return
	}

	// 所有角色都可以查看合同详情

	utils.Success(c, contract)
}

// UpdateContractRequest 更新合同请求
type UpdateContractRequest struct {
	ContractName  string  `json:"contract_name"`
	PartyA        string  `json:"party_a"`
	PartyB        string  `json:"party_b"`
	Amount        float64 `json:"amount"`
	SignDate      string  `json:"sign_date"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
}

// Update 更新合同
func (cc *ContractController) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	db := config.GetDB()
	var contract models.Contract
	if err := db.First(&contract, id).Error; err != nil {
		utils.NotFound(c, "合同不存在")
		return
	}

	updates := make(map[string]interface{})
	if req.ContractName != "" {
		updates["contract_name"] = req.ContractName
	}
	if req.PartyA != "" {
		updates["party_a"] = req.PartyA
	}
	if req.PartyB != "" {
		updates["party_b"] = req.PartyB
	}
	if req.Amount != 0 {
		updates["amount"] = req.Amount
	}
	if req.PaymentMethod != "" {
		updates["payment_method"] = req.PaymentMethod
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.SignDate != "" {
		t, _ := time.Parse("2006-01-02", req.SignDate)
		updates["sign_date"] = t
	}
	if req.StartDate != "" {
		t, _ := time.Parse("2006-01-02", req.StartDate)
		updates["start_date"] = t
	}
	if req.EndDate != "" {
		t, _ := time.Parse("2006-01-02", req.EndDate)
		updates["end_date"] = t
	}

	db.Model(&contract).Updates(updates)

	// 记录日志
	middleware.LogOperation(c, "update", "contract", "contract", contract.ID, contract.ContractName, "更新合同: "+contract.ContractName, "success")

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// UploadFile 上传合同文件
func (cc *ContractController) UploadFile(c *gin.Context) {
	id := c.Param("id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	userID, _ := c.Get("userID")
	db := config.GetDB()

	var contract models.Contract
	if err := db.First(&contract, id).Error; err != nil {
		utils.NotFound(c, "合同不存在")
		return
	}

	// 创建上传目录
	uploadDir := filepath.Join(config.UploadPath, "contracts", time.Now().Format("200601"))
	os.MkdirAll(uploadDir, 0755)

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

	io.Copy(out, file)

	// 更新合同文件路径
	db.Model(&contract).Update("file_path", filePath)

	// 记录日志
	middleware.LogOperation(c, "upload", "contract", "contract", contract.ID, contract.ContractName, "上传合同文件", "success")

	utils.SuccessWithMessage(c, "上传成功", nil)
}

// Delete 删除合同
func (cc *ContractController) Delete(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var contract models.Contract
	if err := db.First(&contract, id).Error; err != nil {
		utils.NotFound(c, "合同不存在")
		return
	}

	// 删除文件
	if contract.FilePath != "" {
		os.Remove(contract.FilePath)
	}

	db.Delete(&contract)

	// 记录日志
	middleware.LogOperation(c, "delete", "contract", "contract", contract.ID, contract.ContractName, "删除合同: "+contract.ContractName, "success")

	utils.SuccessWithMessage(c, "删除成功", nil)
}
