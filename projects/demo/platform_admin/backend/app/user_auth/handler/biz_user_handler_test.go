package handler

import (
	"fmt"
	"net/http"
	"testing"

	"go-admin/app/user_auth/dto"
	"go-admin/app/user_auth/repository"
	"go-admin/app/user_auth/service"
	"go-admin/common/auth/middleware"

	"github.com/gin-gonic/gin"
)

// setupBizUserHandler 创建管理端 biz_user handler 测试环境
func setupBizUserHandler(t *testing.T) (*BizUserHandler, *gin.Engine, *service.BizUserService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	repo := repository.NewBizUserRepository()
	bizUserSvc := service.NewBizUserService(db, repo)
	handler := NewBizUserHandler(bizUserSvc)

	router := gin.New()

	// 模拟管理端认证中间件，注入 AuthContext（tenant_id=1, user_id=100）
	router.Use(func(c *gin.Context) {
		middleware.SetAuthContext(c, &middleware.AuthContext{
			UserID:   100,
			TenantID: 1,
			UserPool: "admin",
		})
		c.Next()
	})

	handler.RegisterRoutes(router.Group("/api/v1/admin"))

	return handler, router, bizUserSvc
}

// seedBizUsers 预置多条测试用户数据
func seedBizUsers(t *testing.T, svc *service.BizUserService) {
	t.Helper()
	// 插入 tenant_id=1 的用户
	for i := 0; i < 3; i++ {
		_, err := svc.CreateBizUser(&dto.CreateBizUserRequest{
			TenantID: 1,
			Phone:    fmt.Sprintf("1380013800%d", i),
			Nickname: fmt.Sprintf("用户%d", i),
		})
		if err != nil {
			t.Fatalf("预置用户数据失败: %v", err)
		}
	}
	// 插入 tenant_id=2 的用户（不应被 tenant_id=1 查到）
	_, err := svc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: 2,
		Phone:    "13900139000",
		Nickname: "其他租户用户",
	})
	if err != nil {
		t.Fatalf("预置其他租户用户失败: %v", err)
	}
}

// TestBizUser_List_TenantIsolation 分页列表 + tenant_id 隔离
func TestBizUser_List_TenantIsolation(t *testing.T) {
	_, router, svc := setupBizUserHandler(t)
	seedBizUsers(t, svc)

	w := performRequest(router, "GET", "/api/v1/admin/biz-users?page=1&page_size=20", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	if code := int(resp["code"].(float64)); code != 0 {
		t.Fatalf("期望 code=0，实际: %d", code)
	}

	data := resp["data"].(map[string]interface{})
	total := int(data["total"].(float64))
	if total != 3 {
		t.Fatalf("期望 total=3（仅 tenant_id=1），实际: %d", total)
	}

	list := data["list"].([]interface{})
	if len(list) != 3 {
		t.Fatalf("期望 list 长度=3，实际: %d", len(list))
	}
}

// TestBizUser_GetByID_Success 获取详情 → 200 + 正确数据
func TestBizUser_GetByID_Success(t *testing.T) {
	_, router, svc := setupBizUserHandler(t)

	// 创建一个用户
	user, err := svc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: 1,
		Phone:    "13800138000",
		Nickname: "测试详情用户",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	path := fmt.Sprintf("/api/v1/admin/biz-users/%d", user.ID)
	w := performRequest(router, "GET", path, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	// ID 以 string 返回（JSON tag 中 ,string）
	if data["phone"] != "13800138000" {
		t.Fatalf("期望 phone=13800138000，实际: %v", data["phone"])
	}
	if data["nickname"] != "测试详情用户" {
		t.Fatalf("期望 nickname=测试详情用户，实际: %v", data["nickname"])
	}
}

// TestBizUser_GetByID_NotFound 获取不存在的用户 → 404
func TestBizUser_GetByID_NotFound(t *testing.T) {
	_, router, _ := setupBizUserHandler(t)

	w := performRequest(router, "GET", "/api/v1/admin/biz-users/999999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("期望 404，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestBizUser_ResetPassword 重置密码 → 响应含明文密码
func TestBizUser_ResetPassword(t *testing.T) {
	_, router, svc := setupBizUserHandler(t)

	user, _ := svc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: 1,
		Phone:    "13800138001",
		Password: "old_password",
	})

	path := fmt.Sprintf("/api/v1/admin/biz-users/%d/reset-password", user.ID)
	w := performRequest(router, "POST", path, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	password, ok := data["password"].(string)
	if !ok || len(password) == 0 {
		t.Fatal("响应中未包含明文密码")
	}
	if len(password) != 8 {
		t.Fatalf("期望密码长度=8，实际: %d", len(password))
	}
}

// TestBizUser_ForceLogout DB token_version 递增
func TestBizUser_ForceLogout(t *testing.T) {
	_, router, svc := setupBizUserHandler(t)

	user, _ := svc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: 1,
		Phone:    "13800138002",
	})

	// 记录初始 token_version
	originalUser, _ := svc.GetBizUser(user.ID)
	originalVersion := originalUser.TokenVersion

	path := fmt.Sprintf("/api/v1/admin/biz-users/%d/force-logout", user.ID)
	w := performRequest(router, "POST", path, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 DB 中 token_version 已递增
	updatedUser, _ := svc.GetBizUser(user.ID)
	if updatedUser.TokenVersion != originalVersion+1 {
		t.Fatalf("期望 token_version=%d，实际: %d", originalVersion+1, updatedUser.TokenVersion)
	}
}

// TestBizUser_ToggleStatus DB status 变更
func TestBizUser_ToggleStatus(t *testing.T) {
	_, router, svc := setupBizUserHandler(t)

	user, _ := svc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: 1,
		Phone:    "13800138003",
	})

	// 默认 status=1，切换到 0
	path := fmt.Sprintf("/api/v1/admin/biz-users/%d/toggle-status", user.ID)
	w := performRequest(router, "POST", path, map[string]int{"status": 0})
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 DB 中 status 变为 0
	updatedUser, _ := svc.GetBizUser(user.ID)
	if updatedUser.Status != 0 {
		t.Fatalf("期望 status=0，实际: %d", updatedUser.Status)
	}

	// 再切换回 1
	w = performRequest(router, "POST", path, map[string]int{"status": 1})
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	updatedUser, _ = svc.GetBizUser(user.ID)
	if updatedUser.Status != 1 {
		t.Fatalf("期望 status=1，实际: %d", updatedUser.Status)
	}
}

// TestBizUser_ToggleStatus_Invalid 无效状态值 → 400
func TestBizUser_ToggleStatus_Invalid(t *testing.T) {
	_, router, svc := setupBizUserHandler(t)

	user, _ := svc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: 1,
		Phone:    "13800138004",
	})

	path := fmt.Sprintf("/api/v1/admin/biz-users/%d/toggle-status", user.ID)
	w := performRequest(router, "POST", path, map[string]int{"status": 99})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestBizUser_Create 创建用户
func TestBizUser_Create(t *testing.T) {
	_, router, _ := setupBizUserHandler(t)

	req := map[string]interface{}{
		"tenant_id": "1",
		"phone":     "13800138005",
		"nickname":  "新用户",
	}
	w := performRequest(router, "POST", "/api/v1/admin/biz-users", req)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	if code := int(resp["code"].(float64)); code != 0 {
		t.Fatalf("期望 code=0，实际: %d", code)
	}

	data := resp["data"].(map[string]interface{})
	if data["phone"] != "13800138005" {
		t.Fatalf("期望 phone=13800138005，实际: %v", data["phone"])
	}
}

// TestBizUser_Update_OptimisticLock 更新（乐观锁）
func TestBizUser_Update_OptimisticLock(t *testing.T) {
	_, router, svc := setupBizUserHandler(t)

	user, _ := svc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: 1,
		Phone:    "13800138006",
		Nickname: "旧昵称",
	})

	path := fmt.Sprintf("/api/v1/admin/biz-users/%d", user.ID)

	// 正确 version → 成功
	req := map[string]interface{}{
		"nickname": "新昵称",
		"version":  user.Version,
	}
	w := performRequest(router, "PUT", path, req)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 错误 version → 409
	req = map[string]interface{}{
		"nickname": "再次修改",
		"version":  user.Version, // 已经过时
	}
	w = performRequest(router, "PUT", path, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("期望 409（乐观锁冲突），实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestBizUser_Delete 软删除
func TestBizUser_Delete(t *testing.T) {
	_, router, svc := setupBizUserHandler(t)

	user, _ := svc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: 1,
		Phone:    "13800138007",
	})

	path := fmt.Sprintf("/api/v1/admin/biz-users/%d", user.ID)
	w := performRequest(router, "DELETE", path, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 删除后再获取 → 404
	w = performRequest(router, "GET", path, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("期望 404（已删除），实际: %d, body: %s", w.Code, w.Body.String())
	}
}
