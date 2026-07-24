package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库并初始化表结构
func setupMenuTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 创建 admin_resource 表
	err = db.Exec(`CREATE TABLE admin_resource (
		id INTEGER PRIMARY KEY,
		parent_id INTEGER,
		type VARCHAR(16) NOT NULL DEFAULT 'MENU',
		name VARCHAR(128) NOT NULL,
		permission_code VARCHAR(128),
		path VARCHAR(256),
		component VARCHAR(256),
		icon VARCHAR(128),
		app_code VARCHAR(64) NOT NULL,
		platform VARCHAR(16) NOT NULL DEFAULT 'admin',
		module_code VARCHAR(64),
		sort_order INTEGER DEFAULT 0,
		status INTEGER DEFAULT 1,
		version INTEGER DEFAULT 1,
		deleted_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	// 创建 admin_tenant_app 表
	err = db.Exec(`CREATE TABLE admin_tenant_app (
		id INTEGER PRIMARY KEY,
		tenant_id INTEGER NOT NULL,
		app_code VARCHAR(64) NOT NULL,
		enabled_modules TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	return db
}

// seedMenuTestData 插入测试数据
func seedMenuTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	// 插入租户订阅的应用
	db.Exec(`INSERT INTO admin_tenant_app (id, tenant_id, app_code, enabled_modules) VALUES (1, 1, 'shop', '["order","product"]')`)
	db.Exec(`INSERT INTO admin_tenant_app (id, tenant_id, app_code, enabled_modules) VALUES (2, 1, 'cms', NULL)`)
	db.Exec(`INSERT INTO admin_tenant_app (id, tenant_id, app_code, enabled_modules) VALUES (3, 2, 'shop', '["order"]')`)

	// 插入 platform=user 的菜单资源
	db.Exec(`INSERT INTO admin_resource (id, parent_id, type, name, path, app_code, platform, module_code, sort_order, status) VALUES (100, NULL, 'MENU', '订单中心', '/orders', 'shop', 'user', 'order', 1, 1)`)
	db.Exec(`INSERT INTO admin_resource (id, parent_id, type, name, path, app_code, platform, module_code, sort_order, status) VALUES (101, 100, 'MENU', '我的订单', '/orders/mine', 'shop', 'user', 'order', 1, 1)`)
	db.Exec(`INSERT INTO admin_resource (id, parent_id, type, name, path, app_code, platform, module_code, sort_order, status) VALUES (102, NULL, 'MENU', '商品浏览', '/products', 'shop', 'user', 'product', 2, 1)`)
	db.Exec(`INSERT INTO admin_resource (id, parent_id, type, name, path, app_code, platform, module_code, sort_order, status) VALUES (103, NULL, 'MENU', '优惠券', '/coupons', 'shop', 'user', 'coupon', 3, 1)`)
	db.Exec(`INSERT INTO admin_resource (id, parent_id, type, name, path, app_code, platform, module_code, sort_order, status) VALUES (200, NULL, 'MENU', '文章列表', '/articles', 'cms', 'user', '', 1, 1)`)

	// 插入 platform=admin 的资源（不应被返回）
	db.Exec(`INSERT INTO admin_resource (id, parent_id, type, name, path, app_code, platform, module_code, sort_order, status) VALUES (300, NULL, 'MENU', '后台管理', '/admin', 'shop', 'admin', '', 1, 1)`)

	// 插入已禁用的 user 资源（不应被返回）
	db.Exec(`INSERT INTO admin_resource (id, parent_id, type, name, path, app_code, platform, module_code, sort_order, status) VALUES (400, NULL, 'MENU', '已禁用功能', '/disabled', 'shop', 'user', 'order', 4, 0)`)
}

// TestGetUserMenu_ReturnsTenantMenu 验证 GetUserMenu 返回该租户的 user 菜单
func TestGetUserMenu_ReturnsTenantMenu(t *testing.T) {
	db := setupMenuTestDB(t)
	seedMenuTestData(t, db)

	svc := NewUserMenuService(db)

	// tenantID=1 订阅了 shop(order,product) + cms(全部)
	tree, err := svc.GetUserMenu(1)
	require.NoError(t, err)
	assert.NotEmpty(t, tree)

	// 应该包含：订单中心(100)、商品浏览(102)、文章列表(200)
	// 不包含：优惠券(103，module_code=coupon 不在 enabled_modules 中)
	// 不包含：后台管理(300，platform=admin)
	// 不包含：已禁用功能(400，status=0)
	names := collectNodeNames(tree)
	assert.Contains(t, names, "订单中心")
	assert.Contains(t, names, "商品浏览")
	assert.Contains(t, names, "文章列表")
	assert.NotContains(t, names, "优惠券")
	assert.NotContains(t, names, "后台管理")
	assert.NotContains(t, names, "已禁用功能")

	// 验证树形结构：订单中心应有子节点"我的订单"
	var orderCenter *ResourceNode
	for _, n := range tree {
		if n.Name == "订单中心" {
			orderCenter = n
			break
		}
	}
	require.NotNil(t, orderCenter)
	require.Len(t, orderCenter.Children, 1)
	assert.Equal(t, "我的订单", orderCenter.Children[0].Name)
}

// TestGetUserMenu_CacheHit 验证第二次调用命中缓存
func TestGetUserMenu_CacheHit(t *testing.T) {
	db := setupMenuTestDB(t)
	seedMenuTestData(t, db)

	svc := NewUserMenuService(db)

	// 第一次调用
	tree1, err := svc.GetUserMenu(1)
	require.NoError(t, err)
	assert.NotEmpty(t, tree1)

	// 删除数据库中的数据（模拟验证缓存命中不查库）
	db.Exec("DELETE FROM admin_resource")
	db.Exec("DELETE FROM admin_tenant_app")

	// 第二次调用应命中缓存，返回相同结果
	tree2, err := svc.GetUserMenu(1)
	require.NoError(t, err)
	assert.Equal(t, len(tree1), len(tree2))
	assert.Equal(t, tree1[0].Name, tree2[0].Name)
}

// TestGetUserMenu_InvalidateCache 验证 Invalidate 后重新查库
func TestGetUserMenu_InvalidateCache(t *testing.T) {
	db := setupMenuTestDB(t)
	seedMenuTestData(t, db)

	svc := NewUserMenuService(db)

	// 第一次调用
	tree1, err := svc.GetUserMenu(1)
	require.NoError(t, err)
	assert.NotEmpty(t, tree1)

	// 新增一条资源
	db.Exec(`INSERT INTO admin_resource (id, parent_id, type, name, path, app_code, platform, module_code, sort_order, status) VALUES (500, NULL, 'MENU', '新功能', '/new', 'cms', 'user', '', 5, 1)`)

	// 未 Invalidate 前仍是旧缓存
	tree2, err := svc.GetUserMenu(1)
	require.NoError(t, err)
	names2 := collectNodeNames(tree2)
	assert.NotContains(t, names2, "新功能")

	// Invalidate 后重新查库
	svc.InvalidateUserMenuCache(1)
	tree3, err := svc.GetUserMenu(1)
	require.NoError(t, err)
	names3 := collectNodeNames(tree3)
	assert.Contains(t, names3, "新功能")
}

// TestGetUserMenu_DifferentTenants 验证不同租户独立缓存
func TestGetUserMenu_DifferentTenants(t *testing.T) {
	db := setupMenuTestDB(t)
	seedMenuTestData(t, db)

	svc := NewUserMenuService(db)

	// tenantID=1 订阅了 shop(order,product) + cms
	tree1, err := svc.GetUserMenu(1)
	require.NoError(t, err)
	names1 := collectNodeNames(tree1)
	assert.Contains(t, names1, "商品浏览") // product 模块

	// tenantID=2 只订阅了 shop(order)
	tree2, err := svc.GetUserMenu(2)
	require.NoError(t, err)
	names2 := collectNodeNames(tree2)
	assert.Contains(t, names2, "订单中心")
	assert.NotContains(t, names2, "商品浏览") // product 不在 tenant 2 的 enabled_modules
	assert.NotContains(t, names2, "文章列表") // cms 不在 tenant 2 订阅中
}

// TestGetUserMenu_CacheTTLExpiry 验证缓存 TTL 过期后重新查库
func TestGetUserMenu_CacheTTLExpiry(t *testing.T) {
	db := setupMenuTestDB(t)
	seedMenuTestData(t, db)

	svc := NewUserMenuService(db)

	// 第一次调用填充缓存
	_, err := svc.GetUserMenu(1)
	require.NoError(t, err)

	// 手动修改缓存时间为过期
	if val, ok := svc.cache.Load(int64(1)); ok {
		entry := val.(*menuCacheEntry)
		entry.cachedAt = time.Now().Add(-11 * time.Minute) // 超过 10 分钟 TTL
	}

	// 新增资源
	db.Exec(`INSERT INTO admin_resource (id, parent_id, type, name, path, app_code, platform, module_code, sort_order, status) VALUES (600, NULL, 'MENU', 'TTL测试', '/ttl', 'cms', 'user', '', 6, 1)`)

	// 缓存过期后应重新查库
	tree, err := svc.GetUserMenu(1)
	require.NoError(t, err)
	names := collectNodeNames(tree)
	assert.Contains(t, names, "TTL测试")
}

// collectNodeNames 递归收集所有节点名称
func collectNodeNames(nodes []*ResourceNode) []string {
	var names []string
	for _, n := range nodes {
		names = append(names, n.Name)
		if len(n.Children) > 0 {
			names = append(names, collectNodeNames(n.Children)...)
		}
	}
	return names
}
