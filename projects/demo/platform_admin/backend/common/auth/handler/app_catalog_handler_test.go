package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-admin/common/auth/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupCatalogTestDB 创建应用目录测试用 SQLite 内存数据库
func setupCatalogTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	err = db.AutoMigrate(
		&model.Application{},
		&model.TenantApp{},
		&model.Role{},
		&model.Resource{},
		&model.ApiPermission{},
		&model.RoleResource{},
		&model.RoleApi{},
		&model.AdminConfig{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// setupCatalogRouter 创建应用目录测试路由（含模拟认证中间件）
func setupCatalogRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewAppCatalogHandler(db, nil)

	api := r.Group("/api/v1/admin")
	api.Use(mockAuthMiddleware())
	h.RegisterRoutes(api)

	return r
}

// TestGetAppCatalog_Success 测试获取应用目录→200 + 含 BUILTIN 和 PLUGIN 应用 + 订阅状态
func TestGetAppCatalog_Success(t *testing.T) {
	db := setupCatalogTestDB(t)
	r := setupCatalogRouter(db)

	// 创建应用
	db.Create(&model.Application{ID: 1, AppCode: "core", Name: "核心平台", AppType: "BUILTIN", Status: 1, Version: 1})
	db.Create(&model.Application{ID: 2, AppCode: "shop", Name: "商城插件", AppType: "PLUGIN", Status: 1, Version: 1})

	// 租户 1 订阅了 core
	db.Create(&model.TenantApp{ID: 100, TenantID: 1, AppCode: "core"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/app-catalog", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int              `json:"code"`
		Data []AppCatalogItem `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	// 返回所有启用的应用
	if len(resp.Data) != 2 {
		t.Fatalf("期望 2 条应用，实际: %d", len(resp.Data))
	}

	// 验证订阅状态
	coreFound := false
	shopFound := false
	for _, item := range resp.Data {
		if item.AppCode == "core" {
			coreFound = true
			if !item.Subscribed {
				t.Fatal("core 应已订阅")
			}
		}
		if item.AppCode == "shop" {
			shopFound = true
			if item.Subscribed {
				t.Fatal("shop 不应已订阅")
			}
		}
	}
	if !coreFound || !shopFound {
		t.Fatal("缺少 core 或 shop 应用")
	}
}

// TestSubscribeApp_Success 测试合法订阅→201
func TestSubscribeApp_Success(t *testing.T) {
	db := setupCatalogTestDB(t)
	r := setupCatalogRouter(db)

	db.Create(&model.Application{ID: 1, AppCode: "crm", Name: "CRM", AppType: "PLUGIN", Status: 1, Version: 1})

	body := map[string]interface{}{"app_code": "crm"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/app-subscriptions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("期望状态码 201，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 DB 记录
	var ta model.TenantApp
	err := db.Where("tenant_id = 1 AND app_code = ?", "crm").First(&ta).Error
	if err != nil {
		t.Fatalf("数据库中未找到订阅记录: %v", err)
	}
}

// TestSubscribeApp_QuotaExceeded 测试超配额→400
func TestSubscribeApp_QuotaExceeded(t *testing.T) {
	db := setupCatalogTestDB(t)
	r := setupCatalogRouter(db)

	// 设置配额为 2
	db.Create(&model.AdminConfig{
		ID:          1,
		ConfigKey:   "quota.max_apps",
		ConfigValue: "2",
		ConfigType:  "number",
		Scope:       "SYSTEM",
		ScopeID:     0,
		TenantID:    0,
		Status:      1,
	})

	// 创建 3 个应用
	db.Create(&model.Application{ID: 1, AppCode: "app1", Name: "App1", AppType: "PLUGIN", Status: 1, Version: 1})
	db.Create(&model.Application{ID: 2, AppCode: "app2", Name: "App2", AppType: "PLUGIN", Status: 1, Version: 1})
	db.Create(&model.Application{ID: 3, AppCode: "app3", Name: "App3", AppType: "PLUGIN", Status: 1, Version: 1})

	// 已订阅 2 个
	db.Create(&model.TenantApp{ID: 100, TenantID: 1, AppCode: "app1"})
	db.Create(&model.TenantApp{ID: 101, TenantID: 1, AppCode: "app2"})

	// 尝试订阅第 3 个
	body := map[string]interface{}{"app_code": "app3"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/app-subscriptions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40000 {
		t.Fatalf("期望 code=40000，实际: %d", resp.Code)
	}
}

// TestUnsubscribeApp_Plugin_Success 测试退订 PLUGIN 类型应用→200 + 级联清理角色绑定
func TestUnsubscribeApp_Plugin_Success(t *testing.T) {
	db := setupCatalogTestDB(t)
	r := setupCatalogRouter(db)

	// 创建应用
	db.Create(&model.Application{ID: 1, AppCode: "shop", Name: "商城", AppType: "PLUGIN", Status: 1, Version: 1})

	// 创建订阅
	db.Create(&model.TenantApp{ID: 100, TenantID: 1, AppCode: "shop"})

	// 创建该 app 的资源和 API
	db.Create(&model.Resource{ID: 201, AppCode: "shop", Name: "商品管理", Type: "menu", Platform: "admin", Status: 1, Version: 1})
	db.Create(&model.ApiPermission{ID: 301, AppCode: "shop", Name: "获取商品", Type: "ENDPOINT", HTTPMethod: "GET", URLPattern: "/api/v1/shop/products"})

	// 创建租户角色和绑定
	db.Create(&model.Role{ID: 401, TenantID: 1, RoleCode: "shop_admin", RoleName: "商城管理员", Status: 1, Version: 1})
	db.Create(&model.RoleResource{ID: 501, RoleID: 401, ResourceID: 201})
	db.Create(&model.RoleApi{ID: 601, RoleID: 401, ApiPermissionID: 301})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/app-subscriptions/shop", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证订阅记录已删除
	var count int64
	db.Model(&model.TenantApp{}).Where("tenant_id = 1 AND app_code = ?", "shop").Count(&count)
	if count != 0 {
		t.Fatal("订阅记录应已删除")
	}

	// 验证级联清理：角色资源绑定已删除
	db.Model(&model.RoleResource{}).Where("role_id = 401 AND resource_id = 201").Count(&count)
	if count != 0 {
		t.Fatal("角色资源绑定应已被级联清理")
	}

	// 验证级联清理：角色 API 绑定已删除
	db.Model(&model.RoleApi{}).Where("role_id = 401 AND api_permission_id = 301").Count(&count)
	if count != 0 {
		t.Fatal("角色 API 绑定应已被级联清理")
	}
}

// TestUnsubscribeApp_Builtin_Rejected 测试退订 BUILTIN 类型→400
func TestUnsubscribeApp_Builtin_Rejected(t *testing.T) {
	db := setupCatalogTestDB(t)
	r := setupCatalogRouter(db)

	db.Create(&model.Application{ID: 1, AppCode: "core", Name: "核心平台", AppType: "BUILTIN", Status: 1, Version: 1})
	db.Create(&model.TenantApp{ID: 100, TenantID: 1, AppCode: "core"})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/app-subscriptions/core", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证订阅记录未被删除
	var count int64
	db.Model(&model.TenantApp{}).Where("tenant_id = 1 AND app_code = ?", "core").Count(&count)
	if count != 1 {
		t.Fatal("BUILTIN 应用的订阅记录不应被删除")
	}
}

// TestGetSubscriptions_Success 测试获取当前租户已订阅列表→200
func TestGetSubscriptions_Success(t *testing.T) {
	db := setupCatalogTestDB(t)
	r := setupCatalogRouter(db)

	// 创建应用
	db.Create(&model.Application{ID: 1, AppCode: "core", Name: "核心平台", AppType: "BUILTIN", Status: 1, Version: 1})
	db.Create(&model.Application{ID: 2, AppCode: "shop", Name: "商城", AppType: "PLUGIN", Status: 1, Version: 1})

	// 租户 1 订阅
	db.Create(&model.TenantApp{ID: 100, TenantID: 1, AppCode: "core"})
	db.Create(&model.TenantApp{ID: 101, TenantID: 1, AppCode: "shop"})

	// 租户 2 订阅（不应出现在结果中）
	db.Create(&model.TenantApp{ID: 200, TenantID: 2, AppCode: "core"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/app-subscriptions", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                `json:"code"`
		Data []SubscriptionItem `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("期望 2 条订阅记录，实际: %d", len(resp.Data))
	}
}
