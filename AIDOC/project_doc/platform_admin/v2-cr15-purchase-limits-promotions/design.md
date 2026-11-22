# CR15 设计 — 套餐购买限购（一期，已实现）+ 限时优惠（二期）

## 0. 现状基线（设计时核实）

- `price_configs`：套餐定价（`price_code` 唯一编码、price 分、original_price 划线价、status 1上架/0下架、sort_order）；
- `orders` / `order_items`：下单时**快照**套餐价格与 `price_config_id`；订单状态 `0 pending / 1 paid / 2 activated / 3 failed / 4 refunded / 5 closed`，pending 30 分钟超时关闭；
- 下单入口 `LicenseService.CreateOrder`（license_service.go）：已有"用户 pending 防重下单、status 下架拦截、price_code 一致性校验、quantity 1~10"；
- C 端列表 `GameService.ListPublicPrices`，管理端 CRUD `ListPrices/CreatePrice/UpdatePrice`（game_price_service.go）。

---

## 一、一期：购买次数限购（已实现）

### 1.1 数据模型

`price_configs` 增两列（GORM AutoMigrate 自动加列，存量行 DEFAULT -1）：

| 列 | 类型 | 语义 |
|---|---|---|
| `max_total_purchases` | INT NOT NULL DEFAULT -1 | 该套餐总可售次数；**-1=不限** |
| `max_per_user_purchases` | INT NOT NULL DEFAULT -1 | 单用户（`user_id`）可购次数；**-1=不限** |

取值字典：`-1` 不限 / `≥1` 上限；`0` 与其他负数**录入即拒绝**（防手滑击杀；停售语义归"下架"开关）。

### 1.2 占位口径（派生计算，不存计数器）

```
已售 = SUM(order_items.quantity)
       JOIN orders ON order_no
       WHERE price_config_id = ? AND orders.tenant_id = ? AND orders.status IN (0,1,2)
```

- **计入**：pending(0)（下单即锁名额）、paid(1)、activated(2)；
- **释放**：failed(3)、refunded(4)、closed(5)（超时关闭/退款自动回补）；
- 标准 SQL，MySQL（生产）与 sqlite（单测）通用。

**"售罄"是派生态，不是 status**：`max_total_purchases > 0 && 已售 ≥ 上限`。
好处：①改大上限即刻恢复销售（满足"可修改重新上架"语义），零状态同步；②管理员显式下架（status=0）与售罄两开关互不纠缠。

### 1.3 服务端拦截（CreateOrder，quantity 确认后、创建订单前）

1. 总量：`remain = max - 已售`；`remain ≤ 0` → "该套餐限量，已被抢光"；`quantity > remain` → "该套餐限量，仅剩 X 份"（QuotaExceeded）；
2. 单人：`该用户已购 + quantity > max` → "该套餐每人限购 N 次，您已购 M 次"（Conflict）；
3. `-1` 天然短路不查库；付款回调不重复校验（下单已锁名额 + 微信金额校验）；
4. 并发：本业务量级下"下单瞬间重查"足够，不加行锁；极端并发靠 quantity≤10 与微信金额校验兜底。

### 1.4 接口增量（向后兼容）

- 管理端 `GET /prices` 行内增量：`sold_count`（占位口径已售=限购口径）、`pending_count`（待支付占用）、`paid_count`（已成交 paid+activated）；响应 items 由 `PriceConfig` 变为嵌入体 `PriceRow{PriceConfig, …}`，JSON 字段拍平，老前端无感；
- 管理端 `POST/PUT /prices` 入参增量：`max_total_purchases?`、`max_per_user_purchases?`（**指针语义：不传=不改**；服务端校验 -1/≥1）；
- C 端 `GET /public/license/price`（公开列表）每项增量：`max_total_purchases`、`max_per_user_purchases`、`sold_out`（老 App 忽略新字段照常工作；购买被拦时服务端中文文案经 App Toast 透传）。

### 1.5 管理界面（Prices.vue）

- 编辑/新增表单：`el-input-number(min=-1)` 两项，文案"-1 = 不限；售满自动售罄，改大即刻恢复"；空值提交归一为 -1，0 客户端先拒（服务端兜底）；
- 列表新增列"成交/占用/限购"：**成交**（绿，paid+activated）/ **占用**（橙，待支付下单，30 分钟超时自动释放）/ 限购上限（-1 显示"不限"）+ 达限红色 **售罄** 标签 + 第二行"每人 X"；
- **行内"改限购"弹窗**（决策③）：显示成交/占用两个数，只改两个上限，保存走 PUT /prices/:id（带整行必填字段）；
- 移动端卡片同步展示"成交/占用/限购 + 改限购"。

### 1.6 实现落点与证据

| 落点 | 文件 |
|---|---|
| 模型两列 | `backend/app/license/models/models.go` PriceConfig |
| 占位统计/校验/列表派生 | `backend/app/license/service/game_price_service.go`（`validatePurchaseLimit` / `priceSoldMap` / `PriceRow` / `PublicPriceItem.SoldOut`） |
| CreateOrder 拦截 | `backend/app/license/service/license_service.go`（`countPricePurchases` + 三段校验） |
| 管理端 UI | `dev-web-admin/src/views/license/Prices.vue`、`src/api/license.ts` |
| 单测 | `backend/app/license/service/price_limit_test.go`（4 用例：值校验 / 创建默认与更新 / 占位统计+售罄派生+改上限恢复 / CreateOrder 三段拦截真跑） |

运行证据（2026-11-22）：`go test ./app/license/service -run '…Limit…'` 4/4 PASS；全量 `go test ./app/license/...` 全 ok；`go vet` 0；`pnpm run build` exit=0。

---

## 二、二期：限时优惠（设计确认，待编码）

### 2.1 数据模型：新表 `price_promotions`（一期一活动一套餐）

| 列 | 类型 | 语义 |
|---|---|---|
| id / tenant_id / name | | 活动名（如"双11 年卡特惠"） |
| price_config_id | INT NOT NULL | 目标套餐（单选；多套餐以后扩关联表） |
| promo_price | INT NOT NULL | **限时一口价（分）**——不用折扣率，免取整口径纠纷；划线价=套餐现价 |
| start_at / end_at | DATETIME | 服务器时间窗，左闭右开 |
| status | INT DEFAULT 1 | 1=启用 0=停用（手动启停） |
| promo_stock | INT NOT NULL DEFAULT -1 | 促销名额（-1 不限）；口径同 1.2（占位含 pending） |
| per_user_limit | INT NOT NULL DEFAULT -1 | 该活动内每人限享次数 |
| condition | VARCHAR(20) DEFAULT 'all' | 字典：`all` / `new_user`（无历史 paid 单）——一期只做两值 |
| created_at / updated_at | | |

约束：同一 `price_config_id` **活动时间窗不得重叠**（保存时检测，重叠即拒绝；不引入优先级字段）。

### 2.2 价格判定链（服务端唯一真相，防改包）

1. **公开列表**：对每个套餐求"当前生效促销"= `status=1 && now∈[start,end) && 促销名额未满 && 用户符合 condition` → 每项增量返回 `promo_price`、`promo_end_at`、`promo_id`；App 显示划线原价 + 倒计时（**用服务器时间差**，不信本机时钟；列表响应已有服务器时间或以响应到达时间起算）；
2. **下单**：`CreateOrder` 当时重查生效促销（App 传 `promo_id` 仅做一致性比对：App 没传而服务端有生效促销 → 自动享；促销已结束 → 回正价）；订单金额=服务端当时计算价，快照写入 `order_items.price`（机制已有），`orders` 可选加 `promotion_id` 溯源列；
3. **支付**：促销期间下单后促销结束——订单按**锁定快照价**，30 分钟支付窗内仍按促销价（行业惯例）；微信回调金额与订单金额一致性校验已存在。

### 2.3 与限购的口径关系（已定）

- 促销成交**同时占用**套餐自身的 `max_total_purchases` / `max_per_user_purchases`（名额同一口径，防"正价买满 N 次 + 促销又 N 次"漏洞）；
- `promo_stock` / `per_user_limit` 是促销的**额外**独立名额（只数促销价成交）；
- 淘宝渠道（激活码兑换）不参与自动促销——商家侧手动定价。

### 2.4 管理界面（二期）

- 新页面"限时促销"（license 菜单，权限码 `license:promotions:*`，套路同 CRUD 菜单注册）：列表（活动名/套餐/促销价/窗口/状态/已享人数）+ 新建弹窗（套餐**下拉选择器**、价格分、`el-date-picker` 区间、条件下拉字典）；
- 保存时服务端时间窗重叠校验失败 → 明确文案"与活动 X（…~…）时间重叠"。

### 2.5 App 端（二期，mjarch）

- 套餐卡片：`promo_price` 存在时显示促销价 + 原价划线 + "剩 HH:MM:SS"倒计时（`promo_end_at - 服务器时间`）；
- 下单响应/错误文案已由服务端定死，App 不自行算价。

### 2.6 边界清单（编码时必测）

窗口边界（start 整分可买 / end 整分不可买）、促销与限购叠加（正价买满后促销期也不能买）、new_user 判定（仅 pending 未付不算老客）、促销名额并发抢占（同 1.3 口径）、App 传旧 promo_id（下单时已结束→按正价成交或明确文案）、倒计时归零前端表现（刷新列表回正价）。
