package repository

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(db *gorm.DB, user *model.User) error
	Update(db *gorm.DB, user *model.User) error
	FindByID(db *gorm.DB, id int64) (*model.User, error)
	FindByUsername(db *gorm.DB, username string) (*model.User, error)
	ListByTenantID(db *gorm.DB, params UserListParams) ([]model.User, int64, error)
	SoftDelete(db *gorm.DB, id int64) error

	CreateUserTenant(db *gorm.DB, ut *model.UserTenant) error
	DeleteUserTenant(db *gorm.DB, userID, tenantID int64) error
	ListUserTenants(db *gorm.DB, userID int64) ([]model.UserTenant, error)
	FindUserTenant(db *gorm.DB, userID, tenantID int64) (*model.UserTenant, error)

	DeleteUserRolesByTenant(db *gorm.DB, userID, tenantID int64) error
}

// UserListParams 用户列表查询参数
type UserListParams struct {
	TenantID int64
	Page     int
	PageSize int
}

// userRepo UserRepository 默认实现
type userRepo struct{}

// NewUserRepository 创建 UserRepository 实例
func NewUserRepository() UserRepository {
	return &userRepo{}
}

// Create 创建用户
func (r *userRepo) Create(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}

// Update 更新用户（乐观锁：WHERE version = oldVersion）
func (r *userRepo) Update(db *gorm.DB, user *model.User) error {
	oldVersion := user.Version
	user.Version = oldVersion + 1
	result := db.Model(user).
		Where("id = ? AND version = ?", user.ID, oldVersion).
		Updates(map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
			"phone":    user.Phone,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"status":   user.Status,
			"version":  user.Version,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}
	return nil
}

// FindByID 根据 ID 查找用户
func (r *userRepo) FindByID(db *gorm.DB, id int64) (*model.User, error) {
	var user model.User
	err := db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrEntityNotFound, "用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *userRepo) FindByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 不存在返回 nil，不报错
		}
		return nil, err
	}
	return &user, nil
}

// ListByTenantID 通过 tenant_id 过滤查询用户列表（JOIN admin_user_tenant）
func (r *userRepo) ListByTenantID(db *gorm.DB, params UserListParams) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := db.Model(&model.User{}).
		Joins("INNER JOIN admin_user_tenant ON admin_user_tenant.user_id = admin_user.id").
		Where("admin_user_tenant.tenant_id = ?", params.TenantID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Select("admin_user.*").Order("admin_user.created_at DESC").Offset(offset).Limit(params.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// SoftDelete 软删除用户
func (r *userRepo) SoftDelete(db *gorm.DB, id int64) error {
	result := db.Delete(&model.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "用户不存在")
	}
	return nil
}

// CreateUserTenant 创建用户-租户关联
func (r *userRepo) CreateUserTenant(db *gorm.DB, ut *model.UserTenant) error {
	return db.Create(ut).Error
}

// DeleteUserTenant 删除用户-租户关联
func (r *userRepo) DeleteUserTenant(db *gorm.DB, userID, tenantID int64) error {
	result := db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Delete(&model.UserTenant{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "用户-租户关联不存在")
	}
	return nil
}

// ListUserTenants 查询用户已关联的租户列表
func (r *userRepo) ListUserTenants(db *gorm.DB, userID int64) ([]model.UserTenant, error) {
	var list []model.UserTenant
	err := db.Where("user_id = ?", userID).Find(&list).Error
	return list, err
}

// FindUserTenant 查找用户-租户关联
func (r *userRepo) FindUserTenant(db *gorm.DB, userID, tenantID int64) (*model.UserTenant, error) {
	var ut model.UserTenant
	err := db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&ut).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &ut, nil
}

// DeleteUserRolesByTenant 级联删除用户在指定租户下的角色绑定
func (r *userRepo) DeleteUserRolesByTenant(db *gorm.DB, userID, tenantID int64) error {
	return db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Delete(&model.UserRole{}).Error
}
