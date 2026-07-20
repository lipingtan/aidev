package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupAppTestDB 创建应用测试用 SQLite 内存数据库
func setupAppTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	err = db.AutoMigrate(
		&model.Application{},
		&model.TenantApp{},
		&model.RoleApp{},
		&model.Role{},
		&model.Tenant{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// setupAppRouter 创建应用测试路由
func setupAppRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	appRepo := repository.NewApplicationRepository()
	roleRepo := repository.NewRoleRepository()
	svc := service.NewApplicationService(db, appRepo)
	h := NewApplicationHandler(svc, roleRepo, db)

	api := r.Group("/api/v1/admin")
	h.RegisterRoutes(api)

	return r
}

// TestCreateApplication_Success 测试创建应用→DB 记录存在
func TestCreateApplication_Success(t *testing.T) {
	db := setupAppTestDB(t)
	r := setupAppRouter(db)

	body := map[string]interface{}{
		"app_code": "crm",
		"name":     "客户管理系统",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/applications", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int              `json:"code"`
		Data model.Application `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}

	// 验证 DB 记录存在
	var app model.Application
	err := db.Where("app_code = ?", "crm").First(&app).Error
	if err != nil {
		t.Fatalf("数据库中未找到应用记录: %v", err)
	}
	if app.Name != "客户管理系统" {
		t.Fatalf("期望 name=客户管理系统，实际: %s", app.Name)
	}
}

// TestSetTenantApps_Success 测试租户订阅→admin_tenant_app 记录正确
func TestSetTenantApps_Success(t *testing.T) {
	db := setupAppTestDB(t)
	r := setupAppRouter(db)

	// 创建租户
	tenant := &model.Tenant{ID: 1001, TenantCode: "t1", Name: "租户1", Status: 1, Version: 1}
	db.Create(tenant)

	// 创建应用
	db.Create(&model.Application{AppCode: "app1", Name: "应用1", Status: 1, Version: 1})
	db.Create(&model.Application{AppCode: "app2", Name: "应用2", Status: 1, Version: 1})

	// 设置订阅
	body := map[string]interface{}{
		"app_codes": []string{"app1", "app2"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/tenants/1001/apps", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 admin_tenant_app 记录
	var tenantApps []model.TenantApp
	db.Where("tenant_id = ?", 1001).Find(&tenantApps)
	if len(tenantApps) != 2 {
		t.Fatalf("期望 2 条订阅记录，实际: %d", len(tenantApps))
	}
}

// TestUnsubscribeCascadeDeleteRoleApps 测试取消订阅→级联清除角色绑定
func TestUnsubscribeCascadeDeleteRoleApps(t *testing.T) {
	db := setupAppTestDB(t)
	r := setupAppRouter(db)

	// 创建租户和角色
	tenant := &model.Tenant{ID: 2001, TenantCode: "t2", Name: "租户2", Status: 1, Version: 1}
	db.Create(tenant)
	role := &model.Role{ID: 3001, TenantID: 2001, RoleCode: "editor", RoleName: "编辑者", Status: 1, Version: 1}
	db.Create(role)

	// 创建应用
	db.Create(&model.Application{AppCode: "appA", Name: "应用A", Status: 1, Version: 1})
	db.Create(&model.Application{AppCode: "appB", Name: "应用B", Status: 1, Version: 1})

	// 先订阅 appA 和 appB
	db.Create(&model.TenantApp{TenantID: 2001, AppCode: "appA"})
	db.Create(&model.TenantApp{TenantID: 2001, AppCode: "appB"})

	// 角色绑定 appA 和 appB
	db.Create(&model.RoleApp{RoleID: 3001, AppCode: "appA"})
	db.Create(&model.RoleApp{RoleID: 3001, AppCode: "appB"})

	// 取消订阅 appB（只保留 appA）
	body := map[string]interface{}{
		"app_codes": []string{"appA"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/tenants/2001/apps", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证角色绑定：appB 应被级联删除，appA 仍存在
	var roleApps []model.RoleApp
	db.Where("role_id = ?", 3001).Find(&roleApps)
	if len(roleApps) != 1 {
		t.Fatalf("期望 1 条角色绑定，实际: %d", len(roleApps))
	}
	if roleApps[0].AppCode != "appA" {
		t.Fatalf("期望保留 appA，实际: %s", roleApps[0].AppCode)
	}
}

// TestSetRoleApps_NotSubscribed 测试角色绑定未订阅应用→返回 400 (code=40005)
func TestSetRoleApps_NotSubscribed(t *testing.T) {
	db := setupAppTestDB(t)
	r := setupAppRouter(db)

	// 创建租户和角色
	tenant := &model.Tenant{ID: 4001, TenantCode: "t4", Name: "租户4", Status: 1, Version: 1}
	db.Create(tenant)
	role := &model.Role{ID: 5001, TenantID: 4001, RoleCode: "viewer", RoleName: "查看者", Status: 1, Version: 1}
	db.Create(role)

	// 租户未订阅任何应用，直接绑定
	body := map[string]interface{}{
		"app_codes": []string{"not_subscribed_app"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/roles/5001/apps", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40005 {
		t.Fatalf("期望 code=40005，实际: %d", resp.Code)
	}
}

// TestListTenantApps_Success 测试查询租户已订阅应用→返回正确列表
func TestListTenantApps_Success(t *testing.T) {
	db := setupAppTestDB(t)
	r := setupAppRouter(db)

	// 创建租户
	tenant := &model.Tenant{ID: 6001, TenantCode: "t6", Name: "租户6", Status: 1, Version: 1}
	db.Create(tenant)

	// 创建应用
	db.Create(&model.Application{AppCode: "x1", Name: "应用X1", Status: 1, Version: 1})
	db.Create(&model.Application{AppCode: "x2", Name: "应用X2", Status: 1, Version: 1})
	db.Create(&model.Application{AppCode: "x3", Name: "应用X3", Status: 1, Version: 1})

	// 直接插入订阅记录
	db.Create(&model.TenantApp{TenantID: 6001, AppCode: "x1"})
	db.Create(&model.TenantApp{TenantID: 6001, AppCode: "x2"})
	db.Create(&model.TenantApp{TenantID: 6001, AppCode: "x3"})

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/admin/tenants/%d/apps", 6001), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                   `json:"code"`
		Data []model.Application   `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if len(resp.Data) != 3 {
		t.Fatalf("期望 3 条记录，实际: %d", len(resp.Data))
	}
}
