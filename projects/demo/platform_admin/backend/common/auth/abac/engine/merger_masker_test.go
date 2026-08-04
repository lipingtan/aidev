package engine

import (
	"testing"
)

// ---- 行权限合并测试 ----

func TestMergeRow_DenyPriority(t *testing.T) {
	rows := []PolicyRowEntry{
		{Effect: EffectAllow, Action: ActionRead, ConditionNode: nil},
		{Effect: EffectDeny, Action: ActionRead, ConditionNode: nil},
	}
	result := MergeRowPolicies(rows, ActionRead, testAuthInfo, testResDef)
	if !result.Deny {
		t.Error("有 DENY 策略时 Deny 应为 true")
	}
}

func TestMergeRow_MultiAllow(t *testing.T) {
	node1 := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: SourceResource, Attr: "dept_id"},
		Op:    OpEq,
		Right: &CondValue{Source: SourceConst, Value: "10"},
	}
	node2 := &CondNode{
		Type:  "expr",
		Left:  &CondValue{Source: SourceResource, Attr: "status"},
		Op:    OpEq,
		Right: &CondValue{Source: SourceConst, Value: "active"},
	}
	rows := []PolicyRowEntry{
		{Effect: EffectAllow, Action: ActionRead, ConditionNode: node1},
		{Effect: EffectAllow, Action: ActionRead, ConditionNode: node2},
	}
	result := MergeRowPolicies(rows, ActionRead, testAuthInfo, testResDef)
	if result.Deny {
		t.Error("多个 ALLOW 不应返回 Deny")
	}
	if len(result.Conditions) != 2 {
		t.Errorf("期望 2 个 ALLOW 条件，got: %d", len(result.Conditions))
	}
}

func TestMergeRow_NoPolicy(t *testing.T) {
	result := MergeRowPolicies(nil, ActionRead, testAuthInfo, testResDef)
	if result.Deny {
		t.Error("无策略时不应 Deny")
	}
	if len(result.Conditions) != 0 {
		t.Error("无策略时不应有条件")
	}
}

func TestMergeRow_ActionFilter(t *testing.T) {
	rows := []PolicyRowEntry{
		{Effect: EffectAllow, Action: ActionCreate, ConditionNode: nil},
	}
	// read action 无匹配策略
	result := MergeRowPolicies(rows, ActionRead, testAuthInfo, testResDef)
	if result.Deny || len(result.Conditions) != 0 {
		t.Error("action 不匹配时应返回空结果")
	}
}

// ---- 列权限合并测试 ----

func TestMergeCol_Priority(t *testing.T) {
	cols := []PolicyColEntry{
		{FieldName: "phone", Effect: ColEffectShow},
		{FieldName: "phone", Effect: ColEffectMask, MaskType: "phone"},
		{FieldName: "phone", Effect: ColEffectHide},
	}
	result := MergeColPolicies(cols)
	if result["phone"].Effect != ColEffectHide {
		t.Errorf("HIDE > MASK > SHOW，期望 HIDE，got: %s", result["phone"].Effect)
	}
}

func TestMergeCol_MaskOverShow(t *testing.T) {
	cols := []PolicyColEntry{
		{FieldName: "email", Effect: ColEffectShow},
		{FieldName: "email", Effect: ColEffectMask, MaskType: "email"},
	}
	result := MergeColPolicies(cols)
	if result["email"].Effect != ColEffectMask {
		t.Errorf("MASK > SHOW，期望 MASK，got: %s", result["email"].Effect)
	}
}

// ---- 脱敏处理测试 ----

func TestMaskPhone(t *testing.T) {
	got := maskPhone("13812345678")
	expected := "138****5678"
	if got != expected {
		t.Errorf("手机脱敏期望 %s，got: %s", expected, got)
	}
}

func TestMaskEmail(t *testing.T) {
	got := maskEmail("user@example.com")
	if got != "***@example.com" {
		t.Errorf("邮箱脱敏期望 ***@example.com，got: %s", got)
	}
}

func TestMaskIDCard(t *testing.T) {
	got := maskIDCard("110101199001011234")
	expected := "110101********1234"
	if got != expected {
		t.Errorf("身份证脱敏期望 %s，got: %s", expected, got)
	}
}

func TestApplyColPolicy_Hide(t *testing.T) {
	data := map[string]interface{}{
		"name":  "张三",
		"phone": "13812345678",
	}
	effects := map[string]PolicyColEntry{
		"phone": {FieldName: "phone", Effect: ColEffectHide},
	}
	result := ApplyColPolicy(data, effects)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("返回类型应为 map")
	}
	if m["phone"] != nil {
		t.Errorf("HIDE 后 phone 应为 nil，got: %v", m["phone"])
	}
	if m["name"] != "张三" {
		t.Error("name 字段不应被修改")
	}
}

func TestApplyColPolicy_Mask(t *testing.T) {
	data := map[string]interface{}{
		"phone": "13812345678",
	}
	effects := map[string]PolicyColEntry{
		"phone": {FieldName: "phone", Effect: ColEffectMask, MaskType: "phone"},
	}
	result := ApplyColPolicy(data, effects)
	m := result.(map[string]interface{})
	if m["phone"] != "138****5678" {
		t.Errorf("MASK phone 期望 138****5678，got: %v", m["phone"])
	}
}
