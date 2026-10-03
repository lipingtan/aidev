# 任务：登录失败 IP 黑名单（v2-cr11）

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 8 |
| 已完成 | 8 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 8/8 (100%) |
| 当前阶段 | Phase 5: 产品经理验收（待触发/执行中） |

> 2026-10-01 全部任务完成。实测记录见文末「执行记录」。

## 任务列表

### Task 1: 黑名单模型 + 计数器核心 ✅

**复杂度**: 中

**设计来源:**
- design.md #数据库设计 admin_login_blacklist
- design.md #核心逻辑 失败计数器 / Guard 结构

**Scope（边界）:**
- 新增: common/auth/model/login_blacklist.go, common/auth/service/login_attempt_guard.go
- 不触碰: 现有 model 文件、auth_service.go

**Constraints（约束）:**
- 雪花 ID 走 model.NextID()（BeforeCreate 钩子，同 LoginLog）
- 计数器必须带 mutex；窗口 30 分钟滚动（go-conventions.md 错误处理与并发约束）

**Acceptance（验证标准）:**
- AC: Model 字段与设计一致，TableName=admin_login_blacklist
- AC: 计数器滚动窗口/清零/组合键隔离单测通过
- AC: go build ./... 零错误

### Task 2: 黑名单服务（Ban/CheckBanned/Unban/List）✅

**复杂度**: 高

**设计来源:**
- design.md #核心逻辑 黑名单服务

**Scope（边界）:**
- 新增: common/auth/service/login_blacklist_service.go
- 不触碰: TokenBlacklist 相关文件

**Constraints（约束）:**
- Redis 走 storage.AdapterCache（sdk.Runtime.GetCacheAdapter()），key=loginbl:{ip_hash}（sha256）
- Ban 幂等（已有未过期记录不刷新）；Unban 双清（DB+Redis）任一失败返回错误
- CheckBanned 分层降级：Redis→DB→放行；DB 故障不阻断（正确性属性）

**Acceptance（验证标准）:**
- AC: SQLite 内存 DB 单测：Ban→CheckBanned 命中；过期后不命中且记录被清理
- AC: 幂等验证：重复 Ban 不改 expires_at
- AC: Unban 后 CheckBanned 立即 false
- AC: go test ./common/auth/service/... 全绿

### Task 3: 登录链路接入 + 错误码 ✅

**复杂度**: 高

**设计来源:**
- design.md #核心逻辑 登录链路改造
- design.md #API 设计 错误码 40305

**Scope（边界）:**
- 修改: common/auth/handler/auth_handler.go, common/auth/errors/errors.go, common/auth/auth.go（装配）
- 不触碰: auth_service.go 的 Login 主体逻辑、user_auth 链路

**Constraints（约束）:**
- 黑名单校验在验证码校验之后、authSvc.Login 之前（RG-3）
- 阈值/时长经 adminConfigSvc.ResolveInt(0, key, default) 读取（缓存链路）
- Ban 失败仅记日志，登录错误响应不受影响（降级放行）
- 拦截时写登录日志 status=0（复用 CreateAsync，RG-4）

**Acceptance（验证标准）:**
- AC: 阈值内失败返回 40101 原样（RG-2）；登录成功清零计数
- AC: 达阈值后同 IP 再登录（任意用户名/正确密码）返回 40305 含剩余分钟
- AC: 阈值=0 时计数与封禁均不发生（RG-5）
- AC: go build ./... 零错误，既有 auth handler 测试不回归（RG-1）

### Task 4: 后端路由 + 菜单/权限/配置 seed ✅

**复杂度**: 中

**设计来源:**
- design.md #API 设计 / #菜单与权限 seed / #数据库设计 配置 seed

**Scope（边界）:**
- 修改: common/auth/router.go（Dependencies+RegisterRoutes）, common/auth/auth.go（autoMigrate+装配+seed 调用）, common/auth/seed.go（SeedCR11Menus+配置 seed）, common/auth/discovery/auto_discover.go（4 张 map）
- 不触碰: 其他 CR 的 seed 函数

**Constraints（约束）:**
- seed 幂等（insert-if-not-exists，模式同 SeedTokenManagementMenu）
- 菜单挂日志管理（/log）下，ModuleCode=log-mgmt，PermissionCode=log:blacklist:list
- AutoMigrate 追加 LoginBlacklist 模型

**Acceptance（验证标准）:**
- AC: GET/DELETE /api/v1/admin/login-blacklist 路由注册成功，permission code 映射齐全
- AC: seed 可重复执行不产生重复行
- AC: go build ./... 零错误

### Task 5: 前端 API + 黑名单列表页 ✅

**复杂度**: 中

**设计来源:**
- design.md #前端 列表页

**Scope（边界）:**
- 修改: dev-web-admin/src/api/log.ts
- 新增: dev-web-admin/src/views/system/log/LoginBlacklistList.vue
- 不触碰: 其他页面组件

**Constraints（约束）:**
- 结构复刻 LoginLogList.vue（搜索卡片+表格+分页，UI/UX 与现有页面一致）
- 剩余时间前端计算（expires_at - now，分钟取整；已过期显示"—"）
- 解封需 confirm 二次确认，成功后刷新列表

**Acceptance（验证标准）:**
- AC: 列表列齐全（IP/触发用户名/失败次数/封禁时间/预计解封时间/剩余时间/状态/操作）
- AC: IP 模糊搜索与分页可用
- AC: vue-tsc/vite build 无类型错误

### Task 6: 前端路由 ✅

**复杂度**: 低

**设计来源:**
- design.md #前端 路由

**Scope（边界）:**
- 修改: dev-web-admin/src/router/static-routes.ts
- 不触碰: guard.ts、其他路由

**Constraints（约束）:**
- 日志管理分组内，permission: 'log:blacklist:list'

**Acceptance（验证标准）:**
- AC: /log/login-blacklist 可达，无权限时不显示
- AC: static-routes 测试不回归

### Task 7: 后端自测（真实性规范）✅

**复杂度**: 高

**设计来源:**
- design.md 全部 AC 场景

**Scope（边界）:**
- 新增: common/auth/service/login_blacklist_service_test.go, common/auth/service/login_attempt_guard_test.go
- 复用: github.com/glebarez/sqlite 内存库模式（同 operation_log_service_test.go）

**Constraints（约束）:**
- 遵循 go-testing.md 自测真实性强制规范：真实 SQLite 数据层（T-5）、验证副作用（T-3）、覆盖异常分支（T-2）、禁止空测试/t.Skip
- Redis 层用可注入的 mock adapter（storage.AdapterCache 接口实现），验证 key/TTL/DEL 行为

**Acceptance（验证标准）:**
- AC: ST-1 计数器：连续失败计数递增、成功清零、窗口过期归零、不同组合互不影响
- AC: ST-2 封禁触发：达阈值 Ban 写库（验证 banned_at/expires_at/fail_count 字段实际值）
- AC: ST-3 幂等：二次达标不刷新 expires_at（查库比对）
- AC: ST-4 拦截：CheckBanned 未过期 true/过期 false；过期记录被异步清理（查库验证）
- AC: ST-5 解封：Unban 后记录软删+缓存 DEL 被调用
- AC: ST-6 降级：Redis nil 与 DB 故障时不 panic、CheckBanned 按 false 处理
- AC: go test ./common/auth/... 全部通过

### Task 8: 构建 + 全链路验证 ✅

**复杂度**: 中

**设计来源:**
- 全部 RG-N 回归验证

**Scope（边界）:**
- 后端 go build/vet/test；前端 vite build
- 起后端实测登录→封禁→解封全链路（若本机可起服务）

**Constraints（约束）:**
- 实测必须覆盖：正常登录、错密码 N 次→封禁→40305→解封→恢复登录

**Acceptance（验证标准）:**
- AC: go build ./... / go vet ./... 零错误零警告
- AC: go test ./common/auth/... 全绿
- AC: 前端构建通过
- AC: 实测链路通过（或记录环境阻塞原因）

## 执行记录

- 既有失败说明：`common/auth/handler/TestDeleteUser_SoftDelete` 在本次改动前即失败（已通过 git stash 验证：无本 CR 改动时同样失败，报"无法删除系统中最后一个用户"，属既有问题，与本 CR 无关，RG-6 佐证）。
- 2026-10-01 设计修正：黑名单删除改硬删（Unscoped）——GORM 软删行仍占据 uk_ip 唯一索引，会导致过期后重新封禁报 UNIQUE 冲突（测试 TestBan_RebanAfterExpiry 实测暴露）；已同步 design.md 正确性属性。
- 2026-10-01 实测修复：BlacklistListItem.ID 原为 string+`,string` tag 导致 JSON 双重引号，已改 int64+`,string`（与项目雪花 ID 序列化惯例一致），8081 实例复测通过。
- 2026-10-01 全链路实测（8081 临时实例，MySQL+Redis 本机实连）：错误密码连续 9 次返回 40101 原样 → 第 10 次触发封禁 → 后续登录（含正确密码）40305 提示剩余约 1440 分钟 → DB 记录字段全对（fail_count=10、时长恰 24h）→ 解封（DB+Redis 双清）后立即恢复登录 200 → 黑名单 List（含 IP 模糊搜索/active/remaining_min 字段）与 Unban 接口实测通过。测试数据已清理，临时实例已停。
- **2026-10-01 需求变更（用户确认）：封禁维度由「整 IP」改为两级。** 一级 = IP+用户名 组合封禁（只拦该组合）；二级 = 同 IP 生效组合封禁数达到 `auth.login_fail_ip_ban_threshold`（默认 5，0=关闭）→ 升级封整 IP，24h 后自动解封。表结构：admin_login_blacklist 加 ban_type（combo/ip）+ (ban_type,ip,username) 目标唯一语义，唯一索引 uk_ip 改 uk_target(ban_type,ip,username)（AutoMigrate 自动变更）；缓存 key 拆两级 loginbl:ip:{hash} / loginbl:combo:{ip_hash}:{user_hash}；新增配置 seed auth.login_fail_ip_ban_threshold。前端列表加「封禁类型」列（IP 整体/IP+用户名 tag）。
- 2026-10-01 两级封禁全链路复测（8081 实例）：user_a 连错 10 次 → 仅 user_a 被拦（40305），同 IP user_b 放行（40101）、user_a 正确密码也拦 → 再封 user_c/d/e/f 共 5 组 → 升级整 IP → 未被封过的 user_x 与正确密码 admin 全部 40305 → DB 7 条记录（6 combo + 1 ip）结构正确。测试数据已清理，实例已停。
