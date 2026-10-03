# 设计：登录失败 IP 黑名单（v2-cr11）

## 技术方案

登录链路（common/auth）内新增 IP 黑名单防护层：失败计数（内存）→ 达标写 DB 黑名单 + Redis 缓存 → 登录前置校验（Redis → 本地缓存 → DB 惰性过期）→ 管理页 CRUD。全部代码落在 common/auth 模块内，不新增顶层模块，与 LoginLog 同构。

### API 设计

| 方法 | 路径 | 描述 | 认证 | 权限码 |
|------|------|------|------|--------|
| GET | /api/v1/admin/login-blacklist | 黑名单分页列表（ip 模糊搜索） | JWT + 动态权限 | log:blacklist:list |
| DELETE | /api/v1/admin/login-blacklist/:id | 手动解封（删记录+清缓存） | JWT + 动态权限 | log:blacklist:delete |

登录接口（POST /auth/login）无新路由，行为增强：命中黑名单返回 HTTP 403，body `{code: 40305, message: "该 IP 已因登录失败次数过多被封禁，剩余约 N 分钟，请稍后重试或联系管理员"}`。新增错误码 `ErrIPBlacklisted = 40305`。

### 数据库设计

新表 admin_login_blacklist（GORM AutoMigrate，随 common/auth 现有 autoMigrate 机制，无需独立 migration 文件）：

```sql
CREATE TABLE admin_login_blacklist (
    id            BIGINT PRIMARY KEY,          -- 雪花 ID（model.NextID）
    ip            VARCHAR(64) NOT NULL,
    ip_hash       VARCHAR(64) NOT NULL,        -- sha256(ip) 小写hex，Redis key 隔离原文
    username      VARCHAR(128) NOT NULL,       -- 触发封禁的用户名
    fail_count    INT NOT NULL,                -- 触发时的失败计数
    banned_at     DATETIME NOT NULL,
    expires_at    DATETIME NOT NULL,
    created_at    DATETIME, updated_at DATETIME, deleted_at DATETIME,
    UNIQUE INDEX uk_ip (ip),                   -- 一 IP 一条有效记录
    INDEX idx_expires (expires_at)
);
```

系统配置 seed（幂等，insert-if-not-exists，AdminConfig 模型复用）：

| config_key | config_value | config_type | scope | scope_id | display_name |
|---|---|---|---|---|---|
| auth.login_fail_ban_threshold | 10 | number | SYSTEM | 0 | 登录失败封禁阈值（次，0=关闭） |
| auth.login_fail_ban_hours | 24 | number | SYSTEM | 0 | 登录失败封禁时长（小时） |

### 核心逻辑

**失败计数器（内存，service 层私有结构）**
```
map[comboKey]string{failCount, windowStart}，comboKey = ip + "|" + username
- 加 mutex；窗口 30 分钟滚动：now - windowStart > 30min 时重置计数
- RecordFailure(ip, username) (count int)：+1 并返回
- Reset(ip, username)：登录成功清零
```
只存于 LoginAttemptGuard（新 service），重启清零（需求已确认可接受）。

**登录链路改造（auth_handler.go handlePasswordLogin）**
```
1. 参数/验证码校验（现状不变）
2. guard.CheckBanned(ip) → 命中：记登录日志(status=0, message="IP被封禁") + 40305 错误返回
3. authSvc.Login(...) 
4. err != nil 且为凭证错误类：
   count = guard.RecordFailure(ip, username)
   threshold = adminConfigSvc.ResolveInt(tenantID=0, "auth.login_fail_ban_threshold", 10)
   if count >= threshold && threshold > 0:
       hours = adminConfigSvc.ResolveInt(0, "auth.login_fail_ban_hours", 24)
       blacklistSvc.Ban(ip, username, count, hours)   // 内部幂等，失败仅记日志
   → 错误返回（现状响应不变）
5. err == nil: guard.Reset(ip, username)
```
黑名单校验放在密码校验之前（省 bcrypt 计算，防刷语义正确）。注入方式：AuthHandler 新增 SetLoginGuard(guard)，auth.go buildDependencies 里装配（adminConfigSvc 已在该作用域存在）。

**黑名单服务（service/login_blacklist_service.go）**
```
Ban(ip, username, failCount, hours):
    upsert admin_login_blacklist WHERE ip=?（已存在且未过期→不改动，幂等）
    写 Redis: SET loginbl:{ip_hash} = expires_at.Unix() EX 到期秒数
    失败：log + 返回 error（调用方降级）
CheckBanned(ip) (banned bool, remaining time.Duration):
    1. Redis GET loginbl:{ip_hash} → 命中且 > now → true
    2. DB 查询 WHERE ip=? AND expires_at > now → 命中→回填 Redis→true
    3. false
Unban(id): 软删记录 + Redis DEL（双重清理，确保立即生效）
List(params): 分页 + ip LIKE + 状态计算（生效中/已过期）
异步清理：CheckBanned 发现过期记录时 go 删除（保持列表整洁）
```
Redis key 用 ip_hash 而非原文：避免特殊字符 key 问题，黑名单值无需可读。

**Guard 结构**
```go
type LoginAttemptGuard struct {
    mu       sync.Mutex
    counters map[string]*failCounter
    blSvc    *LoginBlacklistService
}
```
CheckBanned/RecordFailure/Reset 方法 + 内置窗口逻辑。LoginBlacklistService 独立可测，guard 是组合封装。

### 菜单与权限 seed

SeedCR11Menus（幂等，模式同 SeedTokenManagementMenu）：日志管理(/log)下新增子菜单"登录黑名单"，Path=/log/login-blacklist，PermissionCode=log:blacklist:list，ModuleCode=log-mgmt。
API 权限：discovery/auto_discover.go 三张 map 补 login-blacklists 条目（groupDisplayNames/endpointDisplayNames/endpointPermissionCodes/moduleCodeMap）。

### 前端

- `src/api/log.ts` 增 LoginBlacklistItem 接口 + listLoginBlacklist/deleteLoginBlacklist
- `src/views/system/log/LoginBlacklistList.vue`：复刻 LoginLogList.vue 结构；列 = IP、触发用户名、失败次数、封禁时间、预计解封时间、剩余时间（前端按 expires_at-now 计算，已过期显示"—"）、状态 tag（生效中 danger/已过期 info）、操作列（解封按钮 + confirm）；搜索 = IP 模糊
- `src/router/static-routes.ts` 日志管理组新增路由（permission: 'log:blacklist:list'）

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 正常登录流程不受影响 | 黑名单未命中时 /auth/login 正常返回 token（既有 auth_handler_test 回归） |
| RG-2 | 密码错误的错误响应不变 | 失败未达阈值时错误码/文案与现状一致（40101） |
| RG-3 | 验证码/参数校验先于黑名单拦截 | 空用户名/验证码错误仍返回原错误（顺序不变，除黑名单检查插在验证码后） |
| RG-4 | 登录日志写入不回归 | 封禁拦截也写 admin_login_log（status=0），原有成功/失败日志不受影响 |
| RG-5 | 阈值=0 时功能完全关闭 | 计数、封禁、拦截均不发生，行为与现状一致 |
| RG-6 | TokenBlacklist / user_auth 链路零改动 | git diff 无相关文件 |

## 正确性属性

- 黑名单校验必须发生在密码 bcrypt 计算之前（性能 + 防信息泄露）
- Ban 必须幂等：同 IP 重复达标不刷新到期时间（防攻击者通过继续试错延长封禁探测窗口——保持首次判定）
- Unban 必须同时清理 DB 与 Redis，任一失败返回错误（不可只清一层）
- 黑名单记录删除用硬删（Unscoped）：软删行仍占据 uk_ip 唯一索引会阻断重新封禁；审计依赖 admin_login_log
- CheckBanned 任何一层故障都不得阻断登录主流程（Redis 挂→查 DB；DB 挂→放行+记日志）
- 过期记录不得拦截登录（唯一拦截条件：expires_at > now）
- 黑名单写入失败时登录请求本身必须正常返回（封禁是增强能力，不能变成故障点）
- 所有写黑名单操作记录触发用户名与计数（可审计）

## 正确性属性自演进引用

沿用项目级约束：雪花 ID 主键（NextID）、写操作走 Service 层、软删（DeletedAt）、分页响应 `{list,total,page,page_size}`。
