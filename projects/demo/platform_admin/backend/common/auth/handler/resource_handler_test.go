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
)

// setupResourceRouter 创建资源测试路由（注入 mockAuthMiddleware）
func setupResourceRouter(t *testing.T) (*gin.Engine, *service.ResourceService) {
	db := setupTestDB(t)
	// 额外迁移 Resource 和 RoleResource 表
	err := db.AutoMigrate(&model.Resource{}, &model.RoleResource{})
	if err != nil {
		t.Fatalf("资源表迁移失败: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mockAuthMiddleware())

	cfg := config.DefaultConfig()
	resourceRepo := repository.NewResourceRepository()
	svc := service.NewResourceService(db, cfg, resourceRepo)
	h := NewResourceHandler(svc)

	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	return r, svc
}

// TestCreateResource_Success 测试创建资源成功
func TestCreateResource_Success(t *testing.T) {
	r, _ := setupResourceRouter(t)

	body := map[string]interface{}{
		"tenant_id": 1,
		"type":      "MENU",
		"name":      "系统管理",
		"path":      "/system",
		"app_code":  "default",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
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

// TestCreateResource_Button 测试创建按钮资源
func TestCreateResource_Button(t *testing.T) {
	r, _ := setupResourceRouter(t)

	// 先创建父菜单
	parentBody := map[string]interface{}{
		"tenant_id": 1,
		"type":      "MENU",
		"name":      "用户管理",
		"path":      "/user",
		"app_code":  "default",
	}
	jsonBody, _ := json.Marshal(parentBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	parentID := createResp.Data.ID

	// 创建按钮
	buttonBody := map[string]interface{}{
		"tenant_id":       1,
		"parent_id":       parentID,
		"type":            "BUTTON",
		"name":            "新增用户",
		"permission_code": "user:create",
		"app_code":        "default",
	}
	jsonBody, _ = json.Marshal(buttonBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d, message: %s", resp.Code, resp.Message)
	}
}

// TestGetTree_Success 测试获取资源树
func TestGetTree_Success(t *testing.T) {
	r, _ := setupResourceRouter(t)

	// 创建父菜单
	parentBody := map[string]interface{}{
		"tenant_id": 1,
		"type":      "MENU",
		"name":      "系统管理",
		"path":      "/system",
		"app_code":  "default",
	}
	jsonBody, _ := json.Marshal(parentBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	parentID := createResp.Data.ID

	// 创建子菜单
	childBody := map[string]interface{}{
		"tenant_id": 1,
		"parent_id": parentID,
		"type":      "MENU",
		"name":      "用户管理",
		"path":      "/system/user",
		"app_code":  "default",
	}
	jsonBody, _ = json.Marshal(childBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 查询树（通过 header 设置 tenant）
	req = httptest.NewRequest(http.MethodGet, "/api/v1/resources/tree", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                    `json:"code"`
		Data []service.ResourceNode `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("期望根节点数 1，实际: %d", len(resp.Data))
	}
	if len(resp.Data[0].Children) != 1 {
		t.Fatalf("期望子节点数 1，实际: %d", len(resp.Data[0].Children))
	}
}

// TestGetTree_AppCodeFilter 测试 app_code 过滤
func TestGetTree_AppCodeFilter(t *testing.T) {
	r, _ := setupResourceRouter(t)

	// 创建两个不同 app_code 的资源
	for _, app := range []string{"app1", "app2"} {
		body := map[string]interface{}{
			"tenant_id": 1,
			"type":      "MENU",
			"name":      "菜单_" + app,
			"path":      "/" + app,
			"app_code":  app,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-Tenant-ID", "1")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	// 过滤 app1（通过 header 设置 tenant）
	req := httptest.NewRequest(http.MethodGet, "/api/v1/resources/tree?app_code=app1", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Code int                    `json:"code"`
		Data []service.ResourceNode `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data) != 1 {
		t.Fatalf("期望过滤后根节点数 1，实际: %d", len(resp.Data))
	}
}

// TestDeleteResource_HasChildren 测试删除含子节点的资源返回 40004
func TestDeleteResource_HasChildren(t *testing.T) {
	r, _ := setupResourceRouter(t)

	// 创建父菜单
	parentBody := map[string]interface{}{
		"tenant_id": 1,
		"type":      "MENU",
		"name":      "父菜单",
		"app_code":  "default",
	}
	jsonBody, _ := json.Marshal(parentBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	parentID := createResp.Data.ID

	// 创建子菜单
	childBody := map[string]interface{}{
		"tenant_id": 1,
		"parent_id": parentID,
		"type":      "MENU",
		"name":      "子菜单",
		"app_code":  "default",
	}
	jsonBody, _ = json.Marshal(childBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 尝试删除父菜单
	url := fmt.Sprintf("/api/v1/resources/%d", parentID)
	req = httptest.NewRequest(http.MethodDelete, url, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40004 {
		t.Fatalf("期望 code=40004，实际: %d", resp.Code)
	}
}

// TestDeleteResource_NoChildren 测试删除无子节点的资源成功
func TestDeleteResource_NoChildren(t *testing.T) {
	r, _ := setupResourceRouter(t)

	// 创建资源
	body := map[string]interface{}{
		"tenant_id": 1,
		"type":      "MENU",
		"name":      "独立菜单",
		"app_code":  "default",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	resourceID := createResp.Data.ID

	// 删除
	url := fmt.Sprintf("/api/v1/resources/%d", resourceID)
	req = httptest.NewRequest(http.MethodDelete, url, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestSortResources 测试拖拽排序
func TestSortResources(t *testing.T) {
	r, _ := setupResourceRouter(t)

	// 创建多个资源
	var ids []int64
	for i := 0; i < 3; i++ {
		body := map[string]interface{}{
			"tenant_id":  1,
			"type":       "MENU",
			"name":       fmt.Sprintf("菜单%d", i),
			"sort_order": i,
			"app_code":   "default",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-Tenant-ID", "1")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var createResp struct {
			Code int `json:"code"`
			Data struct {
				ID int64 `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &createResp)
		ids = append(ids, createResp.Data.ID)
	}

	// 调整排序：反转
	sortBody := map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": ids[2], "sort_order": 0, "parent_id": nil},
			{"id": ids[1], "sort_order": 1, "parent_id": nil},
			{"id": ids[0], "sort_order": 2, "parent_id": nil},
		},
	}
	jsonBody, _ := json.Marshal(sortBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/resources/sort", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证排序结果：查询树（通过 header 设置 tenant）
	req = httptest.NewRequest(http.MethodGet, "/api/v1/resources/tree", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Code int                    `json:"code"`
		Data []service.ResourceNode `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data) < 3 {
		t.Fatalf("期望至少 3 个根节点，实际: %d", len(resp.Data))
	}
	// 第一个应该是原来的 ids[2]（sort_order=0）
	if resp.Data[0].ID != ids[2] {
		t.Fatalf("期望第一个节点 ID=%d，实际: %d", ids[2], resp.Data[0].ID)
	}
}

// TestResourceTenantIsolation 测试资源租户隔离
func TestResourceTenantIsolation(t *testing.T) {
	r, _ := setupResourceRouter(t)

	// 租户 1 创建资源
	body1 := map[string]interface{}{
		"tenant_id": 1,
		"type":      "MENU",
		"name":      "租户1菜单",
		"app_code":  "default",
	}
	jsonBody, _ := json.Marshal(body1)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 租户 2 创建资源
	body2 := map[string]interface{}{
		"tenant_id": 2,
		"type":      "MENU",
		"name":      "租户2菜单",
		"app_code":  "default",
	}
	jsonBody, _ = json.Marshal(body2)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/resources", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "2")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 查询租户 1 的树（通过 header 设置 tenant）
	req = httptest.NewRequest(http.MethodGet, "/api/v1/resources/tree", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Code int                    `json:"code"`
		Data []service.ResourceNode `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data) != 1 {
		t.Fatalf("期望租户1只有 1 个资源，实际: %d", len(resp.Data))
	}
	if resp.Data[0].Name != "租户1菜单" {
		t.Fatalf("期望资源名为 '租户1菜单'，实际: %s", resp.Data[0].Name)
	}

	// 查询租户 2 的树
	req = httptest.NewRequest(http.MethodGet, "/api/v1/resources/tree", nil)
	req.Header.Set("X-Test-Tenant-ID", "2")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data) != 1 {
		t.Fatalf("期望租户2只有 1 个资源，实际: %d", len(resp.Data))
	}
	if resp.Data[0].Name != "租户2菜单" {
		t.Fatalf("期望资源名为 '租户2菜单'，实际: %s", resp.Data[0].Name)
	}
}

// TestGetUserMenu 测试获取用户菜单树
func TestGetUserMenu(t *testing.T) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&model.Resource{}, &model.RoleResource{})
	if err != nil {
		t.Fatalf("资源表迁移失败: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mockAuthMiddleware())

	cfg := config.DefaultConfig()
	resourceRepo := repository.NewResourceRepository()
	svc := service.NewResourceService(db, cfg, resourceRepo)
	h := NewResourceHandler(svc)

	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	// 创建资源
	resource1 := &model.Resource{
		Type: "MENU", Name: "有权限菜单",
		Path: "/allowed", AppCode: "default", Status: 1, Version: 1,
	}
	resource2 := &model.Resource{
		Type: "MENU", Name: "无权限菜单",
		Path: "/denied", AppCode: "default", Status: 1, Version: 1,
	}
	db.Create(resource1)
	db.Create(resource2)

	// 创建角色
	role := &model.Role{
		TenantID: 1, RoleCode: "test_role", RoleName: "测试角色",
		RoleType: "NORMAL", Status: 1, Version: 1,
	}
	db.Create(role)

	// 角色关联资源（只关联 resource1）
	roleResource := &model.RoleResource{
		RoleID:     role.ID,
		ResourceID: resource1.ID,
	}
	db.Create(roleResource)

	// 创建用户角色关联
	userRole := &model.UserRole{
		UserID:   100,
		RoleID:   role.ID,
		TenantID: 1,
	}
	db.Create(userRole)

	// 查询用户菜单（通过 header 设置 tenant 和 user）
	req := httptest.NewRequest(http.MethodGet, "/api/v1/resources/user-menu", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	req.Header.Set("X-Test-User-ID", "100")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                    `json:"code"`
		Data []service.ResourceNode `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	// 只应返回有权限的菜单
	if len(resp.Data) != 1 {
		t.Fatalf("期望用户菜单数 1，实际: %d", len(resp.Data))
	}
	if resp.Data[0].Name != "有权限菜单" {
		t.Fatalf("期望菜单名 '有权限菜单'，实际: %s", resp.Data[0].Name)
	}
}
