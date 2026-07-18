package handler

import (
	"strconv"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理 HTTP Handler
type UserHandler struct {
	svc     *service.UserService
	roleSvc *service.UserRoleService
}

// NewUserHandler 创建 UserHandler 实例
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// SetRoleService 注入 UserRoleService
func (h *UserHandler) SetRoleService(roleSvc *service.UserRoleService) {
	h.roleSvc = roleSvc
}

// RegisterRoutes 注册用户管理路由
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
		users.POST("/:id/tenants", h.AssociateTenant)
		users.DELETE("/:id/tenants/:tenantId", h.DissociateTenant)
		users.GET("/:id/tenants", h.ListUserTenants)
		users.POST("/:id/roles", h.AssignRoles)
		users.PUT("/:id/roles", h.ReplaceRoles)
		users.POST("/:id/force-offline", h.ForceOffline)
	}
}

// List 分页查询用户列表（按租户上下文过滤）
// GET /api/v1/users?page=1&page_size=20
func (h *UserHandler) List(c *gin.Context) {
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListUsers(authCtx.TenantID, page, pageSize)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, result)
}

// Create 创建全局用户
// POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	user, err := h.svc.CreateUser(&req)
	if err != nil {
		Error(c, err)
		return
	}

	// 自动关联新用户到当前租户
	authCtx, _ := MustGetAuthContext(c)
	if authCtx != nil && authCtx.TenantID > 0 {
		_ = h.svc.AssociateTenant(user.ID, authCtx.TenantID)
	}

	Success(c, user)
}

// Update 更新用户信息（乐观锁）
// PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的用户 ID"))
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	user, err := h.svc.UpdateUser(id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, user)
}

// Delete 软删除用户
// DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的用户 ID"))
		return
	}

	if err := h.svc.DeleteUser(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// AssociateTenant 关联用户到租户（支持批量）
// POST /api/v1/users/:id/tenants
func (h *UserHandler) AssociateTenant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的用户 ID"))
		return
	}

	var req service.AssociateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	// 兼容单个 tenant_id 和批量 tenant_ids
	ids := req.TenantIDs
	if len(ids) == 0 && req.TenantID > 0 {
		ids = []int64{req.TenantID}
	}
	if len(ids) == 0 {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请至少提供一个租户 ID"))
		return
	}

	for _, tid := range ids {
		if err := h.svc.AssociateTenant(id, tid); err != nil {
			Error(c, err)
			return
		}
	}
	Success(c, nil)
}

// DissociateTenant 解除用户-租户关联（级联删除角色绑定）
// DELETE /api/v1/users/:id/tenants/:tenantId
func (h *UserHandler) DissociateTenant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的用户 ID"))
		return
	}

	tenantID, err := strconv.ParseInt(c.Param("tenantId"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的租户 ID"))
		return
	}

	if err := h.svc.DissociateTenant(id, tenantID); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// ListUserTenants 查询用户已关联租户列表
// GET /api/v1/users/:id/tenants
func (h *UserHandler) ListUserTenants(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的用户 ID"))
		return
	}

	list, err := h.svc.ListUserTenants(id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, list)
}

// AssignRoles 为用户分配角色（追加模式）
// POST /api/v1/users/:id/roles
func (h *UserHandler) AssignRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的用户 ID"))
		return
	}

	var req service.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	if err := h.roleSvc.AssignRoles(id, &req); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// ReplaceRoles 全量替换用户角色
// PUT /api/v1/users/:id/roles
func (h *UserHandler) ReplaceRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的用户 ID"))
		return
	}

	var req service.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	if err := h.roleSvc.ReplaceRoles(id, &req); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// ForceOffline 强制下线用户（禁用账户）
// POST /api/v1/users/:id/force-offline
func (h *UserHandler) ForceOffline(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的用户 ID"))
		return
	}

	if err := h.svc.ForceOffline(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}
