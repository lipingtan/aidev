package errors

import "fmt"

// 认证相关错误码 (401xx)
const (
	ErrInvalidCredentials  = 40101
	ErrTokenExpired        = 40102
	ErrTokenBlacklisted    = 40103
	ErrAccountDisabled     = 40104
	ErrTenantDisabled      = 40105
	ErrNotAssociatedTenant = 40106
)

// 授权相关错误码 (403xx)
const (
	ErrPermissionDenied   = 40301
	ErrDataScopeViolation = 40302
	ErrRoleRequired       = 40303
	ErrForbiddenTenant    = 40304
)

// 业务逻辑错误码 (400xx)
const (
	ErrDuplicateEntity          = 40001
	ErrEntityNotFound           = 40002
	ErrCyclicHierarchy          = 40003
	ErrProtectedEntity          = 40004
	ErrTenantAppNotSubscribed   = 40005
	ErrRoleHasUsers             = 40006
	ErrMaxHierarchyDepth        = 40007
	ErrUserNotInTenant          = 40008
	ErrExceedsParentPermission  = 40009
	ErrInvalidParam             = 40010
	ErrQuotaExceeded            = 40011
)

// AuthError 统一权限错误类型
type AuthError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error 实现 error 接口
func (e *AuthError) Error() string {
	return fmt.Sprintf("[auth-%d] %s", e.Code, e.Message)
}

// NewAuthError 构造 AuthError
func NewAuthError(code int, message string) *AuthError {
	return &AuthError{Code: code, Message: message}
}
