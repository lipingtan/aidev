package repository

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// TenantRepository 租户数据访问接口
type TenantRepository interface {
	Create(db *gorm.DB, tenant *model.Tenant) error
	Update(db *gorm.DB, tenant *model.Tenant) error
	FindByID(db *gorm.DB, id int64) (*model.Tenant, error)
	FindByCode(db *gorm.DB, code string) (*model.Tenant, error)
	List(db *gorm.DB, params TenantListParams) ([]model.Tenant, int64, error)
	SoftDelete(db *gorm.DB, id int64) error
	UpdateStatus(db *gorm.DB, id int64, status int) error
}

// TenantListParams 租户列表查询参数
type TenantListParams struct {
	Page     int
	PageSize int
	Status   *int   // 可选状态筛选
	Name     string // 模糊搜索
}

// tenantRepo TenantRepository 默认实现
type tenantRepo struct{}

// NewTenantRepository 创建 TenantRepository 实例
func NewTenantRepository() TenantRepository {
	return &tenantRepo{}
}

// Create 创建租户
func (r *tenantRepo) Create(db *gorm.DB, tenant *model.Tenant) error {
	return db.Create(tenant).Error
}

// Update 更新租户（乐观锁：WHERE version = oldVersion）
func (r *tenantRepo) Update(db *gorm.DB, tenant *model.Tenant) error {
	oldVersion := tenant.Version
	tenant.Version = oldVersion + 1
	result := db.Model(tenant).
		Where("id = ? AND version = ?", tenant.ID, oldVersion).
		Updates(map[string]interface{}{
			"name":    tenant.Name,
			"config":  tenant.Config,
			"status":  tenant.Status,
			"version": tenant.Version,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}
	return nil
}

// FindByID 根据 ID 查找租户
func (r *tenantRepo) FindByID(db *gorm.DB, id int64) (*model.Tenant, error) {
	var tenant model.Tenant
	err := db.First(&tenant, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrEntityNotFound, "租户不存在")
		}
		return nil, err
	}
	return &tenant, nil
}

// FindByCode 根据 tenant_code 查找租户
func (r *tenantRepo) FindByCode(db *gorm.DB, code string) (*model.Tenant, error) {
	var tenant model.Tenant
	err := db.Where("tenant_code = ?", code).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 不存在返回 nil，不报错
		}
		return nil, err
	}
	return &tenant, nil
}

// List 分页查询租户列表
func (r *tenantRepo) List(db *gorm.DB, params TenantListParams) ([]model.Tenant, int64, error) {
	var tenants []model.Tenant
	var total int64

	query := db.Model(&model.Tenant{})

	// 状态筛选
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	// 名称模糊搜索
	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&tenants).Error; err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

// SoftDelete 硬删除租户（软删除会导致 tenant_code/name unique index 残留）
func (r *tenantRepo) SoftDelete(db *gorm.DB, id int64) error {
	result := db.Unscoped().Delete(&model.Tenant{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "租户不存在")
	}
	return nil
}

// UpdateStatus 更新租户状态
func (r *tenantRepo) UpdateStatus(db *gorm.DB, id int64, status int) error {
	result := db.Model(&model.Tenant{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "租户不存在")
	}
	return nil
}
