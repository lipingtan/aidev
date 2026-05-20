package proto

import "context"

// PluginInfo 插件注册信息
type PluginInfo struct {
	Name           string       `json:"name"`
	Version        string       `json:"version"`
	Description    string       `json:"description"`
	RoutePrefix    string       `json:"route_prefix"`
	Menus          []MenuItem   `json:"menus"`
	Perms          []Permission `json:"perms"`
	FrontendBundle string       `json:"frontend_bundle"`
}

// MenuItem 菜单项
type MenuItem struct {
	Title    string     `json:"title"`
	Icon     string     `json:"icon"`
	Path     string     `json:"path"`
	Sort     int32      `json:"sort"`
	Children []MenuItem `json:"children"`
}

// Permission 权限声明
type Permission struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}

// HttpRequest HTTP 请求（Host → 插件）
type HttpRequest struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	Query       string            `json:"query"`
	UserID      int64             `json:"user_id"`
	TenantID    int64             `json:"tenant_id"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions"`
	DbDsn       string            `json:"db_dsn"`
}

// HttpResponse HTTP 响应（插件 → Host）
type HttpResponse struct {
	StatusCode int32             `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

// CallPluginRequest 插件间调用请求
type CallPluginRequest struct {
	TargetPlugin string `json:"target_plugin"`
	Method       string `json:"method"`
	Payload      []byte `json:"payload"`
}

// CallPluginResponse 插件间调用响应
type CallPluginResponse struct {
	Payload []byte `json:"payload"`
	Error   string `json:"error"`
}

// Event 事件
type Event struct {
	Source    string `json:"source"`
	Type     string `json:"type"`
	Payload  []byte `json:"payload"`
	Timestamp int64  `json:"timestamp"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}

// PluginService 插件必须实现的服务接口
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
