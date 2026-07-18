package shared

import "context"

// PluginEvent 插件变更事件
type PluginEvent struct {
	Type       string `json:"type"`       // install/enable/disable/uninstall/upgrade/rollback
	PluginName string `json:"pluginName"`
	Version    string `json:"version"`
}

// EventPublisher 事件发布接口 - 用于集群间广播插件状态变更
type EventPublisher interface {
	// Publish 发布插件变更事件
	Publish(ctx context.Context, event PluginEvent) error
	// Subscribe 订阅插件变更事件（返回事件 channel）
	Subscribe(ctx context.Context) (<-chan PluginEvent, error)
	// Close 关闭发布者
	Close() error
}

// NoopPublisher 空实现 - 单实例部署时使用
type NoopPublisher struct{}

// NewNoopPublisher 创建空事件发布者
func NewNoopPublisher() *NoopPublisher { return &NoopPublisher{} }

func (p *NoopPublisher) Publish(ctx context.Context, event PluginEvent) error { return nil }
func (p *NoopPublisher) Subscribe(ctx context.Context) (<-chan PluginEvent, error) {
	ch := make(chan PluginEvent)
	return ch, nil
}
func (p *NoopPublisher) Close() error { return nil }
