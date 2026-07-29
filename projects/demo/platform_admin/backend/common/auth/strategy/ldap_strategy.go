package strategy

import (
	"go-admin/common/auth/errors"

	"github.com/gin-gonic/gin"
)

// LDAPStrategy LDAP 认证策略骨架
// 当前仅占位，实际调用返回"未对接"错误
type LDAPStrategy struct{}

// GrantType 返回策略标识
func (s *LDAPStrategy) GrantType() string { return "ldap" }

// Authenticate 认证入口（骨架，返回 501）
func (s *LDAPStrategy) Authenticate(c *gin.Context) (*AuthResult, error) {
	return nil, errors.NewAuthError(errors.ErrNotImplemented, "LDAP 策略暂未对接，请联系管理员")
}
