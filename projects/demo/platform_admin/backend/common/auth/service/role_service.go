package service

import (
	"fmt"

	"go-admin/common/auth/cache"
	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

// RoleService 角色业务逻辑层
type RoleService struct {
	db        *gorm.DB
	cfg       *config.Config
	roleRepo  repository.RoleRepository
	logger    OperationLogger
	permCache *cache.PermissionCache
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

// SetPermCache 设置权限缓存实例（用于级联裁剪后失效缓存）
func (s *RoleService) SetPermCache(pc *cache.PermissionCache) {
	s.permCache = pc
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

// validRoleTypes 合法的角色类型枚举
var validRoleTypes = map[string]bool{
	"":               true, // 允许空（默认）
	"NORMAL":         true,
	"SUPER_ADMIN":    true,
	"PERMISSION_SET": true,
}

// CreateRole 创建角色（校验 tenant_id + role_code 唯一性）
func (s *RoleService) CreateRole(req *CreateRoleRequest) (*model.Role, error) {
	// 校验 role_type 枚举值
	if !validRoleTypes[req.RoleType] {
		return nil, errors.NewAuthError(errors.ErrInvalidParam, "无效的 role_type，允许值: NORMAL/SUPER_ADMIN/PERMISSION_SET")
	}

	// 检查 tenant_id + role_code 唯一性
	existing, err := s.roleRepo.FindByCode(s.db, req.TenantID, req.RoleCode)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "角色编码在该租户下已存在")
	}

	// PERMISSION_SET 类型强制 parent_id 为 NULL（无层级关系）
	if req.RoleType == "PERMISSION_SET" {
		req.ParentID = nil
	}

	// 校验 parent_id 循环引用和深度（PERMISSION_SET 已无 parent_id，不会进入）
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

// DeleteRole 软删除角色（检查用户绑定 / PERMISSION_SET 允许有绑定但需失效缓存）
func (s *RoleService) DeleteRole(id int64) error {
	// 检查角色是否存在
	role, err := s.roleRepo.FindByID(s.db, id)
	if err != nil {
		return err
	}

	// 保护最后一个 SUPER_ADMIN 角色
	if role.RoleType == "SUPER_ADMIN" {
		var superCount int64
		s.db.Model(&model.Role{}).
			Where("role_type = 'SUPER_ADMIN' AND tenant_id = ? AND id != ?", role.TenantID, id).
			Count(&superCount)
		if superCount == 0 {
			return errors.NewAuthError(errors.ErrProtectedEntity, "无法删除最后一个超级管理员角色")
		}
	}

	// PERMISSION_SET 类型：允许有用户绑定的情况下删除（先收集关联用户用于缓存失效）
	if role.RoleType == "PERMISSION_SET" {
		// 收集关联用户 ID，用于删除后失效缓存
		var affectedUserIDs []int64
		s.db.Model(&model.UserRole{}).
			Where("role_id = ?", id).
			Distinct().
			Pluck("user_id", &affectedUserIDs)

		// 删除用户-角色绑定关系
		s.db.Where("role_id = ?", id).Delete(&model.UserRole{})

		// 软删除角色
		if err := s.roleRepo.SoftDelete(s.db, id); err != nil {
			return err
		}

		// 失效关联用户的权限缓存，确保用户实时失去权限集带来的额外权限
		if len(affectedUserIDs) > 0 && s.permCache != nil {
			for _, uid := range affectedUserIDs {
				s.permCache.InvalidateUserPermissions(
					fmt.Sprintf("%d", uid),
					fmt.Sprintf("%d", role.TenantID),
				)
			}
			// 同时失效 L1 中该角色的缓存
			s.permCache.InvalidateRole(fmt.Sprintf("%d", id))
		}

		s.logger.Log(0, "delete_role", "role", id, "删除权限集，影响用户数: "+fmt.Sprintf("%d", len(affectedUserIDs)))
		return nil
	}

	// 非 PERMISSION_SET：检查是否有用户绑定
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

	// 失效 L1 中该角色的缓存（即使无用户绑定，清理残留缓存）
	if s.permCache != nil {
		s.permCache.InvalidateRole(fmt.Sprintf("%d", id))
	}

	s.logger.Log(0, "delete_role", "role", id, "")
	return nil
}

// AssignResourcesRequest 分配菜单权限请求参数
type AssignResourcesRequest struct {
	ResourceIDs StringInt64Slice `json:"resource_ids"`
}

// AffectedChild 级联裁剪影响的子角色信息
type AffectedChild struct {
	RoleID       int64  `json:"role_id,string"`
	RoleName     string `json:"role_name"`
	RemovedCount int    `json:"removed_count"`
}

// AssignResult AssignResources 返回结果
type AssignResult struct {
	AffectedChildren []AffectedChild `json:"affected_children"`
}

// AssignResources 全量替换角色菜单权限（含父角色子集校验 + 级联裁剪）
func (s *RoleService) AssignResources(roleID int64, req *AssignResourcesRequest) (*AssignResult, error) {
	// 校验角色存在
	role, err := s.roleRepo.FindByID(s.db, roleID)
	if err != nil {
		return nil, err
	}

	// 父角色子集校验：非 PERMISSION_SET 且有父角色时校验
	if role.RoleType != "PERMISSION_SET" && role.ParentID != nil {
		parentResIDs := s.getRoleResourceIDsInt64(s.db, *role.ParentID)
		if !isSubset(req.ResourceIDs, parentResIDs) {
			return nil, errors.NewAuthError(errors.ErrExceedsParentPermission, "分配的资源超出父角色权限范围")
		}
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

	// 事务中执行全量替换 + 级联裁剪
	var affected []AffectedChild
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

		// 级联裁剪子角色
		affected, err = s.cascadeTrimChildren(tx, roleID, req.ResourceIDs)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 操作日志：当前角色
	s.logger.Log(0, "assign_resources", "role", roleID, "分配菜单权限")

	// 操作日志：每个被裁剪的子角色
	for _, child := range affected {
		s.logger.Log(0, "cascade_trim_resources", "role", child.RoleID,
			fmt.Sprintf("级联裁剪: 移除 %d 项资源", child.RemovedCount))
	}

	// 缓存失效：收集所有受影响角色关联的用户，批量清除权限缓存
	if len(affected) > 0 && s.permCache != nil {
		affectedRoleIDs := make([]int64, 0, len(affected))
		for _, child := range affected {
			affectedRoleIDs = append(affectedRoleIDs, child.RoleID)
		}
		s.invalidatePermCacheByRoles(affectedRoleIDs)
	}

	return &AssignResult{AffectedChildren: affected}, nil
}

// cascadeTrimChildren 级联裁剪子角色权限（递归）
// 对每个子角色，计算其 resourceIDs 与新父角色 resourceIDs 的交集
// 如果交集小于原来的绑定数，用交集替换子角色绑定
func (s *RoleService) cascadeTrimChildren(tx *gorm.DB, parentRoleID int64, parentResourceIDs []int64) ([]AffectedChild, error) {
	var affected []AffectedChild

	// 查询直接子角色
	var children []model.Role
	if err := tx.Where("parent_id = ?", parentRoleID).Find(&children).Error; err != nil {
		return nil, err
	}

	parentSet := int64ToSet(parentResourceIDs)

	for _, child := range children {
		// PERMISSION_SET 不受约束
		if child.RoleType == "PERMISSION_SET" {
			continue
		}

		// 查询子角色当前绑定的 resourceIDs
		childResIDs := s.getRoleResourceIDsInt64(tx, child.ID)

		// 计算交集
		intersection := intersect(childResIDs, parentSet)

		// 如果交集小于原绑定数，需要裁剪
		removedCount := len(childResIDs) - len(intersection)
		if removedCount > 0 {
			// 删除旧绑定
			if err := tx.Where("role_id = ?", child.ID).Delete(&model.RoleResource{}).Error; err != nil {
				return nil, err
			}
			// 插入交集部分
			for _, resID := range intersection {
				rr := &model.RoleResource{
					RoleID:     child.ID,
					ResourceID: resID,
				}
				if err := tx.Create(rr).Error; err != nil {
					return nil, err
				}
			}

			affected = append(affected, AffectedChild{
				RoleID:       child.ID,
				RoleName:     child.RoleName,
				RemovedCount: removedCount,
			})

			// 递归处理下一层（用裁剪后的交集作为新的父范围）
			subAffected, err := s.cascadeTrimChildren(tx, child.ID, intersection)
			if err != nil {
				return nil, err
			}
			affected = append(affected, subAffected...)
		} else {
			// 即使当前子角色无需裁剪，也要递归检查更深层级
			// （如果当前子角色的资源本身就是父的子集，其子角色可能还是超出）
			// 这里不需要递归，因为父没变小对于子的子来说也没变
		}
	}

	return affected, nil
}

// invalidatePermCacheByRoles 批量失效受影响角色关联用户的权限缓存
func (s *RoleService) invalidatePermCacheByRoles(roleIDs []int64) {
	if s.permCache == nil {
		return
	}

	// 查询这些角色关联的 userIDs
	var userIDs []int64
	s.db.Model(&model.UserRole{}).
		Where("role_id IN ?", roleIDs).
		Distinct().
		Pluck("user_id", &userIDs)

	if len(userIDs) == 0 {
		return
	}

	// 对每个角色调用缓存失效
	for _, roleID := range roleIDs {
		s.permCache.InvalidateRole(fmt.Sprintf("%d", roleID))
	}
}

// getRoleResourceIDsInt64 查询角色绑定的资源 ID 列表（int64）
func (s *RoleService) getRoleResourceIDsInt64(db *gorm.DB, roleID int64) []int64 {
	var ids []int64
	db.Model(&model.RoleResource{}).Where("role_id = ?", roleID).Pluck("resource_id", &ids)
	return ids
}

// isSubset 判断 items 是否为 superset 的子集
func isSubset(items []int64, superset []int64) bool {
	supersetMap := int64ToSet(superset)
	for _, id := range items {
		if !supersetMap[id] {
			return false
		}
	}
	return true
}

// int64ToSet 将 int64 切片转为 map set
func int64ToSet(ids []int64) map[int64]bool {
	m := make(map[int64]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// intersect 计算 items 与 allowedSet 的交集
func intersect(items []int64, allowedSet map[int64]bool) []int64 {
	result := make([]int64, 0, len(items))
	for _, id := range items {
		if allowedSet[id] {
			result = append(result, id)
		}
	}
	return result
}

// ListRoles 按 tenant_id 查询角色列表，支持按 role_type 过滤
func (s *RoleService) ListRoles(tenantID int64, roleType string) ([]model.Role, error) {
	if roleType != "" {
		return s.roleRepo.ListByTenantIDAndType(s.db, tenantID, roleType)
	}
	return s.roleRepo.ListByTenantID(s.db, tenantID)
}

// AssignApisRequest 角色-接口权限分配请求
type AssignApisRequest struct {
	ApiPermissionIDs StringInt64Slice `json:"api_permission_ids" binding:"-"`
}

// AssignApis 全量替换角色的接口权限（含父角色子集校验 + 级联裁剪）
// 逻辑：
// 1. 查询所有传入 ID 的 ApiPermission 记录
// 2. GROUP 类型自动展开为其所有子 ENDPOINT
// 3. 最终只保留 ENDPOINT 类型的 ID
// 4. 父角色子集校验（非 PERMISSION_SET 且有 parent_id 时）
// 5. 事务中删除旧绑定并批量插入新绑定 + 级联裁剪子角色
func (s *RoleService) AssignApis(roleID int64, req *AssignApisRequest) (*AssignResult, error) {
	// 校验角色存在
	role, err := s.roleRepo.FindByID(s.db, roleID)
	if err != nil {
		return nil, err
	}

	// 收集最终的 ENDPOINT ID 列表
	endpointIDs := make([]int64, 0)

	if len(req.ApiPermissionIDs) > 0 {
		// 查询所有传入 ID 的 ApiPermission 记录
		var perms []model.ApiPermission
		if err := s.db.Where("id IN ?", req.ApiPermissionIDs).Find(&perms).Error; err != nil {
			return nil, err
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
				return nil, err
			}
			for _, ep := range childEndpoints {
				endpointIDs = append(endpointIDs, ep.ID)
			}
		}

		// 去重
		endpointIDs = uniqueInt64(endpointIDs)
	}

	// 父角色子集校验：非 PERMISSION_SET 且有父角色时校验
	if role.RoleType != "PERMISSION_SET" && role.ParentID != nil {
		parentApiIDs := s.getRoleApiIDsInt64(s.db, *role.ParentID)
		if !isSubset(endpointIDs, parentApiIDs) {
			return nil, errors.NewAuthError(errors.ErrExceedsParentPermission, "分配的接口权限超出父角色权限范围")
		}
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

	// 事务：删除旧绑定 + 批量插入新绑定 + 级联裁剪
	var affected []AffectedChild
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

		// 级联裁剪子角色 API 绑定
		affected, err = s.cascadeTrimChildrenApis(tx, roleID, endpointIDs)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 操作日志：当前角色
	s.logger.Log(0, "assign_apis", "role", roleID, fmt.Sprintf("分配接口权限，ENDPOINT 数量: %d", len(endpointIDs)))

	// 操作日志：每个被裁剪的子角色
	for _, child := range affected {
		s.logger.Log(0, "cascade_trim_apis", "role", child.RoleID,
			fmt.Sprintf("级联裁剪: 移除 %d 项接口权限", child.RemovedCount))
	}

	// 缓存失效：收集所有受影响角色关联的用户，批量清除权限缓存
	if len(affected) > 0 && s.permCache != nil {
		affectedRoleIDs := make([]int64, 0, len(affected))
		for _, child := range affected {
			affectedRoleIDs = append(affectedRoleIDs, child.RoleID)
		}
		s.invalidatePermCacheByRoles(affectedRoleIDs)
	}

	return &AssignResult{AffectedChildren: affected}, nil
}

// cascadeTrimChildrenApis 级联裁剪子角色 API 权限（递归）
// 对每个子角色，计算其 apiIDs 与新父角色 endpointIDs 的交集
// 如果交集小于原来的绑定数，用交集替换子角色绑定
func (s *RoleService) cascadeTrimChildrenApis(tx *gorm.DB, parentRoleID int64, parentApiIDs []int64) ([]AffectedChild, error) {
	var affected []AffectedChild

	// 查询直接子角色
	var children []model.Role
	if err := tx.Where("parent_id = ?", parentRoleID).Find(&children).Error; err != nil {
		return nil, err
	}

	parentSet := int64ToSet(parentApiIDs)

	for _, child := range children {
		// PERMISSION_SET 不受约束
		if child.RoleType == "PERMISSION_SET" {
			continue
		}

		// 查询子角色当前绑定的 API 权限 ID
		childApiIDs := s.getRoleApiIDsInt64(tx, child.ID)

		// 计算交集
		intersection := intersect(childApiIDs, parentSet)

		// 如果交集小于原绑定数，需要裁剪
		removedCount := len(childApiIDs) - len(intersection)
		if removedCount > 0 {
			// 删除旧绑定
			if err := tx.Where("role_id = ?", child.ID).Delete(&model.RoleApi{}).Error; err != nil {
				return nil, err
			}
			// 插入交集部分
			for _, apiID := range intersection {
				ra := &model.RoleApi{
					RoleID:          child.ID,
					ApiPermissionID: apiID,
				}
				if err := tx.Create(ra).Error; err != nil {
					return nil, err
				}
			}

			affected = append(affected, AffectedChild{
				RoleID:       child.ID,
				RoleName:     child.RoleName,
				RemovedCount: removedCount,
			})

			// 递归处理下一层（用裁剪后的交集作为新的父范围）
			subAffected, err := s.cascadeTrimChildrenApis(tx, child.ID, intersection)
			if err != nil {
				return nil, err
			}
			affected = append(affected, subAffected...)
		}
	}

	return affected, nil
}

// getRoleApiIDsInt64 查询角色绑定的 API 权限 ID 列表（int64）
func (s *RoleService) getRoleApiIDsInt64(db *gorm.DB, roleID int64) []int64 {
	var ids []int64
	db.Model(&model.RoleApi{}).Where("role_id = ?", roleID).Pluck("api_permission_id", &ids)
	return ids
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

// GetAssignableResources 获取角色可分配给子角色的资源 ID 范围
// SUPER_ADMIN 角色：返回系统内全部资源 ID（不受租户订阅限制）
// 顶级角色（parent_id=NULL）：返回租户订阅应用范围内全部资源 ID
// 非顶级角色：返回该角色自身已绑定的资源 ID（子角色可选范围 = 父角色已有权限）
func (s *RoleService) GetAssignableResources(roleID int64, tenantID int64) ([]string, error) {
	role, err := s.roleRepo.FindByID(s.db, roleID)
	if err != nil {
		return nil, err
	}

	var ids []int64
	if role.RoleType == "SUPER_ADMIN" {
		// SUPER_ADMIN：返回全部资源
		if err := s.db.Model(&model.Resource{}).Pluck("id", &ids).Error; err != nil {
			return nil, err
		}
	} else if role.ParentID == nil || *role.ParentID == 0 {
		// 顶级角色：查询租户订阅应用范围内全部资源 ID
		var appCodes []string
		if err := s.db.Model(&model.TenantApp{}).Where("tenant_id = ?", tenantID).Pluck("app_code", &appCodes).Error; err != nil {
			return nil, err
		}
		if len(appCodes) > 0 {
			if err := s.db.Model(&model.Resource{}).Where("app_code IN ?", appCodes).Pluck("id", &ids).Error; err != nil {
				return nil, err
			}
		}
	} else {
		// 非顶级角色：返回自身已绑定的资源 ID
		ids = s.getRoleResourceIDsInt64(s.db, roleID)
	}

	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = fmt.Sprintf("%d", id)
	}
	return result, nil
}

// GetAssignableApis 获取角色可分配给子角色的 API 权限 ID 范围
// SUPER_ADMIN 角色：返回系统内全部 API 权限 ID（不受租户订阅限制）
// 顶级角色（parent_id=NULL）：返回租户订阅应用范围内全部 API 权限 ID
// 非顶级角色：返回该角色自身已绑定的 API 权限 ID（子角色可选范围 = 父角色已有权限）
func (s *RoleService) GetAssignableApis(roleID int64, tenantID int64) ([]string, error) {
	role, err := s.roleRepo.FindByID(s.db, roleID)
	if err != nil {
		return nil, err
	}

	var ids []int64
	if role.RoleType == "SUPER_ADMIN" {
		// SUPER_ADMIN：返回全部 API 权限
		if err := s.db.Model(&model.ApiPermission{}).Pluck("id", &ids).Error; err != nil {
			return nil, err
		}
	} else if role.ParentID == nil || *role.ParentID == 0 {
		// 顶级角色：查询租户订阅应用范围内全部 API 权限 ID
		var appCodes []string
		if err := s.db.Model(&model.TenantApp{}).Where("tenant_id = ?", tenantID).Pluck("app_code", &appCodes).Error; err != nil {
			return nil, err
		}
		if len(appCodes) > 0 {
			if err := s.db.Model(&model.ApiPermission{}).Where("app_code IN ?", appCodes).Pluck("id", &ids).Error; err != nil {
				return nil, err
			}
		}
	} else {
		// 非顶级角色：返回自身已绑定的 API 权限 ID
		ids = s.getRoleApiIDsInt64(s.db, roleID)
	}

	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = fmt.Sprintf("%d", id)
	}
	return result, nil
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
