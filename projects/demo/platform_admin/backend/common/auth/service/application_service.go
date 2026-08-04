package service

import (
	"log"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"
	"go-admin/common/event"

	"gorm.io/gorm"
)

// RegisterApprovalListeners 注册审批结果事件监听，更新 admin_tenant_app.subscription_status。
// 由外部（cmd/api/server.go）在应用启动时调用，不在 init() 中调用，避免循环依赖。
func RegisterApprovalListeners(bus event.Bus, db *gorm.DB) {
	// 审批通过 → subscription_status = 'active'
	bus.Subscribe("approval.completed", func(payload interface{}) {
		e, ok := payload.(event.ApprovalCompletedEvent)
		if !ok {
			log.Printf("[RegisterApprovalListeners] approval.completed: 非预期 payload 类型 %T", payload)
			return
		}
		if e.BizType != "tenant_app_subscription" {
			return
		}
		if err := db.Model(&model.TenantApp{}).
			Where("app_code = ? AND tenant_id = ?", e.BizID, e.TenantID).
			Update("subscription_status", "active").Error; err != nil {
			log.Printf("[RegisterApprovalListeners] approval.completed 更新 subscription_status 失败: %v", err)
		}
	})

	// 审批驳回 → subscription_status = 'rejected'
	bus.Subscribe("approval.rejected", func(payload interface{}) {
		e, ok := payload.(event.ApprovalRejectedEvent)
		if !ok {
			log.Printf("[RegisterApprovalListeners] approval.rejected: 非预期 payload 类型 %T", payload)
			return
		}
		if e.BizType != "tenant_app_subscription" {
			return
		}
		if err := db.Model(&model.TenantApp{}).
			Where("app_code = ? AND tenant_id = ?", e.BizID, e.TenantID).
			Update("subscription_status", "rejected").Error; err != nil {
			log.Printf("[RegisterApprovalListeners] approval.rejected 更新 subscription_status 失败: %v", err)
		}
	})

	// 审批撤销 → subscription_status = 'rejected'（撤销视为未通过）
	bus.Subscribe("approval.cancelled", func(payload interface{}) {
		e, ok := payload.(event.ApprovalCancelledEvent)
		if !ok {
			log.Printf("[RegisterApprovalListeners] approval.cancelled: 非预期 payload 类型 %T", payload)
			return
		}
		if e.BizType != "tenant_app_subscription" {
			return
		}
		if err := db.Model(&model.TenantApp{}).
			Where("app_code = ? AND tenant_id = ?", e.BizID, e.TenantID).
			Update("subscription_status", "rejected").Error; err != nil {
			log.Printf("[RegisterApprovalListeners] approval.cancelled 更新 subscription_status 失败: %v", err)
		}
	})
}

// ApplicationService 应用管理业务逻辑层
type ApplicationService struct {
	db      *gorm.DB
	appRepo repository.ApplicationRepository
}

// NewApplicationService 创建 ApplicationService 实例
func NewApplicationService(db *gorm.DB, appRepo repository.ApplicationRepository) *ApplicationService {
	return &ApplicationService{
		db:      db,
		appRepo: appRepo,
	}
}

// CreateApplicationRequest 创建应用请求
type CreateApplicationRequest struct {
	AppCode     string `json:"app_code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	AppType     string `json:"app_type"` // BUILTIN/PLUGIN/EXTERNAL，默认 BUILTIN
	RoutePrefix string `json:"route_prefix"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
}

// validAppTypes 合法的应用类型枚举
var validAppTypes = map[string]bool{
	"":         true, // 允许空，默认 BUILTIN
	"BUILTIN":  true,
	"PLUGIN":   true,
	"EXTERNAL": true,
}

// UpdateApplicationRequest 更新应用请求
type UpdateApplicationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
	Version     int    `json:"version" binding:"required"`
}

// SetTenantAppsRequest 设置租户订阅应用请求
type SetTenantAppsRequest struct {
	AppCodes []string `json:"app_codes" binding:"required"`
}

// SetRoleAppsRequest 设置角色绑定应用请求
type SetRoleAppsRequest struct {
	AppCodes []string `json:"app_codes" binding:"required"`
}

// CreateApplication 创建应用
func (s *ApplicationService) CreateApplication(req *CreateApplicationRequest) (*model.Application, error) {
	// 校验 app_type 枚举值
	if !validAppTypes[req.AppType] {
		return nil, errors.NewAuthError(errors.ErrInvalidParam, "无效的 app_type，允许值: BUILTIN/PLUGIN/EXTERNAL")
	}

	// 检查 app_code 唯一性
	existing, err := s.appRepo.FindByCode(s.db, req.AppCode)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "应用编码已存在")
	}

	appType := req.AppType
	if appType == "" {
		appType = "BUILTIN"
	}

	app := &model.Application{
		AppCode:     req.AppCode,
		Name:        req.Name,
		Description: req.Description,
		AppType:     appType,
		RoutePrefix: req.RoutePrefix,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		Status:      1,
		Version:     1,
	}

	if err := s.appRepo.Create(s.db, app); err != nil {
		return nil, err
	}
	return app, nil
}

// UpdateApplication 更新应用
func (s *ApplicationService) UpdateApplication(id int64, req *UpdateApplicationRequest) (*model.Application, error) {
	app, err := s.appRepo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}

	// 乐观锁版本校验
	if app.Version != req.Version {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}

	if req.Name != "" {
		app.Name = req.Name
	}
	if req.Description != "" {
		app.Description = req.Description
	}
	if req.Status != nil {
		app.Status = *req.Status
	}

	if err := s.appRepo.Update(s.db, app); err != nil {
		return nil, err
	}
	return app, nil
}

// DeleteApplication 删除应用
func (s *ApplicationService) DeleteApplication(id int64) error {
	return s.appRepo.Delete(s.db, id)
}

// ListApplications 查询全局应用列表
func (s *ApplicationService) ListApplications() ([]model.Application, error) {
	return s.appRepo.List(s.db)
}

// SetTenantApps 设置租户订阅应用（全量替换），取消订阅时级联清除角色绑定
func (s *ApplicationService) SetTenantApps(tenantID int64, req *SetTenantAppsRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取当前订阅列表
		currentApps, err := s.appRepo.ListTenantApps(tx, tenantID)
		if err != nil {
			return err
		}

		// 计算被取消订阅的应用
		newSet := make(map[string]bool, len(req.AppCodes))
		for _, code := range req.AppCodes {
			newSet[code] = true
		}
		var removedCodes []string
		for _, ta := range currentApps {
			if !newSet[ta.AppCode] {
				removedCodes = append(removedCodes, ta.AppCode)
			}
		}

		// 级联清除被取消订阅的应用在该租户角色中的绑定
		if len(removedCodes) > 0 {
			if err := s.appRepo.DeleteRoleAppsByTenantAndCodes(tx, tenantID, removedCodes); err != nil {
				return err
			}
		}

		// 设置新的订阅列表
		return s.appRepo.SetTenantApps(tx, tenantID, req.AppCodes)
	})
}

// ListTenantApps 查询租户已订阅的应用列表（返回完整应用信息）
func (s *ApplicationService) ListTenantApps(tenantID int64) ([]model.Application, error) {
	// 先查租户订阅记录获取 app_codes
	tenantApps, err := s.appRepo.ListTenantApps(s.db, tenantID)
	if err != nil {
		return nil, err
	}
	if len(tenantApps) == 0 {
		return []model.Application{}, nil
	}

	// 根据 app_codes 查询完整应用信息
	codes := make([]string, 0, len(tenantApps))
	for _, ta := range tenantApps {
		codes = append(codes, ta.AppCode)
	}

	var apps []model.Application
	if err := s.db.Where("app_code IN ?", codes).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

// SetRoleApps 角色绑定应用（校验租户已订阅）
func (s *ApplicationService) SetRoleApps(roleID int64, tenantID int64, req *SetRoleAppsRequest) error {
	// 校验所有应用是否已被租户订阅
	notSubscribed, err := s.appRepo.AreTenantSubscribed(s.db, tenantID, req.AppCodes)
	if err != nil {
		return err
	}
	if len(notSubscribed) > 0 {
		return errors.NewAuthError(errors.ErrTenantAppNotSubscribed, "租户未订阅应用: "+notSubscribed[0])
	}

	return s.appRepo.SetRoleApps(s.db, roleID, req.AppCodes)
}
