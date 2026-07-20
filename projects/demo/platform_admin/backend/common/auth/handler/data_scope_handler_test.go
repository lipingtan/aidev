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

// setupDataScopeTestDB 创建 SQLite 内存数据库并迁移数据权限相关表
func setupDataScopeTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	err = db.AutoMigrate(
		&model.DataScopeConfig{},
		&model.DataScope{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// setupDataScopeRouter 创建数据权限测试路由
func setupDataScopeRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	configRepo := repository.NewDataScopeConfigRepository()
	scopeRepo := repository.NewDataScopeRepository()
	svc := service.NewDataScopeService(db, configRepo, scopeRepo)
	h := NewDataScopeHandler(svc)

	api := r.Group("/api/v1/admin")
	h.RegisterRoutes(api)

	return r
}

// TestCreateDataScopeConfig_Success 测试注册维度→DB 记录创建
func TestCreateDataScopeConfig_Success(t *testing.T) {
	db := setupDataScopeTestDB(t)
	r := setupDataScopeRouter(db)

	body := map[string]interface{}{
		"dimension_name": "department",
		"display_name":   "部门",
		"table_column":   "dept_id",
		"value_source":   "admin_department",
		"handler_name":   "DepartmentHandler",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/data-scope-configs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                   `json:"code"`
		Data model.DataScopeConfig `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.DimensionName != "department" {
		t.Fatalf("期望 dimension_name=department，实际: %s", resp.Data.DimensionName)
	}

	// 验证数据库记录
	var count int64
	db.Model(&model.DataScopeConfig{}).Where("dimension_name = ?", "department").Count(&count)
	if count != 1 {
		t.Fatalf("期望 DB 中有 1 条记录，实际: %d", count)
	}
}

// TestCreateDataScopeConfig_DuplicateName 测试重复 dimension_name
func TestCreateDataScopeConfig_DuplicateName(t *testing.T) {
	db := setupDataScopeTestDB(t)
	r := setupDataScopeRouter(db)

	body := map[string]interface{}{
		"dimension_name": "region",
		"display_name":   "区域",
	}
	jsonBody, _ := json.Marshal(body)

	// 第一次创建
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/data-scope-configs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("第一次创建失败: %d", w.Code)
	}

	// 第二次创建（重复）
	jsonBody, _ = json.Marshal(body)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/data-scope-configs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestUpdateDataScopeConfig_Success 测试更新维度→记录更新
func TestUpdateDataScopeConfig_Success(t *testing.T) {
	db := setupDataScopeTestDB(t)
	r := setupDataScopeRouter(db)

	// 先创建
	createBody := map[string]interface{}{
		"dimension_name": "org",
		"display_name":   "组织",
	}
	jsonBody, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/data-scope-configs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("创建失败: %d, body: %s", w.Code, w.Body.String())
	}

	var createResp struct {
		Data model.DataScopeConfig `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	id := createResp.Data.ID

	// 更新
	updateBody := map[string]interface{}{
		"display_name": "组织架构",
		"table_column": "org_id",
	}
	jsonBody, _ = json.Marshal(updateBody)
	url := fmt.Sprintf("/api/v1/admin/data-scope-configs/%d", id)
	req = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("更新失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 DB 中的更新
	var cfg model.DataScopeConfig
	db.First(&cfg, id)
	if cfg.DisplayName != "组织架构" {
		t.Fatalf("期望 display_name=组织架构，实际: %s", cfg.DisplayName)
	}
	if cfg.TableColumn != "org_id" {
		t.Fatalf("期望 table_column=org_id，实际: %s", cfg.TableColumn)
	}
}

// TestDeleteDataScopeConfig_Success 测试删除维度→记录删除
func TestDeleteDataScopeConfig_Success(t *testing.T) {
	db := setupDataScopeTestDB(t)
	r := setupDataScopeRouter(db)

	// 先创建
	createBody := map[string]interface{}{
		"dimension_name": "to_delete",
		"display_name":   "待删除",
	}
	jsonBody, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/data-scope-configs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("创建失败: %d", w.Code)
	}

	var createResp struct {
		Data model.DataScopeConfig `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	id := createResp.Data.ID

	// 删除
	url := fmt.Sprintf("/api/v1/admin/data-scope-configs/%d", id)
	req = httptest.NewRequest(http.MethodDelete, url, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("删除失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 DB 中记录已删除
	var count int64
	db.Model(&model.DataScopeConfig{}).Where("id = ?", id).Count(&count)
	if count != 0 {
		t.Fatalf("期望记录被删除，实际仍存在 %d 条", count)
	}
}

// TestListDataScopeConfigs 测试查询维度列表
func TestListDataScopeConfigs(t *testing.T) {
	db := setupDataScopeTestDB(t)
	r := setupDataScopeRouter(db)

	// 创建 2 个维度
	for _, name := range []string{"dim_a", "dim_b"} {
		body := map[string]interface{}{
			"dimension_name": name,
			"display_name":   name + "_display",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/data-scope-configs", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("创建 %s 失败: %d", name, w.Code)
		}
	}

	// 查询列表
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/data-scope-configs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("查询列表失败: %d", w.Code)
	}

	var resp struct {
		Code int                     `json:"code"`
		Data []model.DataScopeConfig `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Data) != 2 {
		t.Fatalf("期望 2 条记录，实际: %d", len(resp.Data))
	}
}

// TestSetRoleDataScopes_Success 测试角色绑定数据权限→admin_data_scope 记录正确
func TestSetRoleDataScopes_Success(t *testing.T) {
	db := setupDataScopeTestDB(t)
	r := setupDataScopeRouter(db)

	roleID := int64(100)

	body := map[string]interface{}{
		"scopes": []map[string]interface{}{
			{
				"dimension_name":   "department",
				"target_entity":    "orders",
				"dimension_values": []string{"dept_1", "dept_2"},
			},
			{
				"dimension_name":   "region",
				"target_entity":    "orders",
				"dimension_values": []string{"east", "west"},
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	url := fmt.Sprintf("/api/v1/admin/roles/%d/data-scopes", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("绑定数据权限失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 DB 中 admin_data_scope 记录
	var scopes []model.DataScope
	db.Where("role_id = ?", roleID).Find(&scopes)
	if len(scopes) != 2 {
		t.Fatalf("期望 2 条 data_scope 记录，实际: %d", len(scopes))
	}

	// 验证第一条记录
	found := false
	for _, s := range scopes {
		if s.DimensionName == "department" && s.TargetEntity == "orders" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("未找到 department 维度绑定记录")
	}
}

// TestSetRoleDataScopes_ReplaceExisting 测试重复配置会替换旧绑定
func TestSetRoleDataScopes_ReplaceExisting(t *testing.T) {
	db := setupDataScopeTestDB(t)
	r := setupDataScopeRouter(db)

	roleID := int64(200)

	// 第一次绑定
	body1 := map[string]interface{}{
		"scopes": []map[string]interface{}{
			{
				"dimension_name":   "old_dim",
				"target_entity":    "users",
				"dimension_values": []string{"v1"},
			},
		},
	}
	jsonBody, _ := json.Marshal(body1)
	url := fmt.Sprintf("/api/v1/admin/roles/%d/data-scopes", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("第一次绑定失败: %d", w.Code)
	}

	// 第二次绑定（替换）
	body2 := map[string]interface{}{
		"scopes": []map[string]interface{}{
			{
				"dimension_name":   "new_dim",
				"target_entity":    "orders",
				"dimension_values": []string{"v2"},
			},
		},
	}
	jsonBody, _ = json.Marshal(body2)
	req = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("第二次绑定失败: %d", w.Code)
	}

	// 验证旧记录被替换
	var scopes []model.DataScope
	db.Where("role_id = ?", roleID).Find(&scopes)
	if len(scopes) != 1 {
		t.Fatalf("期望替换后只有 1 条记录，实际: %d", len(scopes))
	}
	if scopes[0].DimensionName != "new_dim" {
		t.Fatalf("期望新维度 new_dim，实际: %s", scopes[0].DimensionName)
	}
}

// TestGetRoleDataScopes 测试查询角色数据权限
func TestGetRoleDataScopes(t *testing.T) {
	db := setupDataScopeTestDB(t)
	r := setupDataScopeRouter(db)

	roleID := int64(300)

	// 先绑定
	body := map[string]interface{}{
		"scopes": []map[string]interface{}{
			{
				"dimension_name":   "area",
				"target_entity":    "products",
				"dimension_values": []string{"north"},
			},
		},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/admin/roles/%d/data-scopes", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("绑定失败: %d", w.Code)
	}

	// 查询
	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("查询失败: %d", w.Code)
	}

	var resp struct {
		Code int               `json:"code"`
		Data []model.DataScope `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Data) != 1 {
		t.Fatalf("期望 1 条绑定，实际: %d", len(resp.Data))
	}
	if resp.Data[0].DimensionName != "area" {
		t.Fatalf("期望 dimension_name=area，实际: %s", resp.Data[0].DimensionName)
	}
}
