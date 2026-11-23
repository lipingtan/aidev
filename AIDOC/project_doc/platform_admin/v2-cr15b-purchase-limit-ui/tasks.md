# CR15b 购买页限购体现（售罄 + 单人限购）· 2026-11-23 完成

用户拍板：两种"买不了"状态都要在 mjarch 购买页**前置体现**（此前只有点支付后 Toast 服务端文案）。

## 职责划分（用户问"是 mjarch 端改还是服务端也要改"）
- 售罄：只改 App（服务端 CR15 已下发 sold_out，App 此前没读）
- 单人限购：两端都要——公开套餐接口需按登录态透出 my_purchased/limit_reached

## 服务端（pa feature/cr15b-limit-display @6211949）
- 新增 `OptionalAuthMiddleware`（common/auth/middleware）：无/坏/过期 token 一律按匿名放行绝不 401，有效 token 注入 AuthContext；只挂 license 公开组。顺带修复：公开组"可选 JWT"此前无中间件、PublicGames my_score 恒匿名（代码注释自证）——本次真正激活
- `ListPublicPrices(tenantID, userID)` 透出 `my_purchased`（*int64，与下单限购同口径 pending+paid+activated，退款/超时/失败自动释放天然一致）+ `limit_reached`；匿名 omitempty 不下发，老 App 无感
- 证据：`TestPublicPricesPerUserFields` PASS（匿名不透出/u601 达限/u602 失败单释放=0/不限购5单永不达限）；`./app/license/...` 全绿；auth/handler 唯一失败 TestDeleteUser_SoftDelete 经 stash 基线比对为预存在

## mjarch App（Retro feature/license-phase2b @24239ea6）
- `getPrice` 改 `buildAuthGetRequest("/public/license/price")`：已登录自动带 token，公开路径行为不变
- `PriceInfo` +5 字段（max_total/max_per_user 默认 -1、sold_out、my_purchased(Integer null=匿名)、limit_reached），老服务端缺省=不限不提示无回归
- 卡片新增 `tv_limit_hint`：售罄"限量 N 份，已抢光"(灰)；登录"每人限购 X 次，已购 Y 次[，已达限购]"(达限转灰)；未登录仅公共规则"每人限购 X 次"（点支付仍有服务端兜底文案）
- 按钮：售罄→"已抢光"置灰 / 达限→"已达限购"置灰（自定义背景手动换色 #07C160→#C8C9CC）
- `onResume` 登录态静默刷新套餐（无 loading 闪烁、失败不清列表）——登录/下单/支付回来即反映最新名额
- 证据：assembleMjarchDebug exit=0（仅存量 deprecation 警告），apk mjarchdebug1.4.25.apk 09:59

## 真机四状态联调（2026-11-23 10:42-10:47，6b67cb2 部署后，体验卡设 总限购2/每人1、价格临时 ¥0.01）
1. ✅ 限量2仅剩2："限量 2 份，仅剩 2 份，每人限购 1 次，已购 0 次"（cr15b_s1_remain2.png；年卡/半年卡不限=无剩余行 ✓）
2. ✅ 待支付占名额：banner ORD202611231042542289 + "仅剩 1 份，已购 1 次，已达限购"(灰) + 按钮"已达限购"禁用；onResume 静默刷新生效（cr15b_s2_remain1.png）
3. ✅ 支付成功（用户手动微信支付 ¥0.01）：线上 total_remaining=1 与真机一致；顺带链路证据——普通会员29天(到期2026-12-23)+激活码 DXGZ-RV68-ZNSZ-9VYG（cr15b_s3_paid.png）
4. ✅ 售罄（管理端改总限购1）："限量 1 份，已抢光"(灰) + 按钮"已抢光"禁用，sold_out=true 线上同步（cr15b_s4_soldout.png）
- 对照：ORD202611231022394158（年卡未支付）经支付页"放弃支付"→ CloseOrder 名额释放，已购回 0 ✓
- 收尾：体验卡总限购改回 -1（价格 ¥0.01 为用户测试所改，是否回 9.99 由用户定）

## 补充（用户追加）：总名额剩余显示
- 规则：`max_total_purchases>0` 时卡片显示"限量 N 份，仅剩 M 份"（与单人限购提示逗号拼接）；**-1 不限=服务端不下发 `total_remaining`，App 自然不显示**；售罄仍显"限量 N 份，已抢光"(灰+按钮禁用)
- 服务端 pa `6b67cb2`：`total_remaining` = 上限-占位(pending+paid+activated)，负数钳0，omitempty；TestSoldCountSoldOutAndRecover 扩展断言（满=0/10占4=6/不限=nil）PASS
- App `263c5b6`：PriceInfo.totalRemaining(Integer null=不限) + payment_total_remain_format + adapter 拼装；BUILD OK + 装机 Success
- 真机证据（2026-11-23 10:20/10:27，cr15b_purchase_list.png / cr15b_list_v2.png）：登录态"每人限购 1 次，已购 0 次"透出（证明线上已部署 6211949）；年卡下单 ¥0.01(ORD202611231022394158) 未支付退出后 已购回到 0（名额释放正确）；三个不限套餐均无剩余行 ✓

## 待办
1. 体验卡总限购复原 -1（用户操作）
2. 顺带回归：CS 页与游戏列表（公开组新增中间件不应影响匿名接口）、支付/退款回调
