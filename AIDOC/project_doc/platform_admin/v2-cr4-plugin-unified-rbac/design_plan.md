# 设计计划：V2-CR4 插件系统统一 RBAC + 通信契约

## 设计方向

基于已确认的需求，采用以下技术方案方向：

### 1. Manifest V2 — plugin.json 扩展

proto 保持不变（仅核心 RPC 通信），所有描述性元数据放在 plugin.json 中：
- Host 在安装时解析 plugin.json 完整字段
- Host 在启动时从 plugin.json 读取 menus/apiPermissions/subscribedEvents/exposedActions
- 不再依赖 proto.PluginInfo.Menus/Perms 做资源同步（仅保留用于内存级 Registry 的临时路由注册）

### 2. PluginResourceSyncer — 先清后写的事务同步

- 新建 `backend/common/plugin/syncer.go`
- SyncOnStart：在单个数据库事务中清理旧资源 → 写入新资源 → 清理孤儿角色绑定
- SyncOnUninstall：删除资源 + 级联清理角色绑定 + 软删除 admin_application
- 触发时机：PluginManager.Start() / StartProcess() 成功后调用

### 3. Installer 改造 — 安装时创建 admin_application

- InstallFromFile 中解析 plugin.json V2 字段
- 创建 admin_application（app_type=PLUGIN）
- 废弃 RegisterMenus / UnregisterMenus 方法（标记 Deprecated，不再调用）

### 4. CallPlugin 版本校验 — ActionRegistry

- 新建 ActionRegistry（内存 map：pluginName → []ActionDescriptor）
- 插件启动时从 plugin.json.exposedActions 注册
- EventBus.CallPlugin 中增加版本校验拦截（major 版本匹配）

### 5. EventBus 自动订阅

- PluginManager.Start 中读取 plugin.json.subscribedEvents
- 自动调用 EventBus.Subscribe（替代手动订阅）

### 6. PluginProxy 503 增强

- proxy.go Handler 中区分"未注册"（404）和"已停止"（503 + 结构化 body）

### 7. 前端

- 插件管理页：适配 admin_application 数据源 + 状态感知 UI
- 应用目录页（新建）：展示可订阅应用 + 订阅/退订操作
- 角色权限配置页：已支持按 app_code 分组展示，插件资源写入 admin_resource 后自动出现

## 技术选型

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| Manifest 全 proto | 类型安全、编译期校验 | 扩展需重编译所有插件 | ✗ |
| Manifest 全 plugin.json | 灵活、零耦合、热扩展 | 无编译期校验 | ✗ |
| 混合（proto RPC + plugin.json 元数据） | 运行时通信强类型 + 元数据灵活扩展 | 两处维护 | ✓ |

## 澄清问题

- [Question-1] plugin.json 中的 `menus` 字段结构是否直接复用现有 proto.MenuItem 的 JSON 等价格式（title/icon/path/sort/children），还是采用更丰富的格式（增加 permissionCode/platform/moduleCode/type）？

  **推荐**：采用丰富格式，因为 admin_resource 表需要 platform/module_code/permission_code/type 字段，如果 plugin.json 不声明这些信息，syncer 将无法正确填充。建议格式：
  ```json
  {
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
        "children": [...]
      }
    ]
  }
  ```

  [Answer-1]
丰富格式
- [Question-2] plugin.json 中的 `apiPermissions` 字段如何声明？是否采用类似 admin_api_permission 的树形结构（GROUP + ENDPOINT），还是扁平列表？

  **推荐**：树形结构（与 admin_api_permission 对齐），一级为 GROUP 分组，二级为 ENDPOINT 叶子节点：
  ```json
  {
    "apiPermissions": [
      {
        "type": "GROUP",
        "name": "发票管理",
        "moduleCode": "invoice",
        "children": [
          {
            "type": "ENDPOINT",
            "name": "发票列表",
            "permissionCode": "billing:invoice:list",
            "urlPattern": "/api/v1/billing/invoices",
            "httpMethod": "GET"
          }
        ]
      }
    ]
  }
  ```

  [Answer-2]
同意
- [Question-3] 租户订阅应用接口的权限码如何定义？是所有租户管理员默认可用，还是需要 SUPER_ADMIN 或特定 permission_code？

  **推荐**：
  - `GET /app-catalog` — 所有已认证用户可访问（展示列表，可见性由接口内部按 tenant 过滤）
  - `POST /app-subscriptions` — 需要 `tenant:app:subscribe` 权限码（默认分配给 TENANT_ADMIN）
  - `DELETE /app-subscriptions/:app_code` — 需要 `tenant:app:unsubscribe` 权限码（默认分配给 TENANT_ADMIN）

  [Answer-3]
同意
## 风险点

- [Risk-1] **事务大小**：SyncOnStart 在事务中清理 + 重写，如果插件声明大量资源（>200 条），可能导致事务耗时长。缓解：批量 INSERT 替代逐条 Create。
- [Risk-2] **GetUserMenu 数据源切换**：当前 GetUserMenu 是否仍有从 sys_menu 读取的旧逻辑？需确认 CR-1 已彻底切换到 admin_resource。如果有残留逻辑需在本 CR 清理。
- [Risk-3] **AutoDiscover 刷新时机**：SyncOnStart 写入新 API 后，api_code_map 需要刷新才能被 DynamicPermissionMiddleware 识别。需确认事件触发链路正确。
- [Risk-4] **角色绑定清理的范围**：SyncOnUninstall 级联清理角色绑定时，需处理多租户场景——所有租户的角色绑定都要清理。
