package middleware

import "context"

// dataScopeContextKey context 存取键
type dataScopeContextKey struct{}

// systemOpContextKey 系统操作标记键
type systemOpContextKey struct{}

// DataScopeDimension 单个维度的数据权限配置
type DataScopeDimension struct {
	DimensionName string   // 维度标识
	TargetEntity  string   // 目标实体（表名）
	ColumnName    string   // 过滤字段名
	Values        []string // 允许的值列表
}

// DataScopeContext 数据权限上下文，包含当前用户所有维度配置
type DataScopeContext struct {
	Dimensions []DataScopeDimension
}

// SetDataScopeContext 将数据权限配置写入 context
func SetDataScopeContext(ctx context.Context, dsc *DataScopeContext) context.Context {
	return context.WithValue(ctx, dataScopeContextKey{}, dsc)
}

// GetDataScopeContext 从 context 获取数据权限配置
func GetDataScopeContext(ctx context.Context) *DataScopeContext {
	val, ok := ctx.Value(dataScopeContextKey{}).(*DataScopeContext)
	if !ok {
		return nil
	}
	return val
}

// SystemOpContext 返回标记为系统操作的 context（跳过数据权限注入）
func SystemOpContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, systemOpContextKey{}, true)
}

// IsSystemOp 检查 context 是否标记为系统操作
func IsSystemOp(ctx context.Context) bool {
	val, ok := ctx.Value(systemOpContextKey{}).(bool)
	return ok && val
}
