package plugin

import (
	"strconv"
	"strings"
	"sync"
)

// ActionDescriptor 插件暴露的 Action 描述
type ActionDescriptor struct {
	Name    string // Action 名称
	Version string // 语义化版本，如 "1.0"
}

// ActionRegistry 插件 Action 注册表（内存级，线程安全）
type ActionRegistry struct {
	mu      sync.RWMutex
	actions map[string][]ActionDescriptor // pluginName → actions
}

// NewActionRegistry 创建 Action 注册表
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string][]ActionDescriptor),
	}
}

// Register 注册插件暴露的 Actions
func (r *ActionRegistry) Register(pluginName string, actions []ManifestAction) {
	r.mu.Lock()
	defer r.mu.Unlock()

	descriptors := make([]ActionDescriptor, 0, len(actions))
	for _, a := range actions {
		descriptors = append(descriptors, ActionDescriptor{
			Name:    a.Name,
			Version: a.Version,
		})
	}
	r.actions[pluginName] = descriptors
}

// Unregister 注销插件的所有 Actions
func (r *ActionRegistry) Unregister(pluginName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.actions, pluginName)
}

// FindAction 查找指定插件的指定 Action
func (r *ActionRegistry) FindAction(pluginName, actionName string) (*ActionDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	actions, exists := r.actions[pluginName]
	if !exists {
		return nil, false
	}
	for i := range actions {
		if actions[i].Name == actionName {
			return &actions[i], true
		}
	}
	return nil, false
}

// IsCompatible 检查请求版本与目标插件 Action 版本的兼容性
// 规则：major 版本相同即兼容
func (r *ActionRegistry) IsCompatible(pluginName, actionName, requestVersion string) bool {
	action, found := r.FindAction(pluginName, actionName)
	if !found {
		return false
	}

	reqMajor := parseMajorVersion(requestVersion)
	actionMajor := parseMajorVersion(action.Version)

	return reqMajor == actionMajor && reqMajor >= 0
}

// parseMajorVersion 从版本字符串中提取 major 版本号
// 支持格式："1.0"、"1"、"2.3.1"
func parseMajorVersion(version string) int {
	if version == "" {
		return -1
	}
	parts := strings.SplitN(version, ".", 2)
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}
	return major
}
