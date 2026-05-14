# Task 21-27 前端实现及后端接口补全 变更日志

## 基本信息

| 项目 | 内容 |
|---|---|
| 任务名称 | Task 21-27 前端页面实现 + 后端缺失接口补全 |
| 变更时间 | 2026-04-18 |
| 关联 CR | CR02-用户权限管理 |

## 后端补全（Java 文件变更）

| 文件路径 | 变更类型 | 变更摘要 |
|---|---|---|
| `src/spmp-backend/.../service/UserService.java` | 修改 | 新增 `batchUpdateStatus(List<Long>, Integer)` 接口方法 |
| `src/spmp-backend/.../service/RoleService.java` | 修改 | 新增 `batchDeleteRoles(List<Long>)` 接口方法 |
| `src/spmp-backend/.../service/impl/UserServiceImpl.java` | 修改 | 实现 `batchUpdateStatus` 方法，循环调用 `updateStatus` |
| `src/spmp-backend/.../service/impl/RoleServiceImpl.java` | 修改 | 实现 `batchDeleteRoles` 方法，循环调用 `deleteRole` |
| `src/spmp-backend/.../controller/UserController.java` | 修改 | 新增 `PUT /users/batch-status` 批量状态切换接口 |
| `src/spmp-backend/.../controller/RoleController.java` | 修改 | 新增 `DELETE /roles/batch` 批量删除角色接口 |

## 前端新增（TypeScript/Vue 文件）

| 文件路径 | 变更类型 | 变更摘要 |
|---|---|---|
| `src/spmp-web-pc/src/api/auth.ts` | 新增 | 认证 API（login、logout、refreshToken、getCaptcha、sendSmsCode） |
| `src/spmp-web-pc/src/api/user.ts` | 新增 | 用户管理 API（listUsers、createUser、updateUser、deleteUser、batchDelete、updateStatus、resetPassword） |
| `src/spmp-web-pc/src/api/role.ts` | 新增 | 角色管理 API（listRoles、listAllRoles、createRole、updateRole、deleteRole、getRoleMenuIds、assignMenus、configDataPermission） |
| `src/spmp-web-pc/src/api/menu.ts` | 新增 | 菜单管理 API（getMenuTree、getUserMenuTree、createMenu、updateMenu、deleteMenu） |
| `src/spmp-web-pc/src/api/profile.ts` | 新增 | 个人中心 API（getProfile、updateProfile、updatePassword） |
| `src/spmp-web-pc/src/api/log.ts` | 新增 | 日志 API（listLoginLogs、listOperationLogs） |
| `src/spmp-web-pc/src/api/data-permission.ts` | 新增 | 数据权限 API（getDataPermissionOptions） |
| `src/spmp-web-pc/src/store/modules/user.ts` | 修改 | 从 mock 改为对接真实后端 API，增加 permissions/menus 状态 |
| `src/spmp-web-pc/src/views/system/user/UserList.vue` | 新增 | 用户列表页（表格+搜索+分页+状态切换+重置密码+批量删除） |
| `src/spmp-web-pc/src/views/system/user/UserForm.vue` | 新增 | 用户新增/编辑表单对话框（含角色选择） |
| `src/spmp-web-pc/src/views/system/role/RoleList.vue` | 新增 | 角色列表页（表格+搜索+分页） |
| `src/spmp-web-pc/src/views/system/role/RoleForm.vue` | 新增 | 角色新增/编辑表单对话框 |
| `src/spmp-web-pc/src/views/system/role/MenuAssign.vue` | 新增 | 菜单权限分配（树形勾选对话框） |
| `src/spmp-web-pc/src/views/system/role/DataPermissionConfig.vue` | 新增 | 数据权限配置对话框 |
| `src/spmp-web-pc/src/views/system/menu/MenuList.vue` | 新增 | 菜单树形列表页（树形表格+新增/编辑/删除） |
| `src/spmp-web-pc/src/views/system/menu/MenuForm.vue` | 新增 | 菜单新增/编辑表单对话框 |
| `src/spmp-web-pc/src/views/profile/ProfilePage.vue` | 新增 | 个人信息展示+修改+密码修改 |
| `src/spmp-web-pc/src/views/system/log/LoginLogList.vue` | 新增 | 登录日志列表（表格+搜索+分页） |
| `src/spmp-web-pc/src/views/system/log/OperationLogList.vue` | 新增 | 操作日志列表（表格+搜索+分页） |
| `src/spmp-web-pc/src/router/static-routes.ts` | 修改 | 添加系统管理/日志管理/个人中心子路由 |
| `src/spmp-web-pc/src/router/guard.ts` | 修改 | 完善登录状态检查和用户信息/菜单自动加载 |
| `src/spmp-web-pc/src/layout/AppSidebar.vue` | 修改 | 侧边栏支持分组子菜单渲染（系统管理、日志管理） |

## tasks.md 状态更新

所有必需任务（Task 1-20 后端 + Task 21-27 前端）已标记为 `[x]` 完成。仅剩 12 个可选单元测试子任务（带 `*` 标记）未实现。
