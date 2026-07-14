# 需求文档

## 简介

将 `frontend`（pure-admin-thin）中的管理模块功能迁移到 `dev-web-admin`（自建 Vue3 框架）。保留 dev-web-admin 的页面框架、导航、菜单样式和皮肤不变，将 frontend 已实现的系统管理、权限管理、开发工具、系统工具、任务调度、插件管理等模块在 dev-web-admin 中重新实现。后端 API 不变（/api/v1/xxx）。

## 关键约束

1. **后端技术栈**：现有后端为 Go 语言实现（go-admin 框架），非 Java。dev-web-admin 原配套的 Java 后端源码位于 `projects/demo/game_server/bak/spmp-backend/`，仅作为理解 dev-web-admin 前端逻辑的参考，不用于生产。
2. **前端基座保留**：dev-web-admin 的页面框架（AppLayout）、导航（AppSidebar）、顶栏（AppHeader）、面包屑（AppBreadcrumb）、标签页（TagsView）、皮肤主题、全局样式必须原样保留。
3. **业务模块清理**：dev-web-admin 中偏具体业务场景的模块（base/片区/小区/楼栋/单元/房屋、owner/业主、workorder/工单、billing/缴费、notice/公告）需要移除，这些是原 Java 后端的特定业务，与当前 Go 后端不匹配。
4. **API 对接目标**：所有接口对接现有 Go 后端的 /api/v1/xxx 路由，不使用 Java 后端的接口。
5. **样式统一**：所有新实现的页面必须遵循 dev-web-admin 自身的样式体系，不迁移 frontend 的视觉效果。

## 术语表

- **Dev_Web_Admin**: 目标前端项目，基于 Vue3 + Element Plus + Pinia + vue-i18n 的自建管理后台框架
- **Frontend**: 源前端项目，基于 pure-admin-thin 6.2.0 的管理后台
- **Module_Page**: 一个管理模块的完整页面，包含列表展示、搜索筛选、CRUD 操作
- **API_Layer**: 封装后端 HTTP 请求的 TypeScript 模块，位于 src/api/ 目录
- **Router_Config**: Vue Router 路由配置，定义页面路径和导航结构
- **Permission_Directive**: 前端权限控制指令，用于按钮级别的显示/隐藏控制
- **Plugin_Container**: 动态加载外部插件页面的容器组件

## 需求

### 需求 1：部门管理模块

**用户故事：** 作为系统管理员，我需要管理组织部门的树形结构，以便建立清晰的组织架构。

#### 验收标准

1. WHEN 管理员进入部门管理页面, THE Dev_Web_Admin SHALL 以树形表格展示全部部门层级数据
2. WHEN 管理员点击新增按钮, THE Dev_Web_Admin SHALL 弹出表单对话框，支持填写部门名称、排序、负责人、父级部门等字段并提交至 /api/v1/dept
3. WHEN 管理员点击编辑按钮, THE Dev_Web_Admin SHALL 加载选中部门数据到表单对话框并提交修改至 /api/v1/dept/{id}
4. WHEN 管理员点击删除按钮, THE Dev_Web_Admin SHALL 弹出确认对话框，确认后调用 /api/v1/dept/{id} 删除接口
5. WHEN 管理员在搜索框输入关键字, THE Dev_Web_Admin SHALL 按部门名称过滤展示匹配的部门节点

### 需求 2：岗位管理模块

**用户故事：** 作为系统管理员，我需要管理系统岗位，以便为用户分配对应的职务。

#### 验收标准

1. WHEN 管理员进入岗位管理页面, THE Dev_Web_Admin SHALL 以分页表格展示岗位列表，包含岗位编码、名称、排序、状态字段
2. WHEN 管理员点击新增按钮, THE Dev_Web_Admin SHALL 弹出表单对话框，支持填写岗位编码、名称、排序、状态并提交至 /api/v1/post
3. WHEN 管理员点击编辑按钮, THE Dev_Web_Admin SHALL 加载选中岗位数据并提交修改至 /api/v1/post/{id}
4. WHEN 管理员点击删除按钮, THE Dev_Web_Admin SHALL 弹出确认对话框，确认后调用 /api/v1/post/{id} 删除接口
5. THE Dev_Web_Admin SHALL 支持按岗位名称和状态筛选岗位列表

### 需求 3：字典管理模块

**用户故事：** 作为系统管理员，我需要管理系统字典类型和字典数据，以便维护下拉框等组件的选项数据源。

#### 验收标准

1. WHEN 管理员进入字典管理页面, THE Dev_Web_Admin SHALL 以分页表格展示字典类型列表，包含字典名称、字典类型编码、状态、备注字段
2. WHEN 管理员点击新增字典类型按钮, THE Dev_Web_Admin SHALL 弹出表单对话框，支持创建字典类型并提交至 /api/v1/dict/type
3. WHEN 管理员点击某个字典类型的数据管理按钮, THE Dev_Web_Admin SHALL 展示该字典类型下的字典数据列表
4. WHEN 管理员在字典数据列表中点击新增, THE Dev_Web_Admin SHALL 弹出表单对话框，支持创建字典数据项并提交至 /api/v1/dict/data
5. WHEN 管理员编辑或删除字典类型或数据项, THE Dev_Web_Admin SHALL 调用对应的 PUT 或 DELETE 接口完成操作
6. THE Dev_Web_Admin SHALL 支持按字典名称和类型编码筛选字典类型列表

### 需求 4：系统配置模块

**用户故事：** 作为系统管理员，我需要管理系统参数配置，以便灵活控制系统的运行行为。

#### 验收标准

1. WHEN 管理员进入系统配置页面, THE Dev_Web_Admin SHALL 以分页表格展示配置列表，包含配置名称、配置键、配置值、是否内置字段
2. WHEN 管理员点击新增按钮, THE Dev_Web_Admin SHALL 弹出表单对话框，支持填写配置名称、键、值、备注并提交至 /api/v1/config
3. WHEN 管理员点击编辑按钮, THE Dev_Web_Admin SHALL 加载选中配置数据并提交修改至 /api/v1/config/{id}
4. WHEN 管理员点击删除按钮, THE Dev_Web_Admin SHALL 弹出确认对话框，确认后调用删除接口
5. THE Dev_Web_Admin SHALL 支持按配置名称和配置键筛选配置列表

### 需求 5：接口管理模块

**用户故事：** 作为系统管理员，我需要查看系统已注册的 API 接口列表，以便了解系统接口情况和进行权限分配。

#### 验收标准

1. WHEN 管理员进入接口管理页面, THE Dev_Web_Admin SHALL 以分页表格展示接口列表，包含接口路径、请求方式、接口标题、所属模块字段
2. THE Dev_Web_Admin SHALL 支持按接口路径和标题进行搜索筛选
3. THE Dev_Web_Admin SHALL 为接口列表提供只读展示功能（接口数据由后端自动注册，前端不提供增删改操作）

### 需求 6：权限演示模块

**用户故事：** 作为开发人员，我需要查看按钮级和页面级权限控制的演示页面，以便理解和验证系统权限控制机制。

#### 验收标准

1. WHEN 用户进入按钮权限页面, THE Dev_Web_Admin SHALL 展示当前用户拥有的权限 code 列表
2. THE Dev_Web_Admin SHALL 演示通过 v-auth 指令方式控制按钮可见性
3. THE Dev_Web_Admin SHALL 演示通过 hasAuth 函数方式控制按钮可见性
4. WHEN 用户进入页面权限页面, THE Dev_Web_Admin SHALL 演示基于角色的页面级别访问控制效果

### 需求 7：服务监控模块

**用户故事：** 作为系统管理员，我需要查看服务器运行状态，以便监控系统健康状况。

#### 验收标准

1. WHEN 管理员进入服务监控页面, THE Dev_Web_Admin SHALL 调用 /api/v1/server-monitor 并以卡片形式展示 CPU 信息（核心数、使用率）
2. THE Dev_Web_Admin SHALL 展示内存信息（总内存、已用内存、使用率）
3. THE Dev_Web_Admin SHALL 展示服务器信息（服务器名称、操作系统、服务器 IP）
4. IF 监控接口调用失败, THEN THE Dev_Web_Admin SHALL 展示空数据提示而非页面崩溃

### 需求 8：定时任务管理模块

**用户故事：** 作为系统管理员，我需要管理系统定时任务，以便控制后台周期性作业的执行。

#### 验收标准

1. WHEN 管理员进入定时任务页面, THE Dev_Web_Admin SHALL 以分页表格展示任务列表，包含任务名称、任务组、Cron 表达式、调用目标、状态字段
2. THE Dev_Web_Admin SHALL 支持按任务名称和任务组筛选任务列表
3. WHEN 管理员查看任务执行日志, THE Dev_Web_Admin SHALL 以分页表格展示执行日志，包含任务名称、任务组、调用目标、执行状态、执行时间字段
4. THE Dev_Web_Admin SHALL 通过 /api/v1/sysjob 和 /api/v1/sysjob/log 接口获取数据

### 需求 9：插件容器模块

**用户故事：** 作为系统管理员，我需要通过插件容器加载和展示外部插件页面，以便扩展系统功能。

#### 验收标准

1. WHEN 路由导航到插件页面, THE Dev_Web_Admin SHALL 根据路由 meta 中的 pluginName 或路径段解析插件名称
2. THE Dev_Web_Admin SHALL 从 /static/plugins/{pluginName}/index.js 动态加载插件模块
3. WHEN 插件模块加载成功且含有 routes 配置, THE Dev_Web_Admin SHALL 根据当前路径匹配并渲染对应的插件页面组件
4. WHILE 插件模块正在加载, THE Dev_Web_Admin SHALL 展示加载中的 loading 状态
5. IF 插件加载失败或不存在, THEN THE Dev_Web_Admin SHALL 展示"插件页面加载失败或不存在"的空状态提示

### 需求 10：开发工具模块

**用户故事：** 作为开发人员，我需要在后台管理系统中访问开发辅助工具（代码生成、构建、Swagger 文档），以便提高开发效率。

#### 验收标准

1. WHEN 开发人员进入代码生成页面, THE Dev_Web_Admin SHALL 展示代码生成功能占位页面（该功能暂未开放）
2. WHEN 开发人员进入构建页面, THE Dev_Web_Admin SHALL 展示构建工具功能占位页面
3. WHEN 开发人员进入 Swagger 页面, THE Dev_Web_Admin SHALL 通过 iframe 或跳转方式展示后端 Swagger 文档界面

### 需求 11：路由与导航集成

**用户故事：** 作为用户，我需要通过 dev-web-admin 的侧边栏菜单导航到所有新增模块页面，以便统一访问所有管理功能。

#### 验收标准

1. THE Dev_Web_Admin SHALL 在 Router_Config 中为所有新增模块注册路由，路由结构使用 AppLayout 作为父容器
2. THE Dev_Web_Admin SHALL 为每个新增路由配置 meta 信息，包含 title、icon、permission 字段
3. WHEN 用户无对应菜单权限时, THE Dev_Web_Admin SHALL 在侧边栏中隐藏无权限的菜单项
4. THE Dev_Web_Admin SHALL 保持已有模块（首页、基础数据、业主管理、工单管理、缴费管理、公告管理等）的路由和页面不变

### 需求 12：API 层封装

**用户故事：** 作为开发人员，我需要在 dev-web-admin 的 API 层中新增对应模块的接口封装，以便页面组件统一调用后端接口。

#### 验收标准

1. THE Dev_Web_Admin SHALL 在 src/api/ 目录下为每个新增模块创建对应的 API 文件
2. THE Dev_Web_Admin SHALL 对 go-admin 后端返回的 PageOK 格式（{ data: { list, count } }）进行统一适配
3. THE Dev_Web_Admin SHALL 为所有 API 函数提供 TypeScript 类型定义（请求参数和响应类型）
4. THE Dev_Web_Admin SHALL 复用已有的 axios 实例和请求拦截器，保持与已有模块一致的调用方式

### 需求 13：业务模块清理

**用户故事：** 作为开发人员，我需要移除 dev-web-admin 中与当前 Go 后端不匹配的特定业务模块，以便保持代码库整洁且只包含实际使用的功能。

#### 验收标准

1. THE Dev_Web_Admin SHALL 移除以下业务模块的视图、路由、API 文件和 store：base（片区/小区/楼栋/单元/房屋管理）、owner（业主管理/认证审批）、workorder（工单管理/工单统计）、billing（账单/逾期催收/费用配置/收费统计）、notice（公告列表/发布/审批）
2. THE Dev_Web_Admin SHALL 从 static-routes.ts 中移除上述模块的路由配置
3. THE Dev_Web_Admin SHALL 保留框架基础模块：home（首页）、login（登录）、init（系统初始化）、error（404）、profile（个人中心）、system（用户/角色/菜单/日志）、placeholder（占位页）
4. THE Dev_Web_Admin SHALL 确保移除后项目仍能正常编译运行，无遗留的 import 引用报错

### 需求 14：样式规范统一

**用户故事：** 作为开发人员，我需要所有迁移模块使用 dev-web-admin 自身的样式体系，以便保持整体视觉风格一致。

#### 验收标准

1. THE Dev_Web_Admin SHALL 所有新增模块页面使用 dev-web-admin 现有的样式规范（Element Plus 组件 + 项目 scoped CSS），禁止引入 TailwindCSS 或 frontend 的自定义样式
2. THE Dev_Web_Admin SHALL 新增页面的布局、间距、表格、表单、对话框等组件样式与 dev-web-admin 已有页面（如用户管理、角色管理）保持一致
3. THE Dev_Web_Admin SHALL 仅迁移 frontend 的业务逻辑和数据交互，不迁移其视觉样式和 UI 设计
