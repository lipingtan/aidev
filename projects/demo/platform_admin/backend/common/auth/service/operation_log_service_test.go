package service

import (
	"testing"
	"time"

	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupLogTestDB 创建 SQLite 内存数据库并迁移 OperationLog 表
func setupLogTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.OperationLog{}); err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// TestAsyncOperationLogger_WriteSuccess 测试异步写入成功
func TestAsyncOperationLogger_WriteSuccess(t *testing.T) {
	db := setupLogTestDB(t)
	repo := repository.NewOperationLogRepository()

	logger := NewAsyncOperationLogger(db, repo)
	logger.Start()

	now := time.Now()
	entry := &model.OperationLog{
		UserID:     1,
		TenantID:   100,
		Module:     "user",
		Action:     "CREATE",
		TargetType: "user",
		TargetID:   "42",
		Summary:    "创建用户[测试]",
		ClientIP:   "127.0.0.1",
		UserAgent:  "test-agent",
		CreatedAt:  &now,
	}

	logger.Log(entry)

	// 等待 worker 处理
	time.Sleep(200 * time.Millisecond)
	logger.Stop()

	// 验证数据已写入
	var count int64
	db.Model(&model.OperationLog{}).Count(&count)
	if count != 1 {
		t.Fatalf("期望 1 条日志记录，实际: %d", count)
	}

	var log model.OperationLog
	db.First(&log)
	if log.Module != "user" {
		t.Fatalf("期望 module=user，实际: %s", log.Module)
	}
	if log.Action != "CREATE" {
		t.Fatalf("期望 action=CREATE，实际: %s", log.Action)
	}
}

// TestAsyncOperationLogger_ChannelFullFallback 测试 channel 满时降级，不 panic、不阻塞
func TestAsyncOperationLogger_ChannelFullFallback(t *testing.T) {
	db := setupLogTestDB(t)
	repo := repository.NewOperationLogRepository()

	// 创建 logger 但修改 channel 为小 buffer 来测试降级
	logger := &AsyncOperationLogger{
		db:     db,
		repo:   repo,
		ch:     make(chan *model.OperationLog, 2), // 小 buffer
		stopCh: make(chan struct{}),
	}
	// 注意：不调用 Start()，模拟 worker 未消费的情况

	now := time.Now()
	// 写入 5 条，前 2 条进入 channel，后 3 条降级
	for i := 0; i < 5; i++ {
		logger.Log(&model.OperationLog{
			UserID:     int64(i),
			TenantID:   1,
			Module:     "test",
			Action:     "TEST",
			TargetType: "test",
			TargetID:   "1",
			Summary:    "测试降级",
			CreatedAt:  &now,
		})
	}

	// 此处能正常到达，说明没有阻塞和 panic
	// 启动 worker 消费 channel 中的 2 条
	logger.Start()
	time.Sleep(200 * time.Millisecond)
	logger.Stop()

	// 数据库中应该有 2 条记录（channel 容量为 2）
	var count int64
	db.Model(&model.OperationLog{}).Count(&count)
	if count != 2 {
		t.Fatalf("期望 2 条日志记录（channel 容量），实际: %d", count)
	}
}

// TestAsyncOperationLogger_QueryByFilter 测试查询接口筛选
func TestAsyncOperationLogger_QueryByFilter(t *testing.T) {
	db := setupLogTestDB(t)
	repo := repository.NewOperationLogRepository()

	logger := NewAsyncOperationLogger(db, repo)
	logger.Start()

	now := time.Now()
	// 写入不同模块的日志
	entries := []*model.OperationLog{
		{UserID: 1, TenantID: 1, Module: "user", Action: "CREATE", TargetType: "user", TargetID: "1", Summary: "创建用户A", CreatedAt: &now},
		{UserID: 1, TenantID: 1, Module: "role", Action: "UPDATE", TargetType: "role", TargetID: "2", Summary: "更新角色B", CreatedAt: &now},
		{UserID: 2, TenantID: 1, Module: "user", Action: "DELETE", TargetType: "user", TargetID: "3", Summary: "删除用户C", CreatedAt: &now},
	}
	for _, e := range entries {
		logger.Log(e)
	}

	time.Sleep(300 * time.Millisecond)
	logger.Stop()

	// 按 module=user 筛选
	logs, total, err := repo.List(db, repository.OperationLogListParams{
		Page:     1,
		PageSize: 10,
		Module:   "user",
	})
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if total != 2 {
		t.Fatalf("期望 module=user 的记录数为 2，实际: %d", total)
	}
	if len(logs) != 2 {
		t.Fatalf("期望返回 2 条记录，实际: %d", len(logs))
	}

	// 按 action=CREATE 筛选
	_, total, err = repo.List(db, repository.OperationLogListParams{
		Page:     1,
		PageSize: 10,
		Action:   "CREATE",
	})
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if total != 1 {
		t.Fatalf("期望 action=CREATE 的记录数为 1，实际: %d", total)
	}

	// 按 user_id 筛选
	uid := int64(2)
	logs, total, err = repo.List(db, repository.OperationLogListParams{
		Page:     1,
		PageSize: 10,
		UserID:   &uid,
	})
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if total != 1 {
		t.Fatalf("期望 user_id=2 的记录数为 1，实际: %d", total)
	}
	if logs[0].Summary != "删除用户C" {
		t.Fatalf("期望 summary=删除用户C，实际: %s", logs[0].Summary)
	}
}

// TestAsyncOperationLogger_LogSimple 测试 LogSimple 方法
func TestAsyncOperationLogger_LogSimple(t *testing.T) {
	db := setupLogTestDB(t)
	repo := repository.NewOperationLogRepository()

	logger := NewAsyncOperationLogger(db, repo)
	logger.Start()

	logger.LogSimple(10, "CREATE", "tenant", 99, "创建租户[测试公司]")

	time.Sleep(200 * time.Millisecond)
	logger.Stop()

	var count int64
	db.Model(&model.OperationLog{}).Count(&count)
	if count != 1 {
		t.Fatalf("期望 1 条日志记录，实际: %d", count)
	}

	var log model.OperationLog
	db.First(&log)
	if log.UserID != 10 {
		t.Fatalf("期望 user_id=10，实际: %d", log.UserID)
	}
	if log.TargetID != "99" {
		t.Fatalf("期望 target_id=99，实际: %s", log.TargetID)
	}
}
