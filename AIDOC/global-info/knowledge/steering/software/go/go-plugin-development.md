# Go 插件开发规范

## 架构概述

本底座采用 **hashicorp/go-plugin** 实现插件子进程隔离：
- 插件编译为独立二进制，通过 stdio gRPC 与 Host 通信
- 插件崩溃不影响 Host
- 支持运行时安装/启动/停止/卸载/升级
- 菜单通过 plugin.json 声明，PluginManager 启动时自动同步到 admin_resource 表
- 支持插件间 Action 调用（带版本兼容校验）
- 支持安全升级与自动回滚

## 插件工程结构

```
projects/{solution-name}/{project-name}/plugins/{plugin-name}/
├── go.mod              # module game-server/plugins/{plugin-name}
├── main.go             # 插件入口，实现 PluginService 接口
├── handler.go          # HTTP 请求路由分发
├── db.go               # 数据库初始化（sync.Once + AutoMigrate）
├── models.go           # 数据模型（自定义基础结构，不依赖 Host）
├── service_*.go        # 各子模块业务逻辑
├── plugin.json         # 插件描述文件（name/version/description）
└── frontend/           # 插件前端工程（独立 Vite 项目）
    ├── package.json
    ├── vite.config.ts  # ES module 格式，vue/element-plus external + shim 映射
    ├── tsconfig.json
    └── src/
        ├── index.ts    # 导出 { routes, menus }
        ├── env.d.ts
        └── views/      # 页面组件
```

## plugin.json 格式（V2）

```json
{
  "name": "game",
  "version": "2.1.0",
  "displayName": "游戏管理",
  "description": "游戏管理插件，提供游戏/DLC/CDK等完整管理能力",
  "routePrefix": "game",
  "platforms": ["windows/amd64", "linux/amd64"],
  "modules": ["game", "dlc", "cdk"],
  "frontends": [
    { "entry": "frontend/dist/index.js", "framework": "vue3" }
  ],
  "menus": [
    {
      "title": "游戏管理",
      "icon": "ep:box",
      "path": "/plugin/game",
      "sort": 10,
      "children": [
        { "title": "游戏列表", "icon": "ep:list", "path": "/plugin/game/list", "sort": 1 },
        { "title": "DLC管理", "icon": "ep:folder", "path": "/plugin/game/dlc", "sort": 2 }
      ]
    }
  ],
  "apiPermissions": [
    { "key": "game:list", "title": "游戏列表", "method": "GET", "path": "/list" },
    { "key": "game:create", "title": "创建游戏", "method": "POST", "path": "/game" }
  ],
  "exposedActions": [
    { "name": "GetGameInfo", "version": "1.0.0", "description": "获取游戏详情" },
    { "name": "SyncInventory", "version": "2.0.0", "description": "同步库存数据" }
  ],
  "subscribedEvents": ["tenant.created", "user.deleted"],
  "breakingUpgrade": false,
  "minUpgradeFrom": "1.5.0",
  "migrationNotes": "v2.0 重构了 game_game 表结构，需执行数据迁移"
}
```

### 字段说明

| 字段 | 必填 | 说明 |
|------|------|------|
| name | ✅ | 插件唯一标识，英文 kebab-case |
| version | ✅ | 语义化版本号 |
| displayName | ✅ | 在管理界面显示的中文名称 |
| description | ❌ | 插件功能描述 |
| routePrefix | ✅ | API 路由前缀，`/api/v1/plugin/{routePrefix}/*` |
| platforms | ❌ | 支持的编译目标平台 |
| modules | ❌ | 插件内部子模块列表（信息性） |
| frontends | ❌ | 前端入口声明 |
| menus | ✅ | 菜单声明（V2 替代 Register 中的菜单返回） |
| apiPermissions | ❌ | API 权限声明，同步到 admin_resource 表 |
| exposedActions | ❌ | 对外暴露的 Action 及版本 |
| subscribedEvents | ❌ | 订阅的 Host 事件列表 |
| breakingUpgrade | ❌ | 是否为破坏性升级（默认 false） |
| minUpgradeFrom | ❌ | 允许直接升级的最低源版本 |
| migrationNotes | ❌ | 升级注意事项/迁移说明 |

**重要**：安装时后端自动从 plugin.json 提取插件名，用户不需要手动填写。同名插件不允许重复安装。

## 后端开发规范

### main.go 模板

```go
package main

import (
    "context"
    "game-server/plugin-sdk/helper"
    "game-server/plugin-sdk/proto"
)

type MyPlugin struct{}

func (p *MyPlugin) Register(ctx context.Context) (*proto.PluginInfo, error) {
    return &proto.PluginInfo{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "我的插件",
        RoutePrefix: "my-plugin",
        Menus: []*proto.MenuItem{
            {
                Title: "我的功能",
                Icon:  "ep:box",
                Path:  "/plugin/my-plugin",
                Sort:  10,
                Children: []*proto.MenuItem{
                    {Title: "功能A", Icon: "ep:list", Path: "/plugin/my-plugin/a", Sort: 1},
                    {Title: "功能B", Icon: "ep:edit", Path: "/plugin/my-plugin/b", Sort: 2},
                },
            },
        },
        Perms: []*proto.Permission{
            {Key: "my-plugin:list", Title: "列表"},
        },
    }, nil
}

func (p *MyPlugin) HandleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
    return handleRequest(ctx, req)
}

func (p *MyPlugin) Healthcheck(ctx context.Context) (*proto.HealthResponse, error) {
    return &proto.HealthResponse{Healthy: true, Message: "ok"}, nil
}

func (p *MyPlugin) OnEvent(ctx context.Context, event *proto.Event) error {
    return nil
}

func (p *MyPlugin) CallPlugin(ctx context.Context, req *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
    return &proto.CallPluginResponse{Payload: []byte(`{}`)}, nil
}

func main() {
    helper.Serve(&MyPlugin{})
}
```

### 菜单注册机制

#### ⚠️ 旧方式（已废弃）

> 以下方式在 V1 中使用，V2 已废弃。保留说明仅供参考。

- 插件在 `Register()` 中通过 gRPC 返回菜单树
- Host 启动插件后写入 `sys_menu` 表
- 调用 `RegisterMenus` 方法同步

**废弃原因**：命令式注册存在幂等性问题、事务不安全、插件未启动时菜单丢失。

#### 新方式（V2 声明式）

V2 中菜单通过 `plugin.json` 的 `menus` 字段声明，PluginManager 启动时自动同步：

1. PluginManager 在 `SyncOnStart` 阶段读取所有已安装插件的 `plugin.json`
2. 解析 `menus` 字段，与 `admin_resource` 表对比差异
3. 在同一事务中完成新增/更新/软删除操作
4. 菜单统一挂在"扩展功能"（PluginExtensions）父节点下

**好处**：
- **声明式**：菜单结构即配置，无需编写注册代码
- **幂等**：多次启动结果一致，不产生重复记录
- **事务安全**：菜单同步在单事务中完成，失败自动回滚
- **离线可见**：即使插件未运行，菜单结构仍可从 plugin.json 读取

**注意**：`Register()` 方法仍保留用于返回 PluginInfo（名称、版本等元信息），但菜单和权限声明以 plugin.json 为准。

### 数据库规范

- 插件通过 `req.DbDsn` 获取数据库连接串
- 使用 `sync.Once` 确保只初始化一次
- 表名使用 `{plugin-name}_` 前缀（如 `game_game`、`game_dlc`）
- 插件自行 AutoMigrate，不依赖 Host 迁移
- 不依赖 `go-admin/common/models`，自定义 Model/ModelTime/ControlBy/TenantBy 基础结构

### 请求处理

- Host 代理 `/api/v1/plugin/{name}/*` 到插件的 `HandleRequest`
- `req.Path` 为去掉前缀后的路径（如 `/list`、`/game/1`）
- `req.UserId`、`req.TenantId`、`req.Roles`、`req.Permissions` 由 Host 自动注入
- 响应格式统一：`{"code":200,"msg":"ok","data":...}`
- 所有查询必须带 `tenant_id = req.TenantId` 租户隔离

### protobuf 字段名注意

protobuf 生成的 Go struct 字段名：
- `UserId`（不是 UserID）
- `TenantId`（不是 TenantID）
- `StatusCode`（不是 StatusCode）
- `DbDsn`（不是 DbDSN）

## 前端开发规范

### 构建配置

使用 Vite library 模式，vue/element-plus 作为 external，通过 shim 共享 Host 实例：

```typescript
// vite.config.ts
export default defineConfig({
  plugins: [vue()],
  define: {
    "process.env.NODE_ENV": JSON.stringify("production"),
    "process.env": JSON.stringify({})
  },
  build: {
    lib: {
      entry: "src/index.ts",
      name: "MyPlugin",
      fileName: () => "index.js",
      formats: ["es"]
    },
    rollupOptions: {
      external: ["vue", "element-plus"],
      output: {
        paths: {
          vue: "/plugin-shims/vue.js",
          "element-plus": "/plugin-shims/element-plus.js"
        }
      }
    },
    outDir: "dist"
  }
});
```

### shim 机制

Host 前端在 `main.ts` 中将 Vue 和 Element Plus 挂载到 `window`：
```typescript
import * as Vue from "vue";
import * as ElementPlus from "element-plus";
(window as any).__HOST_VUE__ = Vue;
(window as any).__HOST_ELEMENT_PLUS__ = ElementPlus;
```

Host 的 `public/plugin-shims/vue.js` 从 window 重新导出所有 Vue API。
插件 bundle 中的 `import { ref } from "vue"` 会被重定向到 `/plugin-shims/vue.js`，从而共享 Host 的 Vue 实例。

### index.ts 导出格式

```typescript
import PageA from "./views/PageA.vue";
import PageB from "./views/PageB.vue";

export const routes = [
  { path: "/plugin/my-plugin/a", name: "MyPluginA", component: PageA, meta: { title: "功能A" } },
  { path: "/plugin/my-plugin/b", name: "MyPluginB", component: PageB, meta: { title: "功能B" } },
];

export const menus = []; // 菜单由后端 sys_menu 管理，前端不需要导出
```

### 页面组件规范

- 使用原生 `fetch` 发请求（不依赖 Host 的 axios 封装）
- API 前缀：`/api/v1/plugin/{plugin-name}/`
- token 从 `localStorage.getItem("user-info")` 获取
- Element Plus 组件正常使用（通过 shim 共享）
- 路由路径必须与 `Register()` 中声明的菜单 path 一致

### container.vue 机制

Host 前端有一个 `views/plugin/container.vue` 容器页面：
- sys_menu 中插件菜单的 component 统一为 `plugin/container`
- 用户点击菜单时加载 container.vue
- container.vue 根据当前路由路径从 `/static/plugins/{name}/index.js` 加载插件 bundle
- 匹配 `routes` 数组中 path 对应的组件进行渲染

## 插件升级

### 升级相关字段

| 字段 | 类型 | 说明 |
|------|------|------|
| breakingUpgrade | bool | 标记为破坏性升级时，Host 在升级前向管理员发出警告确认 |
| minUpgradeFrom | string | 允许直接升级的最低源版本。低于此版本需先升级到中间版本 |
| migrationNotes | string | 升级说明，显示在管理界面的升级确认弹窗中 |

### 版本校验规则

- `minUpgradeFrom` 为空时，允许从任意旧版本直接升级
- 当前安装版本 < `minUpgradeFrom` 时，拒绝升级并提示需先升级到中间版本
- `breakingUpgrade = true` 时强制管理员二次确认

### 升级流程

```
上传新版 zip
    ↓
校验：解析 plugin.json → 版本比较 → minUpgradeFrom 检查
    ↓
安全文件替换：
  1. 新文件写入 {plugin-dir}/{plugin-name}.exe.new
  2. 当前文件重命名为 {plugin-name}.exe.bak
  3. .new 重命名为 {plugin-name}.exe
  4. 前端 bundle 同样执行 .new/.bak 替换
    ↓
重启插件进程
    ↓
SyncOnStart：同步 plugin.json 中的菜单/权限到数据库
    ↓
升级完成，清理 .bak 文件
```

### 回滚机制

- 插件进程启动失败（gRPC 握手超时或 Healthcheck 返回 unhealthy）时自动触发回滚
- 回滚步骤：停止新进程 → 恢复 .bak 文件 → 重启旧版本 → 记录回滚日志
- 管理界面显示升级失败原因和回滚状态

### Host 中途崩溃恢复（RecoverStaleUpgrade）

Host 启动时调用 `RecoverStaleUpgrade` 检测未完成的升级：

1. 扫描所有插件目录，查找残留的 `.new` 或 `.bak` 文件
2. 若存在 `.new` 文件（替换未完成）：删除 `.new`，保持当前版本不变
3. 若存在 `.bak` 且当前 exe 不可用：恢复 `.bak` 为当前版本
4. 记录恢复日志，标记插件状态为 `needs_attention`

## Action 版本声明

### exposedActions 字段格式

在 `plugin.json` 中声明插件对外暴露的 Action：

```json
{
  "exposedActions": [
    {
      "name": "GetGameInfo",
      "version": "1.0.0",
      "description": "获取游戏详情"
    },
    {
      "name": "SyncInventory",
      "version": "2.0.0",
      "description": "同步库存数据"
    }
  ]
}
```

### 版本兼容规则

采用语义化版本的 **major 兼容原则**：

- major 相同即视为兼容（如 `1.0.0` 与 `1.3.2` 兼容）
- major 不同则不兼容（如 `1.x` 与 `2.x` 不兼容）
- 调用方可指定所需的 Action 版本进行兼容性校验

### CallPlugin 调用时的版本校验

其他插件或 Host 通过 `CallPlugin` 调用时，可携带 `ActionVersion` 字段：

```go
resp, err := hostClient.CallPlugin(ctx, &proto.CallPluginRequest{
    TargetPlugin: "game",
    Action:       "GetGameInfo",
    ActionVersion: "1.0.0",  // 可选：期望的 Action 版本
    Payload:      payload,
})
```

校验逻辑：
- `ActionVersion` 为空：不校验，直接调用
- `ActionVersion` 非空：比较调用方期望版本与目标插件声明版本的 major
  - major 相同 → 允许调用
  - major 不同 → 返回 `ErrActionVersionIncompatible`，附带当前版本信息

## 打包规范

### 构建流程

```powershell
# 1. 编译后端二进制
cd plugins/{plugin-name}
go build -o {plugin-name}.exe .

# 2. 构建前端 bundle
cd frontend
pnpm install
pnpm build
# 产出 frontend/dist/index.js

# 3. 打包为 zip
# zip 包内容必须包含：
#   {plugin-name}.exe
#   plugin.json
#   frontend/dist/index.js
```

### 打包产物路径

```
projects/{solution-name}/{project-name}/
├── dist/
│   ├── {project-name}.exe          # Host 主程序
│   └── plugins/                    # 插件包目录
│       ├── game-plugin.zip
│       └── ...
```

### zip 包内部结构（必须）

```
{plugin-name}-plugin.zip
├── {plugin-name}.exe       # 插件二进制
├── plugin.json             # 插件描述文件（必须包含 name 字段）
└── frontend/
    └── dist/
        └── index.js        # 前端 bundle
```

## 安装后目录结构

通过管理界面上传 zip 后，Host 自动解压到：

```
{运行目录}/
├── plugins/{plugin-name}/          # 插件二进制 + plugin.json
│   ├── {plugin-name}.exe
│   ├── plugin.json
│   └── frontend/dist/index.js
└── static/plugins/{plugin-name}/   # 前端 bundle（复制）
    └── index.js
```

## 完整生命周期

```
上传 zip → 解压 → 读取 plugin.json → 检查重复/版本 → 记录 sys_plugin
    ↓
SyncOnStart → 同步 menus 到 admin_resource → 同步 apiPermissions
    ↓
启动 → go-plugin Client 启动子进程 → gRPC Register → 获取元信息
    ↓
运行中 → HTTP 代理 /api/v1/plugin/{name}/* → HandleRequest
       → 接收 Host 事件 → OnEvent
       → 响应其他插件调用 → CallPlugin（带版本校验）
    ↓
升级 → 上传新包 → 校验版本 → 安全替换(.new/.bak) → 重启 → SyncOnStart
    ↓
停止 → Kill 子进程 → 更新 sys_plugin 状态（菜单保留，标记为不可用）
    ↓
卸载 → Kill 进程 → 删除文件 → 清理 admin_resource → 删除 sys_plugin 记录
```
