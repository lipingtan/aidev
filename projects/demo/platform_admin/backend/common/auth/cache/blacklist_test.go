package cache

import (
	"testing"
	"time"
)

func TestLocalBlacklistStore_AddAndContains(t *testing.T) {
	store := NewLocalBlacklistStore()
	defer store.Stop()

	err := store.Add("token-abc", 5*time.Second)
	if err != nil {
		t.Fatalf("Add() 返回错误: %v", err)
	}

	if !store.Contains("token-abc") {
		t.Error("Contains() 应返回 true，token 已添加且未过期")
	}
}

func TestLocalBlacklistStore_ExpiredToken(t *testing.T) {
	store := NewLocalBlacklistStore()
	defer store.Stop()

	err := store.Add("token-expire", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Add() 返回错误: %v", err)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	if store.Contains("token-expire") {
		t.Error("Contains() 应返回 false，token 已过期")
	}
}

func TestLocalBlacklistStore_NotAdded(t *testing.T) {
	store := NewLocalBlacklistStore()
	defer store.Stop()

	if store.Contains("token-not-exist") {
		t.Error("Contains() 应返回 false，token 未添加")
	}
}
