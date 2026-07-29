package cache

import (
	"sync"
	"testing"
	"time"
)

// --- mock AdapterCache ---

type mockCacheAdapter struct {
	mu    sync.RWMutex
	store map[string]string
}

func newMockAdapter() *mockCacheAdapter {
	return &mockCacheAdapter{store: make(map[string]string)}
}

func (m *mockCacheAdapter) String() string { return "mock" }

func (m *mockCacheAdapter) Get(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.store[key]
	if !ok {
		return "", nil
	}
	return v, nil
}

func (m *mockCacheAdapter) Set(key string, val interface{}, expire int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key] = val.(string)
	return nil
}

func (m *mockCacheAdapter) Del(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, key)
	return nil
}

func (m *mockCacheAdapter) HashGet(hk, key string) (string, error) { return "", nil }
func (m *mockCacheAdapter) HashDel(hk, key string) error           { return nil }
func (m *mockCacheAdapter) Increase(key string) error              { return nil }
func (m *mockCacheAdapter) Decrease(key string) error              { return nil }
func (m *mockCacheAdapter) Expire(key string, dur time.Duration) error { return nil }

// --- 测试 ---

// TestPermCodeCache_Get_Miss 缓存未命中返回 nil, false
func TestPermCodeCache_Get_Miss(t *testing.T) {
	c := NewPermCodeCache(newMockAdapter(), 10*time.Minute)

	codes, ok := c.Get(1, 100)
	if ok {
		t.Fatal("期望缓存未命中，实际命中")
	}
	if codes != nil {
		t.Fatalf("期望 nil，实际 %v", codes)
	}
}

// TestPermCodeCache_SetThenGet 写入后可正常命中
func TestPermCodeCache_SetThenGet(t *testing.T) {
	c := NewPermCodeCache(newMockAdapter(), 10*time.Minute)

	expected := []string{"user:list", "user:create", "role:delete"}
	c.Set(1, 100, expected)

	codes, ok := c.Get(1, 100)
	if !ok {
		t.Fatal("期望缓存命中，实际未命中")
	}
	if len(codes) != len(expected) {
		t.Fatalf("期望 %d 个权限码，实际 %d 个", len(expected), len(codes))
	}
	for i, code := range expected {
		if codes[i] != code {
			t.Fatalf("第 %d 个权限码期望 %s，实际 %s", i, code, codes[i])
		}
	}
}

// TestPermCodeCache_Invalidate 失效后不可命中
func TestPermCodeCache_Invalidate_CacheCleared(t *testing.T) {
	c := NewPermCodeCache(newMockAdapter(), 10*time.Minute)

	c.Set(1, 100, []string{"user:list"})

	// 验证写入成功
	_, ok := c.Get(1, 100)
	if !ok {
		t.Fatal("前置条件：期望缓存命中")
	}

	// 失效
	c.Invalidate(1, 100)

	// 验证已清除
	_, ok = c.Get(1, 100)
	if ok {
		t.Fatal("失效后期望缓存未命中，实际仍命中")
	}
}

// TestPermCodeCache_NilAdapter_GetReturnsFalse adapter 为 nil 时 Get 不 panic 返回 false
func TestPermCodeCache_NilAdapter_GetReturnsFalse(t *testing.T) {
	c := NewPermCodeCache(nil, 10*time.Minute)

	codes, ok := c.Get(1, 100)
	if ok {
		t.Fatal("nil adapter 期望返回 false")
	}
	if codes != nil {
		t.Fatalf("nil adapter 期望返回 nil，实际 %v", codes)
	}
}

// TestPermCodeCache_NilAdapter_SetNoPanic adapter 为 nil 时 Set 不 panic
func TestPermCodeCache_NilAdapter_SetNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil adapter Set 触发 panic: %v", r)
		}
	}()
	c := NewPermCodeCache(nil, 10*time.Minute)
	c.Set(1, 100, []string{"user:list"})
}

// TestPermCodeCache_NilAdapter_InvalidateNoPanic adapter 为 nil 时 Invalidate 不 panic
func TestPermCodeCache_NilAdapter_InvalidateNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil adapter Invalidate 触发 panic: %v", r)
		}
	}()
	c := NewPermCodeCache(nil, 10*time.Minute)
	c.Invalidate(1, 100)
}

// TestPermCodeCache_TenantIsolation 不同 tenant 的 key 互不影响
func TestPermCodeCache_TenantIsolation(t *testing.T) {
	c := NewPermCodeCache(newMockAdapter(), 10*time.Minute)

	c.Set(1, 100, []string{"tenant1:perm"})
	c.Set(2, 100, []string{"tenant2:perm"})

	codes1, ok1 := c.Get(1, 100)
	codes2, ok2 := c.Get(2, 100)

	if !ok1 || !ok2 {
		t.Fatal("两个租户的缓存均应命中")
	}
	if codes1[0] != "tenant1:perm" {
		t.Fatalf("租户1 权限码期望 tenant1:perm，实际 %s", codes1[0])
	}
	if codes2[0] != "tenant2:perm" {
		t.Fatalf("租户2 权限码期望 tenant2:perm，实际 %s", codes2[0])
	}
}

// TestPermCodeCache_EmptyList 空权限码列表可以正常存取
func TestPermCodeCache_EmptyList(t *testing.T) {
	c := NewPermCodeCache(newMockAdapter(), 10*time.Minute)

	c.Set(1, 100, []string{})
	codes, ok := c.Get(1, 100)
	if !ok {
		t.Fatal("期望命中（空列表也是有效缓存）")
	}
	if len(codes) != 0 {
		t.Fatalf("期望空列表，实际 %v", codes)
	}
}
