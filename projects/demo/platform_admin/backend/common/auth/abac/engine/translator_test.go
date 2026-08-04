package engine

import (
	"strings"
	"testing"
)

// 测试用资源定义
var testResDef = &ResourceDef{
	Type:      "order",
	MainTable: "orders",
	Attributes: []ResourceAttrDef{
		{AttrName: "dept_id", QualifiedCol: "orders.dept_id", DataType: "int"},
		{AttrName: "create_by", QualifiedCol: "orders.create_by", DataType: "int"},
		{AttrName: "manager_id", QualifiedCol: "orders.manager_id", DataType: "int"},
		{AttrName: "status", QualifiedCol: "orders.status", DataType: "string"},
		{AttrName: "user_dept", QualifiedCol: "users.dept_id", DataType: "int",
			JoinPath: "SELECT 1 FROM users WHERE users.id = {main_table}.create_by AND {col} {op} {param}"},
	},
}

var testAuthInfo = map[string]interface{}{
	"user_id":   "42",
	"role_ids":  []string{"sales"},
	"dept_ids":  []string{"10", "11"},
	"tenant_id": "1",
}

func TestTranslate_SimpleEq(t *testing.T) {
	node := &CondNode{
		Type: "expr",
		Left: &CondValue{Source: SourceResource, Attr: "dept_id"},
		Op:   OpEq,
		Right: &CondValue{Source: SourceSubject, Attr: "dept_ids"},
	}
	r := Translate(node, testAuthInfo, testResDef)
	if !strings.Contains(r.SQL, "orders.dept_id") {
		t.Errorf("期望包含 orders.dept_id，got: %s", r.SQL)
	}
	if len(r.Args) == 0 {
		t.Error("期望有参数")
	}
}

func TestTranslate_TwoColCompare(t *testing.T) {
	node := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: SourceResource, Attr: "create_by"},
		Op:    OpEq,
		Right: &CondValue{Source: SourceResource, Attr: "manager_id"},
	}
	r := Translate(node, testAuthInfo, testResDef)
	expected := "orders.create_by = orders.manager_id"
	if r.SQL != expected {
		t.Errorf("两列互比期望 %q，got: %q", expected, r.SQL)
	}
	if len(r.Args) != 0 {
		t.Error("两列互比不应有参数绑定")
	}
}

func TestTranslate_JoinPath(t *testing.T) {
	node := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: SourceResource, Attr: "user_dept"},
		Op:    OpEq,
		Right: &CondValue{Source: SourceConst, Value: "10"},
	}
	r := Translate(node, testAuthInfo, testResDef)
	if !strings.HasPrefix(r.SQL, "EXISTS (") {
		t.Errorf("带 JoinPath 应翻译为 EXISTS，got: %s", r.SQL)
	}
	if len(r.Args) == 0 {
		t.Error("EXISTS 子查询应有参数")
	}
}

func TestTranslate_UnknownAttr(t *testing.T) {
	node := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: SourceResource, Attr: "nonexistent_field"},
		Op:    OpEq,
		Right: &CondValue{Source: SourceConst, Value: "x"},
	}
	r := Translate(node, testAuthInfo, testResDef)
	if r.SQL != "1 = 0" {
		t.Errorf("未注册属性期望返回 '1 = 0'，got: %s", r.SQL)
	}
}

func TestTranslate_GroupAndOr(t *testing.T) {
	// 内层 OR group 需要两个以上子节点才会生成 OR 关键词
	node := &CondNode{
		Type:     "group",
		Operator: "AND",
		Children: []*CondNode{
			{
				Type:  "expr",
				Left:  &CondValue{Source: SourceResource, Attr: "dept_id"},
				Op:    OpEq,
				Right: &CondValue{Source: SourceConst, Value: "10"},
			},
			{
				Type:     "group",
				Operator: "OR",
				Children: []*CondNode{
					{
						Type:  "expr",
						Left:  &CondValue{Source: SourceResource, Attr: "status"},
						Op:    OpEq,
						Right: &CondValue{Source: SourceConst, Value: "active"},
					},
					{
						Type:  "expr",
						Left:  &CondValue{Source: SourceResource, Attr: "status"},
						Op:    OpEq,
						Right: &CondValue{Source: SourceConst, Value: "pending"},
					},
				},
			},
		},
	}
	r := Translate(node, testAuthInfo, testResDef)
	if !strings.Contains(r.SQL, "AND") {
		t.Errorf("组合节点应含 AND，got: %s", r.SQL)
	}
	if !strings.Contains(r.SQL, "OR") {
		t.Errorf("嵌套节点应含 OR，got: %s", r.SQL)
	}
	if len(r.Args) != 3 {
		t.Errorf("期望 3 个参数，got: %d", len(r.Args))
	}
}

func TestTranslate_IsNull(t *testing.T) {
	node := &CondNode{
		Type: "expr",
		Left: &CondValue{Source: SourceResource, Attr: "status"},
		Op:   OpIsNull,
	}
	r := Translate(node, testAuthInfo, testResDef)
	if r.SQL != "orders.status IS NULL" {
		t.Errorf("IS NULL 期望 'orders.status IS NULL'，got: %s", r.SQL)
	}
}
