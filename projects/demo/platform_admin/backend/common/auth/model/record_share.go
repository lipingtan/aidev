package model

import (
	"time"

	"gorm.io/gorm"
)

// RecordShare 记录共享规则
type RecordShare struct {
	ID          int64      `gorm:"primaryKey" json:"id,string"`
	TenantID    int64      `gorm:"not null;index:idx_record,priority:1" json:"tenant_id,string"`
	ObjectCode  string     `gorm:"type:varchar(64);not null;index:idx_record,priority:2" json:"object_code"`
	RecordID    int64      `gorm:"not null;index:idx_record,priority:3" json:"record_id,string"`
	ShareToType string     `gorm:"type:varchar(16);not null;index:idx_share_to,priority:1" json:"share_to_type"` // USER/ROLE/DEPT
	ShareToID   int64      `gorm:"not null;index:idx_share_to,priority:2" json:"share_to_id,string"`
	AccessLevel string     `gorm:"type:varchar(16);default:READ;not null" json:"access_level"` // READ/EDIT
	ExpireAt    *time.Time `json:"expire_at"`
	CreatedBy   int64      `gorm:"not null" json:"created_by,string"`
	CreatedAt   *time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (RecordShare) TableName() string {
	return "admin_record_share"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (r *RecordShare) BeforeCreate(tx *gorm.DB) error {
	if r.ID == 0 {
		r.ID = NextID()
	}
	return nil
}
