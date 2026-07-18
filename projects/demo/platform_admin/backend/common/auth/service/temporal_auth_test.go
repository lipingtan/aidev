package service

import (
	"testing"
	"time"

	"go-admin/common/auth/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTemporalTestDB 创建内存 SQLite 测试数据库
func setupTemporalTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.UserRole{}); err != nil {
		t.Fatalf("迁移表结构失败: %v", err)
	}
	return db
}

// TestGetEffectiveRoles_ExpiredEnd 已过期角色不返回
func TestGetEffectiveRoles_ExpiredEnd(t *testing.T) {
	db := setupTemporalTestDB(t)

	pastEnd := time.Now().Add(-1 * time.Hour)
	db.Create(&model.UserRole{
		ID:           1,
		UserID:       100,
		TenantID:     1,
		RoleID:       10,
		EffectiveEnd: &pastEnd,
	})

	roles, err := GetEffectiveRoles(db, 100, 1)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("期望返回 0 个角色，实际返回 %d 个", len(roles))
	}
}

// TestGetEffectiveRoles_FutureStart 未到生效时间的角色不返回
func TestGetEffectiveRoles_FutureStart(t *testing.T) {
	db := setupTemporalTestDB(t)

	futureStart := time.Now().Add(1 * time.Hour)
	db.Create(&model.UserRole{
		ID:             2,
		UserID:         100,
		TenantID:       1,
		RoleID:         20,
		EffectiveStart: &futureStart,
	})

	roles, err := GetEffectiveRoles(db, 100, 1)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("期望返回 0 个角色，实际返回 %d 个", len(roles))
	}
}

// TestGetEffectiveRoles_NullWindow NULL 时间窗口表示永久有效
func TestGetEffectiveRoles_NullWindow(t *testing.T) {
	db := setupTemporalTestDB(t)

	db.Create(&model.UserRole{
		ID:       3,
		UserID:   100,
		TenantID: 1,
		RoleID:   30,
		// EffectiveStart 和 EffectiveEnd 均为 nil
	})

	roles, err := GetEffectiveRoles(db, 100, 1)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("期望返回 1 个角色，实际返回 %d 个", len(roles))
	}
}

// TestGetEffectiveRoles_ActiveWindow 在有效时间窗口内的角色正常返回
func TestGetEffectiveRoles_ActiveWindow(t *testing.T) {
	db := setupTemporalTestDB(t)

	pastStart := time.Now().Add(-1 * time.Hour)
	futureEnd := time.Now().Add(1 * time.Hour)
	db.Create(&model.UserRole{
		ID:             4,
		UserID:         100,
		TenantID:       1,
		RoleID:         40,
		EffectiveStart: &pastStart,
		EffectiveEnd:   &futureEnd,
	})

	roles, err := GetEffectiveRoles(db, 100, 1)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("期望返回 1 个角色，实际返回 %d 个", len(roles))
	}
}

// TestCalculateCacheTTL_ReturnsMinExpiry TTL 不超过最近的 effective_end
func TestCalculateCacheTTL_ReturnsMinExpiry(t *testing.T) {
	now := time.Now()
	end1 := now.Add(10 * time.Minute)
	end2 := now.Add(5 * time.Minute)

	roles := []model.UserRole{
		{EffectiveEnd: &end1},
		{EffectiveEnd: &end2},
	}

	ttl := CalculateCacheTTL(roles, 30*time.Minute)

	// TTL 应约等于 5 分钟（允许 1 秒误差）
	if ttl > 5*time.Minute+time.Second || ttl < 5*time.Minute-time.Second {
		t.Errorf("期望 TTL 约 5 分钟，实际 %v", ttl)
	}
}

// TestCalculateCacheTTL_AllNoLimit 所有角色无时间限制时返回默认 TTL
func TestCalculateCacheTTL_AllNoLimit(t *testing.T) {
	roles := []model.UserRole{
		{EffectiveEnd: nil},
		{EffectiveEnd: nil},
	}

	defaultTTL := 30 * time.Minute
	ttl := CalculateCacheTTL(roles, defaultTTL)

	if ttl != defaultTTL {
		t.Errorf("期望默认 TTL %v，实际 %v", defaultTTL, ttl)
	}
}

// TestCalculateCacheTTL_MixedRoles 混合角色（部分有时间限制）
func TestCalculateCacheTTL_MixedRoles(t *testing.T) {
	now := time.Now()
	end := now.Add(15 * time.Minute)

	roles := []model.UserRole{
		{EffectiveEnd: nil},   // 永久
		{EffectiveEnd: &end},  // 15 分钟后到期
	}

	ttl := CalculateCacheTTL(roles, 30*time.Minute)

	// TTL 应约等于 15 分钟
	if ttl > 15*time.Minute+time.Second || ttl < 15*time.Minute-time.Second {
		t.Errorf("期望 TTL 约 15 分钟，实际 %v", ttl)
	}
}
