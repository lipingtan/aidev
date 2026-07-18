package model

import (
	"time"

	"gorm.io/gorm"
)

// Application 应用模型
type Application struct {
	ID          int64          `gorm:"primaryKey" json:"id,string"`
	AppCode     string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"app_code"` // 应用编码，全局唯一
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`                // 应用名称
	Description string         `gorm:"type:varchar(512)" json:"description"`                  // 应用描述
	Status      int            `gorm:"default:1" json:"status"`                               // 状态：1-启用 0-禁用
	Version     int            `gorm:"default:1" json:"version"`                              // 乐观锁版本号
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
