# Tasks: 集成 hashicorp/go-plugin

## Task Dependency Graph

```
T1 → T2 → T3 → T4 → T5
```

---

- [ ] 1. protoc 生成 gRPC 代码
  - **复杂度**: 中
  - **Scope:** 修改 `plugin-sdk/proto/plugin.proto`（调整 go_package）；生成 `plugin.pb.go` 和 `plugin_grpc.pb.go`；更新 `plugin-sdk/go.mod` 添加 protobuf/grpc 依赖
  - **Constraints:** 保留现有 types.go 中的 Go 接口不变；proto message 名加 Pb 后缀避免与 Go struct 冲突（或使用同名，types.go 中的 struct 改名）
  - **Acceptance:**
  - AC: protoc 生成代码无报错
  - AC: plugin-sdk go build ./... 零错误
  - AC: 现有 Go 接口（PluginService/HostService）保持不变
  - _Requirements: FR-1_

- [ ] 2. 实现 go-plugin 胶水层（shared 包）
  - **复杂度**: 高
  - **Scope:** 新增 `plugin-sdk/shared/` 目录：plugin.go、grpc_server.go、grpc_client.go、convert.go
  - **Constraints:** GRPCClient 必须实现 proto.PluginService 接口；GRPCServer 必须实现生成的 gRPC server 接口；convert.go 处理 Go struct ↔ protobuf message 转换
  - **Acceptance:**
  - AC: PluginGRPCPlugin 实现 goplugin.GRPCPlugin 接口
  - AC: GRPCClient 实现 proto.PluginService 接口
  - AC: GRPCServer 正确委托到 proto.PluginService 实现
  - AC: plugin-sdk go build ./... 零错误
  - _Requirements: FR-2_

- [ ] 3. 改造 helper/serve.go 使用 go-plugin
  - **复杂度**: 低
  - **Scope:** 修改 `plugin-sdk/helper/serve.go`
  - **Acceptance:**
  - AC: Serve() 调用 goplugin.Serve 启动 gRPC server
  - AC: 使用 shared.Handshake 和 shared.PluginGRPCPlugin
  - AC: plugin-sdk go build ./... 零错误
  - _Requirements: FR-2_

- [ ] 4. 改造 Host PluginManager 使用 go-plugin Client
  - **复杂度**: 高
  - **Scope:** 修改 `backend/common/plugin/manager.go`、`types.go`；更新 `backend/go.mod`
  - **Constraints:** Start 方法接收 binaryPath 参数；PluginInstance 增加 goplugin.Client 字段；Stop 时调用 client.Kill()；保持 GetPlugin/ListPlugins/Healthcheck 接口不变
  - **Acceptance:**
  - AC: Start(name, binaryPath) 启动子进程并建立 gRPC 连接
  - AC: Stop(name) 终止子进程
  - AC: 插件崩溃后 GetPlugin 返回 StatusError
  - AC: backend go build ./... 零错误
  - AC: 【回归】现有测试 TestPluginProxy 等需要适配（mock 模式保留）
  - _Requirements: FR-3, FR-4_

- [ ] 5. 端到端验证
  - **复杂度**: 中
  - **Scope:** 用 DLC 插件（或 scaffold 生成的测试插件）编译为二进制，通过 PluginManager 启动/调用/停止
  - **Acceptance:**
  - AC: 插件二进制可独立编译
  - AC: Host 启动插件子进程成功
  - AC: 通过 gRPC 调用 Register/HandleRequest/Healthcheck 正常
  - AC: Stop 后子进程退出
  - AC: 插件强制 kill 后 Host 检测到异常
  - _Requirements: FR-3_
