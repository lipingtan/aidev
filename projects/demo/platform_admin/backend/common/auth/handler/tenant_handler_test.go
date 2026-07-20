package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-admin/common/auth/config"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建 SQLite 内存数据库并自动迁移
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	// 自动迁移所有相关表
	err = db.AutoMigrate(
		&model.Tenant{},
		&model.Role{},
		&model.User{},
		&model.UserTenant{},
		&model.UserRole{},
		&model.TenantApp{},
		&model.ApiPermission{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// setupRouter 创建测试路由
func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := config.DefaultConfig()
	tenantRepo := repository.NewTenantRepository()
	svc := service.NewTenantService(db, cfg, tenantRepo)
	h := NewTenantHandler(svc)

	api := r.Group("/api/v1/admin")
	h.RegisterRoutes(api)

	return r
}

// TestCreateTenant_Success 测试创建租户成功
func TestCreateTenant_Success(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	body := map[string]interface{}{
		"tenant_code": "test_corp",
		"name":        "测试公司",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenants", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d, message: %s", resp.Code, resp.Message)
	}
}

// TestCreateTenant_DuplicateCode 测试重复 tenant_code 返回 400
func TestCreateTenant_DuplicateCode(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	body := map[string]interface{}{
		"tenant_code": "dup_corp",
		"name":        "重复公司",
	}
	jsonBody, _ := json.Marshal(body)

	// 第一次创建
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenants", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("第一次创建失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 第二次创建（重复）
	jsonBody, _ = json.Marshal(body)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenants", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp.Code != 40001 {
		t.Fatalf("期望 code=40001，实际: %d", resp.Code)
	}
}

// TestUpdateStatus 测试启用/禁用租户状态变更
func TestUpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// 先创建租户
	createBody := map[string]interface{}{
		"tenant_code": "status_corp",
		"name":        "状态测试公司",
	}
	jsonBody, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenants", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("创建租户失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 从响应中获取租户 ID
	var createResp struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id,string"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	tenantID := createResp.Data.ID

	// 禁用租户
	statusBody := map[string]interface{}{
		"status": 0,
	}
	jsonBody, _ = json.Marshal(statusBody)
	url := fmt.Sprintf("/api/v1/admin/tenants/%d/status", tenantID)
	req = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("禁用租户失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证状态变更
	var tenant model.Tenant
	db.First(&tenant, tenantID)
	if tenant.Status != 0 {
		t.Fatalf("期望状态 0（禁用），实际: %d", tenant.Status)
	}
}

// TestListTenants_Pagination 测试分页查询
func TestListTenants_Pagination(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// 创建 3 个租户
	for i := 0; i < 3; i++ {
		body := map[string]interface{}{
			"tenant_code": fmt.Sprintf("page_corp_%d", i),
			"name":        fmt.Sprintf("分页测试公司%d", i),
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenants", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("创建租户 %d 失败: %d", i, w.Code)
		}
	}

	// 查询第一页，page_size=2
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants?page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("查询租户列表失败: %d", w.Code)
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List     []model.Tenant `json:"list"`
			Total    int64          `json:"total"`
			Page     int            `json:"page"`
			PageSize int            `json:"page_size"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.Total != 3 {
		t.Fatalf("期望 total=3，实际: %d", resp.Data.Total)
	}
	if len(resp.Data.List) != 2 {
		t.Fatalf("期望列表长度 2，实际: %d", len(resp.Data.List))
	}
	if resp.Data.Page != 1 {
		t.Fatalf("期望 page=1，实际: %d", resp.Data.Page)
	}
}
