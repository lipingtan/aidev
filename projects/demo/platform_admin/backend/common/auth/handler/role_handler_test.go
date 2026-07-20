package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"go-admin/common/auth/config"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// setupRoleRouter 创建角色测试路由（注入 mockAuthMiddleware）
func setupRoleRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mockAuthMiddleware())

	cfg := config.DefaultConfig()
	roleRepo := repository.NewRoleRepository()
	svc := service.NewRoleService(db, cfg, roleRepo)
	h := NewRoleHandler(svc)

	api := r.Group("/api/v1/admin")
	h.RegisterRoutes(api)

	return r
}

// createRole 辅助函数：创建角色并返回角色 ID
func createRole(t *testing.T, r *gin.Engine, tenantID int64, roleCode, roleName string, parentID *int64) int64 {
	body := map[string]interface{}{
		"tenant_id": tenantID,
		"role_code": roleCode,
		"role_name": roleName,
	}
	if parentID != nil {
		body["parent_id"] = fmt.Sprintf("%d", *parentID)
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", strconv.FormatInt(tenantID, 10))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("创建角色失败: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id,string"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.ID
}

// TestCreateRole_Success 测试创建角色成功，关联正确 tenant_id
func TestCreateRole_Success(t *testing.T) {
	db := setupTestDB(t)
	r := setupRoleRouter(db)

	tenantID := int64(100)
	body := map[string]interface{}{
		"tenant_id": tenantID,
		"role_code": "editor",
		"role_name": "编辑者",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", strconv.FormatInt(tenantID, 10))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int        `json:"code"`
		Data model.Role `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.TenantID != tenantID {
		t.Fatalf("期望 tenant_id=%d，实际: %d", tenantID, resp.Data.TenantID)
	}
	if resp.Data.RoleCode != "editor" {
		t.Fatalf("期望 role_code=editor，实际: %s", resp.Data.RoleCode)
	}
}

// TestCreateRole_DuplicateCode 测试重复 role_code + tenant_id 返回 400
func TestCreateRole_DuplicateCode(t *testing.T) {
	db := setupTestDB(t)
	r := setupRoleRouter(db)

	tenantID := int64(100)
	createRole(t, r, tenantID, "dup_role", "重复角色", nil)

	// 第二次创建相同 tenant_id + role_code
	body := map[string]interface{}{
		"tenant_id": tenantID,
		"role_code": "dup_role",
		"role_name": "重复角色2",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", strconv.FormatInt(tenantID, 10))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40001 {
		t.Fatalf("期望 code=40001，实际: %d", resp.Code)
	}
}

// TestCyclicHierarchy 测试循环引用检测：A.parent=B, B.parent=A 返回 400
func TestCyclicHierarchy(t *testing.T) {
	db := setupTestDB(t)
	r := setupRoleRouter(db)

	tenantID := int64(100)

	// 创建角色 A
	roleA := createRole(t, r, tenantID, "role_a", "角色A", nil)
	// 创建角色 B，parent=A
	roleB := createRole(t, r, tenantID, "role_b", "角色B", &roleA)

	// 尝试更新角色 A 的 parent 为 B（形成循环）
	updateBody := map[string]interface{}{
		"parent_id": fmt.Sprintf("%d", roleB),
		"version":   1,
	}
	jsonBody, _ := json.Marshal(updateBody)
	url := fmt.Sprintf("/api/v1/admin/roles/%d", roleA)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400（循环引用），实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40003 {
		t.Fatalf("期望 code=40003（ErrCyclicHierarchy），实际: %d", resp.Code)
	}
}

// TestMaxHierarchyDepth 测试层级深度超限返回 400
func TestMaxHierarchyDepth(t *testing.T) {
	db := setupTestDB(t)

	// 使用 MaxDepth=3 以便快速测试
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(mockAuthMiddleware())
	cfg := config.DefaultConfig()
	cfg.Role.MaxDepth = 3
	roleRepo := repository.NewRoleRepository()
	svc := service.NewRoleService(db, cfg, roleRepo)
	h := NewRoleHandler(svc)
	api := router.Group("/api/v1/admin")
	h.RegisterRoutes(api)

	tenantID := int64(100)

	// 创建 3 层角色链：root -> child1 -> child2
	rootID := createRoleWith(t, router, tenantID, "root", "根角色", nil)
	child1ID := createRoleWith(t, router, tenantID, "child1", "子角色1", &rootID)
	child2ID := createRoleWith(t, router, tenantID, "child2", "子角色2", &child1ID)

	// 尝试创建第 4 层角色（parent=child2），深度超过 MaxDepth=3
	body := map[string]interface{}{
		"tenant_id": tenantID,
		"role_code": "child3",
		"role_name": "子角色3",
		"parent_id": fmt.Sprintf("%d", child2ID),
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", strconv.FormatInt(tenantID, 10))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400（深度超限），实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40007 {
		t.Fatalf("期望 code=40007（ErrMaxHierarchyDepth），实际: %d", resp.Code)
	}
}

// createRoleWith 辅助函数：在指定 router 上创建角色
func createRoleWith(t *testing.T, router *gin.Engine, tenantID int64, roleCode, roleName string, parentID *int64) int64 {
	body := map[string]interface{}{
		"tenant_id": tenantID,
		"role_code": roleCode,
		"role_name": roleName,
	}
	if parentID != nil {
		body["parent_id"] = fmt.Sprintf("%d", *parentID)
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Tenant-ID", strconv.FormatInt(tenantID, 10))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("创建角色 %s 失败: %d, body: %s", roleCode, w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id,string"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.ID
}

// TestDeleteRole_HasUsers 测试有用户绑定时删除返回 400 (code=40006)
func TestDeleteRole_HasUsers(t *testing.T) {
	db := setupTestDB(t)
	r := setupRoleRouter(db)

	tenantID := int64(100)
	roleID := createRole(t, r, tenantID, "bound_role", "绑定角色", nil)

	// 手动在 admin_user_role 表中创建用户绑定
	userRole := &model.UserRole{
		UserID:   1,
		RoleID:   roleID,
		TenantID: tenantID,
	}
	if err := db.Create(userRole).Error; err != nil {
		t.Fatalf("创建用户-角色绑定失败: %v", err)
	}

	// 尝试删除角色
	url := fmt.Sprintf("/api/v1/admin/roles/%d", roleID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400（有用户绑定），实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40006 {
		t.Fatalf("期望 code=40006（ErrRoleHasUsers），实际: %d", resp.Code)
	}
}

// TestDeleteRole_Success 测试无用户绑定时删除成功
func TestDeleteRole_Success(t *testing.T) {
	db := setupTestDB(t)
	r := setupRoleRouter(db)

	tenantID := int64(100)
	roleID := createRole(t, r, tenantID, "free_role", "无绑定角色", nil)

	// 直接删除角色
	url := fmt.Sprintf("/api/v1/admin/roles/%d", roleID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
}

// TestTenantIsolation 测试租户隔离：租户 A 看不到租户 B 的角色
func TestTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	r := setupRoleRouter(db)

	tenantA := int64(100)
	tenantB := int64(200)

	// 在租户 A 下创建角色
	createRole(t, r, tenantA, "role_in_a", "租户A角色", nil)
	// 在租户 B 下创建角色
	createRole(t, r, tenantB, "role_in_b", "租户B角色", nil)

	// 查询租户 A 的角色列表（通过 header 设置 tenant）
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	req.Header.Set("X-Test-Tenant-ID", strconv.FormatInt(tenantA, 10))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("查询失败: %d", w.Code)
	}

	var resp struct {
		Code int          `json:"code"`
		Data []model.Role `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// 验证仅返回租户 A 的角色
	for _, role := range resp.Data {
		if role.TenantID != tenantA {
			t.Fatalf("租户隔离失败：查询租户 A 但返回了 tenant_id=%d 的角色", role.TenantID)
		}
	}
	if len(resp.Data) != 1 {
		t.Fatalf("期望租户 A 有 1 个角色，实际: %d", len(resp.Data))
	}

	// 查询租户 B 的角色列表
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	req.Header.Set("X-Test-Tenant-ID", strconv.FormatInt(tenantB, 10))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Data) != 1 {
		t.Fatalf("期望租户 B 有 1 个角色，实际: %d", len(resp.Data))
	}
	if resp.Data[0].RoleCode != "role_in_b" {
		t.Fatalf("期望角色编码 role_in_b，实际: %s", resp.Data[0].RoleCode)
	}
}

// TestUpdateRole_OptimisticLock 测试乐观锁更新
func TestUpdateRole_OptimisticLock(t *testing.T) {
	db := setupTestDB(t)
	r := setupRoleRouter(db)

	tenantID := int64(100)
	roleID := createRole(t, r, tenantID, "lock_role", "乐观锁角色", nil)

	// 使用正确版本号更新
	updateBody := map[string]interface{}{
		"role_name": "更新后角色",
		"version":   1,
	}
	jsonBody, _ := json.Marshal(updateBody)
	url := fmt.Sprintf("/api/v1/admin/roles/%d", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("更新失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 使用旧版本号再次更新，应该失败
	updateBody["version"] = 1 // 旧版本
	jsonBody, _ = json.Marshal(updateBody)
	req = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望乐观锁冲突返回 400，实际: %d, body: %s", w.Code, w.Body.String())
	}
}
