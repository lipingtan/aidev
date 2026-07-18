package errors

import (
	"testing"
)

// TestErrorCodesUnique 验证所有错误码不重复
func TestErrorCodesUnique(t *testing.T) {
	codes := []int{
		ErrInvalidCredentials,
		ErrTokenExpired,
		ErrTokenBlacklisted,
		ErrAccountDisabled,
		ErrTenantDisabled,
		ErrNotAssociatedTenant,
		ErrPermissionDenied,
		ErrDataScopeViolation,
		ErrRoleRequired,
		ErrDuplicateEntity,
		ErrEntityNotFound,
		ErrCyclicHierarchy,
		ErrProtectedEntity,
		ErrTenantAppNotSubscribed,
		ErrRoleHasUsers,
		ErrMaxHierarchyDepth,
		ErrUserNotInTenant,
	}

	seen := make(map[int]bool, len(codes))
	for _, code := range codes {
		if seen[code] {
			t.Errorf("错误码 %d 重复", code)
		}
		seen[code] = true
	}
}

// TestAuthErrorImplementsError 验证 AuthError 实现 error 接口
func TestAuthErrorImplementsError(t *testing.T) {
	var err error = NewAuthError(ErrTokenExpired, "token已过期")
	if err == nil {
		t.Fatal("NewAuthError 不应返回 nil")
	}

	expected := "[auth-40102] token已过期"
	if err.Error() != expected {
		t.Errorf("Error() 返回值不匹配\n期望: %s\n实际: %s", expected, err.Error())
	}
}

// TestNewAuthError 验证构造函数
func TestNewAuthError(t *testing.T) {
	e := NewAuthError(ErrPermissionDenied, "无权访问")
	if e.Code != ErrPermissionDenied {
		t.Errorf("Code 应为 %d，实际 %d", ErrPermissionDenied, e.Code)
	}
	if e.Message != "无权访问" {
		t.Errorf("Message 应为 无权访问，实际 %s", e.Message)
	}
}
