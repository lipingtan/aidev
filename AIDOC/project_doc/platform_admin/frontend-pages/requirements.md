# 需求：platform_admin 前端页面补全

## 背景

`projects/demo/platform_admin` 的 backend 已完整实现 game 模块六类 API（游戏、DLC、玩家、订单、支付配置、H5页面）。frontend `views/game/` 下目前只有 `index.vue`（游戏管理），其余五个模块页面缺失，需补全。

## 用户故事

- 作为运营人员，我希望管理 DLC 包的上下架和定价，以便控制内容发布节奏
- 作为运营人员，我希望查看玩家详情并执行封禁/解封，以便处理违规行为
- 作为财务人员，我希望查看订单列表、执行退款并导出数据，以便对账和处理纠纷
- 作为技术人员，我希望配置各游戏的支付渠道参数，以便接入不同支付方式
- 作为运营人员，我希望管理游戏内嵌的 H5 页面（商城/活动/公告），以便灵活更新内容

## 功能需求

### FR-1: DLC 管理页面

**描述：** 管理游戏 DLC 包的完整生命周期，包括创建、编辑、上下架。

**验收标准：**
- WHEN 用户访问 DLC 管理页 THEN 系统 SHALL 展示分页列表，含：ID、所属游戏ID、DLC名称、dlcKey、版本、价格（分→元显示）、是否免费标签、状态标签、下载次数
- WHEN 用户按游戏ID或名称搜索 THEN 系统 SHALL 过滤并刷新列表
- WHEN 用户点击新增/编辑 THEN 系统 SHALL 弹出表单，含字段：gameId、dlcKey、name、version、description、price（元，提交时×100存分）、isFree、status、minGameVersion
- WHEN 用户确认删除 THEN 系统 SHALL 调用删除接口并刷新列表
- WHEN price 字段提交 THEN 系统 SHALL 将元转换为分（×100）后发送

### FR-2: 玩家管理页面

**描述：** 查看玩家信息，支持详情查看和封禁/解封操作。

**验收标准：**
- WHEN 用户访问玩家管理页 THEN 系统 SHALL 展示分页列表，含：ID、所属游戏ID、UID、昵称、邮箱、平台、状态标签、最后登录时间、最后登录IP
- WHEN 用户按游戏ID、UID、昵称、状态搜索 THEN 系统 SHALL 过滤并刷新列表
- WHEN 用户点击"详情" THEN 系统 SHALL 弹出详情弹窗，额外展示：platformId、deviceId、banReason
- WHEN 用户对正常玩家点击"封禁" THEN 系统 SHALL 弹出输入框要求填写 banReason，确认后调用 banPlayer(id, {status:2, banReason})
- WHEN 用户对封禁玩家点击"解封" THEN 系统 SHALL 直接调用 banPlayer(id, {status:1, banReason:""})
- WHEN 玩家列表 THEN 系统 SHALL 不提供新增/删除按钮

### FR-3: 订单管理页面

**描述：** 查看支付订单，支持退款和数据导出。

**验收标准：**
- WHEN 用户访问订单管理页 THEN 系统 SHALL 展示分页列表，含：ID、订单号、游戏ID、玩家ID、商品名称、金额（分→元，保留2位小数+货币单位）、支付渠道、状态标签、支付时间
- WHEN 用户按游戏ID、订单号、状态（pending/paid/refunded/failed）搜索 THEN 系统 SHALL 过滤并刷新列表
- WHEN 订单状态为 pending THEN 系统 SHALL 显示"待支付"黄色标签；paid 显示"已支付"绿色；refunded 显示"已退款"灰色；failed 显示"失败"红色
- WHEN 用户点击"退款"并确认 THEN 系统 SHALL 调用 refundOrder 接口并刷新列表
- WHEN 用户点击"导出" THEN 系统 SHALL 将当前查询条件的订单数据导出为 CSV 文件，文件名含日期
- WHEN 订单列表 THEN 系统 SHALL 不提供新增/删除按钮

### FR-4: 支付配置页面

**描述：** 配置各游戏的支付渠道参数（Stripe / 支付宝 / 微信）。

**验收标准：**
- WHEN 用户访问支付配置页 THEN 系统 SHALL 展示分页列表，含：ID、游戏ID、渠道标识、渠道名称、环境、启用状态标签；敏感字段（私钥、密钥）不在列表展示
- WHEN 用户按游戏ID搜索 THEN 系统 SHALL 过滤并刷新列表
- WHEN 用户新增/编辑配置 THEN 系统 SHALL 弹出表单，channel 字段提供 stripe/alipay/wechat 三个选项
- WHEN channel 选择 stripe THEN 系统 SHALL 显示：stripePublishableKey、stripeSecretKey、stripeWebhookSecret
- WHEN channel 选择 alipay THEN 系统 SHALL 显示：alipayAppId、alipayPrivateKey、alipayPublicKey、alipayNotifyUrl
- WHEN channel 选择 wechat THEN 系统 SHALL 显示：wechatAppId、wechatMchId、wechatApiKey、wechatNotifyUrl
- WHEN channel 切换 THEN 系统 SHALL 隐藏其他渠道的字段组

### FR-5: H5 页面管理

**描述：** 管理游戏内嵌的 H5 页面，支持内嵌内容和外链两种模式。

**验收标准：**
- WHEN 用户访问 H5 页面管理 THEN 系统 SHALL 展示分页列表，含：ID、游戏ID、页面标识、名称、页面类型、是否外链标签、状态标签
- WHEN 用户按游戏ID或名称搜索 THEN 系统 SHALL 过滤并刷新列表
- WHEN 用户新增/编辑 THEN 系统 SHALL 弹出表单，含：gameId、pageKey、name、pageType（下拉：custom/商城页/活动页/公告页）、useExternal、status、remark
- WHEN useExternal 为"是（1）" THEN 系统 SHALL 显示 externalUrl 输入框，隐藏 content
- WHEN useExternal 为"否（2）" THEN 系统 SHALL 显示 content textarea（rows=8），隐藏 externalUrl
- WHEN 用户确认删除 THEN 系统 SHALL 调用删除接口并刷新列表

### FR-6: 菜单初始化 SQL

**描述：** 提供菜单注册的初始化 SQL 脚本，将五个新页面注册到系统菜单。

**验收标准：**
- WHEN 执行初始化 SQL THEN 系统 SHALL 在游戏管理父菜单下新增五个子菜单：DLC管理、玩家管理、订单管理、支付配置、H5页面
- WHEN 菜单注册完成 THEN 系统 SHALL 使对应路由（/game/dlc、/game/player、/game/order、/game/payment、/game/h5）在侧边栏可见

## 非功能需求

- 性能：列表页默认分页 10 条，接口响应时间 < 500ms
- 权限：页面级权限控制，不需要细粒度按钮权限
- 多租户：列表页显示 tenantId 字段（或所属租户名称），数据按租户隔离
- 编码规范：复用现有 `views/game/index.vue` 的代码风格（Element Plus + Composition API），不引入新依赖
- 金额规范：所有金额字段以"分"存储，页面显示时÷100保留2位小数
