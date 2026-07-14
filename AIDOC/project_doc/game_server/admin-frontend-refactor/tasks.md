# 任务列表：admin-frontend-refactor

## 执行顺序

```
Task 1（清理旧模块） → Task 2（API 公共适配层） → Task 3~10（各模块并行） → Task 11（路由注册） → Task 12（编译验证）
```

依赖关系：
- Task 2 依赖 Task 1（清理后再新增）
- Task 3~10 依赖 Task 2（API 适配层就绪）
- Task 11 依赖 Task 3~10（所有页面就绪后注册路由）
- Task 12 依赖 Task 11（全量编译验证）

---

### Task 1: 清理旧业务模块

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `src/views/base/`（删除整个目录）
  - `src/views/owner/`（删除整个目录）
  - `src/views/workorder/`（删除整个目录）
  - `src/views/billing/`（删除整个目录）
  - `src/views/notice/`（删除整个目录）
  - `src/api/` 下对应模块的 API 文件（删除）
  - `src/stores/` 下对应模块的 store 文件（删除，如有）
  - `src/router/static-routes.ts`（修改：移除上述模块路由）
- 涉及模块: base, owner, workorder, billing, notice, router
- 不触碰: src/views/home/, src/views/system/(user/role/menu/log), src/views/login/, src/views/init/, src/views/error/, src/views/profile/, src/views/placeholder/, src/layout/, src/components/

**Constraints（约束）:**
- 删除前先确认各模块的文件清单，避免误删公共组件
- static-routes.ts 中只移除对应路由配置块，不修改其他路由
- 保留框架基础模块的所有路由和组件

**Acceptance（验证标准）:**
- AC: base/owner/workorder/billing/notice 的 views、api、store 文件全部移除
- AC: static-routes.ts 中不再包含上述模块的路由配置
- AC: 保留 home/login/init/error/profile/system(user/role/menu/log)/placeholder 路由
- AC: `npm run build` 或 `vite build` 零错误，无遗留 import 引用
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 2: 新增 API 公共适配层

**复杂度**: 低

**Scope（边界）:**
- 涉及文件: `src/api/helpers.ts`（新增）
- 不触碰: `src/utils/request.ts`（已有 axios 实例）

**Acceptance:**
- AC: `adaptPageResponse<T>(res)` 函数正确解析 `{ list: T[], count: number }` 格式
- AC: 数组格式兜底：传入纯数组时返回 `{ list, total: list.length }`
- AC: 空数据兜底：传入 null/undefined 时返回 `{ list: [], total: 0 }`
- AC: 导出 `PageQuery` 接口类型
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 3: 部门管理模块

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `src/api/dept.ts`（新增）
  - `src/views/system/dept/DeptList.vue`（新增）
  - `src/views/system/dept/DeptForm.vue`（新增）
- 涉及模块: system/dept
- 不触碰: src/views/system/user/, src/views/system/role/

**Acceptance（验证标准）:**
- AC: DeptList 以树形表格展示部门层级（el-table + row-key + tree-props）
- AC: 支持按部门名称搜索过滤
- AC: DeptForm 支持新增/编辑模式，通过 `open(row?)` 方法切换
- AC: API 调用路径正确：GET/POST /api/v1/dept, PUT/DELETE /api/v1/dept/{id}
- AC: 提供 getDeptTree 接口用于父级部门下拉选择
- AC: 页面样式遵循 dev-web-admin 规范（搜索卡片 + 表格卡片 + 表单对话框）
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 4: 岗位管理模块

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `src/api/post.ts`（新增）
  - `src/views/system/post/PostList.vue`（新增）
  - `src/views/system/post/PostForm.vue`（新增）
- 涉及模块: system/post
- 不触碰: 其他 system 子模块

**Acceptance（验证标准）:**
- AC: PostList 以分页表格展示岗位列表（编码、名称、排序、状态）
- AC: 支持按名称和状态筛选
- AC: PostForm 支持新增/编辑，通过 `open(row?)` 切换模式
- AC: API 调用路径正确：GET/POST /api/v1/post, PUT/DELETE /api/v1/post/{id}
- AC: 使用 adaptPageResponse 适配分页数据
- AC: 页面样式遵循 dev-web-admin 规范
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 5: 字典管理模块

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `src/api/dict.ts`（新增）
  - `src/views/system/dict/DictTypeList.vue`（新增）
  - `src/views/system/dict/DictTypeForm.vue`（新增）
  - `src/views/system/dict/DictDataList.vue`（新增）
  - `src/views/system/dict/DictDataForm.vue`（新增）
- 涉及模块: system/dict
- 不触碰: 其他 system 子模块

**Constraints（约束）:**
- 字典类型与字典数据为两级关系，需通过路由参数或对话框切换
- 字典数据列表必须关联字典类型的 dictType 字段进行过滤
- API 路径区分：/api/v1/dict/type（类型）和 /api/v1/dict/data（数据）

**Acceptance（验证标准）:**
- AC: DictTypeList 分页展示字典类型（名称、编码、状态、备注）
- AC: 支持按名称和类型编码筛选
- AC: 点击"数据管理"可查看该类型下的字典数据列表
- AC: DictDataList 分页展示字典数据（排序、标签、值、状态）
- AC: 字典数据的增删改调用 /api/v1/dict/data 接口
- AC: 页面样式遵循 dev-web-admin 规范
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 6: 系统配置模块

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `src/api/config.ts`（新增）
  - `src/views/system/config/ConfigList.vue`（新增）
  - `src/views/system/config/ConfigForm.vue`（新增）
- 涉及模块: system/config
- 不触碰: 其他 system 子模块

**Acceptance（验证标准）:**
- AC: ConfigList 分页展示配置列表（名称、键、值、是否内置）
- AC: 支持按配置名称和键筛选
- AC: ConfigForm 支持新增/编辑
- AC: API 调用路径正确：GET/POST /api/v1/config, PUT/DELETE /api/v1/config/{id}
- AC: 使用 adaptPageResponse 适配分页数据
- AC: 页面样式遵循 dev-web-admin 规范
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 7: 接口管理模块

**复杂度**: 低

**Scope（边界）:**
- 涉及文件:
  - `src/api/sys-api.ts`（新增）
  - `src/views/system/api/ApiList.vue`（新增）
- 不触碰: 其他 system 子模块

**Acceptance:**
- AC: ApiList 分页展示接口列表（路径、方法、标题、所属模块）
- AC: 支持按路径和标题搜索
- AC: 只读展示，不提供增删改操作按钮
- AC: API 调用路径：GET /api/v1/sys-api
- AC: 页面样式遵循 dev-web-admin 规范
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 8: 服务监控模块

**复杂度**: 低

**Scope（边界）:**
- 涉及文件:
  - `src/api/monitor.ts`（新增）
  - `src/views/monitor/server/ServerMonitor.vue`（新增）
- 不触碰: 其他模块

**Acceptance:**
- AC: ServerMonitor 以卡片形式展示 CPU、内存、服务器信息
- AC: 接口失败时展示空数据提示，不崩溃
- AC: API 调用路径：GET /api/v1/server-monitor
- AC: 页面样式遵循 dev-web-admin 规范
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 9: 定时任务管理模块

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `src/api/job.ts`（新增）
  - `src/views/job/list/JobList.vue`（新增）
  - `src/views/job/log/JobLog.vue`（新增）
- 涉及模块: job
- 不触碰: 其他模块

**Acceptance（验证标准）:**
- AC: JobList 分页展示任务列表（名称、组、Cron、目标、状态）
- AC: 支持按名称和组筛选
- AC: JobLog 分页展示执行日志（名称、组、目标、状态、时间）
- AC: API 调用路径：GET /api/v1/sysjob, GET /api/v1/sysjob/log
- AC: 使用 adaptPageResponse 适配分页数据
- AC: 页面样式遵循 dev-web-admin 规范
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 10: 权限演示 + 插件容器 + 开发工具模块

**复杂度**: 中

**Scope（边界）:**
- 涉及文件:
  - `src/views/permission/button/ButtonPermission.vue`（新增）
  - `src/views/permission/page/PagePermission.vue`（新增）
  - `src/views/plugin/PluginContainer.vue`（新增）
  - `src/views/devtools/codegen/CodegenPlaceholder.vue`（新增）
  - `src/views/devtools/build/BuildPlaceholder.vue`（新增）
  - `src/views/devtools/swagger/SwaggerView.vue`（新增）
- 涉及模块: permission, plugin, devtools
- 不触碰: 其他模块

**Acceptance（验证标准）:**
- AC: ButtonPermission 展示权限 code 列表，演示 v-auth 指令和 hasAuth 函数
- AC: PagePermission 演示页面级权限控制
- AC: PluginContainer 根据路由 meta 解析 pluginName，动态加载 /static/plugins/{name}/index.js
- AC: 插件加载中显示 loading，失败显示空状态提示
- AC: CodegenPlaceholder/BuildPlaceholder 为占位页
- AC: SwaggerView 通过 iframe 展示后端 Swagger
- AC: 页面样式遵循 dev-web-admin 规范
- AC: `vue-tsc --noEmit` 类型检查通过

---

### Task 11: 路由注册与导航集成

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `src/router/static-routes.ts`（修改：新增所有模块路由）
- 涉及模块: router
- 不触碰: src/router/index.ts 的路由守卫逻辑、src/layout/ 组件

**Constraints（约束）:**
- 所有新增路由使用 AppLayout 作为父容器
- meta 必须包含 title、icon、permission 字段
- 路由 name 使用 kebab-case 格式
- 保持已有 system/user, system/role, system/menu, system/log 路由不变

**Acceptance（验证标准）:**
- AC: 所有新增模块在侧边栏正确显示菜单项
- AC: 每个路由 meta 包含 title + icon + permission
- AC: 插件路由支持 meta.pluginName 参数传递
- AC: 已有路由（home/system-user/system-role/system-menu/system-log）保持不变
- AC: 无权限用户看不到对应菜单
- AC: `vue-tsc --noEmit` 类型检查通过
- AC: 【回归】已有页面（用户管理、角色管理、菜单管理、日志）正常访问

---

### Task 12: 全量编译验证与收尾

**复杂度**: 低

**Acceptance:**
- AC: `vue-tsc --noEmit` 全量类型检查通过
- AC: `vite build` 构建零错误
- AC: 无遗留的未使用 import 或未引用变量
- AC: 旧模块路由访问返回 404
- AC: 所有新模块页面可正常导航到达
