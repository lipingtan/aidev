package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OperationLog 操作日志模型
type OperationLog struct {
	ID         int64          `gorm:"primaryKey" json:"id,string"`
	UserID     int64          `gorm:"index:idx_log_user;not null" json:"user_id,string"`        // 操作人 ID
	TenantID   int64          `gorm:"index:idx_log_tenant;not null" json:"tenant_id,string"`    // 租户 ID
	Module     string         `gorm:"type:varchar(64);index:idx_log_module" json:"module"` // 模块名
	Action     string         `gorm:"type:varchar(64)" json:"action"`                    // 操作动作
	TargetType string         `gorm:"type:varchar(64)" json:"target_type"`               // 操作对象类型
	TargetID   string         `gorm:"type:varchar(64)" json:"target_id"`                 // 操作对象 ID
	Summary    string         `gorm:"type:varchar(512)" json:"summary"`                  // 操作摘要
	OldValue   datatypes.JSON `gorm:"type:json" json:"old_value"`                        // 变更前值
	NewValue   datatypes.JSON `gorm:"type:json" json:"new_value"`                        // 变更后值
	ClientIP   string         `gorm:"type:varchar(64)" json:"client_ip"`                 // 客户端 IP
	UserAgent  string         `gorm:"type:varchar(512)" json:"user_agent"`               // 客户端 UA
	CreatedAt  *time.Time     `gorm:"autoCreateTime;index:idx_log_time" json:"created_at"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "admin_operation_log"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (ol *OperationLog) BeforeCreate(tx *gorm.DB) error {
	if ol.ID == 0 {
		ol.ID = NextID()
	}
	return nil
}
