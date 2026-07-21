package repository

import (
	"go-admin/common/auth/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FieldObjectRepository 字段对象数据访问接口
type FieldObjectRepository interface {
	// UpsertObject 插入字段对象（已存在则忽略）
	UpsertObject(db *gorm.DB, obj *model.FieldObject) error
	// UpsertDefinition 插入字段定义（已存在则忽略，不覆盖已有描述）
	UpsertDefinition(db *gorm.DB, def *model.FieldDefinition) error
	// ListObjects 获取所有字段对象
	ListObjects(db *gorm.DB) ([]model.FieldObject, error)
	// FindObjectByCode 根据 object_code 查找字段对象
	FindObjectByCode(db *gorm.DB, objectCode string) (*model.FieldObject, error)
	// ListDefinitionsByObject 获取指定对象的所有字段定义
	ListDefinitionsByObject(db *gorm.DB, objectCode string) ([]model.FieldDefinition, error)
	// UpdateDefinitionDesc 更新字段定义的描述
	UpdateDefinitionDesc(db *gorm.DB, objectCode, fieldName, description string) error
	// CreateObject 创建字段对象
	CreateObject(db *gorm.DB, obj *model.FieldObject) error
	// CreateDefinition 创建字段定义
	CreateDefinition(db *gorm.DB, def *model.FieldDefinition) error
}

// fieldObjectRepo FieldObjectRepository 默认实现
type fieldObjectRepo struct{}

// NewFieldObjectRepository 创建 FieldObjectRepository 实例
func NewFieldObjectRepository() FieldObjectRepository {
	return &fieldObjectRepo{}
}

// UpsertObject 插入字段对象（已存在则忽略，不更新）
func (r *fieldObjectRepo) UpsertObject(db *gorm.DB, obj *model.FieldObject) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "object_code"}},
		DoNothing: true,
	}).Create(obj).Error
}

// UpsertDefinition 插入字段定义（已存在则忽略，不覆盖自定义描述——仅供自动注册使用）
func (r *fieldObjectRepo) UpsertDefinition(db *gorm.DB, def *model.FieldDefinition) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "object_code"}, {Name: "field_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"description", "source"}),
	}).Create(def).Error
}

// ListObjects 获取所有字段对象
func (r *fieldObjectRepo) ListObjects(db *gorm.DB) ([]model.FieldObject, error) {
	var list []model.FieldObject
	err := db.Order("created_at ASC").Find(&list).Error
	return list, err
}

// FindObjectByCode 根据 object_code 查找字段对象
func (r *fieldObjectRepo) FindObjectByCode(db *gorm.DB, objectCode string) (*model.FieldObject, error) {
	var obj model.FieldObject
	err := db.Where("object_code = ?", objectCode).First(&obj).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &obj, nil
}

// ListDefinitionsByObject 获取指定对象的所有字段定义
func (r *fieldObjectRepo) ListDefinitionsByObject(db *gorm.DB, objectCode string) ([]model.FieldDefinition, error) {
	var list []model.FieldDefinition
	err := db.Where("object_code = ?", objectCode).Order("created_at ASC").Find(&list).Error
	return list, err
}

// UpdateDefinitionDesc 更新字段定义的描述
func (r *fieldObjectRepo) UpdateDefinitionDesc(db *gorm.DB, objectCode, fieldName, description string) error {
	result := db.Model(&model.FieldDefinition{}).
		Where("object_code = ? AND field_name = ?", objectCode, fieldName).
		Update("description", description)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreateObject 创建字段对象
func (r *fieldObjectRepo) CreateObject(db *gorm.DB, obj *model.FieldObject) error {
	return db.Create(obj).Error
}

// CreateDefinition 创建字段定义
func (r *fieldObjectRepo) CreateDefinition(db *gorm.DB, def *model.FieldDefinition) error {
	return db.Create(def).Error
}
