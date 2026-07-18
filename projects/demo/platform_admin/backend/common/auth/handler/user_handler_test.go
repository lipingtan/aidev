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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupUserRouter 创建用户管理测试路由（注入 mockAuthMiddleware）
func setupUserRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mockAuthMiddleware())

	cfg := config.DefaultConfig()
	userRepo := repository.NewUserRepository()
	svc := service.NewUserService(db, cfg, userRepo)
	h := NewUserHandler(svc)

	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	return r
}

// createUserViaAPI 辅助：通过 API 创建测试用户，返回用户 ID
func createUserViaAPI(t *testing.T, r *gin.Engine, username, password string) int64 {
	body := map[string]interface{}{
		"username": username,
		"password": password,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("创建用户失败: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.ID
}

// createTenantForUser 辅助：直接在 DB 创建租户，返回租户 ID
func createTenantForUser(t *testing.T, db *gorm.DB, code, name string) int64 {
	tenant := &model.Tenant{
		TenantCode: code,
		Name:       name,
		Status:     1,
		Version:    1,
	}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("创建租户失败: %v", err)
	}
	return tenant.ID
}

// TestCreateUser_PasswordEncrypted 测试创建用户→密码不明文存储
func TestCreateUser_PasswordEncrypted(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRouter(db)

	plainPassword := "MySecret123"
	userID := createUserViaAPI(t, r, "test_user", plainPassword)

	// 验证数据库中密码不是明文
	var user model.User
	db.Unscoped().First(&user, userID)

	if user.Password == plainPassword {
		t.Fatal("密码以明文存储，不安全")
	}

	// 验证密码是 bcrypt 哈希
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainPassword))
	if err != nil {
		t.Fatalf("密码 bcrypt 验证失败: %v", err)
	}
}

// TestCreateUser_DuplicateUsername 测试重复用户名返回错误
func TestCreateUser_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRouter(db)

	createUserViaAPI(t, r, "dup_user", "pass123")

	// 第二次创建同名用户
	body := map[string]interface{}{
		"username": "dup_user",
		"password": "pass456",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestAssociateTenant_Creates_UserTenant 测试关联租户→admin_user_tenant 记录创建
func TestAssociateTenant_Creates_UserTenant(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRouter(db)

	userID := createUserViaAPI(t, r, "assoc_user", "pass123")
	tenantID := createTenantForUser(t, db, "corp_a", "公司A")

	// 关联用户到租户
	body := map[string]interface{}{
		"tenant_id": tenantID,
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/users/%d/tenants", userID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("关联租户失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 admin_user_tenant 记录已创建
	var count int64
	db.Model(&model.UserTenant{}).Where("user_id = ? AND tenant_id = ?", userID, tenantID).Count(&count)
	if count != 1 {
		t.Fatalf("期望 admin_user_tenant 记录数为 1，实际: %d", count)
	}
}

// TestListUsers_TenantContext 测试租户上下文查询→仅返回已关联用户
func TestListUsers_TenantContext(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRouter(db)

	// 创建两个用户
	userID1 := createUserViaAPI(t, r, "user_in_tenant", "pass123")
	createUserViaAPI(t, r, "user_not_in_tenant", "pass456")

	// 创建租户并关联第一个用户
	tenantID := createTenantForUser(t, db, "filter_corp", "筛选公司")
	ut := &model.UserTenant{UserID: userID1, TenantID: tenantID}
	db.Create(ut)

	// 按租户上下文查询用户列表（通过 header 设置 tenant）
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1&page_size=20", nil)
	req.Header.Set("X-Test-Tenant-ID", fmt.Sprintf("%d", tenantID))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("查询用户列表失败: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List     []model.User `json:"list"`
			Total    int64        `json:"total"`
			Page     int          `json:"page"`
			PageSize int          `json:"page_size"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.Total != 1 {
		t.Fatalf("期望 total=1（仅关联用户），实际: %d", resp.Data.Total)
	}
	if len(resp.Data.List) != 1 {
		t.Fatalf("期望列表长度 1，实际: %d", len(resp.Data.List))
	}
	if resp.Data.List[0].Username != "user_in_tenant" {
		t.Fatalf("期望返回 user_in_tenant，实际: %s", resp.Data.List[0].Username)
	}
}

// TestUpdateUser_OptimisticLock 测试乐观锁冲突→返回错误
func TestUpdateUser_OptimisticLock(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRouter(db)

	userID := createUserViaAPI(t, r, "lock_user", "pass123")

	// 用错误的 version 更新（用户创建后 version=1，传 version=99 模拟冲突）
	body := map[string]interface{}{
		"nickname": "新昵称",
		"version":  99,
	}
	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("/api/v1/users/%d", userID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400（乐观锁冲突），实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40001 {
		t.Fatalf("期望错误码 40001，实际: %d", resp.Code)
	}
}

// TestDeleteUser_SoftDelete 测试软删除→记录仍存在但 deleted_at 有值
func TestDeleteUser_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRouter(db)

	userID := createUserViaAPI(t, r, "delete_user", "pass123")

	// 软删除
	url := fmt.Sprintf("/api/v1/users/%d", userID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("软删除失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证记录仍存在但 deleted_at 有值
	var user model.User
	result := db.Unscoped().First(&user, userID)
	if result.Error != nil {
		t.Fatalf("Unscoped 查询失败: %v", result.Error)
	}
	if !user.DeletedAt.Valid {
		t.Fatal("期望 deleted_at 有值（软删除），实际为空")
	}

	// 正常查询应查不到
	var normalUser model.User
	result = db.First(&normalUser, userID)
	if result.Error == nil {
		t.Fatal("正常查询不应该能找到已软删除的用户")
	}
}

// TestDissociateTenant_CascadeDeleteRoles 测试解除关联→级联删除该租户下角色绑定
func TestDissociateTenant_CascadeDeleteRoles(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRouter(db)

	userID := createUserViaAPI(t, r, "cascade_user", "pass123")
	tenantID := createTenantForUser(t, db, "cascade_corp", "级联公司")

	// 关联用户到租户
	ut := &model.UserTenant{UserID: userID, TenantID: tenantID}
	db.Create(ut)

	// 创建角色绑定（模拟用户在该租户下有角色）
	userRole := &model.UserRole{
		UserID:   userID,
		RoleID:   999,
		TenantID: tenantID,
	}
	db.Create(userRole)

	// 验证角色绑定存在
	var roleCount int64
	db.Model(&model.UserRole{}).Where("user_id = ? AND tenant_id = ?", userID, tenantID).Count(&roleCount)
	if roleCount != 1 {
		t.Fatalf("期望角色绑定数为 1，实际: %d", roleCount)
	}

	// 解除关联
	url := fmt.Sprintf("/api/v1/users/%d/tenants/%d", userID, tenantID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("解除关联失败: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证用户-租户关联已删除
	var utCount int64
	db.Model(&model.UserTenant{}).Where("user_id = ? AND tenant_id = ?", userID, tenantID).Count(&utCount)
	if utCount != 0 {
		t.Fatalf("期望 user_tenant 记录数为 0，实际: %d", utCount)
	}

	// 验证角色绑定已级联删除
	db.Model(&model.UserRole{}).Where("user_id = ? AND tenant_id = ?", userID, tenantID).Count(&roleCount)
	if roleCount != 0 {
		t.Fatalf("期望角色绑定已级联删除，实际仍有: %d", roleCount)
	}
}

// TestListUserTenants 测试查询用户已关联租户列表
func TestListUserTenants(t *testing.T) {
	db := setupTestDB(t)
	r := setupUserRouter(db)

	userID := createUserViaAPI(t, r, "multi_tenant_user", "pass123")
	tenantID1 := createTenantForUser(t, db, "corp_x", "公司X")
	tenantID2 := createTenantForUser(t, db, "corp_y", "公司Y")

	// 关联到两个租户
	db.Create(&model.UserTenant{UserID: userID, TenantID: tenantID1})
	db.Create(&model.UserTenant{UserID: userID, TenantID: tenantID2})

	// 查询用户租户列表
	url := fmt.Sprintf("/api/v1/users/%d/tenants", userID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("查询用户租户列表失败: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                `json:"code"`
		Data []model.UserTenant `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Data) != 2 {
		t.Fatalf("期望关联租户数为 2，实际: %d", len(resp.Data))
	}
}
