package repository

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// DataScopeConfigRepository 数据权限维度配置数据访问接口
type DataScopeConfigRepository interface {
	Create(db *gorm.DB, cfg *model.DataScopeConfig) error
	Update(db *gorm.DB, cfg *model.DataScopeConfig) error
	FindByID(db *gorm.DB, id int64) (*model.DataScopeConfig, error)
	FindByDimensionName(db *gorm.DB, name string) (*model.DataScopeConfig, error)
	List(db *gorm.DB) ([]model.DataScopeConfig, error)
	Delete(db *gorm.DB, id int64) error
}

// DataScopeRepository 角色数据权限绑定数据访问接口
type DataScopeRepository interface {
	ReplaceByRole(db *gorm.DB, roleID int64, scopes []model.DataScope) error
	FindByRole(db *gorm.DB, roleID int64) ([]model.DataScope, error)
}

// dataScopeConfigRepo DataScopeConfigRepository 默认实现
type dataScopeConfigRepo struct{}

// NewDataScopeConfigRepository 创建实例
func NewDataScopeConfigRepository() DataScopeConfigRepository {
	return &dataScopeConfigRepo{}
}

// Create 创建维度配置
func (r *dataScopeConfigRepo) Create(db *gorm.DB, cfg *model.DataScopeConfig) error {
	return db.Create(cfg).Error
}

// Update 更新维度配置
func (r *dataScopeConfigRepo) Update(db *gorm.DB, cfg *model.DataScopeConfig) error {
	result := db.Model(cfg).Where("id = ?", cfg.ID).Updates(map[string]interface{}{
		"dimension_name": cfg.DimensionName,
		"display_name":   cfg.DisplayName,
		"table_column":   cfg.TableColumn,
		"value_source":   cfg.ValueSource,
		"handler_name":   cfg.HandlerName,
		"status":         cfg.Status,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "维度配置不存在")
	}
	return nil
}

// FindByID 根据 ID 查找
func (r *dataScopeConfigRepo) FindByID(db *gorm.DB, id int64) (*model.DataScopeConfig, error) {
	var cfg model.DataScopeConfig
	err := db.First(&cfg, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrEntityNotFound, "维度配置不存在")
		}
		return nil, err
	}
	return &cfg, nil
}

// FindByDimensionName 根据维度标识查找
func (r *dataScopeConfigRepo) FindByDimensionName(db *gorm.DB, name string) (*model.DataScopeConfig, error) {
	var cfg model.DataScopeConfig
	err := db.Where("dimension_name = ?", name).First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// List 查询所有维度配置
func (r *dataScopeConfigRepo) List(db *gorm.DB) ([]model.DataScopeConfig, error) {
	var list []model.DataScopeConfig
	err := db.Order("created_at DESC").Find(&list).Error
	return list, err
}

// Delete 删除维度配置
func (r *dataScopeConfigRepo) Delete(db *gorm.DB, id int64) error {
	result := db.Delete(&model.DataScopeConfig{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "维度配置不存在")
	}
	return nil
}

// dataScopeRepo DataScopeRepository 默认实现
type dataScopeRepo struct{}

// NewDataScopeRepository 创建实例
func NewDataScopeRepository() DataScopeRepository {
	return &dataScopeRepo{}
}

// ReplaceByRole 替换角色的数据权限绑定（先删后插）
func (r *dataScopeRepo) ReplaceByRole(db *gorm.DB, roleID int64, scopes []model.DataScope) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 先删除该角色的所有数据权限绑定
		if err := tx.Where("role_id = ?", roleID).Delete(&model.DataScope{}).Error; err != nil {
			return err
		}
		// 批量插入新绑定
		for i := range scopes {
			scopes[i].RoleID = roleID
			if err := tx.Create(&scopes[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByRole 查询角色的数据权限绑定
func (r *dataScopeRepo) FindByRole(db *gorm.DB, roleID int64) ([]model.DataScope, error) {
	var list []model.DataScope
	err := db.Where("role_id = ?", roleID).Find(&list).Error
	return list, err
}
