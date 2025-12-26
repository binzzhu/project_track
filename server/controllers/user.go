package controllers

import (
	"project-flow/config"
	"project-flow/middleware"
	"project-flow/models"
	"project-flow/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Department string `json:"department"`
	RoleID     uint   `json:"role_id" binding:"required"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Department string `json:"department"`
	RoleID     uint   `json:"role_id"`
	Status     *int   `json:"status"`
}

// List 获取用户列表
func (uc *UserController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	roleID := c.Query("role_id")
	status := c.Query("status")

	db := config.GetDB()
	var users []models.User
	var total int64

	query := db.Model(&models.User{}).Preload("Role")

	if keyword != "" {
		query = query.Where("username LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if roleID != "" {
		query = query.Where("role_id = ?", roleID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&users)

	utils.SuccessPage(c, users, total, page, pageSize)
}

// Create 创建用户
func (uc *UserController) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写完整信息")
		return
	}

	// 验证密码复杂度
	if !utils.ValidatePassword(req.Password) {
		utils.BadRequest(c, "密码至少8位，需包含大小写字母、数字和特殊符号")
		return
	}

	// 验证部门是否合法
	if req.Department != "" {
		validDept := false
		for _, dept := range config.AllowedDepartments {
			if req.Department == dept {
				validDept = true
				break
			}
		}
		if !validDept {
			utils.BadRequest(c, "部门信息不合法，请选择：BMS研发部、PACK研发部、综合部、外部单位")
			return
		}
	}

	db := config.GetDB()

	// 检查用户名是否存在
	var existing models.User
	if db.Where("username = ?", req.Username).First(&existing).RowsAffected > 0 {
		utils.Error(c, 400, "用户名已存在")
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ServerError(c, "密码加密失败")
		return
	}

	user := models.User{
		Username:   req.Username,
		Password:   hashedPassword,
		Name:       req.Name,
		Email:      req.Email,
		Phone:      req.Phone,
		Department: req.Department,
		RoleID:     req.RoleID,
		Status:     1,
	}

	if err := db.Create(&user).Error; err != nil {
		utils.ServerError(c, "创建用户失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "create", "user", "user", user.ID, user.Name, "创建用户: "+user.Name, "success")

	utils.SuccessWithMessage(c, "创建成功", user)
}

// Get 获取用户详情
func (uc *UserController) Get(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var user models.User
	if err := db.Preload("Role").First(&user, id).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// Update 更新用户
func (uc *UserController) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	db := config.GetDB()
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Department != "" {
		// 验证部门是否合法
		validDept := false
		for _, dept := range config.AllowedDepartments {
			if req.Department == dept {
				validDept = true
				break
			}
		}
		if !validDept {
			utils.BadRequest(c, "部门信息不合法，请选择：BMS研发部、PACK研发部、综合部、外部单位")
			return
		}
		updates["department"] = req.Department
	}
	if req.RoleID != 0 {
		updates["role_id"] = req.RoleID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := db.Model(&user).Updates(updates).Error; err != nil {
		utils.ServerError(c, "更新失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "update", "user", "user", user.ID, user.Name, "更新用户: "+user.Name, "success")

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// Delete 删除用户
func (uc *UserController) Delete(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	if err := db.Delete(&user).Error; err != nil {
		utils.ServerError(c, "删除失败")
		return
	}

	// 记录日志
	middleware.LogOperation(c, "delete", "user", "user", user.ID, user.Name, "删除用户: "+user.Name, "success")

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// ResetPassword 重置密码
func (uc *UserController) ResetPassword(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	// 重置为默认密码
	hashedPassword, _ := utils.HashPassword("Reset@123")
	user.Password = hashedPassword
	user.FailedLogins = 0
	user.LockedUntil = nil
	db.Save(&user)

	// 记录日志
	middleware.LogOperation(c, "reset_password", "user", "user", user.ID, user.Name, "重置密码: "+user.Name, "success")

	utils.SuccessWithMessage(c, "密码已重置为: Reset@123", nil)
}

// GetRoles 获取角色列表
func (uc *UserController) GetRoles(c *gin.Context) {
	db := config.GetDB()
	var roles []models.Role
	db.Find(&roles)
	utils.Success(c, roles)
}
