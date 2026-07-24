package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-admin/app/user_auth/dto"
	"go-admin/app/user_auth/model"
	"go-admin/app/user_auth/repository"
	"go-admin/app/user_auth/service"
	"go-admin/app/user_auth/spi"
	"go-admin/common/auth/config"
	authStrategy "go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建内存 SQLite 测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	// 创建 biz_user 表
	db.AutoMigrate(&model.BizUser{})

	// 创建 admin_tenant 表
	db.Exec(`CREATE TABLE IF NOT EXISTS admin_tenant (
		id INTEGER PRIMARY KEY,
		tenant_code VARCHAR(64) UNIQUE NOT NULL,
		name VARCHAR(128) NOT NULL,
		status INTEGER DEFAULT 1
	)`)

	// 插入有效租户
	db.Exec(`INSERT INTO admin_tenant (id, tenant_code, name, status) VALUES (1, 'valid_tenant', '测试租户', 1)`)
	// 插入禁用租户
	db.Exec(`INSERT INTO admin_tenant (id, tenant_code, name, status) VALUES (2, 'disabled_tenant', '禁用租户', 0)`)

	return db
}

// setupHandler 创建测试用 Handler 和相关依赖
func setupHandler(t *testing.T, db *gorm.DB) (*UserAuthHandler, *service.SmsService) {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.JWT.Secret = "test-secret-key-for-testing-only"
	cfg.JWT.AccessTokenTTL = 2 * time.Hour

	store := service.NewMemoryCodeStore()
	sender := &spi.ConsoleMockSender{}
	smsService := service.NewSmsService(sender, store)

	repo := repository.NewBizUserRepository()
	bizUserSvc := service.NewBizUserService(db, repo)

	handler := NewUserAuthHandler(smsService, bizUserSvc, service.NewUserMenuService(db), cfg, db)
	return handler, smsService
}

// performRequest 执行 HTTP 请求
func performRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// parseResponse 解析 JSON 响应
func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v, body: %s", err, w.Body.String())
	}
	return resp
}

// TestSendCode_Success 发送验证码成功
func TestSendCode_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler, _ := setupHandler(t, db)

	router := gin.New()
	router.POST("/api/v1/user/auth/send-code", handler.SendCode)

	req := dto.SendCodeRequest{
		Phone:      "13800138000",
		TenantCode: "valid_tenant",
	}
	w := performRequest(router, "POST", "/api/v1/user/auth/send-code", req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("响应 data 字段解析失败")
	}
	expiresIn := int(data["expires_in"].(float64))
	if expiresIn != 300 {
		t.Fatalf("期望 expires_in=300，实际: %d", expiresIn)
	}
}

// TestSendCode_TooFrequent 60秒内重复发送 → 429
func TestSendCode_TooFrequent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler, _ := setupHandler(t, db)

	router := gin.New()
	router.POST("/api/v1/user/auth/send-code", handler.SendCode)

	req := dto.SendCodeRequest{
		Phone:      "13800138001",
		TenantCode: "valid_tenant",
	}

	// 第一次发送成功
	w := performRequest(router, "POST", "/api/v1/user/auth/send-code", req)
	if w.Code != http.StatusOK {
		t.Fatalf("第一次发送期望 200，实际: %d", w.Code)
	}

	// 第二次发送 → 429
	w = performRequest(router, "POST", "/api/v1/user/auth/send-code", req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("第二次发送期望 429，实际: %d, body: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	retryAfter, ok := resp["retry_after"].(float64)
	if !ok || retryAfter <= 0 {
		t.Fatalf("期望 retry_after > 0，实际: %v", resp["retry_after"])
	}
}

// TestSendCode_InvalidTenant 无效 tenant_code → 400
func TestSendCode_InvalidTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler, _ := setupHandler(t, db)

	router := gin.New()
	router.POST("/api/v1/user/auth/send-code", handler.SendCode)

	req := dto.SendCodeRequest{
		Phone:      "13800138000",
		TenantCode: "non_existent",
	}
	w := performRequest(router, "POST", "/api/v1/user/auth/send-code", req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestLogin_ExistingUser 正确验证码 + 已存在用户 → 200 + token
func TestLogin_ExistingUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler, smsService := setupHandler(t, db)

	// 预创建用户
	db.Create(&model.BizUser{
		ID:           10001,
		TenantID:     1,
		Phone:        "13800138000",
		Nickname:     "测试用户",
		Status:       1,
		TokenVersion: 1,
		Version:      1,
	})

	router := gin.New()
	router.POST("/api/v1/user/auth/login", handler.Login)

	// 通过 service 直接发送验证码获取 code
	code, err := smsService.SendCode("13800138000", "valid_tenant")
	if err != nil {
		t.Fatalf("发送验证码失败: %v", err)
	}

	// 登录
	loginReq := dto.SmsLoginRequest{
		Phone:      "13800138000",
		Code:       code,
		TenantCode: "valid_tenant",
	}
	w := performRequest(router, "POST", "/api/v1/user/auth/login", loginReq)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("响应 data 字段解析失败")
	}

	accessToken, ok := data["access_token"].(string)
	if !ok || accessToken == "" {
		t.Fatal("access_token 为空")
	}

	// 验证 token 中 UserPool = "user"
	claims := parseTestToken(t, accessToken, "test-secret-key-for-testing-only")
	if claims.UserPool != "user" {
		t.Fatalf("期望 UserPool=user，实际: %s", claims.UserPool)
	}
	if claims.UserID != 10001 {
		t.Fatalf("期望 UserID=10001，实际: %d", claims.UserID)
	}
}

// TestLogin_NewUser 正确验证码 + 未注册用户 → 自动注册 + 200 + token
func TestLogin_NewUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler, smsService := setupHandler(t, db)

	router := gin.New()
	router.POST("/api/v1/user/auth/login", handler.Login)

	// 通过 service 直接发送验证码
	code, err := smsService.SendCode("13900139000", "valid_tenant")
	if err != nil {
		t.Fatalf("发送验证码失败: %v", err)
	}

	// 登录（新用户）
	loginReq := dto.SmsLoginRequest{
		Phone:      "13900139000",
		Code:       code,
		TenantCode: "valid_tenant",
	}
	w := performRequest(router, "POST", "/api/v1/user/auth/login", loginReq)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证 DB 中新增了 biz_user
	var count int64
	db.Model(&model.BizUser{}).Where("phone = ? AND tenant_id = ?", "13900139000", 1).Count(&count)
	if count != 1 {
		t.Fatalf("期望新增 1 条 biz_user，实际: %d", count)
	}

	// 验证返回了有效 token
	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	accessToken := data["access_token"].(string)
	claims := parseTestToken(t, accessToken, "test-secret-key-for-testing-only")
	if claims.UserPool != "user" {
		t.Fatalf("期望 UserPool=user，实际: %s", claims.UserPool)
	}
}

// TestLogin_WrongCode 错误验证码 → 401
func TestLogin_WrongCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler, smsService := setupHandler(t, db)

	router := gin.New()
	router.POST("/api/v1/user/auth/login", handler.Login)

	// 先发送验证码
	_, err := smsService.SendCode("13800138002", "valid_tenant")
	if err != nil {
		t.Fatalf("发送验证码失败: %v", err)
	}

	// 用错误验证码登录
	loginReq := dto.SmsLoginRequest{
		Phone:      "13800138002",
		Code:       "0000",
		TenantCode: "valid_tenant",
	}
	w := performRequest(router, "POST", "/api/v1/user/auth/login", loginReq)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestLogin_InvalidTenant 无效 tenant_code → 400
func TestLogin_InvalidTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler, _ := setupHandler(t, db)

	router := gin.New()
	router.POST("/api/v1/user/auth/login", handler.Login)

	loginReq := dto.SmsLoginRequest{
		Phone:      "13800138000",
		Code:       "1234",
		TenantCode: "invalid_tenant",
	}
	w := performRequest(router, "POST", "/api/v1/user/auth/login", loginReq)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestLogin_DisabledTenant 禁用租户 → 400
func TestLogin_DisabledTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	handler, _ := setupHandler(t, db)

	router := gin.New()
	router.POST("/api/v1/user/auth/login", handler.Login)

	loginReq := dto.SmsLoginRequest{
		Phone:      "13800138000",
		Code:       "1234",
		TenantCode: "disabled_tenant",
	}
	w := performRequest(router, "POST", "/api/v1/user/auth/login", loginReq)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// ─── 辅助函数 ─────────────────────────────────────────────────────────────────

// parseTestToken 解析测试 token 中的 claims
func parseTestToken(t *testing.T, tokenStr string, secret string) *authStrategy.AccessClaims {
	t.Helper()
	token, err := jwt.ParseWithClaims(tokenStr, &authStrategy.AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("解析 token 失败: %v", err)
	}
	claims, ok := token.Claims.(*authStrategy.AccessClaims)
	if !ok || !token.Valid {
		t.Fatal("token claims 无效")
	}
	return claims
}
