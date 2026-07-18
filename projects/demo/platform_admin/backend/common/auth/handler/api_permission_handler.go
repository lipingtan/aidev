package handler

import (
	"strconv"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// ApiPermissionHandler 接口权限管理 HTTP Handler
type ApiPermissionHandler struct {
	svc *service.ApiPermissionService
}

// NewApiPermissionHandler 创建 ApiPermissionHandler 实例
func NewApiPermissionHandler(svc *service.ApiPermissionService) *ApiPermissionHandler {
	return &ApiPermissionHandler{svc: svc}
}

// RegisterRoutes 注册接口权限管理路由
func (h *ApiPermissionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	perms := rg.Group("/api-permissions")
	{
		perms.GET("/tree", h.GetTree)
		perms.GET("/unassigned", h.ListUnassigned)
		perms.POST("", h.Create)
		perms.PUT("/:id", h.Update)
		perms.DELETE("/:id", h.Delete)
		perms.PUT("/:id/move", h.Move)
		perms.PUT("/:id/visible", h.ToggleVisible)
	}
}

// GetTree 获取接口权限树
// GET /api/v1/api-permissions/tree?app_code=xxx
func (h *ApiPermissionHandler) GetTree(c *gin.Context) {
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

// Create 创建接口权限节点
// POST /api/v1/api-permissions
func (h *ApiPermissionHandler) Create(c *gin.Context) {
	var req service.CreateApiPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	perm, err := h.svc.Create(&req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, perm)
}

// Update 更新接口权限节点
// PUT /api/v1/api-permissions/:id
func (h *ApiPermissionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的节点 ID"))
		return
	}

	var req service.UpdateApiPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	perm, err := h.svc.Update(id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, perm)
}

// Delete 删除接口权限节点
// DELETE /api/v1/api-permissions/:id
func (h *ApiPermissionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的节点 ID"))
		return
	}

	if err := h.svc.Delete(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// Move 移动节点到指定 GROUP 下
// PUT /api/v1/api-permissions/:id/move
func (h *ApiPermissionHandler) Move(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的节点 ID"))
		return
	}

	var req service.MoveApiPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	if err := h.svc.Move(id, &req); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// ToggleVisible 切换节点显示/隐藏状态
// PUT /api/v1/api-permissions/:id/visible
func (h *ApiPermissionHandler) ToggleVisible(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的节点 ID"))
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "参数错误"))
		return
	}
	visible, ok := body["visible"]
	if !ok {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "缺少 visible 字段"))
		return
	}
	v := int(visible.(float64))

	if err := h.svc.SetVisible(id, v); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"id": id, "visible": v})
}

// ListUnassigned 获取未分组 ENDPOINT 列表
// GET /api/v1/api-permissions/unassigned?app_code=xxx
func (h *ApiPermissionHandler) ListUnassigned(c *gin.Context) {
	appCode := c.Query("app_code")
	if appCode == "" {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "app_code 参数必传"))
		return
	}

	list, err := h.svc.ListUnassigned(appCode)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, list)
}
