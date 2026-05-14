# Task 6/16 登录日志记录补全 变更日志

## 基本信息

| 项目 | 内容 |
|---|---|
| 任务名称 | Task 6 认证模块 + Task 16 登录日志模块 — 登录日志记录逻辑补全 |
| 变更时间 | 2026-04-18 |
| 关联 CR | CR02-用户权限管理 |

## 变更的代码文件列表

| 文件路径 | 变更类型 | 变更摘要 |
|---|---|---|
| `src/spmp-backend/.../service/impl/AuthServiceImpl.java` | 修改 | 注入 LoginLogService 依赖，登录成功时调用 saveLoginLog 记录成功日志，登录失败（recordLoginFail）时记录失败日志，新增 saveLoginLog 和 getClientIp 辅助方法 |
