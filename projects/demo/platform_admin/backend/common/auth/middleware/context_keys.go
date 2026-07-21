package middleware

import "context"

type contextKey string

const objectCodeKey contextKey = "record_share_object_code"

// WithObjectCode 在 context 中注入 object_code，供 DataScopeCallback 使用
func WithObjectCode(ctx context.Context, objectCode string) context.Context {
	return context.WithValue(ctx, objectCodeKey, objectCode)
}

// GetObjectCode 从 context 中获取 object_code
func GetObjectCode(ctx context.Context) string {
	if v, ok := ctx.Value(objectCodeKey).(string); ok {
		return v
	}
	return ""
}

// authInfoKey context 存取键（用于在 GORM Callback 中获取用户身份信息）
const authInfoKey contextKey = "auth_info_for_share"

// AuthInfo 用户身份信息，存入 context.Context 供 DataScopeCallback 共享规则使用
type AuthInfo struct {
	UserID   int64
	TenantID int64
	RoleIDs  []int64
	DeptIDs  []int64
}

// WithAuthInfo 在 context 中注入用户身份信息
func WithAuthInfo(ctx context.Context, info *AuthInfo) context.Context {
	return context.WithValue(ctx, authInfoKey, info)
}

// GetAuthInfo 从 context 中获取用户身份信息
func GetAuthInfo(ctx context.Context) *AuthInfo {
	if v, ok := ctx.Value(authInfoKey).(*AuthInfo); ok {
		return v
	}
	return nil
}
