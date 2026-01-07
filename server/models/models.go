package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"unique;not null;size:50" json:"username"`
	Password     string         `gorm:"not null;size:255" json:"-"`
	Name         string         `gorm:"size:50" json:"name"`
	Email        string         `gorm:"size:100" json:"email"`
	Phone        string         `gorm:"size:20" json:"phone"`
	Department   string         `gorm:"size:100" json:"department"`
	RoleID       uint           `json:"role_id"`
	Role         *Role          `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Status       int            `gorm:"default:1" json:"status"` // 1:启用 0:禁用
	FailedLogins int            `gorm:"default:0" json:"-"`
	LockedUntil  *time.Time     `json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Role 角色模型
type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"unique;not null;size:50" json:"name"`
	Code        string         `gorm:"unique;not null;size:50" json:"code"`
	Description string         `gorm:"size:255" json:"description"`
	Permissions string         `gorm:"type:text" json:"permissions"` // JSON格式存储权限列表
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Project 项目模型
type Project struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ProjectNo       string         `gorm:"unique;not null;size:50" json:"project_no"` // 项目编号
	Name            string         `gorm:"not null;size:200" json:"name"`             // 项目名称
	ProjectType     string         `gorm:"size:50;not null" json:"project_type"`      // 项目类型：成本性/资本性
	ManagerID       uint           `gorm:"not null" json:"manager_id"`                // 项目负责人ID
	Manager         *User          `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	SubManagerID    *uint          `json:"sub_manager_id"` // 子负责人ID（可选，通过项目成员维护）
	SubManager      *User          `gorm:"foreignKey:SubManagerID" json:"sub_manager,omitempty"`
	ContractNo      string         `gorm:"size:100" json:"contract_no"`                       // 合同编号（非必填）
	BudgetCode      string         `gorm:"size:100" json:"budget_code"`                       // 预算编码（非必填）
	InnovationCode  string         `gorm:"size:100" json:"innovation_code"`                   // 创新项目编码（非必填）
	InitiationDate  *time.Time     `gorm:"not null" json:"initiation_date"`                   // 立项日期
	ClosingDate     *time.Time     `gorm:"not null" json:"closing_date"`                      // 结项日期
	LaborCost       float64        `gorm:"not null;default:0" json:"labor_cost"`              // 人工费用
	DirectCost      float64        `gorm:"not null;default:0" json:"direct_cost"`             // 直接投入费用
	OutsourcingCost float64        `gorm:"not null;default:0" json:"outsourcing_cost"`        // 委托研发费用
	OtherCost       float64        `gorm:"not null;default:0" json:"other_cost"`              // 其他费用
	CurrentPhase    string         `gorm:"size:50;default:'initiation'" json:"current_phase"` // 当前阶段
	Status          string         `gorm:"size:50;default:'not_started'" json:"status"`       // 项目状态
	CreatedBy       uint           `json:"created_by"`
	Creator         *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Phases          []ProjectPhase `gorm:"foreignKey:ProjectID" json:"phases,omitempty"`
	Tasks           []Task         `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
}

// ProjectPhase 项目阶段模型
type ProjectPhase struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ProjectID   uint       `json:"project_id"`
	PhaseName   string     `gorm:"size:50;not null" json:"phase_name"` // 阶段名称
	PhaseOrder  int        `json:"phase_order"`                        // 阶段顺序
	IsFixed     bool       `gorm:"default:false" json:"is_fixed"`      // 是否固定阶段（固定阶段不可删除）
	Status      string     `gorm:"size:50;default:'not_started'" json:"status"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CompletedAt *time.Time `json:"completed_at"`
	Remark      string     `gorm:"type:text" json:"remark"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Task 任务模型
type Task struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ProjectID     uint           `json:"project_id"`
	Project       *Project       `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	PhaseID       uint           `json:"phase_id"` // 所属阶段ID
	Phase         *ProjectPhase  `gorm:"foreignKey:PhaseID" json:"phase,omitempty"`
	TaskName      string         `gorm:"size:200;not null" json:"task_name"` // 任务名称
	Description   string         `gorm:"type:text" json:"description"`       // 任务描述
	TaskType      string         `gorm:"size:50" json:"task_type"`           // 任务类型
	AssigneeID    uint           `json:"assignee_id"`                        // 责任人ID
	Assignee      *User          `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	AssigneeType  string         `gorm:"size:50" json:"assignee_type"` // 责任主体类型
	Deadline      *time.Time     `json:"deadline"`                     // 截止日期
	Status        string         `gorm:"size:50;default:'not_started'" json:"status"`
	Priority      int            `gorm:"default:2" json:"priority"`       // 优先级 1高 2中 3低
	Deliverables  string         `gorm:"type:text" json:"deliverables"`   // 交付件要求
	ReviewStatus  string         `gorm:"size:50" json:"review_status"`    // 审核状态
	ReviewComment string         `gorm:"type:text" json:"review_comment"` // 审核意见
	ReviewedBy    uint           `json:"reviewed_by"`
	ReviewedAt    *time.Time     `json:"reviewed_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	CreatedBy     uint           `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Documents     []Document     `gorm:"foreignKey:TaskID" json:"documents,omitempty"`
}

// Document 资料/文档模型
type Document struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	ProjectID  uint           `json:"project_id"`
	PhaseID    uint           `json:"phase_id"` // 所属阶段ID
	Phase      *ProjectPhase  `gorm:"foreignKey:PhaseID" json:"phase,omitempty"`
	TaskID     uint           `json:"task_id"`
	DocName    string         `gorm:"size:255;not null" json:"doc_name"`       // 资料名称
	DocType    string         `gorm:"size:50" json:"doc_type"`                 // 资料类型
	FilePath   string         `gorm:"size:500" json:"file_path"`               // 文件路径
	FileSize   int64          `json:"file_size"`                               // 文件大小
	MimeType   string         `gorm:"size:100" json:"mime_type"`               // MIME类型
	Version    string         `gorm:"size:20;default:'1.0'" json:"version"`    // 版本号
	Status     string         `gorm:"size:50;default:'pending'" json:"status"` // 状态:pending/approved/archived
	UploadedBy uint           `json:"uploaded_by"`
	Uploader   *User          `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	Remark     string         `gorm:"type:text" json:"remark"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// Contract 合同模型
type Contract struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	ProjectID     uint       `json:"project_id"`
	ContractNo    string     `gorm:"unique;size:50" json:"contract_no"`     // 合同编号
	ContractName  string     `gorm:"size:200" json:"contract_name"`         // 合同名称
	PartyA        string     `gorm:"size:200" json:"party_a"`               // 甲方
	PartyB        string     `gorm:"size:200" json:"party_b"`               // 乙方
	Amount        float64    `json:"amount"`                                // 合同金额
	SignDate      *time.Time `json:"sign_date"`                             // 签订日期
	StartDate     *time.Time `json:"start_date"`                            // 有效期开始
	EndDate       *time.Time `json:"end_date"`                              // 有效期结束
	PaymentMethod string     `gorm:"size:100" json:"payment_method"`        // 付款方式
	Status        string     `gorm:"size:50;default:'draft'" json:"status"` // 状态
	FilePath      string     `gorm:"size:500" json:"file_path"`             // 合同文件路径
	CreatedBy     uint       `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// KnowledgeBase 知识库资料模型
type KnowledgeBase struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Title         string         `gorm:"size:255;not null" json:"title"` // 资料标题
	CategoryID    uint           `json:"category_id"`                    // 分类ID
	Category      *KBCategory    `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Keywords      string         `gorm:"size:500" json:"keywords"`     // 关键词
	Description   string         `gorm:"type:text" json:"description"` // 资料描述
	FilePath      string         `gorm:"size:500" json:"file_path"`    // 文件路径
	FileSize      int64          `json:"file_size"`
	MimeType      string         `gorm:"size:100" json:"mime_type"`
	Version       string         `gorm:"size:20;default:'1.0'" json:"version"`
	Status        string         `gorm:"size:50;default:'published'" json:"status"` // published/draft
	ViewCount     int            `gorm:"default:0" json:"view_count"`               // 查看次数
	DownloadCount int            `gorm:"default:0" json:"download_count"`           // 下载次数
	UploadedBy    uint           `json:"uploaded_by"`
	Uploader      *User          `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// KBCategory 知识库分类模型
type KBCategory struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	ParentID    uint           `json:"parent_id"`
	Description string         `gorm:"size:255" json:"description"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// KBVersion 知识库版本记录
type KBVersion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	KnowledgeID uint      `json:"knowledge_id"`
	Version     string    `gorm:"size:20" json:"version"`
	FilePath    string    `gorm:"size:500" json:"file_path"`
	ChangeNote  string    `gorm:"type:text" json:"change_note"`
	UploadedBy  uint      `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// OperationLog 操作日志模型
type OperationLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"user_id"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action      string    `gorm:"size:100;not null" json:"action"` // 操作类型
	Module      string    `gorm:"size:50" json:"module"`           // 模块名称
	TargetType  string    `gorm:"size:50" json:"target_type"`      // 目标类型
	TargetID    uint      `json:"target_id"`                       // 目标ID
	TargetName  string    `gorm:"size:255" json:"target_name"`     // 目标名称
	Description string    `gorm:"type:text" json:"description"`    // 操作描述
	Result      string    `gorm:"size:50" json:"result"`           // 结果:success/failed
	IPAddress   string    `gorm:"size:50" json:"ip_address"`
	UserAgent   string    `gorm:"size:500" json:"user_agent"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProjectMember 项目成员模型
type ProjectMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `json:"project_id"`
	UserID    uint      `json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RoleType  string    `gorm:"size:50" json:"role_type"` // 在项目中的角色
	JoinDate  time.Time `json:"join_date"`
	CreatedAt time.Time `json:"created_at"`
}

// Expense 费用记录模型（基于财务模板）
type Expense struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// 关联项目信息
	ProjectID   *uint    `gorm:"default:null" json:"project_id"` // 系统内项目ID（可为空，未归类时为null）
	Project     *Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL" json:"project,omitempty"`
	ProjectCode string   `gorm:"size:100" json:"project_code"` // 关联项目编码（来自Excel）

	// 单位信息
	UnitCode string `gorm:"size:100" json:"unit_code"` // 单位编号
	UnitName string `gorm:"size:200" json:"unit_name"` // 单位名称

	// 单据信息
	DocumentNo        string `gorm:"size:100" json:"document_no"`         // 单据编号
	BusinessScene     string `gorm:"size:100" json:"business_scene"`      // 业务场景
	CrossIndustryCode string `gorm:"size:100" json:"cross_industry_code"` // 跨行业业务编码
	IsProjectExpense  string `gorm:"size:10" json:"is_project_expense"`   // 是否属于项目开支
	Summary           string `gorm:"type:text" json:"summary"`            // 摘要

	// 报账人信息
	DepartmentName       string `gorm:"size:100" json:"department_name"` // 部门名称
	ReimbursedBy         uint   `json:"reimbursed_by"`                   // 报账人ID（系统用户）
	ReimbursedUser       *User  `gorm:"foreignKey:ReimbursedBy" json:"reimbursed_user,omitempty"`
	ReimbursedPersonName string `gorm:"size:100" json:"reimbursed_person_name"` // 报账人姓名（来自Excel）

	// 单据状态
	DocumentStatus string `gorm:"size:50" json:"document_status"` // 单据状态
	FrozenStatus   string `gorm:"size:50" json:"frozen_status"`   // 冻结状态

	// 金额信息
	ReimbursementAmount  float64 `gorm:"default:0" json:"reimbursement_amount"`    // 报账金额
	PaymentAmount        float64 `gorm:"default:0" json:"payment_amount"`          // 支付金额
	WriteOffAmount       float64 `gorm:"default:0" json:"write_off_amount"`        // 核销金额
	InvoiceAmountExclTax float64 `gorm:"default:0" json:"invoice_amount_excl_tax"` // 发票不含税金额
	InvoiceAmountInclTax float64 `gorm:"default:0" json:"invoice_amount_incl_tax"` // 发票含税金额

	// 流程信息
	CurrentProcess   string `gorm:"size:100" json:"current_process"`   // 当前处理环节
	CurrentProcessor string `gorm:"size:100" json:"current_processor"` // 当前处理人
	SharedProcess    string `gorm:"size:100" json:"shared_process"`    // 共享处理环节
	SharedProcessor  string `gorm:"size:100" json:"shared_processor"`  // 共享处理人

	// 实物信息
	PhysicalStatus   string `gorm:"size:50" json:"physical_status"`    // 实物状态
	PhysicalLocation string `gorm:"size:200" json:"physical_location"` // 实物位置

	// 单据类型
	DocumentType     string `gorm:"size:50" json:"document_type"`       // 单据类型
	DocumentTypeName string `gorm:"size:100" json:"document_type_name"` // 单据类型名称

	// 供应商信息
	SupplierCode string `gorm:"size:100" json:"supplier_code"` // 供应商编号
	SupplierName string `gorm:"size:200" json:"supplier_name"` // 供应商名称

	// 时间信息
	CreateDocTime *time.Time `json:"create_doc_time"` // 制单时间
	SubmitTime    *time.Time `json:"submit_time"`     // 提交时间

	// 其他信息
	InternalCode   string `gorm:"size:100" json:"internal_code"`   // 内码
	PaymentAccount string `gorm:"size:100" json:"payment_account"` // 付款账号

	// 系统字段
	ExpenseType  string         `gorm:"size:20" json:"expense_type"`        // 费用类型: labor(人工), direct(直接投入), outsourcing(委托研发), other(其他)
	VoucherPath  string         `gorm:"size:1000" json:"voucher_path"`      // 凭证文件路径（多个文件用逗号分隔）
	Remark       string         `gorm:"type:text" json:"remark"`            // 备注
	IsClassified bool           `gorm:"default:false" json:"is_classified"` // 是否已归类到项目
	CreatedBy    uint           `json:"created_by"`
	Creator      *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
