// Package engine 提供 ABAC 策略引擎的核心组件：
// 资源属性 SPI 注册表、条件树翻译器、多策略合并器、列脱敏处理器。
package engine

import "sync"

// ResourceAttrDef 资源属性定义
type ResourceAttrDef struct {
	AttrName     string // 逻辑属性名，在条件表达式中使用
	Display      string // 显示名，前端编辑器展示
	QualifiedCol string // 带表限定符的物理列，如 "orders.dept_id"
	DataType     string // string|int|float|bool
	// JoinPath 关联表属性的 EXISTS 子查询模板。
	// 占位符：{main_table}=主表名，{col}=物理列，{op}=操作符，{param}=?
	// 为空时直接用 QualifiedCol 生成 WHERE 条件。
	JoinPath string
}

// ResourceDef 资源对象定义
type ResourceDef struct {
	Type        string            // 资源类型标识，如 "order"
	DisplayName string            // 显示名
	MainTable   string            // 主表名（用于 GORM Callback 中 resolveTableName 匹配）
	Attributes  []ResourceAttrDef // 属性列表
}

// GetAttr 按逻辑属性名查找属性定义
func (r *ResourceDef) GetAttr(attrName string) (ResourceAttrDef, bool) {
	for _, a := range r.Attributes {
		if a.AttrName == attrName {
			return a, true
		}
	}
	return ResourceAttrDef{}, false
}

// SubjectAttrDef 主体属性定义
type SubjectAttrDef struct {
	AttrName string // 属性名，在条件表达式 source=subject 中使用
	Display  string // 显示名
	DataType string // string|int|[]int 等
	// Resolver 运行时取值函数，authInfo 为当前认证信息 map（含 user_id/role_ids/dept_ids/tenant_id）
	Resolver func(authInfo map[string]interface{}) interface{}
}

// ---- 注册表（内存单例，并发安全）----

var (
	resourceMu  sync.RWMutex
	resourceMap = make(map[string]*ResourceDef)   // resource_type → ResourceDef
	tableMap    = make(map[string]*ResourceDef)   // main_table → ResourceDef（反查用）

	subjectMu      sync.RWMutex
	subjectAttrMap = make(map[string]SubjectAttrDef) // attr_name → SubjectAttrDef
)

// RegisterResource 注册资源对象（插件/业务模块启动时调用，并发安全）
func RegisterResource(def *ResourceDef) {
	resourceMu.Lock()
	defer resourceMu.Unlock()
	resourceMap[def.Type] = def
	if def.MainTable != "" {
		tableMap[def.MainTable] = def
	}
}

// GetResource 按资源类型查询已注册的资源定义
func GetResource(resourceType string) (*ResourceDef, bool) {
	resourceMu.RLock()
	defer resourceMu.RUnlock()
	def, ok := resourceMap[resourceType]
	return def, ok
}

// GetResourceByTable 按主表名反查资源定义（供 GORM Callback 使用）
func GetResourceByTable(tableName string) *ResourceDef {
	resourceMu.RLock()
	defer resourceMu.RUnlock()
	return tableMap[tableName]
}

// ListResources 列出所有已注册资源（供前端编辑器加载属性列表）
func ListResources() []*ResourceDef {
	resourceMu.RLock()
	defer resourceMu.RUnlock()
	result := make([]*ResourceDef, 0, len(resourceMap))
	for _, def := range resourceMap {
		result = append(result, def)
	}
	return result
}

// RegisterSubjectAttr 注册自定义主体属性（支持 SPI 扩展，并发安全）
func RegisterSubjectAttr(def SubjectAttrDef) {
	subjectMu.Lock()
	defer subjectMu.Unlock()
	subjectAttrMap[def.AttrName] = def
}

// GetSubjectAttr 查询主体属性定义
func GetSubjectAttr(attrName string) (SubjectAttrDef, bool) {
	subjectMu.RLock()
	defer subjectMu.RUnlock()
	def, ok := subjectAttrMap[attrName]
	return def, ok
}

// ListSubjectAttrs 列出所有已注册主体属性（供前端编辑器加载）
func ListSubjectAttrs() []SubjectAttrDef {
	subjectMu.RLock()
	defer subjectMu.RUnlock()
	result := make([]SubjectAttrDef, 0, len(subjectAttrMap))
	for _, def := range subjectAttrMap {
		result = append(result, def)
	}
	return result
}

// init 注册内置主体属性
func init() {
	builtins := []SubjectAttrDef{
		{
			AttrName: "user_id",
			Display:  "用户 ID",
			DataType: "string",
			Resolver: func(a map[string]interface{}) interface{} { return a["user_id"] },
		},
		{
			AttrName: "role_ids",
			Display:  "角色列表",
			DataType: "[]string",
			Resolver: func(a map[string]interface{}) interface{} { return a["role_ids"] },
		},
		{
			AttrName: "dept_ids",
			Display:  "部门列表",
			DataType: "[]string",
			Resolver: func(a map[string]interface{}) interface{} { return a["dept_ids"] },
		},
		{
			AttrName: "tenant_id",
			Display:  "租户 ID",
			DataType: "string",
			Resolver: func(a map[string]interface{}) interface{} { return a["tenant_id"] },
		},
	}
	for _, b := range builtins {
		subjectAttrMap[b.AttrName] = b
	}
}
