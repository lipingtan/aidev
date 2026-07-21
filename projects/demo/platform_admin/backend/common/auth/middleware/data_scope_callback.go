package middleware

import (
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// RegisterDataScopeCallback 注册数据权限 GORM Callback
// enabled=false 时不注册，直接返回
func RegisterDataScopeCallback(db *gorm.DB, enabled bool) {
	if !enabled {
		return
	}
	_ = db.Callback().Query().Before("gorm:query").Register("auth:data_scope", dataScopeQueryCallback)
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

	// 维度数据权限注入
	hasDimensionScope := false
	if dsc != nil && len(dsc.Dimensions) > 0 && targetTable != "" {
		hasDimensionScope = injectDimensionScope(db, dsc, targetTable)
	}

	// 记录共享规则扩展：追加 OR id IN (共享命中的 record_id)
	// 无 objectCode 时不注入（RG-4：行为与 CR-1 一致）
	injectRecordShareScope(db, hasDimensionScope)
}

// injectDimensionScope 注入维度数据权限 WHERE 条件，返回是否注入了条件
func injectDimensionScope(db *gorm.DB, dsc *DataScopeContext, targetTable string) bool {
	// 按维度分组合并值（多角色同维度取并集）
	type dimKey struct {
		DimensionName string
		ColumnName    string
	}
	merged := make(map[dimKey]map[string]struct{})

	for _, dim := range dsc.Dimensions {
		// target_entity 不匹配时跳过
		if dim.TargetEntity != targetTable {
			continue
		}
		key := dimKey{DimensionName: dim.DimensionName, ColumnName: dim.ColumnName}
		if merged[key] == nil {
			merged[key] = make(map[string]struct{})
		}
		for _, v := range dim.Values {
			merged[key][v] = struct{}{}
		}
	}

	injected := false
	// 拼接 WHERE 条件
	for key, valSet := range merged {
		if len(valSet) == 0 {
			continue
		}
		values := make([]string, 0, len(valSet))
		for v := range valSet {
			values = append(values, v)
		}
		db.Where(key.ColumnName+" IN (?)", values)
		injected = true
	}
	return injected
}

// injectRecordShareScope 注入记录共享 OR 子查询
// 当 context 中存在 objectCode 和 AuthInfo 时，追加 OR id IN (共享命中) 条件
// hasDimensionScope 表示是否已有维度权限条件（用于决定是否需要用 OR 包裹）
func injectRecordShareScope(db *gorm.DB, _ bool) {
	ctx := db.Statement.Context

	// 获取 objectCode，无则不注入（RG-4）
	objectCode := GetObjectCode(ctx)
	if objectCode == "" {
		return
	}

	// 获取用户身份信息
	authInfo := GetAuthInfo(ctx)
	if authInfo == nil {
		return
	}

	// 构建共享规则子查询：查询 admin_record_share 表中命中的 record_id
	shareSubQuery := db.Session(&gorm.Session{NewDB: true}).
		Model(&model.RecordShare{}).
		Select("record_id").
		Where("tenant_id = ? AND object_code = ?", authInfo.TenantID, objectCode).
		Where("expire_at IS NULL OR expire_at > NOW()")

	// 构建共享目标匹配条件（USER/ROLE/DEPT 任一命中即可）
	targetCond := db.Session(&gorm.Session{NewDB: true}).
		Where("share_to_type = 'USER' AND share_to_id = ?", authInfo.UserID)
	if len(authInfo.RoleIDs) > 0 {
		targetCond = targetCond.Or("share_to_type = 'ROLE' AND share_to_id IN ?", authInfo.RoleIDs)
	}
	if len(authInfo.DeptIDs) > 0 {
		targetCond = targetCond.Or("share_to_type = 'DEPT' AND share_to_id IN ?", authInfo.DeptIDs)
	}
	shareSubQuery = shareSubQuery.Where(targetCond)

	// 追加 OR id IN (共享命中)：不缩小原有数据权限范围
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
