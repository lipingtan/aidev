package proto

import "context"

// PluginService 插件必须实现的服务接口（业务层）
// 注意：这里使用 protobuf 生成的 message 类型
type PluginService interface {
	Register(ctx context.Context) (*PluginInfo, error)
	HandleRequest(ctx context.Context, req *HttpRequest) (*HttpResponse, error)
	Healthcheck(ctx context.Context) (*HealthResponse, error)
	OnEvent(ctx context.Context, event *Event) error
	CallPlugin(ctx context.Context, req *CallPluginRequest) (*CallPluginResponse, error)
}

// HostService Host 提供给插件的反向调用接口
type HostService interface {
	PublishEvent(ctx context.Context, event *Event) error
	CallPlugin(ctx context.Context, req *CallPluginRequest) (*CallPluginResponse, error)
}
