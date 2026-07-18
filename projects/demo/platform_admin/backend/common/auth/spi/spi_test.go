package spi

import (
	"context"
	"testing"
	"time"
)

// TestNoOpUserProvider 验证 NoOpUserProvider 满足 UserProvider 接口
func TestNoOpUserProvider(t *testing.T) {
	var p UserProvider = &NoOpUserProvider{}
	user, err := p.LoadByUsername("test")
	if err != nil {
		t.Fatalf("期望 nil error，得到 %v", err)
	}
	if user != nil {
		t.Fatalf("期望 nil user，得到 %v", user)
	}
}

// TestNoOpRoleProvider 验证 NoOpRoleProvider 满足 RoleProvider 接口
func TestNoOpRoleProvider(t *testing.T) {
	var p RoleProvider = &NoOpRoleProvider{}
	roles, err := p.GetRolesByUser(1, 1)
	if err != nil {
		t.Fatalf("期望 nil error，得到 %v", err)
	}
	if roles != nil {
		t.Fatalf("期望 nil roles，得到 %v", roles)
	}
}

// TestNoOpCacheAdapter 验证 NoOpCacheAdapter 满足 CacheAdapter 接口
func TestNoOpCacheAdapter(t *testing.T) {
	var c CacheAdapter = &NoOpCacheAdapter{}

	val, ok := c.Get("key")
	if ok {
		t.Fatal("期望 ok=false")
	}
	if val != nil {
		t.Fatalf("期望 nil value，得到 %v", val)
	}

	// 以下调用不应 panic
	c.Set("key", "value", time.Minute)
	c.Delete("key1", "key2")
	c.DeleteByPrefix("prefix:")
}

// TestNoOpTokenBlacklistStore 验证 NoOpTokenBlacklistStore 满足 TokenBlacklistStore 接口
func TestNoOpTokenBlacklistStore(t *testing.T) {
	var s TokenBlacklistStore = &NoOpTokenBlacklistStore{}

	err := s.Add("token-abc", time.Hour)
	if err != nil {
		t.Fatalf("期望 nil error，得到 %v", err)
	}
	if s.Contains("token-abc") {
		t.Fatal("期望 Contains 返回 false")
	}
}

// TestNoOpOrganizationProvider 验证 NoOpOrganizationProvider 满足 OrganizationProvider 接口
func TestNoOpOrganizationProvider(t *testing.T) {
	var p OrganizationProvider = &NoOpOrganizationProvider{}

	ids, err := p.GetOrgIds(1, 1)
	if err != nil {
		t.Fatalf("GetOrgIds: 期望 nil error，得到 %v", err)
	}
	if ids == nil || len(ids) != 0 {
		t.Fatalf("GetOrgIds: 期望空切片，得到 %v", ids)
	}

	path, err := p.GetOrgPath(1, 1)
	if err != nil {
		t.Fatalf("GetOrgPath: 期望 nil error，得到 %v", err)
	}
	if path == nil || len(path) != 0 {
		t.Fatalf("GetOrgPath: 期望空切片，得到 %v", path)
	}

	subIds, err := p.GetSubOrgIds(1, 1)
	if err != nil {
		t.Fatalf("GetSubOrgIds: 期望 nil error，得到 %v", err)
	}
	if subIds == nil || len(subIds) != 0 {
		t.Fatalf("GetSubOrgIds: 期望空切片，得到 %v", subIds)
	}
}

// TestNoOpDataScopeHandler 验证 NoOpDataScopeHandler 满足 DataScopeHandler 接口
func TestNoOpDataScopeHandler(t *testing.T) {
	var h DataScopeHandler = &NoOpDataScopeHandler{}
	result, err := h.Handle(context.Background(), "org", 1)
	if err != nil {
		t.Fatalf("期望 nil error，得到 %v", err)
	}
	if result != nil {
		t.Fatalf("期望 nil result，得到 %v", result)
	}
}

// TestNoOpApiDiscoveryStrategy 验证 NoOpApiDiscoveryStrategy 满足 ApiDiscoveryStrategy 接口
func TestNoOpApiDiscoveryStrategy(t *testing.T) {
	var s ApiDiscoveryStrategy = &NoOpApiDiscoveryStrategy{}
	endpoints, err := s.Discover(nil)
	if err != nil {
		t.Fatalf("期望 nil error，得到 %v", err)
	}
	if endpoints != nil {
		t.Fatalf("期望 nil endpoints，得到 %v", endpoints)
	}
}

// TestNoOpEventPublisher 验证 NoOpEventPublisher 满足 EventPublisher 接口
func TestNoOpEventPublisher(t *testing.T) {
	var p EventPublisher = &NoOpEventPublisher{}
	err := p.Publish("user.created", map[string]interface{}{"id": 1})
	if err != nil {
		t.Fatalf("期望 nil error，得到 %v", err)
	}
}

// TestNoOpOperationLogger 验证 NoOpOperationLogger 满足 OperationLogger 接口
func TestNoOpOperationLogger(t *testing.T) {
	var l OperationLogger = &NoOpOperationLogger{}
	err := l.Log(context.Background(), OperationLogEntry{
		Module:     "user",
		Action:     "create",
		TargetType: "user",
		TargetID:   1,
		Summary:    "创建用户",
	})
	if err != nil {
		t.Fatalf("期望 nil error，得到 %v", err)
	}
}

// TestMockReplacement 验证接口可被自定义 mock 替换
func TestMockReplacement(t *testing.T) {
	mock := &mockUserProvider{
		user: &AuthUser{ID: 42, Username: "admin", Status: 1},
	}
	var p UserProvider = mock
	user, err := p.LoadByUsername("admin")
	if err != nil {
		t.Fatalf("期望 nil error，得到 %v", err)
	}
	if user == nil || user.ID != 42 {
		t.Fatalf("期望 ID=42，得到 %v", user)
	}
}

// mockUserProvider 用于验证接口可被 mock 替换
type mockUserProvider struct {
	user *AuthUser
}

func (m *mockUserProvider) LoadByUsername(_ string) (*AuthUser, error) {
	return m.user, nil
}
