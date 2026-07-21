package service

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

// OrgUnitNode 组织架构树节点
type OrgUnitNode struct {
	model.OrgUnit
	Children []*OrgUnitNode `json:"children"`
}

// OrgUnitService 组织架构业务逻辑层
type OrgUnitService struct {
	db      *gorm.DB
	repo    repository.OrgUnitRepository
	logger  OperationLogger
}

// NewOrgUnitService 创建 OrgUnitService 实例
func NewOrgUnitService(db *gorm.DB, repo repository.OrgUnitRepository) *OrgUnitService {
	return &OrgUnitService{db: db, repo: repo, logger: &noopLogger{}}
}

// SetLogger 设置操作日志记录器
func (s *OrgUnitService) SetLogger(logger OperationLogger) {
	s.logger = logger
}

// CreateOrgUnitRequest 创建组织节点请求
type CreateOrgUnitRequest struct {
	TenantID  int64  `json:"-"`
	ParentID  *int64 `json:"parent_id,string"`
	NodeType  string `json:"node_type" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code"`
	SortOrder int    `json:"sort_order"`
}

// UpdateOrgUnitRequest 更新组织节点请求
type UpdateOrgUnitRequest struct {
	ParentID  *int64 `json:"parent_id,string"`
	NodeType  string `json:"node_type"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	SortOrder *int   `json:"sort_order"`
	Status    *int   `json:"status"`
	Version   int    `json:"version" binding:"required"`
}

// SetNodeUsersRequest 设置节点用户请求
type SetNodeUsersRequest struct {
	UserIDs   []int64 `json:"user_ids" binding:"required"`
	IsPrimary int     `json:"is_primary"` // 1=设为主归属
}

// CreateOrgUnit 创建组织节点
func (s *OrgUnitService) CreateOrgUnit(req *CreateOrgUnitRequest) (*model.OrgUnit, error) {
	// code 唯一性校验
	if req.Code != "" {
		existing, err := s.repo.FindByTenantAndCode(s.db, req.TenantID, req.Code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "组织编码已存在")
		}
	}

	unit := &model.OrgUnit{
		TenantID:  req.TenantID,
		ParentID:  req.ParentID,
		NodeType:  req.NodeType,
		Name:      req.Name,
		Code:      req.Code,
		SortOrder: req.SortOrder,
		Status:    1,
		Version:   1,
	}

	if err := s.repo.Create(s.db, unit); err != nil {
		return nil, err
	}

	s.logger.Log(0, "create_org_unit", "org_unit", unit.ID, "创建组织节点: "+unit.Name)
	return unit, nil
}

// UpdateOrgUnit 更新组织节点（含乐观锁校验）
func (s *OrgUnitService) UpdateOrgUnit(id int64, req *UpdateOrgUnitRequest) (*model.OrgUnit, error) {
	unit, err := s.repo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}

	// 乐观锁版本校验
	if unit.Version != req.Version {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}

	if req.ParentID != nil {
		unit.ParentID = req.ParentID
	}
	if req.NodeType != "" {
		unit.NodeType = req.NodeType
	}
	if req.Name != "" {
		unit.Name = req.Name
	}
	if req.Code != "" {
		// code 变更时校验唯一性
		if req.Code != unit.Code {
			existing, err := s.repo.FindByTenantAndCode(s.db, unit.TenantID, req.Code)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "组织编码已存在")
			}
		}
		unit.Code = req.Code
	}
	if req.SortOrder != nil {
		unit.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		unit.Status = *req.Status
	}

	unit.Version++
	if err := s.repo.Update(s.db, unit); err != nil {
		return nil, err
	}

	s.logger.Log(0, "update_org_unit", "org_unit", unit.ID, "更新组织节点: "+unit.Name)
	return unit, nil
}

// DeleteOrgUnit 删除组织节点（有子节点禁止删除）
func (s *OrgUnitService) DeleteOrgUnit(id int64) error {
	_, err := s.repo.FindByID(s.db, id)
	if err != nil {
		return err
	}

	hasChildren, err := s.repo.HasChildren(s.db, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return errors.NewAuthError(errors.ErrProtectedEntity, "该节点下存在子节点，无法删除")
	}

	if err := s.repo.Delete(s.db, id); err != nil {
		return err
	}

	s.logger.Log(0, "delete_org_unit", "org_unit", id, "")
	return nil
}

// GetTree 获取租户组织架构树
func (s *OrgUnitService) GetTree(tenantID int64) ([]*OrgUnitNode, error) {
	units, err := s.repo.ListByTenant(s.db, tenantID)
	if err != nil {
		return nil, err
	}
	return buildOrgTree(units), nil
}

// GetNodeUsers 获取节点下用户 ID 列表
func (s *OrgUnitService) GetNodeUsers(orgUnitID, tenantID int64) ([]model.UserOrg, error) {
	var userOrgs []model.UserOrg
	err := s.db.Where("org_unit_id = ? AND tenant_id = ?", orgUnitID, tenantID).Find(&userOrgs).Error
	return userOrgs, err
}

// SetNodeUsers 设置节点用户（全量替换）
func (s *OrgUnitService) SetNodeUsers(orgUnitID, tenantID int64, req *SetNodeUsersRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除该节点下当前关联
		if err := tx.Where("org_unit_id = ? AND tenant_id = ?", orgUnitID, tenantID).
			Delete(&model.UserOrg{}).Error; err != nil {
			return err
		}
		// 批量插入新关联
		if len(req.UserIDs) == 0 {
			return nil
		}
		userOrgs := make([]model.UserOrg, 0, len(req.UserIDs))
		for _, uid := range req.UserIDs {
			userOrgs = append(userOrgs, model.UserOrg{
				UserID:    uid,
				OrgUnitID: orgUnitID,
				TenantID:  tenantID,
				IsPrimary: req.IsPrimary,
			})
		}
		return tx.Create(&userOrgs).Error
	})
}

// buildOrgTree 构建组织架构树
func buildOrgTree(units []model.OrgUnit) []*OrgUnitNode {
	if len(units) == 0 {
		return []*OrgUnitNode{}
	}
	nodeMap := make(map[int64]*OrgUnitNode)
	for i := range units {
		nodeMap[units[i].ID] = &OrgUnitNode{OrgUnit: units[i]}
	}
	var roots []*OrgUnitNode
	for i := range units {
		node := nodeMap[units[i].ID]
		if node.ParentID == nil || *node.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[*node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	if roots == nil {
		return []*OrgUnitNode{}
	}
	return roots
}
