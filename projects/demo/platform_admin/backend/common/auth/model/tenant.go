package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Tenant 租户模型
type Tenant struct {
	ID         int64          `gorm:"primaryKey" json:"id,string"`
	TenantCode string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"tenant_code" fieldperm:"租户编码"` // 租户编码，全局唯一
	Name       string         `gorm:"type:varchar(128);not null" json:"name" fieldperm:"租户名称"`                   // 租户名称
	Status     int            `gorm:"default:1" json:"status" fieldperm:"状态"`                                    // 状态：1-正常 0-禁用 2-只读 3-注销中
	Config     datatypes.JSON `gorm:"type:json" json:"config"`                                                   // 租户个性化配置
	Timezone   string         `gorm:"type:varchar(64);not null;default:Asia/Shanghai" json:"timezone" fieldperm:"时区"` // 租户时区
	Locale     string         `gorm:"type:varchar(16);not null;default:zh-CN" json:"locale" fieldperm:"语言"`           // 默认语言
	Currency   string         `gorm:"type:varchar(8);not null;default:CNY" json:"currency" fieldperm:"币种"`            // 默认币种
	ExpiredAt  *time.Time     `json:"expired_at" fieldperm:"到期时间"`                                               // 试用/订阅到期时间
	Version    int            `gorm:"default:1" json:"version"`                                                  // 乐观锁版本号
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
