package middleware

import (
	"context"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// testModel 测试用模型
type testModel struct {
	ID       int64  `gorm:"primaryKey"`
	OrgID    string `gorm:"type:varchar(64)"`
	RegionID string `gorm:"type:varchar(64)"`
	Name     string `gorm:"type:varchar(128)"`
}

func (testModel) TableName() string {
	return "test_entities"
}

// setupTestDB 创建内存 SQLite 并注册 Callback
func setupTestDB(t *testing.T, enabled bool) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	RegisterDataScopeCallback(db, enabled, nil)
	return db
}

// TestDataScopeInjectWhere 测试自动注入 WHERE 条件
func TestDataScopeInjectWhere(t *testing.T) {
	db := setupTestDB(t, true)

	dsc := &DataScopeContext{
		Dimensions: []DataScopeDimension{
			{
				DimensionName: "org",
				TargetEntity:  "test_entities",
				ColumnName:    "org_id",
				Values:        []string{"org_1", "org_2"},
			},
		},
	}

	ctx := SetDataScopeContext(context.Background(), dsc)
	stmt := db.WithContext(ctx).Model(&testModel{}).Find(&[]testModel{}).Statement
	sql := stmt.SQL.String()

	if !strings.Contains(sql, "IN") {
		t.Errorf("期望 SQL 包含 IN 子句, 实际: %s", sql)
	}
	if !strings.Contains(sql, "org_id") {
		t.Errorf("期望 SQL 包含 org_id 字段, 实际: %s", sql)
	}
}

// TestDataScopeMultiRoleUnion 测试多角色同维度取并集
func TestDataScopeMultiRoleUnion(t *testing.T) {
	db := setupTestDB(t, true)

	// 模拟多角色：角色1有 org_1, org_2；角色2有 org_2, org_3
	dsc := &DataScopeContext{
		Dimensions: []DataScopeDimension{
			{
				DimensionName: "org",
				TargetEntity:  "test_entities",
				ColumnName:    "org_id",
				Values:        []string{"org_1", "org_2"},
			},
			{
				DimensionName: "org",
				TargetEntity:  "test_entities",
				ColumnName:    "org_id",
				Values:        []string{"org_2", "org_3"},
			},
		},
	}

	ctx := SetDataScopeContext(context.Background(), dsc)
	stmt := db.WithContext(ctx).Model(&testModel{}).Find(&[]testModel{}).Statement
	sql := stmt.SQL.String()

	// 验证 IN 子句存在
	if !strings.Contains(sql, "IN") {
		t.Errorf("期望 SQL 包含 IN 子句, 实际: %s", sql)
	}

	// 验证合并后的值数量：org_1, org_2, org_3 = 3 个（去重并集）
	// GORM 将 []string 展开为独立元素存入 Vars
	vars := stmt.Vars
	valueSet := make(map[string]struct{})
	for _, v := range vars {
		if s, ok := v.(string); ok {
			valueSet[s] = struct{}{}
		}
	}
	if len(valueSet) != 3 {
		t.Errorf("期望并集后有 3 个唯一值, 实际: %d (%v)", len(valueSet), valueSet)
	}
	for _, expected := range []string{"org_1", "org_2", "org_3"} {
		if _, ok := valueSet[expected]; !ok {
			t.Errorf("期望包含 %s, 实际 vars: %v", expected, vars)
		}
	}
}

// TestDataScopeTargetEntityMismatch 测试 target_entity 不匹配时不注入
func TestDataScopeTargetEntityMismatch(t *testing.T) {
	db := setupTestDB(t, true)

	dsc := &DataScopeContext{
		Dimensions: []DataScopeDimension{
			{
				DimensionName: "org",
				TargetEntity:  "other_table", // 不匹配 test_entities
				ColumnName:    "org_id",
				Values:        []string{"org_1"},
			},
		},
	}

	ctx := SetDataScopeContext(context.Background(), dsc)
	stmt := db.WithContext(ctx).Model(&testModel{}).Find(&[]testModel{}).Statement
	sql := stmt.SQL.String()

	if strings.Contains(sql, "IN") {
		t.Errorf("target_entity 不匹配时不应注入 IN 子句, 实际: %s", sql)
	}
}

// TestDataScopeSystemOpSkip 测试 SystemOp 标记跳过注入
func TestDataScopeSystemOpSkip(t *testing.T) {
	db := setupTestDB(t, true)

	dsc := &DataScopeContext{
		Dimensions: []DataScopeDimension{
			{
				DimensionName: "org",
				TargetEntity:  "test_entities",
				ColumnName:    "org_id",
				Values:        []string{"org_1", "org_2"},
			},
		},
	}

	// 先设置 DataScope，再标记 SystemOp
	ctx := SetDataScopeContext(context.Background(), dsc)
	ctx = SystemOpContext(ctx)

	stmt := db.WithContext(ctx).Model(&testModel{}).Find(&[]testModel{}).Statement
	sql := stmt.SQL.String()

	if strings.Contains(sql, "IN") {
		t.Errorf("SystemOp 标记时不应注入 IN 子句, 实际: %s", sql)
	}
}

// TestDataScopeDisabled 测试 enabled=false 时不注册 Callback
func TestDataScopeDisabled(t *testing.T) {
	db := setupTestDB(t, false) // 不注册

	dsc := &DataScopeContext{
		Dimensions: []DataScopeDimension{
			{
				DimensionName: "org",
				TargetEntity:  "test_entities",
				ColumnName:    "org_id",
				Values:        []string{"org_1"},
			},
		},
	}

	ctx := SetDataScopeContext(context.Background(), dsc)
	stmt := db.WithContext(ctx).Model(&testModel{}).Find(&[]testModel{}).Statement
	sql := stmt.SQL.String()

	if strings.Contains(sql, "IN") {
		t.Errorf("disabled 时不应注入 IN 子句, 实际: %s", sql)
	}
}
