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

## 待办
1. CI 部署 pa `branch=feature/cr15b-limit-display`（或先并 main）
2. 手机接回（当前 adb devices 为空）→ adb install -r → 真机四场景截图：限量售罄置灰 / 登录限购已购 N 次 / 达限置灰"已达限购" / 未登录显示公共规则
3. 顺带回归：CS 页与游戏列表（公开组新增中间件不应影响匿名接口）、支付/退款回调
