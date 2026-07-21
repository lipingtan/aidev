package service

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// QuotaService 配额校验服务
type QuotaService struct {
	db        *gorm.DB
	configSvc *AdminConfigService
}

// NewQuotaService 创建配额校验服务
func NewQuotaService(db *gorm.DB, configSvc *AdminConfigService) *QuotaService {
	return &QuotaService{db: db, configSvc: configSvc}
}

// CheckUserQuota 校验租户管理用户数配额
// isSuperAdmin=true 时跳过校验
func (q *QuotaService) CheckUserQuota(tenantID int64, isSuperAdmin bool) error {
	if isSuperAdmin {
		return nil
	}
	var count int64
	q.db.Model(&model.UserTenant{}).Where("tenant_id = ?", tenantID).Count(&count)
	max := q.configSvc.ResolveInt(tenantID, "quota.max_admin_users", 999999)
	if int(count) >= max {
		return errors.NewAuthError(errors.ErrQuotaExceeded, "管理用户数已达上限")
	}
	return nil
}

// CheckRoleQuota 校验租户角色数配额
func (q *QuotaService) CheckRoleQuota(tenantID int64, isSuperAdmin bool) error {
	if isSuperAdmin {
		return nil
	}
	var count int64
	q.db.Model(&model.Role{}).Where("tenant_id = ?", tenantID).Count(&count)
	max := q.configSvc.ResolveInt(tenantID, "quota.max_roles", 999999)
	if int(count) >= max {
		return errors.NewAuthError(errors.ErrQuotaExceeded, "角色数已达上限")
	}
	return nil
}

// CheckAppQuota 校验租户可订阅应用数配额
func (q *QuotaService) CheckAppQuota(tenantID int64, isSuperAdmin bool) error {
	if isSuperAdmin {
		return nil
	}
	var count int64
	q.db.Model(&model.TenantApp{}).Where("tenant_id = ?", tenantID).Count(&count)
	max := q.configSvc.ResolveInt(tenantID, "quota.max_apps", 999999)
	if int(count) >= max {
		return errors.NewAuthError(errors.ErrQuotaExceeded, "可订阅应用数已达上限")
	}
	return nil
}
