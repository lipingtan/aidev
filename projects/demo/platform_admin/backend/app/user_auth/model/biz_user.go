package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	authModel "go-admin/common/auth/model"
)

// BizUser C端用户
type BizUser struct {
	ID           int64      `json:"id,string" gorm:"primaryKey;autoIncrement:false"`
	TenantID     int64      `json:"tenant_id,string" gorm:"not null;default:0;uniqueIndex:uk_tenant_phone;index:idx_biz_user_tenant"`
	Phone        string     `json:"phone" gorm:"size:20;not null;uniqueIndex:uk_tenant_phone;index:idx_biz_user_phone"`
	Password     string     `json:"-" gorm:"size:128;default:''"`
	Nickname     string     `json:"nickname" gorm:"size:64;default:''"`
	Avatar       string     `json:"avatar" gorm:"size:256;default:''"`
	Status       int        `json:"status" gorm:"not null;default:1"`
	TokenVersion int        `json:"token_version" gorm:"not null;default:1"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	LastLoginIP  string     `json:"last_login_ip" gorm:"size:45;default:''"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"deleted_at" gorm:"index:idx_biz_user_deleted"`
	CreateBy     int64      `json:"create_by,string" gorm:"default:0"`
	UpdateBy     int64      `json:"update_by,string" gorm:"default:0"`
	Version      int        `json:"version" gorm:"not null;default:1"`
	ExtFields    datatypes.JSON `json:"ext_fields" gorm:"type:json"` // 扩展字段（JSON）
}

// TableName 指定表名
func (BizUser) TableName() string { return "biz_user" }

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (u *BizUser) BeforeCreate(tx *gorm.DB) error {
	if u.ID == 0 {
		u.ID = authModel.NextID()
	}
	return nil
}
