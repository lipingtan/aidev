package models

import (
	"time"

	authModel "go-admin/common/auth/model"

	"gorm.io/gorm"
)

// AdminApprovalVote 审批投票记录（会签/或签并发安全）
type AdminApprovalVote struct {
	ID       int64      `gorm:"primaryKey" json:"id,string"`
	TenantID int64      `gorm:"not null;index:idx_tenant" json:"tenant_id,string"`                              // 租户ID（从节点取值）
	NodeID   int64      `gorm:"not null;uniqueIndex:idx_node_user,priority:1;index:idx_node" json:"node_id,string"` // 关联节点ID
	UserID   int64      `gorm:"not null;uniqueIndex:idx_node_user,priority:2" json:"user_id,string"`            // 投票用户ID
	Action   string     `gorm:"type:varchar(16);not null" json:"action"`                                        // APPROVE/REJECT
	Comment  string     `gorm:"type:varchar(255)" json:"comment"`                                               // 备注
	VotedAt  *time.Time `gorm:"" json:"voted_at"`                                                               // 投票时间（BeforeCreate 自动设置）
}

// TableName 指定表名
func (AdminApprovalVote) TableName() string {
	return "admin_approval_vote"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID 并设置投票时间
func (m *AdminApprovalVote) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = authModel.NextID()
	}
	if m.VotedAt == nil {
		now := time.Now()
		m.VotedAt = &now
	}
	return nil
}
