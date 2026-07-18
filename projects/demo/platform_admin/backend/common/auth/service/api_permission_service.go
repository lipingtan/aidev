package service

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

// ApiPermissionTreeNode 树形节点（含子节点）
type ApiPermissionTreeNode struct {
	model.ApiPermission
	Children []*ApiPermissionTreeNode `json:"children"`
}

// CreateApiPermissionRequest 创建接口权限请求
type CreateApiPermissionRequest struct {
	ParentID       *int64 `json:"parent_id"`
	Type           string `json:"type" binding:"required,oneof=GROUP ENDPOINT"`
	Name           string `json:"name" binding:"required"`
	PermissionCode string `json:"permission_code"`
	URLPattern     string `json:"url_pattern"`
	HTTPMethod     string `json:"http_method"`
	AppCode        string `json:"app_code" binding:"required"`
	Status         string `json:"status"`
	SortOrder      int    `json:"sort_order"`
}

// UpdateApiPermissionRequest 更新接口权限请求
type UpdateApiPermissionRequest struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	PermissionCode string `json:"permission_code"`
	URLPattern     string `json:"url_pattern"`
	HTTPMethod     string `json:"http_method"`
	AppCode        string `json:"app_code"`
	Status         string `json:"status"`
	SortOrder      int    `json:"sort_order"`
}

// MoveApiPermissionRequest 移动节点请求
type MoveApiPermissionRequest struct {
	ParentID int64 `json:"parent_id" binding:"required"`
}

// ApiPermissionService 接口权限业务逻辑层
type ApiPermissionService struct {
	db   *gorm.DB
	repo repository.ApiPermissionRepository
}

// NewApiPermissionService 创建 ApiPermissionService 实例
func NewApiPermissionService(db *gorm.DB, repo repository.ApiPermissionRepository) *ApiPermissionService {
	return &ApiPermissionService{db: db, repo: repo}
}

// Create 创建接口权限节点
func (s *ApiPermissionService) Create(req *CreateApiPermissionRequest) (*model.ApiPermission, error) {
	status := req.Status
	if status == "" {
		if req.Type == "ENDPOINT" && req.ParentID == nil {
			status = "UNASSIGNED"
		} else {
			status = "ACTIVE"
		}
	}

	perm := &model.ApiPermission{
		ParentID:       req.ParentID,
		Type:           req.Type,
		Name:           req.Name,
		PermissionCode: req.PermissionCode,
		URLPattern:     req.URLPattern,
		HTTPMethod:     req.HTTPMethod,
		AppCode:        req.AppCode,
		Status:         status,
		SortOrder:      req.SortOrder,
	}

	if err := s.repo.Create(s.db, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

// Update 更新接口权限节点
func (s *ApiPermissionService) Update(id int64, req *UpdateApiPermissionRequest) (*model.ApiPermission, error) {
	perm, err := s.repo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		perm.Name = req.Name
	}
	// display_name 始终更新，让前端控制
	perm.DisplayName = req.DisplayName
	if req.PermissionCode != "" {
		perm.PermissionCode = req.PermissionCode
	}
	if req.URLPattern != "" {
		perm.URLPattern = req.URLPattern
	}
	if req.HTTPMethod != "" {
		perm.HTTPMethod = req.HTTPMethod
	}
	if req.AppCode != "" {
		perm.AppCode = req.AppCode
	}
	if req.Status != "" {
		perm.Status = req.Status
	}
	perm.SortOrder = req.SortOrder

	if err := s.repo.Update(s.db, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

// Delete 删除接口权限节点（有子节点时返回 ErrProtectedEntity）
func (s *ApiPermissionService) Delete(id int64) error {
	// 检查节点是否存在
	if _, err := s.repo.FindByID(s.db, id); err != nil {
		return err
	}

	// 检查是否有子节点
	hasChildren, err := s.repo.HasChildren(s.db, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return errors.NewAuthError(errors.ErrProtectedEntity, "该节点下有子节点，无法删除")
	}

	return s.repo.Delete(s.db, id)
}

// GetTree 获取接口权限树（按 app_code）
func (s *ApiPermissionService) GetTree(appCode string) ([]*ApiPermissionTreeNode, error) {
	perms, err := s.repo.ListByAppCode(s.db, appCode)
	if err != nil {
		return nil, err
	}
	return buildApiPermissionTree(perms), nil
}

// Move 移动节点到指定 GROUP 下
func (s *ApiPermissionService) Move(id int64, req *MoveApiPermissionRequest) error {
	// 检查源节点存在
	if _, err := s.repo.FindByID(s.db, id); err != nil {
		return err
	}

	// 校验目标节点必须为 GROUP
	target, err := s.repo.FindByID(s.db, req.ParentID)
	if err != nil {
		return err
	}
	if target.Type != "GROUP" {
		return errors.NewAuthError(errors.ErrProtectedEntity, "目标节点必须为 GROUP 类型")
	}

	return s.repo.Move(s.db, id, req.ParentID)
}

// ListUnassigned 获取未分组 ENDPOINT 列表
func (s *ApiPermissionService) ListUnassigned(appCode string) ([]model.ApiPermission, error) {
	return s.repo.ListUnassigned(s.db, appCode)
}

// SetVisible 设置节点显示/隐藏状态
// 如果是 GROUP，同时更新所有子节点的 visible
func (s *ApiPermissionService) SetVisible(id int64, visible int) error {
	perm, err := s.repo.FindByID(s.db, id)
	if err != nil {
		return err
	}
	// 更新自身
	s.db.Model(&model.ApiPermission{}).Where("id = ?", id).Update("visible", visible)
	// 如果是 GROUP，级联更新子节点
	if perm.Type == "GROUP" {
		s.db.Model(&model.ApiPermission{}).Where("parent_id = ?", id).Update("visible", visible)
	}
	return nil
}

// buildApiPermissionTree 构建树形结构
func buildApiPermissionTree(perms []model.ApiPermission) []*ApiPermissionTreeNode {
	nodeMap := make(map[int64]*ApiPermissionTreeNode, len(perms))
	var roots []*ApiPermissionTreeNode

	// 第一遍：创建所有节点
	for i := range perms {
		nodeMap[perms[i].ID] = &ApiPermissionTreeNode{
			ApiPermission: perms[i],
			Children:      make([]*ApiPermissionTreeNode, 0),
		}
	}

	// 第二遍：建立父子关系
	for i := range perms {
		node := nodeMap[perms[i].ID]
		if perms[i].ParentID == nil || *perms[i].ParentID == 0 {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[*perms[i].ParentID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// 父节点不存在时作为根节点
				roots = append(roots, node)
			}
		}
	}

	if roots == nil {
		roots = make([]*ApiPermissionTreeNode, 0)
	}
	return roots
}
