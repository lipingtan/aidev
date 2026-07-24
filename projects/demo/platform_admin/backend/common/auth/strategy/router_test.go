package strategy

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// mockStrategy 测试用策略
type mockStrategy struct {
	grantType string
}

func (m *mockStrategy) GrantType() string {
	return m.grantType
}

func (m *mockStrategy) Authenticate(c *gin.Context) (*AuthResult, error) {
	return &AuthResult{UserPool: UserPoolAdmin}, nil
}

// TestStrategyRouter_Route_Password 注册 password 策略后能正确路由
func TestStrategyRouter_Route_Password(t *testing.T) {
	router := NewStrategyRouter()
	ps := &mockStrategy{grantType: "password"}
	router.Register(ps)

	s, err := router.Route("password")
	if err != nil {
		t.Fatalf("路由 password 策略失败: %v", err)
	}
	if s.GrantType() != "password" {
		t.Fatalf("期望 GrantType=password，实际: %s", s.GrantType())
	}
}

// TestStrategyRouter_Route_Unknown 查找未注册策略返回 ErrUnsupportedGrantType
func TestStrategyRouter_Route_Unknown(t *testing.T) {
	router := NewStrategyRouter()
	router.Register(&mockStrategy{grantType: "password"})

	_, err := router.Route("unknown")
	if err == nil {
		t.Fatal("期望返回错误，实际为 nil")
	}
	if err != ErrUnsupportedGrantType {
		t.Fatalf("期望 ErrUnsupportedGrantType，实际: %v", err)
	}
}

// TestStrategyRouter_Register_Override 重复注册同类型策略应覆盖
func TestStrategyRouter_Register_Override(t *testing.T) {
	router := NewStrategyRouter()
	router.Register(&mockStrategy{grantType: "password"})

	newPs := &mockStrategy{grantType: "password"}
	router.Register(newPs)

	s, err := router.Route("password")
	if err != nil {
		t.Fatalf("路由失败: %v", err)
	}
	if s != newPs {
		t.Fatal("期望返回新注册的策略实例")
	}
}
