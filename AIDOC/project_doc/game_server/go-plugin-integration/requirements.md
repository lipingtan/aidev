# 需求：集成 hashicorp/go-plugin 实现子进程隔离

## 背景

当前插件架构使用进程内接口调用模式。需要改造为基于 hashicorp/go-plugin 的子进程隔离模式，实现：
- 插件编译为独立二进制
- Host 通过 stdio gRPC 与插件通信
- 插件崩溃不影响 Host
- 支持运行时启停插件

## 用户故事

- 作为系统管理员，我希望插件崩溃不会导致整个服务宕机
- 作为开发者，我希望插件可以独立编译部署，不需要重新编译 Host

## 功能需求

### FR-1: 使用 protoc 生成真正的 gRPC 代码
**验收标准：**
- WHEN 执行 protoc THEN 系统 SHALL 生成 plugin_grpc.pb.go 和 plugin.pb.go
- WHEN 编译 plugin-sdk THEN go build SHALL 零错误

### FR-2: plugin-sdk 实现 go-plugin 的 GRPCPlugin 接口
**验收标准：**
- WHEN 插件调用 helper.Serve() THEN 系统 SHALL 启动 go-plugin gRPC server
- WHEN Host 连接插件 THEN 系统 SHALL 通过 gRPC 调用 PluginService 方法

### FR-3: PluginManager 使用 go-plugin Client 管理子进程
**验收标准：**
- WHEN 调用 Start(name) THEN 系统 SHALL 启动插件子进程并建立 gRPC 连接
- WHEN 调用 Stop(name) THEN 系统 SHALL 优雅关闭子进程
- WHEN 插件进程崩溃 THEN Host SHALL 检测到并标记状态为 Error

### FR-4: 保持现有 API 兼容
**验收标准：**
- WHEN 通过 HTTP 代理调用插件 THEN 行为 SHALL 与改造前一致
- WHEN 事件广播到插件 THEN 行为 SHALL 与改造前一致

## 非功能需求

- go-plugin 版本：v1.8.0（最新稳定版）
- 通信协议：gRPC over stdio
- 插件二进制路径：从 sys_plugin.binary_path 读取
