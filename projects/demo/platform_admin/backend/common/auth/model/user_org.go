package model

import (
	"time"

	"gorm.io/gorm"
)

// UserOrg 用户-组织关联模型
type UserOrg struct {
	ID        int64      `gorm:"primaryKey" json:"id,string"`
	UserID    int64      `gorm:"not null;uniqueIndex:uk_user_org" json:"user_id,string"`
	OrgUnitID int64      `gorm:"not null;uniqueIndex:uk_user_org;index:idx_org_unit" json:"org_unit_id,string"`
	TenantID  int64      `gorm:"not null;uniqueIndex:uk_user_org;index:idx_org_unit" json:"tenant_id,string"`
	IsPrimary int        `gorm:"not null;default:0" json:"is_primary"` // 1=主归属
	CreatedAt *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (UserOrg) TableName() string {
	return "admin_user_org"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (uo *UserOrg) BeforeCreate(tx *gorm.DB) error {
	if uo.ID == 0 {
		uo.ID = NextID()
	}
	return nil
}
