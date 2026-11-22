# CR16 会员综合管理 — 设计

## 1. 总览

一页式会员档案：搜索（R1）→ 360° 详情（R2）+ 就地快捷操作（R3/R4）。
**复用优先**原则：所有既有能力走原路由（禁用/登出/取消订单/退款/码作废/赠码/禁言），
只新增 4 个端点（搜索、详情、G1 赠送、G2 解绑）。

```
Member360.vue ── GET  /admin/license/member-360/search?q=        [新] 候选解析
              ── GET  /admin/license/member-360/detail?user_id=  [新] 聚合档案
              ── POST /admin/license/member-360/grant-membership [新G1] 时长赠送/续期
              ── POST /admin/license/member-360/unbind-device    [新G2] 管理端强制解绑
              ── POST /admin/license/biz-users/:id/status            复用：禁用/启用（双原因）
              ── POST /admin/license/users/:id/force-logout          复用：强制登出
              ── POST /admin/license/orders/:id/cancel               复用：取消待支付
              ── POST /admin/license/refund                          复用：退款（order_no+reason）
              ── GET  /admin/license/order-payment/:order_no         复用：流水/回调抽屉
              ── POST /admin/license/cancel                          复用：激活码作废（code）
              ── POST /admin/license/codes                           复用：生成赠码（type_id）
              ── GET  /admin/license/levels | /code-types            复用：下拉数据源
              ── POST /admin/license/cs-mgmt/session-init            复用：无会话先建会话
              ── POST /admin/license/cs-mgmt/sessions/:id/mute|unmute 复用：禁言/解禁
```

## 2. 搜索（Member360Search）

- 入参 `q`（trim 后空 → 空结果）。**同时**跑两路（订单号与邮箱形态可并存判别成本低，直接双查比形态猜测可靠）：
  - 订单路：`orders.order_no = q` 精确 + `LEFT JOIN biz_user` 带出归属会员昵称/邮箱（LIMIT 1）
  - 用户路：`email/nickname/phone LIKE %q%`，`tenant_id + deleted_at IS NULL`，`id DESC LIMIT 10`
- 候选行 enrichment（仅用户路，≤10 id 的批量查询）：
  - `user_membership` status=1 且未过期 → `membership_summary`（"VIP 至 2027-01-01；SVIP 至 …"）
  - `orders` GROUP BY → `order_count`
- 前端唯一命中自动进详情（1 订单 0 用户 / 0 订单 1 用户），否则点候选行进入。

## 3. 详情聚合（Member360Detail）

`user_id`（string，防 JS 精度）→ 8 区块，全部 `tenant_id` 隔离：

| 区块 | 来源 | 要点 |
|------|------|------|
| base | `biz_user` + `biz_user_status_logs` 最近一条 | `disabled = status!=1`；附最近禁用时间/admin_reason |
| memberships | `user_membership` 全部 | 前端算剩余天数/过期 tag |
| orders | `orders` 近 20 | 附 `order_items` map（order_no → items，套餐名×数量） |
| refunds | `refund_records JOIN orders ON order_no`（where user_id） | 状态字典 PROCESSING/SUCCESS/ABNORMAL/CLOSED |
| activation_codes | `activation_code WHERE used_by=user_id` 近 20 | 状态字典 0未用/1已用/2已恢复/3已取消；作废仅 1/2 可点 |
| devices + bindings | `user_device`；`activation_code_devices JOIN activation_code`（is_active） | 解绑按钮走 bindings |
| cs_sessions | `cs_sessions` 近 5 | 深链 `?session=<id>` 进客服工作台（CustomerService.vue onMounted 自动 selectSession） |
| mute | `cs_mutes` 当前有效行 | `mute_until NULL=永久`；有则显示 tag+解禁 |

## 4. G1 赠送/续期时长（MembershipService.GrantMembership）

- 入参：`user_id`（biz_user 存在性校验）、`level_code`（本租户等级）、`duration_days`（1~3650）。
- 复用激活链的 `upsertMembershipTx`：同等级已有记录 → **有效期内到期日叠加顺延**、过期 → 今日起算；无记录 → 新建。
- 同事务写 `ActivationRecord{Code:"MANUAL_GRANT", CodeID:0, Status:1}`——与真激活码记录同表，后台可审计区分。
- 提交后 `InvalidateMembershipCache`（App 端会员态即时生效）。
- 返回 `expires_at` 供前端 toast 回显。

## 5. G2 管理端强制解绑（LicenseService.AdminUnbindDevice）

与用户自助解绑（受 `MaxUnbindCount`/`UnbindCount` 约束）的区别：**管理员是裁决者**——
- 匹配 `activation_code_devices JOIN activation_code WHERE ac.used_by=? AND device_id=? AND is_active`，逐条 `is_active=false + unbind_time`；
- `device_bind_count` 按解绑条数 Go 侧递减（夹 0，不写 `GREATEST(...)` 保 sqlite 兼容）；
- **不动 `unbind_count`**（不消耗用户自助额度）；
- 备注必填在前端拦，后端 `log.Printf` 审计行留痕（不建新表）。

## 6. 禁言链路（复用 + 口径）

时长字典与客服工作台一致：10 分钟(0.167h)/1h/24h/7d(168h)/永久(0)。
禁言接口挂在会话路由上（校验会话归属），故本页禁言流程：
`detail.cs_sessions[0]?.id` → 无则 `POST cs-mgmt/session-init {user_id}`（自动初始化）→ `POST sessions/:sid/mute`。
解禁同路由 unmute。成功/失败后 `reloadDetail` 刷新 mute tag。

## 7. 权限与菜单（plugin.go 声明式注册）

| 项 | 值 |
|----|----|
| 菜单 | 「会员综合管理」`/license/member360`，icon Search，Sort 4（与会员管理同值紧随其后），PermissionCode `license:member360` |
| 查看 | `license:member360` → search/detail（GET） |
| 操作 | `license:member360:manage` → grant/unbind（POST）+ 复用接口沿用自己的原权限码 |
| 角色 | 超管自动继承全部码；自定义角色需手动勾选两个新码 |

⚠️ 部署验证项：新权限码若超管未自动获得会 403——上线后第一时间验证。

## 8. 前端（Member360.vue，~590 行）

- 布局：搜索卡 → （候选区 | 详情区）切换；详情 7 卡片纵向排。
- 字典全部前端常量（订单/退款/码状态、禁用原因、退款原因、禁言时长、赠送天数 7/30/90/180/365/自定义），杜绝自由文本 ID。
- `user_id` 全链路 string（`json:",string"` 对齐），订单状态等 tag 色与 Orders/Codes 页一致。
- 操作成功统一 toast + `reloadDetail()` 局部刷新。

## 9. 测试设计

| 用例 | 断言 |
|------|------|
| TestGrantMembership | 首次 ≈30 天；二次**顺延**（>旧到期+29d）；单行不重复；2 条 MANUAL_GRANT 记录 |
| 边界 | 0 天 / 3651 天 / 等级不存在 / 用户不存在 全部报错不落库 |
| TestAdminUnbindDevice | 只解指定设备；`device_bind_count` 2→1；`unbind_count` 不动；unbind_time 落；重复解绑报"无有效绑定"；他人设备/跨租户解不到 |

## 10. 顺带修复（独立 commit）

license 测试偶发 `no such table`（`:memory:` 按连接隔离，约 1/3 概率）：
7 个测试建库点统一改 `file:<label>_<纳秒>?mode=memory&cache=shared`（多连接共享、纳秒名隔离数据）。
曾评估 `SetMaxOpenConns(1)`：与事务内外混用查询死锁，弃。
