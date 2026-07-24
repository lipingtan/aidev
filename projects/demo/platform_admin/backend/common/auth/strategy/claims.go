package strategy

import "github.com/golang-jwt/jwt/v5"

const (
	// UserPoolAdmin 管理端用户池
	UserPoolAdmin = "admin"
	// UserPoolUser 普通用户池
	UserPoolUser = "user"
)

// TenantInfo 租户简要信息（嵌入 JWT claims）
type TenantInfo struct {
	ID   int64  `json:"id,string"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// PlatformClaims platform_token 的 JWT claims
type PlatformClaims struct {
	jwt.RegisteredClaims
	UserID  int64        `json:"user_id"`
	Tenants []TenantInfo `json:"tenants"`
}

// AccessClaims access_token 的 JWT claims
// UserPool 和 TokenVersion 为扩展字段，旧 token 解析时为零值不报错
type AccessClaims struct {
	jwt.RegisteredClaims
	UserID       int64   `json:"user_id"`
	TenantID     int64   `json:"tenant_id"`
	Roles        []int64 `json:"roles"`
	UserPool     string  `json:"user_pool,omitempty"`
	TokenVersion int     `json:"token_version,omitempty"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	TokenType     string       `json:"token_type"`               // "platform" 或 "access"
	Token         string       `json:"token"`                    // platform_token 或 access_token
	AccessToken   string       `json:"access_token,omitempty"`   // 单租户时直接返回
	Tenants       []TenantInfo `json:"tenants,omitempty"`        // 可用租户列表
	ExpiresIn     int64        `json:"expires_in"`               // 过期秒数
	PlatformToken string       `json:"platform_token,omitempty"` // 单租户时也返回 platform_token 用于后续刷新
}

// SelectTenantResponse 选择租户响应
type SelectTenantResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}
