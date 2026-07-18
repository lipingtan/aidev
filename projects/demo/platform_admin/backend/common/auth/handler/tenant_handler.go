package handler

import (
	"strconv"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// TenantHandler 租户管理 HTTP Handler
type TenantHandler struct {
	svc *service.TenantService
}

// NewTenantHandler 创建 TenantHandler 实例
func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

// RegisterRoutes 注册租户管理路由
func (h *TenantHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// TODO: 后续添加 SUPER_ADMIN 权限中间件
	tenants := rg.Group("/tenants")
	{
		tenants.GET("", h.List)
		tenants.POST("", h.Create)
		tenants.PUT("/:id", h.Update)
		tenants.PUT("/:id/status", h.UpdateStatus)
		tenants.DELETE("/:id", h.Delete)
	}
}

// List 分页查询租户列表
// GET /api/v1/tenants?page=1&page_size=20&status=1&name=xxx
func (h *TenantHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}
	name := c.Query("name")

	result, err := h.svc.ListTenants(page, pageSize, status, name)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, result)
}

// Create 创建租户
// POST /api/v1/tenants
func (h *TenantHandler) Create(c *gin.Context) {
	var req service.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	tenant, err := h.svc.CreateTenant(&req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, tenant)
}

// Update 更新租户
// PUT /api/v1/tenants/:id
func (h *TenantHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的租户 ID"))
		return
	}

	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	tenant, err := h.svc.UpdateTenant(id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, tenant)
}

// UpdateStatus 启用/禁用租户
// PUT /api/v1/tenants/:id/status
func (h *TenantHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的租户 ID"))
		return
	}

	var req service.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// Delete 删除租户（软删除）
// DELETE /api/v1/tenants/:id
func (h *TenantHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的租户 ID"))
		return
	}

	if err := h.svc.DeleteTenant(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}
