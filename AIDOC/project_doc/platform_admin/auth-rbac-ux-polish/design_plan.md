# 设计计划：auth-rbac UI/UX 交互优化

## 设计方向

纯前端优化，不涉及后端 API 变更。所有修改在 `dev-web-admin/` 中完成，采用渐进增强策略——逐个组件修改，确保每次修改不破坏现有功能。

## 技术选型

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| jsondiffpatch（JSON diff） | 成熟库，支持 HTML 输出，带样式 | 包体积 ~30KB gzip | ✓ |
| 手写 diff | 无依赖 | 工作量大，边界处理复杂 | ✗ |
| el-loading（全局 loading） | 与项目 Element Plus 统一 | 样式稍重 | ✓ |
| NProgress | 轻量顶部进度条 | 与 Element Plus 风格不统一 | ✗ |
| 虚拟滚动方案：el-tree-v2 | Element Plus 原生支持 | API 与 el-tree 略有差异 | ✓ |

## 澄清问题

无（前序 requirements_plan 已全部澄清）

## 风险点

- [Risk-1] el-tree-v2 虚拟滚动不支持拖拽，接口权限页可能需要用 el-tree + 手动分页替代
- [Risk-2] jsondiffpatch 的 HTML 输出需要 v-html 渲染，注意 XSS 防护（diff 数据来自后端 JSON，风险可控）
- [Risk-3] 未保存检测（FR-2）需要在每个权限 Tab 组件中维护 dirty 状态，组件间通信增加复杂度

## 实现策略

按影响范围分 4 波执行：

1. **全局基础设施**（FR-13/14/15）：el-loading 全局指令 + 路由守卫集成 + 网络异常兜底
2. **列表页统一优化**（FR-1/7/8）：操作列重构 + el-empty + page_size 统一
3. **树形/角色页优化**（FR-2/3/9/10/11/12）：Tab 保存检测 + 动态加载 + tooltip + 拖拽引导
4. **认证流程 + 主题**（FR-4/5/6/16/17）：登录焦点 + 租户恢复 + 切换确认 + diff 高亮 + dark 适配
