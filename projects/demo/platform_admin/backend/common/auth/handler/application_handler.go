package handler

import (
	"strconv"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ApplicationHandler 应用管理 HTTP Handler
type ApplicationHandler struct {
	svc      *service.ApplicationService
	roleRepo repository.RoleRepository
	db       *gorm.DB
}

// NewApplicationHandler 创建 ApplicationHandler 实例
func NewApplicationHandler(svc *service.ApplicationService, roleRepo repository.RoleRepository, db *gorm.DB) *ApplicationHandler {
	return &ApplicationHandler{svc: svc, roleRepo: roleRepo, db: db}
}

// RegisterRoutes 注册应用管理路由
func (h *ApplicationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// 全局应用 CRUD（SUPER_ADMIN）
	apps := rg.Group("/applications")
	{
		apps.GET("", h.List)
		apps.POST("", h.Create)
		apps.PUT("/:id", h.Update)
		apps.DELETE("/:id", h.Delete)
	}

	// 租户订阅应用
	tenants := rg.Group("/tenants")
	{
		tenants.PUT("/:id/apps", h.SetTenantApps)
		tenants.GET("/:id/apps", h.ListTenantApps)
	}

	// 角色绑定应用
	roles := rg.Group("/roles")
	{
		roles.PUT("/:id/apps", h.SetRoleApps)
	}
}

// List 查询全局应用列表
// GET /api/v1/applications
func (h *ApplicationHandler) List(c *gin.Context) {
	apps, err := h.svc.ListApplications()
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, apps)
}

// Create 创建应用
// POST /api/v1/applications
func (h *ApplicationHandler) Create(c *gin.Context) {
	var req service.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	app, err := h.svc.CreateApplication(&req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, app)
}

// Update 更新应用
// PUT /api/v1/applications/:id
func (h *ApplicationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的应用 ID"))
		return
	}

	var req service.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	app, err := h.svc.UpdateApplication(id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, app)
}

// Delete 删除应用
// DELETE /api/v1/applications/:id
func (h *ApplicationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的应用 ID"))
		return
	}

	if err := h.svc.DeleteApplication(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// SetTenantApps 设置租户订阅应用
// PUT /api/v1/tenants/:id/apps
func (h *ApplicationHandler) SetTenantApps(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的租户 ID"))
		return
	}

	var req service.SetTenantAppsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	if err := h.svc.SetTenantApps(tenantID, &req); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// ListTenantApps 查询租户已订阅应用列表
// GET /api/v1/tenants/:id/apps
func (h *ApplicationHandler) ListTenantApps(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的租户 ID"))
		return
	}

	apps, err := h.svc.ListTenantApps(tenantID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, apps)
}

// SetRoleApps 角色绑定应用（校验租户已订阅）
// PUT /api/v1/roles/:id/apps
func (h *ApplicationHandler) SetRoleApps(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}

	var req service.SetRoleAppsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	// 查询角色获取 tenant_id
	role, err := h.roleRepo.FindByID(h.db, roleID)
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.SetRoleApps(roleID, role.TenantID, &req); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}
