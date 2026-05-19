# 需求计划：game_server 前端页面补全

## 背景

backend game 模块已完整实现六类 API（游戏、DLC、玩家、订单、支付配置、H5页面），
frontend `views/game/` 下只有 `index.vue`（游戏管理），其余五个模块页面缺失。

## 澄清问题

[Question] Q1: 玩家管理页面是否需要"查看详情"弹窗（展示 platformId、deviceId 等完整字段），还是列表展示已足够？
[Answer]
需要
[Question] Q2: 订单管理是否需要支持导出功能（CSV/Excel），还是仅查看和退款？
[Answer]
需要导出
[Question] Q3: 支付配置的 channel 字段目前有 stripe/alipay/wechat 三种，是否还有其他渠道需要支持？
[Answer]
先就这几种
[Question] Q4: H5页面的 pageType 字段有哪些枚举值（目前 model 只有 default:'custom'）？
[Answer]
"商城页"、"活动页"、"公告页
[Question] Q5: 菜单注册是通过管理界面手动操作，还是需要提供初始化 SQL 脚本？
[Answer]
提供初始化sql
## 非功能性需求（待确认）

- 性能：列表页默认分页 10 条，是否有特殊要求？ ok
- 权限：各页面是否需要细粒度按钮权限控制（如"封禁"按钮单独权限）？ 不要
- 多租户：前端页面是否需要显示租户信息，还是完全透明？ 显示
