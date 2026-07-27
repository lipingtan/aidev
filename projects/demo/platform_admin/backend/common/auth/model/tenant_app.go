package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TenantApp 租户-应用订阅关联（M:N）
type TenantApp struct {
	ID                 int64          `gorm:"primaryKey" json:"id"`
	TenantID           int64          `gorm:"uniqueIndex:idx_tenant_app;not null" json:"tenant_id"`                          // 租户 ID
	AppCode            string         `gorm:"type:varchar(64);uniqueIndex:idx_tenant_app;not null" json:"app_code"`          // 应用编码
	EnabledModules     datatypes.JSON `gorm:"type:json" json:"enabled_modules"`                                              // 启用的模块列表，NULL=全部启用
	SubscriptionMode   string         `gorm:"type:varchar(32);not null;default:direct" json:"subscription_mode"`             // direct=直接生效 approval_required=需审批（静态配置）
	SubscriptionStatus string         `gorm:"type:varchar(32);not null;default:active" json:"subscription_status"`           // active=已激活 pending_approval=审批中 rejected=已驳回（运行态）
	CreatedAt          *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (TenantApp) TableName() string {
	return "admin_tenant_app"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ta *TenantApp) BeforeCreate(tx *gorm.DB) error {
	if ta.ID == 0 {
		ta.ID = NextID()
	}
	return nil
}
