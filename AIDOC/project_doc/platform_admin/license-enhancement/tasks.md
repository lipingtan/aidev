# 任务列表：激活码增强功能

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 20 |
| 已完成 | 0 |
| 进行中 | 0 |
| 未开始 | 20 |
| 完成率 | 0/20 (0%) |
| 当前阶段 | Phase 1: 数据库变更 |
| 预估总工时 | 16 人天 |

---

## Phase 1: 数据库变更

### Task 1: 新增 activation_code_devices 表 ⬜

**复杂度**: 中 | **预估工时**: 0.5天
**依赖**: 无（可与 Task 2、Task 3 并行）

**Scope（边界）:**
- 涉及文件: backend/app/license/models/models.go
- 涉及模块: license
- 不触碰: 其他模块的 model 文件

**Constraints（约束）:**
- 表结构必须符合 design.md 中的 DDL 定义
- 必须包含唯一索引 `idx_code_device (activation_code_id, device_id)`
- GORM Model 的 tag 必须与 DDL 一致

**Acceptance（验证标准）:**
- AC: Model 定义正确，包含所有字段和 GORM tag
- AC: AutoMigrate 后表结构正确
- AC: go build ./... 零错误

### Task 2: 新增 orders 表 ⬜

**复杂度**: 中 | **预估工时**: 0.5天
**依赖**: 无（可与 Task 1、Task 3 并行）

**Scope（边界）:**
- 涉及文件: backend/app/license/models/models.go
- 涉及模块: license
- 不触碰: 其他模块的 model 文件

**Constraints（约束）:**
- 表结构必须符合 design.md 中的 DDL 定义（含 device_id、expire_at、refund 相关字段）
- order_no 必须有唯一索引
- status 枚举：0=pending, 1=paid, 2=activated, 3=failed, 4=refunded, 5=closed
- GORM Model 的 tag 必须与 DDL 一致

**Acceptance（验证标准）:**
- AC: Model 定义正确，包含所有字段和 GORM tag
- AC: AutoMigrate 后表结构正确
- AC: go build ./... 零错误

### Task 3: 修改 activation_codes 表 ⬜

**复杂度**: 中 | **预估工时**: 0.5天
**依赖**: 无（可与 Task 1、Task 2 并行）

**Scope（边界）:**
- 涉及文件: backend/app/license/models/models.go
- 涉及模块: license
- 不触碰: 其他模块的 model 文件

**Constraints（约束）:**
- 新增字段：max_devices, device_bind_count, max_unbind_count, unbind_count, reserved_order_id, max_activations, used_activations
- max_activations 默认值为 1（APP内购），淘宝订单为订单项数量合计
- used_activations 默认值为 0
- 所有新增字段必须有默认值
- GORM Model 的 tag 必须与 DDL 一致

**Acceptance（验证标准）:**
- AC: Model 定义正确，包含新增字段和 GORM tag
- AC: AutoMigrate 后表结构正确
- AC: go build ./... 零错误
- AC: 【回归】现有激活码数据不受影响（RG-1）

---

## Phase 2: 后端 API 开发

### Task 4: 修改激活接口（增加设备绑定逻辑） ⬜

**复杂度**: 高 | **预估工时**: 1.5天
**依赖**: Task 1, Task 3

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 激活流程必须使用数据库事务
- 激活码查询和状态更新必须使用 SELECT FOR UPDATE 行锁
- 错误码必须符合 design.md 中的定义（4001-4013）
- 支持 confirm_extend 参数（同设备延期确认）
- 检查激活次数：used_activations < max_activations

**Acceptance（验证标准）:**
- AC: 激活成功时返回 bound_devices 和 remaining_activations 字段
- AC: 激活码已绑定最大设备数时返回错误码 4001
- AC: 激活码已在此设备激活且激活次数用完时返回错误码 4002
- AC: 激活码已被取消时返回错误码 4003
- AC: 激活码已过期时返回错误码 4004
- AC: 激活次数已用完时返回错误码 4012
- AC: 同设备再次激活（激活次数未用完）且未确认时返回 4013（含延期信息）
- AC: 同设备再次激活且 confirm_extend=true 时消耗激活次数并延长会员
- AC: 并发激活同一激活码时只有一个成功
- AC: go build ./... 零错误
- AC: 【回归】现有激活流程不受影响（RG-2）

### Task 5: 新增设备解绑接口 ⬜

**复杂度**: 中 | **预估工时**: 1天
**依赖**: Task 4

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 解绑流程必须使用数据库事务
- 必须检查解绑次数（unbind_count < max_unbind_count）
- 解绑操作必须使用 SELECT FOR UPDATE 行锁
- 处理并发场景（同时解绑和激活）

**Acceptance（验证标准）:**
- AC: 解绑成功时返回成功响应
- AC: 解绑次数达到上限时返回错误码 4005
- AC: 并发解绑时保证数据一致性
- AC: go build ./... 零错误

### Task 6a: 定义支付抽象层接口（PaymentAdapter） ⬜

**复杂度**: 中 | **预估工时**: 0.5天
**依赖**: 无（可提前开发）

**Scope（边界）:**
- 涉及文件: backend/app/license/service/payment_service.go（新建）
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 定义 PaymentAdapter 接口（CreatePayment、HandleCallback、QueryOrder、Refund、HandleRefundCallback、GetMethod）
- 定义 PaymentService 结构体（路由到对应 Adapter）
- 定义相关的请求/响应结构体

**Acceptance（验证标准）:**
- AC: PaymentAdapter 接口定义正确，包含6个方法（含 HandleRefundCallback）
- AC: PaymentService 结构体定义正确
- AC: go build ./... 零错误

### Task 6b: 实现微信支付适配器（WeChatPayAdapter） ⬜

**复杂度**: 高 | **预估工时**: 2天
**依赖**: Task 6a

**Scope（边界）:**
- 涉及文件: backend/app/license/service/wechat_pay.go（新建）
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 实现微信支付 API v3 H5 下单
- 实现回调验签（RSA-SHA256）和解密（AES-256-GCM）
- 实现主动查单接口
- 实现退款接口
- 实现退款回调处理
- 参考 C:\data\developer\projects\wechatpay 中的已验证代码

**Acceptance（验证标准）:**
- AC: WeChatPayAdapter 正确实现 PaymentAdapter 接口（含 HandleRefundCallback）
- AC: H5 下单返回正确的 h5_url
- AC: 回调验签和解密正确
- AC: 主动查单返回正确状态
- AC: 退款接口调用正确
- AC: 退款回调处理正确（更新订单状态、释放激活码）
- AC: go build ./... 零错误

### Task 7: 新增购买订单创建接口 ⬜

**复杂度**: 高 | **预估工时**: 1天
**依赖**: Task 2, Task 6b

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 创建订单时设置 expire_at = now() + 30分钟
- 创建订单后预分配激活码（设置 reserved_order_id）
- 调用 PaymentService.CreatePayment 获取支付链接
- 订单号必须唯一
- 必须返回 pay_url 和 expire_at 给前端

**Acceptance（验证标准）:**
- AC: 订单创建成功时返回 order_no、pay_url、expire_at
- AC: 订单包含 device_id 字段
- AC: 激活码正确预分配（reserved_order_id）
- AC: 支付API调用失败时订单状态保持 pending
- AC: go build ./... 零错误

### Task 8: 新增支付回调接口 ⬜

**复杂度**: 高 | **预估工时**: 1天
**依赖**: Task 7

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 必须验证支付签名（微信：RSA-SHA256）
- 必须使用订单号作为幂等键
- 检查订单是否已关闭（status=5），如已关闭则拒绝处理
- 重复回调时直接返回成功
- 支付成功后清除激活码的 reserved_order_id 并激活

**Acceptance（验证标准）:**
- AC: 支付成功后更新订单状态为 paid，记录 payment_order_id 和 paid_at
- AC: 支付成功后激活预分配的激活码
- AC: 已关闭订单收到回调时拒绝处理并记录日志
- AC: 重复回调不重复处理（幂等）
- AC: 签名验证失败时返回错误码 4008
- AC: go build ./... 零错误

### Task 9: 新增订单状态查询接口（含主动查单） ⬜

**复杂度**: 中 | **预估工时**: 1天
**依赖**: Task 8

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 支持单个订单查询（by order_no）
- 支持订单列表查询（带分页）
- 订单状态为 pending 且未过期时，主动调用微信查单 API（兜底）
- 查单节流：同一订单60秒内只查一次

**Acceptance（验证标准）:**
- AC: 单个订单查询返回正确数据
- AC: 订单列表查询支持分页
- AC: pending 状态订单触发主动查单
- AC: 查单节流正常工作
- AC: go build ./... 零错误

### Task 10: 新增价格配置查询接口 ⬜

**复杂度**: 低 | **预估工时**: 0.5天
**依赖**: 无（可提前开发）

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 返回所有激活码类型的价格配置（从 activation_code_types 表查询）

**Acceptance（验证标准）:**
- AC: 返回正确的价格配置列表
- AC: go build ./... 零错误

### Task 11: 新增淘宝订单推送接口 ⬜

**复杂度**: 高 | **预估工时**: 1天
**依赖**: Task 4

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 使用现有认证体系（access_token）
- 支持批量推送
- 淘宝订单状态映射：WAIT_SELLER_SEND_GOODS/TRADE_FINISHED → 创建激活码；TRADE_CLOSED/TRADE_CLOSED_BY_TAOBAO → 取消激活码
- 取消激活码时：状态改为 canceled，重新计算用户会员到期时间（replay 逻辑）
- 如果用户会员时长变为0或负数，会员立即失效

**Acceptance（验证标准）:**
- AC: WAIT_SELLER_SEND_GOODS 状态正确创建激活码
- AC: TRADE_CLOSED 状态正确取消激活码
- AC: 取消已激活的激活码时正确减去对应时长
- AC: 用户总时长变为0时会员立即失效
- AC: 批量推送多个订单时正确处理
- AC: 重复推送同一订单时不重复创建激活码
- AC: go build ./... 零错误

### Task 12: 新增订单超时定时任务 ⬜

**复杂度**: 中 | **预估工时**: 0.5天
**依赖**: Task 8
**并行**: 可与 Task 13 并行

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块

**Constraints（约束）:**
- 每分钟执行一次
- 查询 status=0 且 expire_at < now() 的订单
- 使用 FOR UPDATE SKIP LOCKED 避免阻塞
- 关闭订单并释放预分配的激活码

**Acceptance（验证标准）:**
- AC: 超时订单状态正确更新为 closed（status=5）
- AC: 预分配激活码正确释放（reserved_order_id = NULL）
- AC: 记录 closed_at 时间
- AC: go build ./... 零错误

### Task 13: 新增退款接口和自动退款逻辑 ⬜

**复杂度**: 中 | **预估工时**: 1天
**依赖**: Task 8
**并行**: 可与 Task 12 并行

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块

**Constraints（约束）:**
- 管理员手动退款接口
- 激活重试超过3次自动退款
- 调用微信退款 API
- 记录 refund_order_id、refund_reason、refunded_at

**Acceptance（验证标准）:**
- AC: 管理员可通过接口发起退款
- AC: 激活重试3次失败后自动触发退款
- AC: 订单状态正确更新为 refunded（status=4）
- AC: 退款失败时记录日志并通知管理员
- AC: go build ./... 零错误

### Task 13b: 新增退款回调接口 ⬜

**复杂度**: 中 | **预估工时**: 0.5天
**依赖**: Task 13

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块

**Constraints（约束）:**
- 接收微信退款结果通知
- 验签和解密（与支付回调相同）
- 根据退款状态更新订单
- 释放关联的激活码

**Acceptance（验证标准）:**
- AC: 退款成功（SUCCESS）时订单状态更新为 refunded
- AC: 退款异常（ABNORMAL）时记录日志并通知管理员
- AC: 重复回调幂等处理
- AC: 关联激活码正确释放（状态改为 unused）
- AC: go build ./... 零错误

---

## Phase 3: 前端开发

### Task 14: 前端激活码管理页面（设备绑定信息展示） ⬜

**复杂度**: 中 | **预估工时**: 1天
**依赖**: Task 4, Task 5

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/license/Codes.vue, dev-web-admin/src/api/license.ts（或对应 API 模块）
- 涉及模块: license
- 不触碰: 其他模块的 Vue 文件

**Constraints（约束）:**
- 必须使用 @/utils/request（axios wrapper）
- 必须展示设备绑定信息（设备ID、设备名称、绑定时间）
- 必须添加设备解绑 API 调用

**Acceptance（验证标准）:**
- AC: 激活码列表正确展示
- AC: 设备绑定信息正确展示（设备ID、设备名称、绑定时间）
- AC: 解绑操作正常工作
- AC: 【回归】现有激活码列表功能不受影响（RG-3）

### Task 15: 前端订单管理页面 ⬜

**复杂度**: 中 | **预估工时**: 1天
**依赖**: Task 9, Task 13

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/license/Orders.vue（新建）
- 涉及模块: license
- 不触碰: 其他模块的 Vue 文件

**Constraints（约束）:**
- 必须使用 @/utils/request（axios wrapper）
- 必须支持分页
- 必须支持管理员发起退款操作

**Acceptance（验证标准）:**
- AC: 订单列表正确展示（含支付状态、退款状态）
- AC: 分页正常工作
- AC: 订单状态正确展示（pending/paid/activated/failed/refunded/closed）
- AC: 管理员可发起退款操作

---

## Phase 4: 测试

### Task 16: 支付流程集成测试 ⬜

**复杂度**: 高 | **预估工时**: 1天
**依赖**: Task 8, Task 9, Task 12, Task 13, Task 13b

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service_test.go（新建或修改）
- 涉及模块: license
- 不触碰: 其他模块的测试文件

**测试场景：**
1. 正常支付流程：创建订单 → 支付回调 → 激活成功
2. 支付超时：创建订单 → 30分钟后订单自动关闭
3. 支付后订单已关闭：回调时订单已超时关闭，拒绝处理
4. 重复回调：同一订单收到多次回调，只处理一次
5. 激活失败自动退款：Mock 数据库错误模拟激活失败，重试3次后自动退款
6. 主动查单兜底：回调丢失时通过轮询触发补单
7. 查单节流：同一订单60秒内多次轮询，只触发一次查单
8. 退款回调成功：退款成功后订单状态更新，激活码释放
9. 退款回调异常：退款异常时记录日志

**Acceptance（验证标准）:**
- AC: 所有测试场景通过
- AC: go test ./... 零失败

### Task 17: 激活码并发测试 ⬜

**复杂度**: 中 | **预估工时**: 0.5天
**依赖**: Task 4, Task 5

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service_test.go
- 涉及模块: license

**测试场景：**
1. 1000并发激活同一激活码：只有1个成功
2. 同一激活码同时激活和解绑：串行处理，数据一致
3. 同一订单同时收到回调和轮询查单：幂等处理，只激活一次

**Acceptance（验证标准）:**
- AC: 并发场景数据一致性正确
- AC: go test -race ./... 零竞态错误

### Task 18: 微信支付沙箱联调 ⬜

**复杂度**: 中 | **预估工时**: 0.5天
**依赖**: Task 6b

**Scope（边界）:**
- 涉及环境: 微信支付沙箱环境
- 涉及模块: license

**测试内容：**
1. H5 下单接口调用
2. 支付回调验签和解密
3. 主动查单接口
4. 退款接口
5. 退款回调验签和解密

**Acceptance（验证标准）:**
- AC: 沙箱环境下单成功
- AC: 回调处理正确
- AC: 查单和退款接口正常
- AC: 退款回调处理正确

### Task 19: E2E 测试（完整购买流程） ⬜

**复杂度**: 高 | **预估工时**: 0.5天
**依赖**: Task 14, Task 15, Task 18

**Scope（边界）:**
- 涉及模块: license 完整流程

**测试流程：**
1. 用户查看价格 → 创建订单 → 完成支付 → 激活成功 → 查看会员状态
2. 管理员查看订单列表 → 发起退款 → 确认退款成功（含退款回调）
3. Chrome 插件推送淘宝订单 → 创建激活码 → 用户激活

**Acceptance（验证标准）:**
- AC: 完整购买流程正常
- AC: 管理员退款流程正常（含退款回调验证）
- AC: 淘宝订单推送流程正常

---

## 回归测试清单

| ID | 测试项 | 关联任务 | 验证方法 |
|----|--------|---------|---------|
| RG-1 | 现有激活码数据不受影响 | Task 3 | 查询现有激活码，验证字段值正确 |
| RG-2 | 现有激活流程不受影响 | Task 4 | 使用旧版 app 测试激活 |
| RG-3 | 现有激活码列表功能不受影响 | Task 14 | 管理端激活码列表页面正常 |
| RG-4 | 现有会员状态查询不受影响 | Task 4 | 查询现有会员状态，验证正确 |
| RG-5 | 现有激活记录查询不受影响 | Task 4 | 查询现有激活记录，验证正确 |

---

## 并行执行建议

```
Phase 1（可完全并行，1天内完成）:
  Task 1 ──┐
  Task 2 ──┼── 并行执行
  Task 3 ──┘

Phase 2 主线:
  Task 6a ─→ Task 6b ─→ Task 7 ─→ Task 8 ─→ Task 9
                                      │
                                      ├─→ Task 12（并行）
                                      └─→ Task 13 ─→ Task 13b

Phase 2 支线（可提前或并行）:
  Task 10（无依赖，可提前）
  Task 4 ─→ Task 5（依赖 Phase 1）
  Task 4 ─→ Task 11（依赖 Task 4）

Phase 3:
  Task 14（依赖 Task 4, 5）
  Task 15（依赖 Task 9, 13）

Phase 4:
  Task 17（可在 Task 5 完成后开始）
  Task 18（可在 Task 6b 完成后开始）
  Task 16（依赖后端主要功能完成）
  Task 19（最后执行）
```

---

## 工时汇总

| Phase | 任务数 | 预估工时 |
|-------|--------|---------|
| Phase 1: 数据库变更 | 3 | 1.5天（可并行压缩至0.5天） |
| Phase 2: 后端 API 开发 | 12 | 11天 |
| Phase 3: 前端开发 | 2 | 2天 |
| Phase 4: 测试 | 4 | 2.5天 |
| **合计** | **20** | **16天**（考虑并行可压缩至12天） |
