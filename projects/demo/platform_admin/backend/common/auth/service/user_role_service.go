package service

import (
	"time"

	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/event"

	"gorm.io/gorm"
)

// UserRoleService 用户-角色分配业务逻辑层
type UserRoleService struct {
	db     *gorm.DB
	cfg    *config.Config
	logger OperationLogger
}

// NewUserRoleService 创建 UserRoleService 实例
func NewUserRoleService(db *gorm.DB, cfg *config.Config) *UserRoleService {
	return &UserRoleService{
		db:     db,
		cfg:    cfg,
		logger: &noopLogger{},
	}
}

// SetLogger 设置操作日志记录器
func (s *UserRoleService) SetLogger(logger OperationLogger) {
	s.logger = logger
}

// AssignRolesRequest 分配角色请求参数
type AssignRolesRequest struct {
	TenantID int64            `json:"tenant_id,string" binding:"required"`
	Roles    []RoleAssignment `json:"roles" binding:"required"`
}

// RoleAssignment 单个角色分配
type RoleAssignment struct {
	RoleID         int64      `json:"role_id,string" binding:"required"`
	EffectiveStart *time.Time `json:"effective_start"`
	EffectiveEnd   *time.Time `json:"effective_end"`
}

// AssignRoles 为用户分配角色（追加模式）
func (s *UserRoleService) AssignRoles(userID int64, req *AssignRolesRequest) error {
	// 校验用户已关联当前租户
	if err := s.validateUserInTenant(userID, req.TenantID); err != nil {
		return err
	}

	// 校验所有角色属于当前租户
	for _, ra := range req.Roles {
		if err := s.validateRoleBelongsToTenant(ra.RoleID, req.TenantID); err != nil {
			return err
		}
	}

	// 创建用户-角色记录
	for _, ra := range req.Roles {
		ur := &model.UserRole{
			UserID:         userID,
			RoleID:         ra.RoleID,
			TenantID:       req.TenantID,
			EffectiveStart: ra.EffectiveStart,
			EffectiveEnd:   ra.EffectiveEnd,
		}
		if err := s.db.Create(ur).Error; err != nil {
			return err
		}
	}

	s.logger.Log(0, "assign_roles", "user_role", userID, "为用户分配角色")

	// CR-8: 发布权限变更事件（所有 Create 成功即为提交完成）
	event.DefaultBus.Publish(event.EventPermissionChanged, &event.PermissionChangedEvent{
		AffectedUsers: []event.AffectedUser{{UserID: userID, TenantID: req.TenantID}},
		Source:        "assign_roles",
	})

	return nil
}

// ReplaceRoles 全量替换用户角色（删除当前租户下旧记录，创建新记录）
func (s *UserRoleService) ReplaceRoles(userID int64, req *AssignRolesRequest) error {
	// 校验用户已关联当前租户
	if err := s.validateUserInTenant(userID, req.TenantID); err != nil {
		return err
	}

	// 校验所有角色属于当前租户
	for _, ra := range req.Roles {
		if err := s.validateRoleBelongsToTenant(ra.RoleID, req.TenantID); err != nil {
			return err
		}
	}

	// SUPER_ADMIN 保护：如果用户当前是 SUPER_ADMIN，新角色列表必须包含 SUPER_ADMIN
	if s.isSuperAdminUser(userID, req.TenantID) {
		newRoleIDs := make([]int64, 0, len(req.Roles))
		for _, ra := range req.Roles {
			newRoleIDs = append(newRoleIDs, ra.RoleID)
		}
		if len(newRoleIDs) > 0 {
			var superCount int64
			s.db.Model(&model.Role{}).
				Where("id IN ? AND role_type = 'SUPER_ADMIN'", newRoleIDs).
				Count(&superCount)
			if superCount == 0 {
				return errors.NewAuthError(errors.ErrProtectedEntity, "超级管理员不能移除自身的超级管理员角色")
			}
		}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 删除当前租户下所有旧角色
		if err := tx.Where("user_id = ? AND tenant_id = ?", userID, req.TenantID).
			Delete(&model.UserRole{}).Error; err != nil {
			return err
		}

		// 插入新的角色列表
		for _, ra := range req.Roles {
			ur := &model.UserRole{
				UserID:         userID,
				RoleID:         ra.RoleID,
				TenantID:       req.TenantID,
				EffectiveStart: ra.EffectiveStart,
				EffectiveEnd:   ra.EffectiveEnd,
			}
			if err := tx.Create(ur).Error; err != nil {
				return err
			}
		}

		s.logger.Log(0, "replace_roles", "user_role", userID, "全量替换用户角色")
		return nil
	})
	if err != nil {
		return err
	}

	// CR-8: 事务 Commit 成功后发布权限变更事件
	event.DefaultBus.Publish(event.EventPermissionChanged, &event.PermissionChangedEvent{
		AffectedUsers: []event.AffectedUser{{UserID: userID, TenantID: req.TenantID}},
		Source:        "replace_roles",
	})

	return nil
}

// validateUserInTenant 校验用户已关联当前租户
func (s *UserRoleService) validateUserInTenant(userID, tenantID int64) error {
	var count int64
	s.db.Model(&model.UserTenant{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Count(&count)
	if count == 0 {
		return errors.NewAuthError(errors.ErrUserNotInTenant, "用户未关联当前租户")
	}
	return nil
}

// validateRoleBelongsToTenant 校验角色属于当前租户
func (s *UserRoleService) validateRoleBelongsToTenant(roleID, tenantID int64) error {
	var count int64
	s.db.Model(&model.Role{}).
		Where("id = ? AND tenant_id = ?", roleID, tenantID).
		Count(&count)
	if count == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "角色不存在或不属于当前租户")
	}
	return nil
}

// isSuperAdminUser 检查用户是否在指定租户下拥有 SUPER_ADMIN 角色
func (s *UserRoleService) isSuperAdminUser(userID, tenantID int64) bool {
	var count int64
	s.db.Table("admin_user_role ur").
		Joins("JOIN admin_role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND ur.tenant_id = ? AND r.role_type = 'SUPER_ADMIN'", userID, tenantID).
		Count(&count)
	return count > 0
}
