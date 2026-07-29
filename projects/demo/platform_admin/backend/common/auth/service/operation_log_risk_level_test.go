package service

import (
	"testing"
	"time"

	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupRiskLevelTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.OperationLog{}); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	return db
}

// TestAsyncOperationLogger_HighRiskAction_FilledHIGH 高风险 action 自动填充 HIGH
func TestAsyncOperationLogger_HighRiskAction_FilledHIGH(t *testing.T) {
	db := setupRiskLevelTestDB(t)
	repo := repository.NewOperationLogRepository()
	logger := NewAsyncOperationLogger(db, repo)
	logger.Start()
	defer logger.Stop()

	entry := &model.OperationLog{
		UserID:     1,
		TenantID:   1,
		Module:     "tenant",
		Action:     "delete_tenant",
		TargetType: "tenant",
		TargetID:   "100",
		Summary:    "删除租户",
	}
	logger.Log(entry)

	// 等待异步写入
	time.Sleep(50 * time.Millisecond)

	var saved model.OperationLog
	if err := db.Where("action = ?", "delete_tenant").First(&saved).Error; err != nil {
		t.Fatalf("日志未写入 DB: %v", err)
	}
	if saved.RiskLevel != "HIGH" {
		t.Fatalf("期望 RiskLevel=HIGH，实际 %s", saved.RiskLevel)
	}
}

// TestAsyncOperationLogger_NormalAction_FilledLOW 普通 action 自动填充 LOW
func TestAsyncOperationLogger_NormalAction_FilledLOW(t *testing.T) {
	db := setupRiskLevelTestDB(t)
	repo := repository.NewOperationLogRepository()
	logger := NewAsyncOperationLogger(db, repo)
	logger.Start()
	defer logger.Stop()

	entry := &model.OperationLog{
		UserID:     1,
		TenantID:   1,
		Module:     "user",
		Action:     "create_user",
		TargetType: "user",
		TargetID:   "200",
		Summary:    "创建用户",
	}
	logger.Log(entry)

	time.Sleep(50 * time.Millisecond)

	var saved model.OperationLog
	if err := db.Where("action = ?", "create_user").First(&saved).Error; err != nil {
		t.Fatalf("日志未写入 DB: %v", err)
	}
	if saved.RiskLevel != "LOW" {
		t.Fatalf("期望 RiskLevel=LOW，实际 %s", saved.RiskLevel)
	}
}

// TestAsyncOperationLogger_ExplicitRiskLevel_NotOverridden 显式设置 risk_level 不被覆盖
func TestAsyncOperationLogger_ExplicitRiskLevel_NotOverridden(t *testing.T) {
	db := setupRiskLevelTestDB(t)
	repo := repository.NewOperationLogRepository()
	logger := NewAsyncOperationLogger(db, repo)
	logger.Start()
	defer logger.Stop()

	entry := &model.OperationLog{
		UserID:     1,
		TenantID:   1,
		Module:     "user",
		Action:     "create_user",
		TargetType: "user",
		TargetID:   "300",
		Summary:    "创建用户（手动标HIGH）",
		RiskLevel:  "HIGH", // 显式设置
	}
	logger.Log(entry)

	time.Sleep(50 * time.Millisecond)

	var saved model.OperationLog
	if err := db.Where("target_id = ?", "300").First(&saved).Error; err != nil {
		t.Fatalf("日志未写入 DB: %v", err)
	}
	// create_user 本来是 LOW，但显式设置为 HIGH 应保留
	if saved.RiskLevel != "HIGH" {
		t.Fatalf("显式设置的 RiskLevel 被覆盖：期望 HIGH，实际 %s", saved.RiskLevel)
	}
}

// TestAsyncOperationLogger_DeleteRole_FilledHIGH delete_role 是高风险
func TestAsyncOperationLogger_DeleteRole_FilledHIGH(t *testing.T) {
	db := setupRiskLevelTestDB(t)
	repo := repository.NewOperationLogRepository()
	logger := NewAsyncOperationLogger(db, repo)
	logger.Start()
	defer logger.Stop()

	entry := &model.OperationLog{
		UserID:     1,
		TenantID:   1,
		Module:     "role",
		Action:     "delete_role",
		TargetType: "role",
		TargetID:   "50",
	}
	logger.Log(entry)

	time.Sleep(50 * time.Millisecond)

	var saved model.OperationLog
	if err := db.Where("target_id = ? AND action = ?", "50", "delete_role").First(&saved).Error; err != nil {
		t.Fatalf("日志未写入 DB: %v", err)
	}
	if saved.RiskLevel != "HIGH" {
		t.Fatalf("delete_role 期望 HIGH，实际 %s", saved.RiskLevel)
	}
}
