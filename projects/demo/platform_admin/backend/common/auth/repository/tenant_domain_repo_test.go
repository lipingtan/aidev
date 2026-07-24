package repository

import (
	"go-admin/common/auth/model"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTenantDomainDB 初始化内存 SQLite 用于测试（纯 Go 实现，无需 CGO）
func setupTenantDomainDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试 DB 失败: %v", err)
	}
	if err := db.AutoMigrate(&model.TenantDomain{}); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}
	return db
}

// insertTenantDomain 测试辅助：插入一条记录
func insertTenantDomain(t *testing.T, db *gorm.DB, domain string, tenantID int64) *model.TenantDomain {
	t.Helper()
	td := &model.TenantDomain{
		Domain:    domain,
		MatchType: "EXACT",
		TenantID:  tenantID,
		Remark:    "",
		Version:   1,
	}
	if err := td.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate: %v", err)
	}
	if err := db.Create(td).Error; err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}
	return td
}

// TestFindByDomain_Exists 存在且未删除 → 返回正确记录
func TestFindByDomain_Exists(t *testing.T) {
	db := setupTenantDomainDB(t)
	repo := NewTenantDomainRepository()
	td := insertTenantDomain(t, db, "abc.example.com", 1001)

	got, err := repo.FindByDomain(db, "abc.example.com")
	if err != nil {
		t.Fatalf("FindByDomain error: %v", err)
	}
	if got == nil {
		t.Fatal("FindByDomain 返回 nil，期望非 nil")
	}
	if got.ID != td.ID || got.TenantID != td.TenantID {
		t.Errorf("FindByDomain 返回记录不匹配: got %+v, want id=%d tenantID=%d", got, td.ID, td.TenantID)
	}
}

// TestFindByDomain_NotExists 不存在 → 返回 nil, nil
func TestFindByDomain_NotExists(t *testing.T) {
	db := setupTenantDomainDB(t)
	repo := NewTenantDomainRepository()

	got, err := repo.FindByDomain(db, "nonexistent.com")
	if err != nil {
		t.Fatalf("FindByDomain error: %v", err)
	}
	if got != nil {
		t.Errorf("FindByDomain 应返回 nil，got %+v", got)
	}
}

// TestFindByDomain_SoftDeleted 软删除后 → 返回 nil
func TestFindByDomain_SoftDeleted(t *testing.T) {
	db := setupTenantDomainDB(t)
	repo := NewTenantDomainRepository()
	td := insertTenantDomain(t, db, "deleted.example.com", 1001)

	// 软删除
	if err := repo.SoftDelete(db, td.ID); err != nil {
		t.Fatalf("SoftDelete error: %v", err)
	}

	got, err := repo.FindByDomain(db, "deleted.example.com")
	if err != nil {
		t.Fatalf("FindByDomain after soft delete error: %v", err)
	}
	if got != nil {
		t.Errorf("软删除后 FindByDomain 应返回 nil，got %+v", got)
	}
}

// TestList_DomainFuzzySearch domain 模糊搜索只返回匹配记录
func TestList_DomainFuzzySearch(t *testing.T) {
	db := setupTenantDomainDB(t)
	repo := NewTenantDomainRepository()
	insertTenantDomain(t, db, "abc.example.com", 1001)
	insertTenantDomain(t, db, "xyz.example.com", 1002)
	insertTenantDomain(t, db, "other.domain.net", 1003)

	list, total, err := repo.List(db, TenantDomainListParams{
		Page: 1, PageSize: 20, Domain: "example.com",
	})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total != 2 {
		t.Errorf("模糊搜索 total = %d，期望 2", total)
	}
	if len(list) != 2 {
		t.Errorf("模糊搜索 len(list) = %d，期望 2", len(list))
	}
}

// TestSoftDelete_FindByDomainReturnsNil 软删后 FindByDomain 找不到
func TestSoftDelete_FindByDomainReturnsNil(t *testing.T) {
	db := setupTenantDomainDB(t)
	repo := NewTenantDomainRepository()
	td := insertTenantDomain(t, db, "todelete.com", 2001)

	if err := repo.SoftDelete(db, td.ID); err != nil {
		t.Fatalf("SoftDelete error: %v", err)
	}

	var deletedAt *time.Time
	db.Model(&model.TenantDomain{}).Unscoped().Select("deleted_at").Where("id = ?", td.ID).Scan(&deletedAt)
	if deletedAt == nil {
		t.Error("SoftDelete 后 deleted_at 为 nil，期望非 nil")
	}

	got, _ := repo.FindByDomain(db, "todelete.com")
	if got != nil {
		t.Errorf("SoftDelete 后 FindByDomain 应返回 nil，got %+v", got)
	}
}
