package models

import (
	"time"

	authModel "go-admin/common/auth/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AdminApprovalFlow 审批流程定义
type AdminApprovalFlow struct {
	ID          int64          `gorm:"primaryKey" json:"id,string"`
	TenantID    int64          `gorm:"not null;default:0;index:idx_tenant_code,priority:1" json:"tenant_id,string"` // 租户ID，0=全局默认
	FlowCode    string         `gorm:"type:varchar(64);not null;index:idx_tenant_code,priority:2" json:"flow_code"` // 流程标识，唯一性由业务层校验
	FlowName    string         `gorm:"type:varchar(128);not null" json:"flow_name"`                                 // 流程名称
	FlowConfig  datatypes.JSON `gorm:"type:json;not null" json:"flow_config"`                                       // 节点配置列表
	Description string         `gorm:"type:varchar(255)" json:"description"`                                        // 描述
	CreatedAt   *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	CreateBy    int            `gorm:"not null;default:0" json:"create_by"`
	UpdateBy    int            `gorm:"not null;default:0" json:"update_by"`
}

// TableName 指定表名
func (AdminApprovalFlow) TableName() string {
	return "admin_approval_flow"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (m *AdminApprovalFlow) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = authModel.NextID()
	}
	return nil
}
