package service

import (
	"time"

	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

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

	return s.db.Transaction(func(tx *gorm.DB) error {
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
