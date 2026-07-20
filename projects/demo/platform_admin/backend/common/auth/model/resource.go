package model

import (
	"time"

	"gorm.io/gorm"
)

// Resource 菜单/按钮资源树模型
type Resource struct {
	ID             int64          `gorm:"primaryKey" json:"id,string"`
	ParentID       *int64         `gorm:"index" json:"parent_id,string"`                             // 父节点 ID
	Type           string         `gorm:"type:varchar(16);not null" json:"type"`                     // 类型：menu/button/page
	Name           string         `gorm:"type:varchar(128);not null" json:"name"`                    // 资源名称
	PermissionCode string         `gorm:"type:varchar(128)" json:"permission_code"`                  // 权限标识码
	Path           string         `gorm:"type:varchar(256)" json:"path"`                             // 路由路径
	Component      string         `gorm:"type:varchar(256)" json:"component"`                        // 前端组件路径
	Icon           string         `gorm:"type:varchar(128)" json:"icon"`                             // 图标
	AppCode        string         `gorm:"type:varchar(64);not null;index:idx_app_platform" json:"app_code"` // 所属应用编码
	Platform       string         `gorm:"type:varchar(16);not null;default:admin;index:idx_app_platform" json:"platform"` // 归属前端平台：admin/user
	ModuleCode     string         `gorm:"type:varchar(64)" json:"module_code"`                       // 所属功能模块
	SortOrder      int            `gorm:"default:0" json:"sort_order"`                               // 排序号
	Status         int            `gorm:"default:1" json:"status"`                                   // 状态：1-启用 0-禁用
	Version        int            `gorm:"default:1" json:"version"`                                  // 乐观锁版本号
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt      *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Resource) TableName() string {
	return "admin_resource"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	if r.ID == 0 {
		r.ID = NextID()
	}
	return nil
}
