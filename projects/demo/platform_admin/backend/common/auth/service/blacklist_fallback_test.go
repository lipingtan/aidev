package service

import (
	"testing"
	"time"
)

func TestLocalBlacklistFallback_AddAndContains(t *testing.T) {
	bl := &localBlacklistFallback{}

	// 未添加时不在黑名单
	if bl.Contains("jti-1") {
		t.Error("未添加时 Contains 应返回 false")
	}

	// 添加后在黑名单
	_ = bl.Add("jti-1", time.Hour)
	if !bl.Contains("jti-1") {
		t.Error("Add 后 Contains 应返回 true")
	}
}

func TestLocalBlacklistFallback_Expiry(t *testing.T) {
	bl := &localBlacklistFallback{}

	// 添加一个极短 TTL 的 token
	_ = bl.Add("jti-expired", 1*time.Millisecond)

	// 等待过期
	time.Sleep(10 * time.Millisecond)

	// 过期后 Contains 应返回 false（惰性删除）
	if bl.Contains("jti-expired") {
		t.Error("过期后 Contains 应返回 false")
	}
}

func TestLocalBlacklistFallback_MultipleTokens(t *testing.T) {
	bl := &localBlacklistFallback{}

	_ = bl.Add("jti-a", time.Hour)
	_ = bl.Add("jti-b", time.Hour)

	if !bl.Contains("jti-a") {
		t.Error("jti-a 应在黑名单")
	}
	if !bl.Contains("jti-b") {
		t.Error("jti-b 应在黑名单")
	}
	if bl.Contains("jti-c") {
		t.Error("jti-c 未添加，不应在黑名单")
	}
}

func TestAuthService_UsesBlacklist(t *testing.T) {
	// 验证 NewAuthService 使用 localBlacklistFallback 作为降级实现
	svc := NewAuthService(nil, nil)
	if svc.blacklist == nil {
		t.Error("blacklist 不应为 nil")
	}

	// 验证可以正常使用
	_ = svc.blacklist.Add("test-jti", time.Hour)
	if !svc.blacklist.Contains("test-jti") {
		t.Error("添加后 Contains 应返回 true")
	}
}
