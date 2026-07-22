package plugin

import (
	"context"
	"fmt"
	"log"
	"sync"

	"platform-admin/plugin-sdk/proto"
)

// 错误变量
var (
	ErrPluginNotFound            = fmt.Errorf("目标插件不存在")
	ErrPluginNotRunning          = fmt.Errorf("目标插件未运行")
	ErrActionNotFound            = fmt.Errorf("目标 Action 不存在")
	ErrActionVersionIncompatible = fmt.Errorf("Action 版本不兼容")
)

// EventBus 插件间异步事件总线
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]string // eventType → []pluginName
	mgr         *PluginManager
}

// NewEventBus 创建事件总线
func NewEventBus(mgr *PluginManager) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]string),
		mgr:         mgr,
	}
}

// Subscribe 注册插件订阅的事件类型
func (b *EventBus) Subscribe(pluginName string, eventTypes []string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, et := range eventTypes {
		// 避免重复订阅
		exists := false
		for _, name := range b.subscribers[et] {
			if name == pluginName {
				exists = true
				break
			}
		}
		if !exists {
			b.subscribers[et] = append(b.subscribers[et], pluginName)
		}
	}
}

// Unsubscribe 注销插件的所有订阅
func (b *EventBus) Unsubscribe(pluginName string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for et, names := range b.subscribers {
		filtered := make([]string, 0, len(names))
		for _, name := range names {
			if name != pluginName {
				filtered = append(filtered, name)
			}
		}
		if len(filtered) == 0 {
			delete(b.subscribers, et)
		} else {
			b.subscribers[et] = filtered
		}
	}
}

// PublishEvent 广播事件到所有订阅该类型的运行中插件
func (b *EventBus) PublishEvent(ctx context.Context, event *proto.Event) error {
	b.mu.RLock()
	names := make([]string, len(b.subscribers[event.Type]))
	copy(names, b.subscribers[event.Type])
	b.mu.RUnlock()

	for _, name := range names {
		inst, ok := b.mgr.GetPlugin(name)
		if !ok || inst.Status != StatusRunning {
			continue
		}
		if err := inst.Service.OnEvent(ctx, event); err != nil {
			// 事件广播失败不阻塞，记录错误继续
			log.Printf("[EventBus] 向插件 %s 广播事件 %s 失败: %v", name, event.Type, err)
		}
	}
	return nil
}

// CallPlugin 同步调用目标插件（含 Action 版本校验）
func (b *EventBus) CallPlugin(ctx context.Context, req *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
	// 1. 检查插件是否存在
	inst, ok := b.mgr.GetPlugin(req.TargetPlugin)
	if !ok {
		return nil, ErrPluginNotFound
	}
	// 2. 检查插件是否运行
	if inst.Status != StatusRunning {
		return nil, ErrPluginNotRunning
	}
	// 3. Action 版本校验（如果请求带了 ActionVersion）
	if req.ActionVersion != "" {
		ar := b.mgr.ActionRegistry()
		_, found := ar.FindAction(req.TargetPlugin, req.Method)
		if !found {
			return nil, ErrActionNotFound
		}
		if !ar.IsCompatible(req.TargetPlugin, req.Method, req.ActionVersion) {
			return nil, ErrActionVersionIncompatible
		}
	}
	// 4. 转发调用
	return inst.Service.CallPlugin(ctx, req)
}

// HostServiceImpl 实现 proto.HostService 接口，委托给 EventBus
type HostServiceImpl struct {
	bus *EventBus
}

// NewHostService 创建 HostService 实例
func NewHostService(bus *EventBus) *HostServiceImpl {
	return &HostServiceImpl{bus: bus}
}

// PublishEvent 发布事件（委托给 EventBus）
func (h *HostServiceImpl) PublishEvent(ctx context.Context, event *proto.Event) error {
	return h.bus.PublishEvent(ctx, event)
}

// CallPlugin 调用目标插件（委托给 EventBus）
func (h *HostServiceImpl) CallPlugin(ctx context.Context, req *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
	return h.bus.CallPlugin(ctx, req)
}
