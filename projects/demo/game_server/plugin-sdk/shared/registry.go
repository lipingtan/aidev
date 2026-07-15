package shared

import "context"

// PluginRecord 插件数据库记录
type PluginRecord struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
	Status       int    `json:"status"` // 0=installed, 1=running, 2=stopped, 3=error
	FrontendPath string `json:"frontendPath"`
	Config       string `json:"config"`
}

// PluginVersionRecord 插件版本记录
type PluginVersionRecord struct {
	PluginName   string `json:"pluginName"`
	Version      string `json:"version"`
	SnapshotPath string `json:"snapshotPath"`
	MigrationVer int    `json:"migrationVer"`
}

// TenantPluginRecord 租户插件关联记录
type TenantPluginRecord struct {
	TenantId   int    `json:"tenantId"`
	PluginName string `json:"pluginName"`
	Enabled    int    `json:"enabled"`
}

// PluginRegistry 插件注册表接口 - 数据库操作抽象
type PluginRegistry interface {
	// 插件 CRUD
	GetByName(ctx context.Context, name string) (*PluginRecord, error)
	List(ctx context.Context) ([]PluginRecord, error)
	UpdateStatus(ctx context.Context, name string, status int) error

	// 版本管理
	CreateVersion(ctx context.Context, record PluginVersionRecord) error
	ListVersions(ctx context.Context, pluginName string) ([]PluginVersionRecord, error)
	GetLatestVersion(ctx context.Context, pluginName string) (*PluginVersionRecord, error)

	// 租户插件
	ListByTenant(ctx context.Context, tenantId int) ([]TenantPluginRecord, error)
	SetTenantPlugin(ctx context.Context, tenantId int, pluginName string, enabled int) error

	// 权限
	RegisterPermissions(ctx context.Context, pluginName string, permissions []Permission) error
	UnregisterPermissions(ctx context.Context, pluginName string) error
	ListPermissions(ctx context.Context, pluginName string) ([]Permission, error)
}
