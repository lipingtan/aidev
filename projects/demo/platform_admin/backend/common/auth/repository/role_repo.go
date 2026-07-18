package repository

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// RoleRepository 角色数据访问接口
type RoleRepository interface {
	Create(db *gorm.DB, role *model.Role) error
	Update(db *gorm.DB, role *model.Role) error
	FindByID(db *gorm.DB, id int64) (*model.Role, error)
	FindByCode(db *gorm.DB, tenantID int64, code string) (*model.Role, error)
	ListByTenantID(db *gorm.DB, tenantID int64) ([]model.Role, error)
	SoftDelete(db *gorm.DB, id int64) error
	HasUserBinding(db *gorm.DB, roleID int64) (bool, error)
}

// roleRepo RoleRepository 默认实现
type roleRepo struct{}

// NewRoleRepository 创建 RoleRepository 实例
func NewRoleRepository() RoleRepository {
	return &roleRepo{}
}

// Create 创建角色
func (r *roleRepo) Create(db *gorm.DB, role *model.Role) error {
	return db.Create(role).Error
}

// Update 更新角色（乐观锁：WHERE version = oldVersion）
func (r *roleRepo) Update(db *gorm.DB, role *model.Role) error {
	oldVersion := role.Version
	role.Version = oldVersion + 1
	result := db.Model(role).
		Where("id = ? AND version = ?", role.ID, oldVersion).
		Updates(map[string]interface{}{
			"role_name":  role.RoleName,
			"role_type":  role.RoleType,
			"parent_id":  role.ParentID,
			"sort_order": role.SortOrder,
			"status":     role.Status,
			"version":    role.Version,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}
	return nil
}

// FindByID 根据 ID 查找角色
func (r *roleRepo) FindByID(db *gorm.DB, id int64) (*model.Role, error) {
	var role model.Role
	err := db.First(&role, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrEntityNotFound, "角色不存在")
		}
		return nil, err
	}
	return &role, nil
}

// FindByCode 根据 tenant_id + role_code 查找角色
func (r *roleRepo) FindByCode(db *gorm.DB, tenantID int64, code string) (*model.Role, error) {
	var role model.Role
	err := db.Where("tenant_id = ? AND role_code = ?", tenantID, code).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// ListByTenantID 按 tenant_id 查询角色列表
func (r *roleRepo) ListByTenantID(db *gorm.DB, tenantID int64) ([]model.Role, error) {
	var roles []model.Role
	err := db.Where("tenant_id = ?", tenantID).Order("sort_order ASC, created_at ASC").Find(&roles).Error
	return roles, err
}

// SoftDelete 软删除角色
func (r *roleRepo) SoftDelete(db *gorm.DB, id int64) error {
	result := db.Delete(&model.Role{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "角色不存在")
	}
	return nil
}

// HasUserBinding 检查角色是否有用户绑定（admin_user_role 表）
func (r *roleRepo) HasUserBinding(db *gorm.DB, roleID int64) (bool, error) {
	var count int64
	err := db.Model(&model.UserRole{}).Where("role_id = ?", roleID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
