package service

import (
	"fmt"

	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

// RoleService 角色业务逻辑层
type RoleService struct {
	db       *gorm.DB
	cfg      *config.Config
	roleRepo repository.RoleRepository
	logger   OperationLogger
}

// NewRoleService 创建 RoleService 实例
func NewRoleService(db *gorm.DB, cfg *config.Config, roleRepo repository.RoleRepository) *RoleService {
	return &RoleService{
		db:       db,
		cfg:      cfg,
		roleRepo: roleRepo,
		logger:   &noopLogger{},
	}
}

// SetLogger 设置操作日志记录器
func (s *RoleService) SetLogger(logger OperationLogger) {
	s.logger = logger
}

// CreateRoleRequest 创建角色请求参数
type CreateRoleRequest struct {
	TenantID  int64  `json:"tenant_id"`
	RoleCode  string `json:"role_code" binding:"required"`
	RoleName  string `json:"role_name" binding:"required"`
	RoleType  string `json:"role_type"`
	ParentID  *int64 `json:"parent_id,string"`
	SortOrder int    `json:"sort_order"`
}

// UpdateRoleRequest 更新角色请求参数
type UpdateRoleRequest struct {
	RoleName  string `json:"role_name"`
	RoleType  string `json:"role_type"`
	ParentID  *int64 `json:"parent_id,string"`
	SortOrder *int   `json:"sort_order"`
	Status    *int   `json:"status"`
	Version   int    `json:"version" binding:"required"`
}

// CreateRole 创建角色（校验 tenant_id + role_code 唯一性）
func (s *RoleService) CreateRole(req *CreateRoleRequest) (*model.Role, error) {
	// 检查 tenant_id + role_code 唯一性
	existing, err := s.roleRepo.FindByCode(s.db, req.TenantID, req.RoleCode)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "角色编码在该租户下已存在")
	}

	// 校验 parent_id 循环引用和深度
	if req.ParentID != nil {
		if err := s.checkCyclicAndDepth(0, *req.ParentID); err != nil {
			return nil, err
		}
	}

	role := &model.Role{
		TenantID:  req.TenantID,
		RoleCode:  req.RoleCode,
		RoleName:  req.RoleName,
		RoleType:  req.RoleType,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
		Status:    1,
		Version:   1,
	}

	if err := s.roleRepo.Create(s.db, role); err != nil {
		return nil, err
	}

	s.logger.Log(0, "create_role", "role", role.ID, "创建角色: "+role.RoleName)
	return role, nil
}

// UpdateRole 更新角色（乐观锁 + 校验 parent_id 循环引用和深度）
func (s *RoleService) UpdateRole(id int64, req *UpdateRoleRequest) (*model.Role, error) {
	role, err := s.roleRepo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}

	// 乐观锁版本校验
	if role.Version != req.Version {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}

	// 校验 parent_id 变更时的循环引用和深度
	if req.ParentID != nil {
		if err := s.checkCyclicAndDepth(id, *req.ParentID); err != nil {
			return nil, err
		}
	}

	// 更新字段
	if req.RoleName != "" {
		role.RoleName = req.RoleName
	}
	if req.RoleType != "" {
		role.RoleType = req.RoleType
	}
	if req.ParentID != nil {
		role.ParentID = req.ParentID
	}
	if req.SortOrder != nil {
		role.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := s.roleRepo.Update(s.db, role); err != nil {
		return nil, err
	}

	s.logger.Log(0, "update_role", "role", role.ID, "更新角色: "+role.RoleName)
	return role, nil
}

// DeleteRole 软删除角色（检查用户绑定）
func (s *RoleService) DeleteRole(id int64) error {
	// 检查角色是否存在
	_, err := s.roleRepo.FindByID(s.db, id)
	if err != nil {
		return err
	}

	// 检查是否有用户绑定
	hasBinding, err := s.roleRepo.HasUserBinding(s.db, id)
	if err != nil {
		return err
	}
	if hasBinding {
		return errors.NewAuthError(errors.ErrRoleHasUsers, "角色下存在关联用户，无法删除")
	}

	if err := s.roleRepo.SoftDelete(s.db, id); err != nil {
		return err
	}

	s.logger.Log(0, "delete_role", "role", id, "")
	return nil
}

// AssignResourcesRequest 分配菜单权限请求参数
type AssignResourcesRequest struct {
	ResourceIDs StringInt64Slice `json:"resource_ids" binding:"required"`
}

// AssignResources 全量替换角色菜单权限
func (s *RoleService) AssignResources(roleID int64, req *AssignResourcesRequest) error {
	// 校验角色存在
	_, err := s.roleRepo.FindByID(s.db, roleID)
	if err != nil {
		return err
	}

	// 自动绑定资源对应的应用
	if len(req.ResourceIDs) > 0 {
		var resourceAppCodes []string
		s.db.Model(&model.Resource{}).Where("id IN ?", req.ResourceIDs).Distinct().Pluck("app_code", &resourceAppCodes)

		var existingAppCodes []string
		s.db.Model(&model.RoleApp{}).Where("role_id = ?", roleID).Pluck("app_code", &existingAppCodes)
		existingMap := make(map[string]bool)
		for _, c := range existingAppCodes {
			existingMap[c] = true
		}

		for _, appCode := range resourceAppCodes {
			if !existingMap[appCode] {
				s.db.Create(&model.RoleApp{RoleID: roleID, AppCode: appCode})
			}
		}
	}

	// 事务中全量替换
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧记录
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleResource{}).Error; err != nil {
			return err
		}
		// 批量插入新记录
		for _, resID := range req.ResourceIDs {
			rr := &model.RoleResource{
				RoleID:     roleID,
				ResourceID: resID,
			}
			if err := tx.Create(rr).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	s.logger.Log(0, "assign_resources", "role", roleID, "分配菜单权限")
	return nil
}

// ListRoles 按 tenant_id 查询角色列表
func (s *RoleService) ListRoles(tenantID int64) ([]model.Role, error) {
	return s.roleRepo.ListByTenantID(s.db, tenantID)
}

// AssignApisRequest 角色-接口权限分配请求
type AssignApisRequest struct {
	ApiPermissionIDs StringInt64Slice `json:"api_permission_ids"`
}

// AssignApis 全量替换角色的接口权限
// 逻辑：
// 1. 查询所有传入 ID 的 ApiPermission 记录
// 2. GROUP 类型自动展开为其所有子 ENDPOINT
// 3. 最终只保留 ENDPOINT 类型的 ID
// 4. 事务中删除旧绑定并批量插入新绑定
func (s *RoleService) AssignApis(roleID int64, req *AssignApisRequest) error {
	// 校验角色存在
	_, err := s.roleRepo.FindByID(s.db, roleID)
	if err != nil {
		return err
	}

	// 收集最终的 ENDPOINT ID 列表
	endpointIDs := make([]int64, 0)

	if len(req.ApiPermissionIDs) > 0 {
		// 查询所有传入 ID 的 ApiPermission 记录
		var perms []model.ApiPermission
		if err := s.db.Where("id IN ?", req.ApiPermissionIDs).Find(&perms).Error; err != nil {
			return err
		}

		// 收集 GROUP ID，用于展开子 ENDPOINT
		groupIDs := make([]int64, 0)
		for _, perm := range perms {
			if perm.Type == "ENDPOINT" {
				endpointIDs = append(endpointIDs, perm.ID)
			} else if perm.Type == "GROUP" {
				groupIDs = append(groupIDs, perm.ID)
			}
		}

		// 展开 GROUP：查询 parent_id IN groupIDs AND type = 'ENDPOINT'
		if len(groupIDs) > 0 {
			var childEndpoints []model.ApiPermission
			if err := s.db.Where("parent_id IN ? AND type = ?", groupIDs, "ENDPOINT").Find(&childEndpoints).Error; err != nil {
				return err
			}
			for _, ep := range childEndpoints {
				endpointIDs = append(endpointIDs, ep.ID)
			}
		}

		// 去重
		endpointIDs = uniqueInt64(endpointIDs)
	}

	// 自动绑定接口权限对应的应用
	if len(endpointIDs) > 0 {
		var apiAppCodes []string
		s.db.Model(&model.ApiPermission{}).Where("id IN ?", endpointIDs).Distinct().Pluck("app_code", &apiAppCodes)

		var existingAppCodes []string
		s.db.Model(&model.RoleApp{}).Where("role_id = ?", roleID).Pluck("app_code", &existingAppCodes)
		existingMap := make(map[string]bool)
		for _, c := range existingAppCodes {
			existingMap[c] = true
		}

		for _, appCode := range apiAppCodes {
			if !existingMap[appCode] {
				s.db.Create(&model.RoleApp{RoleID: roleID, AppCode: appCode})
			}
		}
	}

	// 事务：删除旧绑定 + 批量插入新绑定
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧绑定
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleApi{}).Error; err != nil {
			return err
		}

		// 批量插入新绑定
		for _, epID := range endpointIDs {
			ra := &model.RoleApi{
				RoleID:          roleID,
				ApiPermissionID: epID,
			}
			if err := tx.Create(ra).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	s.logger.Log(0, "assign_apis", "role", roleID, fmt.Sprintf("分配接口权限，ENDPOINT 数量: %d", len(endpointIDs)))
	return nil
}

// uniqueInt64 对 int64 切片去重
func uniqueInt64(ids []int64) []int64 {
	seen := make(map[int64]bool, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

// GetRoleResourceIDs 查询角色已分配的资源 ID 列表（返回字符串格式）
func (s *RoleService) GetRoleResourceIDs(roleID int64) ([]string, error) {
	var ids []int64
	err := s.db.Model(&model.RoleResource{}).Where("role_id = ?", roleID).Pluck("resource_id", &ids).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = fmt.Sprintf("%d", id)
	}
	return result, nil
}

// GetRoleApiIDs 查询角色已分配的接口权限 ID 列表（返回字符串格式）
func (s *RoleService) GetRoleApiIDs(roleID int64) ([]string, error) {
	var ids []int64
	err := s.db.Model(&model.RoleApi{}).Where("role_id = ?", roleID).Pluck("api_permission_id", &ids).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = fmt.Sprintf("%d", id)
	}
	return result, nil
}

// GetRoleAppCodes 查询角色已绑定的应用编码列表
func (s *RoleService) GetRoleAppCodes(roleID int64) ([]string, error) {
	var codes []string
	err := s.db.Model(&model.RoleApp{}).Where("role_id = ?", roleID).Pluck("app_code", &codes).Error
	return codes, err
}

// AppPermissionSummary 角色在某应用下的权限统计摘要
type AppPermissionSummary struct {
	AppCode        string `json:"app_code"`
	AppName        string `json:"app_name"`
	Bound          bool   `json:"bound"`
	MenuCount      int64  `json:"menu_count"`
	ApiCount       int64  `json:"api_count"`
	DataScopeCount int64  `json:"data_scope_count"`
}

// GetPermissionSummary 查询角色在每个应用下的权限统计
func (s *RoleService) GetPermissionSummary(roleID int64) ([]AppPermissionSummary, error) {
	// 1. 查所有启用的应用
	var apps []model.Application
	if err := s.db.Where("status = 1").Find(&apps).Error; err != nil {
		return nil, err
	}

	// 2. 查角色已绑定的应用编码
	var boundCodes []string
	s.db.Model(&model.RoleApp{}).Where("role_id = ?", roleID).Pluck("app_code", &boundCodes)
	boundMap := make(map[string]bool)
	for _, c := range boundCodes {
		boundMap[c] = true
	}

	// 3. 查角色已分配的资源 ID
	var resourceIDs []int64
	s.db.Model(&model.RoleResource{}).Where("role_id = ?", roleID).Pluck("resource_id", &resourceIDs)

	// 4. 按 app_code 统计菜单数
	menuCountMap := make(map[string]int64)
	if len(resourceIDs) > 0 {
		type countResult struct {
			AppCode string
			Count   int64
		}
		var results []countResult
		s.db.Model(&model.Resource{}).
			Select("app_code as app_code, count(*) as count").
			Where("id IN ?", resourceIDs).
			Group("app_code").
			Find(&results)
		for _, r := range results {
			menuCountMap[r.AppCode] = r.Count
		}
	}

	// 5. 查角色已分配的 API ID
	var apiIDs []int64
	s.db.Model(&model.RoleApi{}).Where("role_id = ?", roleID).Pluck("api_permission_id", &apiIDs)

	// 按 app_code 统计 API 数
	apiCountMap := make(map[string]int64)
	if len(apiIDs) > 0 {
		type countResult struct {
			AppCode string
			Count   int64
		}
		var results []countResult
		s.db.Model(&model.ApiPermission{}).
			Select("app_code as app_code, count(*) as count").
			Where("id IN ?", apiIDs).
			Group("app_code").
			Find(&results)
		for _, r := range results {
			apiCountMap[r.AppCode] = r.Count
		}
	}

	// 6. 组装结果
	summaries := make([]AppPermissionSummary, 0, len(apps))
	for _, app := range apps {
		summaries = append(summaries, AppPermissionSummary{
			AppCode:        app.AppCode,
			AppName:        app.Name,
			Bound:          boundMap[app.AppCode],
			MenuCount:      menuCountMap[app.AppCode],
			ApiCount:       apiCountMap[app.AppCode],
			DataScopeCount: 0,
		})
	}
	return summaries, nil
}

// checkCyclicAndDepth 校验循环引用和层级深度
// roleID: 当前角色 ID（新建时传 0）
// newParentID: 新设置的 parent_id
func (s *RoleService) checkCyclicAndDepth(roleID int64, newParentID int64) error {
	maxDepth := s.cfg.Role.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 5
	}

	visited := make(map[int64]bool)
	if roleID != 0 {
		visited[roleID] = true
	}

	current := &newParentID
	depth := 1

	for current != nil && *current != 0 {
		if visited[*current] {
			return errors.NewAuthError(errors.ErrCyclicHierarchy, "角色层级存在循环引用")
		}
		if depth >= maxDepth {
			return errors.NewAuthError(errors.ErrMaxHierarchyDepth, "角色层级深度超过最大限制")
		}

		parent, err := s.roleRepo.FindByID(s.db, *current)
		if err != nil {
			// 父角色不存在时，中断循环
			return errors.NewAuthError(errors.ErrEntityNotFound, "父角色不存在")
		}
		visited[*current] = true
		current = parent.ParentID
		depth++
	}

	return nil
}
