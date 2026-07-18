# 技术设计文档：插件化架构规范（Plugin Architecture）

## 概述

将系统从单体架构升级为"最小平台框架 + 插件化扩展"架构。设计覆盖三端：Go 后端、dev-web-admin（管理端）、dev-web-user（用户端）。

设计核心原则：
- 统一 Plugin Interface + 适配器模式（LocalAdapter 当前实现，RemoteAdapter 预留）
- 前端双模式加载（源码编译 + 运行时动态 import）
- DB 为插件状态真相源，事件广播保证集群一致性
- 框架面向接口编程，后续扩展为增量适配

## 架构

### 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        Platform Shell                            │
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌────────────────┐ │
│  │  Go Backend     │  │  dev-web-admin   │  │  dev-web-user  │ │
│  │  (Plugin Host)  │  │  (Admin Shell)   │  │  (User Shell)  │ │
│  └────────┬────────┘  └────────┬────────┘  └───────┬────────┘ │
│           │                     │                    │          │
│           ▼                     ▼                    ▼          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Plugin Manager (统一调度层)                   │  │
│  │                                                          │  │
│  │  ┌──────────────┐        ┌──────────────────┐          │  │
│  │  │ LocalAdapter │        │  RemoteAdapter   │          │  │
│  │  │ (内嵌编译)    │        │ (gRPC, 预留)     │          │  │
│  │  └──────┬───────┘        └───────┬──────────┘          │  │
│  │         │                         │                      │  │
│  │         ▼                         ▼                      │  │
│  │  ┌──────────────────────────────────────────────────┐   │  │
│  │  │           Unified Plugin Interface               │   │  │
│  │  │  Name() / Version() / Manifest()                 │   │  │
│  │  │  RegisterRoutes() / Permissions()                │   │  │
│  │  │  OnEnable() / OnDisable()                        │   │  │
│  │  │  Migrate(up/down)                                │   │  │
│  │  └──────────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Plugin Registry (DB)                    │  │
│  │  sys_plugin / sys_plugin_version / sys_tenant_plugin      │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 前端插件加载架构

```
┌────────────────────────────────────────────────────┐
│                Frontend Shell                       │
│                                                    │
│  ┌──────────────────────────────────────────────┐ │
│  │           Plugin Loader                       │ │
│  │                                              │ │
│  │  ┌─────────────────┐  ┌──────────────────┐  │ │
│  │  │ Source Mode      │  │ Runtime Mode     │  │ │
│  │  │ src/plugins/xxx  │  │ /static/plugins/ │  │ │
│  │  │ (构建时包含)      │  │ (动态 import)    │  │ │
│  │  └────────┬─────────┘  └───────┬──────────┘  │ │
│  │           │                     │             │ │
│  │           ▼                     ▼             │ │
│  │  ┌──────────────────────────────────────┐    │ │
│  │  │     definePlugin() 标准输出           │    │ │
│  │  │  { routes, menus, permissions, ... } │    │ │
│  │  └──────────────────────────────────────┘    │ │
│  └──────────────────────────────────────────────┘ │
│                         │                          │
│                         ▼                          │
│  ┌─────────────┐ ┌────────────┐ ┌─────────────┐ │
│  │ Router合并   │ │ Menu合并    │ │ EventBus    │ │
│  └─────────────┘ └────────────┘ └─────────────┘ │
└────────────────────────────────────────────────────┘
```

## 组件与接口

### 后端 Plugin Interface

```go
// plugin-sdk/plugin.go

package pluginsdk

import (
    "context"
    "github.com/gin-gonic/gin"
)

// Plugin 统一插件接口
type Plugin interface {
    // 元数据
    Name() string
    Version() string
    Manifest() Manifest

    // 生命周期
    OnInstall(ctx context.Context) error
    OnEnable(ctx context.Context) error
    OnDisable(ctx context.Context) error
    OnUninstall(ctx context.Context) error

    // 路由注册
    RegisterRoutes(group *gin.RouterGroup)

    // 权限声明
    Permissions() []Permission

    // 数据库迁移
    MigrateUp(ctx context.Context) error
    MigrateDown(ctx context.Context) error
}

// Manifest 插件描述
type Manifest struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    DisplayName  string            `json:"displayName"`
    Description  string            `json:"description"`
    Dependencies map[string]string `json:"dependencies"` // name -> semver
    Platforms    []string          `json:"platforms"`     // admin/user/both
    FrontendEntry string           `json:"frontendEntry"` // 前端入口路径
}

// Permission 权限声明
type Permission struct {
    Code        string `json:"code"`
    DisplayName string `json:"displayName"`
    Group       string `json:"group"`
}
```

### Plugin Manager

```go
// plugin-sdk/manager.go

package pluginsdk

// PluginAdapter 适配器接口（支持本地和远程）
type PluginAdapter interface {
    Load(name string) (Plugin, error)
    Unload(name string) error
}

// LocalAdapter 内嵌插件适配器（当前实现）
type LocalAdapter struct {
    plugins map[string]Plugin
}

func (a *LocalAdapter) Register(p Plugin) {
    a.plugins[p.Name()] = p
}

func (a *LocalAdapter) Load(name string) (Plugin, error) {
    p, ok := a.plugins[name]
    if !ok {
        return nil, fmt.Errorf("plugin %s not registered", name)
    }
    return p, nil
}

// RemoteAdapter gRPC 远程插件适配器（预留，不实现）
// type RemoteAdapter struct { ... }

// Manager 插件管理器
type Manager struct {
    adapters []PluginAdapter
    registry PluginRegistry // DB 操作接口
    eventBus EventPublisher // Redis Pub/Sub 或 noop
}

func (m *Manager) EnablePlugin(ctx context.Context, name string) error {
    // 1. 从适配器加载插件
    // 2. 执行 OnEnable
    // 3. 注册路由
    // 4. 注册权限
    // 5. 更新 DB 状态
    // 6. 广播变更事件
}
```

### 前端 Plugin SDK

```typescript
// src/plugin-sdk/index.ts

export interface PluginManifest {
  name: string
  version: string
  displayName: string
  description: string
  platforms: ('admin' | 'user' | 'both')[]
}

export interface PluginRoute {
  path: string
  name: string
  component: () => Promise<any>
  meta: { title: string; icon?: string; permission?: string }
}

export interface PluginMenu {
  title: string
  icon: string
  path: string
  sort: number
  group?: string
  permission?: string
}

export interface PluginConfig {
  manifest: PluginManifest
  routes?: PluginRoute[]
  menus?: PluginMenu[]
  permissions?: string[]
  extensions?: Record<string, any>
  setup?: (ctx: PluginContext) => void | Promise<void>
}

export interface PluginContext {
  currentUser: Readonly<UserInfo>
  currentTenant: Readonly<TenantInfo>
  permissions: Readonly<string[]>
  eventBus: EventBus
  registerExtension: (point: string, component: any, sort?: number) => void
}

export interface EventBus {
  emit: (event: string, payload?: any) => void
  on: (event: string, handler: (payload: any) => void) => void
  off: (event: string, handler?: (payload: any) => void) => void
}

/** 插件入口函数 */
export function definePlugin(config: PluginConfig) {
  return config
}

/** 框架注入的插件上下文 */
export function usePluginContext(): PluginContext {
  // 由框架 provide/inject 实现
}
```

### 前端 Plugin Loader

```typescript
// src/core/plugin-loader.ts

import type { PluginConfig } from '@/plugin-sdk'

interface LoadedPlugin {
  name: string
  config: PluginConfig
}

class PluginLoader {
  private loaded: Map<string, LoadedPlugin> = new Map()

  /** 加载源码模式插件 */
  async loadSource(name: string): Promise<PluginConfig> {
    // Vite glob import: src/plugins/{name}/index.ts
    const modules = import.meta.glob('@/plugins/*/index.ts')
    const entry = modules[`/src/plugins/${name}/index.ts`]
    if (!entry) throw new Error(`Plugin ${name} not found in source`)
    const mod = await entry()
    return mod.default || mod
  }

  /** 加载运行时模式插件 */
  async loadRuntime(name: string): Promise<PluginConfig> {
    const mod = await import(/* @vite-ignore */ `/static/plugins/${name}/index.js`)
    return mod.default || mod
  }

  /** 统一加载入口 */
  async load(name: string, mode: 'source' | 'runtime'): Promise<PluginConfig> {
    if (this.loaded.has(name)) return this.loaded.get(name)!.config
    const config = mode === 'source'
      ? await this.loadSource(name)
      : await this.loadRuntime(name)
    this.loaded.set(name, { name, config })
    return config
  }

  /** 批量加载已启用插件 */
  async loadAll(enabledPlugins: Array<{ name: string; mode: 'source' | 'runtime' }>) {
    const results: PluginConfig[] = []
    for (const p of enabledPlugins) {
      try {
        const config = await this.load(p.name, p.mode)
        results.push(config)
      } catch (e) {
        console.error(`[PluginLoader] Failed to load ${p.name}:`, e)
      }
    }
    return results
  }
}

export const pluginLoader = new PluginLoader()
```

### 前端动态路由与菜单合并

```typescript
// src/core/plugin-router.ts

import { type Router } from 'vue-router'
import type { PluginConfig, PluginRoute } from '@/plugin-sdk'

/** 将插件路由注册到 Vue Router */
export function mergePluginRoutes(router: Router, plugins: PluginConfig[]) {
  for (const plugin of plugins) {
    if (!plugin.routes) continue
    for (const route of plugin.routes) {
      router.addRoute('app-layout', {
        path: route.path,
        name: route.name,
        component: route.component,
        meta: { ...route.meta, pluginName: plugin.manifest.name }
      })
    }
  }
}

/** 将插件菜单合并到侧边栏 */
export function mergePluginMenus(plugins: PluginConfig[], userPermissions: string[]) {
  const menus: PluginMenu[] = []
  for (const plugin of plugins) {
    if (!plugin.menus) continue
    for (const menu of plugin.menus) {
      // 权限过滤
      if (menu.permission && !userPermissions.includes(menu.permission)) continue
      menus.push(menu)
    }
  }
  return menus.sort((a, b) => a.sort - b.sort)
}
```

### EventBus 实现

```typescript
// src/core/event-bus.ts
import mitt from 'mitt'
import type { EventBus } from '@/plugin-sdk'

const emitter = mitt()

export const pluginEventBus: EventBus = {
  emit: (event, payload) => emitter.emit(event, payload),
  on: (event, handler) => emitter.on(event, handler),
  off: (event, handler) => emitter.off(event, handler)
}
```

## 数据模型

### 数据库表设计

```sql
-- 插件注册表
CREATE TABLE sys_plugin (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  name        VARCHAR(100) NOT NULL UNIQUE,
  version     VARCHAR(50) NOT NULL,
  display_name VARCHAR(200),
  description TEXT,
  status      TINYINT NOT NULL DEFAULT 0, -- 0=disabled, 1=enabled
  config      JSON,
  frontend_entry VARCHAR(500),
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 插件版本历史
CREATE TABLE sys_plugin_version (
  id            BIGINT PRIMARY KEY AUTO_INCREMENT,
  plugin_name   VARCHAR(100) NOT NULL,
  version       VARCHAR(50) NOT NULL,
  snapshot_path VARCHAR(500),  -- 旧版本文件快照路径
  migration_ver INT NOT NULL DEFAULT 0,  -- 数据库迁移版本号
  installed_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_name_version (plugin_name, version)
);

-- 租户-插件关联
CREATE TABLE sys_tenant_plugin (
  id         BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id  BIGINT NOT NULL,
  plugin_name VARCHAR(100) NOT NULL,
  enabled    TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_tenant_plugin (tenant_id, plugin_name)
);

-- 插件权限注册表
CREATE TABLE sys_plugin_permission (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  plugin_name VARCHAR(100) NOT NULL,
  code        VARCHAR(200) NOT NULL,
  display_name VARCHAR(200),
  group_name  VARCHAR(100),
  UNIQUE KEY uk_plugin_code (plugin_name, code)
);
```

### 插件 Manifest 完整格式

```json
{
  "name": "dict-manager",
  "version": "1.0.0",
  "displayName": "字典管理",
  "description": "系统字典类型和字典数据管理",
  "platforms": ["admin"],
  "dependencies": {},
  "frontendEntry": "dict-manager/index.js",
  "permissions": [
    { "code": "system:dict:list", "displayName": "字典列表", "group": "字典管理" },
    { "code": "system:dict:create", "displayName": "创建字典", "group": "字典管理" },
    { "code": "system:dict:update", "displayName": "编辑字典", "group": "字典管理" },
    { "code": "system:dict:delete", "displayName": "删除字典", "group": "字典管理" }
  ],
  "menus": [
    { "title": "字典管理", "icon": "Collection", "path": "/system/dict", "sort": 50, "group": "系统管理", "permission": "system:dict:list" }
  ],
  "routes": [
    { "path": "system/dict", "name": "system-dict", "component": "./views/DictTypeList.vue", "meta": { "title": "字典管理", "icon": "Collection", "permission": "system:dict:list" } }
  ],
  "migrations": ["001_create_dict_tables.sql"]
}
```

### 插件目录结构（源码模式）

```
src/plugins/dict-manager/
├── manifest.json          # 插件描述
├── index.ts               # 入口（definePlugin）
├── api/
│   └── dict.ts            # API 封装
├── views/
│   ├── DictTypeList.vue   # 页面组件
│   ├── DictTypeForm.vue
│   ├── DictDataList.vue
│   └── DictDataForm.vue
└── migrations/
    ├── 001_create_dict_tables.up.sql
    └── 001_create_dict_tables.down.sql
```

## 错误处理

| 场景 | 处理方式 |
|------|----------|
| 插件加载失败 | console.error 记录，跳过该插件，不影响框架和其他插件 |
| 插件 manifest 格式错误 | 拒绝安装/启用，返回具体校验错误 |
| 插件依赖不满足 | 拒绝启用，提示缺少哪些依赖 |
| 插件 migration up 失败 | 自动执行 down 回滚，标记安装失败 |
| 插件 migration down 失败 | 记录错误日志，标记为"需人工干预" |
| Redis 广播失败 | 兜底定时轮询（30s），记录 warning 日志 |
| 远程插件连接失败（未来） | 标记为不可用，不注册路由，定时重连 |

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 框架核心功能（登录/用户/角色/菜单/日志）不依赖任何插件 | 禁用所有插件后核心功能正常 |
| RG-2 | 单个插件故障不影响框架和其他插件 | 故意制造插件加载错误，验证其他功能正常 |
| RG-3 | 已有 admin-frontend-refactor 的页面继续可用 | 重构前后页面功能一致 |
| RG-4 | 多租户数据隔离不被破坏 | 不同租户只能看到自己授权的插件 |
| RG-5 | 插件禁用后其路由不可访问 | 禁用插件后访问其 URL 返回 404 |

## 实施阶段划分

| 阶段 | 内容 | 产出 |
|------|------|------|
| Phase 1 | 后端 Plugin Interface + LocalAdapter + Manager + DB 表 | 后端插件 SDK 可用 |
| Phase 2 | 前端 Plugin SDK + Loader + 动态路由/菜单合并 + EventBus | 前端插件加载机制可用 |
| Phase 3 | 管理端插件管理页面（安装/启用/禁用/版本管理） | 可视化插件管理 |
| Phase 4 | dev-web-user 框架瘦身 + 插件容器 + 响应式布局 | 用户端插件化就绪 |
| Phase 5 | 将现有模块（部门/岗位/字典等）拆为插件 | 验证架构可行性 |
| Phase 6（未来） | RemoteAdapter 实现 + 集群 Redis 广播 | 支持远程插件热加载 |
