# CR16 会员综合管理 — 任务清单

分支：`platform_admin` 仓库 `feature/cr16-member360`（基线 = CR15 的 869f594）。
合 main 待线上验证通过后执行。

| ID | 任务 | 验收标准 | 状态 |
|----|------|----------|------|
| T1 | 搜索端点 search | 订单号精确+模糊三路 LIKE；候选带会员摘要/订单数；租户隔离 | ✅ 代码+构建证据 |
| T2 | 详情端点 detail | 8 区块聚合一次返回；user_id string；租户隔离 | ✅ 代码+构建证据 |
| T3 | G1 grant-membership | 顺延/叠加/过期今起算；1~3650 边界；MANUAL_GRANT 留痕；缓存失效 | ✅ TestGrantMembership 通过 |
| T4 | G2 unbind-device | 不耗自助额度；bind_count 递减；备注留痕；重复/他人/跨租户报错 | ✅ TestAdminUnbindDevice 通过 |
| T5 | Member360.vue | 搜索/候选/详情/9 种快捷操作/流水抽屉/客服深链；字典下拉无手填 ID | ✅ pnpm build 通过，待线上点验 |
| T6 | 菜单+权限注册 | 会员管理后一位；license:member360(:manage) | ✅ 代码，⚠️ 部署后验证超管 403 |
| T7 | 客服工作台深链 | ?session=<id> 自动选中会话 | ✅ 代码，待线上点验 |
| T8 | （顺带）测试库 flaky 根治 | count=3 全绿 | ✅ f300f00 三轮 3×3 包全绿 |

## 联调验证清单（部署 feature/cr16-member360，target=all）

1. 菜单出现「会员综合管理」，打开无 403（权限码自动发放验证）
2. 订单号搜索 → 直达详情；邮箱片段搜索 → 候选列表点选
3. 赠送时长：选等级+天数 → 到期顺延、激活记录页出现 MANUAL_GRANT
4. 解绑：备注必填；解后绑定数-1、用户自助解绑次数不变
5. 禁言：无会话会员直接禁言成功（自动 init）；解禁；工作台深链自动选中
6. 取消待支付订单 / 退款 → 订单列表与本页状态同步；退款记录卡出现
7. 作废激活码 / 生成赠码（复制）；禁用/启用（双原因）+ 强制登出
8. 本分支同时携带 CR15（限购+成交/占用）与 7e91fa2（客服 mode 字段）——一并回归

## 遗留/风险

- 超管若不自动获得新权限码 → 页面 403，需在角色管理手动勾选（上线首日确认）
- 详情各区块 LIMIT 20/10/5，超大订单量会员要回模块页翻（本期不分页）
- 解绑审计目前仅服务端日志，若要 UI 追溯需建操作日志表（下期）

## TODO（用户 2026-11-22 记录，未排期；实现前先立需求+设计，届时建议单开 CR17）

### T9 web user 端：微信 Chat Bot 扫码登录 + 客服消息接入微信

**需求**：C 端（dev-web-user）增加「微信扫码登录」入口——经微信 Chat Bot
（go-rag-agent iLink bot，网关 :9201）扫码确认即登录；并且智能客服（bot）
回复与人工客服回复同步推送到该微信，用户在微信里就能收/发客服消息，
不必开着 web。

**方案粗案（只记录，未设计评审）**：
1. 扫码登录：web 向网关申请 bot 登录二维码 → 前端展示+轮询确认 →
   网关回微信身份（wechat_user_id）→ pa 后端查/建微信身份与 biz_user 的
   绑定（倾向新表 `biz_user_wechat_binds`，一码一用户，支持解绑重绑）→
   复用现有登录成功链路签发 go-admin token。
2. 消息互通：CsSession/CsMessage 模型不动，新增投递渠道 = 「该 user 已绑微信」：
   - bot/人工回复 → 除 WS 外经 bot 主动发微信消息（人工回复以 bot 账号代发，
     文案标注[客服]）；
   - 用户在微信直接发消息 → bot 收取 → 网关转 pa 写入 cs_messages
     （`source=wechat`，与 miniapp 同构）→ 自动定位/初始化其客服会话；
   - bot 模式照常走 ai_agent 链，命中转人工规则后微信侧提示"已转人工"。
3. 边界：微信身份换绑/一号多设备、bot 收发频率限制（微信侧节流，摘要式
   推送）、用户未登录 web 时消息经微信兜底、已登录时 WS 优先避免双推。

**依赖**：go-rag-agent 微信登录态（agent 登录页已有）+ iLink 收发消息能力；
pa 侧复用 CR14 客服会话/转人工/禁言全套。
**状态**：TODO——按质量门规则 3，写码前需先出完整设计（数据模型/字典/交互）给用户确认。

## CR16.5 退款增强（2026-11-22 追加，已实现 6983c74）

**用户拍板**：已激活订单可退（无时间窗）；金额可自定义（≤剩余）；可退多次直到
原单金额退完；退满才作废激活码(status3)+回收权益（1a：按整单天数扣回，扣穿钳到
当前）；部分退款视为补偿不回收权益。
**追加硬要求**：所有操作发起的和接收的报文都要有明细流水可查、能互相关联，供审计对账。

| ID | 任务 | 验收 | 状态 |
|----|------|------|------|
| T10 | 多次部分退款 + 已激活可退 | 状态门禁{1,2,3}；额度=原单−(SUCCESS+PROCESSING 占用)；退满才置4并处置码/权益 | ✅ refund_test.go 5 组用例全绿 |
| T11 | 退款全链路报文审计 | refund_records 三段报文（发起/响应原文/回调原始+解密）；payment_transactions.out_refund_no 回标；payment_callbacks 补记 REFUND 回调（验签失败也入账）；关联链 订单号↔商户退款单号↔微信退款单号 | ✅ 代码+单测断言 |
| T12 | Member360 退款 UI | 金额输入(默认/max=剩余,0.01 下限)+已激活提示；「报文明细」弹窗 | ✅ pnpm build，待线上点验 |

**审计关联设计**：`orders.order_no` ↔ `refund_records.refund_no`（=发给微信的
out_refund_no，同时回标到 `payment_transactions.out_refund_no` 列）↔
`refund_records.refund_id`（微信退款单号）；回调原文入 `payment_callbacks`
(callback_type=REFUND, headers+验签+处理结果)，解密明文并入 `refund_records.callback_body`。
**顺带修复**：适配器原先自造退款单号（与记录表对不上）；回调按订单号全量覆盖记录
（多次退款会互相踩）；ABNORMAL 渠道响应误记 PROCESSING；SUCCESS 记录防迟到 CLOSED 降级。
**遗留**：PROCESSING 后渠道回调 ABNORMAL 时款未退但订单未置终态（金额口径按记录走，
无需人工回滚权益——只有 SUCCESS 累计退满才回收）；退款按钮入口目前仅 Member360 页。
