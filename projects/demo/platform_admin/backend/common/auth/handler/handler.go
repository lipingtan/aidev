// Package handler Gin HTTP Handler
package handler

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/middleware"

	"github.com/gin-gonic/gin"
)

// MustGetAuthContext 从 gin.Context 获取 AuthContext，不存在或 TenantID 为 0 时返回错误
func MustGetAuthContext(c *gin.Context) (*middleware.AuthContext, error) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		return nil, errors.NewAuthError(errors.ErrInvalidCredentials, "未认证")
	}
	if authCtx.TenantID == 0 {
		return nil, errors.NewAuthError(errors.ErrForbiddenTenant, "缺少租户上下文")
	}
	return authCtx, nil
}
