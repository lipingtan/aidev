package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Tenant 租户模型
type Tenant struct {
	ID         int64          `gorm:"primaryKey" json:"id,string"`
	TenantCode string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"tenant_code"` // 租户编码，全局唯一
	Name       string         `gorm:"type:varchar(128);not null" json:"name"`                   // 租户名称
	Status     int            `gorm:"default:1" json:"status"`                                  // 状态：1-启用 0-禁用
	Config     datatypes.JSON `gorm:"type:json" json:"config"`                                  // 租户个性化配置
	Version    int            `gorm:"default:1" json:"version"`                                 // 乐观锁版本号
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt  *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "admin_tenant"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == 0 {
		t.ID = NextID()
	}
	return nil
}
