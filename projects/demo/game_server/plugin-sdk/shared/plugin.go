package shared

import (
	"context"

	goplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"game-server/plugin-sdk/proto"
)

// Handshake 握手配置，Host 和插件必须匹配
var Handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GAME_SERVER_PLUGIN",
	MagicCookieValue: "plugin_v1",
}

// PluginMap 插件类型映射（Host 端使用，不需要 Impl）
var PluginMap = map[string]goplugin.Plugin{
	"plugin_service": &PluginGRPCPlugin{},
}

// PluginGRPCPlugin 实现 goplugin.GRPCPlugin 接口
type PluginGRPCPlugin struct {
	goplugin.Plugin
	Impl proto.PluginService // 插件端提供具体实现，Host 端为 nil
}

// GRPCServer 插件端调用，注册 gRPC server
func (p *PluginGRPCPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterPluginServiceServer(s, &GRPCPluginServer{Impl: p.Impl})
	return nil
}

// GRPCClient Host 端调用，返回实现了 proto.PluginService 的 client
func (p *PluginGRPCPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCPluginClient{client: proto.NewPluginServiceClient(c)}, nil
}
