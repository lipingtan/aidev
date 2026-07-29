package strategy

import (
	"go-admin/common/auth/errors"

	"github.com/gin-gonic/gin"
)

// OAuth2Strategy OAuth2 认证策略骨架
// 当前仅占位，实际调用返回"未对接"错误
type OAuth2Strategy struct{}

// GrantType 返回策略标识
func (s *OAuth2Strategy) GrantType() string { return "oauth2" }

// Authenticate 认证入口（骨架，返回 501）
func (s *OAuth2Strategy) Authenticate(c *gin.Context) (*AuthResult, error) {
	return nil, errors.NewAuthError(errors.ErrNotImplemented, "OAuth2 策略暂未对接，请联系管理员")
}
