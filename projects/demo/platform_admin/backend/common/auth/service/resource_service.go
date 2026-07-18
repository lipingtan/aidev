package service

import (
	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

// ResourceNode 资源树节点
type ResourceNode struct {
	model.Resource
	Children []*ResourceNode `json:"children"`
}

// ResourceService 资源业务逻辑层
type ResourceService struct {
	db           *gorm.DB
	cfg          *config.Config
	resourceRepo repository.ResourceRepository
	logger       OperationLogger
}

// NewResourceService 创建 ResourceService 实例
func NewResourceService(db *gorm.DB, cfg *config.Config, resourceRepo repository.ResourceRepository) *ResourceService {
	return &ResourceService{
		db:           db,
		cfg:          cfg,
		resourceRepo: resourceRepo,
		logger:       &noopLogger{},
	}
}

// SetLogger 设置操作日志记录器
func (s *ResourceService) SetLogger(logger OperationLogger) {
	s.logger = logger
}

// CreateResourceRequest 创建资源请求参数
type CreateResourceRequest struct {
	ParentID       *int64 `json:"parent_id"`
	Type           string `json:"type" binding:"required"`
	Name           string `json:"name" binding:"required"`
	PermissionCode string `json:"permission_code"`
	Path           string `json:"path"`
	Component      string `json:"component"`
	Icon           string `json:"icon"`
	AppCode        string `json:"app_code" binding:"required"`
	SortOrder      int    `json:"sort_order"`
}

// UpdateResourceRequest 更新资源请求参数
type UpdateResourceRequest struct {
	ParentID       *int64  `json:"parent_id"`
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	PermissionCode string  `json:"permission_code"`
	Path           string  `json:"path"`
	Component      string  `json:"component"`
	Icon           string  `json:"icon"`
	AppCode        string  `json:"app_code"`
	SortOrder      *int    `json:"sort_order"`
	Status         *int    `json:"status"`
	Version        int     `json:"version" binding:"required"`
}

// SortResourcesRequest 排序请求参数
type SortResourcesRequest struct {
	Items []repository.SortItem `json:"items" binding:"required"`
}

// CreateResource 创建资源
func (s *ResourceService) CreateResource(req *CreateResourceRequest) (*model.Resource, error) {
	// 校验 parent_id 循环引用
	if req.ParentID != nil && *req.ParentID != 0 {
		if err := s.checkCyclicAndDepth(0, *req.ParentID); err != nil {
			return nil, err
		}
	}

	resource := &model.Resource{
		ParentID:       req.ParentID,
		Type:           req.Type,
		Name:           req.Name,
		PermissionCode: req.PermissionCode,
		Path:           req.Path,
		Component:      req.Component,
		Icon:           req.Icon,
		AppCode:        req.AppCode,
		SortOrder:      req.SortOrder,
		Status:         1,
		Version:        1,
	}

	if err := s.resourceRepo.Create(s.db, resource); err != nil {
		return nil, err
	}

	s.logger.Log(0, "create_resource", "resource", resource.ID, "创建资源: "+resource.Name)
	return resource, nil
}

// UpdateResource 更新资源（乐观锁 + 循环引用校验）
func (s *ResourceService) UpdateResource(id int64, req *UpdateResourceRequest) (*model.Resource, error) {
	resource, err := s.resourceRepo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}

	// 乐观锁版本校验
	if resource.Version != req.Version {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}

	// 校验 parent_id 变更时的循环引用
	if req.ParentID != nil {
		if err := s.checkCyclicAndDepth(id, *req.ParentID); err != nil {
			return nil, err
		}
	}

	// 更新字段
	if req.ParentID != nil {
		resource.ParentID = req.ParentID
	}
	if req.Type != "" {
		resource.Type = req.Type
	}
	if req.Name != "" {
		resource.Name = req.Name
	}
	if req.PermissionCode != "" {
		resource.PermissionCode = req.PermissionCode
	}
	if req.Path != "" {
		resource.Path = req.Path
	}
	if req.Component != "" {
		resource.Component = req.Component
	}
	if req.Icon != "" {
		resource.Icon = req.Icon
	}
	if req.AppCode != "" {
		resource.AppCode = req.AppCode
	}
	if req.SortOrder != nil {
		resource.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		resource.Status = *req.Status
	}

	if err := s.resourceRepo.Update(s.db, resource); err != nil {
		return nil, err
	}

	s.logger.Log(0, "update_resource", "resource", resource.ID, "更新资源: "+resource.Name)
	return resource, nil
}

// DeleteResource 删除资源（含子节点检查）
func (s *ResourceService) DeleteResource(id int64) error {
	// 检查资源是否存在
	_, err := s.resourceRepo.FindByID(s.db, id)
	if err != nil {
		return err
	}

	// 检查是否有子节点
	hasChildren, err := s.resourceRepo.HasChildren(s.db, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return errors.NewAuthError(errors.ErrProtectedEntity, "资源下存在子节点，无法删除")
	}

	if err := s.resourceRepo.SoftDelete(s.db, id); err != nil {
		return err
	}

	s.logger.Log(0, "delete_resource", "resource", id, "")
	return nil
}

// GetTree 获取资源树
func (s *ResourceService) GetTree(appCode string) ([]*ResourceNode, error) {
	resources, err := s.resourceRepo.ListByAppCode(s.db, appCode)
	if err != nil {
		return nil, err
	}
	return buildTree(resources), nil
}

// SortResources 批量更新排序
func (s *ResourceService) SortResources(items []repository.SortItem) error {
	return s.resourceRepo.BatchUpdateSort(s.db, items)
}

// GetUserMenu 获取用户有权限的菜单树
// SUPER_ADMIN 角色直接返回所有 MENU 资源
// 普通角色：查角色绑定应用 → 查角色资源 → 过滤 app_code 范围
func (s *ResourceService) GetUserMenu(tenantID int64, userID int64) ([]*ResourceNode, error) {
	// 查询用户在该租户下的角色 ID 列表
	var roleIDs []int64
	err := s.db.Model(&model.UserRole{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Pluck("role_id", &roleIDs).Error
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []*ResourceNode{}, nil
	}

	// 检查是否包含 SUPER_ADMIN 角色 → 直接返回全部菜单
	if s.cfg != nil && s.cfg.Permission.SuperAdminRole != "" {
		var superCount int64
		s.db.Model(&model.Role{}).
			Where("id IN ? AND role_code = ?", roleIDs, s.cfg.Permission.SuperAdminRole).
			Count(&superCount)
		if superCount > 0 {
			var resources []model.Resource
			err = s.db.Where("type = ?", "MENU").
				Order("sort_order ASC, created_at ASC").
				Find(&resources).Error
			if err != nil {
				return nil, err
			}
			return buildTree(resources), nil
		}
	}

	// 普通角色：查角色绑定的应用
	var appCodes []string
	err = s.db.Model(&model.RoleApp{}).
		Where("role_id IN ?", roleIDs).
		Pluck("app_code", &appCodes).Error
	if err != nil {
		return nil, err
	}
	if len(appCodes) == 0 {
		return []*ResourceNode{}, nil
	}

	// 查角色已分配的资源 ID
	var resourceIDs []int64
	err = s.db.Model(&model.RoleResource{}).
		Where("role_id IN ?", roleIDs).
		Pluck("resource_id", &resourceIDs).Error
	if err != nil {
		return nil, err
	}
	if len(resourceIDs) == 0 {
		return []*ResourceNode{}, nil
	}

	// 查询资源详情（仅 MENU 类型，且 app_code 在角色绑定范围内）
	var resources []model.Resource
	err = s.db.Where("id IN ? AND app_code IN ? AND type = ?", resourceIDs, appCodes, "MENU").
		Order("sort_order ASC, created_at ASC").
		Find(&resources).Error
	if err != nil {
		return nil, err
	}

	return buildTree(resources), nil
}

// buildTree 构建资源树（保持 sort_order 排序）
func buildTree(resources []model.Resource) []*ResourceNode {
	if len(resources) == 0 {
		return []*ResourceNode{}
	}
	nodeMap := make(map[int64]*ResourceNode)
	for i := range resources {
		nodeMap[resources[i].ID] = &ResourceNode{Resource: resources[i]}
	}

	var roots []*ResourceNode
	// 按原始 slice 顺序（sort_order ASC）遍历，保证排序
	for i := range resources {
		node := nodeMap[resources[i].ID]
		if node.ParentID == nil || *node.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[*node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	if roots == nil {
		return []*ResourceNode{}
	}
	return roots
}

// checkCyclicAndDepth 校验循环引用和层级深度
func (s *ResourceService) checkCyclicAndDepth(resourceID int64, newParentID int64) error {
	if newParentID == 0 {
		return nil
	}

	maxDepth := s.cfg.Role.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 5
	}

	visited := make(map[int64]bool)
	if resourceID != 0 {
		visited[resourceID] = true
	}

	current := &newParentID
	depth := 1

	for current != nil && *current != 0 {
		if visited[*current] {
			return errors.NewAuthError(errors.ErrCyclicHierarchy, "资源层级存在循环引用")
		}
		if depth >= maxDepth {
			return errors.NewAuthError(errors.ErrMaxHierarchyDepth, "资源层级深度超过最大限制")
		}

		parent, err := s.resourceRepo.FindByID(s.db, *current)
		if err != nil {
			return errors.NewAuthError(errors.ErrEntityNotFound, "父资源不存在")
		}
		visited[*current] = true
		current = parent.ParentID
		depth++
	}

	return nil
}
