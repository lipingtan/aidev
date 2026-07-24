package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"go-admin/common/auth/config"
	"go-admin/common/auth/service"
	"go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// mockBizUserChecker 用于测试的 BizUserChecker mock 实现
type mockBizUserChecker struct {
	status       int
	tokenVersion int
	err          error
}

func (m *mockBizUserChecker) CheckBizUser(userID int64) (int, int, error) {
	return m.status, m.tokenVersion, m.err
}

// setupTestRouter 创建测试路由（含 AuthMiddleware）
// 使用内存 SQLite 避免 nil DB panic，tenant 和 user 表查询会返回 err（跳过状态检查逻辑）
func setupTestRouter(checker BizUserChecker) (*gin.Engine, *service.AuthService) {
	gin.SetMode(gin.TestMode)

	cfg := config.DefaultConfig()
	cfg.JWT.Secret = "test-secret-key-for-unit-test"

	// 使用 glebarez/sqlite 内存数据库，避免 nil DB panic
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	authSvc := service.NewAuthService(db, cfg)

	r := gin.New()
	if checker != nil {
		r.Use(AuthMiddleware(authSvc, checker))
	} else {
		r.Use(AuthMiddleware(authSvc))
	}
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	return r, authSvc
}

// generateTestToken 生成测试 token
func generateTestToken(cfg *config.Config, userPool string, tokenVersion int) string {
	token, _ := strategy.GenerateAccessToken(cfg, &strategy.AccessTokenOptions{
		UserID:       1001,
		TenantID:     1,
		Roles:        []int64{1},
		UserPool:     userPool,
		TokenVersion: tokenVersion,
	})
	return token
}

// parseResponseBody 解析响应体
func parseResponseBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	return body
}

// TestBizUser_ValidToken_Pass C端用户 token_version 匹配 + status=1 → 放行
func TestBizUser_ValidToken_Pass(t *testing.T) {
	// 清理缓存避免测试间干扰
	bizUserCacheMap = syncMapNew()

	checker := &mockBizUserChecker{status: 1, tokenVersion: 1}
	r, authSvc := setupTestRouter(checker)
	cfg := authSvc.GetConfig()

	token := generateTestToken(cfg, strategy.UserPoolUser, 1)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际 %d, body: %s", w.Code, w.Body.String())
	}
}

// TestBizUser_TokenVersionMismatch C端用户 token_version 不匹配 → 401
func TestBizUser_TokenVersionMismatch(t *testing.T) {
	bizUserCacheMap = syncMapNew()

	// DB 中 tokenVersion=2，token 中 tokenVersion=1 → 不匹配
	checker := &mockBizUserChecker{status: 1, tokenVersion: 2}
	r, authSvc := setupTestRouter(checker)
	cfg := authSvc.GetConfig()

	token := generateTestToken(cfg, strategy.UserPoolUser, 1)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401，实际 %d", w.Code)
	}
	body := parseResponseBody(t, w)
	if code, _ := body["code"].(float64); int(code) != 40103 {
		t.Errorf("期望 code=40103，实际 %v", body["code"])
	}
	if msg, _ := body["message"].(string); msg != "token 已失效" {
		t.Errorf("期望 message='token 已失效'，实际 '%s'", msg)
	}
}

// TestBizUser_StatusDisabled C端用户 status=0 → 401（账号已禁用）
func TestBizUser_StatusDisabled(t *testing.T) {
	bizUserCacheMap = syncMapNew()

	checker := &mockBizUserChecker{status: 0, tokenVersion: 1}
	r, authSvc := setupTestRouter(checker)
	cfg := authSvc.GetConfig()

	token := generateTestToken(cfg, strategy.UserPoolUser, 1)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401，实际 %d", w.Code)
	}
	body := parseResponseBody(t, w)
	if code, _ := body["code"].(float64); int(code) != 40104 {
		t.Errorf("期望 code=40104，实际 %v", body["code"])
	}
	if msg, _ := body["message"].(string); msg != "账号已禁用" {
		t.Errorf("期望 message='账号已禁用'，实际 '%s'", msg)
	}
}

// TestAdminUser_SkipBizCheck admin token → 不查 token_version（直接放行到下一个中间件）
func TestAdminUser_SkipBizCheck(t *testing.T) {
	bizUserCacheMap = syncMapNew()

	// 即使 checker 返回异常值，admin 用户也不应执行校验
	checker := &mockBizUserChecker{status: 0, tokenVersion: 999}
	r, authSvc := setupTestRouter(checker)
	cfg := authSvc.GetConfig()

	token := generateTestToken(cfg, strategy.UserPoolAdmin, 0)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望 200（admin 不校验 biz user），实际 %d, body: %s", w.Code, w.Body.String())
	}
}

// TestBizUser_CheckerError checker 返回 error → 401（fail-closed）
func TestBizUser_CheckerError(t *testing.T) {
	bizUserCacheMap = syncMapNew()

	checker := &mockBizUserChecker{err: errors.New("db connection failed")}
	r, authSvc := setupTestRouter(checker)
	cfg := authSvc.GetConfig()

	token := generateTestToken(cfg, strategy.UserPoolUser, 1)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401（fail-closed），实际 %d", w.Code)
	}
	body := parseResponseBody(t, w)
	if msg, _ := body["message"].(string); msg != "用户信息查询失败" {
		t.Errorf("期望 message='用户信息查询失败'，实际 '%s'", msg)
	}
}

// TestBizUser_NilChecker_Skip 未注入 checker 时跳过C端校验
func TestBizUser_NilChecker_Skip(t *testing.T) {
	bizUserCacheMap = syncMapNew()

	// 不传 checker
	r, authSvc := setupTestRouter(nil)
	cfg := authSvc.GetConfig()

	// 即使是 user pool 的 token，没有 checker 也应放行
	token := generateTestToken(cfg, strategy.UserPoolUser, 1)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望 200（无 checker 跳过校验），实际 %d, body: %s", w.Code, w.Body.String())
	}
}

// TestInvalidateBizUserCache 缓存失效后重新查库
func TestInvalidateBizUserCache(t *testing.T) {
	bizUserCacheMap = syncMapNew()

	checker := &mockBizUserChecker{status: 1, tokenVersion: 1}
	r, authSvc := setupTestRouter(checker)
	cfg := authSvc.GetConfig()
	token := generateTestToken(cfg, strategy.UserPoolUser, 1)

	// 第一次请求，写入缓存
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("第一次请求期望 200，实际 %d", w.Code)
	}

	// 修改 checker 返回值（模拟强制登出后 DB 已更新）
	checker.tokenVersion = 2

	// 未失效缓存 → 仍用旧缓存 → 放行
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("缓存未失效时期望 200，实际 %d", w2.Code)
	}

	// 主动失效缓存
	InvalidateBizUserCache(1001)

	// 缓存失效后 → 重新查库 → tokenVersion 不匹配 → 401
	req3 := httptest.NewRequest("GET", "/api/test", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("缓存失效后期望 401，实际 %d", w3.Code)
	}
}

// syncMapNew 返回新的空 sync.Map（用于测试间隔离缓存）
func syncMapNew() sync.Map {
	return sync.Map{}
}
