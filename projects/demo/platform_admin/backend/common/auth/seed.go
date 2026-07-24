package auth

import (
	"go-admin/common/auth/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedInitialData 创建初始数据：超级管理员 + 默认租户 + SUPER_ADMIN 角色 + 应用 + 菜单
// 仅在 admin_user 表无记录时执行（幂等）
func SeedInitialData(db *gorm.DB) error {
	// 幂等检查
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// 1. 创建超级管理员
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	admin := &model.User{
		Username: "admin",
		Password: string(hashedPwd),
		Nickname: "超级管理员",
		Status:   1,
		Version:  1,
	}
	if err := db.Create(admin).Error; err != nil {
		return err
	}

	// 2. 创建默认租户（含 timezone/locale/currency）
	tenant := &model.Tenant{
		TenantCode: "default",
		Name:       "默认租户",
		Status:     1,
		Timezone:   "Asia/Shanghai",
		Locale:     "zh-CN",
		Currency:   "CNY",
		Version:    1,
	}
	if err := db.Create(tenant).Error; err != nil {
		return err
	}

	// 3. 关联 admin → default 租户
	if err := db.Create(&model.UserTenant{UserID: admin.ID, TenantID: tenant.ID}).Error; err != nil {
		return err
	}

	// 4. 创建 SUPER_ADMIN 角色
	role := &model.Role{
		TenantID: tenant.ID,
		RoleCode: "SUPER_ADMIN",
		RoleName: "超级管理员",
		RoleType: "SUPER_ADMIN",
		Status:   1,
		Version:  1,
	}
	if err := db.Create(role).Error; err != nil {
		return err
	}

	// 5. 分配角色
	if err := db.Create(&model.UserRole{UserID: admin.ID, RoleID: role.ID, TenantID: tenant.ID}).Error; err != nil {
		return err
	}

	// 6. 创建内置应用 platform_admin（含 app_type/route_prefix/platforms）
	app := &model.Application{
		AppCode:     "platform_admin",
		Name:        "平台管理",
		Description: "平台管理后台应用",
		AppType:     "BUILTIN",
		RoutePrefix: "/api/v1/admin",
		Platforms:   []byte(`["admin:pc","admin:h5"]`),
		Status:      1,
		Version:     1,
	}
	if err := db.Create(app).Error; err != nil {
		return err
	}

	// 7. 角色绑定应用
	if err := db.Create(&model.RoleApp{RoleID: role.ID, AppCode: "platform_admin"}).Error; err != nil {
		return err
	}

	// 8. 租户订阅应用
	if err := db.Create(&model.TenantApp{TenantID: tenant.ID, AppCode: "platform_admin"}).Error; err != nil {
		return err
	}

	// 9. 创建管理端菜单（全部 platform=admin, app_code=platform_admin）
	if err := seedMenus(db); err != nil {
		return err
	}

	return nil
}

// seedMenus 创建管理端菜单资源
func seedMenus(db *gorm.DB) error {
	// 一级菜单
	resources := []*model.Resource{
		{Type: "MENU", Name: "首页", Path: "/home", Icon: "HomeFilled", AppCode: "platform_admin", Platform: "admin", SortOrder: 0, Status: 1, Version: 1},
		{Type: "MENU", Name: "系统管理", Path: "/system", Icon: "Setting", AppCode: "platform_admin", Platform: "admin", ModuleCode: "system", SortOrder: 10, Status: 1, Version: 1},
		{Type: "MENU", Name: "日志管理", Path: "/log", Icon: "Document", AppCode: "platform_admin", Platform: "admin", ModuleCode: "log-mgmt", SortOrder: 20, Status: 1, Version: 1},
		{Type: "MENU", Name: "监控", Path: "/monitor", Icon: "Monitor", AppCode: "platform_admin", Platform: "admin", ModuleCode: "monitor", SortOrder: 30, Status: 1, Version: 1},
	}
	if err := db.Create(&resources).Error; err != nil {
		return err
	}

	// 系统管理子菜单
	sysParentID := resources[1].ID
	sysChildren := []*model.Resource{
		{ParentID: &sysParentID, Type: "MENU", Name: "用户管理", Path: "/system/users", Icon: "User", PermissionCode: "user:user:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "user-mgmt", SortOrder: 0, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "角色管理", Path: "/system/roles", Icon: "UserFilled", PermissionCode: "user:role:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "role-mgmt", SortOrder: 1, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "租户管理", Path: "/system/tenants", Icon: "Tickets", PermissionCode: "system:tenant:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "tenant-mgmt", SortOrder: 3, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "应用管理", Path: "/system/applications", Icon: "Box", PermissionCode: "system:app:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "app-mgmt", SortOrder: 4, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "数据权限配置", Path: "/system/data-scope", Icon: "Key", PermissionCode: "system:data-scope:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "data-scope-mgmt", SortOrder: 5, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "系统配置", Path: "/system/config", Icon: "Setting", PermissionCode: "system:config:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "config-mgmt", SortOrder: 6, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "接口管理", Path: "/system/api", Icon: "Connection", PermissionCode: "system:api:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "api-perm-mgmt", SortOrder: 7, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "插件管理", Path: "/system/plugins", Icon: "Box", PermissionCode: "system:plugin:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "plugin-mgmt", SortOrder: 8, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "字段对象管理", Path: "/system/field-objects", Icon: "Grid", PermissionCode: "system:field-object:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "field-perm-mgmt", SortOrder: 9, Status: 1, Version: 1},
	}
	if err := db.Create(&sysChildren).Error; err != nil {
		return err
	}

	// 日志管理子菜单
	logParentID := resources[2].ID
	logChildren := []*model.Resource{
		{ParentID: &logParentID, Type: "MENU", Name: "登录日志", Path: "/log/login-logs", Icon: "Document", PermissionCode: "user:log:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "log-mgmt", SortOrder: 0, Status: 1, Version: 1},
		{ParentID: &logParentID, Type: "MENU", Name: "操作日志", Path: "/log/operation-logs", Icon: "Notebook", PermissionCode: "user:log:list", AppCode: "platform_admin", Platform: "admin", ModuleCode: "log-mgmt", SortOrder: 1, Status: 1, Version: 1},
	}
	if err := db.Create(&logChildren).Error; err != nil {
		return err
	}

	// 监控子菜单
	monParentID := resources[3].ID
	monChildren := []*model.Resource{
		{ParentID: &monParentID, Type: "MENU", Name: "服务监控", Path: "/monitor/server", Icon: "Monitor", PermissionCode: "monitor:server", AppCode: "platform_admin", Platform: "admin", ModuleCode: "monitor", SortOrder: 0, Status: 1, Version: 1},
	}
	if err := db.Create(&monChildren).Error; err != nil {
		return err
	}

	return nil
}

// SeedCR5Menus 补充 CR-5 新增菜单：C端用户管理（幂等，可多次执行）
func SeedCR5Menus(db *gorm.DB) error {
	var count int64
	db.Model(&model.Resource{}).Where("path = ?", "/system/biz-user").Count(&count)
	if count > 0 {
		return nil // 已存在，幂等跳过
	}
	var sysParent model.Resource
	if err := db.Where("path = ? AND app_code = ?", "/system", "platform_admin").First(&sysParent).Error; err != nil {
		return nil // 系统管理父菜单不存在时跳过
	}
	sysID := sysParent.ID
	menu := &model.Resource{
		ParentID:       &sysID,
		Type:           "MENU",
		Name:           "C端用户管理",
		Path:           "/system/biz-user",
		Icon:           "Avatar",
		PermissionCode: "biz:user:list",
		AppCode:        "platform_admin",
		Platform:       "admin",
		ModuleCode:     "biz-user-mgmt",
		SortOrder:      10,
		Status:         1,
		Version:        1,
	}
	return db.Create(menu).Error
}

// SeedCR2Menus 补充 CR-2 新增菜单（幂等，可多次执行）
func SeedCR2Menus(db *gorm.DB) error {
	// 字段对象管理菜单（系统管理下）
	var fieldObjCount int64
	db.Model(&model.Resource{}).Where("path = ?", "/system/field-objects").Count(&fieldObjCount)
	if fieldObjCount == 0 {
		var sysParent model.Resource
		if err := db.Where("path = ? AND app_code = ?", "/system", "platform_admin").First(&sysParent).Error; err == nil {
			sysID := sysParent.ID
			fieldObjMenu := &model.Resource{
				ParentID: &sysID, Type: "MENU", Name: "字段对象管理", Path: "/system/field-objects",
				Icon: "Grid", PermissionCode: "system:field-object:list",
				AppCode: "platform_admin", Platform: "admin", ModuleCode: "field-perm-mgmt",
				SortOrder: 9, Status: 1, Version: 1,
			}
			db.Create(fieldObjMenu)
		}
	}

	// 清理旧的"权限管理/权限演示"菜单组（如果存在）
	var permParent model.Resource
	if err := db.Where("path = ? AND app_code = ?", "/permission", "platform_admin").First(&permParent).Error; err == nil {
		// 删除子菜单
		db.Where("parent_id = ?", permParent.ID).Delete(&model.Resource{})
		// 删除父菜单
		db.Delete(&permParent)
	}

	return nil
}

// SeedConfigDefaults 创建默认配额和功能开关种子数据（幂等）
func SeedConfigDefaults(db *gorm.DB) error {
	defaults := []model.AdminConfig{
		{ConfigKey: "quota.max_admin_users", ConfigValue: "999999", ConfigType: "number", Scope: "SYSTEM", ScopeID: 0, TenantID: 0, DisplayName: "最大管理端用户数", Description: "租户最大管理端用户数配额，999999=不限制", Status: 1},
		{ConfigKey: "quota.max_roles", ConfigValue: "999999", ConfigType: "number", Scope: "SYSTEM", ScopeID: 0, TenantID: 0, DisplayName: "最大角色数", Description: "租户最大角色数配额，999999=不限制", Status: 1},
		{ConfigKey: "quota.max_apps", ConfigValue: "999999", ConfigType: "number", Scope: "SYSTEM", ScopeID: 0, TenantID: 0, DisplayName: "最大可订阅应用数", Description: "租户最大可订阅应用数配额，999999=不限制", Status: 1},
	}

	for _, cfg := range defaults {
		var count int64
		db.Model(&model.AdminConfig{}).
			Where("config_key = ? AND scope = 'SYSTEM' AND scope_id = 0 AND tenant_id = 0", cfg.ConfigKey).
			Count(&count)
		if count > 0 {
			continue
		}
		db.Create(&cfg)
	}
	return nil
}

// SeedCR6Menus 补充 CR-6 新增菜单：域名管理（幂等，可多次执行）
func SeedCR6Menus(db *gorm.DB) error {
	var count int64
	db.Model(&model.Resource{}).Where("path = ?", "/system/tenant-domain").Count(&count)
	if count > 0 {
		return nil // 已存在，幂等跳过
	}
	var sysParent model.Resource
	if err := db.Where("path = ? AND app_code = ?", "/system", "platform_admin").First(&sysParent).Error; err != nil {
		return nil // 系统管理父菜单不存在时跳过
	}
	sysID := sysParent.ID
	menu := &model.Resource{
		ParentID:       &sysID,
		Type:           "MENU",
		Name:           "域名管理",
		Path:           "/system/tenant-domain",
		Icon:           "Link",
		PermissionCode: "system:tenant-domain:list",
		AppCode:        "platform_admin",
		Platform:       "admin",
		ModuleCode:     "tenant-domain-mgmt",
		SortOrder:      11,
		Status:         1,
		Version:        1,
	}
	return db.Create(menu).Error
}
