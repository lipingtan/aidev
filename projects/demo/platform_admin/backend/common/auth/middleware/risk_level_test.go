package middleware

import "testing"

// TestGetRiskLevel_HighRiskPaths 高风险路径返回 HIGH
func TestGetRiskLevel_HighRiskPaths(t *testing.T) {
	tests := []struct {
		method   string
		fullPath string
	}{
		{"DELETE", "/api/v1/admin/tenants/:id"},
		{"PUT", "/api/v1/admin/users/:id/status"},
		{"PUT", "/api/v1/admin/users/:id/reset-pwd"},
		{"DELETE", "/api/v1/admin/roles/:id"},
		{"PUT", "/api/v1/admin/roles/:id/resources"},
		{"PUT", "/api/v1/admin/roles/:id/apis"},
		{"DELETE", "/api/v1/admin/applications/:id"},
	}

	for _, tt := range tests {
		t.Run(tt.method+":"+tt.fullPath, func(t *testing.T) {
			level := GetRiskLevel(tt.method, tt.fullPath)
			if level != "HIGH" {
				t.Errorf("期望 HIGH，实际 %s", level)
			}
		})
	}
}

// TestGetRiskLevel_WriteLow 非白名单写操作返回 LOW
func TestGetRiskLevel_WriteLow(t *testing.T) {
	tests := []struct {
		method   string
		fullPath string
	}{
		{"POST", "/api/v1/admin/tenants"},
		{"PUT", "/api/v1/admin/tenants/:id"},
		{"POST", "/api/v1/admin/users"},
		{"PUT", "/api/v1/admin/users/:id"},
		{"POST", "/api/v1/admin/roles"},
	}

	for _, tt := range tests {
		t.Run(tt.method+":"+tt.fullPath, func(t *testing.T) {
			level := GetRiskLevel(tt.method, tt.fullPath)
			if level != "LOW" {
				t.Errorf("期望 LOW，实际 %s", level)
			}
		})
	}
}

// TestGetRiskLevel_GetEmpty GET 请求返回空字符串（不记录日志）
func TestGetRiskLevel_GetEmpty(t *testing.T) {
	tests := []string{
		"/api/v1/admin/users",
		"/api/v1/admin/roles",
		"/api/v1/admin/tenants/:id",
	}

	for _, path := range tests {
		t.Run("GET:"+path, func(t *testing.T) {
			level := GetRiskLevel("GET", path)
			if level != "" {
				t.Errorf("GET 请求期望返回空，实际 %s", level)
			}
		})
	}
}
