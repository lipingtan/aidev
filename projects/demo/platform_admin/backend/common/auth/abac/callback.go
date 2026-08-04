// Package abac 提供 ABAC 策略引擎的 GORM Callback 注册和公开 API。
package abac

import (
	"context"
	"fmt"
	"strings"

	"go-admin/common/auth/abac/engine"
	"go-admin/common/auth/abac/service"
	"go-admin/common/auth/middleware"

	"gorm.io/gorm"
)

// policyLoader 策略加载器（由 Init 注入）
var policyLoader *service.AbacService

// RegisterAbacCallback 注册 ABAC GORM Callback。
// 必须在 RegisterDataScopeCallback 之后调用。
func RegisterAbacCallback(db *gorm.DB, svc *service.AbacService) {
	if svc == nil {
		return
	}
	policyLoader = svc
	_ = db.Callback().Query().
		After("auth:data_scope").
		Before("gorm:query").
		Register("auth:abac", abacQueryCallback)
}

// abacQueryCallback GORM Callback：根据 ABAC 策略注入行过滤 WHERE 条件，并缓存列权限到 context。
func abacQueryCallback(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Context == nil {
		return
	}
	ctx := db.Statement.Context

	// SystemOp 标记跳过
	if middleware.IsSystemOp(ctx) {
		return
	}

	// 获取认证信息
	authInfo := middleware.GetAuthInfo(ctx)
	if authInfo == nil {
		return
	}

	// 解析目标表名
	targetTable := resolveTableName(db)
	if targetTable == "" {
		return
	}

	// 通过主表名反查资源定义
	resDef := engine.GetResourceByTable(targetTable)
	if resDef == nil {
		return // 未注册 → 不注入
	}

	// 加载策略（含缓存）
	userIDStr := fmt.Sprintf("%d", authInfo.UserID)
	rowEntries, colEntries := policyLoader.LoadPoliciesForCallback(authInfo.TenantID, userIDStr, resDef.Type)

	// 无任何策略配置 → 不注入（不影响现有查询逻辑，RG-1）
	if len(rowEntries) == 0 && len(colEntries) == 0 {
		return
	}

	// 构建 authInfo map 供翻译器使用
	authInfoMap := buildAuthInfoMap(authInfo)

	// 获取当前 action
	action := engine.GetAbacAction(ctx)

	// 合并行权限
	rowResult := engine.MergeRowPolicies(rowEntries, action, authInfoMap, resDef)
	if rowResult.Deny {
		db.Where("1 = 0")
		return
	}

	// 注入 ALLOW 条件（多个取 OR）
	if len(rowResult.Conditions) > 0 {
		orParts := make([]string, 0, len(rowResult.Conditions))
		var allArgs []interface{}
		for _, cond := range rowResult.Conditions {
			orParts = append(orParts, "("+cond.SQL+")")
			allArgs = append(allArgs, cond.Args...)
		}
		db.Where(strings.Join(orParts, " OR "), allArgs...)
	}

	// 合并列权限，写入 context 供 ApplyColPolicy 使用
	if len(colEntries) > 0 {
		colEffects := engine.MergeColPolicies(colEntries)
		// context 是 immutable 的，但 db.Statement.Context 可以被替换
		newCtx := engine.SetAbacColContext(ctx, colEffects)
		db.Statement.Context = newCtx
	}
}

// resolveTableName 从 GORM Statement 解析目标表名
func resolveTableName(db *gorm.DB) string {
	if db.Statement.Table != "" {
		return db.Statement.Table
	}
	if db.Statement.Schema != nil {
		return db.Statement.Schema.Table
	}
	return ""
}

// buildAuthInfoMap 将 middleware.AuthInfo 转换为翻译器需要的 map
func buildAuthInfoMap(authInfo *middleware.AuthInfo) map[string]interface{} {
	roleIDs := make([]string, 0, len(authInfo.RoleIDs))
	for _, r := range authInfo.RoleIDs {
		roleIDs = append(roleIDs, fmt.Sprintf("%d", r))
	}
	deptIDs := make([]string, 0, len(authInfo.DeptIDs))
	for _, d := range authInfo.DeptIDs {
		deptIDs = append(deptIDs, fmt.Sprintf("%d", d))
	}
	return map[string]interface{}{
		"user_id":   fmt.Sprintf("%d", authInfo.UserID),
		"role_ids":  roleIDs,
		"dept_ids":  deptIDs,
		"tenant_id": fmt.Sprintf("%d", authInfo.TenantID),
	}
}

// ApplyColPolicy 公开函数：对响应数据应用列权限（Handler 主动调用一行）。
// ctx 需为 GORM Callback 已处理的 context（含 AbacColContext）。
func ApplyColPolicy(ctx context.Context, data interface{}) interface{} {
	colEffects := engine.GetAbacColContext(ctx)
	if len(colEffects) == 0 {
		return data
	}
	return engine.ApplyColPolicy(data, colEffects)
}
