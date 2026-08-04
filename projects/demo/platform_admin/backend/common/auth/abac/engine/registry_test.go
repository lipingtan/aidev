package engine

import (
	"sync"
	"testing"
)

func TestRegisterResource(t *testing.T) {
	RegisterResource(&ResourceDef{
		Type:        "order_test",
		DisplayName: "测试订单",
		MainTable:   "orders_test",
		Attributes: []ResourceAttrDef{
			{AttrName: "dept_id", Display: "部门 ID", QualifiedCol: "orders_test.dept_id", DataType: "int"},
		},
	})

	def, ok := GetResource("order_test")
	if !ok {
		t.Fatal("RegisterResource 后应能查到资源")
	}
	if def.DisplayName != "测试订单" {
		t.Errorf("DisplayName 不符，got %s", def.DisplayName)
	}

	// 反查
	byTable := GetResourceByTable("orders_test")
	if byTable == nil {
		t.Fatal("GetResourceByTable 应能按 MainTable 反查到资源")
	}

	// GetAttr
	attr, found := def.GetAttr("dept_id")
	if !found || attr.QualifiedCol != "orders_test.dept_id" {
		t.Errorf("GetAttr 未返回正确属性: %+v", attr)
	}
}

func TestBuiltinSubjectAttrs(t *testing.T) {
	builtins := []string{"user_id", "role_ids", "dept_ids", "tenant_id"}
	for _, name := range builtins {
		if _, ok := GetSubjectAttr(name); !ok {
			t.Errorf("内置主体属性 %s 未注册", name)
		}
	}
}

func TestRegisterSubjectAttr(t *testing.T) {
	RegisterSubjectAttr(SubjectAttrDef{
		AttrName: "custom_level",
		Display:  "自定义等级",
		DataType: "int",
		Resolver: func(a map[string]interface{}) interface{} { return a["custom_level"] },
	})
	if _, ok := GetSubjectAttr("custom_level"); !ok {
		t.Fatal("自定义主体属性注册后应能查到")
	}
}

func TestRegisterResource_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "concurrent_res"
			RegisterResource(&ResourceDef{Type: key, MainTable: "concurrent_tbl"})
			GetResource(key)
			GetResourceByTable("concurrent_tbl")
		}(i)
	}
	wg.Wait()
}
