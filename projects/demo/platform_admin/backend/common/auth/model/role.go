package model

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID        int64          `gorm:"primaryKey" json:"id,string"`
	TenantID  int64          `gorm:"uniqueIndex:idx_tenant_role_code;not null" json:"tenant_id,string"` // 所属租户
	RoleCode  string         `gorm:"type:varchar(64);uniqueIndex:idx_tenant_role_code;not null" json:"role_code"` // 角色编码
	RoleName  string         `gorm:"type:varchar(128);not null" json:"role_name"`                // 角色名称
	RoleType  string         `gorm:"type:varchar(32)" json:"role_type"`                          // 角色类型：platform/tenant/custom
	ParentID  *int64         `gorm:"index" json:"parent_id,string"`                              // 父角色 ID（层级结构）
	SortOrder int            `gorm:"default:0" json:"sort_order"`                                // 排序号
	Status    int            `gorm:"default:1" json:"status"`                                    // 状态：1-启用 0-禁用
	Version   int            `gorm:"default:1" json:"version"`                                   // 乐观锁版本号
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "admin_role"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == 0 {
		r.ID = NextID()
	}
	return nil
}
