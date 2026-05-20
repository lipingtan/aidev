# 设计：SaaS 底座插件架构

## 技术方案

### 整体架构

```
Host 进程（Gin + go-plugin PluginManager）
├── core/           # 基础能力（用户/角色/菜单/字典/租户）
├── plugin_mgr/     # 插件管理器
│   ├── manager.go      # 插件生命周期管理
│   ├── registry.go     # 路由/菜单/权限注册表
│   ├── proxy.go        # HTTP → gRPC 请求代理
│   └── event_bus.go    # 插件间异步事件总线
├── proto/          # gRPC 协议定义
│   └── plugin.proto
└── static/plugins/ # 插件前端 bundle 存储

插件进程（独立二进制，通过 stdio gRPC 与 Host 通信）
├── main.go         # 插件入口，实现 PluginService gRPC 接口
├── handler/        # 业务 HTTP handler
├── model/          # 数据模型（表名前缀 {pluginName}_）
└── frontend/dist/  # 前端 bundle（构建产物）
```

### API 设计

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/plugins | 插件列表 | JWT + 管理员 |
| POST | /api/v1/plugins/install | 安装插件（上传/URL） | JWT + 管理员 |
| POST | /api/v1/plugins/:name/start | 启动插件 | JWT + 管理员 |
| POST | /api/v1/plugins/:name/stop | 停止插件 | JWT + 管理员 |
| DELETE | /api/v1/plugins/:name | 卸载插件 | JWT + 管理员 |
| GET | /api/v1/plugins/:name/health | 插件健康检查 | JWT |
| ANY | /api/v1/plugin/{name}/* | 代理到插件 | JWT + 租户 |

### gRPC 协议设计（proto）

```protobuf
syntax = "proto3";
package plugin;

// 插件注册信息
message PluginInfo {
  string name = 1;
  string version = 2;
  string description = 3;
  string route_prefix = 4;          // 路由前缀，如 "game"
  repeated MenuItem menus = 5;       // 菜单声明
  repeated Permission perms = 6;     // 权限声明
  string frontend_bundle = 7;        // 前端 bundle 相对路径
}

message MenuItem {
  string title = 1;
  string icon = 2;
  string path = 3;
  int32 sort = 4;
  repeated MenuItem children = 5;
}

message Permission {
  string key = 1;       // 如 "game:dlc:list"
  string title = 2;
}

// HTTP 请求/响应
message HttpRequest {
  string method = 1;
  string path = 2;
  map<string, string> headers = 3;
  bytes body = 4;
  string query = 5;
  // 上下文注入
  int64 user_id = 10;
  int64 tenant_id = 11;
  repeated string roles = 12;
  repeated string permissions = 13;
  string db_dsn = 14;              // 数据库连接串
}

message HttpResponse {
  int32 status_code = 1;
  map<string, string> headers = 2;
  bytes body = 3;
}

// 插件间通信
message CallPluginRequest {
  string target_plugin = 1;
  string method = 2;
  bytes payload = 3;
}

message CallPluginResponse {
  bytes payload = 1;
  string error = 2;
}

// 事件
message Event {
  string source = 1;
  string type = 2;
  bytes payload = 3;
  int64 timestamp = 4;
}

message Empty {}

// 插件必须实现的服务
service PluginService {
  rpc Register(Empty) returns (PluginInfo);
  rpc HandleRequest(HttpRequest) returns (HttpResponse);
  rpc Healthcheck(Empty) returns (HealthResponse);
  rpc OnEvent(Event) returns (Empty);                    // 接收事件
  rpc CallPlugin(CallPluginRequest) returns (CallPluginResponse); // 调用其他插件
}

message HealthResponse {
  bool healthy = 1;
  string message = 2;
}

// Host 提供给插件的服务（反向调用）
service HostService {
  rpc PublishEvent(Event) returns (Empty);               // 发布事件
  rpc CallPlugin(CallPluginRequest) returns (CallPluginResponse); // 通过 Host 中转调用其他插件
}
```

### 数据库设计

```sql
-- 插件注册表（Host 管理）
CREATE TABLE sys_plugin (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE COMMENT '插件名称',
  version VARCHAR(32) NOT NULL COMMENT '版本号',
  description VARCHAR(512) COMMENT '描述',
  status INT DEFAULT 0 COMMENT '状态 0=已安装 1=运行中 2=已停止 3=异常',
  binary_path VARCHAR(256) NOT NULL COMMENT '二进制文件路径',
  frontend_path VARCHAR(256) COMMENT '前端bundle路径',
  config TEXT COMMENT '插件配置JSON',
  installed_at DATETIME COMMENT '安装时间',
  create_by INT DEFAULT 0,
  update_by INT DEFAULT 0,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 核心逻辑

#### 插件管理器生命周期

```
Install → (解压到 plugins/{name}/) → 记录 sys_plugin
Start   → go-plugin.Client 启动子进程 → gRPC Register → 注册路由/菜单/权限 → 复制前端bundle到static
Stop    → gRPC 优雅关闭 → 注销路由/菜单/权限
Uninstall → Stop → 删除文件 → 可选清除数据表 → 删除 sys_plugin 记录
```

#### HTTP 请求代理流程

```
Client → Gin Router → /api/v1/plugin/{name}/* 匹配
  → proxy middleware 查找 registry 中对应插件的 gRPC client
  → 构造 HttpRequest（注入 user_id/tenant_id/roles/permissions/db_dsn）
  → 调用 PluginService.HandleRequest
  → 将 HttpResponse 写回 gin.Context
```

#### 插件间通信

```
同步调用：Plugin A → Host.CallPlugin(target="B", method, payload) → Host 查找 B 的 client → B.CallPlugin → 返回
异步事件：Plugin A → Host.PublishEvent(event) → Host EventBus 广播 → 所有订阅该 type 的插件收到 OnEvent
```

#### 前端动态加载

```
1. Host 前端启动时请求 GET /api/v1/plugins（获取运行中插件列表）
2. 遍历插件，对每个有 frontend_bundle 的插件：
   - dynamic import(`/static/plugins/{name}/index.js`)
   - 插件 bundle 导出 { routes, menus }
   - 注册 routes 到 Vue Router
   - 注入 menus 到侧边栏
3. 插件停止时，从 Router 和菜单中移除
```

### 文件结构变更

```
backend/
├── common/plugin/              # 新增：插件框架
│   ├── manager.go              # 插件管理器
│   ├── registry.go             # 路由/菜单/权限注册表
│   ├── proxy.go                # HTTP→gRPC 代理中间件
│   ├── event_bus.go            # 事件总线
│   ├── installer.go            # 安装/卸载逻辑
│   └── types.go                # 类型定义
├── common/plugin/proto/        # 新增：protobuf 定义
│   └── plugin.proto
├── app/plugin/                 # 新增：插件管理 API
│   ├── apis/plugin.go
│   ├── models/sys_plugin.go
│   ├── service/plugin.go
│   └── router/router.go
├── plugins/                    # 插件二进制存放目录
└── static/plugins/             # 插件前端 bundle 存放目录

frontend/
├── src/utils/plugin-loader.ts  # 新增：插件前端加载器
├── src/store/modules/plugin.ts # 新增：插件状态管理
└── src/views/admin/plugin/     # 新增：插件管理页面
    └── index.vue

scaffold/                       # 新增：插件开发脚手架 CLI
├── main.go
├── templates/
│   ├── backend/
│   └── frontend/
└── README.md
```

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有用户/角色/菜单/字典 CRUD 不受影响 | 管理功能正常 |
| RG-2 | 现有登录/JWT 认证流程不变 | 登录正常返回 token |
| RG-3 | 租户隔离不被破坏 | 不同租户数据不互通 |
| RG-4 | 无插件运行时 Host 功能完全正常 | 空插件目录启动无报错 |
| RG-5 | 现有 game 模块功能不受影响（渐进式改造） | game API 正常 |

## 正确性属性

- 插件进程隔离：插件崩溃不影响 Host，go-plugin 自动检测进程退出
- 路由隔离：插件只能处理自己前缀下的请求，无法越权访问其他插件或 Host 内部路由
- 数据隔离：插件表名强制 `{pluginName}_` 前缀，Host 不会误操作插件数据
- 权限集成：插件声明的权限自动注册到 Host RBAC（casbin），遵循统一权限模型
- 事件最终一致：异步事件通过内存 EventBus 广播，插件离线时事件丢失（可接受）
