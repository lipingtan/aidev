package helper

import (
	goplugin "github.com/hashicorp/go-plugin"

	"platform-admin/plugin-sdk/proto"
	"platform-admin/plugin-sdk/shared"
)

// Serve 启动插件 gRPC 服务（基于 hashicorp/go-plugin）
// 插件的 main() 函数调用此方法
func Serve(impl proto.PluginService) {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]goplugin.Plugin{
			"plugin_service": &shared.PluginGRPCPlugin{Impl: impl},
		},
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}
