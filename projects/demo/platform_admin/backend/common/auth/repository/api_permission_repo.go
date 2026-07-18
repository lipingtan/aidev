package repository

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// ApiPermissionRepository 接口权限数据访问接口
type ApiPermissionRepository interface {
	Create(db *gorm.DB, perm *model.ApiPermission) error
	Update(db *gorm.DB, perm *model.ApiPermission) error
	FindByID(db *gorm.DB, id int64) (*model.ApiPermission, error)
	ListByAppCode(db *gorm.DB, appCode string) ([]model.ApiPermission, error)
	Delete(db *gorm.DB, id int64) error
	HasChildren(db *gorm.DB, id int64) (bool, error)
	ListUnassigned(db *gorm.DB, appCode string) ([]model.ApiPermission, error)
	Move(db *gorm.DB, id int64, parentID int64) error
}

// apiPermissionRepo ApiPermissionRepository 默认实现
type apiPermissionRepo struct{}

// NewApiPermissionRepository 创建 ApiPermissionRepository 实例
func NewApiPermissionRepository() ApiPermissionRepository {
	return &apiPermissionRepo{}
}

// Create 创建接口权限节点
func (r *apiPermissionRepo) Create(db *gorm.DB, perm *model.ApiPermission) error {
	return db.Create(perm).Error
}

// Update 更新接口权限节点
func (r *apiPermissionRepo) Update(db *gorm.DB, perm *model.ApiPermission) error {
	result := db.Model(perm).Where("id = ?", perm.ID).Updates(map[string]interface{}{
		"name":            perm.Name,
		"permission_code": perm.PermissionCode,
		"url_pattern":     perm.URLPattern,
		"http_method":     perm.HTTPMethod,
		"app_code":        perm.AppCode,
		"status":          perm.Status,
		"sort_order":      perm.SortOrder,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "接口权限节点不存在")
	}
	return nil
}

// FindByID 根据 ID 查找接口权限节点
func (r *apiPermissionRepo) FindByID(db *gorm.DB, id int64) (*model.ApiPermission, error) {
	var perm model.ApiPermission
	err := db.First(&perm, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrEntityNotFound, "接口权限节点不存在")
		}
		return nil, err
	}
	return &perm, nil
}

// ListByAppCode 按 app_code 查询所有接口权限节点
func (r *apiPermissionRepo) ListByAppCode(db *gorm.DB, appCode string) ([]model.ApiPermission, error) {
	var perms []model.ApiPermission
	err := db.Where("app_code = ?", appCode).Order("sort_order ASC, id ASC").Find(&perms).Error
	return perms, err
}

// Delete 硬删除接口权限节点（无 DeletedAt 字段）
func (r *apiPermissionRepo) Delete(db *gorm.DB, id int64) error {
	result := db.Unscoped().Delete(&model.ApiPermission{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "接口权限节点不存在")
	}
	return nil
}

// HasChildren 检查节点是否有子节点
func (r *apiPermissionRepo) HasChildren(db *gorm.DB, id int64) (bool, error) {
	var count int64
	err := db.Model(&model.ApiPermission{}).Where("parent_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListUnassigned 查询未分组 ENDPOINT 列表（status=UNASSIGNED）
func (r *apiPermissionRepo) ListUnassigned(db *gorm.DB, appCode string) ([]model.ApiPermission, error) {
	var perms []model.ApiPermission
	err := db.Where("app_code = ? AND status = ? AND type = ?", appCode, "UNASSIGNED", "ENDPOINT").
		Order("sort_order ASC, id ASC").Find(&perms).Error
	return perms, err
}

// Move 移动节点到指定 GROUP 下，更新 parent_id + status=ACTIVE
func (r *apiPermissionRepo) Move(db *gorm.DB, id int64, parentID int64) error {
	result := db.Model(&model.ApiPermission{}).Where("id = ?", id).Updates(map[string]interface{}{
		"parent_id": parentID,
		"status":    "ACTIVE",
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "接口权限节点不存在")
	}
	return nil
}
