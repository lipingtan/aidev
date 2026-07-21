package handler

import (
	"strconv"
	"time"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

// RoleHandler 角色管理 HTTP Handler
type RoleHandler struct {
	svc *service.RoleService
}

// NewRoleHandler 创建 RoleHandler 实例
func NewRoleHandler(svc *service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// RegisterRoutes 注册角色管理路由
func (h *RoleHandler) RegisterRoutes(rg *gin.RouterGroup) {
	roles := rg.Group("/roles")
	{
		roles.GET("", h.List)
		roles.POST("", h.Create)
		roles.PUT("/:id", h.Update)
		roles.DELETE("/:id", h.Delete)
		roles.GET("/:id/resources", h.GetResources)
		roles.PUT("/:id/resources", h.AssignResources)
		roles.GET("/:id/apis", h.GetApis)
		roles.PUT("/:id/apis", h.AssignApis)
		roles.GET("/:id/apps", h.GetApps)
		roles.GET("/:id/permission-summary", h.GetPermissionSummary)
		roles.GET("/:id/assignable-resources", h.AssignableResources)
		roles.GET("/:id/assignable-apis", h.AssignableApis)
	}
}

// RoleTreeNode 角色树节点（含子节点）
type RoleTreeNode struct {
	ID        int64          `json:"id,string"`
	TenantID  int64          `json:"tenant_id,string"`
	RoleCode  string         `json:"role_code"`
	RoleName  string         `json:"role_name"`
	RoleType  string         `json:"role_type"`
	ParentID  *int64         `json:"parent_id,string"`
	SortOrder int            `json:"sort_order"`
	Status    int            `json:"status"`
	Version   int            `json:"version"`
	CreatedAt *time.Time     `json:"created_at"`
	Children  []*RoleTreeNode `json:"children"`
}

// List 查询当前租户角色列表（树形结构）
// GET /api/v1/roles?role_type=PERMISSION_SET
func (h *RoleHandler) List(c *gin.Context) {
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}

	// 从 query param 读取 role_type 过滤条件
	roleType := c.Query("role_type")

	roles, err := h.svc.ListRoles(authCtx.TenantID, roleType)
	if err != nil {
		Error(c, err)
		return
	}

	// 构建树形结构
	tree := buildRoleTree(roles)
	Success(c, tree)
}

// buildRoleTree 将扁平角色列表构建为树形结构
func buildRoleTree(roles []model.Role) []*RoleTreeNode {
	nodeMap := make(map[int64]*RoleTreeNode)
	var roots []*RoleTreeNode

	// 创建所有节点
	for i := range roles {
		r := &roles[i]
		node := &RoleTreeNode{
			ID:        r.ID,
			TenantID:  r.TenantID,
			RoleCode:  r.RoleCode,
			RoleName:  r.RoleName,
			RoleType:  r.RoleType,
			ParentID:  r.ParentID,
			SortOrder: r.SortOrder,
			Status:    r.Status,
			Version:   r.Version,
			CreatedAt: r.CreatedAt,
			Children:  make([]*RoleTreeNode, 0),
		}
		nodeMap[r.ID] = node
	}

	// 构建树
	for _, r := range roles {
		node := nodeMap[r.ID]
		if r.ParentID != nil && *r.ParentID != 0 {
			if parent, ok := nodeMap[*r.ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	return roots
}

// Create 创建角色
// POST /api/v1/roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}
	req.TenantID = authCtx.TenantID // 强制覆写

	role, err := h.svc.CreateRole(&req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, role)
}

// Update 更新角色（乐观锁）
// PUT /api/v1/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	role, err := h.svc.UpdateRole(id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, role)
}

// Delete 删除角色（软删除，有用户绑定时拒绝）
// DELETE /api/v1/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}

	if err := h.svc.DeleteRole(id); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// AssignResources 分配角色菜单权限（全量替换）
// PUT /api/v1/roles/:id/resources
func (h *RoleHandler) AssignResources(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}

	var req service.AssignResourcesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	result, err := h.svc.AssignResources(id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, result)
}

// AssignApis 分配角色接口权限（全量替换，含父角色子集校验 + 级联裁剪）
// PUT /api/v1/roles/:id/apis
func (h *RoleHandler) AssignApis(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}

	var req service.AssignApisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewAuthError(errors.ErrDuplicateEntity, "请求参数无效: "+err.Error()))
		return
	}

	result, err := h.svc.AssignApis(id, &req)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, result)
}

// GetResources 查询角色已分配的资源 ID 列表
// GET /api/v1/roles/:id/resources
func (h *RoleHandler) GetResources(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}
	ids, err := h.svc.GetRoleResourceIDs(id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, ids)
}

// GetApis 查询角色已分配的接口权限 ID 列表
// GET /api/v1/roles/:id/apis
func (h *RoleHandler) GetApis(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}
	ids, err := h.svc.GetRoleApiIDs(id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, ids)
}

// GetApps 查询角色已绑定的应用编码列表
// GET /api/v1/roles/:id/apps
func (h *RoleHandler) GetApps(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}
	codes, err := h.svc.GetRoleAppCodes(id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, codes)
}

// AssignableResources 查询角色可分配给子角色的资源 ID 范围
// GET /api/v1/roles/:id/assignable-resources
func (h *RoleHandler) AssignableResources(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}
	ids, err := h.svc.GetAssignableResources(id, authCtx.TenantID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, ids)
}

// AssignableApis 查询角色可分配给子角色的 API 权限 ID 范围
// GET /api/v1/roles/:id/assignable-apis
func (h *RoleHandler) AssignableApis(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}
	authCtx, err := MustGetAuthContext(c)
	if err != nil {
		Error(c, err)
		return
	}
	ids, err := h.svc.GetAssignableApis(id, authCtx.TenantID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, ids)
}

// GetPermissionSummary 查询角色在每个应用下的权限统计摘要
// GET /api/v1/roles/:id/permission-summary
func (h *RoleHandler) GetPermissionSummary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewAuthError(errors.ErrEntityNotFound, "无效的角色 ID"))
		return
	}
	summary, err := h.svc.GetPermissionSummary(id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, summary)
}
