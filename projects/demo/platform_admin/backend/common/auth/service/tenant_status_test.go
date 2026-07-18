package service

import (
	"fmt"
	"testing"
	"time"

	"go-admin/common/auth/cache"
	"go-admin/common/auth/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建内存 SQLite 用于测试
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	// 建表
	if err := db.AutoMigrate(&model.Tenant{}); err != nil {
		t.Fatalf("迁移 Tenant 表失败: %v", err)
	}
	return db
}

// createTestTenant 插入测试租户
func createTestTenant(t *testing.T, db *gorm.DB, id int64, status int) {
	t.Helper()
	tenant := &model.Tenant{
		ID:         id,
		TenantCode: "test_tenant_" + fmt.Sprintf("%d", id),
		Name:       "测试租户",
		Status:     status,
		Version:    1,
	}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("创建测试租户失败: %v", err)
	}
	// status=0 时 GORM 可能忽略零值，需强制更新
	if status == 0 {
		if err := db.Model(&model.Tenant{}).Where("id = ?", id).Update("status", 0).Error; err != nil {
			t.Fatalf("更新租户状态失败: %v", err)
		}
	}
}

func TestIsTenantActive_Enabled(t *testing.T) {
	db := setupTestDB(t)
	createTestTenant(t, db, 1001, 1) // status=1 启用

	lc := cache.NewLocalCache()
	defer lc.Stop()

	svc := NewTenantStatusService(db, lc)
	if !svc.IsTenantActive(1001) {
		t.Fatal("期望启用的租户返回 true")
	}
}

func TestIsTenantActive_Disabled(t *testing.T) {
	db := setupTestDB(t)
	createTestTenant(t, db, 1002, 0) // status=0 禁用

	lc := cache.NewLocalCache()
	defer lc.Stop()

	svc := NewTenantStatusService(db, lc)
	if svc.IsTenantActive(1002) {
		t.Fatal("期望禁用的租户返回 false")
	}
}

func TestIsTenantActive_NotFound(t *testing.T) {
	db := setupTestDB(t)

	lc := cache.NewLocalCache()
	defer lc.Stop()

	svc := NewTenantStatusService(db, lc)
	if svc.IsTenantActive(9999) {
		t.Fatal("期望不存在的租户返回 false")
	}
}

func TestIsTenantActive_CacheHit(t *testing.T) {
	db := setupTestDB(t)
	createTestTenant(t, db, 1003, 1)

	lc := cache.NewLocalCache()
	defer lc.Stop()

	svc := NewTenantStatusService(db, lc)

	// 首次查询，写入缓存
	svc.IsTenantActive(1003)

	// 修改 DB 状态为禁用（模拟直接 DB 变更）
	db.Model(&model.Tenant{}).Where("id = ?", 1003).Update("status", 0)

	// 第二次查询应命中缓存，仍返回 true（未过期）
	if !svc.IsTenantActive(1003) {
		t.Fatal("期望缓存命中，重复调用不重复查 DB，仍返回 true")
	}
}

func TestOnTenantDisabled_ClearsL2Cache(t *testing.T) {
	db := setupTestDB(t)

	statusCache := cache.NewLocalCache()
	defer statusCache.Stop()

	l2Cache := cache.NewLocalCache()
	defer l2Cache.Stop()

	// 模拟 L2 缓存中有该租户的数据
	// L2 key 格式: user:{uid}:tenant:{tid}:app:{app}
	l2Cache.Set("user:1:tenant:100:app:default", "perms1", 5*time.Minute)
	l2Cache.Set("user:2:tenant:100:app:app2", "perms2", 5*time.Minute)
	l2Cache.Set("user:3:tenant:200:app:default", "perms3", 5*time.Minute) // 其他租户

	svc := NewTenantStatusService(db, statusCache)
	svc.OnTenantDisabled(100, l2Cache)

	// 该租户的 L2 缓存应被清除
	if _, ok := l2Cache.Get("user:1:tenant:100:app:default"); ok {
		t.Fatal("期望 tenant:100 的 L2 缓存被清除")
	}
	if _, ok := l2Cache.Get("user:2:tenant:100:app:app2"); ok {
		t.Fatal("期望 tenant:100 的 L2 缓存被清除")
	}
	// 其他租户不受影响
	if _, ok := l2Cache.Get("user:3:tenant:200:app:default"); !ok {
		t.Fatal("期望 tenant:200 的 L2 缓存不受影响")
	}
}

func TestOnTenantDisabled_StatusCacheSetToFalse(t *testing.T) {
	db := setupTestDB(t)
	createTestTenant(t, db, 1004, 1)

	statusCache := cache.NewLocalCache()
	defer statusCache.Stop()

	l2Cache := cache.NewLocalCache()
	defer l2Cache.Stop()

	svc := NewTenantStatusService(db, statusCache)

	// 先查询一次，缓存为 true
	if !svc.IsTenantActive(1004) {
		t.Fatal("期望首次查询返回 true")
	}

	// 禁用
	svc.OnTenantDisabled(1004, l2Cache)

	// 禁用后查询应返回 false（缓存中已设为 false）
	if svc.IsTenantActive(1004) {
		t.Fatal("期望 OnTenantDisabled 后返回 false")
	}
}

func TestOnTenantEnabled_ClearsStatusCache(t *testing.T) {
	db := setupTestDB(t)
	createTestTenant(t, db, 1005, 0) // 初始禁用

	statusCache := cache.NewLocalCache()
	defer statusCache.Stop()

	svc := NewTenantStatusService(db, statusCache)

	// 首次查询，缓存为 false
	if svc.IsTenantActive(1005) {
		t.Fatal("期望禁用租户返回 false")
	}

	// DB 中启用
	db.Model(&model.Tenant{}).Where("id = ?", 1005).Update("status", 1)

	// 调用 OnTenantEnabled 清除缓存
	svc.OnTenantEnabled(1005)

	// 再次查询应重新从 DB 加载，返回 true
	if !svc.IsTenantActive(1005) {
		t.Fatal("期望 OnTenantEnabled 后重新加载 DB，返回 true")
	}
}
