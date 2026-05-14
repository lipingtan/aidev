# SPMP user-center API 文档

> 模块包名：`com.spmp.user`  
> API 前缀：`/api/v1/user/`  
> 统一响应：`Result<T>` / `PageResult<T>`

---

## 一、认证接口（白名单，无需 Token）

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/user/auth/captcha` | 获取图形验证码 | 公开 |
| POST | `/api/v1/user/auth/login` | 用户名密码登录 | 公开 |
| POST | `/api/v1/user/auth/login/sms` | 手机号验证码登录 | 公开 |
| POST | `/api/v1/user/auth/sms-code` | 发送短信验证码 | 公开 |
| POST | `/api/v1/user/auth/refresh` | 刷新 Token | 公开 |
| POST | `/api/v1/user/auth/logout` | 登出 | 需认证 |

---

## 二、用户管理接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/users` | 用户分页查询 | `user:user:list` |
| POST | `/api/v1/user/users` | 新增用户 | `user:user:create` |
| PUT | `/api/v1/user/users/{id}` | 编辑用户 | `user:user:edit` |
| DELETE | `/api/v1/user/users/{id}` | 删除用户（逻辑删除） | `user:user:delete` |
| DELETE | `/api/v1/user/users/batch` | 批量删除用户 | `user:user:delete` |
| PUT | `/api/v1/user/users/{id}/status` | 用户状态切换 | `user:user:edit` |
| PUT | `/api/v1/user/users/batch-status` | 批量启用/禁用 | `user:user:edit` |
| PUT | `/api/v1/user/users/{id}/reset-password` | 重置密码 | `user:user:reset-pwd` |

---

## 三、角色管理接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/roles` | 角色分页查询 | `user:role:list` |
| GET | `/api/v1/user/roles/list` | 角色全量列表（下拉用） | `user:role:list` |
| POST | `/api/v1/user/roles` | 新增角色 | `user:role:create` |
| PUT | `/api/v1/user/roles/{id}` | 编辑角色 | `user:role:edit` |
| DELETE | `/api/v1/user/roles/{id}` | 删除角色 | `user:role:delete` |
| DELETE | `/api/v1/user/roles/batch` | 批量删除角色 | `user:role:delete` |
| GET | `/api/v1/user/roles/{id}/menus` | 查询角色已分配菜单 ID | `user:role:list` |
| PUT | `/api/v1/user/roles/{id}/menus` | 分配菜单权限 | `user:role:assign` |
| PUT | `/api/v1/user/roles/{id}/data-permission` | 配置数据权限 | `user:role:assign` |

---

## 四、菜单管理接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/menus/tree` | 菜单树查询（完整） | `user:menu:list` |
| GET | `/api/v1/user/menus/user-tree` | 当前用户菜单树 | 需认证 |
| POST | `/api/v1/user/menus` | 新增菜单 | `user:menu:create` |
| PUT | `/api/v1/user/menus/{id}` | 编辑菜单 | `user:menu:edit` |
| DELETE | `/api/v1/user/menus/{id}` | 删除菜单 | `user:menu:delete` |

---

## 五、个人中心接口

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/user/profile` | 获取个人信息 | 需认证 |
| PUT | `/api/v1/user/profile` | 修改个人信息 | 需认证 |
| PUT | `/api/v1/user/profile/password` | 修改密码 | 需认证 |

---

## 六、数据权限接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/data-permission/options` | 数据权限可选范围 | `user:role:assign` |

---

## 七、日志查询接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/login-logs` | 登录日志分页查询 | `user:log:list` |
| GET | `/api/v1/user/operation-logs` | 操作日志分页查询 | `user:log:list` |

---

## 八、对外 API 接口（供其他模块调用）

单体应用内通过注入接口调用，定义在 `com.spmp.user.api` 包下。

### UserApi

| 方法 | 签名 | 说明 | 调用方 |
|------|------|------|--------|
| getUserById | `UserBriefDTO getUserById(Long userId)` | 查询用户基本信息 | workorder、notice、billing |
| getUsersByRoleCode | `List<UserBriefDTO> getUsersByRoleCode(String roleCode)` | 按角色查用户列表 | workorder（查维修人员） |
| getUsersByIds | `List<UserBriefDTO> getUsersByIds(List<Long> userIds)` | 批量查询用户信息 | 所有业务模块 |

### PermissionApi

| 方法 | 签名 | 说明 | 调用方 |
|------|------|------|--------|
| getDataPermission | `DataPermissionDTO getDataPermission(Long userId)` | 获取用户数据权限范围 | 所有业务模块 |
| checkPermission | `boolean checkPermission(Long userId, String permCode)` | 校验用户权限标识 | 所有业务模块 |
| getUserRoles | `List<String> getUserRoles(Long userId)` | 获取用户角色编码列表 | 所有业务模块 |

### 跨模块 DTO（`com.spmp.user.api.dto`）

- `UserBriefDTO`：id、username、realName、phone（脱敏）
- `DataPermissionDTO`：level、scopeMap、userId
