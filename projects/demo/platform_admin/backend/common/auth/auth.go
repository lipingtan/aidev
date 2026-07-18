package auth

import (
	"log"

	pluginRouter "go-admin/app/plugin/router"
	"go-admin/common/auth/config"
	"go-admin/common/auth/discovery"
	"go-admin/common/auth/handler"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Init 初始化 auth-rbac 模块
// 完成全部初始化：AutoMigrate + 创建依赖 + 注册路由 + 启动 API 发现
func Init(cfg *config.Config, db *gorm.DB, engine *gin.Engine) error {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if !cfg.Enabled {
		log.Println("[auth-rbac] 模块已禁用，跳过初始化")
		return nil
	}

	// AutoMigrate
	if err := autoMigrate(db); err != nil {
		return err
	}

	// 创建依赖
	deps := buildDependencies(cfg, db)

	// 将 AuthService 注入到插件路由（供 InitPluginRouter 使用）
	pluginRouter.SetAuthService(deps.AuthService)

	// 注册路由
	rg := engine.Group("")
	RegisterRoutes(rg, deps)

	// 启动 API 自动发现
	if cfg.APIDiscovery.Enabled {
		discovery.AutoDiscover(engine, db, cfg.APIDiscovery.AsyncThreshold)
	}

	log.Printf("[auth-rbac] 模块初始化完成 auth-type=%s cache-type=%s", cfg.AuthType, cfg.CacheType)
	return nil
}

// AutoMigratePublic 导出的表迁移函数，供 setup 等外部包调用
func AutoMigratePublic(db *gorm.DB) error {
	return autoMigrate(db)
}

// autoMigrate 执行数据库表自动迁移
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Tenant{},
		&model.Role{},
		&model.UserTenant{},
		&model.UserRole{},
		&model.TenantApp{},
		&model.ApiPermission{},
		&model.Resource{},
		&model.Application{},
		&model.RoleApi{},
		&model.RoleResource{},
		&model.RoleApp{},
		&model.OperationLog{},
		&model.DataScope{},
		&model.DataScopeConfig{},
		&model.SysConfig{},
		&model.LoginLog{},
	)
}

// buildDependencies 创建所有 service/handler 依赖
func buildDependencies(cfg *config.Config, db *gorm.DB) *Dependencies {
	// Repositories
	tenantRepo := repository.NewTenantRepository()
	userRepo := repository.NewUserRepository()
	roleRepo := repository.NewRoleRepository()
	apiPermRepo := repository.NewApiPermissionRepository()
	resourceRepo := repository.NewResourceRepository()
	appRepo := repository.NewApplicationRepository()
	dataScopeConfigRepo := repository.NewDataScopeConfigRepository()
	dataScopeRepo := repository.NewDataScopeRepository()
	operationLogRepo := repository.NewOperationLogRepository()

	// Services
	authSvc := service.NewAuthService(db, cfg)
	tenantSvc := service.NewTenantService(db, cfg, tenantRepo)
	userSvc := service.NewUserService(db, cfg, userRepo)
	roleSvc := service.NewRoleService(db, cfg, roleRepo)
	userRoleSvc := service.NewUserRoleService(db, cfg)
	apiPermSvc := service.NewApiPermissionService(db, apiPermRepo)
	resourceSvc := service.NewResourceService(db, cfg, resourceRepo)
	appSvc := service.NewApplicationService(db, appRepo)
	dataScopeSvc := service.NewDataScopeService(db, dataScopeConfigRepo, dataScopeRepo)
	opLogQuerySvc := service.NewOperationLogQueryService(db, operationLogRepo)
	configSvc := service.NewConfigService(db)
	loginLogSvc := service.NewLoginLogService(db)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	tenantHandler := handler.NewTenantHandler(tenantSvc)
	userHandler := handler.NewUserHandler(userSvc)
	userHandler.SetRoleService(userRoleSvc)
	roleHandler := handler.NewRoleHandler(roleSvc)
	apiPermHandler := handler.NewApiPermissionHandler(apiPermSvc)
	resourceHandler := handler.NewResourceHandler(resourceSvc)
	appHandler := handler.NewApplicationHandler(appSvc, roleRepo, db)
	dataScopeHandler := handler.NewDataScopeHandler(dataScopeSvc)
	opLogHandler := handler.NewOperationLogHandler(opLogQuerySvc)
	configHandler := handler.NewConfigHandler(configSvc)
	loginLogHandler := handler.NewLoginLogHandler(loginLogSvc)

	return &Dependencies{
		DB:  db,
		Cfg: cfg,

		AuthService:    authSvc,
		TenantService:  tenantSvc,
		UserService:    userSvc,
		RoleService:    roleSvc,
		UserRoleService: userRoleSvc,

		AuthHandler:   authHandler,
		TenantHandler: tenantHandler,
		UserHandler:   userHandler,
		RoleHandler:   roleHandler,

		ApiPermissionService: apiPermSvc,
		ApiPermissionHandler: apiPermHandler,

		ResourceService: resourceSvc,
		ResourceHandler: resourceHandler,

		ApplicationService: appSvc,
		ApplicationHandler: appHandler,

		DataScopeService: dataScopeSvc,
		DataScopeHandler: dataScopeHandler,

		OperationLogQueryService: opLogQuerySvc,
		OperationLogHandler:      opLogHandler,

		ConfigService:  configSvc,
		ConfigHandler:  configHandler,

		LoginLogService: loginLogSvc,
		LoginLogHandler: loginLogHandler,
	}
}
