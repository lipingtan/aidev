package auth

import (
	"log"

	pluginRouter "go-admin/app/plugin/router"
	"go-admin/common/auth/config"
	"go-admin/common/auth/discovery"
	"go-admin/common/auth/handler"
	"go-admin/common/auth/middleware"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"
	"go-admin/common/auth/spi"
	"go-admin/common/plugin"

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

	// Seed 初始数据
	if err := SeedInitialData(db); err != nil {
		log.Printf("[auth-rbac] Seed 失败: %v", err)
	}
	// Seed 配额默认值
	if err := SeedConfigDefaults(db); err != nil {
		log.Printf("[auth-rbac] Seed 配置默认值失败: %v", err)
	}

	// 清理 sys_menu 中遗留的插件菜单（一次性迁移）
	cleanLegacyPluginMenus(db)

	// 字段权限自动注册
	registerFieldPermModels(deps.FieldRegistry)

	// 启动 API 自动发现
	if cfg.APIDiscovery.Enabled {
		discovery.AutoDiscover(engine, db, cfg.APIDiscovery.AsyncThreshold)
	}

	// 加载中间件缓存（AutoDiscover 写入 module_code 后再加载）
	deps.ModuleCodeCache.Load(db)
	deps.AppPrefixMap.Load(db)

	// sys_config → admin_config 数据迁移（幂等）
	if err := deps.AdminConfigService.MigrateFromSysConfig(); err != nil {
		log.Printf("[auth-rbac] sys_config 迁移失败: %v", err)
	}

	// 注册数据权限 GORM Callback（含 OrganizationProvider）
	orgProvider := &spi.DefaultOrganizationProvider{DB: db}
	middleware.RegisterDataScopeCallback(db, cfg.DataScope.Enabled, orgProvider)

	// 加载字段权限路由映射
	deps.FieldObjectRegistry.LoadRoutes(db)
	// 硬编码已知路由
	deps.FieldObjectRegistry.RegisterRoute("GET", "/api/v1/admin/users", "user")
	deps.FieldObjectRegistry.RegisterRoute("GET", "/api/v1/admin/users/:id", "user")
	deps.FieldObjectRegistry.RegisterRoute("GET", "/api/v1/admin/tenants", "tenant")
	deps.FieldObjectRegistry.RegisterRoute("GET", "/api/v1/admin/tenants/:id", "tenant")

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
		&model.FieldObject{},
		&model.FieldDefinition{},
		&model.FieldPermission{},
		&model.RecordShare{},
		&model.OrgUnit{},
		&model.UserOrg{},
		&model.AdminConfig{},
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
	fieldObjectRepo := repository.NewFieldObjectRepository()
	fieldPermRepo := repository.NewFieldPermissionRepository()
	recordShareRepo := repository.NewRecordShareRepository()

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
	fieldRegistrySvc := service.NewFieldRegistry(db, fieldObjectRepo)
	fieldPermSvc := service.NewFieldPermissionService(db, fieldObjectRepo, fieldPermRepo)
	recordShareSvc := service.NewRecordShareService(db, recordShareRepo)
	orgUnitRepo := repository.NewOrgUnitRepository()
	orgUnitSvc := service.NewOrgUnitService(db, orgUnitRepo)
	adminConfigSvc := service.NewAdminConfigService(db)

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
	fieldPermHandler := handler.NewFieldPermissionHandler(fieldPermSvc)
	recordShareHandler := handler.NewRecordShareHandler(recordShareSvc)
	orgUnitHandler := handler.NewOrgUnitHandler(orgUnitSvc)
	adminConfigHandler := handler.NewAdminConfigHandler(adminConfigSvc)

	// 插件系统依赖
	pluginsDir := "./plugins"
	staticDir := "./static/plugins"
	pluginSyncer := plugin.NewPluginResourceSyncer(db)
	pluginMgr := plugin.NewPluginManager(db, pluginsDir)
	pluginInstaller := plugin.NewInstaller(pluginsDir, staticDir, db, pluginSyncer)

	// 清理残留的升级目录
	if err := pluginInstaller.RecoverStaleUpgrade(); err != nil {
		log.Printf("[auth-rbac] 插件升级残留清理失败: %v", err)
	}

	// 插件管理 Handler
	pluginHandler := handler.NewPluginHandler(pluginMgr, pluginInstaller)
	appCatalogHandler := handler.NewAppCatalogHandler(db, pluginMgr)

	return &Dependencies{
		DB:  db,
		Cfg: cfg,

		AuthService:     authSvc,
		TenantService:   tenantSvc,
		UserService:     userSvc,
		RoleService:     roleSvc,
		UserRoleService: userRoleSvc,
		FieldRegistry:   fieldRegistrySvc,

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

		FieldPermissionService: fieldPermSvc,
		FieldPermissionHandler: fieldPermHandler,

		RecordShareService: recordShareSvc,
		RecordShareHandler: recordShareHandler,

		OrgUnitService:      orgUnitSvc,
		OrgUnitHandler:      orgUnitHandler,

		AdminConfigService: adminConfigSvc,
		AdminConfigHandler: adminConfigHandler,

		AppCatalogHandler: appCatalogHandler,
		PluginHandler:     pluginHandler,

		AppPrefixMap:         middleware.NewAppPrefixMap(),
		ModuleCodeCache:      middleware.NewModuleCodeCache(),
		FieldObjectRegistry:  middleware.NewFieldObjectRegistry(),
		FieldPermissionRepo:  fieldPermRepo,
	}
}

// registerFieldPermModels 注册所有需要字段权限管控的 model
func registerFieldPermModels(registry *service.FieldRegistry) {
	registry.AutoRegister("user", "用户", model.User{})
	registry.AutoRegister("tenant", "租户", model.Tenant{})
}

// cleanLegacyPluginMenus 清理 sys_menu 中遗留的插件菜单
// 将 menu_name 以 plugin_ 开头的记录执行软删除（设置 deleted_at）
// 幂等：已软删除的记录不会再次处理
func cleanLegacyPluginMenus(db *gorm.DB) {
	result := db.Exec("UPDATE sys_menu SET deleted_at = NOW() WHERE menu_name LIKE 'plugin_%' AND deleted_at IS NULL")
	if result.Error != nil {
		log.Printf("[auth-rbac] 清理遗留插件菜单失败: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("[auth-rbac] 清理了 %d 条遗留插件菜单", result.RowsAffected)
	}
}
