package model

import (
	"time"

	"gorm.io/gorm"
)

// TenantDomain 域名-租户映射
// 唯一性由 Service 层业务校验（CREATE/UPDATE 前先查活跃记录），不依赖 DB 唯一索引
// 原因：MySQL 中 NULL 不参与唯一约束比较，联合 (domain, deleted_at) 索引无法真正防重
type TenantDomain struct {
	ID        int64          `json:"id,string" gorm:"primaryKey;autoIncrement:false"`
	Domain    string         `json:"domain" gorm:"size:255;not null;index:idx_domain"` // 普通索引，唯一性由 Service 层保证
	MatchType string         `json:"match_type" gorm:"size:16;not null;default:EXACT"` // EXACT | WILDCARD（预留）
	TenantID  int64          `json:"tenant_id,string" gorm:"not null;index:idx_tenant_id"`
	Remark    string         `json:"remark" gorm:"size:256;default:''"`
	Version   int            `json:"version" gorm:"not null;default:1"` // 乐观锁
	CreatedAt *time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt *time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (TenantDomain) TableName() string { return "tenant_domain" }

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (td *TenantDomain) BeforeCreate(tx *gorm.DB) error {
	if td.ID == 0 {
		td.ID = NextID()
	}
	return nil
}
