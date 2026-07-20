package model

import (
	"time"

	"gorm.io/gorm"
)

// ApiPermission 接口权限树模型
type ApiPermission struct {
	ID             int64      `gorm:"primaryKey" json:"id,string"`
	ParentID       *int64     `gorm:"index" json:"parent_id,string"`                         // 父节点 ID
	Type           string     `gorm:"type:varchar(16);not null" json:"type"`                 // 类型：GROUP/ENDPOINT
	Name           string     `gorm:"type:varchar(128);not null" json:"name"`                // 接口名称/分组名
	DisplayName    string     `gorm:"type:varchar(128)" json:"display_name"`                 // 中文显示名称
	PermissionCode string     `gorm:"type:varchar(128)" json:"permission_code"`              // 权限标识码
	URLPattern     string     `gorm:"type:varchar(256)" json:"url_pattern"`                  // URL 匹配模式
	HTTPMethod     string     `gorm:"type:varchar(16)" json:"http_method"`                   // HTTP 方法：GET/POST/PUT/DELETE
	AppCode        string     `gorm:"type:varchar(64);not null;index" json:"app_code"`       // 所属应用编码
	ModuleCode     string     `gorm:"type:varchar(64)" json:"module_code"`                   // 所属功能模块
	Status         string     `gorm:"type:varchar(16);default:'ACTIVE'" json:"status"`       // 状态：UNASSIGNED/ACTIVE/DEPRECATED
	Visible        *int       `gorm:"default:1" json:"visible"`                              // 1=显示 0=隐藏（界面配置可见性）
	AuthRequired   *int       `gorm:"default:1" json:"auth_required"`                        // 1=需要权限校验 0=免检（所有登录用户可调）
	SortOrder      int        `gorm:"default:0" json:"sort_order"`                           // 排序号
	CreatedAt      *time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ApiPermission) TableName() string {
	return "admin_api_permission"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ap *ApiPermission) BeforeCreate(tx *gorm.DB) error {
	if ap.ID == 0 {
		ap.ID = NextID()
	}
	return nil
}
