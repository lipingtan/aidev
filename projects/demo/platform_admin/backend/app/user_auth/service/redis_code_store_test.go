package service

import (
	"fmt"
	"testing"
	"time"
)

// mockRedisAdapter 用于测试的 mock adapter
type mockRedisAdapter struct {
	data map[string]string
}

func newMockRedisAdapter() *mockRedisAdapter {
	return &mockRedisAdapter{data: make(map[string]string)}
}

func (m *mockRedisAdapter) String() string { return "mock" }
func (m *mockRedisAdapter) Get(key string) (string, error) {
	v, ok := m.data[key]
	if !ok {
		return "", fmt.Errorf("not found")
	}
	return v, nil
}
func (m *mockRedisAdapter) Set(key string, val interface{}, _ int) error {
	m.data[key] = fmt.Sprintf("%v", val)
	return nil
}
func (m *mockRedisAdapter) Del(key string) error {
	delete(m.data, key)
	return nil
}
func (m *mockRedisAdapter) HashGet(_, _ string) (string, error) { return "", nil }
func (m *mockRedisAdapter) HashDel(_, _ string) error           { return nil }
func (m *mockRedisAdapter) Increase(_ string) error             { return nil }
func (m *mockRedisAdapter) Decrease(_ string) error             { return nil }
func (m *mockRedisAdapter) Expire(_ string, _ time.Duration) error { return nil }

// ---- RedisCodeStore 测试 ----

func TestRedisCodeStore_SetAndGet(t *testing.T) {
	adapter := newMockRedisAdapter()
	store := NewRedisCodeStore(adapter)

	if err := store.Set("code:13800:tenant1", "1234", time.Minute); err != nil {
		t.Fatalf("Set 不应报错: %v", err)
	}

	val, err := store.Get("code:13800:tenant1")
	if err != nil {
		t.Fatalf("Get 不应报错: %v", err)
	}
	if val != "1234" {
		t.Errorf("期望 '1234'，got: %s", val)
	}
}

func TestRedisCodeStore_GetNotFound(t *testing.T) {
	adapter := newMockRedisAdapter()
	store := NewRedisCodeStore(adapter)

	_, err := store.Get("nonexistent-key")
	if err == nil {
		t.Error("不存在的 key 应返回错误")
	}
}

func TestRedisCodeStore_Delete(t *testing.T) {
	adapter := newMockRedisAdapter()
	store := NewRedisCodeStore(adapter)

	_ = store.Set("code:test", "5678", time.Minute)
	_ = store.Delete("code:test")

	_, err := store.Get("code:test")
	if err == nil {
		t.Error("Delete 后 Get 应返回错误")
	}
}

func TestRedisCodeStore_Exists(t *testing.T) {
	adapter := newMockRedisAdapter()
	store := NewRedisCodeStore(adapter)

	if store.Exists("key-not-set") {
		t.Error("未设置时 Exists 应返回 false")
	}

	_ = store.Set("key-set", "value", time.Minute)
	if !store.Exists("key-set") {
		t.Error("设置后 Exists 应返回 true")
	}
}

func TestRedisCodeStore_NilAdapter(t *testing.T) {
	store := NewRedisCodeStore(nil)

	err := store.Set("k", "v", time.Minute)
	if err == nil {
		t.Error("nil adapter 时 Set 应返回错误")
	}

	_, err = store.Get("k")
	if err == nil {
		t.Error("nil adapter 时 Get 应返回错误")
	}

	if store.Exists("k") {
		t.Error("nil adapter 时 Exists 应返回 false")
	}
}
