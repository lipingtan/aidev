package repository

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// OrgUnitRepository 组织架构节点数据访问接口
type OrgUnitRepository interface {
	Create(db *gorm.DB, unit *model.OrgUnit) error
	Update(db *gorm.DB, unit *model.OrgUnit) error
	Delete(db *gorm.DB, id int64) error
	FindByID(db *gorm.DB, id int64) (*model.OrgUnit, error)
	ListByTenant(db *gorm.DB, tenantID int64) ([]model.OrgUnit, error)
	HasChildren(db *gorm.DB, id int64) (bool, error)
	FindByTenantAndCode(db *gorm.DB, tenantID int64, code string) (*model.OrgUnit, error)
}

// orgUnitRepo OrgUnitRepository 默认实现
type orgUnitRepo struct{}

// NewOrgUnitRepository 创建 OrgUnitRepository 实例
func NewOrgUnitRepository() OrgUnitRepository {
	return &orgUnitRepo{}
}

func (r *orgUnitRepo) Create(db *gorm.DB, unit *model.OrgUnit) error {
	return db.Create(unit).Error
}

func (r *orgUnitRepo) Update(db *gorm.DB, unit *model.OrgUnit) error {
	result := db.Model(unit).Where("id = ? AND version = ?", unit.ID, unit.Version-1).Updates(unit)
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}
	return result.Error
}

func (r *orgUnitRepo) Delete(db *gorm.DB, id int64) error {
	// 硬删除：软删除会导致 unique index 残留（tenant_id+code），相同编码无法复用
	return db.Unscoped().Delete(&model.OrgUnit{}, id).Error
}

func (r *orgUnitRepo) FindByID(db *gorm.DB, id int64) (*model.OrgUnit, error) {
	var unit model.OrgUnit
	if err := db.First(&unit, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrEntityNotFound, "组织节点不存在")
		}
		return nil, err
	}
	return &unit, nil
}

func (r *orgUnitRepo) ListByTenant(db *gorm.DB, tenantID int64) ([]model.OrgUnit, error) {
	var units []model.OrgUnit
	err := db.Where("tenant_id = ? AND status = 1", tenantID).
		Order("sort_order ASC, created_at ASC").Find(&units).Error
	return units, err
}

func (r *orgUnitRepo) HasChildren(db *gorm.DB, id int64) (bool, error) {
	var count int64
	err := db.Model(&model.OrgUnit{}).Where("parent_id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *orgUnitRepo) FindByTenantAndCode(db *gorm.DB, tenantID int64, code string) (*model.OrgUnit, error) {
	var unit model.OrgUnit
	err := db.Where("tenant_id = ? AND code = ?", tenantID, code).First(&unit).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &unit, nil
}
