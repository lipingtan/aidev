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

// setupRoleApiTestDB 创建包含 RoleApi 和 ApiPermission 表的测试数据库
func setupRoleApiTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	err = db.AutoMigrate(
		&model.Tenant{},
		&model.Role{},
		&model.User{},
		&model.UserTenant{},
		&model.UserRole{},
		&model.TenantApp{},
		&model.ApiPermission{},
		&model.RoleApi{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// setupRoleApiRouter 创建角色接口权限测试路由
func setupRoleApiRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := config.DefaultConfig()
	roleRepo := repository.NewRoleRepository()
	svc := service.NewRoleService(db, cfg, roleRepo)
	h := NewRoleHandler(svc)

	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	return r
}

// createTestRole 辅助：在 DB 创建角色，返回角色 ID
func createTestRole(t *testing.T, db *gorm.DB, tenantID int64, code, name string) int64 {
	role := &model.Role{
		TenantID: tenantID,
		RoleCode: code,
		RoleName: name,
		RoleType: "NORMAL",
		Status:   1,
		Version:  1,
	}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}
	return role.ID
}

// createTestApiPermission 辅助：在 DB 创建接口权限节点，返回节点 ID
func createTestApiPermission(t *testing.T, db *gorm.DB, tenantID int64, parentID *int64, permType, name string) int64 {
	perm := &model.ApiPermission{
		ParentID: parentID,
		Type:     permType,
		Name:     name,
		AppCode:  "default",
		Status:   "ACTIVE",
	}
	if err := db.Create(perm).Error; err != nil {
		t.Fatalf("创建接口权限节点失败: %v", err)
	}
	return perm.ID
}

// TestAssignApis_BindEndpoint 测试绑定 ENDPOINT→admin_role_api 记录正确
func TestAssignApis_BindEndpoint(t *testing.T) {
	db := setupRoleApiTestDB(t)
	r := setupRoleApiRouter(db)

	tenantID := int64(100)
	roleID := createTestRole(t, db, tenantID, "api_role", "接口角色")

	// 创建 2 个 ENDPOINT 节点
	ep1 := createTestApiPermission(t, db, tenantID, nil, "ENDPOINT", "获取用户列表")
	ep2 := createTestApiPermission(t, db, tenantID, nil, "ENDPOINT", "创建用户")

	// 绑定 ENDPOINT
	body := map[string]interface{}{
		"api_permission_ids": []int64{ep1, ep2},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/roles/%d/apis", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 admin_role_api 表中有 2 条记录
	var roleApis []model.RoleApi
	db.Where("role_id = ?", roleID).Find(&roleApis)
	if len(roleApis) != 2 {
		t.Fatalf("期望 admin_role_api 有 2 条记录，实际: %d", len(roleApis))
	}

	// 验证绑定的 api_permission_id 正确
	idSet := make(map[int64]bool)
	for _, ra := range roleApis {
		idSet[ra.ApiPermissionID] = true
	}
	if !idSet[ep1] || !idSet[ep2] {
		t.Fatalf("绑定的 api_permission_id 不正确: %v", roleApis)
	}
}

// TestAssignApis_GroupAutoExpand 测试绑定 GROUP id→自动展开为子 ENDPOINT
func TestAssignApis_GroupAutoExpand(t *testing.T) {
	db := setupRoleApiTestDB(t)
	r := setupRoleApiRouter(db)

	tenantID := int64(100)
	roleID := createTestRole(t, db, tenantID, "group_role", "分组角色")

	// 创建 GROUP 节点
	groupID := createTestApiPermission(t, db, tenantID, nil, "GROUP", "用户管理")

	// 在 GROUP 下创建 2 个 ENDPOINT 子节点
	childEp1 := createTestApiPermission(t, db, tenantID, &groupID, "ENDPOINT", "获取用户")
	childEp2 := createTestApiPermission(t, db, tenantID, &groupID, "ENDPOINT", "删除用户")

	// 绑定 GROUP ID（应自动展开为子 ENDPOINT）
	body := map[string]interface{}{
		"api_permission_ids": []int64{groupID},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/roles/%d/apis", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 admin_role_api 表中有 2 条子 ENDPOINT 记录（GROUP 本身不出现）
	var roleApis []model.RoleApi
	db.Where("role_id = ?", roleID).Find(&roleApis)
	if len(roleApis) != 2 {
		t.Fatalf("期望 admin_role_api 有 2 条记录（展开后的 ENDPOINT），实际: %d", len(roleApis))
	}

	// 验证绑定的是子 ENDPOINT 的 ID
	idSet := make(map[int64]bool)
	for _, ra := range roleApis {
		idSet[ra.ApiPermissionID] = true
	}
	if !idSet[childEp1] || !idSet[childEp2] {
		t.Fatalf("展开后的 api_permission_id 不正确，期望 %d 和 %d，实际: %v", childEp1, childEp2, roleApis)
	}
	// GROUP 本身不应出现在绑定中
	if idSet[groupID] {
		t.Fatalf("GROUP 节点不应出现在 admin_role_api 绑定中")
	}
}

// TestAssignApis_EmptyArray 测试绑定空数组→旧记录全部清除（全量替换）
func TestAssignApis_EmptyArray(t *testing.T) {
	db := setupRoleApiTestDB(t)
	r := setupRoleApiRouter(db)

	tenantID := int64(100)
	roleID := createTestRole(t, db, tenantID, "clear_role", "清除角色")

	// 创建 ENDPOINT 并先绑定
	ep1 := createTestApiPermission(t, db, tenantID, nil, "ENDPOINT", "旧接口1")
	ep2 := createTestApiPermission(t, db, tenantID, nil, "ENDPOINT", "旧接口2")

	// 第一次绑定
	body := map[string]interface{}{
		"api_permission_ids": []int64{ep1, ep2},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/roles/%d/apis", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("第一次绑定失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证有 2 条记录
	var count int64
	db.Model(&model.RoleApi{}).Where("role_id = ?", roleID).Count(&count)
	if count != 2 {
		t.Fatalf("第一次绑定后期望 2 条记录，实际: %d", count)
	}

	// 第二次绑定空数组
	body = map[string]interface{}{
		"api_permission_ids": []int64{},
	}
	jsonBody, _ = json.Marshal(body)
	req = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("清除绑定失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证记录全部清除
	db.Model(&model.RoleApi{}).Where("role_id = ?", roleID).Count(&count)
	if count != 0 {
		t.Fatalf("清除后期望 0 条记录，实际: %d", count)
	}
}
