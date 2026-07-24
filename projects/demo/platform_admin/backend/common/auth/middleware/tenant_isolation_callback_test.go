package middleware

import (
	"context"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// tenantTestModel 含 tenant_id 的测试模型
type tenantTestModel struct {
	ID       int64  `gorm:"primaryKey"`
	TenantID int64  `gorm:"column:tenant_id"`
	Name     string `gorm:"type:varchar(128)"`
}

func (tenantTestModel) TableName() string {
	return "tenant_test_entities"
}

// noTenantTestModel 不含 tenant_id 的测试模型
type noTenantTestModel struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(128)"`
}

func (noTenantTestModel) TableName() string {
	return "no_tenant_test_entities"
}

// setupTenantTestDB 创建内存 SQLite 并注册 TenantIsolation + DataScope Callback
func setupTenantTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	RegisterTenantIsolationCallback(db)
	RegisterDataScopeCallback(db, true, nil)
	return db
}

// setupTenantTestDBReal 创建真实内存 SQLite（非 DryRun，用于 Create 测试）
func setupTenantTestDBReal(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&tenantTestModel{}, &noTenantTestModel{}); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}
	RegisterTenantIsolationCallback(db)
	RegisterDataScopeCallback(db, true, nil)
	return db
}

// withTenantContext 创建含 tenant_id 的 context
func withTenantContext(tenantID int64) context.Context {
	ctx := context.Background()
	return WithAuthInfo(ctx, &AuthInfo{TenantID: tenantID})
}

// TestTenantQuery_WithTenantID 含 tenant_id 表 + tenantID=1 → SQL 包含 WHERE tenant_id=1
func TestTenantQuery_WithTenantID(t *testing.T) {
	db := setupTenantTestDB(t)

	ctx := withTenantContext(1)
	stmt := db.WithContext(ctx).Model(&tenantTestModel{}).Find(&[]tenantTestModel{}).Statement
	sql := stmt.SQL.String()

	if !strings.Contains(sql, "tenant_id") {
		t.Errorf("期望 SQL 包含 tenant_id 条件, 实际: %s", sql)
	}
}

// TestTenantQuery_ZeroTenantID 含 tenant_id 表 + tenantID=0 → SQL 不包含 tenant_id 条件
func TestTenantQuery_ZeroTenantID(t *testing.T) {
	db := setupTenantTestDB(t)

	ctx := withTenantContext(0)
	stmt := db.WithContext(ctx).Model(&tenantTestModel{}).Find(&[]tenantTestModel{}).Statement
	sql := stmt.SQL.String()

	if strings.Contains(sql, "tenant_id") {
		t.Errorf("tenantID=0 时不应注入 tenant_id 条件, 实际: %s", sql)
	}
}

// TestTenantQuery_NoTenantField 不含 tenant_id 表 → SQL 无 tenant_id 条件
func TestTenantQuery_NoTenantField(t *testing.T) {
	db := setupTenantTestDB(t)

	ctx := withTenantContext(1)
	stmt := db.WithContext(ctx).Model(&noTenantTestModel{}).Find(&[]noTenantTestModel{}).Statement
	sql := stmt.SQL.String()

	if strings.Contains(sql, "tenant_id") {
		t.Errorf("不含 tenant_id 字段的表不应注入条件, 实际: %s", sql)
	}
}

// TestTenantCreate_WithTenantID Create 含 tenant_id 表 + tenantID=1 → 记录 tenant_id=1
func TestTenantCreate_WithTenantID(t *testing.T) {
	db := setupTenantTestDBReal(t)

	ctx := withTenantContext(1)
	record := &tenantTestModel{Name: "test"}
	if err := db.WithContext(ctx).Create(record).Error; err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	// 使用 SystemOp 绕过隔离查询验证
	sysCtx := SystemOpContext(context.Background())
	var found tenantTestModel
	if err := db.WithContext(sysCtx).First(&found, record.ID).Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if found.TenantID != 1 {
		t.Errorf("期望 tenant_id=1, 实际: %d", found.TenantID)
	}
}

// TestTenantCreate_ZeroTenantID Create + tenantID=0 → 记录 tenant_id 不被填充
func TestTenantCreate_ZeroTenantID(t *testing.T) {
	db := setupTenantTestDBReal(t)

	ctx := withTenantContext(0)
	record := &tenantTestModel{Name: "test_zero"}
	if err := db.WithContext(ctx).Create(record).Error; err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	sysCtx := SystemOpContext(context.Background())
	var found tenantTestModel
	if err := db.WithContext(sysCtx).First(&found, record.ID).Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if found.TenantID != 0 {
		t.Errorf("tenantID=0 时不应填充, 期望 0, 实际: %d", found.TenantID)
	}
}

// TestTenantIsolation_SystemOpSkip SystemOp 标记 → 跳过所有注入
func TestTenantIsolation_SystemOpSkip(t *testing.T) {
	db := setupTenantTestDB(t)

	ctx := withTenantContext(1)
	ctx = SystemOpContext(ctx)

	stmt := db.WithContext(ctx).Model(&tenantTestModel{}).Find(&[]tenantTestModel{}).Statement
	sql := stmt.SQL.String()

	if strings.Contains(sql, "tenant_id") {
		t.Errorf("SystemOp 标记时不应注入 tenant_id 条件, 实际: %s", sql)
	}
}

// TestTenantIsolation_DataScopeStillWorks DataScopeCallback 仍然正常工作
func TestTenantIsolation_DataScopeStillWorks(t *testing.T) {
	db := setupTenantTestDB(t)

	// 同时设置 tenant 和 data_scope
	ctx := withTenantContext(1)
	dsc := &DataScopeContext{
		Dimensions: []DataScopeDimension{
			{
				DimensionName: "org",
				TargetEntity:  "tenant_test_entities",
				ColumnName:    "org_id",
				Values:        []string{"org_1"},
			},
		},
	}
	ctx = SetDataScopeContext(ctx, dsc)

	stmt := db.WithContext(ctx).Model(&tenantTestModel{}).Find(&[]tenantTestModel{}).Statement
	sql := stmt.SQL.String()

	// 两个条件都应注入
	if !strings.Contains(sql, "tenant_id") {
		t.Errorf("期望 SQL 包含 tenant_id 条件, 实际: %s", sql)
	}
	if !strings.Contains(sql, "org_id") {
		t.Errorf("期望 SQL 包含 org_id 条件（DataScope）, 实际: %s", sql)
	}
}
