package shared

import (
	"fmt"
	"sync"
)

// LocalAdapter 内嵌插件适配器 - 管理编译期注册的本地插件
type LocalAdapter struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// NewLocalAdapter 创建本地适配器实例
func NewLocalAdapter() *LocalAdapter {
	return &LocalAdapter{
		plugins: make(map[string]Plugin),
	}
}

// Register 注册一个本地插件
func (a *LocalAdapter) Register(p Plugin) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.plugins[p.Name()] = p
}

// Load 加载指定名称的插件
func (a *LocalAdapter) Load(name string) (Plugin, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	p, ok := a.plugins[name]
	if !ok {
		return nil, fmt.Errorf("插件 %s 未注册", name)
	}
	return p, nil
}

// Unload 卸载指定名称的插件
func (a *LocalAdapter) Unload(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.plugins[name]; !ok {
		return fmt.Errorf("插件 %s 未注册", name)
	}
	delete(a.plugins, name)
	return nil
}

// List 列出所有已注册的插件名
func (a *LocalAdapter) List() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	names := make([]string, 0, len(a.plugins))
	for name := range a.plugins {
		names = append(names, name)
	}
	return names
}
