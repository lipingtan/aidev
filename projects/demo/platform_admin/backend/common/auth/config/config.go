package config

import "time"

// Config auth-rbac 模块完整配置
type Config struct {
	Enabled      bool            `yaml:"enabled"`
	AuthType     string          `yaml:"auth-type"`
	CacheType    string          `yaml:"cache-type"`
	IDStrategy   string          `yaml:"id-strategy"`
	JWT          JWTConfig       `yaml:"jwt"`
	Permission   PermConfig      `yaml:"permission"`
	Role         RoleConfig      `yaml:"role"`
	DataScope    DataScopeConfig `yaml:"data-scope"`
	APIDiscovery APIDiscovery    `yaml:"api-discovery"`
	App          AppConfig       `yaml:"app"`
	Tenant       TenantConfig    `yaml:"tenant"`
	Blacklist    BlacklistConfig `yaml:"blacklist"`
	Redis        RedisConfig     `yaml:"redis"`
}

// JWTConfig JWT 相关配置
type JWTConfig struct {
	Secret             string        `yaml:"secret"`
	PlatformTokenTTL   time.Duration `yaml:"platform-token-ttl"`
	AccessTokenTTL     time.Duration `yaml:"access-token-ttl"`
	UserAccessTokenTTL time.Duration `yaml:"user-access-token-ttl"` // C端 token 有效期，默认 7 天
	Issuer             string        `yaml:"issuer"`
}

// PermConfig 权限相关配置
type PermConfig struct {
	WildcardEnabled     bool   `yaml:"wildcard-enabled"`
	SuperAdminRole      string `yaml:"super-admin-role"`
	EnforcementStrategy string `yaml:"enforcement-strategy"`
}

// RoleConfig 角色相关配置
type RoleConfig struct {
	HierarchyEnabled bool `yaml:"hierarchy-enabled"`
	MaxDepth         int  `yaml:"max-depth"`
}

// DataScopeConfig 数据权限配置
type DataScopeConfig struct {
	Enabled    bool `yaml:"enabled"`
	AutoInject bool `yaml:"auto-inject"`
}

// APIDiscovery API 自动发现配置
type APIDiscovery struct {
	Enabled        bool `yaml:"enabled"`
	AsyncThreshold int  `yaml:"async-threshold"`
}

// AppConfig 应用配置
type AppConfig struct {
	DefaultAppCode string `yaml:"default-app-code"`
}

// TemplateRole 模板角色定义
type TemplateRole struct {
	Code string `yaml:"code"`
	Name string `yaml:"name"`
}

// TenantConfig 租户配置
type TenantConfig struct {
	TemplateRoles []TemplateRole `yaml:"template-roles"`
	DefaultApps   []string       `yaml:"default-apps"`
}

// BlacklistConfig 黑名单配置
type BlacklistConfig struct {
	StoreType string `yaml:"store-type"`
}

// RedisConfig Redis 连接配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// DefaultConfig 返回带默认值的配置
func DefaultConfig() *Config {
	return &Config{
		Enabled:    true,
		AuthType:   "jwt",
		CacheType:  "local",
		IDStrategy: "snowflake",
		JWT: JWTConfig{
			Secret:             "change-me-in-production",
			PlatformTokenTTL:   168 * time.Hour,
			AccessTokenTTL:     2 * time.Hour,
			UserAccessTokenTTL: 7 * 24 * time.Hour,
			Issuer:             "auth-rbac",
		},
		Permission: PermConfig{
			WildcardEnabled:     true,
			SuperAdminRole:      "SUPER_ADMIN",
			EnforcementStrategy: "annotation-first",
		},
		Role: RoleConfig{
			HierarchyEnabled: true,
			MaxDepth:         5,
		},
		DataScope: DataScopeConfig{
			Enabled:    false,
			AutoInject: true,
		},
		APIDiscovery: APIDiscovery{
			Enabled:        true,
			AsyncThreshold: 500,
		},
		App: AppConfig{
			DefaultAppCode: "default",
		},
		Tenant: TenantConfig{
			TemplateRoles: []TemplateRole{
				{Code: "tenant_admin", Name: "租户管理员"},
				{Code: "editor", Name: "编辑者"},
				{Code: "viewer", Name: "查看者"},
			},
			DefaultApps: []string{"default"},
		},
		Blacklist: BlacklistConfig{
			StoreType: "local",
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}
}
