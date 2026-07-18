package repository

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// ResourceRepository 资源数据访问接口
type ResourceRepository interface {
	Create(db *gorm.DB, resource *model.Resource) error
	Update(db *gorm.DB, resource *model.Resource) error
	FindByID(db *gorm.DB, id int64) (*model.Resource, error)
	ListByAppCode(db *gorm.DB, appCode string) ([]model.Resource, error)
	SoftDelete(db *gorm.DB, id int64) error
	HasChildren(db *gorm.DB, id int64) (bool, error)
	BatchUpdateSort(db *gorm.DB, items []SortItem) error
	ListByIDs(db *gorm.DB, ids []int64) ([]model.Resource, error)
}

// SortItem 排序条目
type SortItem struct {
	ID        int64  `json:"id"`
	SortOrder int    `json:"sort_order"`
	ParentID  *int64 `json:"parent_id"`
}

// resourceRepo ResourceRepository 默认实现
type resourceRepo struct{}

// NewResourceRepository 创建 ResourceRepository 实例
func NewResourceRepository() ResourceRepository {
	return &resourceRepo{}
}

// Create 创建资源
func (r *resourceRepo) Create(db *gorm.DB, resource *model.Resource) error {
	return db.Create(resource).Error
}

// Update 更新资源（乐观锁：WHERE version = oldVersion）
func (r *resourceRepo) Update(db *gorm.DB, resource *model.Resource) error {
	oldVersion := resource.Version
	resource.Version = oldVersion + 1
	result := db.Model(resource).
		Where("id = ? AND version = ?", resource.ID, oldVersion).
		Updates(map[string]interface{}{
			"parent_id":       resource.ParentID,
			"type":            resource.Type,
			"name":            resource.Name,
			"permission_code": resource.PermissionCode,
			"path":            resource.Path,
			"component":       resource.Component,
			"icon":            resource.Icon,
			"app_code":        resource.AppCode,
			"sort_order":      resource.SortOrder,
			"status":          resource.Status,
			"version":         resource.Version,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}
	return nil
}

// FindByID 根据 ID 查找资源
func (r *resourceRepo) FindByID(db *gorm.DB, id int64) (*model.Resource, error) {
	var resource model.Resource
	err := db.First(&resource, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrEntityNotFound, "资源不存在")
		}
		return nil, err
	}
	return &resource, nil
}

// ListByAppCode 按 app_code 查询资源列表
func (r *resourceRepo) ListByAppCode(db *gorm.DB, appCode string) ([]model.Resource, error) {
	var resources []model.Resource
	err := db.Where("app_code = ?", appCode).Order("sort_order ASC, created_at ASC").Find(&resources).Error
	return resources, err
}

// SoftDelete 软删除资源
func (r *resourceRepo) SoftDelete(db *gorm.DB, id int64) error {
	result := db.Delete(&model.Resource{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "资源不存在")
	}
	return nil
}

// HasChildren 检查资源是否有子节点
func (r *resourceRepo) HasChildren(db *gorm.DB, id int64) (bool, error) {
	var count int64
	err := db.Model(&model.Resource{}).Where("parent_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// BatchUpdateSort 批量更新排序（sort_order + parent_id）
func (r *resourceRepo) BatchUpdateSort(db *gorm.DB, items []SortItem) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			result := tx.Model(&model.Resource{}).
				Where("id = ?", item.ID).
				Updates(map[string]interface{}{
					"sort_order": item.SortOrder,
					"parent_id":  item.ParentID,
				})
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
}

// ListByIDs 根据 ID 列表查询资源
func (r *resourceRepo) ListByIDs(db *gorm.DB, ids []int64) ([]model.Resource, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var resources []model.Resource
	err := db.Where("id IN ?", ids).Order("sort_order ASC, created_at ASC").Find(&resources).Error
	return resources, err
}
