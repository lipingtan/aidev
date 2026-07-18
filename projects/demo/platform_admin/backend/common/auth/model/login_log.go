package model

import (
	"time"

	"gorm.io/gorm"
)

// LoginLog 登录日志模型
type LoginLog struct {
	ID        int64      `gorm:"primaryKey" json:"id,string"`
	UserID    int64      `gorm:"index" json:"user_id,string"`
	Username  string     `gorm:"type:varchar(128)" json:"username"`
	IP        string     `gorm:"type:varchar(64)" json:"ip"`
	Location  string     `gorm:"type:varchar(256)" json:"location"`
	Browser   string     `gorm:"type:varchar(256)" json:"browser"`
	OS        string     `gorm:"type:varchar(128)" json:"os"`
	Status    int        `gorm:"default:1" json:"status"` // 1=成功 0=失败
	Message   string     `gorm:"type:varchar(512)" json:"message"`
	LoginTime *time.Time `gorm:"autoCreateTime" json:"login_time"`
}

// TableName 指定表名
func (LoginLog) TableName() string { return "admin_login_log" }

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (l *LoginLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == 0 {
		l.ID = NextID()
	}
	return nil
}
