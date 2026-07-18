# 任务列表：auth-rbac-ux-polish（UI/UX 交互优化）

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 8 |
| 已完成 | 8 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 8/8 (100%) |
| 当前阶段 | ✅ 全部完成 |

### 状态说明

| 标记 | 状态 | 含义 |
|------|------|------|
| ✅ | 已完成 | 代码修改完成 + 手动验证通过 |
| 🔨 | 进行中 | 正在编码 |
| ⬜ | 未开始 | 尚未启动 |

---

## Wave 1: 全局基础设施（无依赖，可并行）

### Task 1: 全局 Loading 服务 + 路由切换 Loading ✅

**复杂度**: 中

**Scope:**
- 涉及文件: dev-web-admin/src/utils/loading.ts(新建), src/router/guard.ts(增强), src/utils/request.ts(增强)

**Acceptance:**
- AC: 新建 loading.ts 提供 showLoading/hideLoading（引用计数防重复）
- AC: Token 刷新期间显示全屏 el-loading
- AC: 路由切换时显示/隐藏 loading
- AC: 不影响现有 request 拦截器逻辑（RG-3）

---

### Task 2: 网络异常兜底页 ✅

**复杂度**: 低

**Scope:**
- 涉及文件: dev-web-admin/src/views/error/NetworkError.vue(新建), src/router/static-routes.ts(增加路由), src/utils/request.ts(增强)

**Acceptance:**
- AC: 请求超时或 network error 时跳转 /network-error
- AC: 页面显示图标 + "网络异常" + 重试按钮
- AC: 点击重试回到上一页面

---

## Wave 2: 列表页统一优化（无依赖，可并行）

### Task 3: 操作列优化 + 空状态 + page_size 统一 ✅

**复杂度**: 中

**Scope:**
- 涉及文件: src/views/system/user/UserList.vue, src/views/system/tenant/TenantList.vue, src/views/system/operation-log/OperationLogList.vue, src/views/system/application/ApplicationList.vue, src/views/system/data-scope/DataScopeConfigList.vue

**Acceptance:**
- AC: UserList 操作列改为"编辑/删除" + "更多"下拉（FR-1）
- AC: 全部列表页 el-table 增加 el-empty 空状态（FR-7）
- AC: 全部列表页 page_size 默认值统一为 20（FR-8）
- AC: 不影响现有 CRUD 功能（RG-1）

---

## Wave 3: 树形/角色页优化

### Task 4: 角色管理 Tab 未保存检测 + 错误码判断 ✅

**依赖**: 无

**复杂度**: 高

**Scope:**
- 涉及文件: src/views/system/role/RoleList.vue, src/views/system/role/MenuPermTab.vue, src/views/system/role/ApiPermTab.vue, src/views/system/role/DataScopeTab.vue, src/views/system/role/AppBindTab.vue

**Acceptance:**
- AC: 4 个 Tab 组件暴露 isDirty 状态（FR-2）
- AC: 切换 Tab 时检测 isDirty，弹出确认框
- AC: 角色树节点名称截断 + tooltip（FR-9）
- AC: 删除角色错误判断改用 code === 40006（FR-10）

---

### Task 5: 资源树动态 appList + 拖拽引导 ✅

**依赖**: 无

**复杂度**: 中

**Scope:**
- 涉及文件: src/views/system/resource/ResourceTree.vue, src/views/system/api-permission/ApiPermissionTree.vue

**Acceptance:**
- AC: ResourceTree appList 从 GET /api/v1/applications 动态加载（FR-3）
- AC: 树节点增加拖拽 handle 图标（FR-12）
- AC: 不影响拖拽排序功能（RG-2）

---

### Task 6: 接口权限右侧分页 ✅

**依赖**: 无

**复杂度**: 中

**Scope:**
- 涉及文件: src/views/system/api-permission/ApiPermissionTree.vue

**Acceptance:**
- AC: 右侧未分配列表增加搜索框 + 分页（每页 50 条）（FR-11）
- AC: 数据量大时不卡顿

---

## Wave 4: 认证流程 + 主题适配

### Task 7: 认证流程优化 ✅

**依赖**: Task 1

**复杂度**: 中

**Scope:**
- 涉及文件: src/views/login/LoginView.vue, src/views/tenant-select/index.vue, src/layout/AppHeader.vue, src/store/modules/auth.ts

**Acceptance:**
- AC: 登录失败后 focus 密码框 + 清空密码（FR-4）
- AC: 租户选择页刷新时从 localStorage 恢复 tenants 或重新请求（FR-5）
- AC: 切换租户增加二次确认对话框（FR-6）
- AC: 不影响路由守卫认证逻辑（RG-4）

---

### Task 8: JSON Diff 高亮 + 暗色主题适配 ✅

**依赖**: 无

**复杂度**: 中

**Scope:**
- 涉及文件: src/utils/json-diff.ts(新建), src/views/system/operation-log/OperationLogList.vue, src/styles/global.css

**Acceptance:**
- AC: 安装 jsondiffpatch 依赖
- AC: 日志详情弹窗用 diff 高亮替代纯 JSON 展示（FR-16）
- AC: el-transfer 在 luxury-dark 主题下样式适配（FR-17）
- AC: 暗色主题已有样式不退化（RG-5）

---

## 全局约束

- 所有修改在 `dev-web-admin/` 目录下
- 修改后确认无 TypeScript 编译错误
- 每个 Task 完成后手动验证对应 AC
- 不修改后端代码
