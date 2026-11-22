# CR15 任务清单

## 一期：购买次数限购（2026-11-22 完成，feature 分支待测试合入）

- [x] price_configs 增 `max_total_purchases` / `max_per_user_purchases`（-1 不限，DEFAULT -1）
- [x] 占位统计 `priceSoldMap` / `countPricePurchases`（pending+paid+activated 计入，refunded/closed/failed 释放）
- [x] CreateOrder 三段拦截（抢光 / 仅剩 X 份 / 每人限购 N 次已购 M 次）
- [x] 限购值字典校验（-1 或 ≥1；0 与其他负数拒绝）
- [x] 管理端列表 `sold_count` + 售罄标签 + 行内"改限购"弹窗
- [x] C 端公开列表增量字段（max_total_purchases / max_per_user_purchases / sold_out）
- [x] 单测 4 用例真跑全绿 + license 全包回归 ok + vet 0 + 管理端 build exit=0
- [ ] 用户 feature 分支实测（部署 dev/测试环境 → 设上限购买 → 售罄 → 改上限恢复）→ 合 main

## 二期：限时优惠（设计已确认，待编码）

- [ ] 新表 `price_promotions`（design.md §2.1，字段/字典见设计）
- [ ] 保存时同套餐时间窗重叠校验
- [ ] 公开列表注入生效促销（promo_price / promo_end_at / promo_id，服务器时间口径）
- [ ] CreateOrder 接促销价（服务端重查 + 快照 + 可选 orders.promotion_id）
- [ ] 促销名额独立计数（promo_stock / per_user_limit）与套餐限购叠加口径联测
- [ ] 管理端"限时促销"页面 + 权限码注册
- [ ] mjarch App 套餐卡片促销价/划线价/倒计时
- [ ] 边界清单测试（design.md §2.6）
