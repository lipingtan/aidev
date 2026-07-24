package handler

import (
	"strconv"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/middleware"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// TenantDomainHandler 域名-租户映射 HTTP Handler
type TenantDomainHandler struct {
	svc *service.TenantDomainService
}

// NewTenantDomainHandler 创建 TenantDomainHandler 实例
func NewTenantDomainHandler(svc *service.TenantDomainService) *TenantDomainHandler {
	return &TenantDomainHandler{svc: svc}
}

// QueryByDomain 公开接口：根据域名查询租户信息
// GET /api/v1/public/tenant-domain?domain=xxx
func (h *TenantDomainHandler) QueryByDomain(c *gin.Context) {
	domain := c.Query("domain")
	result, err := h.svc.QueryByDomain(domain)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, result)
}

// List 分页查询域名映射列表
// GET /api/v1/admin/tenant-domains?page=1&page_size=20&domain=xxx&tenant_id=123
func (h *TenantDomainHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	domain := c.Query("domain")
	var tenantID int64
	if tid := c.Query("tenant_id"); tid != "" {
		tenantID, _ = strconv.ParseInt(tid, 10, 64)
	}

	params := repository.TenantDomainListParams{
		Page:     page,
		PageSize: pageSize,
		Domain:   domain,
		TenantID: tenantID,
	}

	list, total, err := h.svc.List(params)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// Create 创建域名映射
// POST /api/v1/admin/tenant-domains
func (h *TenantDomainHandler) Create(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		Error(c, errors.NewAuthError(errors.ErrInvalidCredentials, "未认证"))
		return
	}

	var req service.CreateTenantDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrInvalidParam, "请求参数无效: "+err.Error()))
		return
	}

	td, err := h.svc.Create(authCtx.UserID, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, td)
}

// Update 更新域名映射
// PUT /api/v1/admin/tenant-domains/:id
func (h *TenantDomainHandler) Update(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		Error(c, errors.NewAuthError(errors.ErrInvalidCredentials, "未认证"))
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrInvalidParam, "无效的 ID"))
		return
	}

	var req service.UpdateTenantDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrInvalidParam, "请求参数无效: "+err.Error()))
		return
	}

	td, err := h.svc.Update(authCtx.UserID, id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, td)
}

// Delete 删除域名映射
// DELETE /api/v1/admin/tenant-domains/:id
func (h *TenantDomainHandler) Delete(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		Error(c, errors.NewAuthError(errors.ErrInvalidCredentials, "未认证"))
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrInvalidParam, "无效的 ID"))
		return
	}

	if err := h.svc.Delete(authCtx.UserID, id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// RegisterRoutes 注册域名映射 CRUD 路由到 admin 路由组
func (h *TenantDomainHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/tenant-domains")
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
