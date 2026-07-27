package models

import (
	"time"

	authModel "go-admin/common/auth/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AdminApproval 审批实例
type AdminApproval struct {
	ID           int64          `gorm:"primaryKey" json:"id,string"`
	TenantID     int64          `gorm:"not null;index:idx_tenant_status,priority:1;index:idx_tenant_applicant,priority:1" json:"tenant_id,string"` // 租户ID
	FlowID       int64          `gorm:"not null" json:"flow_id,string"`                                                                             // 关联的流程定义ID
	FlowCode     string         `gorm:"type:varchar(64);not null" json:"flow_code"`                                                                 // 流程标识快照
	FlowSnapshot datatypes.JSON `gorm:"type:json;not null" json:"flow_snapshot"`                                                                    // 发起时的flow_config快照
	BizType      string         `gorm:"type:varchar(64);not null;index:idx_biz,priority:1" json:"biz_type"`                                         // 业务类型
	BizID        string         `gorm:"type:varchar(64);not null;index:idx_biz,priority:2" json:"biz_id"`                                           // 业务对象ID
	Status       string         `gorm:"type:varchar(32);not null;default:PENDING;index:idx_tenant_status,priority:2" json:"status"`                 // PENDING/APPROVED/REJECTED/CANCELLED
	ApplicantID  int64          `gorm:"not null;index:idx_tenant_applicant,priority:2" json:"applicant_id,string"`                                  // 发起人用户ID
	CancelBy     *int64         `gorm:"default:null" json:"cancel_by,string"`                                                                       // 撤销操作人ID
	CancelReason string         `gorm:"type:varchar(255)" json:"cancel_reason"`                                                                     // 撤销原因
	CreatedAt    *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	CreateBy     int            `gorm:"not null;default:0" json:"create_by"`
	UpdateBy     int            `gorm:"not null;default:0" json:"update_by"`
}

// TableName 指定表名
func (AdminApproval) TableName() string {
	return "admin_approval"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (m *AdminApproval) BeforeCreate(_ *gorm.DB) error {
	if m.ID == 0 {
		m.ID = authModel.NextID()
	}
	return nil
}
