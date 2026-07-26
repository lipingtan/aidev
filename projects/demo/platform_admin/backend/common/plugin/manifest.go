package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FrontendConfig 插件前端 bundle 配置（V2 多端支持）
type FrontendConfig struct {
	Platform string `json:"platform"` // "admin" 或 "user"
	Device   string `json:"device"`   // "pc" 或 "h5"
	Entry    string `json:"entry"`    // 相对路径，如 "admin-pc/bundle.js"
}

// ManifestV2 plugin.json 完整结构（V2 Manifest）
type ManifestV2 struct {
	Name             string                  `json:"name"`
	Version          string                  `json:"version"`
	DisplayName      string                  `json:"displayName"`
	Description      string                  `json:"description"`
	RoutePrefix      string                  `json:"routePrefix"`
	Platforms        []string                `json:"platforms"`
	Modules          []ManifestModule        `json:"modules"`
	Frontends        []FrontendConfig        `json:"frontends"`
	Menus            []ManifestMenu          `json:"menus"`
	ApiPermissions   []ManifestApiPermission `json:"apiPermissions"`
	ExposedActions   []ManifestAction        `json:"exposedActions"`
	SubscribedEvents []string                `json:"subscribedEvents"`
	BreakingUpgrade  bool                    `json:"breakingUpgrade"`
	MinUpgradeFrom   string                  `json:"minUpgradeFrom"`
	MigrationNotes   string                  `json:"migrationNotes"`
}

// ManifestModule 功能模块声明
type ManifestModule struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// ManifestMenu 菜单/按钮资源声明
type ManifestMenu struct {
	Type           string         `json:"type"`           // menu/button/page
	Name           string         `json:"name"`           // 资源名称
	PermissionCode string         `json:"permissionCode"` // 权限标识码
	Path           string         `json:"path"`           // 路由路径
	Icon           string         `json:"icon"`           // 图标
	Component      string         `json:"component"`      // 前端组件路径
	Platform       string         `json:"platform"`       // admin/user
	ModuleCode     string         `json:"moduleCode"`     // 所属功能模块
	Sort           int            `json:"sort"`           // 排序号
	Children       []ManifestMenu `json:"children"`       // 子菜单/按钮
}

// ManifestApiPermission API 权限声明（树形：GROUP + ENDPOINT）
type ManifestApiPermission struct {
	Type           string                  `json:"type"`           // GROUP/ENDPOINT
	Name           string                  `json:"name"`           // 接口名称/分组名
	DisplayName    string                  `json:"displayName"`    // 中文显示名称
	PermissionCode string                  `json:"permissionCode"` // 权限标识码
	URLPattern     string                  `json:"urlPattern"`     // URL 匹配模式
	HTTPMethod     string                  `json:"httpMethod"`     // HTTP 方法
	ModuleCode     string                  `json:"moduleCode"`     // 所属功能模块
	Children       []ManifestApiPermission `json:"children"`       // 子节点（ENDPOINT）
}

// ManifestAction 插件暴露的可调用 Action
type ManifestAction struct {
	Name         string `json:"name"`         // Action 名称
	Version      string `json:"version"`      // 语义化版本，如 "1.0"
	InputSchema  string `json:"inputSchema"`  // 输入 JSON Schema
	OutputSchema string `json:"outputSchema"` // 输出 JSON Schema
}

// ParseManifest 从插件目录中读取并解析 plugin.json（V2 格式）
func ParseManifest(pluginDir string) (*ManifestV2, error) {
	path := filepath.Join(pluginDir, "plugin.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 plugin.json 失败: %w", err)
	}

	var manifest ManifestV2
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("解析 plugin.json 失败: %w", err)
	}

	if manifest.Name == "" {
		return nil, fmt.Errorf("plugin.json 中 name 字段不能为空")
	}

	// displayName 缺失时 fallback 到 name
	if manifest.DisplayName == "" {
		manifest.DisplayName = manifest.Name
	}

	return &manifest, nil
}
