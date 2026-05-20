package plugin

import (
	"context"
	"fmt"
	"sync"

	"game-server/plugin-sdk/proto"
)

// PluginManager 插件生命周期管理器
type PluginManager struct {
	mu       sync.RWMutex
	plugins  map[string]*PluginInstance // name → instance
	registry *Registry
	eventBus *EventBus
}

// NewPluginManager 创建插件管理器
func NewPluginManager() *PluginManager {
	m := &PluginManager{
		plugins:  make(map[string]*PluginInstance),
		registry: NewRegistry(),
	}
	m.eventBus = NewEventBus(m)
	return m
}

// Registry 获取注册表实例
func (m *PluginManager) Registry() *Registry {
	return m.registry
}

// EventBus 获取事件总线实例
func (m *PluginManager) EventBus() *EventBus {
	return m.eventBus
}

// HostService 获取 HostService 实例
func (m *PluginManager) HostService() proto.HostService {
	return NewHostService(m.eventBus)
}

// Register 注册插件（进程内），仅将服务实例加入管理，不启动
func (m *PluginManager) Register(name string, svc proto.PluginService) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("插件 %s 已注册", name)
	}

	m.plugins[name] = &PluginInstance{
		Service: svc,
		Status:  StatusStopped,
	}
	return nil
}

// Start 启动指定插件，调用 Register 获取 PluginInfo 并注册到 Registry
func (m *PluginManager) Start(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("插件 %s 未注册", name)
	}
	if inst.Status == StatusRunning {
		return fmt.Errorf("插件 %s 已在运行", name)
	}

	// 调用插件的 Register 方法获取插件信息
	info, err := inst.Service.Register(context.Background())
	if err != nil {
		inst.Status = StatusError
		return fmt.Errorf("插件 %s 注册失败: %w", name, err)
	}

	inst.Info = info
	inst.Status = StatusRunning

	// 注册路由/菜单/权限到 Registry
	m.registry.RegisterPlugin(info)

	return nil
}

// Stop 停止指定插件，从 Registry 注销
func (m *PluginManager) Stop(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("插件 %s 未注册", name)
	}
	if inst.Status != StatusRunning {
		return fmt.Errorf("插件 %s 未在运行", name)
	}

	// 从 Registry 注销
	m.registry.UnregisterPlugin(name)

	// 注销事件订阅
	m.eventBus.Unsubscribe(name)

	inst.Status = StatusStopped
	return nil
}

// StopAll 停止所有运行中的插件
func (m *PluginManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, inst := range m.plugins {
		if inst.Status == StatusRunning {
			m.registry.UnregisterPlugin(name)
			m.eventBus.Unsubscribe(name)
			inst.Status = StatusStopped
		}
	}
}

// GetPlugin 获取指定插件实例
func (m *PluginManager) GetPlugin(name string) (*PluginInstance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inst, exists := m.plugins[name]
	return inst, exists
}

// ListPlugins 列出所有插件实例
func (m *PluginManager) ListPlugins() []*PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*PluginInstance, 0, len(m.plugins))
	for _, inst := range m.plugins {
		list = append(list, inst)
	}
	return list
}

// Healthcheck 对指定插件执行健康检查
func (m *PluginManager) Healthcheck(name string) (*proto.HealthResponse, error) {
	m.mu.RLock()
	inst, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("插件 %s 未注册", name)
	}
	if inst.Status != StatusRunning {
		return &proto.HealthResponse{Healthy: false, Message: "插件未运行"}, nil
	}

	return inst.Service.Healthcheck(context.Background())
}
