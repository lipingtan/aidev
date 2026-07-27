package models

import (
	"time"

	authModel "go-admin/common/auth/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AdminApprovalNode 审批节点
type AdminApprovalNode struct {
	ID              int64          `gorm:"primaryKey" json:"id,string"`
	TenantID        int64          `gorm:"not null" json:"tenant_id,string"`                                                             // 租户ID
	ApprovalID      int64          `gorm:"not null;index:idx_approval" json:"approval_id,string"`                                        // 关联审批实例ID
	NodeOrder       int            `gorm:"not null" json:"node_order"`                                                                   // 节点顺序，从1开始
	NodeType        string         `gorm:"type:varchar(32);not null" json:"node_type"`                                                   // SINGLE/AND_SIGN/OR_SIGN
	AssigneeType    string         `gorm:"type:varchar(32);not null" json:"assignee_type"`                                               // USER/ROLE/DEPT_HEAD
	AssigneeIDs     datatypes.JSON `gorm:"column:assignee_ids;type:json;not null" json:"assignee_ids"`                                   // 审批人ID列表（原始配置）
	AssigneeUserIDs datatypes.JSON `gorm:"column:assignee_user_ids;type:json;not null" json:"assignee_user_ids"`                         // 展开后的实际用户ID列表
	Status          string         `gorm:"type:varchar(32);not null;default:WAITING;index:idx_pending_timeout,priority:1;index:idx_notified,priority:1" json:"status"` // WAITING/PENDING/APPROVED/REJECTED/SKIPPED/TIMEOUT
	ApproveComment  string         `gorm:"type:varchar(255)" json:"approve_comment"`                                                     // 审批通过时的备注
	RejectReason    string         `gorm:"type:varchar(255)" json:"reject_reason"`                                                       // 驳回原因
	AssigneeNote    string         `gorm:"type:varchar(255)" json:"assignee_note"`                                                       // 审计注释
	TimeoutHours    int            `gorm:"not null;default:0" json:"timeout_hours"`                                                      // 超时时限（小时），0=不超时
	TimeoutAction   string         `gorm:"type:varchar(32)" json:"timeout_action"`                                                       // AUTO_APPROVE/AUTO_REJECT/ESCALATE
	EscalateTo      string         `gorm:"type:varchar(64)" json:"escalate_to"`                                                          // 超时升级目标用户/角色ID
	TimeoutAt       *time.Time     `gorm:"index:idx_pending_timeout,priority:2" json:"timeout_at"`                                       // 超时触发时间
	NotifiedAt      *time.Time     `gorm:"index:idx_notified,priority:2" json:"notified_at"`                                             // 审批人通知时间，NULL=待通知
	CreatedAt       *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (AdminApprovalNode) TableName() string {
	return "admin_approval_node"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (m *AdminApprovalNode) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = authModel.NextID()
	}
	return nil
}
