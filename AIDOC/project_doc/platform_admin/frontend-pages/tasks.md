# Tasks: platform_admin 前端页面补全

## Task Dependency Graph

```
T1 → T2 → T3 → T4 → T5 → T6（串行，后续页面可参考前面的实现）
```

---

- [x] 1. 实现 DLC 管理页面
  - 创建 `projects/demo/platform_admin/platform_admin/frontend/src/views/game/dlc.vue`
  - 实现 DLC 列表、搜索（gameId/name）、新增、编辑、删除功能
  - 复用 `getDlcList / createDlc / updateDlc / deleteDlc`
  - 价格字段：表单输入"元"，提交时×100转"分"；列表显示时÷100保留2位小数
  - isFree（1=免费/2=付费）、status（1=上架/2=下架）用 el-tag 展示
  - 使用 `defineOptions({ name: "DlcManage" })`，风格与 `views/game/index.vue` 一致
  - _Requirements: FR-1_

- [x] 2. 实现玩家管理页面
  - 创建 `projects/demo/platform_admin/platform_admin/frontend/src/views/game/player.vue`
  - 实现玩家列表、搜索（gameId/uid/nickname/status）、详情弹窗、封禁/解封功能
  - 详情弹窗（只读）：展示全部字段含 platformId、deviceId、banReason
  - 封禁弹窗：输入 banReason，调用 `banPlayer(id, { status: 2, banReason })`
  - 解封：直接调用 `banPlayer(id, { status: 1, banReason: "" })`，无需弹窗
  - 无新增/删除按钮，使用 `defineOptions({ name: "PlayerManage" })`
  - _Requirements: FR-2_

- [x] 3. 实现订单管理页面
  - 创建 `projects/demo/platform_admin/platform_admin/frontend/src/views/game/order.vue`
  - 实现订单列表、搜索（gameId/orderNo/status）、退款、CSV导出功能
  - 金额显示：`(row.amount / 100).toFixed(2) + ' ' + row.currency`
  - status 颜色：pending→warning、paid→success、refunded→info、failed→danger
  - 退款：el-popconfirm 确认后调用 refundOrder
  - 导出 CSV：将当前搜索条件下的列表数据转为 CSV 并触发下载，文件名含日期
  - 无新增/删除按钮，使用 `defineOptions({ name: "OrderManage" })`
  - _Requirements: FR-3_

- [x] 4. 实现支付配置页面
  - 创建 `projects/demo/platform_admin/platform_admin/frontend/src/views/game/payment.vue`
  - 实现支付配置列表、搜索（gameId）、新增、编辑功能
  - 列表不展示敏感字段（*Key/*Secret/*PrivateKey/*PrivateKey）
  - 表单中 channel（stripe/alipay/wechat）切换时动态显示对应字段组
  - stripe 字段组：stripePublishableKey、stripeSecretKey、stripeWebhookSecret
  - alipay 字段组：alipayAppId、alipayPrivateKey、alipayPublicKey、alipayNotifyUrl
  - wechat 字段组：wechatAppId、wechatMchId、wechatApiKey、wechatNotifyUrl
  - enabled 用 el-tag 展示，使用 `defineOptions({ name: "PaymentManage" })`
  - _Requirements: FR-4_

- [x] 5. 实现 H5 页面管理
  - 创建 `projects/demo/platform_admin/platform_admin/frontend/src/views/game/h5.vue`
  - 实现 H5 页面列表、搜索（gameId/name）、新增、编辑、删除功能
  - pageType 下拉选项：custom / 商城页 / 活动页 / 公告页
  - useExternal=1 时显示 externalUrl 输入框，=2 时显示 content textarea（rows=8）
  - useExternal、status 用 el-tag 展示，使用 `defineOptions({ name: "H5PageManage" })`
  - _Requirements: FR-5_

- [x] 6. 生成菜单初始化 SQL
  - 创建 `projects/demo/platform_admin/backend/config/menu_init.sql`
  - 包含游戏管理父菜单（若不存在则插入）及五个子菜单的 INSERT 语句
  - 子菜单：DLC管理、玩家管理、订单管理、支付配置、H5页面
  - 路由路径：game/dlc、game/player、game/order、game/payment、game/h5
  - 使用子查询获取父菜单 ID，避免硬编码
  - _Requirements: FR-6_
