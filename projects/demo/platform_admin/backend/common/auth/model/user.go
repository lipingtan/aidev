package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User 全局用户模型
type User struct {
	ID        int64          `gorm:"primaryKey" json:"id,string"`
	Username  string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username" fieldperm:"用户名"` // 用户名，全局唯一
	Password  string         `gorm:"type:varchar(256);not null" json:"-"`                                   // 密码哈希
	Email     string         `gorm:"type:varchar(128)" json:"email" fieldperm:"邮箱"`                         // 邮箱
	Phone     string         `gorm:"type:varchar(32)" json:"phone" fieldperm:"手机号"`                         // 手机号
	Nickname  string         `gorm:"type:varchar(64)" json:"nickname" fieldperm:"昵称"`                       // 昵称
	Avatar    string         `gorm:"type:varchar(512)" json:"avatar" fieldperm:"头像"`                        // 头像 URL
	Status    int            `gorm:"default:1" json:"status" fieldperm:"状态"`                                // 状态：1-启用 0-禁用
	ExtFields datatypes.JSON `gorm:"type:json" json:"ext_fields"`                                           // 扩展字段（JSON）
	Version   int            `gorm:"default:1" json:"version"`                                              // 乐观锁版本号
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt *time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "admin_user"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == 0 {
		u.ID = NextID()
	}
	return nil
}
