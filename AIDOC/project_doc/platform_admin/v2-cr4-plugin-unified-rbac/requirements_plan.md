# 需求计划：V2-CR4 插件系统统一 RBAC + 通信契约

## 需求理解

- **目标**：将插件系统的权限管理从 `sys_menu` 旧体系迁移到 V2 的 `admin_resource` / `admin_api_permission` 统一 RBAC 体系，同时升级插件通信契约（Manifest V2 + Action 版本化校验）
- **范围**：插件生命周期管理、权限同步（PluginResourceSyncer）、Manifest V2 解析、插件通信版本校验、EventBus 自动订阅、前端插件管理/应用目录
- **预期效果**：
  1. 插件安装时自动创建 `admin_application`（app_type=PLUGIN）
  2. 插件启动时通过 PluginResourceSyncer 将菜单/API 同步到 `admin_resource` / `admin_api_permission`
  3. 角色权限配置页统一展示插件应用的权限（与内置应用无差别）
  4. 废弃 `Installer.RegisterMenus`（不再写 sys_menu）
  5. 插件间调用支持 Action 版本兼容性校验
  6. EventBus 基于 Manifest.subscribedEvents 自动注册订阅
  7. 插件停止时 PluginProxy 返回 503

## 假设列表

- [假设-1] 现有已安装的插件（sys_plugin 表中的记录）需要在本次升级后重新启动以触发资源同步，无需自动迁移历史数据
- [假设-2] Manifest V2 通过扩展现有 proto.PluginInfo（增加新字段）实现，不破坏现有 protobuf 向后兼容性
- [假设-3] `sys_menu` 表中由插件注册的菜单在迁移完成后保持软删除状态，不物理删除（以便回滚）
- [假设-4] 插件的 app_code 直接使用插件的 name 字段（与现有 routePrefix 命名保持一致）
- [假设-5] 本 CR 不实现"应用市场"功能，租户侧的"应用目录"仅展示当前平台已上架且运行中的应用列表

## 澄清问题

- [Question-1] Manifest V2 的 `modules` / `platforms` / `frontends` / `exposedActions` / `subscribedEvents` 字段——是在 proto 层面扩展（修改 .proto 文件 + 重新生成），还是仅在 plugin.json 层面解析（proto 保持不变，plugin.json 作为补充元数据）？
  
  **行业实践参考**：
  - **gRPC proto 扩展**：结构化强、类型安全、适合强契约通信场景；但需要所有插件重新编译
  - **JSON manifest 补充**：灵活、无需重新编译旧插件、适合描述性元数据；缺点是缺少编译期校验
  - **推荐**：混合方案——核心通信字段（name/version/routePrefix/menus/perms）保留 proto；扩展描述性字段（modules/platforms/frontends/exposedActions/subscribedEvents）放在 plugin.json 中由 Host 解析。理由：旧插件无需重新编译即可运行，新插件可选声明 V2 字段。

  [Answer-1] 
混合方案
- [Question-2] 插件安装时创建 `admin_application` 记录，其中 `platforms` / `modules` 字段从哪读取？如果旧插件没有 Manifest V2 字段，默认值是什么？
  
  [Answer-2] 
不考虑旧插件，因为还没上生产，按最佳设计来做
- [Question-3] 插件停止后，前端角色配置页中该插件的权限是否仍可见（可取消分配）？还是灰显不可操作？
  
  [Answer-3] 
仍可见，但配置那里红字提示插件已停止
- [Question-4] CallPlugin 的 Action 版本兼容性检查——是仅 major 版本必须相同（语义化版本），还是有更严格的规则（如 major.minor 必须匹配）？
  
  [Answer-4] 
major即可
- [Question-5] 是否需要在本 CR 实现「租户侧应用目录」页面（租户管理员浏览可用应用 + 申请开通）？还是仅实现平台侧的插件管理改造？
  
  [Answer-5] 
需要在本 CR 实现「租户侧应用目录」页面（租户管理员浏览可用应用 + 申请开通）
## 非功能需求建议

- **性能**：PluginResourceSyncer.SyncOnStart 在事务中执行，大型插件（50+ 菜单 + 100+ API）应控制在 2s 内完成
- **幂等性**：SyncOnStart 每次全量同步（先清后写），确保重复启动不产生脏数据
- **向后兼容**：旧插件（无 Manifest V2 字段）仍可正常安装/启动，使用默认值填充
- **安全**：插件注册的 API 权限默认 auth_required=1，需要角色分配后才能访问
- **多租户**：插件资源属于平台级（不带 tenant_id），通过 tenant_app 订阅关系控制可见性

## 影响范围预判

- **涉及模块**：
  - `backend/common/plugin/`：syncer.go（新建）、manager.go（集成 syncer）、installer.go（创建 application）、registry.go（可能调整）、proxy.go（503 增强）
  - `backend/common/auth/`：model（无新增）、service/resource_service.go（查询时含插件资源）、handler
  - `plugin-sdk/proto/`：可能扩展 PluginInfo 字段
  - `frontend/`：插件管理页改造、应用目录页（如需）、角色权限配置页展示插件权限
- **涉及文件（预估）**：~15-20 个文件
- **可能的副作用**：
  - 废弃 RegisterMenus 后，旧 sys_menu 中的插件菜单不再更新（需确认 GetUserMenu 不再从 sys_menu 读取插件菜单）
  - 角色权限配置页数据量增加（含插件应用的资源树）
