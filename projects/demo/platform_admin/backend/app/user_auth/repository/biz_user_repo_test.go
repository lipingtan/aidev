package repository

import (
	"testing"

	"go-admin/app/user_auth/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 初始化 SQLite in-memory 测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.BizUser{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	return db
}

// TestCreate 创建用户无错误且ID > 0
func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBizUserRepository()

	user := &model.BizUser{
		ID:       1001,
		TenantID: 100,
		Phone:    "13800000001",
	}
	err := repo.Create(db, user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if user.ID <= 0 {
		t.Fatalf("期望 ID > 0，实际: %d", user.ID)
	}
}

// TestFindByTenantPhone_Exists 存在的记录返回正确
func TestFindByTenantPhone_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBizUserRepository()

	user := &model.BizUser{
		ID:       2001,
		TenantID: 200,
		Phone:    "13900000001",
		Nickname: "测试用户",
	}
	_ = repo.Create(db, user)

	found, err := repo.FindByTenantPhone(db, 200, "13900000001")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if found.ID != 2001 {
		t.Fatalf("期望 ID=2001，实际: %d", found.ID)
	}
	if found.Nickname != "测试用户" {
		t.Fatalf("期望 Nickname='测试用户'，实际: %s", found.Nickname)
	}
}

// TestFindByTenantPhone_NotExists 不存在返回 ErrRecordNotFound
func TestFindByTenantPhone_NotExists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBizUserRepository()

	_, err := repo.FindByTenantPhone(db, 999, "00000000000")
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("期望 ErrRecordNotFound，实际: %v", err)
	}
}

// TestIncrementTokenVersion 递增后 token_version + 1
func TestIncrementTokenVersion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBizUserRepository()

	user := &model.BizUser{
		ID:           3001,
		TenantID:     300,
		Phone:        "13700000001",
		TokenVersion: 1,
	}
	_ = repo.Create(db, user)

	err := repo.IncrementTokenVersion(db, 3001)
	if err != nil {
		t.Fatalf("递增 token_version 失败: %v", err)
	}

	found, _ := repo.FindByID(db, 3001)
	if found.TokenVersion != 2 {
		t.Fatalf("期望 token_version=2，实际: %d", found.TokenVersion)
	}
}

// TestCreate_DuplicatePhone 同租户重复手机号应返回唯一约束错误
func TestCreate_DuplicatePhone(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBizUserRepository()

	user1 := &model.BizUser{
		ID:       4001,
		TenantID: 400,
		Phone:    "13600000001",
	}
	_ = repo.Create(db, user1)

	user2 := &model.BizUser{
		ID:       4002,
		TenantID: 400,
		Phone:    "13600000001",
	}
	err := repo.Create(db, user2)
	if err == nil {
		t.Fatal("期望唯一约束错误，但创建成功")
	}
}
