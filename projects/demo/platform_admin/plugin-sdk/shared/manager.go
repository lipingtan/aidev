package shared

import (
	"context"
	"fmt"
	"sync"
)

// Manager 插件管理器 - 协调适配器、注册表和事件发布
type Manager struct {
	mu       sync.RWMutex
	adapters []PluginAdapter
	enabled  map[string]Plugin // 当前已启用的插件
	eventBus EventPublisher
}

// NewManager 创建插件管理器
func NewManager(eventBus EventPublisher) *Manager {
	if eventBus == nil {
		eventBus = NewNoopPublisher()
	}
	return &Manager{
		adapters: make([]PluginAdapter, 0),
		enabled:  make(map[string]Plugin),
		eventBus: eventBus,
	}
}

// AddAdapter 添加插件适配器
func (m *Manager) AddAdapter(adapter PluginAdapter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.adapters = append(m.adapters, adapter)
}

// LoadPlugin 从适配器中加载插件（不启用）
func (m *Manager) LoadPlugin(name string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, adapter := range m.adapters {
		p, err := adapter.Load(name)
		if err == nil {
			return p, nil
		}
	}
	return nil, fmt.Errorf("插件 %s 在所有适配器中均未找到", name)
}

// EnablePlugin 启用插件
func (m *Manager) EnablePlugin(ctx context.Context, name string) error {
	p, err := m.LoadPlugin(name)
	if err != nil {
		return err
	}
	if err := p.OnEnable(ctx); err != nil {
		return fmt.Errorf("启用插件 %s 失败: %w", name, err)
	}
	m.mu.Lock()
	m.enabled[name] = p
	m.mu.Unlock()
	// 广播事件
	_ = m.eventBus.Publish(ctx, PluginEvent{
		Type:       "enable",
		PluginName: name,
		Version:    p.Version(),
	})
	return nil
}

// DisablePlugin 禁用插件
func (m *Manager) DisablePlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	p, ok := m.enabled[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("插件 %s 未启用", name)
	}
	delete(m.enabled, name)
	m.mu.Unlock()

	if err := p.OnDisable(ctx); err != nil {
		return fmt.Errorf("禁用插件 %s 失败: %w", name, err)
	}
	_ = m.eventBus.Publish(ctx, PluginEvent{
		Type:       "disable",
		PluginName: name,
		Version:    p.Version(),
	})
	return nil
}

// ListEnabled 列出当前已启用的插件
func (m *Manager) ListEnabled() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	plugins := make([]Plugin, 0, len(m.enabled))
	for _, p := range m.enabled {
		plugins = append(plugins, p)
	}
	return plugins
}

// ListAvailable 列出所有可用的插件名（来自所有适配器）
func (m *Manager) ListAvailable() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var names []string
	for _, adapter := range m.adapters {
		names = append(names, adapter.List()...)
	}
	return names
}

// IsEnabled 检查插件是否已启用
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.enabled[name]
	return ok
}
