package plugin

import (
	"strings"
	"sync"

	"platform-admin/plugin-sdk/proto"
)

// Registry 插件路由/菜单/权限注册表
type Registry struct {
	mu     sync.RWMutex
	routes map[string]string               // routePrefix → pluginName
	menus  map[string][]*proto.MenuItem    // pluginName → menus
	perms  map[string][]*proto.Permission  // pluginName → permissions
}

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		routes: make(map[string]string),
		menus:  make(map[string][]*proto.MenuItem),
		perms:  make(map[string][]*proto.Permission),
	}
}

// RegisterPlugin 注册插件的路由/菜单/权限
func (r *Registry) RegisterPlugin(info *proto.PluginInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if info.RoutePrefix != "" {
		r.routes[info.RoutePrefix] = info.Name
	}
	if len(info.Menus) > 0 {
		r.menus[info.Name] = info.Menus
	}
	if len(info.Perms) > 0 {
		r.perms[info.Name] = info.Perms
	}
}

// UnregisterPlugin 注销插件的路由/菜单/权限
func (r *Registry) UnregisterPlugin(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 删除路由映射
	for prefix, pluginName := range r.routes {
		if pluginName == name {
			delete(r.routes, prefix)
			break
		}
	}
	delete(r.menus, name)
	delete(r.perms, name)
}

// FindPluginByRoute 根据请求路径查找对应的插件名称
func (r *Registry) FindPluginByRoute(path string) (pluginName string, found bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for prefix, name := range r.routes {
		if strings.HasPrefix(path, prefix) {
			return name, true
		}
	}
	return "", false
}

// GetAllMenus 获取所有插件的菜单
func (r *Registry) GetAllMenus() map[string][]*proto.MenuItem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 返回副本避免外部修改
	result := make(map[string][]*proto.MenuItem, len(r.menus))
	for k, v := range r.menus {
		result[k] = v
	}
	return result
}

// GetAllPermissions 获取所有插件的权限（合并为一个列表）
func (r *Registry) GetAllPermissions() []*proto.Permission {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*proto.Permission
	for _, perms := range r.perms {
		all = append(all, perms...)
	}
	return all
}
