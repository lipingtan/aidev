package model

import (
	"time"

	"gorm.io/gorm"
)

// SysConfig 系统配置模型
type SysConfig struct {
	ID          int64          `gorm:"primaryKey" json:"id,string"`
	ConfigName  string         `gorm:"type:varchar(128)" json:"config_name"`
	ConfigKey   string         `gorm:"type:varchar(128);uniqueIndex" json:"config_key"`
	ConfigValue string         `gorm:"type:text" json:"config_value"`
	ConfigType  int            `gorm:"default:0" json:"config_type"` // 0=默认 1=系统内置
	Remark      string         `gorm:"type:varchar(512)" json:"remark"`
	Status      int            `gorm:"default:1" json:"status"`
	CreatedAt   *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (SysConfig) TableName() string { return "sys_config" }

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (s *SysConfig) BeforeCreate(tx *gorm.DB) error {
	if s.ID == 0 {
		s.ID = NextID()
	}
	return nil
}
