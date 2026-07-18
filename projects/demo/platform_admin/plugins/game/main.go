package main

import (
	"context"
	"platform-admin/plugin-sdk/helper"
	"platform-admin/plugin-sdk/proto"
)

// GamePlugin 游戏管理插件
type GamePlugin struct{}

func (p *GamePlugin) Register(ctx context.Context) (*proto.PluginInfo, error) {
	return &proto.PluginInfo{
		Name:        "game",
		Version:     "1.0.0",
		Description: "游戏管理插件 - 提供游戏列表、DLC、玩家、订单、支付配置、H5 页面管理功能",
		RoutePrefix: "game",
		Menus: []*proto.MenuItem{
			{
				Title: "游戏管理",
				Icon:  "ep:game-pad",
				Path:  "/plugin/game",
				Sort:  10,
				Children: []*proto.MenuItem{
					{Title: "游戏列表", Icon: "ep:list", Path: "/plugin/game/list", Sort: 1},
					{Title: "DLC管理", Icon: "ep:box", Path: "/plugin/game/dlc", Sort: 2},
					{Title: "玩家管理", Icon: "ep:user", Path: "/plugin/game/player", Sort: 3},
					{Title: "订单管理", Icon: "ep:document", Path: "/plugin/game/order", Sort: 4},
					{Title: "支付配置", Icon: "ep:credit-card", Path: "/plugin/game/payment", Sort: 5},
					{Title: "H5页面", Icon: "ep:monitor", Path: "/plugin/game/h5", Sort: 6},
				},
			},
		},
		Perms: []*proto.Permission{
			{Key: "game:list", Title: "游戏列表"},
			{Key: "game:create", Title: "创建游戏"},
			{Key: "game:update", Title: "更新游戏"},
			{Key: "game:delete", Title: "删除游戏"},
			{Key: "dlc:list", Title: "DLC列表"},
			{Key: "dlc:create", Title: "创建DLC"},
			{Key: "dlc:update", Title: "更新DLC"},
			{Key: "dlc:delete", Title: "删除DLC"},
			{Key: "dlc:upload", Title: "上传PCK"},
			{Key: "player:list", Title: "玩家列表"},
			{Key: "player:ban", Title: "封禁玩家"},
			{Key: "order:list", Title: "订单列表"},
			{Key: "order:refund", Title: "订单退款"},
			{Key: "payment:list", Title: "支付配置列表"},
			{Key: "payment:save", Title: "保存支付配置"},
			{Key: "h5:list", Title: "H5页面列表"},
			{Key: "h5:create", Title: "创建H5页面"},
			{Key: "h5:update", Title: "更新H5页面"},
			{Key: "h5:delete", Title: "删除H5页面"},
		},
	}, nil
}

func (p *GamePlugin) HandleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	return handleRequest(ctx, req)
}

func (p *GamePlugin) Healthcheck(ctx context.Context) (*proto.HealthResponse, error) {
	return &proto.HealthResponse{Healthy: true, Message: "ok"}, nil
}

func (p *GamePlugin) OnEvent(ctx context.Context, event *proto.Event) error {
	return nil
}

func (p *GamePlugin) CallPlugin(ctx context.Context, req *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
	return &proto.CallPluginResponse{Payload: []byte(`{}`)}, nil
}

func main() {
	helper.Serve(&GamePlugin{})
}
