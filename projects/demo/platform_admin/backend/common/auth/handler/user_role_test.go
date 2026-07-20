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
	"gorm.io/gorm"
)

// setupUserRoleRouter 创建包含 UserRoleService 的用户管理测试路由
func setupUserRoleRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := config.DefaultConfig()
	userRepo := repository.NewUserRepository()
	userSvc := service.NewUserService(db, cfg, userRepo)
	roleSvc := service.NewUserRoleService(db, cfg)

	h := NewUserHandler(userSvc)
	h.SetRoleService(roleSvc)

	api := r.Group("/api/v1/admin")
	h.RegisterRoutes(api)

	return r
}

// createRoleInDB 辅助：在 DB 创建角色，返回角色 ID
func createRoleInDB(t *testing.T, db *gorm.DB, tenantID int64, code, name string) int64 {
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

// createUserViaRoleAPI 辅助：通过 API 创建测试用户（用于角色测试路由器），返回用户 ID
func createUserViaRoleAPI(t *testing.T, r *gin.Engine, username, password string) int64 {
	body := map[string]interface{}{
		"username": username,
		"password": password,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("创建用户失败: %d, body: %s", w.Code, w.Body.String())
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

// TestAssignRoles_Success 测试分配角色→admin_user_role 记录创建
func TestAssignRoles_Success(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRoleRouter(db)

	userID := createUserViaRoleAPI(t, r, "role_user", "pass123")
	tenantID := createTenantForUser(t, db, "role_corp", "角色公司")
	roleID := createRoleInDB(t, db, tenantID, "editor", "编辑者")

	// 关联用户到租户
	db.Create(&model.UserTenant{UserID: userID, TenantID: tenantID})

	// 分配角色
	body := map[string]interface{}{
		"tenant_id": fmt.Sprintf("%d", tenantID),
		"roles": []map[string]interface{}{
			{"role_id": fmt.Sprintf("%d", roleID)},
		},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/admin/users/%d/roles", userID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("分配角色失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 admin_user_role 记录已创建
	var count int64
	db.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ? AND tenant_id = ?", userID, roleID, tenantID).Count(&count)
	if count != 1 {
		t.Fatalf("期望 admin_user_role 记录数为 1，实际: %d", count)
	}
}

// TestAssignRoles_UserNotInTenant 测试用户未关联租户时分配→返回 400 (code=40008)
func TestAssignRoles_UserNotInTenant(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRoleRouter(db)

	userID := createUserViaRoleAPI(t, r, "no_tenant_user", "pass123")
	tenantID := createTenantForUser(t, db, "no_assoc_corp", "未关联公司")
	roleID := createRoleInDB(t, db, tenantID, "viewer", "查看者")

	// 不关联用户到租户，直接分配角色
	body := map[string]interface{}{
		"tenant_id": fmt.Sprintf("%d", tenantID),
		"roles": []map[string]interface{}{
			{"role_id": fmt.Sprintf("%d", roleID)},
		},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/admin/users/%d/roles", userID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40008 {
		t.Fatalf("期望错误码 40008，实际: %d", resp.Code)
	}
}

// TestAssignRoles_RoleNotInTenant 测试角色不属于当前租户→返回 400
func TestAssignRoles_RoleNotInTenant(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRoleRouter(db)

	userID := createUserViaRoleAPI(t, r, "wrong_role_user", "pass123")
	tenantID := createTenantForUser(t, db, "my_corp", "我的公司")
	otherTenantID := createTenantForUser(t, db, "other_corp", "其他公司")
	roleID := createRoleInDB(t, db, otherTenantID, "other_role", "其他角色")

	// 关联用户到租户
	db.Create(&model.UserTenant{UserID: userID, TenantID: tenantID})

	// 尝试分配其他租户的角色
	body := map[string]interface{}{
		"tenant_id": fmt.Sprintf("%d", tenantID),
		"roles": []map[string]interface{}{
			{"role_id": fmt.Sprintf("%d", roleID)},
		},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/admin/users/%d/roles", userID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40002 {
		t.Fatalf("期望错误码 40002，实际: %d", resp.Code)
	}
}

// TestReplaceRoles_Success 测试全量替换→旧记录删除、新记录创建
func TestReplaceRoles_Success(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRoleRouter(db)

	userID := createUserViaRoleAPI(t, r, "replace_user", "pass123")
	tenantID := createTenantForUser(t, db, "replace_corp", "替换公司")
	oldRoleID := createRoleInDB(t, db, tenantID, "old_role", "旧角色")
	newRoleID := createRoleInDB(t, db, tenantID, "new_role", "新角色")

	// 关联用户到租户
	db.Create(&model.UserTenant{UserID: userID, TenantID: tenantID})

	// 先分配旧角色
	db.Create(&model.UserRole{UserID: userID, RoleID: oldRoleID, TenantID: tenantID})

	// 验证旧角色存在
	var count int64
	db.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ? AND tenant_id = ?", userID, oldRoleID, tenantID).Count(&count)
	if count != 1 {
		t.Fatalf("旧角色应存在，实际: %d", count)
	}

	// 全量替换为新角色
	body := map[string]interface{}{
		"tenant_id": fmt.Sprintf("%d", tenantID),
		"roles": []map[string]interface{}{
			{"role_id": fmt.Sprintf("%d", newRoleID)},
		},
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/admin/users/%d/roles", userID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("全量替换失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证旧角色已删除
	db.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ? AND tenant_id = ?", userID, oldRoleID, tenantID).Count(&count)
	if count != 0 {
		t.Fatalf("旧角色应已删除，实际: %d", count)
	}

	// 验证新角色已创建
	db.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ? AND tenant_id = ?", userID, newRoleID, tenantID).Count(&count)
	if count != 1 {
		t.Fatalf("新角色应已创建，实际: %d", count)
	}
}
