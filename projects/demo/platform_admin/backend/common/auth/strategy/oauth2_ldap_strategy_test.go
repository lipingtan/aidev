package strategy

import (
	"net/http/httptest"
	"testing"

	"go-admin/common/auth/errors"

	"github.com/gin-gonic/gin"
)

// setupTestCtx 创建测试用 gin.Context
func setupTestCtx() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}

// TestOAuth2Strategy_GrantType 返回 "oauth2"
func TestOAuth2Strategy_GrantType(t *testing.T) {
	s := &OAuth2Strategy{}
	if s.GrantType() != "oauth2" {
		t.Fatalf("期望 oauth2，实际 %s", s.GrantType())
	}
}

// TestOAuth2Strategy_Authenticate_Returns501 调用返回 ErrNotImplemented
func TestOAuth2Strategy_Authenticate_Returns501(t *testing.T) {
	s := &OAuth2Strategy{}
	c := setupTestCtx()

	result, err := s.Authenticate(c)
	if result != nil {
		t.Fatal("期望 result 为 nil")
	}
	if err == nil {
		t.Fatal("期望返回错误，实际 nil")
	}

	authErr, ok := err.(*errors.AuthError)
	if !ok {
		t.Fatalf("期望 *errors.AuthError，实际 %T", err)
	}
	if authErr.Code != errors.ErrNotImplemented {
		t.Fatalf("期望错误码 %d，实际 %d", errors.ErrNotImplemented, authErr.Code)
	}
	if authErr.Message == "" {
		t.Fatal("错误消息不应为空")
	}
}

// TestLDAPStrategy_GrantType 返回 "ldap"
func TestLDAPStrategy_GrantType(t *testing.T) {
	s := &LDAPStrategy{}
	if s.GrantType() != "ldap" {
		t.Fatalf("期望 ldap，实际 %s", s.GrantType())
	}
}

// TestLDAPStrategy_Authenticate_Returns501 调用返回 ErrNotImplemented
func TestLDAPStrategy_Authenticate_Returns501(t *testing.T) {
	s := &LDAPStrategy{}
	c := setupTestCtx()

	result, err := s.Authenticate(c)
	if result != nil {
		t.Fatal("期望 result 为 nil")
	}
	if err == nil {
		t.Fatal("期望返回错误，实际 nil")
	}

	authErr, ok := err.(*errors.AuthError)
	if !ok {
		t.Fatalf("期望 *errors.AuthError，实际 %T", err)
	}
	if authErr.Code != errors.ErrNotImplemented {
		t.Fatalf("期望错误码 %d，实际 %d", errors.ErrNotImplemented, authErr.Code)
	}
}

// TestStrategyRouter_OAuth2Ldap_Registered 注册后路由到骨架策略
func TestStrategyRouter_OAuth2Ldap_Registered(t *testing.T) {
	router := NewStrategyRouter()
	router.Register(&OAuth2Strategy{})
	router.Register(&LDAPStrategy{})

	// 验证 oauth2 可路由
	s1, err := router.Route("oauth2")
	if err != nil {
		t.Fatalf("oauth2 策略未注册: %v", err)
	}
	if s1.GrantType() != "oauth2" {
		t.Fatalf("期望 oauth2，实际 %s", s1.GrantType())
	}

	// 验证 ldap 可路由
	s2, err := router.Route("ldap")
	if err != nil {
		t.Fatalf("ldap 策略未注册: %v", err)
	}
	if s2.GrantType() != "ldap" {
		t.Fatalf("期望 ldap，实际 %s", s2.GrantType())
	}
}

// TestStrategyRouter_ExistingStrategiesUnaffected 注册骨架不影响已有策略
func TestStrategyRouter_ExistingStrategiesUnaffected(t *testing.T) {
	router := NewStrategyRouter()

	// 注册 stub 已有策略
	stub := &stubStrategy{grantType: "password"}
	router.Register(stub)
	router.Register(&OAuth2Strategy{})
	router.Register(&LDAPStrategy{})

	// password 策略不受影响
	s, err := router.Route("password")
	if err != nil {
		t.Fatalf("password 策略应仍可用: %v", err)
	}
	if s.GrantType() != "password" {
		t.Fatalf("期望 password，实际 %s", s.GrantType())
	}
}

// stubStrategy 用于测试的 stub 策略（避免与 router_test.go 中的 mockStrategy 冲突）
type stubStrategy struct {
	grantType string
}

func (m *stubStrategy) GrantType() string { return m.grantType }
func (m *stubStrategy) Authenticate(c *gin.Context) (*AuthResult, error) {
	return &AuthResult{UserID: 1}, nil
}
