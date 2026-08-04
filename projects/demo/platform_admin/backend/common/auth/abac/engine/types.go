package engine

// CondNode 条件表达式节点（JSON 可序列化的递归树）
// type="expr" → 叶子节点（属性比较）
// type="group" → 组合节点（AND/OR）
type CondNode struct {
	Type     string      `json:"type"`               // "expr" | "group"
	Left     *CondValue  `json:"left,omitempty"`     // expr 专用：左值
	Op       string      `json:"op,omitempty"`       // expr 专用：操作符
	Right    *CondValue  `json:"right,omitempty"`    // expr 专用：右值
	Operator string      `json:"operator,omitempty"` // group 专用："AND" | "OR"
	Children []*CondNode `json:"children,omitempty"` // group 专用：子节点
}

// CondValue 条件值（左值或右值）
type CondValue struct {
	Source string      `json:"source"`          // "resource" | "subject" | "const"
	Attr   string      `json:"attr,omitempty"`  // source=resource/subject 时的属性名
	Value  interface{} `json:"value,omitempty"` // source=const 时的字面量
}

// 支持的操作符常量
const (
	OpEq        = "eq"
	OpNe        = "ne"
	OpGt        = "gt"
	OpLt        = "lt"
	OpGte       = "gte"
	OpLte       = "lte"
	OpIn        = "in"
	OpNotIn     = "not_in"
	OpContains  = "contains"
	OpIsNull    = "is_null"
	OpIsNotNull = "is_not_null"
)

// source 类型常量
const (
	SourceResource = "resource"
	SourceSubject  = "subject"
	SourceConst    = "const"
)

// effect 常量
const (
	EffectAllow = "ALLOW"
	EffectDeny  = "DENY"
)

// col effect 常量
const (
	ColEffectShow = "SHOW"
	ColEffectHide = "HIDE"
	ColEffectMask = "MASK"
)

// action 常量
const (
	ActionRead   = "read"
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// subject type 常量
const (
	SubjectRole          = "ROLE"
	SubjectPermissionSet = "PERMISSION_SET"
	SubjectUser          = "USER"
	SubjectDept          = "DEPT"
)
