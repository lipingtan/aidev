package plugin

import "game-server/plugin-sdk/proto"

// PluginStatus 插件运行状态
type PluginStatus int

const (
	// StatusStopped 已停止
	StatusStopped PluginStatus = iota
	// StatusRunning 运行中
	StatusRunning
	// StatusError 异常
	StatusError
)

// PluginInstance 运行中的插件实例
type PluginInstance struct {
	Info    *proto.PluginInfo
	Service proto.PluginService
	Status  PluginStatus
}
