# Tasks: SaaS 底座插件架构

## Task Dependency Graph

```
T1 → T2 → T3 → T4 → T5 → T6 → T7 → T8 → T9
```

---

- [x] 1. 定义 gRPC 协议（proto）
  - **复杂度**: 中
  - **Scope:** 新增 `common/plugin/proto/plugin.proto` 及生成的 `.pb.go` 文件；不触碰现有代码
  - **Constraints:** 使用 proto3 语法；PluginService 和 HostService 分开定义；message 命名遵循 Go 惯例
  - **Acceptance:**
  - AC: proto 文件定义 PluginService（Register/HandleRequest/Healthcheck/OnEvent/CallPlugin）
  - AC: proto 文件定义 HostService（PublishEvent/CallPlugin）
  - AC: protoc 生成 Go 代码无报错
  - AC: go build ./... 零错误
  - _Requirements: FR-2_

- [x] 2. 实现插件管理器核心
  - **复杂度**: 高
  - **Scope:** 新增 `common/plugin/types.go`、`manager.go`、`registry.go`；涉及模块 common/plugin；不触碰 app/ 下业务代码
  - **Constraints:** 基于 hashicorp/go-plugin 库；支持运行时启停；Manager 必须线程安全（sync.RWMutex）；插件崩溃不 panic Host
  - **Acceptance:**
  - AC: PluginManager 可启动/停止插件子进程
  - AC: 启动时调用 Register RPC 获取 PluginInfo
  - AC: Registry 正确存储路由/菜单/权限映射
  - AC: 插件停止时注销对应条目
  - AC: go build ./... 零错误
  - _Requirements: FR-1_

- [x] 3. 实现 HTTP 请求代理
  - **复杂度**: 高
  - **Scope:** 新增 `common/plugin/proxy.go`；修改主路由注册（注册代理中间件）；不触碰现有 API handler
  - **Constraints:** 路由匹配 `/api/v1/plugin/{name}/*`；必须注入 user_id/tenant_id/roles/permissions/db_dsn；插件未运行返回 503
  - **Acceptance:**
  - AC: 匹配插件路由前缀的请求被正确转发
  - AC: HttpRequest 包含完整上下文信息
  - AC: 插件响应正确写回 HTTP
  - AC: 插件未运行时返回 503
  - AC: go build ./... 零错误
  - _Requirements: FR-3_

- [x] 4. 实现事件总线和插件间通信
  - **复杂度**: 高
  - **Scope:** 新增 `common/plugin/event_bus.go`；实现 HostService gRPC server；不触碰业务代码
  - **Constraints:** 同步 CallPlugin 通过 Host 中转；异步事件内存广播（不持久化）；事件类型为字符串匹配
  - **Acceptance:**
  - AC: PublishEvent 广播到所有订阅插件的 OnEvent
  - AC: CallPlugin 同步调用目标插件并返回响应
  - AC: 目标插件不存在时返回错误
  - AC: go build ./... 零错误
  - _Requirements: FR-2_

- [x] 5. 实现插件安装/卸载逻辑
  - **复杂度**: 高
  - **Scope:** 新增 `common/plugin/installer.go`、`app/plugin/models/sys_plugin.go`；涉及模块 common/plugin + app/plugin；不触碰现有 model
  - **Constraints:** 支持 zip/tar.gz 解压；支持远程 URL 下载；安装目录 `plugins/{name}/`；前端 bundle 复制到 `static/plugins/{name}/`；卸载时可选清除 `{name}_*` 表
  - **Acceptance:**
  - AC: 本地上传 zip 安装成功，文件解压到正确目录
  - AC: 远程 URL 下载安装成功
  - AC: sys_plugin 表记录正确
  - AC: 卸载时文件删除、记录清除
  - AC: go build ./... 零错误
  - _Requirements: FR-1, FR-6_

- [x] 6. 实现插件管理 API 和路由
  - **复杂度**: 高
  - **Scope:** 新增 `app/plugin/apis/plugin.go`、`service/plugin.go`、`router/router.go`；修改 setup.go 的 runMigrations 加入 sys_plugin
  - **Constraints:** 所有接口需 JWT + 管理员权限；启动/停止调用 PluginManager；安装调用 Installer
  - **Acceptance:**
  - AC: GET /api/v1/plugins 返回插件列表
  - AC: POST /api/v1/plugins/install 安装成功
  - AC: POST /api/v1/plugins/:name/start 启动成功
  - AC: POST /api/v1/plugins/:name/stop 停止成功
  - AC: DELETE /api/v1/plugins/:name 卸载成功
  - AC: go build ./... 零错误
  - AC: 【回归】现有 API 不受影响（RG-1~RG-5）
  - _Requirements: FR-1, FR-6_

- [x] 7. 实现前端插件加载器
  - **复杂度**: 中
  - **Scope:** 新增 `frontend/src/utils/plugin-loader.ts`、`src/store/modules/plugin.ts`；修改 `src/router/index.ts`
  - **Constraints:** 使用 dynamic import 加载 `/static/plugins/{name}/index.js`；插件 bundle 必须导出 `{ routes, menus }`；加载失败不影响 Host 前端
  - **Acceptance:**
  - AC: 启动时请求插件列表并加载运行中插件的前端 bundle
  - AC: 插件路由注册到 Vue Router
  - AC: 插件菜单注入侧边栏
  - AC: 加载失败时 console.error 但不阻塞
  - _Requirements: FR-4_

- [x] 8. 实现插件管理界面
  - **复杂度**: 中
  - **Scope:** 新增 `frontend/src/views/admin/plugin/index.vue`；修改 menu_init.sql 注册菜单
  - **Constraints:** 风格与现有管理页面一致；操作按钮需确认弹窗；上传支持 zip/tar.gz
  - **Acceptance:**
  - AC: 展示插件列表（名称/版本/状态/描述）
  - AC: 安装支持本地上传和远程 URL
  - AC: 启动/停止/卸载操作正常
  - AC: 菜单在侧边栏正确显示
  - _Requirements: FR-6_

- [x] 9. 实现插件开发脚手架 CLI
  - **复杂度**: 中
  - **Scope:** 新增 `scaffold/` 目录（独立 Go 模块）；不触碰 Host 代码
  - **Constraints:** 使用 cobra 或简单 flag 解析；模板使用 text/template；生成的插件项目可独立编译
  - **Acceptance:**
  - AC: `go run scaffold/main.go new-plugin myplugin` 生成完整项目骨架
  - AC: 生成的项目包含 main.go、proto、frontend 组件、构建脚本、README
  - AC: 生成的项目可独立 go build 成功
  - _Requirements: FR-7_
