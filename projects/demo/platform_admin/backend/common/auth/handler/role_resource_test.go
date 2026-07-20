package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-admin/common/auth/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// setupRoleResourceRouter 创建角色资源分配测试路由（额外迁移 Resource 和 RoleResource 表）
func setupRoleResourceRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	db := setupTestDB(t)
	// 额外迁移 Resource、RoleResource、RoleApp 表
	err := db.AutoMigrate(&model.Resource{}, &model.RoleResource{}, &model.RoleApp{})
	if err != nil {
		t.Fatalf("资源表迁移失败: %v", err)
	}
	r := setupRoleRouter(db)
	return r, db
}

// TestAssignResources_Success 测试分配菜单权限成功，admin_role_resource 记录正确
func TestAssignResources_Success(t *testing.T) {
	r, db := setupRoleResourceRouter(t)

	tenantID := int64(100)
	roleID := createRole(t, r, tenantID, "res_role", "资源角色", nil)

	// 创建资源（同租户）
	res1 := &model.Resource{Type: "menu", Name: "菜单1", AppCode: "default"}
	res2 := &model.Resource{Type: "menu", Name: "菜单2", AppCode: "default"}
	db.Create(res1)
	db.Create(res2)

	// 分配资源
	body := map[string]interface{}{
		"resource_ids": []int64{res1.ID, res2.ID},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/admin/roles/%d/resources", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 admin_role_resource 记录
	var records []model.RoleResource
	db.Where("role_id = ?", roleID).Find(&records)
	if len(records) != 2 {
		t.Fatalf("期望 2 条记录，实际: %d", len(records))
	}

	// 验证资源 ID 正确
	ids := make(map[int64]bool)
	for _, rec := range records {
		ids[rec.ResourceID] = true
	}
	if !ids[res1.ID] || !ids[res2.ID] {
		t.Fatalf("记录中的 resource_id 不匹配，期望 %v 和 %v", res1.ID, res2.ID)
	}
}

// TestAssignResources_CrossTenant 测试分配不同 app_code 的资源→成功（资源是全局的）
func TestAssignResources_CrossTenant(t *testing.T) {
	r, db := setupRoleResourceRouter(t)

	tenantA := int64(100)
	roleID := createRole(t, r, tenantA, "cross_role", "跨应用角色", nil)

	// 创建 other_app 的资源
	resCross := &model.Resource{Type: "menu", Name: "其他应用菜单", AppCode: "other_app"}
	db.Create(resCross)

	// 分配其他应用资源→应成功（自动绑定对应 app）
	body := map[string]interface{}{
		"resource_ids": []int64{resCross.ID},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/admin/roles/%d/resources", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证角色自动绑定了 other_app
	var roleApps []model.RoleApp
	db.Where("role_id = ?", roleID).Find(&roleApps)
	if len(roleApps) != 1 || roleApps[0].AppCode != "other_app" {
		t.Fatalf("期望角色自动绑定 other_app，实际: %+v", roleApps)
	}
}

// TestAssignResources_FullReplace 测试全量替换→旧记录清除
func TestAssignResources_FullReplace(t *testing.T) {
	r, db := setupRoleResourceRouter(t)

	tenantID := int64(100)
	roleID := createRole(t, r, tenantID, "replace_role", "替换角色", nil)

	// 创建 3 个资源
	res1 := &model.Resource{Type: "menu", Name: "菜单A", AppCode: "default"}
	res2 := &model.Resource{Type: "menu", Name: "菜单B", AppCode: "default"}
	res3 := &model.Resource{Type: "menu", Name: "菜单C", AppCode: "default"}
	db.Create(res1)
	db.Create(res2)
	db.Create(res3)

	// 第一次分配：res1, res2
	body := map[string]interface{}{
		"resource_ids": []int64{res1.ID, res2.ID},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/admin/roles/%d/resources", roleID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("第一次分配失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 第二次分配：res2, res3（全量替换，res1 应被清除）
	body = map[string]interface{}{
		"resource_ids": []int64{res2.ID, res3.ID},
	}
	jsonBody, _ = json.Marshal(body)
	req = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("第二次分配失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证：只剩 res2, res3
	var records []model.RoleResource
	db.Where("role_id = ?", roleID).Find(&records)
	if len(records) != 2 {
		t.Fatalf("期望 2 条记录，实际: %d", len(records))
	}

	ids := make(map[int64]bool)
	for _, rec := range records {
		ids[rec.ResourceID] = true
	}
	if ids[res1.ID] {
		t.Fatalf("旧记录 res1 未被清除")
	}
	if !ids[res2.ID] || !ids[res3.ID] {
		t.Fatalf("新记录不正确，期望 res2=%d 和 res3=%d", res2.ID, res3.ID)
	}
}
