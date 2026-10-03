package handler

import (
	"bytes"
"fmt"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-admin/common/auth/config"
	"go-admin/common/auth/middleware"
	"go-admin/common/auth/model"
	"go-admin/common/auth/service"
	"go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupAuthTestDB 创建 auth 测试专用数据库
func setupAuthTestDB(t *testing.T) *gorm.DB {
	return setupTestDB(t)
}

// setupAuthRouter 创建认证测试路由
func setupAuthRouter(db *gorm.DB) (*gin.Engine, *service.AuthService) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := config.DefaultConfig()
	authSvc := service.NewAuthService(db, cfg)
	h := NewAuthHandler(authSvc)

	// 注册 /auth 路由
	h.RegisterRoutes(r.Group(""))

	// 注册一个受保护的测试端点
	protected := r.Group("/api/v1/admin")
	protected.Use(middleware.AuthMiddleware(authSvc))
	protected.GET("/me", func(c *gin.Context) {
		authCtx := middleware.GetAuthContext(c)
		Success(c, map[string]interface{}{
			"user_id":   authCtx.UserID,
			"tenant_id": authCtx.TenantID,
			"roles":     authCtx.Roles,
		})
	})

	return r, authSvc
}

// createAuthUser 创建测试用户（bcrypt 密码），用于认证测试
func createAuthUser(t *testing.T, db *gorm.DB, username, password string) *model.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("生成密码哈希失败: %v", err)
	}
	user := &model.User{
		Username: username,
		Password: string(hash),
		Status:   1,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
	return user
}

// createAuthTenant 创建测试租户，用于认证测试
func createAuthTenant(t *testing.T, db *gorm.DB, code, name string) *model.Tenant {
	tenant := &model.Tenant{
		TenantCode: code,
		Name:       name,
		Status:     1,
	}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("创建测试租户失败: %v", err)
	}
	return tenant
}

// createAuthUserTenant 创建用户-租户关联，用于认证测试
func createAuthUserTenant(t *testing.T, db *gorm.DB, userID, tenantID int64) {
	ut := &model.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
	}
	if err := db.Create(ut).Error; err != nil {
		t.Fatalf("创建用户-租户关联失败: %v", err)
	}
}

// TestLogin_Success_MultiTenant 测试多租户登录返回 platform_token
func TestLogin_Success_MultiTenant(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant1 := createAuthTenant(t, db, "corp_a", "公司A")
	tenant2 := createAuthTenant(t, db, "corp_b", "公司B")
	createAuthUserTenant(t, db, user.ID, tenant1.ID)
	createAuthUserTenant(t, db, user.ID, tenant2.ID)

	body := map[string]interface{}{
		"username": "admin",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			TokenType string             `json:"token_type"`
			Token     string             `json:"token"`
			Tenants   []strategy.TenantInfo `json:"tenants"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.TokenType != "platform" {
		t.Fatalf("期望 token_type=platform，实际: %s", resp.Data.TokenType)
	}
	if resp.Data.Token == "" {
		t.Fatal("platform_token 不应为空")
	}
	if len(resp.Data.Tenants) != 2 {
		t.Fatalf("期望 2 个租户，实际: %d", len(resp.Data.Tenants))
	}
}

// TestLogin_WrongPassword 测试错误密码返回 401
func TestLogin_WrongPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant := createAuthTenant(t, db, "corp_a", "公司A")
	createAuthUserTenant(t, db, user.ID, tenant.ID)

	body := map[string]interface{}{
		"username": "admin",
		"password": "wrong_password",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望状态码 401，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 40101 {
		t.Fatalf("期望 code=40101，实际: %d", resp.Code)
	}
}

// TestSelectTenant_Success 测试选择租户返回 access_token
func TestSelectTenant_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant1 := createAuthTenant(t, db, "corp_a", "公司A")
	tenant2 := createAuthTenant(t, db, "corp_b", "公司B")
	createAuthUserTenant(t, db, user.ID, tenant1.ID)
	createAuthUserTenant(t, db, user.ID, tenant2.ID)

	// 先登录获取 platform_token
	loginBody := map[string]interface{}{
		"username": "admin",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	platformToken := loginResp.Data.Token

	// 选择租户
	selectBody := map[string]interface{}{
		"tenant_id": fmt.Sprintf("%d", tenant1.ID),
	}
	jsonBody, _ = json.Marshal(selectBody)
	req = httptest.NewRequest(http.MethodPost, "/auth/tenant/select", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+platformToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var selectResp struct {
		Code int `json:"code"`
		Data struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int64  `json:"expires_in"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &selectResp)

	if selectResp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", selectResp.Code)
	}
	if selectResp.Data.AccessToken == "" {
		t.Fatal("access_token 不应为空")
	}
	if selectResp.Data.ExpiresIn <= 0 {
		t.Fatal("expires_in 应大于 0")
	}
}

// TestSelectTenant_NotAssociated 测试选择未关联租户返回错误
func TestSelectTenant_NotAssociated(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant1 := createAuthTenant(t, db, "corp_a", "公司A")
	tenant2 := createAuthTenant(t, db, "corp_b", "公司B")
	createAuthUserTenant(t, db, user.ID, tenant1.ID)
	// 用户未关联 tenant2

	// 登录
	loginBody := map[string]interface{}{
		"username": "admin",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var loginResp struct {
		Data struct {
			Token         string `json:"token"`
			PlatformToken string `json:"platform_token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)

	// 单租户时直接返回 access_token，取 platform_token 做后续测试
	platformToken := loginResp.Data.PlatformToken
	if platformToken == "" {
		platformToken = loginResp.Data.Token
	}

	// 尝试选择未关联的租户
	selectBody := map[string]interface{}{
		"tenant_id": fmt.Sprintf("%d", tenant2.ID),
	}
	jsonBody, _ = json.Marshal(selectBody)
	req = httptest.NewRequest(http.MethodPost, "/auth/tenant/select", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+platformToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 应返回认证错误（platform_token 中不包含 tenant2）
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望状态码 401，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestLogin_SingleTenant 测试单租户直接返回 access_token
func TestLogin_SingleTenant(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant := createAuthTenant(t, db, "corp_a", "公司A")
	createAuthUserTenant(t, db, user.ID, tenant.ID)

	body := map[string]interface{}{
		"username": "admin",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			TokenType   string `json:"token_type"`
			AccessToken string `json:"access_token"`
			Token       string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.TokenType != "access" {
		t.Fatalf("单租户期望 token_type=access，实际: %s", resp.Data.TokenType)
	}
	if resp.Data.AccessToken == "" {
		t.Fatal("单租户时 access_token 不应为空")
	}
}

// TestLogout_BlacklistToken 测试 logout 后 token 被加入黑名单
func TestLogout_BlacklistToken(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant := createAuthTenant(t, db, "corp_a", "公司A")
	createAuthUserTenant(t, db, user.ID, tenant.ID)

	// 登录（单租户，直接获取 access_token）
	loginBody := map[string]interface{}{
		"username": "admin",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var loginResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
			Token       string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	accessToken := loginResp.Data.AccessToken
	if accessToken == "" {
		accessToken = loginResp.Data.Token
	}

	// 先验证 access_token 有效
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("logout 前访问应成功，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// logout
	req = httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("logout 应返回 200，实际: %d", w.Code)
	}

	// 再次使用已注销的 token 访问受保护端点
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("logout 后期望 401，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestRefreshToken 测试刷新 access_token
func TestRefreshToken(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	user := createAuthUser(t, db, "admin", "password123")
	tenant1 := createAuthTenant(t, db, "corp_a", "公司A")
	createAuthUserTenant(t, db, user.ID, tenant1.ID)

	// 登录获取 token（单租户直接返回 access_token + refresh_token）
	loginBody := map[string]interface{}{
		"username": "admin",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(loginBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var loginResp struct {
		Code int `json:"code"`
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)

	if loginResp.Code != 0 {
		t.Fatalf("登录失败: code=%d, body=%s", loginResp.Code, w.Body.String())
	}
	if loginResp.Data.AccessToken == "" {
		t.Fatal("登录后 access_token 不应为空")
	}
	if loginResp.Data.RefreshToken == "" {
		t.Fatal("登录后 refresh_token 不应为空")
	}

	// 使用 refresh_token 刷新
	refreshBody := map[string]interface{}{
		"refresh_token": loginResp.Data.RefreshToken,
	}
	jsonBody, _ = json.Marshal(refreshBody)
	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("刷新期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var refreshResp struct {
		Code int `json:"code"`
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &refreshResp)

	if refreshResp.Code != 0 {
		t.Fatalf("刷新期望 code=0，实际: %d, body: %s", refreshResp.Code, w.Body.String())
	}
	if refreshResp.Data.AccessToken == "" {
		t.Fatal("刷新后 access_token 不应为空")
	}
	if refreshResp.Data.RefreshToken == "" {
		t.Fatal("刷新后 refresh_token 不应为空")
	}
	if refreshResp.Data.AccessToken == loginResp.Data.AccessToken {
		t.Fatal("刷新后 access_token 应该变化")
	}
	if refreshResp.Data.RefreshToken == loginResp.Data.RefreshToken {
		t.Fatal("刷新后 refresh_token 应该变化（旋转刷新）")
	}

	// 验证旧 refresh_token 已失效（黑名单）
	oldRefreshBody := map[string]interface{}{
		"refresh_token": loginResp.Data.RefreshToken,
	}
	jsonBody, _ = json.Marshal(oldRefreshBody)
	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		// 旧 token 应该被拒绝
		var oldRefreshResp struct {
			Code int `json:"code"`
		}
		json.Unmarshal(w.Body.Bytes(), &oldRefreshResp)
		if oldRefreshResp.Code == 0 {
			t.Fatal("旧 refresh_token 应该被拒绝（黑名单），但刷新成功了")
		}
	}
}

// TestMiddleware_Unauthorized 测试未认证请求返回 401
func TestMiddleware_Unauthorized(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	// 不带 token 访问受保护端点
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestMiddleware_InvalidToken 测试无效 token 返回 401
func TestMiddleware_InvalidToken(t *testing.T) {
	db := setupAuthTestDB(t)
	r, _ := setupAuthRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际: %d, body: %s", w.Code, w.Body.String())
	}
}
