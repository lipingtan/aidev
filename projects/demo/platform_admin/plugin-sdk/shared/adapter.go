package shared

// PluginAdapter 插件适配器接口 - 支持本地和远程两种模式
type PluginAdapter interface {
	// Load 加载指定名称的插件
	Load(name string) (Plugin, error)
	// Unload 卸载指定名称的插件
	Unload(name string) error
	// List 列出所有已注册的插件名
	List() []string
}
