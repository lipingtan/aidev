package engine

import (
	"strings"
	"testing"
)

// ---- 翻译器边界用例 ----

func TestTranslate_NilNode(t *testing.T) {
	r := Translate(nil, testAuthInfo, testResDef)
	if r.SQL != "1 = 1" {
		t.Errorf("nil 节点应返回 '1 = 1'，got: %s", r.SQL)
	}
}

func TestTranslate_EmptyGroup(t *testing.T) {
	node := &CondNode{Type: "group", Operator: "AND", Children: []*CondNode{}}
	r := Translate(node, testAuthInfo, testResDef)
	if r.SQL != "1 = 1" {
		t.Errorf("空 group 节点应返回 '1 = 1'，got: %s", r.SQL)
	}
}

func TestTranslate_UnknownNodeType(t *testing.T) {
	node := &CondNode{Type: "unknown"}
	r := Translate(node, testAuthInfo, testResDef)
	if r.SQL != "1 = 0" {
		t.Errorf("未知节点类型应返回 '1 = 0'，got: %s", r.SQL)
	}
}

func TestTranslate_LeftSourceNotResource(t *testing.T) {
	node := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: "subject", Attr: "user_id"},
		Op:    OpEq,
		Right: &CondValue{Source: SourceConst, Value: "1"},
	}
	r := Translate(node, testAuthInfo, testResDef)
	if r.SQL != "1 = 0" {
		t.Errorf("左值不是 resource 应返回 '1 = 0'，got: %s", r.SQL)
	}
}

func TestTranslate_NilLeft(t *testing.T) {
	node := &CondNode{Type: "expr", Op: OpEq}
	r := Translate(node, testAuthInfo, testResDef)
	if r.SQL != "1 = 0" {
		t.Errorf("nil left 应返回 '1 = 0'，got: %s", r.SQL)
	}
}

func TestTranslate_IsNotNull(t *testing.T) {
	node := &CondNode{
		Type: "expr",
		Left: &CondValue{Source: SourceResource, Attr: "status"},
		Op:   OpIsNotNull,
	}
	r := Translate(node, testAuthInfo, testResDef)
	if r.SQL != "orders.status IS NOT NULL" {
		t.Errorf("IS NOT NULL 期望正确 SQL，got: %s", r.SQL)
	}
}

func TestTranslate_ContainsOp(t *testing.T) {
	node := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: SourceResource, Attr: "status"},
		Op:    OpContains,
		Right: &CondValue{Source: SourceConst, Value: "active"},
	}
	r := Translate(node, testAuthInfo, testResDef)
	if !strings.Contains(r.SQL, "LIKE") {
		t.Errorf("contains 操作符应翻译为 LIKE，got: %s", r.SQL)
	}
	if len(r.Args) == 0 || !strings.Contains(r.Args[0].(string), "active") {
		t.Errorf("LIKE 参数应包含 active，got: %v", r.Args)
	}
}

func TestTranslate_InOp(t *testing.T) {
	node := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: SourceResource, Attr: "dept_id"},
		Op:    OpIn,
		Right: &CondValue{Source: SourceSubject, Attr: "dept_ids"},
	}
	r := Translate(node, testAuthInfo, testResDef)
	if !strings.Contains(r.SQL, "IN (?)") {
		t.Errorf("in 操作符应翻译为 IN (?)，got: %s", r.SQL)
	}
}

func TestTranslate_UnknownSubjectAttr(t *testing.T) {
	node := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: SourceResource, Attr: "dept_id"},
		Op:    OpEq,
		Right: &CondValue{Source: SourceSubject, Attr: "nonexistent_subject_attr"},
	}
	r := Translate(node, testAuthInfo, testResDef)
	// 未知主体属性 resolver 返回 nil，参数为 nil 但 SQL 仍应合法
	if !strings.Contains(r.SQL, "orders.dept_id") {
		t.Errorf("未知主体属性时左值仍应正确，got: %s", r.SQL)
	}
}

// ---- 合并器边界用例 ----

func TestMergeRow_NilAuthInfo(t *testing.T) {
	rows := []PolicyRowEntry{
		{Effect: EffectAllow, Action: ActionRead, ConditionNode: nil},
	}
	// nil authInfo 时不应 panic，ALLOW 无条件直接返回
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("nil authInfo 不应 panic: %v", r)
		}
	}()
	result := MergeRowPolicies(rows, ActionRead, nil, testResDef)
	if result.Deny {
		t.Error("无条件 ALLOW 不应 Deny")
	}
}

func TestMergeRow_EmptyAction(t *testing.T) {
	rows := []PolicyRowEntry{
		{Effect: EffectAllow, Action: ActionRead, ConditionNode: nil},
	}
	// action="" 不过滤，所有行权限均命中
	result := MergeRowPolicies(rows, "", testAuthInfo, testResDef)
	if result.Deny {
		t.Error("action='' 时不应 Deny")
	}
}

func TestMergeRow_DenyBeatsMultipleAllow(t *testing.T) {
	node := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: SourceResource, Attr: "status"},
		Op:    OpEq,
		Right: &CondValue{Source: SourceConst, Value: "active"},
	}
	rows := []PolicyRowEntry{
		{Effect: EffectAllow, Action: ActionRead, ConditionNode: node},
		{Effect: EffectAllow, Action: ActionRead, ConditionNode: node},
		{Effect: EffectDeny, Action: ActionRead, ConditionNode: nil},
		{Effect: EffectAllow, Action: ActionRead, ConditionNode: node},
	}
	result := MergeRowPolicies(rows, ActionRead, testAuthInfo, testResDef)
	if !result.Deny {
		t.Error("任一 DENY 应立即返回 Deny=true")
	}
}

// ---- 脱敏边界用例 ----

func TestMaskPhone_ShortInput(t *testing.T) {
	got := maskPhone("123")
	if got != "***" {
		t.Errorf("短手机号应返回 '***'，got: %s", got)
	}
}

func TestMaskEmail_NoAt(t *testing.T) {
	got := maskEmail("notanemail")
	if got != "***" {
		t.Errorf("无 @ 邮箱应返回 '***'，got: %s", got)
	}
}

func TestMaskCustom_InvalidPattern(t *testing.T) {
	original := "test123"
	got := maskCustom(original, "[invalid(")
	if got != original {
		t.Errorf("无效正则应保留原值，got: %s", got)
	}
}

func TestMaskCustom_ValidPattern(t *testing.T) {
	got := maskCustom("phone:13812345678", `\d{4,}`)
	if !strings.Contains(got, "***") {
		t.Errorf("有效正则应替换匹配部分，got: %s", got)
	}
}

func TestApplyColPolicy_EmptyEffects(t *testing.T) {
	data := map[string]interface{}{"name": "张三"}
	result := ApplyColPolicy(data, nil)
	m, ok := result.(map[string]interface{})
	if !ok || m["name"] != "张三" {
		t.Error("空 effects 时原样返回")
	}
}

func TestApplyColPolicy_Slice(t *testing.T) {
	data := []map[string]interface{}{
		{"phone": "13812345678"},
		{"phone": "13900000000"},
	}
	effects := map[string]PolicyColEntry{
		"phone": {FieldName: "phone", Effect: ColEffectHide},
	}
	result := ApplyColPolicy(data, effects)
	list, ok := result.([]map[string]interface{})
	if !ok {
		t.Fatal("slice 应返回 []map")
	}
	for _, item := range list {
		if item["phone"] != nil {
			t.Errorf("HIDE 后 phone 应为 nil，got: %v", item["phone"])
		}
	}
}

func TestMergeCol_EmptyInput(t *testing.T) {
	result := MergeColPolicies(nil)
	if len(result) != 0 {
		t.Error("空输入应返回空 map")
	}
}

func TestMergeCol_ShowDoesNotOverrideMask(t *testing.T) {
	cols := []PolicyColEntry{
		{FieldName: "email", Effect: ColEffectMask, MaskType: "email"},
		{FieldName: "email", Effect: ColEffectShow},
	}
	result := MergeColPolicies(cols)
	if result["email"].Effect != ColEffectMask {
		t.Errorf("SHOW 不应覆盖 MASK，期望 MASK，got: %s", result["email"].Effect)
	}
}
