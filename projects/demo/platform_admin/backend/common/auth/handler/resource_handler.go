package handler

import (
	"strconv"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// ResourceHandler 资源管理 HTTP Handler
type ResourceHandler struct {
	svc *service.ResourceService
}

// NewResourceHandler 创建 ResourceHandler 实例
func NewResourceHandler(svc *service.ResourceService) *ResourceHandler {
	return &ResourceHandler{svc: svc}
}

// RegisterRoutes 注册资源管理路由（admin 组下的 CRUD 接口）
func (h *ResourceHandler) RegisterRoutes(rg *gin.RouterGroup) {
	resources := rg.Group("/resources")
	{
		resources.GET("/tree", h.GetTree)
		resources.POST("", h.Create)
		resources.PUT("/sort", h.Sort)
		resources.PUT("/:id", h.Update)
		resources.DELETE("/:id", h.Delete)
	}
}

// GetTree 获取资源树
// GET /api/v1/resources/tree?app_code=xxx
func (h *ResourceHandler) GetTree(c *gin.Context) {
	appCode := c.Query("app_code")
	if appCode == "" {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "app_code 参数必传"))
		return
	}

	tree, err := h.svc.GetTree(appCode)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, tree)
}

// GetUserMenu 获取用户有权限的菜单树
// GET /api/v1/common/user-menu?platform=admin
func (h *ResourceHandler) GetUserMenu(c *gin.Context) {
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}

	platform := c.DefaultQuery("platform", "admin")

	tree, err := h.svc.GetUserMenu(authCtx.TenantID, authCtx.UserID, platform)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, tree)
}

// Create 创建资源
// POST /api/v1/resources
func (h *ResourceHandler) Create(c *gin.Context) {
	var req service.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	resource, err := h.svc.CreateResource(&req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, resource)
}

// Update 更新资源
// PUT /api/v1/resources/:id
func (h *ResourceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的资源 ID"))
		return
	}

	var req service.UpdateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	resource, err := h.svc.UpdateResource(id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, resource)
}

// Delete 删除资源（含子节点检查）
// DELETE /api/v1/resources/:id
func (h *ResourceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的资源 ID"))
		return
	}

	if err := h.svc.DeleteResource(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// Sort 拖拽排序（批量更新 sort_order + parent_id）
// PUT /api/v1/resources/sort
func (h *ResourceHandler) Sort(c *gin.Context) {
	var req service.SortResourcesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	if err := h.svc.SortResources(req.Items); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}
