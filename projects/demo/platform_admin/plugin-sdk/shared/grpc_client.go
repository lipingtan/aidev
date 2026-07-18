package shared

import (
	"context"

	"platform-admin/plugin-sdk/proto"
)

// 编译时接口检查：确保 GRPCPluginClient 实现 proto.PluginService
var _ proto.PluginService = (*GRPCPluginClient)(nil)

// GRPCPluginClient 将 gRPC client 适配为业务 PluginService 接口
type GRPCPluginClient struct {
	client proto.PluginServiceClient
}

func (c *GRPCPluginClient) Register(ctx context.Context) (*proto.PluginInfo, error) {
	return c.client.Register(ctx, &proto.Empty{})
}

func (c *GRPCPluginClient) HandleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	return c.client.HandleRequest(ctx, req)
}

func (c *GRPCPluginClient) Healthcheck(ctx context.Context) (*proto.HealthResponse, error) {
	return c.client.Healthcheck(ctx, &proto.Empty{})
}

func (c *GRPCPluginClient) OnEvent(ctx context.Context, event *proto.Event) error {
	_, err := c.client.OnEvent(ctx, event)
	return err
}

func (c *GRPCPluginClient) CallPlugin(ctx context.Context, req *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
	return c.client.CallPlugin(ctx, req)
}
