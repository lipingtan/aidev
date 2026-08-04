package auth

import (
	"log"
	"time"

	userAuth "go-admin/app/user_auth"
	userAuthHandler "go-admin/app/user_auth/handler"
	userAuthModel "go-admin/app/user_auth/model"
	userAuthRepo "go-admin/app/user_auth/repository"
	userAuthService "go-admin/app/user_auth/service"
	userAuthSpi "go-admin/app/user_auth/spi"
	userStrategy "go-admin/app/user_auth/strategy"
	pluginRouter "go-admin/app/plugin/router"
	authCache "go-admin/common/auth/cache"
	abacmodel "go-admin/common/auth/abac/model"
	abachandler "go-admin/common/auth/abac/handler"
	abacrepo "go-admin/common/auth/abac/repository"
	abacservice "go-admin/common/auth/abac/service"
	"go-admin/common/auth/abac"
	"go-admin/common/auth/config"
	"go-admin/common/auth/discovery"
	"go-admin/common/auth/handler"
	"go-admin/common/auth/middleware"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"
	"go-admin/common/auth/service"
	"go-admin/common/auth/spi"
	"go-admin/common/event"
	"go-admin/common/plugin"

	"github.com/go-admin-team/go-admin-core/sdk"
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

	// 创建依赖（传入 Redis blacklist，有 CacheAdapter 则用 Redis，否则降级到内存）
	var blacklistStore spi.TokenBlacklistStore
	if cacheAdapterForBL := sdk.Runtime.GetCacheAdapter(); cacheAdapterForBL != nil {
		blacklistStore = authCache.NewRedisBlacklistStore(cacheAdapterForBL)
		log.Println("[auth-rbac] Token 黑名单使用 Redis 存储（多 Pod 安全）")
	} else {
		log.Println("[auth-rbac] Token 黑名单降级为进程内存储（单实例模式）")
	}
	deps := buildDependencies(cfg, db, blacklistStore)

	// CR-10: 初始化 ABAC 依赖（必须在 RegisterRoutes 之前，否则路由注册时 AbacHandler 为 nil）
	buildAbacDependencies(db, deps)

	// 将 AuthService 注入到插件路由（供 InitPluginRouter 使用）
	pluginRouter.SetAuthService(deps.AuthService)

	// 注册路由
	rg := engine.Group("")
	RegisterRoutes(rg, deps)

	// 注册 user_auth 域路由
	initUserAuth(cfg, deps, engine)

	// Seed 初始数据
	if err := SeedInitialData(db); err != nil {
		log.Printf("[auth-rbac] Seed 失败: %v", err)
	}
	// Seed 配额默认值
	if err := SeedConfigDefaults(db); err != nil {
		log.Printf("[auth-rbac] Seed 配置默认值失败: %v", err)
	}
	// Seed CR-2 权限管理菜单（增量，幂等）
	if err := SeedCR2Menus(db); err != nil {
		log.Printf("[auth-rbac] Seed CR-2 菜单失败: %v", err)
	}
	// Seed CR-5 C端用户管理菜单（增量，幂等）
	if err := SeedCR5Menus(db); err != nil {
		log.Printf("[auth-rbac] Seed CR-5 菜单失败: %v", err)
	}
	// Seed CR-6 域名管理菜单（增量，幂等）
	if err := SeedCR6Menus(db); err != nil {
		log.Printf("[auth-rbac] Seed CR-6 菜单失败: %v", err)
	}
	// Seed CR-10 ABAC 策略管理菜单（增量，幂等）
	if err := SeedCR10Menus(db); err != nil {
		log.Printf("[auth-rbac] Seed CR-10 菜单失败: %v", err)
	}

	// 字段权限自动注册
	registerFieldPermModels(deps.FieldRegistry)

	// 启动 API 自动发现
	if cfg.APIDiscovery.Enabled {
		discovery.AutoDiscover(engine, db, cfg.APIDiscovery.AsyncThreshold)
	}

	// 加载中间件缓存（AutoDiscover 写入 module_code 后再加载）
	deps.ModuleCodeCache.Load(db)
	deps.AppPrefixMap.Load(db)

	// 注册租户隔离 GORM Callback（必须在 DataScope 之前注册）
	middleware.RegisterTenantIsolationCallback(db)

	// 注册数据权限 GORM Callback（含 OrganizationProvider）
	orgProvider := &spi.DefaultOrganizationProvider{DB: db}
	middleware.RegisterDataScopeCallback(db, cfg.DataScope.Enabled, orgProvider)

	// CR-10: 注册 ABAC GORM Callback（在 DataScope 之后，已在前面初始化 deps.AbacService）
	abac.RegisterAbacCallback(db, deps.AbacService)

	// 加载字段权限路由映射
	deps.FieldObjectRegistry.LoadRoutes(db)
	// 硬编码已知路由
	deps.FieldObjectRegistry.RegisterRoute("GET", "/api/v1/admin/users", "user")
	deps.FieldObjectRegistry.RegisterRoute("GET", "/api/v1/admin/users/:id", "user")
	deps.FieldObjectRegistry.RegisterRoute("GET", "/api/v1/admin/tenants", "tenant")
	deps.FieldObjectRegistry.RegisterRoute("GET", "/api/v1/admin/tenants/:id", "tenant")

	// CR-8: 初始化 PermCodeCache 并注册事件订阅
	var permCodeCache *authCache.PermCodeCache
	cacheAdapter := sdk.Runtime.GetCacheAdapter()
	if cacheAdapter != nil {
		permCodeCache = authCache.NewPermCodeCache(cacheAdapter, 10*time.Minute)
	}
	// 注册权限变更事件订阅（缓存失效）
	if permCodeCache != nil {
		event.DefaultBus.Subscribe(event.EventPermissionChanged, func(payload interface{}) {
			ev, ok := payload.(*event.PermissionChangedEvent)
			if !ok {
				return
			}
			for _, u := range ev.AffectedUsers {
				permCodeCache.Invalidate(u.TenantID, u.UserID)
			}
		})
	}
	// 将 permCodeCache 保存供中间件使用
	deps.PermCodeCache = permCodeCache

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
		&model.LoginLog{},
		&model.FieldObject{},
		&model.FieldDefinition{},
		&model.FieldPermission{},
		&model.RecordShare{},
		&model.OrgUnit{},
		&model.UserOrg{},
		&model.AdminConfig{},
		// CR5: C端用户表
		&userAuthModel.BizUser{},
		// CR6: 域名-租户映射表
		&model.TenantDomain{},
		// CR8: 自定义字段表（仅 DDL）
		&model.CustomField{},
		// CR10: ABAC 策略引擎表
		&abacmodel.AbacPolicy{},
		&abacmodel.AbacRowPolicy{},
		&abacmodel.AbacColPolicy{},
	)
}

// buildDependencies 创建所有 service/handler 依赖
func buildDependencies(cfg *config.Config, db *gorm.DB, blacklistStore ...spi.TokenBlacklistStore) *Dependencies {
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
	authSvc := service.NewAuthService(db, cfg, blacklistStore...)
	tenantSvc := service.NewTenantService(db, cfg, tenantRepo)
	userSvc := service.NewUserService(db, cfg, userRepo)
	roleSvc := service.NewRoleService(db, cfg, roleRepo)
	userRoleSvc := service.NewUserRoleService(db, cfg)
	apiPermSvc := service.NewApiPermissionService(db, apiPermRepo)
	resourceSvc := service.NewResourceService(db, cfg, resourceRepo)
	appSvc := service.NewApplicationService(db, appRepo)
	dataScopeSvc := service.NewDataScopeService(db, dataScopeConfigRepo, dataScopeRepo)
	opLogQuerySvc := service.NewOperationLogQueryService(db, operationLogRepo)
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

	// 域名-租户映射
	domainRepo := repository.NewTenantDomainRepository()
	tenantDomainCache := authCache.NewLocalCache()
	tenantDomainSvc := service.NewTenantDomainService(db, domainRepo, tenantRepo, tenantDomainCache)
	tenantDomainHandler := handler.NewTenantDomainHandler(tenantDomainSvc)

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

		TenantDomainService: tenantDomainSvc,
		TenantDomainHandler: tenantDomainHandler,

		AppPrefixMap:         middleware.NewAppPrefixMap(),
		ModuleCodeCache:      middleware.NewModuleCodeCache(),
		FieldObjectRegistry:  middleware.NewFieldObjectRegistry(),
		FieldPermissionRepo:  fieldPermRepo,
	}
}

// buildAbacDependencies 初始化 ABAC 模块依赖并注入到 deps
func buildAbacDependencies(db *gorm.DB, deps *Dependencies) {
	abacPolicyRepo := abacrepo.NewAbacPolicyRepository()
	abacRowRepo := abacrepo.NewAbacRowPolicyRepository()
	abacColRepo := abacrepo.NewAbacColPolicyRepository()
	cacheAdapter := sdk.Runtime.GetCacheAdapter() // storage.AdapterCache，可为 nil
	abacSvc := abacservice.NewAbacService(db, abacPolicyRepo, abacRowRepo, abacColRepo, cacheAdapter)
	deps.AbacService = abacSvc
	deps.AbacHandler = abachandler.NewAbacHandler(abacSvc)
}

// registerFieldPermModels 注册所有需要字段权限管控的 model
func registerFieldPermModels(registry *service.FieldRegistry) {
	registry.AutoRegister("user", "用户", model.User{})
	registry.AutoRegister("tenant", "租户", model.Tenant{})
}

// initUserAuth 初始化 user_auth 域：实例化依赖、注册路由、注册 SmsStrategy
func initUserAuth(cfg *config.Config, deps *Dependencies, engine *gin.Engine) {
	db := deps.DB

	// 实例化 user_auth 域依赖
	smsSender := &userAuthSpi.ConsoleMockSender{}
	// 有 Redis 时使用 Redis CodeStore（多 Pod 安全），否则降级到内存
	var codeStore userAuthService.CodeStore
	if redisAdapter := sdk.Runtime.GetCacheAdapter(); redisAdapter != nil {
		codeStore = userAuthService.NewRedisCodeStore(redisAdapter)
		log.Println("[user_auth] SMS CodeStore 使用 Redis（多 Pod 安全）")
	} else {
		codeStore = userAuthService.NewMemoryCodeStore()
		log.Println("[user_auth] SMS CodeStore 降级为内存（单实例模式）")
	}
	smsSvc := userAuthService.NewSmsService(smsSender, codeStore)
	bizUserRepo := userAuthRepo.NewBizUserRepository()
	bizUserSvc := userAuthService.NewBizUserService(db, bizUserRepo)
	menuSvc := userAuthService.NewUserMenuService(db)
	smsStrategy := userStrategy.NewSmsStrategy(smsSvc, bizUserSvc, bizUserRepo, cfg, db)

	// 实例化 Handler
	uaHandler := userAuthHandler.NewUserAuthHandler(smsSvc, bizUserSvc, menuSvc, cfg, db)
	bizHandler := userAuthHandler.NewBizUserHandler(bizUserSvc)

	// C端公开路由（无需认证）
	publicGroup := engine.Group("/api/v1/user/auth")

	// C端认证路由（需 auth 中间件 + C端用户检查）
	userGroup := engine.Group("/api/v1/user/auth")
	userGroup.Use(middleware.AuthMiddleware(deps.AuthService, bizUserSvc))

	// 管理端路由（需 auth + 应用解析 + 动态权限中间件）
	adminGroup := engine.Group("/api/v1/admin")
	adminGroup.Use(middleware.AuthMiddleware(deps.AuthService))
	if deps.AppPrefixMap != nil && deps.ModuleCodeCache != nil {
		adminGroup.Use(middleware.AppResolveMiddleware(deps.AppPrefixMap, deps.ModuleCodeCache, db))
	}
	adminGroup.Use(middleware.DynamicPermissionMiddleware(db, cfg, deps.AdminConfigService))

	// 注册路由
	userAuth.RegisterRoutes(publicGroup, userGroup, adminGroup, uaHandler, bizHandler)

	// 注册 SmsStrategy 到 AuthHandler 的 StrategyRouter
	userAuth.RegisterSmsStrategy(deps.AuthHandler.GetStrategyRouter(), smsStrategy)

	log.Println("[user_auth] C端认证域路由注册完成")
}
