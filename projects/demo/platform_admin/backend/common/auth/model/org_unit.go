package model

import (
	"time"

	"gorm.io/gorm"
)

// OrgUnit 组织架构节点模型
type OrgUnit struct {
	ID        int64      `gorm:"primaryKey" json:"id,string"`
	TenantID  int64      `gorm:"not null;index:idx_tenant_parent" json:"tenant_id,string"`
	ParentID  *int64     `gorm:"index:idx_tenant_parent" json:"parent_id,string"`                                  // NULL=顶级节点
	NodeType  string     `gorm:"type:varchar(32);not null;default:'DEPARTMENT'" json:"node_type"`                   // COMPANY/BRANCH/DEPARTMENT/GROUP/TEAM
	Name      string     `gorm:"type:varchar(128);not null" json:"name"`
	Code      string     `gorm:"type:varchar(64);uniqueIndex:uk_tenant_code" json:"code"`                           // 组织编码（租户内唯一）
	SortOrder int        `gorm:"default:0" json:"sort_order"`
	Status    int        `gorm:"default:1" json:"status"`                                                           // 1=启用 0=禁用
	Version   int        `gorm:"not null;default:1" json:"version"`                                                 // 乐观锁版本号
	CreatedAt *time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (OrgUnit) TableName() string {
	return "admin_org_unit"
}

// BeforeCreate 创建前钩子，自动生成雪花 ID
func (o *OrgUnit) BeforeCreate(tx *gorm.DB) error {
	if o.ID == 0 {
		o.ID = NextID()
	}
	return nil
}
