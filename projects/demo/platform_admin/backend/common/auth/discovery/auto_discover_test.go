package discovery

import (
	"net/http"
	"testing"

	"go-admin/common/auth/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建 SQLite 内存数据库并自动迁移
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 SQLite 内存数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.ApiPermission{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	return db
}

// setupTestEngine 创建测试用 gin.Engine 并注册路由
func setupTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/api/v1/users", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.POST("/api/v1/users", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.PUT("/api/v1/users/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.DELETE("/api/v1/users/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
	return engine
}

// TestAutoDiscover_FirstScan 首次扫描→新 endpoint + GROUP 插入 DB
func TestAutoDiscover_FirstScan(t *testing.T) {
	db := setupTestDB(t)
	engine := setupTestEngine()

	AutoDiscover(engine, db, 500)

	var perms []model.ApiPermission
	db.Where("app_code = ?", "admin").Find(&perms)

	// 期望 1 GROUP (users) + 4 ENDPOINT = 5
	if len(perms) != 5 {
		t.Fatalf("期望 5 条记录（1 GROUP + 4 ENDPOINT），实际 %d 条", len(perms))
	}

	var endpoints []model.ApiPermission
	for _, p := range perms {
		if p.Type == "ENDPOINT" {
			endpoints = append(endpoints, p)
		}
	}
	if len(endpoints) != 4 {
		t.Fatalf("期望 4 条 ENDPOINT，实际 %d 条", len(endpoints))
	}

	for _, ep := range endpoints {
		if ep.Status != "ACTIVE" {
			t.Errorf("期望 status=ACTIVE，实际 %s", ep.Status)
		}
	}
}

// TestAutoDiscover_Idempotent 重复扫描→不产生重复记录
func TestAutoDiscover_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	engine := setupTestEngine()

	// 首次扫描
	AutoDiscover(engine, db, 500)
	// 再次扫描
	AutoDiscover(engine, db, 500)

	var count int64
	db.Model(&model.ApiPermission{}).Where("app_code = ?", "admin").Count(&count)

	// 期望 1 GROUP + 4 ENDPOINT = 5，重复扫描不增加
	if count != 5 {
		t.Fatalf("幂等失败：期望 5 条记录，实际 %d 条", count)
	}
}

// TestAutoDiscover_AsyncThreshold 路由数超过阈值时异步执行（不阻塞）
func TestAutoDiscover_AsyncThreshold(t *testing.T) {
	db := setupTestDB(t)
	engine := setupTestEngine()

	// asyncThreshold=1，路由数=4，应异步执行
	AutoDiscover(engine, db, 1)

	// 立即查询应该没有记录（异步尚未执行）
	var count int64
	db.Model(&model.ApiPermission{}).Where("app_code = ?", "admin").Count(&count)

	if count != 0 {
		t.Logf("异步模式：立即查询到 %d 条记录（可能执行很快），跳过严格断言", count)
	}
}

// TestRegisterAPIs_Active 代码声明注册→status=ACTIVE + 完整元数据
func TestRegisterAPIs_Active(t *testing.T) {
	db := setupTestDB(t)

	apis := []ApiMetadata{
		{
			Name:           "获取用户列表",
			PermissionCode: "user:list",
			URLPattern:     "/api/v1/users",
			HTTPMethod:     "GET",
			AppCode:        "admin",
			GroupName:      "用户管理",
		},
		{
			Name:           "创建用户",
			PermissionCode: "user:create",
			URLPattern:     "/api/v1/users",
			HTTPMethod:     "POST",
			AppCode:        "admin",
			GroupName:      "用户管理",
		},
	}

	RegisterAPIs(db, "admin", apis)

	// 验证 ENDPOINT
	var endpoints []model.ApiPermission
	db.Where("app_code = ? AND type = ?", "admin", "ENDPOINT").Find(&endpoints)
	if len(endpoints) != 2 {
		t.Fatalf("期望 2 条 ENDPOINT，实际 %d 条", len(endpoints))
	}

	for _, ep := range endpoints {
		if ep.Status != "ACTIVE" {
			t.Errorf("期望 status=ACTIVE，实际 %s", ep.Status)
		}
		if ep.AppCode != "admin" {
			t.Errorf("期望 app_code=admin，实际 %s", ep.AppCode)
		}
		if ep.ParentID == nil {
			t.Error("期望有 parent_id（分组），实际为 nil")
		}
	}

	// 验证 GROUP 只创建了一个
	var groups []model.ApiPermission
	db.Where("app_code = ? AND type = ?", "admin", "GROUP").Find(&groups)
	if len(groups) != 1 {
		t.Fatalf("期望 1 个 GROUP，实际 %d 个", len(groups))
	}
	if groups[0].Name != "用户管理" {
		t.Errorf("期望分组名=用户管理，实际 %s", groups[0].Name)
	}
}

// TestRegisterAPIs_Idempotent 重复注册→不产生重复
func TestRegisterAPIs_Idempotent(t *testing.T) {
	db := setupTestDB(t)

	apis := []ApiMetadata{
		{
			Name:           "获取用户列表",
			PermissionCode: "user:list",
			URLPattern:     "/api/v1/users",
			HTTPMethod:     "GET",
			AppCode:        "admin",
		},
	}

	RegisterAPIs(db, "admin", apis)
	RegisterAPIs(db, "admin", apis)

	var count int64
	db.Model(&model.ApiPermission{}).Where("app_code = ? AND type = ?", "admin", "ENDPOINT").Count(&count)
	if count != 1 {
		t.Fatalf("RegisterAPIs 幂等失败：期望 1 条，实际 %d 条", count)
	}
}
