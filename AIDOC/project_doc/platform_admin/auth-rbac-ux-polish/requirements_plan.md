# 需求计划：auth-rbac UI/UX 交互优化

## 需求理解

- 目标：对 auth-rbac 前端管理页面进行产品经理 Review 后的 UI/UX 优化，提升用户操作效率和体验一致性
- 范围：dev-web-admin 前端工程，涉及认证流程、列表页交互、角色管理、树形管理、全局体验 5 个维度
- 预期效果：修复 18 个已识别的交互问题，使系统达到生产可用的体验标准

## 假设列表

- [假设-1] 所有优化仅涉及前端，不需要修改后端 API
- [假设-2] 优化不改变现有功能逻辑，仅改善交互体验
- [假设-3] 暗色主题（luxury-dark）需要同步适配

## 澄清问题

- [Question-1] 撤回删除功能（#7）是否需要后端支持"恢复软删除"接口，还是仅做前端延迟删除？
  [Answer-1]
不做撤回删除
- [Question-2] 全局 loading 进度条（#15）选用 NProgress 还是 Element Plus 的 el-loading？
  [Answer-2]
el-loading 
- [Question-3] 操作日志 JSON diff 高亮（#17）是否引入第三方 diff 库（如 jsondiffpatch），还是手写简单比较？
  [Answer-3]
引入 jsondiffpatch
## 非功能需求建议

- 性能：虚拟滚动（#13）应在列表超过 100 条时启用
- 一致性：分页 page_size 全局统一为 20
- 可访问性：拖拽区域需提供键盘操作替代方案

## 影响范围预判

- 涉及模块：layout、views/system/*、utils/request.ts、router/guard.ts
- 涉及文件（预估）：~15 个 Vue 组件 + 2 个工具文件
- 可能的副作用：无，纯 UI 层优化

## 问题清单（按优先级排序）

### 高优先级（影响核心流程）

| # | 问题 | 页面 | 建议 |
|---|------|------|------|
| 5 | 用户页操作列 5 按钮溢出 | UserList | 改为"更多"下拉菜单 |
| 10 | 角色权限 Tab 切换无保存提示 | RoleList | 切换前检测未保存变更 |
| 12 | 资源树 appList 硬编码 | ResourceTree | 从后端动态加载 |

### 中优先级（影响效率/一致性）

| # | 问题 | 页面 | 建议 |
|---|------|------|------|
| 1 | 登录失败焦点不回到密码框 | LoginView | 失败后 focus 密码输入框 |
| 2 | 租户选择页刷新后 tenants 丢失 | tenant-select | 从 localStorage 恢复或重新请求 |
| 3 | 切换租户无二次确认 | AppHeader | 增加 confirm 对话框 |
| 6 | 表格无空状态 | 全部列表页 | 添加 el-empty |
| 8 | 分页 page_size 不统一 | 全部列表页 | 统一为 20 |
| 9 | 角色树固定 320px 截断 | RoleList | 增加 tooltip 或可拖拽宽度 |
| 11 | 删除角色错误判断用字符串匹配 | RoleList | 改用 code === 40006 |
| 13 | 接口权限右侧无分页/虚拟滚动 | ApiPermissionTree | 数据多时添加虚拟滚动 |
| 14 | 拖拽无视觉引导 | ResourceTree / ApiPermission | 增加 drag handle + 引导提示 |

### 低优先级（美观/体验增强）

| # | 问题 | 页面 | 建议 |
|---|------|------|------|
| 4 | Token 刷新无交互反馈 | request.ts | 全局 loading 遮罩 |
| 7 | 删除无撤回机会 | 全部列表页 | 5s 内可撤回 |
| 15 | 无全局路由切换进度条 | 全局 | NProgress |
| 16 | 无网络异常兜底页 | 全局 | 错误检测 + 重试 |
| 17 | 操作日志 JSON 无 diff 高亮 | OperationLogList | 红绿色差异标注 |
| 18 | 暗色主题 el-transfer 未适配 | UserList | 补充 dark 模式样式 |
