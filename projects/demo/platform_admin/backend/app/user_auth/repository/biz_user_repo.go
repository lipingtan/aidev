package repository

import (
	"go-admin/app/user_auth/model"

	"gorm.io/gorm"
)

// BizUserRepository C端用户数据访问接口
type BizUserRepository interface {
	Create(db *gorm.DB, user *model.BizUser) error
	FindByID(db *gorm.DB, id int64) (*model.BizUser, error)
	FindByTenantPhone(db *gorm.DB, tenantID int64, phone string) (*model.BizUser, error)
	Update(db *gorm.DB, user *model.BizUser) error
	Delete(db *gorm.DB, id int64) error
	ListByTenant(db *gorm.DB, tenantID int64, page, pageSize int, phone string) ([]model.BizUser, int64, error)
	IncrementTokenVersion(db *gorm.DB, id int64) error
}

// bizUserRepo BizUserRepository 默认实现
type bizUserRepo struct{}

// NewBizUserRepository 创建 BizUserRepository 实例
func NewBizUserRepository() BizUserRepository {
	return &bizUserRepo{}
}

// Create 创建C端用户
func (r *bizUserRepo) Create(db *gorm.DB, user *model.BizUser) error {
	return db.Create(user).Error
}

// FindByID 根据 ID 查找C端用户
func (r *bizUserRepo) FindByID(db *gorm.DB, id int64) (*model.BizUser, error) {
	var user model.BizUser
	err := db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByTenantPhone 根据租户ID和手机号查找C端用户
func (r *bizUserRepo) FindByTenantPhone(db *gorm.DB, tenantID int64, phone string) (*model.BizUser, error) {
	var user model.BizUser
	err := db.Where("tenant_id = ? AND phone = ? AND deleted_at IS NULL", tenantID, phone).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新C端用户（乐观锁）
func (r *bizUserRepo) Update(db *gorm.DB, user *model.BizUser) error {
	oldVersion := user.Version
	user.Version = oldVersion + 1
	result := db.Model(user).
		Where("id = ? AND version = ? AND deleted_at IS NULL", user.ID, oldVersion).
		Updates(map[string]interface{}{
			"phone":         user.Phone,
			"password":      user.Password,
			"nickname":      user.Nickname,
			"avatar":        user.Avatar,
			"status":        user.Status,
			"last_login_at": user.LastLoginAt,
			"last_login_ip": user.LastLoginIP,
			"update_by":     user.UpdateBy,
			"version":       user.Version,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 软删除C端用户
func (r *bizUserRepo) Delete(db *gorm.DB, id int64) error {
	now := db.NowFunc()
	result := db.Model(&model.BizUser{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListByTenant 按租户分页查询C端用户，支持手机号模糊搜索
func (r *bizUserRepo) ListByTenant(db *gorm.DB, tenantID int64, page, pageSize int, phone string) ([]model.BizUser, int64, error) {
	var users []model.BizUser
	var total int64

	query := db.Model(&model.BizUser{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)
	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// IncrementTokenVersion 递增 token 版本号（用于强制下线）
func (r *bizUserRepo) IncrementTokenVersion(db *gorm.DB, id int64) error {
	result := db.Model(&model.BizUser{}).
		Where("id = ? AND deleted_at IS NULL", id).
		UpdateColumn("token_version", gorm.Expr("token_version + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
