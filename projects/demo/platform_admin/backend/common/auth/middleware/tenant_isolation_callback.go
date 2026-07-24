package middleware

import (
	"gorm.io/gorm"
)

// RegisterTenantIsolationCallback 注册租户隔离 GORM Callback
// 顺序保证：先于 DataScopeCallback 注册，确保执行顺序 tenant_isolation → data_scope → gorm:query
// 注意：必须在 RegisterDataScopeCallback 之前调用
func RegisterTenantIsolationCallback(db *gorm.DB) {
	db.Callback().Create().Before("gorm:create").Register("auth:tenant_create", tenantCreateCallback)
	db.Callback().Query().Before("gorm:query").Register("auth:tenant_query", tenantQueryCallback)
	db.Callback().Update().Before("gorm:update").Register("auth:tenant_update", tenantUpdateCallback)
	db.Callback().Delete().Before("gorm:delete").Register("auth:tenant_delete", tenantDeleteCallback)
}

// hasTenantField 检测表是否含 tenant_id 字段（通过 GORM Schema 反射，零配置）
func hasTenantField(db *gorm.DB) bool {
	if db.Statement == nil || db.Statement.Schema == nil {
		return false
	}
	return db.Statement.Schema.LookUpField("TenantID") != nil
}

// getTenantIDFromContext 从 GORM Statement 的 context 中获取 tenant_id
// 优先从 AuthInfo 获取（与 DataScopeCallback 共享同一 context 注入路径）
func getTenantIDFromGormContext(db *gorm.DB) int64 {
	if db.Statement == nil || db.Statement.Context == nil {
		return 0
	}
	info := GetAuthInfo(db.Statement.Context)
	if info != nil {
		return info.TenantID
	}
	return 0
}

// tenantCreateCallback Create 操作：自动填充 tenant_id
func tenantCreateCallback(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Context == nil {
		return
	}
	if IsSystemOp(db.Statement.Context) {
		return
	}
	if !hasTenantField(db) {
		return
	}
	tenantID := getTenantIDFromGormContext(db)
	if tenantID == 0 {
		return
	}
	db.Statement.SetColumn("TenantID", tenantID)
}

// tenantQueryCallback Query 操作：自动注入 WHERE tenant_id = ?
func tenantQueryCallback(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Context == nil {
		return
	}
	if IsSystemOp(db.Statement.Context) {
		return
	}
	if !hasTenantField(db) {
		return
	}
	tenantID := getTenantIDFromGormContext(db)
	if tenantID == 0 {
		return
	}
	db.Where("tenant_id = ?", tenantID)
}

// tenantUpdateCallback Update 操作：自动注入 WHERE tenant_id = ?
func tenantUpdateCallback(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Context == nil {
		return
	}
	if IsSystemOp(db.Statement.Context) {
		return
	}
	if !hasTenantField(db) {
		return
	}
	tenantID := getTenantIDFromGormContext(db)
	if tenantID == 0 {
		return
	}
	db.Where("tenant_id = ?", tenantID)
}

// tenantDeleteCallback Delete 操作：自动注入 WHERE tenant_id = ?
func tenantDeleteCallback(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Context == nil {
		return
	}
	if IsSystemOp(db.Statement.Context) {
		return
	}
	if !hasTenantField(db) {
		return
	}
	tenantID := getTenantIDFromGormContext(db)
	if tenantID == 0 {
		return
	}
	db.Where("tenant_id = ?", tenantID)
}
