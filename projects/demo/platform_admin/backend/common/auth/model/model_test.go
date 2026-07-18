package model

import (
	"strings"
	"testing"
)

// TestAllTableNames 验证所有模型的 TableName() 返回值以 admin_ 开头
func TestAllTableNames(t *testing.T) {
	// 定义所有模型及其期望表名
	tests := []struct {
		name      string
		tableName string
	}{
		{"Tenant", Tenant{}.TableName()},
		{"User", User{}.TableName()},
		{"UserTenant", UserTenant{}.TableName()},
		{"Role", Role{}.TableName()},
		{"UserRole", UserRole{}.TableName()},
		{"Application", Application{}.TableName()},
		{"TenantApp", TenantApp{}.TableName()},
		{"RoleApp", RoleApp{}.TableName()},
		{"Resource", Resource{}.TableName()},
		{"RoleResource", RoleResource{}.TableName()},
		{"ApiPermission", ApiPermission{}.TableName()},
		{"RoleApi", RoleApi{}.TableName()},
		{"DataScopeConfig", DataScopeConfig{}.TableName()},
		{"DataScope", DataScope{}.TableName()},
		{"OperationLog", OperationLog{}.TableName()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.HasPrefix(tt.tableName, "admin_") {
				t.Errorf("%s.TableName() = %q, 期望以 admin_ 开头", tt.name, tt.tableName)
			}
		})
	}
}

// TestTableNameUniqueness 验证所有表名不重复
func TestTableNameUniqueness(t *testing.T) {
	tableNames := []string{
		Tenant{}.TableName(),
		User{}.TableName(),
		UserTenant{}.TableName(),
		Role{}.TableName(),
		UserRole{}.TableName(),
		Application{}.TableName(),
		TenantApp{}.TableName(),
		RoleApp{}.TableName(),
		Resource{}.TableName(),
		RoleResource{}.TableName(),
		ApiPermission{}.TableName(),
		RoleApi{}.TableName(),
		DataScopeConfig{}.TableName(),
		DataScope{}.TableName(),
		OperationLog{}.TableName(),
	}

	seen := make(map[string]bool)
	for _, name := range tableNames {
		if seen[name] {
			t.Errorf("表名重复: %s", name)
		}
		seen[name] = true
	}
}
