# 设计：集成 hashicorp/go-plugin

## 技术方案

### 架构变更

```
改造前：
Host 进程 → 直接调用 proto.PluginService 接口 → 插件代码（同进程）

改造后：
Host 进程 → go-plugin Client → stdio gRPC → 插件子进程（独立二进制）
         ← go-plugin 自动重连/检测崩溃 ←
```

### 核心组件

#### 1. plugin-sdk/proto/ — protobuf + gRPC 生成

用 protoc 从 plugin.proto 生成：
- `plugin.pb.go` — message 定义
- `plugin_grpc.pb.go` — gRPC service stub

保留 `types.go` 中的 Go 接口定义（PluginService/HostService），作为业务层接口。

#### 2. plugin-sdk/shared/ — go-plugin 胶水层

```go
// shared/plugin.go
package shared

import (
    "context"
    goplugin "github.com/hashicorp/go-plugin"
    "google.golang.org/grpc"
    
    "platform-admin/plugin-sdk/proto"
)

// Handshake 握手配置
var Handshake = goplugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "GAME_SERVER_PLUGIN",
    MagicCookieValue: "plugin_v1",
}

// PluginMap 插件类型映射
var PluginMap = map[string]goplugin.Plugin{
    "plugin": &PluginGRPCPlugin{},
}

// PluginGRPCPlugin 实现 goplugin.GRPCPlugin 接口
type PluginGRPCPlugin struct {
    goplugin.Plugin
    Impl proto.PluginService // 插件端提供具体实现
}

func (p *PluginGRPCPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
    // 注册 gRPC server（插件端调用）
    proto.RegisterPluginServiceServer(s, &GRPCServer{Impl: p.Impl})
    return nil
}

func (p *PluginGRPCPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
    // 返回 gRPC client（Host 端调用）
    return &GRPCClient{client: proto.NewPluginServiceClient(c)}, nil
}
```

#### 3. plugin-sdk/shared/grpc_server.go — gRPC Server 适配

将 `proto.PluginService` Go 接口适配为 gRPC server handler：

```go
type GRPCServer struct {
    proto.UnimplementedPluginServiceServer
    Impl proto.PluginService
}

func (s *GRPCServer) Register(ctx context.Context, req *proto.Empty) (*proto.PluginInfoPb, error) {
    info, err := s.Impl.Register(ctx)
    // 将 Go struct 转为 protobuf message
    return toPluginInfoPb(info), err
}

func (s *GRPCServer) HandleRequest(ctx context.Context, req *proto.HttpRequestPb) (*proto.HttpResponsePb, error) {
    goReq := fromHttpRequestPb(req)
    resp, err := s.Impl.HandleRequest(ctx, goReq)
    return toHttpResponsePb(resp), err
}
// ... 其他方法类似
```

#### 4. plugin-sdk/shared/grpc_client.go — gRPC Client 适配

将 gRPC client stub 适配为 `proto.PluginService` Go 接口：

```go
type GRPCClient struct {
    client proto.PluginServiceClient
}

func (c *GRPCClient) Register(ctx context.Context) (*proto.PluginInfo, error) {
    resp, err := c.client.Register(ctx, &proto.Empty{})
    return fromPluginInfoPb(resp), err
}

func (c *GRPCClient) HandleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
    pbReq := toHttpRequestPb(req)
    resp, err := c.client.HandleRequest(ctx, pbReq)
    return fromHttpResponsePb(resp), err
}
// ... 其他方法类似
```

#### 5. plugin-sdk/helper/serve.go — 插件端启动

```go
func Serve(impl proto.PluginService) {
    goplugin.Serve(&goplugin.ServeConfig{
        HandshakeConfig: shared.Handshake,
        Plugins: map[string]goplugin.Plugin{
            "plugin": &shared.PluginGRPCPlugin{Impl: impl},
        },
        GRPCServer: goplugin.DefaultGRPCServer,
    })
}
```

#### 6. Host PluginManager 改造

```go
// Start 启动插件子进程
func (m *PluginManager) Start(name string, binaryPath string) error {
    client := goplugin.NewClient(&goplugin.ClientConfig{
        HandshakeConfig: shared.Handshake,
        Plugins:         shared.PluginMap,
        Cmd:             exec.Command(binaryPath),
        AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
    })
    
    rpcClient, err := client.Client()
    raw, err := rpcClient.Dispense("plugin")
    svc := raw.(proto.PluginService)
    
    // 调用 Register 获取插件信息
    info, err := svc.Register(ctx)
    
    // 存储实例
    m.plugins[name] = &PluginInstance{
        Info:    info,
        Service: svc,
        Status:  StatusRunning,
        client:  client, // 保存 client 用于 Stop 时 Kill
    }
}

// Stop 停止插件
func (m *PluginManager) Stop(name string) error {
    inst := m.plugins[name]
    inst.client.Kill() // 终止子进程
    inst.Status = StatusStopped
}
```

### 文件结构变更

```
plugin-sdk/
├── go.mod              # 新增 hashicorp/go-plugin、google.golang.org/grpc 依赖
├── proto/
│   ├── plugin.proto    # 保持不变
│   ├── plugin.pb.go    # 新增：protoc 生成
│   ├── plugin_grpc.pb.go # 新增：protoc 生成
│   └── types.go        # 保持：Go 接口定义
├── shared/             # 新增：go-plugin 胶水层
│   ├── plugin.go       # Handshake + PluginGRPCPlugin
│   ├── grpc_server.go  # GRPCServer 适配器
│   ├── grpc_client.go  # GRPCClient 适配器
│   └── convert.go      # Go struct ↔ protobuf message 转换
└── helper/
    └── serve.go        # 改造：调用 goplugin.Serve

backend/common/plugin/
├── manager.go          # 改造：使用 goplugin.Client 管理子进程
├── types.go            # 修改：PluginInstance 增加 client 字段
└── ...                 # 其他文件不变
```

### 依赖新增

plugin-sdk/go.mod:
```
require (
    github.com/hashicorp/go-plugin v1.8.0
    google.golang.org/grpc v1.72.0
    google.golang.org/protobuf v1.36.6
)
```

backend/go.mod:
```
require (
    github.com/hashicorp/go-plugin v1.8.0  // 新增
)
```

## 不变行为清单

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | HTTP 代理转发行为不变 | 插件请求正确代理 |
| RG-2 | 事件广播行为不变 | OnEvent 正确触发 |
| RG-3 | 插件管理 API 行为不变 | CRUD 接口正常 |
| RG-4 | 现有业务不受影响 | game/admin 模块正常 |

## 正确性属性

- 进程隔离：插件 panic 不影响 Host（go-plugin 自动检测进程退出）
- 通信可靠：gRPC over stdio，无网络端口占用
- 优雅关闭：Stop 时 Kill 子进程，释放资源
