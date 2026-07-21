package repository

import (
	"go-admin/common/auth/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FieldPermissionRepository 字段权限数据访问接口
type FieldPermissionRepository interface {
	// GetByRolesAndObject 根据多个角色ID和对象代码查询字段权限
	GetByRolesAndObject(db *gorm.DB, roleIDs []int64, objectCode string) ([]model.FieldPermission, error)
	// BatchUpsert 批量插入或更新字段权限
	BatchUpsert(db *gorm.DB, items []model.FieldPermission) error
	// Delete 根据ID删除字段权限
	Delete(db *gorm.DB, id int64) error
	// ListByRoleAndObject 根据角色ID和对象代码查询字段权限列表
	ListByRoleAndObject(db *gorm.DB, roleID int64, objectCode string) ([]model.FieldPermission, error)
	// DeleteByRoleAndObject 删除指定角色对指定对象的所有字段权限
	DeleteByRoleAndObject(db *gorm.DB, roleID int64, objectCode string) error
}

// fieldPermissionRepo FieldPermissionRepository 默认实现
type fieldPermissionRepo struct{}

// NewFieldPermissionRepository 创建 FieldPermissionRepository 实例
func NewFieldPermissionRepository() FieldPermissionRepository {
	return &fieldPermissionRepo{}
}

// GetByRolesAndObject 根据多个角色ID和对象代码查询字段权限
func (r *fieldPermissionRepo) GetByRolesAndObject(db *gorm.DB, roleIDs []int64, objectCode string) ([]model.FieldPermission, error) {
	var list []model.FieldPermission
	err := db.Where("role_id IN ? AND object_code = ?", roleIDs, objectCode).Find(&list).Error
	return list, err
}

// BatchUpsert 批量插入或更新字段权限（冲突时更新 access）
func (r *fieldPermissionRepo) BatchUpsert(db *gorm.DB, items []model.FieldPermission) error {
	if len(items) == 0 {
		return nil
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}, {Name: "object_code"}, {Name: "field_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"access"}),
	}).Create(&items).Error
}

// Delete 根据ID删除字段权限
func (r *fieldPermissionRepo) Delete(db *gorm.DB, id int64) error {
	result := db.Delete(&model.FieldPermission{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListByRoleAndObject 根据角色ID和对象代码查询字段权限列表
func (r *fieldPermissionRepo) ListByRoleAndObject(db *gorm.DB, roleID int64, objectCode string) ([]model.FieldPermission, error) {
	var list []model.FieldPermission
	err := db.Where("role_id = ? AND object_code = ?", roleID, objectCode).Find(&list).Error
	return list, err
}

// DeleteByRoleAndObject 删除指定角色对指定对象的所有字段权限
func (r *fieldPermissionRepo) DeleteByRoleAndObject(db *gorm.DB, roleID int64, objectCode string) error {
	return db.Where("role_id = ? AND object_code = ?", roleID, objectCode).Delete(&model.FieldPermission{}).Error
}
