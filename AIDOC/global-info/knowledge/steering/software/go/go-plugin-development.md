# Go 插件开发规范

## 架构概述

本底座采用 **hashicorp/go-plugin** 实现插件子进程隔离：
- 插件编译为独立二进制，通过 stdio gRPC 与 Host 通信
- 插件崩溃不影响 Host
- 支持运行时安装/启动/停止/卸载
- 菜单在插件启动时自动注册到 sys_menu 表，挂在"扩展功能"目录下

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

## plugin.json 格式

```json
{
  "name": "game",
  "version": "1.0.0",
  "description": "游戏管理插件"
}
```

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

- 插件在 `Register()` 中声明菜单树（支持多级）
- Host 启动插件后自动将菜单写入 `sys_menu` 表
- 菜单统一挂在"扩展功能"（PluginExtensions）目录下
- 停止/卸载时自动注销菜单（软删除）
- 菜单 component 统一为 `plugin/container`（Host 前端的插件容器页面）

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
上传 zip → 解压 → 读取 plugin.json → 检查重复 → 记录 sys_plugin
    ↓
启动 → go-plugin Client 启动子进程 → gRPC Register → 获取菜单 → 写入 sys_menu
    ↓
运行中 → HTTP 代理 /api/v1/plugin/{name}/* → HandleRequest
    ↓
停止 → Kill 子进程 → 注销 sys_menu → 更新 sys_plugin 状态
    ↓
卸载 → Kill 进程 → 删除文件 → 注销菜单 → 删除 sys_plugin 记录
```
