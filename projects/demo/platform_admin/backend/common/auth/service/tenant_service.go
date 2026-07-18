package service

import (
	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// OperationLogger 操作日志记录器接口（占位，后续实现）
type OperationLogger interface {
	Log(operatorID int64, action string, targetType string, targetID int64, detail string)
}

// noopLogger 空实现，占位用
type noopLogger struct{}

func (n *noopLogger) Log(operatorID int64, action string, targetType string, targetID int64, detail string) {
	// TODO: 实现操作日志记录
}

// TenantService 租户业务逻辑层
type TenantService struct {
	db         *gorm.DB
	cfg        *config.Config
	tenantRepo repository.TenantRepository
	logger     OperationLogger
}

// NewTenantService 创建 TenantService 实例
func NewTenantService(db *gorm.DB, cfg *config.Config, tenantRepo repository.TenantRepository) *TenantService {
	return &TenantService{
		db:         db,
		cfg:        cfg,
		tenantRepo: tenantRepo,
		logger:     &noopLogger{},
	}
}

// SetLogger 设置操作日志记录器
func (s *TenantService) SetLogger(logger OperationLogger) {
	s.logger = logger
}

// CreateTenantRequest 创建租户请求参数
type CreateTenantRequest struct {
	TenantCode string `json:"tenant_code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	AdminUserID *int64 `json:"admin_user_id"` // 可选，指定已有用户作为管理员
}

// UpdateTenantRequest 更新租户请求参数
type UpdateTenantRequest struct {
	Name    string `json:"name"`
	Config  string `json:"config"`  // JSON 字符串
	Version int    `json:"version" binding:"required"` // 乐观锁版本号
}

// UpdateStatusRequest 更新状态请求参数
type UpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

// CreateTenant 创建租户 + 自动初始化
func (s *TenantService) CreateTenant(req *CreateTenantRequest) (*model.Tenant, error) {
	// 检查 tenant_code 唯一性
	existing, err := s.tenantRepo.FindByCode(s.db, req.TenantCode)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "租户编码已存在")
	}

	tenant := &model.Tenant{
		TenantCode: req.TenantCode,
		Name:       req.Name,
		Status:     1,
		Version:    1,
	}

	// 事务中完成创建 + 初始化
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建租户
		if err := s.tenantRepo.Create(tx, tenant); err != nil {
			return err
		}

		// 2. 创建预置角色
		var tenantAdminRoleID int64
		for _, tmpl := range s.cfg.Tenant.TemplateRoles {
			role := &model.Role{
				TenantID: tenant.ID,
				RoleCode: tmpl.Code,
				RoleName: tmpl.Name,
				RoleType: "tenant",
				Status:   1,
				Version:  1,
			}
			if err := tx.Create(role).Error; err != nil {
				return err
			}
			// 记录 tenant_admin 角色 ID
			if tmpl.Code == "tenant_admin" {
				tenantAdminRoleID = role.ID
			}
		}

		// 3. 创建或关联默认管理员
		var adminUserID int64
		if req.AdminUserID != nil {
			// 使用指定的已有用户
			adminUserID = *req.AdminUserID
		} else {
			// 创建默认管理员
			defaultPassword := "Admin@123"
			hashedPwd, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			adminUser := &model.User{
				Username: req.TenantCode + "_admin",
				Password: string(hashedPwd),
				Nickname: req.Name + " 管理员",
				Status:   1,
				Version:  1,
			}
			if err := tx.Create(adminUser).Error; err != nil {
				return err
			}
			adminUserID = adminUser.ID
		}

		// 4. 关联管理员到租户
		userTenant := &model.UserTenant{
			UserID:   adminUserID,
			TenantID: tenant.ID,
		}
		if err := tx.Create(userTenant).Error; err != nil {
			return err
		}

		// 5. 分配 tenant_admin 角色
		if tenantAdminRoleID > 0 {
			userRole := &model.UserRole{
				UserID:   adminUserID,
				RoleID:   tenantAdminRoleID,
				TenantID: tenant.ID,
			}
			if err := tx.Create(userRole).Error; err != nil {
				return err
			}
		}

		// 6. 订阅默认应用
		for _, appCode := range s.cfg.Tenant.DefaultApps {
			tenantApp := &model.TenantApp{
				TenantID: tenant.ID,
				AppCode:  appCode,
			}
			if err := tx.Create(tenantApp).Error; err != nil {
				return err
			}
		}

		// 7. API 权限现在是应用级别，不再需要按租户复制

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 记录操作日志
	s.logger.Log(0, "create_tenant", "tenant", tenant.ID, "创建租户: "+tenant.Name)

	return tenant, nil
}

// UpdateTenant 更新租户信息（乐观锁）
func (s *TenantService) UpdateTenant(id int64, req *UpdateTenantRequest) (*model.Tenant, error) {
	tenant, err := s.tenantRepo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}

	// 乐观锁版本校验
	if tenant.Version != req.Version {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}

	// 更新字段
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Config != "" {
		tenant.Config = []byte(req.Config)
	}

	if err := s.tenantRepo.Update(s.db, tenant); err != nil {
		return nil, err
	}

	s.logger.Log(0, "update_tenant", "tenant", tenant.ID, "更新租户: "+tenant.Name)
	return tenant, nil
}

// UpdateStatus 启用/禁用租户
func (s *TenantService) UpdateStatus(id int64, status int) error {
	if err := s.tenantRepo.UpdateStatus(s.db, id, status); err != nil {
		return err
	}
	action := "enable_tenant"
	if status == 0 {
		action = "disable_tenant"
	}
	s.logger.Log(0, action, "tenant", id, "")
	return nil
}

// DeleteTenant 软删除租户
func (s *TenantService) DeleteTenant(id int64) error {
	if err := s.tenantRepo.SoftDelete(s.db, id); err != nil {
		return err
	}
	s.logger.Log(0, "delete_tenant", "tenant", id, "")
	return nil
}

// TenantListResult 分页结果
type TenantListResult struct {
	List     []model.Tenant `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ListTenants 分页查询租户列表
func (s *TenantService) ListTenants(page, pageSize int, status *int, name string) (*TenantListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	params := repository.TenantListParams{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Name:     name,
	}

	list, total, err := s.tenantRepo.List(s.db, params)
	if err != nil {
		return nil, err
	}

	return &TenantListResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
