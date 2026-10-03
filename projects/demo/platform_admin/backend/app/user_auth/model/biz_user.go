package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	authModel "go-admin/common/auth/model"
)

// BizUser C端用户
// Phone 和 Email 都是可选字段，至少填一个
// 唯一约束：(tenant_id, phone) 和 (tenant_id, email)，NULL 值不参与唯一约束
type BizUser struct {
	ID           int64      `json:"id,string" gorm:"primaryKey;autoIncrement:false"`
	TenantID     int64      `json:"tenant_id,string" gorm:"not null;default:0;uniqueIndex:uk_tenant_phone;uniqueIndex:uk_tenant_email;index:idx_biz_user_tenant"`
	Phone        *string    `json:"phone" gorm:"size:20;uniqueIndex:uk_tenant_phone;index:idx_biz_user_phone"`
	Email        *string    `json:"email" gorm:"size:255;uniqueIndex:uk_tenant_email"`
	Password     string     `json:"-" gorm:"size:128;default:''"`
	Nickname     string     `json:"nickname" gorm:"size:64;default:''"`
	Avatar       string     `json:"avatar" gorm:"size:256;default:''"`
	Status       int        `json:"status" gorm:"not null;default:1"`
	// FR-21：禁用/解禁历史与原因在 biz_user_status_logs 表（最新一条 = 当前态），biz_user 不加列
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
	// Token 过期时间（分钟），NULL 时使用全局配置
	AccessTokenExpireMinutes *int `json:"access_token_expire_minutes"`
	RefreshTokenExpireMinutes *int `json:"refresh_token_expire_minutes"`
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

// BizUserStatusLog FR-21 C端账号启用/禁用操作历史
type BizUserStatusLog struct {
	ID          int64      `json:"id,string" gorm:"primaryKey;autoIncrement:false"`
	TenantID    int64      `json:"tenant_id,string" gorm:"not null;default:0"`
	UserID      int64      `json:"user_id,string" gorm:"not null;default:0;index:idx_busl_user,priority:1"`
	Action      int        `json:"action" gorm:"not null"` // 0=禁用 1=启用
	AdminReason string     `json:"admin_reason" gorm:"size:500;default:''"` // 管理端内部原因（禁用时）
	UserReason  string     `json:"user_reason" gorm:"size:500;default:''"`  // 用户展示原因（禁用时）
	OperatorID  int64      `json:"operator_id,string" gorm:"not null;default:0"`
	OperatorName string    `json:"operator_name" gorm:"size:64;default:''"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;index:idx_busl_user"`
}

func (BizUserStatusLog) TableName() string { return "biz_user_status_logs" }

// BeforeCreate 雪花 ID
func (l *BizUserStatusLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == 0 {
		l.ID = authModel.NextID()
	}
	return nil
}
