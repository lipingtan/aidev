package strategy

import (
	"testing"
	"time"

	"go-admin/common/auth/config"

	"github.com/golang-jwt/jwt/v5"
)

// testConfig 返回测试用配置
func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-for-strategy",
			PlatformTokenTTL: 168 * time.Hour,
			AccessTokenTTL:   2 * time.Hour,
			Issuer:           "test-issuer",
		},
	}
}

// TestGenerateAccessToken_Admin 生成 admin token → 解析后 UserPool="admin"
func TestGenerateAccessToken_Admin(t *testing.T) {
	cfg := testConfig()
	opts := &AccessTokenOptions{
		UserID:       100,
		TenantID:     200,
		Roles:        []int64{1, 2, 3},
		UserPool:     UserPoolAdmin,
		TokenVersion: 0,
	}

	tokenStr, err := GenerateAccessToken(cfg, opts)
	if err != nil {
		t.Fatalf("生成 admin access_token 失败: %v", err)
	}

	claims, err := ParseAccessToken(cfg, tokenStr)
	if err != nil {
		t.Fatalf("解析 admin access_token 失败: %v", err)
	}

	if claims.UserID != 100 {
		t.Fatalf("期望 UserID=100，实际: %d", claims.UserID)
	}
	if claims.TenantID != 200 {
		t.Fatalf("期望 TenantID=200，实际: %d", claims.TenantID)
	}
	if claims.UserPool != UserPoolAdmin {
		t.Fatalf("期望 UserPool=%q，实际: %q", UserPoolAdmin, claims.UserPool)
	}
	if claims.TokenVersion != 0 {
		t.Fatalf("期望 TokenVersion=0，实际: %d", claims.TokenVersion)
	}
	if len(claims.Roles) != 3 {
		t.Fatalf("期望 3 个角色，实际: %d", len(claims.Roles))
	}
}

// TestGenerateAccessToken_User 生成 user token with version=3 → 解析后 TokenVersion=3
func TestGenerateAccessToken_User(t *testing.T) {
	cfg := testConfig()
	opts := &AccessTokenOptions{
		UserID:       500,
		TenantID:     600,
		Roles:        []int64{10, 20},
		UserPool:     UserPoolUser,
		TokenVersion: 3,
	}

	tokenStr, err := GenerateAccessToken(cfg, opts)
	if err != nil {
		t.Fatalf("生成 user access_token 失败: %v", err)
	}

	claims, err := ParseAccessToken(cfg, tokenStr)
	if err != nil {
		t.Fatalf("解析 user access_token 失败: %v", err)
	}

	if claims.UserPool != UserPoolUser {
		t.Fatalf("期望 UserPool=%q，实际: %q", UserPoolUser, claims.UserPool)
	}
	if claims.TokenVersion != 3 {
		t.Fatalf("期望 TokenVersion=3，实际: %d", claims.TokenVersion)
	}
	if claims.UserID != 500 {
		t.Fatalf("期望 UserID=500，实际: %d", claims.UserID)
	}
	if claims.TenantID != 600 {
		t.Fatalf("期望 TenantID=600，实际: %d", claims.TenantID)
	}
}

// TestGeneratePlatformToken 生成 platform token → 解析后 Tenants 列表正确
func TestGeneratePlatformToken(t *testing.T) {
	cfg := testConfig()
	tenants := []TenantInfo{
		{ID: 1, Code: "corp_a", Name: "公司A"},
		{ID: 2, Code: "corp_b", Name: "公司B"},
		{ID: 3, Code: "corp_c", Name: "公司C"},
	}

	tokenStr, err := GeneratePlatformToken(cfg, 42, tenants)
	if err != nil {
		t.Fatalf("生成 platform_token 失败: %v", err)
	}

	claims, err := ParsePlatformToken(cfg, tokenStr)
	if err != nil {
		t.Fatalf("解析 platform_token 失败: %v", err)
	}

	if claims.UserID != 42 {
		t.Fatalf("期望 UserID=42，实际: %d", claims.UserID)
	}
	if len(claims.Tenants) != 3 {
		t.Fatalf("期望 3 个租户，实际: %d", len(claims.Tenants))
	}
	if claims.Tenants[0].Code != "corp_a" {
		t.Fatalf("期望第一个租户 code=corp_a，实际: %s", claims.Tenants[0].Code)
	}
	if claims.Tenants[1].Name != "公司B" {
		t.Fatalf("期望第二个租户 name=公司B，实际: %s", claims.Tenants[1].Name)
	}
	if claims.Tenants[2].ID != 3 {
		t.Fatalf("期望第三个租户 ID=3，实际: %d", claims.Tenants[2].ID)
	}
}

// TestParseAccessToken_Legacy 手动构造不含 UserPool 字段的旧 JWT → 解析后 UserPool="" 不报错
func TestParseAccessToken_Legacy(t *testing.T) {
	cfg := testConfig()

	// 手动构造旧格式 claims（不含 user_pool 和 token_version 字段）
	now := time.Now()
	legacyClaims := jwt.MapClaims{
		"jti":       "legacy-jti-123",
		"iss":       cfg.JWT.Issuer,
		"sub":       "access",
		"iat":       now.Unix(),
		"exp":       now.Add(2 * time.Hour).Unix(),
		"user_id":   float64(999),
		"tenant_id": float64(888),
		"roles":     []interface{}{float64(1), float64(2)},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, legacyClaims)
	tokenStr, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		t.Fatalf("构造旧 JWT 失败: %v", err)
	}

	// 解析旧格式 token
	claims, err := ParseAccessToken(cfg, tokenStr)
	if err != nil {
		t.Fatalf("解析旧格式 access_token 不应报错，实际: %v", err)
	}

	// UserPool 应为零值空字符串
	if claims.UserPool != "" {
		t.Fatalf("旧 token 期望 UserPool=\"\"，实际: %q", claims.UserPool)
	}
	// TokenVersion 应为零值 0
	if claims.TokenVersion != 0 {
		t.Fatalf("旧 token 期望 TokenVersion=0，实际: %d", claims.TokenVersion)
	}
	// 其他字段应正确解析
	if claims.UserID != 999 {
		t.Fatalf("期望 UserID=999，实际: %d", claims.UserID)
	}
	if claims.TenantID != 888 {
		t.Fatalf("期望 TenantID=888，实际: %d", claims.TenantID)
	}
	if len(claims.Roles) != 2 {
		t.Fatalf("期望 2 个角色，实际: %d", len(claims.Roles))
	}
}
