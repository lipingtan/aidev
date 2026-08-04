package model

import (
	"time"

	authmodel "go-admin/common/auth/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AbacPolicy ABAC 策略主表
type AbacPolicy struct {
	ID           int64          `gorm:"primaryKey"                                json:"id,string"`
	TenantID     int64          `gorm:"index;not null"                            json:"tenant_id,string"`
	Name         string         `gorm:"type:varchar(128);not null"                json:"name"`
	ResourceType string         `gorm:"type:varchar(64);not null"                 json:"resource_type"`
	SubjectType  string         `gorm:"type:varchar(16);not null"                 json:"subject_type"`  // ROLE|PERMISSION_SET|USER|DEPT
	SubjectID    string         `gorm:"type:varchar(64);not null"                 json:"subject_id"`
	Effect       string         `gorm:"type:varchar(8);not null"                  json:"effect"`        // ALLOW|DENY
	Priority     int            `gorm:"not null;default:100"                      json:"priority"`
	Status       int            `gorm:"not null;default:1"                        json:"status"`        // 1=启用 0=禁用
	Version      int            `gorm:"not null;default:1"                        json:"version"`       // 乐观锁
	Description  string         `gorm:"type:varchar(512)"                         json:"description"`
	CreatedAt    *time.Time     `gorm:"autoCreateTime"                            json:"created_at"`
	UpdatedAt    *time.Time     `gorm:"autoUpdateTime"                            json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                                     json:"-"`
}

// TableName 指定表名
func (AbacPolicy) TableName() string {
	return "abac_policy"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (p *AbacPolicy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == 0 {
		p.ID = authmodel.NextID()
	}
	return nil
}

// AbacRowPolicy 行权限（每条策略可有多个 action 的行权限）
type AbacRowPolicy struct {
	ID            int64          `gorm:"primaryKey"             json:"id,string"`
	PolicyID      int64          `gorm:"index;not null"         json:"policy_id,string"`
	Action        string         `gorm:"type:varchar(16);not null" json:"action"` // read|create|update|delete
	ConditionExpr datatypes.JSON `gorm:"type:json"              json:"condition_expr"`
	CreatedAt     *time.Time     `gorm:"autoCreateTime"         json:"created_at"`
}

// TableName 指定表名
func (AbacRowPolicy) TableName() string {
	return "abac_row_policy"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (p *AbacRowPolicy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == 0 {
		p.ID = authmodel.NextID()
	}
	return nil
}

// AbacColPolicy 列权限（每条策略可有多个字段配置）
type AbacColPolicy struct {
	ID          int64      `gorm:"primaryKey"                json:"id,string"`
	PolicyID    int64      `gorm:"index;not null"            json:"policy_id,string"`
	FieldName   string     `gorm:"type:varchar(64);not null" json:"field_name"`
	Effect      string     `gorm:"type:varchar(8);not null"  json:"effect"`      // SHOW|HIDE|MASK
	MaskType    string     `gorm:"type:varchar(16)"          json:"mask_type"`   // phone|email|id_card|custom
	MaskPattern string     `gorm:"type:varchar(256)"         json:"mask_pattern"`
	CreatedAt   *time.Time `gorm:"autoCreateTime"            json:"created_at"`
}

// TableName 指定表名
func (AbacColPolicy) TableName() string {
	return "abac_col_policy"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (p *AbacColPolicy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == 0 {
		p.ID = authmodel.NextID()
	}
	return nil
}
