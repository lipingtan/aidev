package middleware

import (
	"go-admin/common/auth/model"
	"go-admin/common/auth/spi"

	"gorm.io/gorm"
)

// orgProvider 组织架构提供者（通过 RegisterDataScopeCallback 注入）
var dataScopeOrgProvider spi.OrganizationProvider

// RegisterDataScopeCallback 注册数据权限 GORM Callback
// enabled=false 时不注册，直接返回
func RegisterDataScopeCallback(db *gorm.DB, enabled bool, orgProvider spi.OrganizationProvider) {
	if !enabled {
		return
	}
	dataScopeOrgProvider = orgProvider
	_ = db.Callback().Query().After("auth:tenant_query").Before("gorm:query").Register("auth:data_scope", dataScopeQueryCallback)
}

// dataScopeQueryCallback GORM Callback 实现：自动注入数据权限 WHERE 条件
func dataScopeQueryCallback(db *gorm.DB) {
	if db.Statement == nil || db.Statement.Context == nil {
		return
	}

	ctx := db.Statement.Context

	// SystemOp 标记跳过
	if IsSystemOp(ctx) {
		return
	}

	// 获取数据权限上下文
	dsc := GetDataScopeContext(ctx)

	// 获取当前查询的目标表名
	targetTable := resolveTableName(db)

	// scope_type 分支注入
	hasDimensionScope := false
	if dsc != nil && len(dsc.Dimensions) > 0 && targetTable != "" {
		hasDimensionScope = injectScopeTypeWhere(db, dsc, targetTable)
	}

	// 记录共享规则扩展：追加 OR id IN (共享命中的 record_id)
	injectRecordShareScope(db, hasDimensionScope)
}

// injectScopeTypeWhere 根据 scope_type 注入 WHERE 条件
func injectScopeTypeWhere(db *gorm.DB, dsc *DataScopeContext, targetTable string) bool {
	// 先检查是否有 ALL 短路（任一角色对该实体配置 ALL → 不注入任何条件）
	for _, dim := range dsc.Dimensions {
		if dim.TargetEntity == targetTable && dim.ScopeType == "ALL" {
			return false
		}
	}

	injected := false
	for _, dim := range dsc.Dimensions {
		if dim.TargetEntity != targetTable {
			continue
		}
		switch dim.ScopeType {
		case "ALL":
			continue // 不注入（已在短路中处理，此处防御性保留）
		case "SELF":
			db.Where("create_by = ?", dim.UserID)
			injected = true
		case "DEPT":
			if dataScopeOrgProvider != nil {
				orgIDs, _ := dataScopeOrgProvider.GetOrgIds(dim.UserID, dim.TenantID)
				if len(orgIDs) > 0 {
					db.Where(dim.ColumnName+" IN ?", orgIDs)
				} else {
					db.Where("1 = 0")
				}
				injected = true
			}
		case "DEPT_TREE":
			if dataScopeOrgProvider != nil {
				orgIDs, _ := dataScopeOrgProvider.GetOrgIds(dim.UserID, dim.TenantID)
				allIDs := make([]int64, 0)
				for _, oid := range orgIDs {
					subIDs, _ := dataScopeOrgProvider.GetSubOrgIds(oid, dim.TenantID)
					allIDs = append(allIDs, subIDs...)
				}
				// 去重
				allIDs = uniqueInt64(allIDs)
				if len(allIDs) > 0 {
					db.Where(dim.ColumnName+" IN ?", allIDs)
				} else {
					db.Where("1 = 0")
				}
				injected = true
			}
		case "CUSTOM":
			// 保持现有 CUSTOM 逻辑
			if len(dim.Values) > 0 {
				db.Where(dim.ColumnName+" IN ?", dim.Values)
				injected = true
			}
		default:
			// 未知 scope_type 按 CUSTOM 处理
			if len(dim.Values) > 0 {
				db.Where(dim.ColumnName+" IN ?", dim.Values)
				injected = true
			}
		}
	}
	return injected
}

// injectRecordShareScope 注入记录共享 OR 子查询
func injectRecordShareScope(db *gorm.DB, _ bool) {
	ctx := db.Statement.Context

	objectCode := GetObjectCode(ctx)
	if objectCode == "" {
		return
	}

	authInfo := GetAuthInfo(ctx)
	if authInfo == nil {
		return
	}

	shareSubQuery := db.Session(&gorm.Session{NewDB: true}).
		Model(&model.RecordShare{}).
		Select("record_id").
		Where("tenant_id = ? AND object_code = ?", authInfo.TenantID, objectCode).
		Where("expire_at IS NULL OR expire_at > NOW()")

	targetCond := db.Session(&gorm.Session{NewDB: true}).
		Where("share_to_type = 'USER' AND share_to_id = ?", authInfo.UserID)
	if len(authInfo.RoleIDs) > 0 {
		targetCond = targetCond.Or("share_to_type = 'ROLE' AND share_to_id IN ?", authInfo.RoleIDs)
	}
	if len(authInfo.DeptIDs) > 0 {
		targetCond = targetCond.Or("share_to_type = 'DEPT' AND share_to_id IN ?", authInfo.DeptIDs)
	}
	shareSubQuery = shareSubQuery.Where(targetCond)

	db.Or("id IN (?)", shareSubQuery)
}

// resolveTableName 从 GORM Statement 中解析目标表名
func resolveTableName(db *gorm.DB) string {
	if db.Statement.Table != "" {
		return db.Statement.Table
	}
	if db.Statement.Schema != nil {
		return db.Statement.Schema.Table
	}
	return ""
}

// uniqueInt64 去重 int64 切片
func uniqueInt64(input []int64) []int64 {
	seen := make(map[int64]struct{})
	result := make([]int64, 0, len(input))
	for _, v := range input {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
