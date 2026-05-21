package main

import (
	"context"
	"game-server/plugin-sdk/helper"
	"game-server/plugin-sdk/proto"
)

// DlcPlugin DLC 管理插件
type DlcPlugin struct{}

func (p *DlcPlugin) Register(ctx context.Context) (*proto.PluginInfo, error) {
	return &proto.PluginInfo{
		Name:        "dlc",
		Version:     "1.0.0",
		Description: "DLC 管理插件 - 提供 DLC 包的 CRUD、文件上传下载、统计功能",
		RoutePrefix: "dlc",
		Menus: []*proto.MenuItem{
			{
				Title: "DLC管理",
				Icon:  "ep:box",
				Path:  "/plugin/dlc",
				Sort:  10,
				Children: []*proto.MenuItem{
					{Title: "DLC列表", Icon: "ep:list", Path: "/plugin/dlc/list", Sort: 1},
					{Title: "DLC统计", Icon: "ep:data-analysis", Path: "/plugin/dlc/stats", Sort: 2},
				},
			},
		},
		Perms: []*proto.Permission{
			{Key: "dlc:list", Title: "DLC列表"},
			{Key: "dlc:create", Title: "创建DLC"},
			{Key: "dlc:update", Title: "更新DLC"},
			{Key: "dlc:delete", Title: "删除DLC"},
			{Key: "dlc:upload", Title: "上传PCK"},
		},
	}, nil
}

func (p *DlcPlugin) HandleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	return handleRequest(ctx, req)
}

func (p *DlcPlugin) Healthcheck(ctx context.Context) (*proto.HealthResponse, error) {
	return &proto.HealthResponse{Healthy: true, Message: "ok"}, nil
}

func (p *DlcPlugin) OnEvent(ctx context.Context, event *proto.Event) error {
	return nil
}

func (p *DlcPlugin) CallPlugin(ctx context.Context, req *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
	return &proto.CallPluginResponse{Payload: []byte(`{}`)}, nil
}

func main() {
	helper.Serve(&DlcPlugin{})
}
