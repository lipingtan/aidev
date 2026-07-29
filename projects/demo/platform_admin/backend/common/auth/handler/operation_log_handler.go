package handler

import (
	"strconv"
	"time"

	"go-admin/common/auth/middleware"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// OperationLogHandler 操作日志 HTTP Handler
type OperationLogHandler struct {
	querySvc *service.OperationLogQueryService
}

// NewOperationLogHandler 创建 OperationLogHandler 实例
func NewOperationLogHandler(querySvc *service.OperationLogQueryService) *OperationLogHandler {
	return &OperationLogHandler{querySvc: querySvc}
}

// RegisterRoutes 注册操作日志路由
func (h *OperationLogHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/operation-logs", h.List)
}

// List 分页查询操作日志
// CR-8: 新增 risk_level 过滤 + 多租户隔离（SUPER_ADMIN 全量 / TENANT_ADMIN 本租户 / 普通用户仅本人）
func (h *OperationLogHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	params := repository.OperationLogListParams{
		Page:       page,
		PageSize:   pageSize,
		Module:     c.Query("module"),
		Action:     c.Query("action"),
		TargetType: c.Query("target_type"),
		RiskLevel:  c.Query("risk_level"),
	}

	if uid := c.Query("user_id"); uid != "" {
		if v, err := strconv.ParseInt(uid, 10, 64); err == nil {
			params.UserID = &v
		}
	}
	if st := c.Query("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			params.StartTime = &t
		}
	}
	if et := c.Query("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			params.EndTime = &t
		}
	}

	// CR-8: 多租户隔离（按角色分级查询范围）
	authCtx := middleware.GetAuthContext(c)
	if authCtx != nil {
		if !middleware.IsSuperAdmin(c) {
			// 非 SUPER_ADMIN：强制按租户隔离
			params.TenantID = &authCtx.TenantID
			// 如果不是 TENANT_ADMIN（即普通用户），进一步限制为仅查本人日志
			if !middleware.IsTenantAdmin(c) {
				params.UserID = &authCtx.UserID
			}
		}
		// SUPER_ADMIN：不加 tenant_id 过滤，支持按 query 参数选择性过滤
		if middleware.IsSuperAdmin(c) {
			if tid := c.Query("tenant_id"); tid != "" {
				if v, err := strconv.ParseInt(tid, 10, 64); err == nil {
					params.TenantID = &v
				}
			}
		}
	}

	result, err := h.querySvc.ListLogs(params)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, result)
}
