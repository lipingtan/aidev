package shared

// Manifest 插件描述文件结构
type Manifest struct {
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	DisplayName   string            `json:"displayName"`
	Description   string            `json:"description"`
	Dependencies  map[string]string `json:"dependencies"`  // name -> semver
	Platforms     []string          `json:"platforms"`      // admin/user/both
	FrontendEntry string            `json:"frontendEntry"`  // 前端入口路径
	Menus         []MenuEntry       `json:"menus"`
	Routes        []RouteEntry      `json:"routes"`
	Migrations    []string          `json:"migrations"`     // 迁移文件列表
}

// Permission 权限声明
type Permission struct {
	Code        string `json:"code"`
	DisplayName string `json:"displayName"`
	Group       string `json:"group"`
}

// MenuEntry 菜单项声明
type MenuEntry struct {
	Title      string `json:"title"`
	Icon       string `json:"icon"`
	Path       string `json:"path"`
	Sort       int    `json:"sort"`
	Group      string `json:"group"`
	Permission string `json:"permission"`
}

// RouteEntry 路由声明
type RouteEntry struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Component string    `json:"component"`
	Meta      RouteMeta `json:"meta"`
}

// RouteMeta 路由元数据
type RouteMeta struct {
	Title      string `json:"title"`
	Icon       string `json:"icon,omitempty"`
	Permission string `json:"permission,omitempty"`
	Hidden     bool   `json:"hidden,omitempty"`
}
