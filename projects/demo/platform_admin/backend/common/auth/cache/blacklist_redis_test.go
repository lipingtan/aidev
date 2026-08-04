package cache

import (
	"fmt"
	"testing"
	"time"
)

// mockRedisBlacklistAdapter 用于测试的 mock adapter（避免与同包其他测试文件冲突）
type mockRedisBlacklistAdapter struct {
	data map[string]string
}

func newMockBlacklistCache() *mockRedisBlacklistAdapter {
	return &mockRedisBlacklistAdapter{data: make(map[string]string)}
}

func (m *mockRedisBlacklistAdapter) String() string { return "mock" }

func (m *mockRedisBlacklistAdapter) Get(key string) (string, error) {
	v, ok := m.data[key]
	if !ok {
		return "", fmt.Errorf("not found")
	}
	return v, nil
}

func (m *mockRedisBlacklistAdapter) Set(key string, val interface{}, _ int) error {
	m.data[key] = fmt.Sprintf("%v", val)
	return nil
}

func (m *mockRedisBlacklistAdapter) Del(key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockRedisBlacklistAdapter) HashGet(_, _ string) (string, error) { return "", nil }
func (m *mockRedisBlacklistAdapter) HashDel(_, _ string) error           { return nil }
func (m *mockRedisBlacklistAdapter) Increase(_ string) error             { return nil }
func (m *mockRedisBlacklistAdapter) Decrease(_ string) error             { return nil }
func (m *mockRedisBlacklistAdapter) Expire(_ string, _ time.Duration) error { return nil }

// ---- RedisBlacklistStore 测试 ----

func TestRedisBlacklist_AddAndContains(t *testing.T) {
	adapter := newMockBlacklistCache()
	store := NewRedisBlacklistStore(adapter)

	// 未添加时不在黑名单
	if store.Contains("jti-abc") {
		t.Error("未添加时 Contains 应返回 false")
	}

	// 添加后在黑名单
	if err := store.Add("jti-abc", time.Hour); err != nil {
		t.Fatalf("Add 不应报错: %v", err)
	}
	if !store.Contains("jti-abc") {
		t.Error("Add 后 Contains 应返回 true")
	}
}

func TestRedisBlacklist_DifferentKeys(t *testing.T) {
	adapter := newMockBlacklistCache()
	store := NewRedisBlacklistStore(adapter)

	_ = store.Add("jti-1", time.Hour)

	if store.Contains("jti-2") {
		t.Error("jti-2 未添加，Contains 应返回 false")
	}
}

func TestRedisBlacklist_NilAdapter(t *testing.T) {
	store := NewRedisBlacklistStore(nil)

	// nil adapter 时所有操作均静默处理
	if err := store.Add("jti-x", time.Hour); err != nil {
		t.Errorf("nil adapter 时 Add 应返回 nil，got: %v", err)
	}
	if store.Contains("jti-x") {
		t.Error("nil adapter 时 Contains 应返回 false")
	}
}

func TestRedisBlacklist_KeyFormat(t *testing.T) {
	adapter := newMockBlacklistCache()
	store := NewRedisBlacklistStore(adapter)

	_ = store.Add("my-jti", time.Hour)

	// 验证 key 格式为 blacklist:{jti}
	if _, ok := adapter.data["blacklist:my-jti"]; !ok {
		t.Error("key 格式应为 blacklist:{jti}")
	}
}
