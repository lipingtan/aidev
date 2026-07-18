//go:build integration

package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-admin/common/auth/config"
	"go-admin/common/auth/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupIntegrationDB 创建 SQLite 内存数据库并 AutoMigrate 全表
func setupIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 SQLite 内存 DB 失败: %v", err)
	}
	if err := autoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}
	return db
}

// setupIntegrationRouter 调用 auth.Init() 获得完整路由
func setupIntegrationRouter(t *testing.T, cfg *config.Config, db *gorm.DB) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if err := Init(cfg, db, engine); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	return engine
}

// defaultTestConfig 返回测试用配置
func defaultTestConfig(enabled bool) *config.Config {
	cfg := config.DefaultConfig()
	cfg.Enabled = enabled
	cfg.JWT.Secret = "test-secret-key-for-integration"
	cfg.APIDiscovery.Enabled = false
	return cfg
}

// seedBootstrapAdmin 直接向数据库写入一个超级管理员用户，用于测试启动时获取第一个 token
func seedBootstrapAdmin(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte("Bootstrap@123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt 加密失败: %v", err)
	}
	user := &model.User{
		Username: "bootstrap_admin",
		Password: string(hashedPwd),
		Nickname: "启动管理员",
		Status:   1,
		Version:  1,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建 bootstrap admin 失败: %v", err)
	}
	return user
}

// seedBootstrapTenant 直接向数据库写入一个租户，并关联 bootstrap admin
func seedBootstrapTenant(t *testing.T, db *gorm.DB, userID int64) *model.Tenant {
	t.Helper()
	tenant := &model.Tenant{
		TenantCode: "bootstrap_tenant",
		Name:       "启动租户",
		Status:     1,
		Version:    1,
	}
	if err := db.Create(tenant).Error; err != nil {
		t.Fatalf("创建 bootstrap tenant 失败: %v", err)
	}
	ut := &model.UserTenant{
		UserID:   userID,
		TenantID: tenant.ID,
	}
	if err := db.Create(ut).Error; err != nil {
		t.Fatalf("关联 bootstrap admin 到 tenant 失败: %v", err)
	}
	return tenant
}

// jsonBody 辅助构造 JSON 请求体
func jsonBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}
	return bytes.NewBuffer(data)
}

// testResponse 统一响应解析
type testResponse struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) *testResponse {
	t.Helper()
	var resp testResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v, body: %s", err, w.Body.String())
	}
	return &resp
}

// doRequest 发送 HTTP 请求的辅助函数
func doRequest(engine *gin.Engine, method, path string, body *bytes.Buffer, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, path, body)
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	engine.ServeHTTP(w, req)
	return w
}

// TestIntegration_FullFlow 端到端正向流程
// 1. POST /api/v1/tenants 创建租户
// 2. POST /api/v1/users 创建用户
// 3. POST /api/v1/users/:id/tenants 关联用户到租户
// 4. POST /auth/login 登录获取 platform_token
// 5. POST /auth/tenant/select 选择租户获取 access_token
// 6. GET /api/v1/roles (带 Bearer access_token) → 200
func TestIntegration_FullFlow(t *testing.T) {
	db := setupIntegrationDB(t)
	cfg := defaultTestConfig(true)

	// 种子数据：bootstrap admin + tenant（用于获取首个 access_token 调用管理接口）
	bootstrapUser := seedBootstrapAdmin(t, db)
	bootstrapTenant := seedBootstrapTenant(t, db, bootstrapUser.ID)

	engine := setupIntegrationRouter(t, cfg, db)

	// --- 获取 bootstrap admin 的 access_token ---
	loginResp := doRequest(engine, "POST", "/auth/login", jsonBody(t, map[string]string{
		"username": "bootstrap_admin",
		"password": "Bootstrap@123",
	}), "")
	if loginResp.Code != http.StatusOK {
		t.Fatalf("bootstrap 登录期望 200, 实际: %d, body: %s", loginResp.Code, loginResp.Body.String())
	}
	resp := parseResponse(t, loginResp)
	if resp.Code != 0 {
		t.Fatalf("bootstrap 登录业务码期望 0, 实际: %d, msg: %s", resp.Code, resp.Message)
	}
	var bootstrapLogin struct {
		AccessToken   string `json:"access_token"`
		PlatformToken string `json:"platform_token"`
		TokenType     string `json:"token_type"`
	}
	json.Unmarshal(resp.Data, &bootstrapLogin)

	adminToken := bootstrapLogin.AccessToken
	if adminToken == "" {
		// 如果是 platform token（不该出现在单租户情况），尝试 select tenant
		t.Fatal("bootstrap 登录未返回 access_token")
	}

	// --- Step 1: POST /api/v1/tenants 创建租户 ---
	w := doRequest(engine, "POST", "/api/v1/tenants", jsonBody(t, map[string]string{
		"tenant_code": "test_corp",
		"name":        "测试公司",
	}), adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("创建租户期望 200, 实际: %d, body: %s", w.Code, w.Body.String())
	}
	resp = parseResponse(t, w)
	if resp.Code != 0 {
		t.Fatalf("创建租户业务码期望 0, 实际: %d, msg: %s", resp.Code, resp.Message)
	}
	var tenantData struct {
		ID         int64  `json:"id"`
		TenantCode string `json:"tenant_code"`
		Name       string `json:"name"`
	}
	json.Unmarshal(resp.Data, &tenantData)
	if tenantData.ID == 0 {
		t.Fatal("创建租户返回的 ID 不应为 0")
	}
	if tenantData.TenantCode != "test_corp" {
		t.Fatalf("创建租户返回 tenant_code 期望 test_corp, 实际: %s", tenantData.TenantCode)
	}
	_ = bootstrapTenant // suppress unused

	// --- Step 2: POST /api/v1/users 创建用户 ---
	w = doRequest(engine, "POST", "/api/v1/users", jsonBody(t, map[string]string{
		"username": "test_user",
		"password": "Test@12345",
	}), adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("创建用户期望 200, 实际: %d, body: %s", w.Code, w.Body.String())
	}
	resp = parseResponse(t, w)
	if resp.Code != 0 {
		t.Fatalf("创建用户业务码期望 0, 实际: %d, msg: %s", resp.Code, resp.Message)
	}
	var userData struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	json.Unmarshal(resp.Data, &userData)
	if userData.ID == 0 {
		t.Fatal("创建用户返回的 ID 不应为 0")
	}
	if userData.Username != "test_user" {
		t.Fatalf("创建用户返回 username 期望 test_user, 实际: %s", userData.Username)
	}

	// --- Step 3: POST /api/v1/users/:id/tenants 关联用户到租户 ---
	associatePath := fmt.Sprintf("/api/v1/users/%d/tenants", userData.ID)
	w = doRequest(engine, "POST", associatePath, jsonBody(t, map[string]int64{
		"tenant_id": tenantData.ID,
	}), adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("关联用户到租户期望 200, 实际: %d, body: %s", w.Code, w.Body.String())
	}
	resp = parseResponse(t, w)
	if resp.Code != 0 {
		t.Fatalf("关联用户到租户业务码期望 0, 实际: %d, msg: %s", resp.Code, resp.Message)
	}

	// --- Step 4: POST /auth/login 登录获取 platform_token ---
	w = doRequest(engine, "POST", "/auth/login", jsonBody(t, map[string]string{
		"username": "test_user",
		"password": "Test@12345",
	}), "")
	if w.Code != http.StatusOK {
		t.Fatalf("用户登录期望 200, 实际: %d, body: %s", w.Code, w.Body.String())
	}
	resp = parseResponse(t, w)
	if resp.Code != 0 {
		t.Fatalf("用户登录业务码期望 0, 实际: %d, msg: %s", resp.Code, resp.Message)
	}
	var loginData struct {
		TokenType     string `json:"token_type"`
		Token         string `json:"token"`
		AccessToken   string `json:"access_token"`
		PlatformToken string `json:"platform_token"`
		ExpiresIn     int64  `json:"expires_in"`
		Tenants       []struct {
			ID   int64  `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"tenants"`
	}
	json.Unmarshal(resp.Data, &loginData)

	// 用户只有一个租户，应返回 access_token 直接可用
	if loginData.TokenType != "access" {
		t.Fatalf("单租户登录期望 token_type=access, 实际: %s", loginData.TokenType)
	}
	if loginData.AccessToken == "" {
		t.Fatal("access_token 不应为空")
	}
	if loginData.PlatformToken == "" {
		t.Fatal("platform_token 不应为空（单租户也返回用于刷新）")
	}
	if len(loginData.Tenants) != 1 {
		t.Fatalf("期望 1 个租户, 实际: %d", len(loginData.Tenants))
	}
	if loginData.Tenants[0].Code != "test_corp" {
		t.Fatalf("期望租户 code=test_corp, 实际: %s", loginData.Tenants[0].Code)
	}

	// --- Step 5: POST /auth/tenant/select 选择租户获取 access_token ---
	w = doRequest(engine, "POST", "/auth/tenant/select", jsonBody(t, map[string]int64{
		"tenant_id": loginData.Tenants[0].ID,
	}), loginData.PlatformToken)
	if w.Code != http.StatusOK {
		t.Fatalf("选择租户期望 200, 实际: %d, body: %s", w.Code, w.Body.String())
	}
	resp = parseResponse(t, w)
	if resp.Code != 0 {
		t.Fatalf("选择租户业务码期望 0, 实际: %d, msg: %s", resp.Code, resp.Message)
	}
	var selectData struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	json.Unmarshal(resp.Data, &selectData)
	if selectData.AccessToken == "" {
		t.Fatal("选择租户返回的 access_token 不应为空")
	}
	if selectData.ExpiresIn <= 0 {
		t.Fatalf("expires_in 应 > 0, 实际: %d", selectData.ExpiresIn)
	}

	// --- Step 6: GET /api/v1/roles (带 Bearer access_token) → 200 ---
	rolesPath := fmt.Sprintf("/api/v1/roles?tenant_id=%d", loginData.Tenants[0].ID)
	w = doRequest(engine, "GET", rolesPath, nil, selectData.AccessToken)
	if w.Code != http.StatusOK {
		t.Fatalf("访问受保护资源期望 200, 实际: %d, body: %s", w.Code, w.Body.String())
	}
	resp = parseResponse(t, w)
	if resp.Code != 0 {
		t.Fatalf("受保护资源业务码期望 0, 实际: %d, msg: %s", resp.Code, resp.Message)
	}
}

// TestIntegration_NoToken_Unauthorized 未认证测试：GET /api/v1/roles 不带 token → 401
func TestIntegration_NoToken_Unauthorized(t *testing.T) {
	db := setupIntegrationDB(t)
	cfg := defaultTestConfig(true)
	engine := setupIntegrationRouter(t, cfg, db)

	w := doRequest(engine, "GET", "/api/v1/roles?tenant_id=1", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401, 实际: %d, body: %s", w.Code, w.Body.String())
	}
	resp := parseResponse(t, w)
	if resp.Code != 40101 {
		t.Fatalf("期望业务码 40101, 实际: %d", resp.Code)
	}
	if resp.Message == "" {
		t.Fatal("错误消息不应为空")
	}
}

// TestIntegration_Disabled_NoRoutes auth.enabled=false 时 Init 不注册路由，路径返回 404
func TestIntegration_Disabled_NoRoutes(t *testing.T) {
	db := setupIntegrationDB(t)
	cfg := defaultTestConfig(false)
	engine := setupIntegrationRouter(t, cfg, db)

	// /auth/login 应返回 404
	w := doRequest(engine, "POST", "/auth/login", jsonBody(t, map[string]string{
		"username": "admin",
		"password": "123456",
	}), "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("/auth/login 期望 404, 实际: %d, body: %s", w.Code, w.Body.String())
	}

	// /api/v1/roles 也应返回 404
	w = doRequest(engine, "GET", "/api/v1/roles?tenant_id=1", nil, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("/api/v1/roles 期望 404, 实际: %d, body: %s", w.Code, w.Body.String())
	}
}
