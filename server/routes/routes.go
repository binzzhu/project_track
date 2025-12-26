package routes

import (
	"project-flow/config"
	"project-flow/controllers"
	"project-flow/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// 控制器实例
	authCtrl := &controllers.AuthController{}
	userCtrl := &controllers.UserController{}
	projectCtrl := &controllers.ProjectController{}
	taskCtrl := &controllers.TaskController{}
	docCtrl := &controllers.DocumentController{}
	kbCtrl := &controllers.KnowledgeController{}
	contractCtrl := &controllers.ContractController{}
	logCtrl := &controllers.LogController{}
	expenseCtrl := &controllers.ExpenseController{}

	// API路由组
	api := r.Group("/api")
	{
		// 健康检查接口
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "message": "服务运行正常"})
		})

		// 公开接口
		api.POST("/login", authCtrl.Login)

		// 需要认证的接口
		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			// 认证相关
			auth.GET("/user/current", authCtrl.GetCurrentUser)
			auth.POST("/user/change-password", authCtrl.ChangePassword)
			auth.POST("/logout", authCtrl.Logout)

			// 用户管理
			users := auth.Group("/users")
			{
				// 查询用户列表（所有人可查询，用于项目负责人选择等）
				users.GET("", userCtrl.List)
				users.GET("/:id", userCtrl.Get)

				// 用户管理操作（仅管理员和部门经理）
				users.POST("", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager), userCtrl.Create)
				users.PUT("/:id", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager), userCtrl.Update)
				users.DELETE("/:id", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager), userCtrl.Delete)
				users.POST("/:id/reset-password", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager), userCtrl.ResetPassword)
			}
			auth.GET("/roles", userCtrl.GetRoles)

			// 项目管理（所有用户可查看，权限控制在控制器中实现）
			projects := auth.Group("/projects")
			{
				projects.GET("", projectCtrl.List)
				projects.GET("/statistics", projectCtrl.GetStatistics)
				projects.GET("/:id", projectCtrl.Get)

				// 创建项目（组长和组员）
				projects.POST("", middleware.RoleMiddleware(config.RoleTeamLeader, config.RoleTeamMember), projectCtrl.Create)
				// 修改/删除项目（权限在控制器中检查）
				projects.PUT("/:id", projectCtrl.Update)
				projects.DELETE("/:id", projectCtrl.Delete)

				// 阶段管理
				projects.GET("/:id/phases", projectCtrl.GetPhases)
				projects.POST("/:id/phases", projectCtrl.AddPhase) // 添加自定义阶段
				projects.PUT("/:id/phases/:phaseId", projectCtrl.UpdatePhase)
				projects.DELETE("/:id/phases/:phaseId", projectCtrl.DeletePhase) // 删除自定义阶段

				// 成员管理
				projects.GET("/:id/members", projectCtrl.GetMembers)
				projects.POST("/:id/members", projectCtrl.AddMember)
				projects.DELETE("/:id/members/:memberId", projectCtrl.RemoveMember)
			}

			// 任务管理（所有用户可查看）
			tasks := auth.Group("/tasks")
			{
				tasks.GET("", taskCtrl.List)
				tasks.GET("/statistics", taskCtrl.GetTaskStatistics)
				tasks.GET("/my", taskCtrl.GetMyTasks)
				tasks.GET(":id", taskCtrl.Get)

				// 创建/分配任务（组长和组员）
				tasks.POST("", middleware.RoleMiddleware(config.RoleTeamLeader, config.RoleTeamMember), taskCtrl.Create)
				tasks.POST("/batch", middleware.RoleMiddleware(config.RoleTeamLeader, config.RoleTeamMember), taskCtrl.BatchCreate)
				tasks.PUT("/:id", taskCtrl.Update)
				tasks.DELETE("/:id", taskCtrl.Delete) // 权限在控制器中检查（项目负责人）

				// 状态更新（所有人可更新自己的任务状态）
				tasks.PUT("/:id/status", taskCtrl.UpdateStatus)

				// 审核（组长和组员）
				tasks.POST("/:id/review", middleware.RoleMiddleware(config.RoleTeamLeader, config.RoleTeamMember), taskCtrl.ReviewTask)
			}

			// 文档管理（所有用户可查看和下载，上传权限在控制器中检查）
			docs := auth.Group("/documents")
			{
				docs.GET("", docCtrl.List)
				docs.GET("/:id", docCtrl.Get)
				docs.GET("/:id/download", docCtrl.Download)
				docs.POST("/upload", docCtrl.Upload)
				docs.PUT("/:id", docCtrl.Update)
				docs.DELETE("/:id", docCtrl.Delete)
				docs.POST("/:id/archive", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager), docCtrl.Archive)
			}

			// 合同管理（组长和组员可操作）
			contracts := auth.Group("/contracts")
			{
				contracts.GET("", contractCtrl.List)
				contracts.GET("/:id", contractCtrl.Get)
				contracts.POST("", middleware.RoleMiddleware(config.RoleTeamLeader, config.RoleTeamMember), contractCtrl.Create)
				contracts.PUT("/:id", middleware.RoleMiddleware(config.RoleTeamLeader, config.RoleTeamMember), contractCtrl.Update)
				contracts.POST("/:id/upload", middleware.RoleMiddleware(config.RoleTeamLeader, config.RoleTeamMember), contractCtrl.UploadFile)
				contracts.DELETE("/:id", middleware.RoleMiddleware(config.RoleTeamLeader, config.RoleTeamMember), contractCtrl.Delete)
			}

			// 知识库管理（所有用户可查看和下载）
			kb := auth.Group("/knowledge")
			{
				kb.GET("", kbCtrl.List)
				kb.GET("/hot", kbCtrl.GetHotItems)
				kb.GET("/categories", kbCtrl.GetCategories)
				kb.GET("/:id", kbCtrl.Get)
				kb.GET("/:id/download", kbCtrl.Download)
				kb.GET("/:id/versions", kbCtrl.GetVersions)

				// 上传和编辑（部门经理、组长、组员）
				kb.POST("/upload", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager, config.RoleTeamLeader, config.RoleTeamMember), kbCtrl.Upload)
				kb.PUT("/:id", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager), kbCtrl.Update)
				kb.POST("/:id/version", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager, config.RoleTeamLeader, config.RoleTeamMember), kbCtrl.NewVersion)
				// 删除（权限在控制器中检查：管理员/部门经理可删除所有，其他用户只能删除自己的）
				kb.DELETE("/:id", kbCtrl.Delete)

				// 分类管理（管理员、部门经理）
				kb.POST("/categories", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager), kbCtrl.CreateCategory)
				kb.PUT("/categories/:id", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager), kbCtrl.UpdateCategory)
				kb.DELETE("/categories/:id", middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager), kbCtrl.DeleteCategory)
			}

			// 操作日志（管理员和部门经理可查看）
			logs := auth.Group("/logs")
			logs.Use(middleware.RoleMiddleware(config.RoleAdmin, config.RoleDeptManager))
			{
				logs.GET("", logCtrl.List)
				logs.GET("/actions", logCtrl.GetActions)
				logs.GET("/modules", logCtrl.GetModules)
				logs.GET("/statistics", logCtrl.GetStatistics)
			}

			// 费用管理（所有用户可查看自己的费用，管理员可查看所有）
			expenses := auth.Group("/expenses")
			{
				expenses.GET("", expenseCtrl.List)
				expenses.GET("/statistics", expenseCtrl.GetStatistics)
				expenses.GET("/comparison", expenseCtrl.GetProjectComparison)
				expenses.GET("/:id", expenseCtrl.Get)
				expenses.POST("", expenseCtrl.Create)
				expenses.PUT("/:id", expenseCtrl.Update)
				expenses.DELETE("/:id", expenseCtrl.Delete)
				expenses.POST("/:id/voucher", expenseCtrl.UploadVoucher)
				expenses.GET("/:id/voucher", expenseCtrl.DownloadVoucher)
				expenses.DELETE("/:id/voucher", expenseCtrl.DeleteVoucher)
			}
		}
	}
}
