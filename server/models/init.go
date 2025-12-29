package models

import (
	"log"
	"project-flow/config"

	"golang.org/x/crypto/bcrypt"
)

// AutoMigrate 自动迁移数据库表
func AutoMigrate() {
	db := config.GetDB()

	err := db.AutoMigrate(
		&User{},
		&Role{},
		&Project{},
		&ProjectPhase{},
		&Task{},
		&Document{},
		&Contract{},
		&KnowledgeBase{},
		&KBCategory{},
		&KBVersion{},
		&OperationLog{},
		&ProjectMember{},
		&Expense{},
	)
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}
	log.Println("数据库迁移成功")
}

// InitDefaultData 初始化默认数据
func InitDefaultData() {
	db := config.GetDB()

	// 初始化角色（组织级）
	roles := []Role{
		{
			Name:        "系统管理员",
			Code:        config.RoleAdmin,
			Description: "系统管理员，拥有所有权限",
			Permissions: `["all"]`,
		},
		{
			Name:        "部门经理",
			Code:        config.RoleDeptManager,
			Description: "部门经理，可管理用户和查看所有项目",
			Permissions: `["user:manage","project:view","project:create","document:view","document:download","kb:all","log:view"]`,
		},
		{
			Name:        "组长",
			Code:        config.RoleTeamLeader,
			Description: "组长，可创建项目并查看所有项目",
			Permissions: `["project:view","project:create","document:view","document:download","kb:view","kb:download"]`,
		},
		{
			Name:        "组员",
			Code:        config.RoleTeamMember,
			Description: "组员，可查看所有项目和资料",
			Permissions: `["project:view","document:view","document:download","kb:view","kb:download"]`,
		},
	}

	for _, role := range roles {
		var existing Role
		if db.Where("code = ?", role.Code).First(&existing).RowsAffected == 0 {
			db.Create(&role)
		}
	}

	// 初始化管理员账号
	var adminRole Role
	db.Where("code = ?", config.RoleAdmin).First(&adminRole)

	var adminUser User
	if db.Where("username = ?", "admin").First(&adminUser).RowsAffected == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
		admin := User{
			Username:   "admin",
			Password:   string(hashedPassword),
			Name:       "系统管理员",
			Email:      "admin@example.com",
			Department: "系统部门",
			RoleID:     adminRole.ID,
			Status:     1,
		}
		db.Create(&admin)
	}

	// 初始化知识库分类
	categories := []KBCategory{
		{Name: "项目模板", Description: "项目相关模板文档", SortOrder: 1},
		{Name: "政策文件", Description: "政策法规文件", SortOrder: 2},
		{Name: "技术规范", Description: "技术规范标准", SortOrder: 3},
		{Name: "案例资料", Description: "项目案例资料", SortOrder: 4},
		{Name: "培训材料", Description: "培训学习材料", SortOrder: 5},
	}

	for _, cat := range categories {
		var existing KBCategory
		if db.Where("name = ?", cat.Name).First(&existing).RowsAffected == 0 {
			db.Create(&cat)
		}
	}

	log.Println("默认数据初始化完成")
}
