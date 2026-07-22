# 设计：V2-CR4 插件系统统一 RBAC + 通信契约

## 技术方案

### plugin.json Manifest V2 格式

```json
{
  "name": "billing",
  "version": "1.2.0",
  "displayName": "计费管理",
  "description": "发票、支付、退款管理",
  "routePrefix": "/api/v1/billing",
  "platforms": ["admin:pc", "user:pc", "user:h5"],
  "modules": [
    {"code": "invoice", "name": "发票管理"},
    {"code": "payment", "name": "支付管理"},
    {"code": "refund", "name": "退款管理"}
  ],
  "frontends": {
    "admin": {"pc": "dist/admin-pc/"},
    "user": {"pc": "dist/user-pc/", "h5": "dist/user-h5/"}
  },
  "menus": [
    {
      "type": "menu",
      "name": "发票管理",
      "permissionCode": "billing:invoice:menu",
      "path": "/plugin/billing/invoice",
      "icon": "ep:document",
      "platform": "admin",
      "moduleCode": "invoice",
      "sort": 1,
      "component": "plugin/container",
      "children": [
        {
          "type": "button",
          "name": "创建发票",
          "permissionCode": "billing:invoice:create",
          "platform": "admin",
          "moduleCode": "invoice",
          "sort": 1
        }
      ]
    }
  ],
  "apiPermissions": [
    {
      "type": "GROUP",
      "name": "发票管理",
      "moduleCode": "invoice",
      "children": [
        {
          "type": "ENDPOINT",
          "name": "发票列表",
          "displayName": "获取发票列表",
          "permissionCode": "billing:invoice:list",
          "urlPattern": "/api/v1/billing/invoices",
          "httpMethod": "GET"
        },
        {
          "type": "ENDPOINT",
          "name": "创建发票",
          "displayName": "创建新发票",
          "permissionCode": "billing:invoice:create",
          "urlPattern": "/api/v1/billing/invoices",
          "httpMethod": "POST"
        }
      ]
    }
  ],
  "exposedActions": [
    {
      "name": "getInvoice",
      "version": "1.0",
      "inputSchema": {"invoiceId": "int64"},
      "outputSchema": {"invoice": "InvoiceDTO"}
    }
  ],
  "subscribedEvents": ["order.created", "payment.completed"]
}
```

### API 设计

| 方法 | 路径 | 描述 | 认证 | 权限码 |
|------|------|------|------|--------|
| GET | /api/v1/admin/plugins | 插件列表（含状态） | JWT | platform:plugin:list |
| POST | /api/v1/admin/plugins/upload | 上传安装插件 | JWT | platform:plugin:install |
| POST | /api/v1/admin/plugins/:name/start | 启动插件 | JWT | platform:plugin:start |
| POST | /api/v1/admin/plugins/:name/stop | 停止插件 | JWT | platform:plugin:stop |
| DELETE | /api/v1/admin/plugins/:name | 卸载插件 | JWT | platform:plugin:uninstall |
| PUT | /api/v1/admin/plugins/:name/upgrade | 升级插件（上传新包） | JWT | platform:plugin:install |
| GET | /api/v1/admin/plugins/:name/health | 健康检查 | JWT | platform:plugin:list |
| GET | /api/v1/admin/app-catalog | 应用目录（含订阅状态） | JWT | auth_required=0（免检） |
| POST | /api/v1/admin/app-subscriptions | 租户订阅应用 | JWT | tenant:app:subscribe |
| DELETE | /api/v1/admin/app-subscriptions/:app_code | 租户退订应用 | JWT | tenant:app:unsubscribe |
| GET | /api/v1/admin/app-subscriptions | 当前租户已订阅列表 | JWT | auth_required=0（免检） |
| PUT | /api/v1/admin/app-subscriptions/:app_code/modules | 更新租户已订阅应用的启用模块 | JWT | tenant:app:subscribe |

### 数据库设计

本 CR 无 DDL 变更。使用现有表：
- `admin_application`（app_type=PLUGIN 记录）
- `admin_resource`（插件菜单资源）
- `admin_api_permission`（插件 API 权限）
- `admin_tenant_app`（租户订阅关系）
- `admin_role_resource` / `admin_role_api`（角色绑定）
- `sys_plugin`（插件运行时状态）

### 核心逻辑

#### 1. 安装流程改造

```
InstallFromFile
  ├── 解压 → 读取 plugin.json（V2 完整字段）
  ├── 重名校验（两层）：
  │     ├── sys_plugin WHERE name = ? → 已安装同名插件则拒绝
  │     └── admin_application WHERE app_code = ? AND deleted_at IS NULL → 与内置/外部应用冲突则拒绝
  │     错误信息："插件名称 xxx 与已有应用冲突，无法安装"
  ├── 写入 sys_plugin 记录
  ├── 创建 admin_application:
  │     app_code = plugin.name
  │     app_type = "PLUGIN"
  │     name = plugin.displayName || plugin.name
  │     route_prefix = plugin.routePrefix
  │     platforms = plugin.platforms (JSON)
  │     modules = plugin.modules (JSON)
  │     status = 1
  ├── 部署前端 bundle（按 frontends 配置）
  └── 不再调用 RegisterMenus
```

#### 2. 启动流程改造

```
PluginManager.Start / StartProcess
  ├── 调用 Plugin.Register()（gRPC）→ 获取 PluginInfo
  ├── 注册路由到内存 Registry（保留，用于 PluginProxy 路由匹配）
  ├── 读取 plugin.json（从 pluginsDir/{name}/plugin.json）
  ├── PluginResourceSyncer.SyncOnStart(pluginName, manifest):
  │     ├── 事务开始
  │     ├── 硬删除 admin_resource WHERE app_code = ?（Unscoped，事务内不可见于外部查询）
  │     ├── 硬删除 admin_api_permission WHERE app_code = ?（Unscoped）
  │     ├── 批量 INSERT admin_resource（from manifest.menus）
  │     ├── 批量 INSERT admin_api_permission（from manifest.apiPermissions）
  │     ├── 清理孤儿角色绑定（admin_role_resource/admin_role_api 中引用了已删除 ID 的记录）
  │     ├── 同步更新 admin_application 元数据（syncApplicationMeta）
  │     ├── 事务提交
  │     └── 事务提交后同步调用 AutoDiscover.Refresh() 刷新 api_code_map
  ├── ActionRegistry.Register(pluginName, manifest.exposedActions)
  ├── EventBus.Subscribe(pluginName, manifest.subscribedEvents)
  └── 更新 sys_plugin.status = Running
```

#### 3. 停止流程

```
PluginManager.Stop
  ├── 从内存 Registry 注销路由
  ├── EventBus.Unsubscribe(pluginName)
  ├── ActionRegistry.Unregister(pluginName)
  ├── 终止子进程（如有）
  ├── 更新 sys_plugin.status = Stopped
  └── admin_resource 中的记录保留（不删除，停止不清理资源）
```

#### 4. 卸载流程改造

```
Installer.Uninstall
  ├── Kill 进程
  ├── PluginResourceSyncer.SyncOnUninstall(pluginName):
  │     ├── 事务开始
  │     ├── 查出 admin_resource IDs WHERE app_code = ?
  │     ├── 查出 admin_api_permission IDs WHERE app_code = ?
  │     ├── DELETE admin_role_resource WHERE resource_id IN (...)
  │     ├── DELETE admin_role_api WHERE api_permission_id IN (...)
  │     ├── DELETE admin_role_app WHERE app_code = ?
  │     ├── DELETE admin_resource WHERE app_code = ?
  │     ├── DELETE admin_api_permission WHERE app_code = ?
  │     ├── 软删除 admin_application WHERE app_code = ?
  │     ├── DELETE admin_tenant_app WHERE app_code = ?
  │     └── 事务提交
  ├── 删除文件目录
  ├── 删除 sys_plugin 记录
  └── 触发 AutoDiscover 刷新
```

#### 5. CallPlugin 版本校验

**proto 变更**：修改 plugin.proto 的 CallPluginRequest，增加 `action_version` 字段：
```protobuf
message CallPluginRequest {
    string target_plugin = 1;
    string method = 2;       // action name
    bytes payload = 3;
    string action_version = 4; // 新增：请求的 action 版本，如 "1.0"
}
```

需重新生成 Go 代码（protoc-gen-go）。

```go
// ActionRegistry 存储插件暴露的 Action 信息
type ActionRegistry struct {
    mu      sync.RWMutex
    actions map[string][]ActionDescriptor // pluginName → actions
}

type ActionDescriptor struct {
    Name    string
    Version string // 语义化版本 "1.0"
}

// CallPlugin 增强
func (b *EventBus) CallPlugin(ctx context.Context, req *CallPluginRequest) (*CallPluginResponse, error) {
    // 1. 检查插件是否运行
    inst, ok := b.mgr.GetPlugin(req.TargetPlugin)
    if !ok { return nil, ErrPluginNotFound }
    if inst.Status != StatusRunning { return nil, ErrPluginNotRunning }

    // 2. Action 版本校验（如果请求带了 ActionVersion）
    if req.ActionVersion != "" {
        if !b.actionRegistry.IsCompatible(req.TargetPlugin, req.Method, req.ActionVersion) {
            return nil, ErrActionVersionIncompatible
        }
    }

    // 3. 转发调用
    return inst.Service.CallPlugin(ctx, req)
}

// IsCompatible: major 版本相同即兼容
func (r *ActionRegistry) IsCompatible(plugin, action, requestVersion string) bool {
    // 查找目标插件的该 action
    // 比较 major 版本
}
```

#### 6. PluginProxy 503 增强

```go
func (p *PluginProxy) Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        name := c.Param("name")
        
        inst, exists := p.mgr.GetPlugin(name)
        if !exists {
            c.JSON(404, gin.H{"code": 40400, "msg": "插件不存在"})
            return
        }
        if inst.Status != StatusRunning {
            c.JSON(503, gin.H{
                "code": 50301,
                "msg":  "插件维护中，请稍后再试",
                "data": gin.H{"plugin": name, "status": "stopped"},
            })
            return
        }
        // 正常代理...
    }
}
```

**SUPER_ADMIN 权限可见性规则**：
- SUPER_ADMIN 在角色配置页可看到所有应用权限（含未被任何租户订阅的 PLUGIN 应用）
- 普通租户管理员仅能看到本租户已订阅应用的权限

#### 7. 租户应用订阅

```go
// AppCatalogHandler — GET /api/v1/admin/app-catalog
func (h *Handler) GetAppCatalog(c *gin.Context) {
    tenantID := middleware.GetTenantId(c)
    
    // 查所有 status=1 的应用
    apps := h.appRepo.ListActive()
    
    // 查当前租户已订阅
    subscribed := h.tenantAppRepo.ListByTenant(tenantID)
    subscribedMap := toMap(subscribed)
    
    // 合并订阅状态
    for _, app := range apps {
        app.Subscribed = subscribedMap[app.AppCode] != nil
    }
    
    // 附加插件运行状态（PLUGIN 类型）
    for _, app := range apps {
        if app.AppType == "PLUGIN" {
            inst, ok := h.pluginMgr.GetPlugin(app.AppCode)
            if ok {
                app.PluginStatus = inst.Status
            }
        }
    }
}

// SubscribeApp — POST /api/v1/admin/app-subscriptions
func (h *Handler) SubscribeApp(c *gin.Context) {
    // 验证应用存在且 status=1
    // 检查配额（quota.max_apps）
    // 创建 admin_tenant_app 记录（enabled_modules=NULL 表示全部启用）
}

// UnsubscribeApp — DELETE /api/v1/admin/app-subscriptions/:app_code
func (h *Handler) UnsubscribeApp(c *gin.Context) {
    appCode := c.Param("app_code")
    
    // 禁止退订 BUILTIN 类型应用
    app := h.appRepo.FindByCode(appCode)
    if app.AppType == "BUILTIN" {
        c.JSON(400, gin.H{"code": 40001, "msg": "内置应用不允许退订"})
        return
    }
    
    // 删除 admin_tenant_app
    // 级联清理该租户下引用该 app_code 资源的角色绑定
}
```

#### 8. 插件升级流程

```
PUT /api/v1/admin/plugins/:name/upgrade (multipart file)
  ├── 校验：目标插件必须存在（已安装）
  ├── 校验：上传包 plugin.json.name 必须与 URL :name 一致
  ├── 如果插件运行中 → 先停止（Stop）
  ├── 替换文件（安全策略：先部署新目录再原子替换）：
  │     ├── 解压新包到 plugins/{name}.new/
  │     ├── 重新部署前端 bundle 到 static/plugins/{name}.new/
  │     ├── 原子 rename: plugins/{name} → plugins/{name}.bak
  │     ├── 原子 rename: plugins/{name}.new → plugins/{name}
  │     ├── 原子 rename: static/plugins/{name} → static/plugins/{name}.bak (如有)
  │     ├── 原子 rename: static/plugins/{name}.new → static/plugins/{name} (如有)
  ├── 更新 sys_plugin 记录（version/binary_path/frontend_path）
  ├── 更新 admin_application（version/modules/platforms/route_prefix/description）
  ├── 启动插件（由 Handler 层调用 manager.Start，非 installer 直接调）
  │     └── 启动触发 SyncOnStart（自动同步新菜单/API）
  ├── 启动成功 → 清理 .bak 目录 → 返回升级结果（含新旧版本号）
  └── 启动失败 → 回滚：rename .bak 回原名 → 恢复 sys_plugin 版本号 → 返回错误
```

**升级 vs 卸载重装的区别**：
| 维度 | 升级 | 卸载+重装 |
|------|------|----------|
| 插件数据表 | 保留（插件内部 AutoMigrate 处理 DDL 变更） | 可选清除 |
| 租户订阅关系 | 保留 | 丢失 |
| 角色绑定 | SyncOnStart 会清理已删除资源的绑定，保留仍存在资源的绑定 | 全部丢失 |
| admin_application ID | 不变 | 新 ID |

**大版本不兼容升级处理**：

plugin.json 支持声明破坏性升级标记：
```json
{
  "version": "2.0.0",
  "breakingUpgrade": true,
  "minUpgradeFrom": "1.0.0",
  "migrationNotes": "V2 重构了订单表结构，旧数据将自动迁移，预计耗时 1-5 分钟"
}
```

升级接口行为分支：
```
PUT /api/v1/admin/plugins/:name/upgrade
  ├── 解析新包 plugin.json
  ├── 版本校验：
  │     ├── minUpgradeFrom 有值 && 当前版本 < minUpgradeFrom → 拒绝，返回：
  │     │   "当前版本过低，请先升级到 {minUpgradeFrom} 再执行此升级"
  │     └── breakingUpgrade = true → 返回 needConfirm=true + migrationNotes
  │           前端收到后弹窗二次确认
  ├── 用户确认（请求带 confirm=true）→ 执行升级
  └── 插件启动后自行完成数据迁移
```

**数据迁移责任分层**：
| 层 | 责任 | 机制 |
|---|------|------|
| Host | 版本兼容性校验 + 破坏性升级提示 + 二次确认 | breakingUpgrade / minUpgradeFrom 字段 |
| 插件 | 旧表结构检测 + 数据迁移执行 | Register() 或首次 HandleRequest 中自行处理 |
| Host | 升级失败回滚（保留旧二进制备份） | 升级前 rename 旧目录为 .bak，失败时 rename 回原名 |

**回滚机制**：
- 升级时 Host 将旧 `plugins/{name}/` rename 为 `plugins/{name}.bak/`（原子操作）
- 新包部署到 `plugins/{name}.new/` → rename 为 `plugins/{name}/`
- 如果新版插件启动失败（Register 超时或报错）→ rename .bak 恢复 + 恢复 sys_plugin 版本号
- 备份目录在启动成功后立即清理
- **Host 中途崩溃恢复**：启动时检测 .new/.bak 残留目录，如果 plugins/{name}/ 不存在但 .bak 存在 → 自动恢复

#### 9. 启动时同步 admin_application 元数据

```go
// 在 SyncOnStart 中追加：更新 admin_application 的版本/模块/平台等字段
func (s *PluginResourceSyncer) syncApplicationMeta(tx *gorm.DB, manifest *ManifestV2) error {
    updates := map[string]interface{}{
        "name":         manifest.DisplayName,
        "description":  manifest.Description,
        "route_prefix": manifest.RoutePrefix,
        "platforms":    datatypes.JSON(marshalJSON(manifest.Platforms)),
        "modules":      datatypes.JSON(marshalJSON(manifest.Modules)),
    }
    return tx.Model(&model.Application{}).
        Where("app_code = ?", manifest.Name).
        Updates(updates).Error
}
```

#### 10. sys_menu 旧数据清理

启动时执行一次性清理：
```go
// cleanLegacyPluginMenus 清理 sys_menu 中遗留的插件菜单
func cleanLegacyPluginMenus(db *gorm.DB) {
    db.Exec("UPDATE sys_menu SET deleted_at = NOW() WHERE menu_name LIKE 'plugin_%' AND deleted_at IS NULL")
}
```

### pluginJSON V2 Go 结构体

```go
// ManifestV2 plugin.json 完整结构
type ManifestV2 struct {
    Name             string                    `json:"name"`
    Version          string                    `json:"version"`
    DisplayName      string                    `json:"displayName"`
    Description      string                    `json:"description"`
    RoutePrefix      string                    `json:"routePrefix"`
    Platforms        []string                  `json:"platforms"`
    Modules          []ManifestModule          `json:"modules"`
    Frontends        map[string]map[string]string `json:"frontends"`
    Menus            []ManifestMenu            `json:"menus"`
    ApiPermissions   []ManifestApiPermission   `json:"apiPermissions"`
    ExposedActions   []ManifestAction          `json:"exposedActions"`
    SubscribedEvents []string                  `json:"subscribedEvents"`
    BreakingUpgrade  bool                      `json:"breakingUpgrade"`  // 是否为破坏性升级
    MinUpgradeFrom   string                    `json:"minUpgradeFrom"`   // 最低可升级来源版本
    MigrationNotes   string                    `json:"migrationNotes"`   // 迁移说明（前端展示）
}

type ManifestModule struct {
    Code string `json:"code"`
    Name string `json:"name"`
}

type ManifestMenu struct {
    Type           string         `json:"type"`           // menu/button/page
    Name           string         `json:"name"`
    PermissionCode string         `json:"permissionCode"`
    Path           string         `json:"path"`
    Icon           string         `json:"icon"`
    Component      string         `json:"component"`
    Platform       string         `json:"platform"`       // admin/user
    ModuleCode     string         `json:"moduleCode"`
    Sort           int            `json:"sort"`
    Children       []ManifestMenu `json:"children"`
}

type ManifestApiPermission struct {
    Type           string                    `json:"type"`   // GROUP/ENDPOINT
    Name           string                    `json:"name"`
    DisplayName    string                    `json:"displayName"`
    PermissionCode string                    `json:"permissionCode"`
    URLPattern     string                    `json:"urlPattern"`
    HTTPMethod     string                    `json:"httpMethod"`
    ModuleCode     string                    `json:"moduleCode"`
    Children       []ManifestApiPermission   `json:"children"`
}

type ManifestAction struct {
    Name         string `json:"name"`
    Version      string `json:"version"`
    InputSchema  string `json:"inputSchema"`
    OutputSchema string `json:"outputSchema"`
}
```

### Handler 改动清单

| 文件 | 改动 |
|------|------|
| `backend/common/plugin/syncer.go` | 新建 — PluginResourceSyncer |
| `backend/common/plugin/manifest.go` | 新建 — ManifestV2 结构体 + 解析函数 |
| `backend/common/plugin/action_registry.go` | 新建 — ActionRegistry |
| `backend/common/plugin/manager.go` | 改造 — Start/StartProcess 集成 syncer + eventbus + action |
| `backend/common/plugin/installer.go` | 改造 — 安装时创建 admin_application，升级接口，废弃 RegisterMenus |
| `backend/common/plugin/proxy.go` | 改造 — 503 结构化返回 |
| `backend/common/plugin/event_bus.go` | 改造 — CallPlugin 增加版本校验 |
| `backend/common/auth/handler/app_catalog_handler.go` | 新建 — 应用目录 + 订阅接口 |
| `backend/common/auth/router.go` | 改造 — 注册新路由 + 权限码 |
| `backend/common/auth/service/tenant_service.go` | 改造 — 退订时级联清理 |
| `frontend/src/views/plugin/` | 改造 — 插件管理页 |
| `frontend/src/views/app-catalog/` | 新建 — 应用目录页 |
| `frontend/src/api/plugin.ts` | 改造 — 适配新接口 |
| `frontend/src/api/app-catalog.ts` | 新建 — 应用目录 API |

---

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有插件安装/卸载文件操作不受影响 | 上传 zip → 正确解压到 plugins/ 目录 |
| RG-2 | 现有插件 gRPC 通信正常 | Plugin.Register() + HandleRequest 正常工作 |
| RG-3 | 现有角色权限配置页对 BUILTIN 应用的展示不变 | platform_admin 应用的资源树正常展示 |
| RG-4 | 现有租户订阅逻辑（admin_tenant_app）不被破坏 | 已订阅租户仍可正常使用 |
| RG-5 | 现有路由注册不受影响 | 所有 /api/v1/admin/ 接口可正常访问 |
| RG-6 | 现有 AutoDiscover 逻辑不受影响 | 内置应用的 api_code_map 正确生成 |
| RG-7 | EventBus 现有事件广播行为不变 | PublishEvent 正常广播到已订阅插件 |

## 正确性属性

- 插件资源写入 admin_resource 时 app_code 必须与插件 name 一致
- 插件 API 权限写入 admin_api_permission 时 app_code 必须与插件 name 一致
- SyncOnStart 必须在事务中执行，保证原子性
- 卸载时级联清理必须覆盖所有租户的角色绑定（无 tenant_id 过滤）
- CallPlugin 版本校验不影响不带 ActionVersion 字段的旧调用（向后兼容）
- 插件停止后 admin_resource 中的记录保留不删除
- 应用目录列表对 C 端用户不可见（仅管理端）
