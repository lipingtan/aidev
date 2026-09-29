# 任务列表：激活码增强功能

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 13 |
| 已完成 | 0 |
| 进行中 | 0 |
| 未开始 | 13 |
| 完成率 | 0/13 (0%) |
| 当前阶段 | Phase 1: 数据库变更 |

---

## Phase 1: 数据库变更

### Task 1: 新增 activation_code_devices 表 ⬜

**复杂度**: 中

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

**复杂度**: 中
**依赖**: Task 1

**Scope（边界）:**
- 涉及文件: backend/app/license/models/models.go
- 涉及模块: license
- 不触碰: 其他模块的 model 文件

**Constraints（约束）:**
- 表结构必须符合 design.md 中的 DDL 定义
- order_no 必须有唯一索引
- GORM Model 的 tag 必须与 DDL 一致

**Acceptance（验证标准）:**
- AC: Model 定义正确，包含所有字段和 GORM tag
- AC: AutoMigrate 后表结构正确
- AC: go build ./... 零错误

### Task 3: 修改 activation_codes 表 ⬜

**复杂度**: 中
**依赖**: Task 2

**Scope（边界）:**
- 涉及文件: backend/app/license/models/models.go
- 涉及模块: license
- 不触碰: 其他模块的 model 文件

**Constraints（约束）:**
- 新增字段：max_devices, device_bind_count, max_unbind_count, unbind_count, version
- 所有新增字段必须有默认值
- version 字段用于乐观锁，默认值为 0
- GORM Model 的 tag 必须与 DDL 一致

**Acceptance（验证标准）:**
- AC: Model 定义正确，包含新增字段和 GORM tag
- AC: AutoMigrate 后表结构正确
- AC: go build ./... 零错误
- AC: 【回归】现有激活码数据不受影响（RG-1）

---

## Phase 2: 后端 API 开发

### Task 4: 修改激活接口（增加设备绑定逻辑） ⬜

**复杂度**: 高
**依赖**: Task 3

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 激活流程必须使用数据库事务
- 设备绑定检查必须使用 SELECT FOR UPDATE 行锁
- 激活码状态更新必须使用乐观锁/版本号
- 错误码必须符合 design.md 中的定义

**Acceptance（验证标准）:**
- AC: 激活成功时返回 bound_devices 字段
- AC: 设备绑定数量达到上限时返回错误码 4001
- AC: 激活码已被取消时返回错误码 4003
- AC: 并发激活同一激活码时只有一个成功
- AC: go build ./... 零错误
- AC: 【回归】现有激活流程不受影响（RG-2）

### Task 5: 新增设备解绑接口 ⬜

**复杂度**: 中
**依赖**: Task 4

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 解绑流程必须使用数据库事务
- 必须检查解绑次数（unbind_count < max_unbind_count）
- 解绑操作必须使用 SELECT FOR UPDATE 行锁

**Acceptance（验证标准）:**
- AC: 解绑成功时返回成功响应
- AC: 解绑次数达到上限时返回错误码 4005
- AC: go build ./... 零错误

### Task 6: 新增购买订单创建接口 ⬜

**复杂度**: 高
**依赖**: Task 5

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 创建订单后必须调用微信H5支付API
- 订单号必须唯一
- 必须返回 pay_url 给前端
- 微信API调用失败时订单状态保持 pending，用户可以重新尝试

**Acceptance（验证标准）:**
- AC: 订单创建成功时返回 order_no 和 pay_url
- AC: 微信API调用失败时订单状态保持 pending
- AC: go build ./... 零错误

### Task 7: 新增微信支付回调接口 ⬜

**复杂度**: 高
**依赖**: Task 6

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 必须验证微信签名（RSA-SHA256）
- 必须使用订单号作为幂等键
- 重复回调时直接返回成功

**Acceptance（验证标准）:**
- AC: 支付成功后更新订单状态为 paid
- AC: 支付成功后生成/分配激活码
- AC: 重复回调不重复处理
- AC: 签名验证失败时返回错误
- AC: go build ./... 零错误

### Task 8: 新增订单状态查询接口 ⬜

**复杂度**: 中
**依赖**: Task 7

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 支持单个订单查询（by order_no）
- 支持订单列表查询（带分页）

**Acceptance（验证标准）:**
- AC: 单个订单查询返回正确数据
- AC: 订单列表查询支持分页
- AC: go build ./... 零错误

### Task 9: 新增价格配置查询接口 ⬜

**复杂度**: 低
**依赖**: Task 8

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 返回所有激活码类型的价格配置

**Acceptance（验证标准）:**
- AC: 返回正确的价格配置列表
- AC: go build ./... 零错误

### Task 10: 新增淘宝订单推送接口 ⬜

**复杂度**: 高
**依赖**: Task 9

**Scope（边界）:**
- 涉及文件: backend/app/license/service/license_service.go, backend/app/license/apis/license.go, backend/app/license/plugin.go
- 涉及模块: license
- 不触碰: 其他模块的 service/handler 文件

**Constraints（约束）:**
- 支持批量推送
- order_status = paid 时创建或确认激活码
- order_status = refunded/closed 时取消激活码
- 会员时长合并计算：用户会员到期时间 = 所有激活码时长的总和
- 取消激活码时：激活码状态改为 canceled，从用户总时长中减去该激活码的时长（直到0为止）
- 如果用户总时长变为0或负数，会员立即失效

**Acceptance（验证标准）:**
- AC: paid 状态正确创建/确认激活码
- AC: refunded/closed 状态正确取消激活码
- AC: 取消已激活的激活码时正确减去对应时长
- AC: 用户总时长变为0时会员立即失效
- AC: 批量推送多个订单时正确处理
- AC: 重复推送同一订单时不重复创建激活码
- AC: go build ./... 零错误

---

## Phase 3: 前端开发

### Task 11: 前端激活码管理页面（设备绑定信息展示） ⬜

**复杂度**: 中
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

### Task 12: 前端购买流程页面（微信H5支付） ⬜

**复杂度**: 高
**依赖**: Task 6, Task 7, Task 8, Task 9

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/license/Purchase.vue（新建）
- 涉及模块: license
- 不触碰: 其他模块的 Vue 文件

**Constraints（约束）:**
- 必须使用 @/utils/request（axios wrapper）
- 必须支持 WebView 打开支付链接
- 必须轮询订单状态（超时5分钟）
- 支付失败判断：轮询超时或用户主动取消
- 支付失败后用户可以重新支付

**Acceptance（验证标准）:**
- AC: 购买流程正常工作
- AC: 支付成功后正确显示激活状态
- AC: 轮询超时（5分钟）后提示支付失败，用户可以重新支付
- AC: 用户主动取消支付后提示支付失败，用户可以重新支付

### Task 13: 前端订单管理页面 ⬜

**复杂度**: 中
**依赖**: Task 8

**Scope（边界）:**
- 涉及文件: dev-web-admin/src/views/license/Orders.vue（新建）
- 涉及模块: license
- 不触碰: 其他模块的 Vue 文件

**Constraints（约束）:**
- 必须使用 @/utils/request（axios wrapper）
- 必须支持分页

**Acceptance（验证标准）:**
- AC: 订单列表正确展示
- AC: 分页正常工作
- AC: 订单状态正确展示
