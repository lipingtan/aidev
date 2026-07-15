# 需求文档：插件化架构规范（Plugin Architecture）

## 背景

当前系统为单体前端 + 固定业务模块结构。需要升级为"最小平台框架 + 插件化扩展"架构，使所有具体业务功能通过插件方式安装/卸载，框架仅保留用户体系、权限、多租户、插件管理等核心能力。

涉及三端：
- dev-web-admin（管理端，PC）
- dev-web-user（用户端，PC + H5 响应式，Vant + 响应式增强）
- Go 后端（plugin-sdk 扩展）

## 关键约束

1. **框架最小化**：框架只包含用户体系 + 权限 + 多租户 + 插件管理 + 系统配置/日志
2. **插件化扩展**：所有业务功能（部门/岗位/字典/定时任务/监控等）均为插件
3. **双模式构建**：插件支持源码编译（`src/plugins/` 子目录）和独立构建打包（`/static/plugins/{name}/index.js`）两种模式
4. **插件间通信**：共享上下文（只读全局状态）+ EventBus + 扩展点，插件不直接互调
5. **集群一致性**：DB 为真相源 + 启动全量加载 + Redis Pub/Sub 广播变更（初期预留接口，单实例先不接 Redis）
6. **用户端响应式**：dev-web-user 使用 Vant + CSS 响应式增强，同一套代码适配 PC 和 H5
7. **向后兼容**：admin-frontend-refactor 已实现的模块暂保持框架内，后续再拆为插件
8. **后端插件双适配器**：定义统一 Plugin Interface，实现 LocalAdapter（内嵌编译，当前使用）和 RemoteAdapter（gRPC/HTTP，预留接口不实现）。框架面向接口编程，后续新增远程插件只需实现适配器，框架零改动
9. **前端资产分发**：插件前端构建产物与后端一起打包到容器镜像中
10. **插件回滚含 DB 迁移**：插件需提供 up/down migration，回滚时恢复旧版本文件 + 执行 down migration
11. **前端源码模式目录**：源码编译模式的插件放置在主项目 `src/plugins/{name}/` 目录下

## 术语表

- **Plugin_Shell**：平台框架骨架，包含核心功能和插件加载基础设施
- **Plugin_Module**：一个可安装/卸载的业务功能单元，包含前后端代码
- **Plugin_Manifest**：插件描述文件（manifest.json），声明插件元数据、权限、路由、依赖
- **Plugin_SDK**：供插件开发使用的类型定义和工具库
- **Plugin_Registry**：后端插件注册表（sys_plugin 表），存储插件状态和配置
- **EventBus**：框架提供的发布/订阅事件总线，用于插件间松耦合通信
- **Extension_Point**：框架定义的扩展槽位，插件向其注册内容由框架聚合展示

## 需求

### 需求 1：插件 Manifest 规范

**用户故事：** 作为插件开发者，我需要一个标准的描述文件格式，以便框架能识别和加载我的插件。

#### 验收标准

1. THE Plugin_Module SHALL 包含 `manifest.json` 文件，声明以下必填字段：name、version、displayName、description
2. THE Plugin_Manifest SHALL 支持声明 permissions 数组，定义插件所需的权限 code 列表
3. THE Plugin_Manifest SHALL 支持声明 routes 数组，定义插件提供的前端路由（path、component、meta）
4. THE Plugin_Manifest SHALL 支持声明 menus 数组，定义插件在侧边栏中注册的菜单项（title、icon、path、sort）
5. THE Plugin_Manifest SHALL 支持声明 dependencies 对象，标明对其他插件的依赖（name + semver）
6. THE Plugin_Manifest SHALL 支持声明 platforms 数组，标明插件适用的端（admin/user/both）
7. WHEN manifest.json 格式不合法或缺少必填字段, THE Plugin_Shell SHALL 拒绝加载并报告具体错误

### 需求 2：前端插件 SDK

**用户故事：** 作为插件开发者，我需要一个 TypeScript SDK，以便在统一接口下注册路由、菜单、Store 和组件。

#### 验收标准

1. THE Plugin_SDK SHALL 导出 `definePlugin(config)` 函数，接收插件配置并返回标准插件对象
2. THE Plugin_SDK SHALL 提供 `usePluginContext()` 组合式函数，返回当前用户、租户、权限列表等只读全局状态
3. THE Plugin_SDK SHALL 提供 `useEventBus()` 组合式函数，支持 emit/on/off 事件操作
4. THE Plugin_SDK SHALL 提供 TypeScript 类型定义，包含 PluginManifest、PluginRoute、PluginMenu 等接口
5. THE Plugin_SDK SHALL 提供 `registerExtension(pointName, component)` 函数，用于向框架扩展点注册内容
6. THE Plugin_SDK SHALL 在 npm 包或项目内模块形式提供，支持 `import { definePlugin } from '@/plugin-sdk'`

### 需求 3：前端插件加载机制

**用户故事：** 作为框架，我需要在运行时动态加载已启用的插件，以便用户访问插件提供的页面。

#### 验收标准

1. WHEN 用户登录成功后, THE Plugin_Shell SHALL 从后端获取当前租户已启用的插件列表
2. THE Plugin_Shell SHALL 支持两种加载模式：
   - 源码模式：从 `src/plugins/{name}/index.ts` 静态导入（构建时包含）
   - 运行时模式：从 `/static/plugins/{name}/index.js` 动态 import
3. WHEN 插件加载成功, THE Plugin_Shell SHALL 将插件声明的路由动态添加到 Vue Router
4. WHEN 插件加载成功, THE Plugin_Shell SHALL 将插件声明的菜单项合并到侧边栏菜单树
5. WHILE 插件正在加载, THE Plugin_Shell SHALL 展示 loading 状态
6. IF 插件加载失败, THEN THE Plugin_Shell SHALL 在控制台输出错误日志并跳过该插件，不影响其他插件和框架运行

### 需求 4：前端插件菜单动态注册

**用户故事：** 作为系统管理员，我需要启用插件后自动在侧边栏看到插件的菜单项，无需手动配置路由。

#### 验收标准

1. WHEN 插件启用后, THE Plugin_Shell SHALL 根据 manifest 中的 menus 配置自动在侧边栏渲染菜单项
2. THE Plugin_Shell SHALL 支持菜单排序（通过 sort 字段）
3. THE Plugin_Shell SHALL 支持菜单分组（通过 group 字段归入已有菜单组或创建新组）
4. WHEN 插件禁用或卸载后, THE Plugin_Shell SHALL 从侧边栏移除该插件的菜单项
5. THE Plugin_Shell SHALL 根据用户权限过滤菜单项（用户无插件声明的 permission 时不显示对应菜单）

### 需求 5：插件权限自动注册

**用户故事：** 作为系统管理员，我需要插件安装后其权限自动注册到权限系统，以便通过角色分配控制访问。

#### 验收标准

1. WHEN 插件安装时, THE Plugin_Registry SHALL 读取 manifest 中的 permissions 列表并写入权限表
2. THE Plugin_Registry SHALL 为每个权限记录关联所属插件（plugin_name 字段）
3. WHEN 插件卸载时, THE Plugin_Registry SHALL 移除该插件注册的所有权限记录
4. THE Plugin_Shell SHALL 在角色权限分配页面展示插件注册的权限项（按插件分组）
5. WHEN 插件升级且权限列表变化时, THE Plugin_Registry SHALL 自动增量更新权限（新增的加入、移除的标记废弃）

### 需求 6：后端插件生命周期管理

**用户故事：** 作为系统管理员，我需要在管理端安装、启用、禁用、卸载、升级插件。

#### 验收标准

1. THE Plugin_Registry SHALL 提供以下 API：
   - POST /api/v1/plugin/install — 安装插件
   - PUT /api/v1/plugin/{name}/enable — 启用
   - PUT /api/v1/plugin/{name}/disable — 禁用
   - DELETE /api/v1/plugin/{name} — 卸载
   - PUT /api/v1/plugin/{name}/upgrade — 升级
   - PUT /api/v1/plugin/{name}/rollback — 回滚到上一版本
2. WHEN 插件状态变更, THE Plugin_Registry SHALL 更新 sys_plugin 表并发布变更事件
3. THE Plugin_Registry SHALL 在 sys_plugin 表中记录：name、version、status、config、installed_at、updated_at
4. WHEN 集群部署时, THE Plugin_Registry SHALL 通过 Redis Pub/Sub 通知其他实例重载插件状态
5. IF Redis 不可用, THEN THE Plugin_Registry SHALL 兜底通过定时轮询 DB（30s 间隔）检测变更
6. WHEN 插件升级时, THE Plugin_Registry SHALL 保留上一版本快照以支持回滚
7. THE Plugin_Registry SHALL 定义统一 Plugin Interface，通过 LocalAdapter（内嵌编译）加载本地插件
8. THE Plugin_Registry SHALL 预留 RemoteAdapter 接口定义（gRPC/HTTP），使后续新增远程插件无需修改框架代码

### 需求 7：插件版本管理与回滚

**用户故事：** 作为系统管理员，我需要能升级插件到新版本，并在出现问题时回滚到上一版本。

#### 验收标准

1. THE Plugin_Registry SHALL 维护 sys_plugin_version 表记录所有已安装版本（plugin_name、version、snapshot_path、installed_at）
2. WHEN 执行升级时, THE Plugin_Registry SHALL 先保存当前版本快照（二进制 + 前端 bundle）再加载新版本
3. WHEN 执行回滚时, THE Plugin_Registry SHALL 恢复上一版本文件并重载
4. THE Plugin_Shell SHALL 在管理端展示插件版本历史列表
5. IF 升级过程中出错, THEN THE Plugin_Registry SHALL 自动回滚并报告错误原因
6. THE Plugin_Module SHALL 提供 up/down migration 脚本，升级时执行 up migration，回滚时执行 down migration
7. WHEN 回滚时, THE Plugin_Registry SHALL 按顺序执行 down migration 恢复数据库到上一版本状态

### 需求 8：插件间通信机制

**用户故事：** 作为插件开发者，我需要与其他插件进行松耦合的数据交互，而不直接依赖对方内部实现。

#### 验收标准

1. THE Plugin_SDK SHALL 提供全局 EventBus，支持 `emit(eventName, payload)` 和 `on(eventName, handler)`
2. THE Plugin_SDK SHALL 提供 `usePluginContext()` 返回只读共享状态（currentUser、currentTenant、permissions）
3. THE Plugin_Shell SHALL 定义标准扩展点（Extension Points），如：
   - `user-detail-tabs`：用户详情页附加 Tab
   - `dashboard-widgets`：首页仪表盘卡片
   - `global-actions`：全局操作按钮
4. THE Plugin_SDK SHALL 提供 `registerExtension(pointName, { component, sort })` 注册扩展内容
5. THE Plugin_Shell SHALL 在扩展点位置聚合所有插件注册的内容并按 sort 排序渲染

### 需求 9：dev-web-user 框架瘦身

**用户故事：** 作为开发人员，我需要清理 dev-web-user 中的物业业务模块，保留用户体系骨架和插件容器。

#### 验收标准

1. THE dev-web-user SHALL 移除以下业务模块：workorder（工单/报修）、billing（缴费）、notice（公告）、certify（认证）、repair（维修工台）、owner/base/family/property 相关 API
2. THE dev-web-user SHALL 保留框架骨架：登录/注册、个人中心（个人信息/修改密码）、TabBar 布局、错误页
3. THE dev-web-user SHALL 集成插件容器，支持动态加载用户端插件页面
4. THE dev-web-user SHALL 使用 Vant + CSS 响应式增强，同一套代码适配 PC（≥768px）和 H5（<768px）
5. THE dev-web-user SHALL 在 PC 端展示更宽的内容区域和适当的侧边导航（如有插件注册菜单）

### 需求 10：插件开发脚手架

**用户故事：** 作为插件开发者，我需要一个标准的项目模板和开发流程，以便快速创建新插件。

#### 验收标准

1. THE Plugin_SDK SHALL 提供插件项目模板（scaffold），包含 manifest.json、index.ts、示例路由和页面
2. THE Plugin_SDK SHALL 提供开发文档说明：插件目录结构、manifest 字段说明、开发/构建/发布流程
3. WHEN 使用源码模式开发时, THE Plugin_Shell SHALL 支持插件代码的 HMR 热更新
4. WHEN 使用独立构建模式时, THE Plugin_SDK SHALL 提供构建脚本将插件打包为 ES Module 输出到 `/static/plugins/{name}/`
5. THE Plugin_SDK SHALL 提供类型定义使插件开发时有完整的 TypeScript 类型提示

### 需求 11：插件 CSS 隔离

**用户故事：** 作为框架维护者，我需要确保插件的样式不会污染框架和其他插件。

#### 验收标准

1. THE Plugin_Shell SHALL 要求插件组件使用 scoped CSS 或 CSS Modules
2. THE Plugin_SDK SHALL 在开发文档中明确禁止插件使用全局样式覆盖
3. IF 插件使用运行时加载模式, THEN THE Plugin_Shell SHALL 将插件样式限定在插件容器 DOM 范围内
4. THE Plugin_Shell SHALL 提供公共 CSS 变量（主题色、间距、字号），插件通过 CSS 变量继承框架主题

### 需求 12：多租户插件可见性

**用户故事：** 作为平台运营者，我需要控制不同租户可使用哪些插件。

#### 验收标准

1. THE Plugin_Registry SHALL 在 sys_tenant_plugin 表中记录租户与插件的关联关系
2. WHEN 查询当前租户已启用插件列表时, THE Plugin_Registry SHALL 只返回该租户已授权且已启用的插件
3. THE Plugin_Shell SHALL 在管理端提供"租户插件分配"页面，支持为租户启用/禁用特定插件
4. WHEN 租户未被授权某插件时, THE Plugin_Shell SHALL 隐藏该插件的菜单、路由不可访问

## 非功能需求

- **性能**：未启用插件不加载代码，首屏不受插件数量影响
- **可靠性**：单个插件加载失败不影响框架和其他插件运行
- **安全**：插件权限受 RBAC 控制，无权限用户无法访问插件功能
- **可维护性**：Plugin SDK 提供完整 TypeScript 类型定义
- **开发效率**：源码模式下插件支持 HMR
- **集群安全**：插件状态变更通过 DB + 事件广播保证最终一致性（30s 内）
