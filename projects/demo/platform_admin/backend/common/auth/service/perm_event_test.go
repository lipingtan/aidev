package service

import (
	"sync"
	"testing"
	"time"

	"go-admin/common/auth/config"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"
	"go-admin/common/event"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupPermEventTestDB 创建包含角色/用户角色/API权限相关表的测试数据库
func setupPermEventTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Role{},
		&model.UserRole{},
		&model.UserTenant{},
		&model.RoleResource{},
		&model.RoleApi{},
		&model.ApiPermission{},
		&model.Resource{},
		&model.TenantApp{},
		&model.RoleApp{},
	); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	return db
}

// capturePermChangedEvents 订阅 PermissionChanged 事件，返回捕获的事件列表
func capturePermChangedEvents(bus event.Bus) *[]event.PermissionChangedEvent {
	var mu sync.Mutex
	var events []event.PermissionChangedEvent

	bus.Subscribe(event.EventPermissionChanged, func(payload interface{}) {
		if ev, ok := payload.(*event.PermissionChangedEvent); ok {
			mu.Lock()
			events = append(events, *ev)
			mu.Unlock()
		}
	})
	return &events
}

// TestUserRoleService_AssignRoles_PublishesEvent 分配角色后发布权限变更事件
func TestUserRoleService_AssignRoles_PublishesEvent(t *testing.T) {
	db := setupPermEventTestDB(t)

	// 创建测试数据
	tenant := &model.Tenant{TenantCode: "t1", Name: "租户1", Status: 1}
	db.Create(tenant)

	user := &model.User{Username: "user1"}
	db.Create(user)

	role := &model.Role{TenantID: tenant.ID, RoleCode: "ROLE1", RoleName: "角色1", Status: 1, Version: 1}
	db.Create(role)

	// 关联用户到租户
	db.Create(&model.UserTenant{UserID: user.ID, TenantID: tenant.ID})

	// 创建自定义 EventBus 并订阅
	bus := event.NewBus()
	capturedEvents := capturePermChangedEvents(bus)

	// 创建 service（使用自定义 bus）
	svc := NewUserRoleService(db, config.DefaultConfig())

	// 替换 AssignRoles 中对 DefaultBus 的依赖（需要验证事件发布机制）
	// 由于当前实现使用 event.DefaultBus，我们订阅 DefaultBus
	defaultCaptured := capturePermChangedEvents(event.DefaultBus)

	err := svc.AssignRoles(user.ID, &AssignRolesRequest{
		TenantID: tenant.ID,
		Roles:    []RoleAssignment{{RoleID: role.ID}},
	})
	if err != nil {
		t.Fatalf("AssignRoles 失败: %v", err)
	}

	// 等待异步事件处理
	time.Sleep(50 * time.Millisecond)

	// 验证事件已发布
	if len(*defaultCaptured) == 0 {
		t.Fatal("期望 AssignRoles 发布 PermissionChanged 事件，实际未收到")
	}

	// 验证 payload 包含正确的 userID 和 tenantID
	found := false
	for _, ev := range *defaultCaptured {
		for _, u := range ev.AffectedUsers {
			if u.UserID == user.ID && u.TenantID == tenant.ID {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("事件 payload 未包含 userID=%d tenantID=%d", user.ID, tenant.ID)
	}

	// bus 变量已用于其他测试，避免编译警告
	_ = bus
	_ = capturedEvents
}

// TestUserRoleService_ReplaceRoles_PublishesEvent 全量替换角色后发布权限变更事件
func TestUserRoleService_ReplaceRoles_PublishesEvent(t *testing.T) {
	db := setupPermEventTestDB(t)

	// 准备数据
	tenant := &model.Tenant{TenantCode: "t2", Name: "租户2", Status: 1}
	db.Create(tenant)

	user := &model.User{Username: "user2"}
	db.Create(user)

	role := &model.Role{TenantID: tenant.ID, RoleCode: "ROLE2", RoleName: "角色2", Status: 1, Version: 1}
	db.Create(role)

	db.Create(&model.UserTenant{UserID: user.ID, TenantID: tenant.ID})

	// 先分配一个角色
	db.Create(&model.UserRole{UserID: user.ID, RoleID: role.ID, TenantID: tenant.ID})

	svc := NewUserRoleService(db, config.DefaultConfig())

	// 订阅事件
	captured := capturePermChangedEvents(event.DefaultBus)

	err := svc.ReplaceRoles(user.ID, &AssignRolesRequest{
		TenantID: tenant.ID,
		Roles:    []RoleAssignment{{RoleID: role.ID}},
	})
	if err != nil {
		t.Fatalf("ReplaceRoles 失败: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if len(*captured) == 0 {
		t.Fatal("期望 ReplaceRoles 发布 PermissionChanged 事件，实际未收到")
	}

	found := false
	for _, ev := range *captured {
		for _, u := range ev.AffectedUsers {
			if u.UserID == user.ID {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("事件未包含受影响用户 userID=%d", user.ID)
	}

	// 验证 DB 中角色绑定已替换（副作用验证）
	var count int64
	db.Model(&model.UserRole{}).Where("user_id = ? AND tenant_id = ?", user.ID, tenant.ID).Count(&count)
	if count != 1 {
		t.Fatalf("期望 1 条角色绑定，实际 %d 条", count)
	}
}

// TestRoleService_AssignResources_PublishesEvent 分配资源权限后发布事件
func TestRoleService_AssignResources_PublishesEvent(t *testing.T) {
	db := setupPermEventTestDB(t)

	// 准备数据
	tenant := &model.Tenant{TenantCode: "t3", Name: "租户3", Status: 1}
	db.Create(tenant)

	user := &model.User{Username: "user3"}
	db.Create(user)

	role := &model.Role{TenantID: tenant.ID, RoleCode: "ROLE3", RoleName: "角色3", Status: 1, Version: 1}
	db.Create(role)

	// 关联用户到角色
	db.Create(&model.UserRole{UserID: user.ID, RoleID: role.ID, TenantID: tenant.ID})

	roleRepo := repository.NewRoleRepository()
	svc := NewRoleService(db, config.DefaultConfig(), roleRepo)

	captured := capturePermChangedEvents(event.DefaultBus)

	_, err := svc.AssignResources(role.ID, &AssignResourcesRequest{
		ResourceIDs: []int64{},
	})
	if err != nil {
		t.Fatalf("AssignResources 失败: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// 有用户绑定时才发事件；此处 role 有绑定 user3，应该发布
	if len(*captured) == 0 {
		t.Fatal("期望 AssignResources 发布 PermissionChanged 事件，实际未收到")
	}

	found := false
	for _, ev := range *captured {
		for _, u := range ev.AffectedUsers {
			if u.UserID == user.ID {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("事件未包含受影响用户 userID=%d", user.ID)
	}
}
