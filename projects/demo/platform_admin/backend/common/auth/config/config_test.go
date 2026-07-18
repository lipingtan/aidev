package config

import (
	"testing"
	"time"
)

// TestDefaultConfig 验证默认配置值正确
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// 基本字段
	if !cfg.Enabled {
		t.Error("默认 Enabled 应为 true")
	}
	if cfg.AuthType != "jwt" {
		t.Errorf("默认 AuthType 应为 jwt，实际 %s", cfg.AuthType)
	}
	if cfg.CacheType != "local" {
		t.Errorf("默认 CacheType 应为 local，实际 %s", cfg.CacheType)
	}
	if cfg.IDStrategy != "snowflake" {
		t.Errorf("默认 IDStrategy 应为 snowflake，实际 %s", cfg.IDStrategy)
	}

	// JWT
	if cfg.JWT.Secret != "change-me-in-production" {
		t.Error("JWT Secret 默认值错误")
	}
	if cfg.JWT.PlatformTokenTTL != 168*time.Hour {
		t.Errorf("PlatformTokenTTL 应为 168h，实际 %v", cfg.JWT.PlatformTokenTTL)
	}
	if cfg.JWT.AccessTokenTTL != 2*time.Hour {
		t.Errorf("AccessTokenTTL 应为 2h，实际 %v", cfg.JWT.AccessTokenTTL)
	}
	if cfg.JWT.Issuer != "auth-rbac" {
		t.Errorf("Issuer 应为 auth-rbac，实际 %s", cfg.JWT.Issuer)
	}

	// Permission
	if !cfg.Permission.WildcardEnabled {
		t.Error("默认 WildcardEnabled 应为 true")
	}
	if cfg.Permission.SuperAdminRole != "SUPER_ADMIN" {
		t.Errorf("SuperAdminRole 应为 SUPER_ADMIN，实际 %s", cfg.Permission.SuperAdminRole)
	}
	if cfg.Permission.EnforcementStrategy != "annotation-first" {
		t.Errorf("EnforcementStrategy 应为 annotation-first，实际 %s", cfg.Permission.EnforcementStrategy)
	}

	// Role
	if !cfg.Role.HierarchyEnabled {
		t.Error("默认 HierarchyEnabled 应为 true")
	}
	if cfg.Role.MaxDepth != 5 {
		t.Errorf("MaxDepth 应为 5，实际 %d", cfg.Role.MaxDepth)
	}

	// DataScope
	if cfg.DataScope.Enabled {
		t.Error("默认 DataScope.Enabled 应为 false")
	}
	if !cfg.DataScope.AutoInject {
		t.Error("默认 DataScope.AutoInject 应为 true")
	}

	// APIDiscovery
	if !cfg.APIDiscovery.Enabled {
		t.Error("默认 APIDiscovery.Enabled 应为 true")
	}
	if cfg.APIDiscovery.AsyncThreshold != 500 {
		t.Errorf("AsyncThreshold 应为 500，实际 %d", cfg.APIDiscovery.AsyncThreshold)
	}

	// App
	if cfg.App.DefaultAppCode != "default" {
		t.Errorf("DefaultAppCode 应为 default，实际 %s", cfg.App.DefaultAppCode)
	}

	// Tenant
	if len(cfg.Tenant.TemplateRoles) != 3 {
		t.Errorf("TemplateRoles 应有 3 个，实际 %d", len(cfg.Tenant.TemplateRoles))
	}
	if cfg.Tenant.TemplateRoles[0].Code != "tenant_admin" {
		t.Errorf("第一个模板角色 Code 应为 tenant_admin，实际 %s", cfg.Tenant.TemplateRoles[0].Code)
	}
	if len(cfg.Tenant.DefaultApps) != 1 || cfg.Tenant.DefaultApps[0] != "default" {
		t.Error("DefaultApps 应为 [default]")
	}

	// Blacklist
	if cfg.Blacklist.StoreType != "local" {
		t.Errorf("Blacklist.StoreType 应为 local，实际 %s", cfg.Blacklist.StoreType)
	}

	// Redis
	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("Redis.Addr 应为 localhost:6379，实际 %s", cfg.Redis.Addr)
	}
	if cfg.Redis.Password != "" {
		t.Error("Redis.Password 默认应为空")
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("Redis.DB 应为 0，实际 %d", cfg.Redis.DB)
	}
}
