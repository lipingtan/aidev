package strategy

import "github.com/gin-gonic/gin"

// AuthResult 认证结果
type AuthResult struct {
	UserID       int64
	TenantID     int64
	UserPool     string // "admin" | "user"
	Roles        []int64
	TokenVersion int
}

// AuthenticationStrategy 认证策略接口
type AuthenticationStrategy interface {
	GrantType() string
	Authenticate(c *gin.Context) (*AuthResult, error)
}
