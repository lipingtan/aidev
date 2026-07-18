package engine

import "testing"

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name     string
		codes    []string
		required string
		want     bool
	}{
		// 精确匹配
		{name: "精确匹配-命中", codes: []string{"sys:user:add"}, required: "sys:user:add", want: true},
		{name: "精确匹配-未命中", codes: []string{"sys:user:add"}, required: "sys:user:edit", want: false},
		{name: "精确匹配-多码命中", codes: []string{"sys:user:add", "sys:user:edit"}, required: "sys:user:edit", want: true},

		// * 单层通配符
		{name: "*匹配单层-命中", codes: []string{"sys:user:*"}, required: "sys:user:add", want: true},
		{name: "*匹配单层-不跨层", codes: []string{"sys:user:*"}, required: "sys:user:detail:view", want: false},
		{name: "*在中间-命中", codes: []string{"sys:*:add"}, required: "sys:user:add", want: true},
		{name: "*在中间-不匹配", codes: []string{"sys:*:add"}, required: "sys:user:edit", want: false},
		{name: "*多个-命中", codes: []string{"sys:*:*"}, required: "sys:user:add", want: true},
		{name: "*多个-不跨层", codes: []string{"sys:*:*"}, required: "sys:user:detail:view", want: false},

		// ** 多层通配符
		{name: "**匹配多层-末尾", codes: []string{"sys:**"}, required: "sys:user:add", want: true},
		{name: "**匹配多层-深层", codes: []string{"sys:**"}, required: "sys:role:list", want: true},
		{name: "**匹配多层-四层", codes: []string{"sys:**"}, required: "sys:user:detail:view", want: true},
		{name: "**不匹配前缀不同", codes: []string{"sys:**"}, required: "app:user:add", want: false},
		{name: "**单独匹配一切", codes: []string{"**"}, required: "sys:user:add", want: true},
		{name: "**单独匹配单层", codes: []string{"**"}, required: "anything", want: true},

		// ** 在中间
		{name: "**在中间-命中", codes: []string{"sys:**:add"}, required: "sys:user:add", want: true},
		{name: "**在中间-多层命中", codes: []string{"sys:**:add"}, required: "sys:user:detail:add", want: true},
		{name: "**在中间-末尾不匹配", codes: []string{"sys:**:add"}, required: "sys:user:edit", want: false},

		// 边界场景
		{name: "空码列表", codes: []string{}, required: "sys:user:add", want: false},
		{name: "空串码被忽略", codes: []string{""}, required: "sys:user:add", want: false},
		{name: "单层码精确匹配", codes: []string{"dashboard"}, required: "dashboard", want: true},
		{name: "单层码不匹配", codes: []string{"dashboard"}, required: "settings", want: false},
		{name: "空格trim", codes: []string{" sys:user:add "}, required: "sys:user:add", want: true},

		// 混合模式
		{name: "精确+通配符-精确命中", codes: []string{"sys:user:add", "sys:role:*"}, required: "sys:user:add", want: true},
		{name: "精确+通配符-通配命中", codes: []string{"sys:user:add", "sys:role:*"}, required: "sys:role:list", want: true},
		{name: "精确+通配符-都不命中", codes: []string{"sys:user:add", "sys:role:*"}, required: "app:menu:view", want: false},

		// ** 匹配一切场景
		{name: "matchAll优先级", codes: []string{"**", "sys:user:add"}, required: "anything:at:all", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPermissionEngine(tt.codes)
			got := engine.HasPermission(tt.required)
			if got != tt.want {
				t.Errorf("HasPermission(%q) = %v, want %v (codes: %v)", tt.required, got, tt.want, tt.codes)
			}
		})
	}
}

func BenchmarkHasPermission(b *testing.B) {
	// 模拟真实场景：50个精确码 + 5个通配符
	codes := []string{
		"sys:user:add", "sys:user:edit", "sys:user:delete", "sys:user:list", "sys:user:view",
		"sys:role:add", "sys:role:edit", "sys:role:delete", "sys:role:list", "sys:role:view",
		"sys:menu:add", "sys:menu:edit", "sys:menu:delete", "sys:menu:list", "sys:menu:view",
		"sys:tenant:add", "sys:tenant:edit", "sys:tenant:delete", "sys:tenant:list", "sys:tenant:view",
		"sys:log:list", "sys:log:view", "sys:log:export",
		"app:order:add", "app:order:edit", "app:order:delete", "app:order:list", "app:order:view",
		"app:product:add", "app:product:edit", "app:product:delete", "app:product:list", "app:product:view",
		"app:customer:add", "app:customer:edit", "app:customer:delete", "app:customer:list", "app:customer:view",
		"app:report:daily", "app:report:weekly", "app:report:monthly", "app:report:export",
		"app:config:view", "app:config:edit",
		"app:notification:send", "app:notification:list", "app:notification:delete",
		"app:file:upload", "app:file:download", "app:file:delete", "app:file:list",
		// 通配符模式
		"monitor:*:view",
		"dashboard:**",
		"sys:audit:*",
		"app:stats:**",
		"plugin:*:config",
	}

	engine := NewPermissionEngine(codes)

	b.Run("精确命中", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine.HasPermission("sys:user:add")
		}
	})

	b.Run("精确未命中_走通配符", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine.HasPermission("monitor:cpu:view")
		}
	})

	b.Run("通配符命中_**", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine.HasPermission("dashboard:main:overview")
		}
	})

	b.Run("全部未命中", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine.HasPermission("unknown:module:action")
		}
	})
}
