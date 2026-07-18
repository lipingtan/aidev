package handler

import (
	"strconv"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// DataScopeHandler 数据权限维度管理 HTTP Handler
type DataScopeHandler struct {
	svc *service.DataScopeService
}

// NewDataScopeHandler 创建 DataScopeHandler 实例
func NewDataScopeHandler(svc *service.DataScopeService) *DataScopeHandler {
	return &DataScopeHandler{svc: svc}
}

// RegisterRoutes 注册数据权限维度管理路由
func (h *DataScopeHandler) RegisterRoutes(rg *gin.RouterGroup) {
	configs := rg.Group("/data-scope-configs")
	{
		configs.GET("", h.List)
		configs.POST("", h.Create)
		configs.PUT("/:id", h.Update)
		configs.DELETE("/:id", h.Delete)
	}

	// 角色数据权限路由
	roles := rg.Group("/roles")
	{
		roles.PUT("/:id/data-scopes", h.SetRoleDataScopes)
		roles.GET("/:id/data-scopes", h.GetRoleDataScopes)
	}
}

// List 查询维度注册列表
// GET /api/v1/data-scope-configs
func (h *DataScopeHandler) List(c *gin.Context) {
	list, err := h.svc.ListConfigs()
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, list)
}

// Create 注册维度
// POST /api/v1/data-scope-configs
func (h *DataScopeHandler) Create(c *gin.Context) {
	var req service.CreateDataScopeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	cfg, err := h.svc.CreateConfig(&req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, cfg)
}

// Update 更新维度
// PUT /api/v1/data-scope-configs/:id
func (h *DataScopeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的维度 ID"))
		return
	}

	var req service.UpdateDataScopeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	cfg, err := h.svc.UpdateConfig(id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, cfg)
}

// Delete 删除维度
// DELETE /api/v1/data-scope-configs/:id
func (h *DataScopeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的维度 ID"))
		return
	}

	if err := h.svc.DeleteConfig(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// SetRoleDataScopes 角色配置数据权限
// PUT /api/v1/roles/:id/data-scopes
func (h *DataScopeHandler) SetRoleDataScopes(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}

	var req service.SetRoleDataScopesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	if err := h.svc.SetRoleDataScopes(id, &req); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// GetRoleDataScopes 查询角色数据权限绑定
// GET /api/v1/roles/:id/data-scopes
func (h *DataScopeHandler) GetRoleDataScopes(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}

	scopes, err := h.svc.GetRoleDataScopes(id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, scopes)
}
