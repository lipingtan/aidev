package middleware

import (
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
	if dsc == nil || len(dsc.Dimensions) == 0 {
		return
	}

	// 获取当前查询的目标表名
	targetTable := resolveTableName(db)
	if targetTable == "" {
		return
	}

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
	}
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
