package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// setupOperationLogRouter 创建操作日志测试路由和 logger
func setupOperationLogRouter(t *testing.T) (*gin.Engine, *service.AsyncOperationLogger) {
	db := setupTestDB(t)

	// 迁移操作日志表
	if err := db.AutoMigrate(&model.OperationLog{}); err != nil {
		t.Fatalf("迁移 OperationLog 表失败: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	repo := repository.NewOperationLogRepository()
	logger := service.NewAsyncOperationLogger(db, repo)
	logger.Start()

	querySvc := service.NewOperationLogQueryService(db, repo)
	h := NewOperationLogHandler(querySvc)

	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	return r, logger
}

// TestAsyncOperationLogger_WriteAndQuery 测试异步写入后可查询到记录
func TestAsyncOperationLogger_WriteAndQuery(t *testing.T) {
	r, logger := setupOperationLogRouter(t)
	defer logger.Stop()

	// 写入日志
	logger.Log(&model.OperationLog{
		UserID:     1001,
		TenantID:   1,
		Module:     "tenant",
		Action:     "create",
		TargetType: "tenant",
		TargetID:   "100",
		Summary:    "创建租户测试",
	})

	// 等待异步写入完成
	time.Sleep(100 * time.Millisecond)

	// 查询验证
	req := httptest.NewRequest(http.MethodGet, "/api/v1/operation-logs?module=tenant&action=create", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List     []model.OperationLog `json:"list"`
			Total    int64                `json:"total"`
			Page     int                  `json:"page"`
			PageSize int                  `json:"page_size"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.Total != 1 {
		t.Fatalf("期望 total=1，实际: %d", resp.Data.Total)
	}
	if resp.Data.List[0].Module != "tenant" {
		t.Fatalf("期望 module=tenant，实际: %s", resp.Data.List[0].Module)
	}
	if resp.Data.List[0].Action != "create" {
		t.Fatalf("期望 action=create，实际: %s", resp.Data.List[0].Action)
	}
	if resp.Data.List[0].UserID != 1001 {
		t.Fatalf("期望 user_id=1001，实际: %d", resp.Data.List[0].UserID)
	}
}

// TestOperationLogQuery_FilterByModuleAndAction 测试按 module/action 筛选
func TestOperationLogQuery_FilterByModuleAndAction(t *testing.T) {
	r, logger := setupOperationLogRouter(t)
	defer logger.Stop()

	// 写入多条不同模块的日志
	logger.Log(&model.OperationLog{
		UserID:     1,
		TenantID:   1,
		Module:     "user",
		Action:     "create",
		TargetType: "user",
		TargetID:   "10",
		Summary:    "创建用户",
	})
	logger.Log(&model.OperationLog{
		UserID:     1,
		TenantID:   1,
		Module:     "role",
		Action:     "update",
		TargetType: "role",
		TargetID:   "20",
		Summary:    "更新角色",
	})
	logger.Log(&model.OperationLog{
		UserID:     2,
		TenantID:   1,
		Module:     "user",
		Action:     "delete",
		TargetType: "user",
		TargetID:   "11",
		Summary:    "删除用户",
	})

	// 等待异步写入完成
	time.Sleep(100 * time.Millisecond)

	// 按 module=user 筛选
	req := httptest.NewRequest(http.MethodGet, "/api/v1/operation-logs?module=user", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("查询失败: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List  []model.OperationLog `json:"list"`
			Total int64                `json:"total"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.Total != 2 {
		t.Fatalf("按 module=user 筛选期望 total=2，实际: %d", resp.Data.Total)
	}

	// 按 module=user&action=create 筛选
	req = httptest.NewRequest(http.MethodGet, "/api/v1/operation-logs?module=user&action=create", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.Total != 1 {
		t.Fatalf("按 module=user&action=create 筛选期望 total=1，实际: %d", resp.Data.Total)
	}
}

// TestOperationLogQuery_Pagination 测试分页
func TestOperationLogQuery_Pagination(t *testing.T) {
	r, logger := setupOperationLogRouter(t)
	defer logger.Stop()

	// 写入 5 条日志
	for i := 0; i < 5; i++ {
		logger.Log(&model.OperationLog{
			UserID:     1,
			TenantID:   1,
			Module:     "test",
			Action:     "page_test",
			TargetType: "item",
			TargetID:   "1",
			Summary:    "分页测试",
		})
	}

	time.Sleep(100 * time.Millisecond)

	// page=1, page_size=2
	req := httptest.NewRequest(http.MethodGet, "/api/v1/operation-logs?module=test&page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List     []model.OperationLog `json:"list"`
			Total    int64                `json:"total"`
			Page     int                  `json:"page"`
			PageSize int                  `json:"page_size"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.Total != 5 {
		t.Fatalf("期望 total=5，实际: %d", resp.Data.Total)
	}
	if len(resp.Data.List) != 2 {
		t.Fatalf("期望列表长度 2，实际: %d", len(resp.Data.List))
	}
	if resp.Data.Page != 1 {
		t.Fatalf("期望 page=1，实际: %d", resp.Data.Page)
	}
	if resp.Data.PageSize != 2 {
		t.Fatalf("期望 page_size=2，实际: %d", resp.Data.PageSize)
	}
}
