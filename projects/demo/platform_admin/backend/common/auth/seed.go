package auth

import (
	"go-admin/common/auth/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedInitialData 创建初始数据：超级管理员 + 默认租户 + SUPER_ADMIN 角色 + 关联
// 仅在 admin_user 表无记录时执行（幂等）
func SeedInitialData(db *gorm.DB) error {
	// 幂等检查
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // 已有数据，跳过
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

	// 2. 创建默认租户
	tenant := &model.Tenant{
		TenantCode: "default",
		Name:       "默认租户",
		Status:     1,
		Version:    1,
	}
	if err := db.Create(tenant).Error; err != nil {
		return err
	}

	// 3. 关联 admin → default 租户
	ut := &model.UserTenant{
		UserID:   admin.ID,
		TenantID: tenant.ID,
	}
	if err := db.Create(ut).Error; err != nil {
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
	userRole := &model.UserRole{
		UserID:   admin.ID,
		RoleID:   role.ID,
		TenantID: tenant.ID,
	}
	if err := db.Create(userRole).Error; err != nil {
		return err
	}

	// 6. 创建默认应用 admin
	app := &model.Application{
		AppCode:     "admin",
		Name:        "平台管理",
		Description: "平台管理应用",
		Status:      1,
		Version:     1,
	}
	if err := db.Create(app).Error; err != nil {
		return err
	}

	// 7. 角色绑定应用
	roleApp := &model.RoleApp{RoleID: role.ID, AppCode: "admin"}
	if err := db.Create(roleApp).Error; err != nil {
		return err
	}

	// 8. 租户订阅应用
	tenantApp := &model.TenantApp{TenantID: tenant.ID, AppCode: "admin"}
	if err := db.Create(tenantApp).Error; err != nil {
		return err
	}

	// 9. 创建菜单资源（对齐前端 static-routes.ts）
	resources := []*model.Resource{
		{Type: "MENU", Name: "首页", Path: "/home", Icon: "HomeFilled", AppCode: "admin", SortOrder: 0, Status: 1, Version: 1},
		{Type: "MENU", Name: "系统管理", Path: "/system", Icon: "Setting", AppCode: "admin", SortOrder: 10, Status: 1, Version: 1},
		{Type: "MENU", Name: "日志管理", Path: "/log", Icon: "Document", AppCode: "admin", SortOrder: 20, Status: 1, Version: 1},
		{Type: "MENU", Name: "监控", Path: "/monitor", Icon: "Monitor", AppCode: "admin", SortOrder: 30, Status: 1, Version: 1},
	}
	if err := db.Create(&resources).Error; err != nil {
		return err
	}

	// 系统管理子菜单
	sysParentID := resources[1].ID
	sysChildren := []*model.Resource{
		{ParentID: &sysParentID, Type: "MENU", Name: "用户管理", Path: "/system/users", Icon: "User", PermissionCode: "user:user:list", AppCode: "admin", SortOrder: 0, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "角色管理", Path: "/system/roles", Icon: "UserFilled", PermissionCode: "user:role:list", AppCode: "admin", SortOrder: 1, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "租户管理", Path: "/system/tenants", Icon: "Tickets", PermissionCode: "system:tenant:list", AppCode: "admin", SortOrder: 3, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "应用管理", Path: "/system/applications", Icon: "Box", PermissionCode: "system:app:list", AppCode: "admin", SortOrder: 4, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "数据权限配置", Path: "/system/data-scope", Icon: "Key", PermissionCode: "system:data-scope:list", AppCode: "admin", SortOrder: 5, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "系统配置", Path: "/system/config", Icon: "Setting", PermissionCode: "system:config:list", AppCode: "admin", SortOrder: 6, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "接口管理", Path: "/system/api", Icon: "Connection", PermissionCode: "system:api:list", AppCode: "admin", SortOrder: 7, Status: 1, Version: 1},
		{ParentID: &sysParentID, Type: "MENU", Name: "插件管理", Path: "/system/plugins", Icon: "Box", PermissionCode: "system:plugin:list", AppCode: "admin", SortOrder: 8, Status: 1, Version: 1},
	}
	if err := db.Create(&sysChildren).Error; err != nil {
		return err
	}

	// 日志管理子菜单
	logParentID := resources[2].ID
	logChildren := []*model.Resource{
		{ParentID: &logParentID, Type: "MENU", Name: "登录日志", Path: "/log/login-logs", Icon: "Document", PermissionCode: "user:log:list", AppCode: "admin", SortOrder: 0, Status: 1, Version: 1},
		{ParentID: &logParentID, Type: "MENU", Name: "操作日志", Path: "/log/operation-logs", Icon: "Notebook", PermissionCode: "user:log:list", AppCode: "admin", SortOrder: 1, Status: 1, Version: 1},
	}
	if err := db.Create(&logChildren).Error; err != nil {
		return err
	}

	// 监控子菜单
	monParentID := resources[3].ID
	monChildren := []*model.Resource{
		{ParentID: &monParentID, Type: "MENU", Name: "服务监控", Path: "/monitor/server", Icon: "Monitor", PermissionCode: "monitor:server", AppCode: "admin", SortOrder: 0, Status: 1, Version: 1},
	}
	if err := db.Create(&monChildren).Error; err != nil {
		return err
	}

	return nil
}
