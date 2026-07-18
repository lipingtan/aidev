// Package model 定义 GORM 模型（admin_* 表）
package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 公共基础字段（含乐观锁版本号和软删除）
type BaseModel struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	Version   int            `gorm:"default:1" json:"version"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// BaseModelNoVersion 无版本号的公共基础字段
type BaseModelNoVersion struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}
