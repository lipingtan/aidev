package service

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

// SetPermissionItem 批量设置字段权限的单项
type SetPermissionItem struct {
	FieldName string `json:"field_name" binding:"required"`
	Access    string `json:"access" binding:"required"` // VISIBLE/EDITABLE/HIDDEN
}

// FieldPermissionService 字段权限业务服务
type FieldPermissionService struct {
	db             *gorm.DB
	fieldObjectRepo repository.FieldObjectRepository
	fieldPermRepo  repository.FieldPermissionRepository
}

// NewFieldPermissionService 创建 FieldPermissionService 实例
func NewFieldPermissionService(
	db *gorm.DB,
	fieldObjectRepo repository.FieldObjectRepository,
	fieldPermRepo repository.FieldPermissionRepository,
) *FieldPermissionService {
	return &FieldPermissionService{
		db:             db,
		fieldObjectRepo: fieldObjectRepo,
		fieldPermRepo:  fieldPermRepo,
	}
}

// ListObjects 返回已注册对象列表
func (s *FieldPermissionService) ListObjects() ([]model.FieldObject, error) {
	return s.fieldObjectRepo.ListObjects(s.db)
}

// ListFields 返回对象字段列表
func (s *FieldPermissionService) ListFields(objectCode string) ([]model.FieldDefinition, error) {
	return s.fieldObjectRepo.ListDefinitionsByObject(s.db, objectCode)
}

// UpdateFieldDesc 修改字段描述
func (s *FieldPermissionService) UpdateFieldDesc(objectCode, fieldName, description string) error {
	return s.fieldObjectRepo.UpdateDefinitionDesc(s.db, objectCode, fieldName, description)
}

// GetPermissions 查询角色字段权限
func (s *FieldPermissionService) GetPermissions(roleID int64, objectCode string) ([]model.FieldPermission, error) {
	return s.fieldPermRepo.ListByRoleAndObject(s.db, roleID, objectCode)
}

// validAccessLevels 合法的字段权限级别
var validAccessLevels = map[string]bool{
	"HIDDEN":   true,
	"VISIBLE":  true,
	"EDITABLE": true,
}

// SetPermissions 批量设置角色字段权限（全量替换该角色对该对象的字段权限）
func (s *FieldPermissionService) SetPermissions(roleID int64, objectCode string, items []SetPermissionItem) error {
	// 校验 object_code 是否已注册
	obj, err := s.fieldObjectRepo.FindObjectByCode(s.db, objectCode)
	if err != nil {
		return err
	}
	if obj == nil {
		return errors.NewAuthError(errors.ErrInvalidParam, "对象 "+objectCode+" 未注册")
	}

	// 校验每项的 access 枚举值
	for _, item := range items {
		if !validAccessLevels[item.Access] {
			return errors.NewAuthError(errors.ErrInvalidParam, "无效的 access 值: "+item.Access+"，允许值: HIDDEN/VISIBLE/EDITABLE")
		}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除该角色对该对象的所有现有权限
		if err := s.fieldPermRepo.DeleteByRoleAndObject(tx, roleID, objectCode); err != nil {
			return err
		}
		// 批量插入新权限
		if len(items) == 0 {
			return nil
		}
		perms := make([]model.FieldPermission, 0, len(items))
		for _, item := range items {
			perms = append(perms, model.FieldPermission{
				RoleID:     roleID,
				ObjectCode: objectCode,
				FieldName:  item.FieldName,
				Access:     item.Access,
			})
		}
		return s.fieldPermRepo.BatchUpsert(tx, perms)
	})
}

// DeletePermission 删除字段权限配置
func (s *FieldPermissionService) DeletePermission(id int64) error {
	return s.fieldPermRepo.Delete(s.db, id)
}

// ManualRegisterObject 手动注册对象（已存在则返回冲突错误）
func (s *FieldPermissionService) ManualRegisterObject(objectCode, objectName string) error {
	// 先检查是否已存在
	existing, err := s.fieldObjectRepo.FindObjectByCode(s.db, objectCode)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.NewAuthError(errors.ErrDuplicateEntity, "object_code "+objectCode+" 已存在")
	}

	obj := &model.FieldObject{
		ObjectCode: objectCode,
		ObjectName: objectName,
		Source:     "MANUAL",
	}
	return s.fieldObjectRepo.CreateObject(s.db, obj)
}

// ManualRegisterField 手动注册字段（已存在则更新描述）
func (s *FieldPermissionService) ManualRegisterField(objectCode, fieldName, description string) error {
	def := &model.FieldDefinition{
		ObjectCode:  objectCode,
		FieldName:   fieldName,
		Description: description,
		Source:      "MANUAL",
	}
	return s.fieldObjectRepo.UpsertDefinition(s.db, def)
}
