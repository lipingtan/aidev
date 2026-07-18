package cache

import (
	"testing"
	"time"
)

// ========== LocalCache 基本操作测试 ==========

func TestLocalCache_SetAndGet(t *testing.T) {
	lc := NewLocalCache()
	defer lc.Stop()

	lc.Set("key1", "value1", 5*time.Second)
	val, ok := lc.Get("key1")
	if !ok {
		t.Fatal("期望获取到 key1")
	}
	if val != "value1" {
		t.Fatalf("期望 value1，得到 %v", val)
	}
}

func TestLocalCache_GetMiss(t *testing.T) {
	lc := NewLocalCache()
	defer lc.Stop()

	_, ok := lc.Get("nonexistent")
	if ok {
		t.Fatal("期望未命中")
	}
}

func TestLocalCache_Delete(t *testing.T) {
	lc := NewLocalCache()
	defer lc.Stop()

	lc.Set("key1", "value1", 5*time.Second)
	lc.Delete("key1")
	_, ok := lc.Get("key1")
	if ok {
		t.Fatal("期望删除后未命中")
	}
}

func TestLocalCache_DeleteByPrefix(t *testing.T) {
	lc := NewLocalCache()
	defer lc.Stop()

	lc.Set("role:1", "perms1", 5*time.Second)
	lc.Set("role:2", "perms2", 5*time.Second)
	lc.Set("user:1", "data1", 5*time.Second)

	lc.DeleteByPrefix("role:")

	_, ok1 := lc.Get("role:1")
	_, ok2 := lc.Get("role:2")
	_, ok3 := lc.Get("user:1")

	if ok1 || ok2 {
		t.Fatal("期望 role: 前缀的 key 被删除")
	}
	if !ok3 {
		t.Fatal("期望 user:1 仍存在")
	}
}

// ========== TTL 过期测试 ==========

func TestLocalCache_TTLExpiry(t *testing.T) {
	lc := NewLocalCache()
	defer lc.Stop()

	lc.Set("expire_key", "data", 100*time.Millisecond)

	// 立即获取应命中
	_, ok := lc.Get("expire_key")
	if !ok {
		t.Fatal("期望 TTL 内命中")
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)
	_, ok = lc.Get("expire_key")
	if ok {
		t.Fatal("期望 TTL 过期后未命中")
	}
}

// ========== TwoLevelCache L1/L2 逻辑测试 ==========

func TestTwoLevelCache_L2MissThenL1Hit(t *testing.T) {
	l1 := NewLocalCache()
	l2 := NewLocalCache()
	defer l1.Stop()
	defer l2.Stop()

	tc := NewTwoLevelCache(l1, l2)
	pc := NewPermissionCache(tc, 5*time.Minute, 2*time.Minute)

	// 设置 L1 角色权限
	pc.SetRolePermissions("100", []interface{}{"perm:read", "perm:write"})

	// L2 未命中，应从 L1 合并
	result, ok := pc.GetUserPermissions("1", "10", "default", []string{"100"})
	if !ok {
		t.Fatal("期望从 L1 获取成功")
	}
	perms := result.([]interface{})
	if len(perms) != 2 {
		t.Fatalf("期望 2 个权限，得到 %d", len(perms))
	}

	// 再次获取应命中 L2
	result2, ok := pc.GetUserPermissions("1", "10", "default", []string{"100"})
	if !ok {
		t.Fatal("期望从 L2 命中")
	}
	perms2 := result2.([]interface{})
	if len(perms2) != 2 {
		t.Fatalf("期望 L2 命中 2 个权限，得到 %d", len(perms2))
	}
}

// ========== 失效测试 ==========

func TestPermissionCache_InvalidateRole(t *testing.T) {
	l1 := NewLocalCache()
	l2 := NewLocalCache()
	defer l1.Stop()
	defer l2.Stop()

	tc := NewTwoLevelCache(l1, l2)
	pc := NewPermissionCache(tc, 5*time.Minute, 2*time.Minute)

	// 设置 L1 并触发 L2 写入
	pc.SetRolePermissions("200", []interface{}{"perm:admin"})
	pc.GetUserPermissions("5", "20", "app1", []string{"200"})

	// 验证 L1 存在
	_, ok := l1.Get(L1Key("200"))
	if !ok {
		t.Fatal("期望 L1 中有 role:200")
	}

	// 失效角色
	pc.InvalidateRole("200")

	// L1 应被清除
	_, ok = l1.Get(L1Key("200"))
	if ok {
		t.Fatal("期望 InvalidateRole 后 L1 清除")
	}

	// 等待异步 L2 清除完成
	time.Sleep(300 * time.Millisecond)

	// L2 应被清除
	_, ok = l2.Get(L2Key("5", "20", "app1"))
	if ok {
		t.Fatal("期望 InvalidateRole 后 L2 清除")
	}
}

func TestPermissionCache_InvalidateUserPermissions(t *testing.T) {
	l1 := NewLocalCache()
	l2 := NewLocalCache()
	defer l1.Stop()
	defer l2.Stop()

	tc := NewTwoLevelCache(l1, l2)
	pc := NewPermissionCache(tc, 5*time.Minute, 2*time.Minute)

	// 设置 L2
	l2.Set(L2Key("10", "30", "default"), []interface{}{"perm:x"}, 5*time.Minute)
	l2.Set(L2Key("10", "30", "app2"), []interface{}{"perm:y"}, 5*time.Minute)
	l2.Set(L2Key("99", "30", "default"), []interface{}{"perm:z"}, 5*time.Minute)

	// 清除 user:10 tenant:30 的所有缓存
	pc.InvalidateUserPermissions("10", "30")

	_, ok1 := l2.Get(L2Key("10", "30", "default"))
	_, ok2 := l2.Get(L2Key("10", "30", "app2"))
	_, ok3 := l2.Get(L2Key("99", "30", "default"))

	if ok1 || ok2 {
		t.Fatal("期望 user:10:tenant:30 的缓存被清除")
	}
	if !ok3 {
		t.Fatal("期望 user:99 的缓存不受影响")
	}
}
