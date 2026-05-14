package models

// TenantBy 租户字段 mixin
// 嵌入到所有需要租户隔离的业务 model 中
type TenantBy struct {
	// 租户 ID（0=超级管理员，不隔离）
	TenantId int `json:"tenantId" gorm:"index;default:0;comment:租户ID"`
}

// SetTenantId 设置租户 ID
func (t *TenantBy) SetTenantId(tenantId int) {
	t.TenantId = tenantId
}

// GetTenantId 获取租户 ID
func (t *TenantBy) GetTenantId() int {
	return t.TenantId
}
