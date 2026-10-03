# 需求：登录失败 IP 黑名单（连续失败 N 次封禁 24 小时）

## 背景

web admin 管理端登录接口暴露在公网时可被暴力试错。需要在该场景下增加防护：同一来源连续登录失败达到阈值后，将该 IP 加入登录黑名单，限制一段时间内不可登录。阈值与封禁时长均为系统配置项（后端已有 admin_config 系统配置能力，阈值默认 10 次）。

## 用户故事

- 作为平台管理员，我希望同一 IP+用户名 连续登录失败达到配置阈值后系统自动封禁该 IP，以便抵御暴力破解。
- 作为平台管理员，我希望在黑名单列表页看到被封 IP、封禁时间、预计解封时间，以便掌握封禁情况。
- 作为平台管理员，我希望能手动解封误封的 IP（如 NAT 出口），以便恢复正常用户访问。
- 作为运维，我希望阈值和封禁时长可在系统配置中调整，以便按环境调节防护强度。

## 功能需求

### FR-1: 登录失败计数
**描述：** 密码登录失败时，按 IP+用户名 组合维度累计失败次数（内存计数，滚动窗口）。登录成功清零该组合的计数。
**验收标准：**
- WHEN 同一 IP+用户名 登录失败 THEN 计数 +1
- WHEN 任一组合登录成功 THEN 该 IP+用户名 计数清零
- WHEN 计数窗口过期（30 分钟无失败）THEN 计数自动归零
- WHEN 服务重启 THEN 计数清零（可接受，黑名单不受影响）

### FR-2: 达标封禁（两级）
**描述：** 密码登录失败时按 IP+用户名 组合累计；某组合达到阈值（`auth.login_fail_ban_threshold`，默认 10，0=关闭）→ 封禁该 **IP+用户名 组合**；当同一 IP 下被封禁的用户名组合数达到二级阈值（`auth.login_fail_ip_ban_threshold`，默认 5，0=关闭）→ 升级封禁**整个 IP**。封禁时长取 `auth.login_fail_ban_hours`（默认 24 小时）。
**验收标准：**
- WHEN 某组合失败达到阈值 THEN 系统 SHALL 创建组合封禁记录（ban_type=combo），仅拦截该 IP+用户名
- WHEN 同一 IP 下生效中的组合封禁数达到二级阈值 THEN 系统 SHALL 创建整 IP 封禁记录（ban_type=ip），拦截该 IP 全部登录
- WHEN 任一阈值配置为 0 THEN 系统 SHALL 关闭对应层级
- WHEN 目标已有生效中的封禁 THEN 系统 SHALL 不重复创建（沿用原到期时间）
- WHEN 黑名单写入失败 THEN 系统 SHALL 记录错误日志但不阻断登录主流程（降级放行）

### FR-3: 黑名单拦截
**描述：** 登录入口在密码校验前按两级查黑名单：先查整 IP 封禁，再查 IP+用户名 组合封禁；命中且未过期直接拒绝。
**验收标准：**
- WHEN 组合封禁生效 THEN 该 IP 使用**被封用户名**登录 SHALL 被拒（40305）；同 IP 换**其他用户名** SHALL 正常走登录流程
- WHEN 整 IP 封禁生效 THEN 该 IP 任意用户名（含密码正确）SHALL 全部被拒（40305）
- WHEN 黑名单记录过期（expires_at < now）THEN 系统 SHALL 视为未命中，正常走登录流程
- WHEN 缓存不可用 THEN 系统 SHALL 回源 DB 查询（可用性优先）

### FR-4: 到期自动解封
**描述：** 黑名单校验为惰性过期判断（查询时比较 expires_at），到期后自动失效，无需定时任务。
**验收标准：**
- WHEN now > expires_at THEN 该 IP 登录不再被拦截
- WHEN 校验时发现过期记录 THEN 系统 SHALL 异步删除该记录（保持列表整洁）

### FR-5: 黑名单管理页（菜单）
**描述：** web admin 新增"登录黑名单"菜单页（日志管理分组下），展示被封 IP 列表并支持手动解封。
**验收标准：**
- WHEN 管理员打开页面 THEN SHALL 展示列表：IP、触发用户名、封禁类型（组合/IP 整体）、失败次数、封禁时间（banned_at）、预计解封时间（expires_at）、剩余时间、状态（生效中/已过期）、操作（解封）
- WHEN 管理员点击解封 THEN 系统 SHALL 删除黑名单记录并同步清除 Redis/缓存，该 IP 立即可登录
- WHEN 普通角色（无权限码）访问 THEN SHALL 无法看到该菜单且接口返回 403
- WHEN 支持 IP 模糊搜索与分页 THEN SHALL 与登录日志页交互一致

### FR-6: 配置项 seed
**描述：** 系统配置中新增两个平台级配置项（scope=SYSTEM, tenant_id=0），管理员可在现有系统配置页修改。
**验收标准：**
- WHEN 初始化（幂等 seed）THEN SHALL 存在 `auth.login_fail_ban_threshold`（默认 10，number 型）、`auth.login_fail_ip_ban_threshold`（默认 5，number 型，同 IP 组合封禁数达此值升级整 IP 封禁，0=关闭）和 `auth.login_fail_ban_hours`（默认 24，number 型）
- WHEN 修改阈值/时长 THEN 后续封禁行为 SHALL 使用新值（读取走 AdminConfigService 缓存链路）
- WHEN 配置不存在 THEN SHALL 使用内置默认值（阈值 10 / 时长 24），功能默认开启

## 非功能需求

- 性能：黑名单校验 Redis 命中路径额外耗时 < 2ms（P99）；未命中回源 DB 单表 ip 索引查询
- 安全：黑名单接口权限码 `log:blacklist:list`（查询）/ `log:blacklist:delete`（解封），走现有动态权限中间件；封禁提示不泄露计数进度
- 多租户：黑名单为平台级全局数据（IP 无租户归属），表不带 tenant_id；配置项为平台级（SYSTEM/0）
- 可用性：Redis 不可用时降级本地缓存，DB 不可用时登录主流程不阻断（封禁功能失效但不影响正常登录）
- 兼容性：配置为 0 或模块异常时，登录流程行为与现状完全一致

## 影响范围

- 后端：common/auth（model/repository/service/handler/router/seed/autoMigrate/discovery）
- 前端：dev-web-admin（api/log.ts、views/system/log/LoginBlacklistList.vue、router/static-routes.ts）
- 不触碰：app/user_auth 业务用户登录链路、TokenBlacklist（token 吊销）、现有登录成功/失败响应结构
