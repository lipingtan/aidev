package repository

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// ApplicationRepository 应用数据访问接口
type ApplicationRepository interface {
	Create(db *gorm.DB, app *model.Application) error
	Update(db *gorm.DB, app *model.Application) error
	FindByID(db *gorm.DB, id int64) (*model.Application, error)
	FindByCode(db *gorm.DB, code string) (*model.Application, error)
	List(db *gorm.DB) ([]model.Application, error)
	Delete(db *gorm.DB, id int64) error

	// 租户订阅
	SetTenantApps(db *gorm.DB, tenantID int64, appCodes []string) error
	ListTenantApps(db *gorm.DB, tenantID int64) ([]model.TenantApp, error)
	IsTenantSubscribed(db *gorm.DB, tenantID int64, appCode string) (bool, error)
	AreTenantSubscribed(db *gorm.DB, tenantID int64, appCodes []string) ([]string, error)

	// 角色-应用绑定
	SetRoleApps(db *gorm.DB, roleID int64, appCodes []string) error
	ListRoleApps(db *gorm.DB, roleID int64) ([]model.RoleApp, error)
	DeleteRoleAppsByTenantAndCodes(db *gorm.DB, tenantID int64, appCodes []string) error
}

// applicationRepo ApplicationRepository 默认实现
type applicationRepo struct{}

// NewApplicationRepository 创建 ApplicationRepository 实例
func NewApplicationRepository() ApplicationRepository {
	return &applicationRepo{}
}

// Create 创建应用
func (r *applicationRepo) Create(db *gorm.DB, app *model.Application) error {
	return db.Create(app).Error
}

// Update 更新应用（乐观锁）
func (r *applicationRepo) Update(db *gorm.DB, app *model.Application) error {
	oldVersion := app.Version
	app.Version = oldVersion + 1
	result := db.Model(app).
		Where("id = ? AND version = ?", app.ID, oldVersion).
		Updates(map[string]interface{}{
			"name":        app.Name,
			"description": app.Description,
			"status":      app.Status,
			"version":     app.Version,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}
	return nil
}

// FindByID 根据 ID 查找应用
func (r *applicationRepo) FindByID(db *gorm.DB, id int64) (*model.Application, error) {
	var app model.Application
	err := db.First(&app, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrEntityNotFound, "应用不存在")
		}
		return nil, err
	}
	return &app, nil
}

// FindByCode 根据 app_code 查找应用
func (r *applicationRepo) FindByCode(db *gorm.DB, code string) (*model.Application, error) {
	var app model.Application
	err := db.Where("app_code = ?", code).First(&app).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

// List 查询全部应用列表
func (r *applicationRepo) List(db *gorm.DB) ([]model.Application, error) {
	var apps []model.Application
	err := db.Order("created_at DESC").Find(&apps).Error
	return apps, err
}

// Delete 硬删除应用（软删除会导致 app_code unique index 残留，相同编码无法复用）
func (r *applicationRepo) Delete(db *gorm.DB, id int64) error {
	result := db.Unscoped().Delete(&model.Application{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NewAuthError(errors.ErrEntityNotFound, "应用不存在")
	}
	return nil
}

// SetTenantApps 设置租户订阅的应用列表（全量替换）
func (r *applicationRepo) SetTenantApps(db *gorm.DB, tenantID int64, appCodes []string) error {
	// 删除旧订阅
	if err := db.Where("tenant_id = ?", tenantID).Delete(&model.TenantApp{}).Error; err != nil {
		return err
	}
	// 创建新订阅
	for _, code := range appCodes {
		ta := &model.TenantApp{
			TenantID: tenantID,
			AppCode:  code,
		}
		if err := db.Create(ta).Error; err != nil {
			return err
		}
	}
	return nil
}

// ListTenantApps 查询租户已订阅应用
func (r *applicationRepo) ListTenantApps(db *gorm.DB, tenantID int64) ([]model.TenantApp, error) {
	var apps []model.TenantApp
	err := db.Where("tenant_id = ?", tenantID).Find(&apps).Error
	return apps, err
}

// IsTenantSubscribed 检查租户是否订阅了某应用
func (r *applicationRepo) IsTenantSubscribed(db *gorm.DB, tenantID int64, appCode string) (bool, error) {
	var count int64
	err := db.Model(&model.TenantApp{}).Where("tenant_id = ? AND app_code = ?", tenantID, appCode).Count(&count).Error
	return count > 0, err
}

// AreTenantSubscribed 批量检查租户是否订阅了给定应用，返回未订阅的列表
func (r *applicationRepo) AreTenantSubscribed(db *gorm.DB, tenantID int64, appCodes []string) ([]string, error) {
	if len(appCodes) == 0 {
		return nil, nil
	}
	var subscribedCodes []string
	err := db.Model(&model.TenantApp{}).
		Where("tenant_id = ? AND app_code IN ?", tenantID, appCodes).
		Pluck("app_code", &subscribedCodes).Error
	if err != nil {
		return nil, err
	}
	// 找出未订阅的
	subscribedSet := make(map[string]bool, len(subscribedCodes))
	for _, c := range subscribedCodes {
		subscribedSet[c] = true
	}
	var notSubscribed []string
	for _, c := range appCodes {
		if !subscribedSet[c] {
			notSubscribed = append(notSubscribed, c)
		}
	}
	return notSubscribed, nil
}

// SetRoleApps 设置角色绑定的应用列表（全量替换）
func (r *applicationRepo) SetRoleApps(db *gorm.DB, roleID int64, appCodes []string) error {
	// 删除旧绑定
	if err := db.Where("role_id = ?", roleID).Delete(&model.RoleApp{}).Error; err != nil {
		return err
	}
	// 创建新绑定
	for _, code := range appCodes {
		ra := &model.RoleApp{
			RoleID:  roleID,
			AppCode: code,
		}
		if err := db.Create(ra).Error; err != nil {
			return err
		}
	}
	return nil
}

// ListRoleApps 查询角色绑定的应用
func (r *applicationRepo) ListRoleApps(db *gorm.DB, roleID int64) ([]model.RoleApp, error) {
	var apps []model.RoleApp
	err := db.Where("role_id = ?", roleID).Find(&apps).Error
	return apps, err
}

// DeleteRoleAppsByTenantAndCodes 删除指定租户下所有角色对指定应用的绑定
func (r *applicationRepo) DeleteRoleAppsByTenantAndCodes(db *gorm.DB, tenantID int64, appCodes []string) error {
	if len(appCodes) == 0 {
		return nil
	}
	// 查找该租户的所有角色 ID
	var roleIDs []int64
	err := db.Model(&model.Role{}).Where("tenant_id = ?", tenantID).Pluck("id", &roleIDs).Error
	if err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	// 删除这些角色对指定应用的绑定
	return db.Where("role_id IN ? AND app_code IN ?", roleIDs, appCodes).Delete(&model.RoleApp{}).Error
}
