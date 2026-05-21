package shared

import (
	"context"

	"game-server/plugin-sdk/proto"
)

// GRPCPluginServer 将业务 PluginService 适配为 gRPC server handler
type GRPCPluginServer struct {
	proto.UnimplementedPluginServiceServer
	Impl proto.PluginService
}

func (s *GRPCPluginServer) Register(ctx context.Context, _ *proto.Empty) (*proto.PluginInfo, error) {
	return s.Impl.Register(ctx)
}

func (s *GRPCPluginServer) HandleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	return s.Impl.HandleRequest(ctx, req)
}

func (s *GRPCPluginServer) Healthcheck(ctx context.Context, _ *proto.Empty) (*proto.HealthResponse, error) {
	return s.Impl.Healthcheck(ctx)
}

func (s *GRPCPluginServer) OnEvent(ctx context.Context, event *proto.Event) (*proto.Empty, error) {
	err := s.Impl.OnEvent(ctx, event)
	return &proto.Empty{}, err
}

func (s *GRPCPluginServer) CallPlugin(ctx context.Context, req *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
	return s.Impl.CallPlugin(ctx, req)
}
