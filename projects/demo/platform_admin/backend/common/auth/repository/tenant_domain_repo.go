package repository

import (
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// TenantDomainRepository 域名-租户映射数据访问接口
// 遵循单表原则：不做跨表 JOIN，tenant_code/name 由 Service 层通过 TenantRepository 补充
type TenantDomainRepository interface {
	// FindByDomain 按域名查找活跃的精确匹配记录（deleted_at IS NULL AND match_type = 'EXACT'）
	FindByDomain(db *gorm.DB, domain string) (*model.TenantDomain, error)
	// FindByID 按 ID 查找记录
	FindByID(db *gorm.DB, id int64) (*model.TenantDomain, error)
	// FindActiveByDomain 检查活跃记录是否已存在（用于唯一性校验）
	FindActiveByDomain(db *gorm.DB, domain string) (*model.TenantDomain, error)
	// Create 创建记录
	Create(db *gorm.DB, td *model.TenantDomain) error
	// Update 更新记录（乐观锁）
	Update(db *gorm.DB, td *model.TenantDomain, oldVersion int) error
	// SoftDelete 软删除
	SoftDelete(db *gorm.DB, id int64) error
	// List 分页查询
	List(db *gorm.DB, params TenantDomainListParams) ([]model.TenantDomain, int64, error)
}

// TenantDomainListParams 域名列表查询参数
type TenantDomainListParams struct {
	Page     int
	PageSize int
	Domain   string // 模糊搜索
	TenantID int64  // 精确筛选（0 表示不筛选）
}

// tenantDomainRepo TenantDomainRepository GORM 实现
type tenantDomainRepo struct{}

// NewTenantDomainRepository 创建 TenantDomainRepository 实例
func NewTenantDomainRepository() TenantDomainRepository {
	return &tenantDomainRepo{}
}

// FindByDomain 按域名查找活跃精确匹配记录
func (r *tenantDomainRepo) FindByDomain(db *gorm.DB, domain string) (*model.TenantDomain, error) {
	var td model.TenantDomain
	err := db.Where("domain = ? AND match_type = 'EXACT'", domain).First(&td).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &td, nil
}

// FindByID 按 ID 查找（包含软删除过滤，GORM 默认处理）
func (r *tenantDomainRepo) FindByID(db *gorm.DB, id int64) (*model.TenantDomain, error) {
	var td model.TenantDomain
	err := db.First(&td, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &td, nil
}

// FindActiveByDomain 查找活跃记录（用于创建/更新前唯一性校验）
func (r *tenantDomainRepo) FindActiveByDomain(db *gorm.DB, domain string) (*model.TenantDomain, error) {
	var td model.TenantDomain
	err := db.Where("domain = ?", domain).First(&td).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &td, nil
}

// Create 创建域名映射记录
func (r *tenantDomainRepo) Create(db *gorm.DB, td *model.TenantDomain) error {
	return db.Create(td).Error
}

// Update 更新记录（乐观锁：WHERE id = ? AND version = oldVersion）
func (r *tenantDomainRepo) Update(db *gorm.DB, td *model.TenantDomain, oldVersion int) error {
	result := db.Model(&model.TenantDomain{}).
		Where("id = ? AND version = ?", td.ID, oldVersion).
		Updates(map[string]interface{}{
			"domain":     td.Domain,
			"tenant_id":  td.TenantID,
			"remark":     td.Remark,
			"version":    td.Version,
			"updated_at": td.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 乐观锁冲突，由 Service 层转换错误
	}
	return nil
}

// SoftDelete 软删除
func (r *tenantDomainRepo) SoftDelete(db *gorm.DB, id int64) error {
	result := db.Delete(&model.TenantDomain{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 分页查询
func (r *tenantDomainRepo) List(db *gorm.DB, params TenantDomainListParams) ([]model.TenantDomain, int64, error) {
	var list []model.TenantDomain
	var total int64

	query := db.Model(&model.TenantDomain{})

	if params.Domain != "" {
		query = query.Where("domain LIKE ?", "%"+params.Domain+"%")
	}
	if params.TenantID > 0 {
		query = query.Where("tenant_id = ?", params.TenantID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
