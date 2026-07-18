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

// setupApiPermTestDB 创建测试数据库
func setupApiPermTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	err = db.AutoMigrate(&model.ApiPermission{})
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// setupApiPermRouter 创建测试路由（注入 mockAuthMiddleware）
func setupApiPermRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mockAuthMiddleware())

	repo := repository.NewApiPermissionRepository()
	svc := service.NewApiPermissionService(db, repo)
	h := NewApiPermissionHandler(svc)

	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	return r
}

// TestCreateGroup_Success 测试创建 GROUP 节点
func TestCreateGroup_Success(t *testing.T) {
	db := setupApiPermTestDB(t)
	r := setupApiPermRouter(db)

	body := map[string]interface{}{
		"tenant_id": 1,
		"type":      "GROUP",
		"name":      "用户管理",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                  `json:"code"`
		Data model.ApiPermission `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.Type != "GROUP" {
		t.Fatalf("期望 type=GROUP，实际: %s", resp.Data.Type)
	}
	if resp.Data.Status != "ACTIVE" {
		t.Fatalf("期望 status=ACTIVE，实际: %s", resp.Data.Status)
	}
}

// TestCreateEndpoint_Success 测试创建 ENDPOINT 节点
func TestCreateEndpoint_Success(t *testing.T) {
	db := setupApiPermTestDB(t)
	r := setupApiPermRouter(db)

	// 先创建 GROUP
	groupBody := map[string]interface{}{
		"tenant_id": 1,
		"type":      "GROUP",
		"name":      "用户管理",
	}
	jsonBody, _ := json.Marshal(groupBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var groupResp struct {
		Data model.ApiPermission `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &groupResp)
	groupID := groupResp.Data.ID

	// 创建 ENDPOINT
	endpointBody := map[string]interface{}{
		"tenant_id":       1,
		"parent_id":       groupID,
		"type":            "ENDPOINT",
		"name":            "获取用户列表",
		"permission_code": "user:list",
		"url_pattern":     "/api/v1/users",
		"http_method":     "GET",
		"status":          "ACTIVE",
	}
	jsonBody, _ = json.Marshal(endpointBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                  `json:"code"`
		Data model.ApiPermission `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.Type != "ENDPOINT" {
		t.Fatalf("期望 type=ENDPOINT，实际: %s", resp.Data.Type)
	}
	if resp.Data.URLPattern != "/api/v1/users" {
		t.Fatalf("期望 url_pattern=/api/v1/users，实际: %s", resp.Data.URLPattern)
	}
}

// TestGetTree 测试获取权限树
func TestGetTree(t *testing.T) {
	db := setupApiPermTestDB(t)
	r := setupApiPermRouter(db)

	// 创建 GROUP
	groupBody := map[string]interface{}{
		"tenant_id": 1,
		"type":      "GROUP",
		"name":      "用户管理",
	}
	jsonBody, _ := json.Marshal(groupBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var groupResp struct {
		Data model.ApiPermission `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &groupResp)
	groupID := groupResp.Data.ID

	// 创建子 ENDPOINT
	endpointBody := map[string]interface{}{
		"tenant_id":       1,
		"parent_id":       groupID,
		"type":            "ENDPOINT",
		"name":            "创建用户",
		"permission_code": "user:create",
		"url_pattern":     "/api/v1/users",
		"http_method":     "POST",
		"status":          "ACTIVE",
	}
	jsonBody, _ = json.Marshal(endpointBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 获取树（通过 header 设置 tenant）
	req = httptest.NewRequest(http.MethodGet, "/api/v1/api-permissions/tree", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var treeResp struct {
		Code int `json:"code"`
		Data []struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			Children []struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"children"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &treeResp)

	if treeResp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", treeResp.Code)
	}
	if len(treeResp.Data) != 1 {
		t.Fatalf("期望根节点数 1，实际: %d", len(treeResp.Data))
	}
	if treeResp.Data[0].Name != "用户管理" {
		t.Fatalf("期望根节点名称=用户管理，实际: %s", treeResp.Data[0].Name)
	}
	if len(treeResp.Data[0].Children) != 1 {
		t.Fatalf("期望子节点数 1，实际: %d", len(treeResp.Data[0].Children))
	}
}

// TestMoveEndpoint 测试移动节点到 GROUP 下
func TestMoveEndpoint(t *testing.T) {
	db := setupApiPermTestDB(t)
	r := setupApiPermRouter(db)

	// 创建 GROUP
	groupBody := map[string]interface{}{
		"tenant_id": 1,
		"type":      "GROUP",
		"name":      "角色管理",
	}
	jsonBody, _ := json.Marshal(groupBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var groupResp struct {
		Data model.ApiPermission `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &groupResp)
	groupID := groupResp.Data.ID

	// 创建未分组 ENDPOINT（无 parent_id）
	endpointBody := map[string]interface{}{
		"tenant_id":       1,
		"type":            "ENDPOINT",
		"name":            "获取角色列表",
		"permission_code": "role:list",
		"url_pattern":     "/api/v1/roles",
		"http_method":     "GET",
	}
	jsonBody, _ = json.Marshal(endpointBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var epResp struct {
		Data model.ApiPermission `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &epResp)
	epID := epResp.Data.ID

	// 验证初始状态为 UNASSIGNED
	if epResp.Data.Status != "UNASSIGNED" {
		t.Fatalf("期望初始 status=UNASSIGNED，实际: %s", epResp.Data.Status)
	}

	// 移动到 GROUP 下
	moveBody := map[string]interface{}{
		"parent_id": groupID,
	}
	jsonBody, _ = json.Marshal(moveBody)
	url := fmt.Sprintf("/api/v1/api-permissions/%d/move", epID)
	req = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证移动后状态
	var perm model.ApiPermission
	db.First(&perm, epID)
	if perm.Status != "ACTIVE" {
		t.Fatalf("期望移动后 status=ACTIVE，实际: %s", perm.Status)
	}
	if perm.ParentID == nil || *perm.ParentID != groupID {
		t.Fatalf("期望移动后 parent_id=%d，实际: %v", groupID, perm.ParentID)
	}
}

// TestListUnassigned 测试获取未分组 ENDPOINT 列表
func TestListUnassigned(t *testing.T) {
	db := setupApiPermTestDB(t)
	r := setupApiPermRouter(db)

	// 创建两个未分组 ENDPOINT
	for i := 0; i < 2; i++ {
		body := map[string]interface{}{
			"tenant_id":       1,
			"type":            "ENDPOINT",
			"name":            fmt.Sprintf("endpoint_%d", i),
			"permission_code": fmt.Sprintf("test:ep%d", i),
			"url_pattern":     fmt.Sprintf("/api/v1/test%d", i),
			"http_method":     "GET",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-Tenant-ID", "1")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("创建 endpoint_%d 失败: %d", i, w.Code)
		}
	}

	// 创建一个已分组的 ENDPOINT（有 parent_id，状态 ACTIVE）
	groupBody := map[string]interface{}{
		"tenant_id": 1,
		"type":      "GROUP",
		"name":      "分组",
	}
	jsonBody, _ := json.Marshal(groupBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var gResp struct {
		Data model.ApiPermission `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &gResp)

	assignedBody := map[string]interface{}{
		"tenant_id":   1,
		"parent_id":   gResp.Data.ID,
		"type":        "ENDPOINT",
		"name":        "已分组接口",
		"url_pattern": "/api/v1/assigned",
		"http_method": "POST",
		"status":      "ACTIVE",
	}
	jsonBody, _ = json.Marshal(assignedBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 查询未分组列表（通过 header 设置 tenant）
	req = httptest.NewRequest(http.MethodGet, "/api/v1/api-permissions/unassigned", nil)
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                    `json:"code"`
		Data []model.ApiPermission `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	// 只有 2 个未分组的
	if len(resp.Data) != 2 {
		t.Fatalf("期望未分组数 2，实际: %d", len(resp.Data))
	}
}

// TestDeleteGroupWithChildren 测试删除有子节点的 GROUP 返回 ErrProtectedEntity
func TestDeleteGroupWithChildren(t *testing.T) {
	db := setupApiPermTestDB(t)
	r := setupApiPermRouter(db)

	// 创建 GROUP
	groupBody := map[string]interface{}{
		"tenant_id": 1,
		"type":      "GROUP",
		"name":      "不可删分组",
	}
	jsonBody, _ := json.Marshal(groupBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var groupResp struct {
		Data model.ApiPermission `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &groupResp)
	groupID := groupResp.Data.ID

	// 创建子 ENDPOINT
	childBody := map[string]interface{}{
		"tenant_id":   1,
		"parent_id":   groupID,
		"type":        "ENDPOINT",
		"name":        "子接口",
		"url_pattern": "/api/v1/child",
		"http_method": "GET",
		"status":      "ACTIVE",
	}
	jsonBody, _ = json.Marshal(childBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/api-permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", "1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 尝试删除有子节点的 GROUP
	url := fmt.Sprintf("/api/v1/api-permissions/%d", groupID)
	req = httptest.NewRequest(http.MethodDelete, url, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40004 {
		t.Fatalf("期望 code=40004(ErrProtectedEntity)，实际: %d", resp.Code)
	}
}
