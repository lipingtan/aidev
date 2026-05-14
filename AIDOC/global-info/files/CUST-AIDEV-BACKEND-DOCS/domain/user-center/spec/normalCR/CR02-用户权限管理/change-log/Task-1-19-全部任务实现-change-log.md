# Task 1-19 用户权限管理全部任务实现 - 变更记录

## 基本信息

| 项目 | 内容 |
|---|---|
| 任务名称 | 任务 1-19：实现 user-center 模块全部组件 |
| 变更时间 | 2026-04-18 |
| 关联 CR | CR02-用户权限管理 |

## 变更概要

共新增/修改约 90+ 个文件，涵盖 common 模块变更 + user 模块完整实现。

### 任务 1：common 模块变更

| 文件路径 | 操作 | 变更内容摘要 |
|---|---|---|
| com/spmp/common/security/DataPermissionLevel.java | 修改 | 增加 SELF 枚举值 |
| com/spmp/common/security/DataPermissionContext.java | 修改 | 重构为通用化设计（scopeMap + userId + username + ThreadLocal） |
| com/spmp/common/security/DataPermission.java | 修改 | 增加 selfField 参数 |
| com/spmp/common/security/DataPermissionInterceptor.java | 修改 | 支持 SELF 级别、scopeMap 遍历、IN 查询 |
| com/spmp/common/security/JwtTokenProvider.java | 修改 | generateToken/generateRefreshToken 增加 jti |
| com/spmp/common/security/JwtAuthenticationFilter.java | 修改 | 增加 Token 黑名单校验（Redis 降级放行） |
| com/spmp/common/security/SecurityConfig.java | 修改 | 白名单追加 user 模块认证路径 |
