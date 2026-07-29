package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Application 应用模型（V2 统一应用模型）
type Application struct {
	ID          int64          `gorm:"primaryKey" json:"id,string"`
	AppCode     string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"app_code"`       // 应用编码，全局唯一
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`                      // 应用名称
	Description string         `gorm:"type:varchar(512)" json:"description"`                        // 应用描述
	AppType     string         `gorm:"type:varchar(16);not null;default:BUILTIN" json:"app_type"`   // 类型：BUILTIN/PLUGIN/EXTERNAL
	RoutePrefix string         `gorm:"type:varchar(128)" json:"route_prefix"`                       // API 路由前缀，如 /api/v1/admin
	Platforms   datatypes.JSON `gorm:"type:json" json:"platforms"`                                   // 支持的平台 ["admin:pc","user:h5"]
	Modules     datatypes.JSON `gorm:"type:json" json:"modules"`                                    // 功能模块声明 [{"code":"xxx","name":"xxx"}]
	AppConfig   datatypes.JSON `gorm:"type:json" json:"app_config"`                                 // 应用级配置（EXTERNAL 存 OAuth2 等）
	Icon        string         `gorm:"type:varchar(256)" json:"icon"`                               // 应用图标
	SortOrder   int            `gorm:"default:0" json:"sort_order"`                                 // 排序号
	Status      int            `gorm:"default:1" json:"status"`                                     // 状态：1-启用 0-禁用
	ExtFields   datatypes.JSON `gorm:"type:json" json:"ext_fields"`                                 // 扩展字段（JSON）
	Version     int            `gorm:"default:1" json:"version"`                                    // 乐观锁版本号
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt   *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Application) TableName() string {
	return "admin_application"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.ID == 0 {
		a.ID = NextID()
	}
	return nil
}
