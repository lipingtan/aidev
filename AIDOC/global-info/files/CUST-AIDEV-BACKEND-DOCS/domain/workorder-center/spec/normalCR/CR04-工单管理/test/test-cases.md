# 工单管理模块测试用例

| 字段 | 内容 |
|------|------|
| 文档名称 | 工单管理模块测试用例 |
| CR 编号 | CR04 |
| 版本 | 1.0 |
| 创建日期 | 2026-04-19 |
| 负责人 | 技术团队 |
| 状态 | 草稿 |
| 需求文档 | requirements.md（版本 2.1） |
| 设计文档 | design.md（版本 1.1） |
| 测试策略 | test-strategy.md（版本 1.0） |

---

## 一、PC 管理端测试用例

### 1.1 工单列表查询

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| PC-001 | 正常查询工单列表 | 1. 调用 `/api/v1/workorder/orders` 接口<br>2. 传入分页参数 pageNum=1, pageSize=10 | 1. 返回 200 OK<br>2. 响应包含工单列表和分页信息<br>3. 数据权限正确过滤 | 高 |
| PC-002 | 按状态筛选 | 1. 调用 `/api/v1/workorder/orders` 接口<br>2. 传入 status=PENDING_DISPATCH | 1. 只返回待派发状态的工单 | 高 |
| PC-003 | 按类型筛选 | 1. 调用 `/api/v1/workorder/orders` 接口<br>2. 传入 orderType=WATER_ELECTRIC | 1. 只返回水电维修类型的工单 | 高 |
| PC-004 | 按时间范围筛选 | 1. 调用 `/api/v1/workorder/orders` 接口<br>2. 传入 startDate=2026-04-01, endDate=2026-04-30 | 1. 只返回 4 月份的工单 | 高 |
| PC-005 | 按小区/楼栋筛选 | 1. 调用 `/api/v1/workorder/orders` 接口<br>2. 传入 communityId=1, buildingId=2 | 1. 只返回指定小区和楼栋的工单 | 高 |
| PC-006 | 关键词搜索 | 1. 调用 `/api/v1/workorder/orders` 接口<br>2. 传入 keyword=WO20260419 | 1. 只返回工单编号包含 WO20260419 的工单 | 高 |
| PC-007 | 无权限访问 | 1. 使用无 workorder:list 权限的用户调用接口 | 1. 返回 403 Forbidden | 高 |

### 1.2 工单详情查询

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| PC-008 | 正常查询工单详情 | 1. 调用 `/api/v1/workorder/orders/{id}` 接口<br>2. 传入存在的工单 ID | 1. 返回 200 OK<br>2. 响应包含工单完整信息，包括图片、派发记录、材料、评价、操作日志 | 高 |
| PC-009 | 查询不存在的工单 | 1. 调用 `/api/v1/workorder/orders/{id}` 接口<br>2. 传入不存在的工单 ID | 1. 返回 404 Not Found | 高 |
| PC-010 | 无权限访问 | 1. 使用无 workorder:detail 权限的用户调用接口 | 1. 返回 403 Forbidden | 高 |

### 1.3 派发工单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| PC-011 | 正常派发工单 | 1. 调用 `/api/v1/workorder/orders/{id}/dispatch` 接口<br>2. 传入维修人员 ID 和备注 | 1. 返回 200 OK<br>2. 工单状态变为 PENDING_ACCEPT<br>3. 生成派发记录<br>4. 发送通知 | 高 |
| PC-012 | 派发不存在的工单 | 1. 调用 `/api/v1/workorder/orders/{id}/dispatch` 接口<br>2. 传入不存在的工单 ID | 1. 返回 404 Not Found | 高 |
| PC-013 | 派发非待派发状态的工单 | 1. 调用 `/api/v1/workorder/orders/{id}/dispatch` 接口<br>2. 传入非 PENDING_DISPATCH 状态的工单 ID | 1. 返回 400 Bad Request<br>2. 提示状态不允许变更 | 高 |
| PC-014 | 无权限派发 | 1. 使用无 workorder:dispatch 权限的用户调用接口 | 1. 返回 403 Forbidden | 高 |

### 1.4 取消/关闭工单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| PC-015 | 取消待派发工单 | 1. 调用 `/api/v1/workorder/orders/{id}/cancel` 接口<br>2. 传入取消原因 | 1. 返回 200 OK<br>2. 工单状态变为 CANCELLED | 高 |
| PC-016 | 取消待接单工单 | 1. 调用 `/api/v1/workorder/orders/{id}/cancel` 接口<br>2. 传入取消原因 | 1. 返回 200 OK<br>2. 工单状态变为 CANCELLED<br>3. 通知维修人员 | 高 |
| PC-017 | 强制关闭处理中工单 | 1. 调用 `/api/v1/workorder/orders/{id}/cancel` 接口<br>2. 传入关闭原因，cancelType=FORCE_CLOSE | 1. 返回 200 OK<br>2. 工单状态变为 FORCE_CLOSED | 高 |
| PC-018 | 强制关闭待验收工单 | 1. 调用 `/api/v1/workorder/orders/{id}/cancel` 接口<br>2. 传入关闭原因，cancelType=FORCE_CLOSE | 1. 返回 200 OK<br>2. 工单状态变为 FORCE_CLOSED | 高 |
| PC-019 | 无权限取消 | 1. 使用无 workorder:cancel 权限的用户调用接口 | 1. 返回 403 Forbidden | 高 |

### 1.5 工单统计

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| PC-020 | 查询今日统计 | 1. 调用 `/api/v1/workorder/orders/statistics` 接口<br>2. 传入 timeRange=TODAY | 1. 返回 200 OK<br>2. 响应包含今日的待处理数、处理中数、完成数等统计数据 | 高 |
| PC-021 | 查询本周统计 | 1. 调用 `/api/v1/workorder/orders/statistics` 接口<br>2. 传入 timeRange=WEEK | 1. 返回 200 OK<br>2. 响应包含本周的统计数据 | 高 |
| PC-022 | 查询本月统计 | 1. 调用 `/api/v1/workorder/orders/statistics` 接口<br>2. 传入 timeRange=MONTH | 1. 返回 200 OK<br>2. 响应包含本月的统计数据 | 高 |
| PC-023 | 按小区统计 | 1. 调用 `/api/v1/workorder/orders/statistics` 接口<br>2. 传入 communityId=1 | 1. 返回 200 OK<br>2. 响应包含指定小区的统计数据 | 高 |
| PC-024 | 无权限访问 | 1. 使用无 workorder:statistics 权限的用户调用接口 | 1. 返回 403 Forbidden | 高 |

### 1.6 工单导出

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| PC-025 | 正常导出工单 | 1. 调用 `/api/v1/workorder/orders/export` 接口<br>2. 传入查询参数 | 1. 返回 200 OK<br>2. 响应为 Excel 文件<br>3. 文件包含正确的工单数据 | 高 |
| PC-026 | 导出数量限制 | 1. 调用 `/api/v1/workorder/orders/export` 接口<br>2. 传入查询参数，预期结果超过 1000 条 | 1. 返回 400 Bad Request<br>2. 提示导出数量超过限制 | 高 |
| PC-027 | 无权限导出 | 1. 使用无 workorder:export 权限的用户调用接口 | 1. 返回 403 Forbidden | 高 |

### 1.7 维修人员列表

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| PC-028 | 查询维修人员列表 | 1. 调用 `/api/v1/workorder/staff` 接口 | 1. 返回 200 OK<br>2. 响应包含角色为 REPAIR_STAFF 的用户列表 | 高 |
| PC-029 | 按小区查询维修人员 | 1. 调用 `/api/v1/workorder/staff` 接口<br>2. 传入 communityId=1 | 1. 返回 200 OK<br>2. 响应包含指定小区的维修人员列表 | 高 |
| PC-030 | 无权限访问 | 1. 使用无 workorder:dispatch 权限的用户调用接口 | 1. 返回 403 Forbidden | 高 |

---

## 二、H5 业主端测试用例

### 2.1 提交报修

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-OWNER-001 | 正常提交房屋报修 | 1. 调用 `/api/v1/workorder/h5/orders` 接口<br>2. 传入报修类型、问题描述、房屋信息、图片 | 1. 返回 200 OK<br>2. 生成工单编号<br>3. 工单状态为 PENDING_DISPATCH | 高 |
| H5-OWNER-002 | 正常提交公共区域报修 | 1. 调用 `/api/v1/workorder/h5/orders` 接口<br>2. 传入报修类型、问题描述、公共区域信息、图片 | 1. 返回 200 OK<br>2. 生成工单编号<br>3. 工单状态为 PENDING_DISPATCH | 高 |
| H5-OWNER-003 | 缺少必填字段 | 1. 调用 `/api/v1/workorder/h5/orders` 接口<br>2. 不传入报修类型 | 1. 返回 400 Bad Request<br>2. 提示报修类型不能为空 | 高 |
| H5-OWNER-004 | 图片数量超限 | 1. 调用 `/api/v1/workorder/h5/orders` 接口<br>2. 传入超过 5 张图片 | 1. 返回 400 Bad Request<br>2. 提示图片数量超过限制 | 高 |
| H5-OWNER-005 | 未认证业主提交 | 1. 使用未认证的业主调用接口 | 1. 返回 401 Unauthorized<br>2. 提示需要认证 | 高 |

### 2.2 我的工单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-OWNER-006 | 查询我的工单列表 | 1. 调用 `/api/v1/workorder/h5/orders/mine` 接口<br>2. 传入分页参数 | 1. 返回 200 OK<br>2. 只返回当前业主的工单 | 高 |
| H5-OWNER-007 | 按状态筛选 | 1. 调用 `/api/v1/workorder/h5/orders/mine` 接口<br>2. 传入 status=IN_PROGRESS | 1. 只返回处理中状态的工单 | 高 |
| H5-OWNER-008 | 未认证业主查询 | 1. 使用未认证的业主调用接口 | 1. 返回 401 Unauthorized | 高 |

### 2.3 工单详情

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-OWNER-009 | 查询自己的工单详情 | 1. 调用 `/api/v1/workorder/h5/orders/{id}` 接口<br>2. 传入自己的工单 ID | 1. 返回 200 OK<br>2. 响应包含工单完整信息 | 高 |
| H5-OWNER-010 | 查询他人的工单 | 1. 调用 `/api/v1/workorder/h5/orders/{id}` 接口<br>2. 传入他人的工单 ID | 1. 返回 403 Forbidden | 高 |
| H5-OWNER-011 | 未认证业主查询 | 1. 使用未认证的业主调用接口 | 1. 返回 401 Unauthorized | 高 |

### 2.4 验收工单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-OWNER-012 | 验收通过 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/verify` 接口<br>2. 传入 passed=true, score=5, evaluateContent=满意 | 1. 返回 200 OK<br>2. 工单状态变为 COMPLETED<br>3. 生成评价记录 | 高 |
| H5-OWNER-013 | 验收不通过 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/verify` 接口<br>2. 传入 passed=false, rejectReason=未修复, rejectImageUrls=[图片URL] | 1. 返回 200 OK<br>2. 工单状态变为 IN_PROGRESS<br>3. 通知维修人员<br>4. reject_count+1 | 高 |
| H5-OWNER-014 | 验收非待验收工单 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/verify` 接口<br>2. 传入非 PENDING_VERIFY 状态的工单 ID | 1. 返回 400 Bad Request<br>2. 提示状态不允许变更 | 高 |
| H5-OWNER-015 | 未认证业主验收 | 1. 使用未认证的业主调用接口 | 1. 返回 401 Unauthorized | 高 |

### 2.5 评价工单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-OWNER-016 | 正常评价 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/evaluate` 接口<br>2. 传入 score=5, content=非常满意 | 1. 返回 200 OK<br>2. 生成评价记录 | 高 |
| H5-OWNER-017 | 评价非已完成工单 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/evaluate` 接口<br>2. 传入非 COMPLETED 状态的工单 ID | 1. 返回 400 Bad Request<br>2. 提示工单状态不允许评价 | 高 |
| H5-OWNER-018 | 未认证业主评价 | 1. 使用未认证的业主调用接口 | 1. 返回 401 Unauthorized | 高 |

### 2.6 催单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-OWNER-019 | 正常催单 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/urge` 接口<br>2. 传入待派发状态的工单 ID | 1. 返回 200 OK<br>2. 生成催单记录<br>3. 通知楼栋管家和维修人员<br>4. urge_count+1 | 高 |
| H5-OWNER-020 | 催单已完成工单 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/urge` 接口<br>2. 传入已完成状态的工单 ID | 1. 返回 400 Bad Request<br>2. 提示工单状态不允许催单 | 高 |
| H5-OWNER-021 | 未认证业主催单 | 1. 使用未认证的业主调用接口 | 1. 返回 401 Unauthorized | 高 |

### 2.7 取消工单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-OWNER-022 | 取消待派发工单 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/cancel` 接口<br>2. 传入取消原因 | 1. 返回 200 OK<br>2. 工单状态变为 CANCELLED | 高 |
| H5-OWNER-023 | 取消待接单工单 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/cancel` 接口<br>2. 传入取消原因 | 1. 返回 200 OK<br>2. 工单状态变为 CANCELLED<br>3. 通知维修人员 | 高 |
| H5-OWNER-024 | 取消处理中工单 | 1. 调用 `/api/v1/workorder/h5/orders/{id}/cancel` 接口<br>2. 传入取消原因 | 1. 返回 400 Bad Request<br>2. 提示状态不允许取消 | 高 |
| H5-OWNER-025 | 未认证业主取消 | 1. 使用未认证的业主调用接口 | 1. 返回 401 Unauthorized | 高 |

---

## 三、H5 维修人员端测试用例

### 3.1 工作台概览

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-REPAIR-001 | 查询工作台概览 | 1. 调用 `/api/v1/workorder/h5/repair/dashboard` 接口 | 1. 返回 200 OK<br>2. 响应包含今日待处理数、本月完成数等统计数据 | 高 |
| H5-REPAIR-002 | 非维修人员访问 | 1. 使用非 REPAIR_STAFF 角色的用户调用接口 | 1. 返回 403 Forbidden | 高 |

### 3.2 待处理工单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-REPAIR-003 | 查询待处理工单 | 1. 调用 `/api/v1/workorder/h5/repair/pending` 接口 | 1. 返回 200 OK<br>2. 只返回分配给当前维修人员的待接单和处理中工单 | 高 |
| H5-REPAIR-004 | 非维修人员访问 | 1. 使用非 REPAIR_STAFF 角色的用户调用接口 | 1. 返回 403 Forbidden | 高 |

### 3.3 历史工单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-REPAIR-005 | 查询历史工单 | 1. 调用 `/api/v1/workorder/h5/repair/history` 接口 | 1. 返回 200 OK<br>2. 只返回当前维修人员的已完成工单 | 高 |
| H5-REPAIR-006 | 非维修人员访问 | 1. 使用非 REPAIR_STAFF 角色的用户调用接口 | 1. 返回 403 Forbidden | 高 |

### 3.4 接单

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-REPAIR-007 | 正常接单 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/accept` 接口<br>2. 传入待接单状态的工单 ID | 1. 返回 200 OK<br>2. 工单状态变为 IN_PROGRESS<br>3. 通知业主 | 高 |
| H5-REPAIR-008 | 接单非待接单工单 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/accept` 接口<br>2. 传入非 PENDING_ACCEPT 状态的工单 ID | 1. 返回 400 Bad Request<br>2. 提示状态不允许变更 | 高 |
| H5-REPAIR-009 | 接单非分配给自己的工单 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/accept` 接口<br>2. 传入分配给其他维修人员的工单 ID | 1. 返回 403 Forbidden | 高 |

### 3.5 完成维修

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-REPAIR-010 | 正常完成维修 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/complete` 接口<br>2. 传入处理说明、维修时长、材料、图片 | 1. 返回 200 OK<br>2. 工单状态变为 PENDING_VERIFY<br>3. 生成维修材料记录<br>4. 通知业主 | 高 |
| H5-REPAIR-011 | 缺少处理说明 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/complete` 接口<br>2. 不传入 repairDescription | 1. 返回 400 Bad Request<br>2. 提示处理说明不能为空 | 高 |
| H5-REPAIR-012 | 完成非处理中工单 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/complete` 接口<br>2. 传入非 IN_PROGRESS 状态的工单 ID | 1. 返回 400 Bad Request<br>2. 提示状态不允许变更 | 高 |
| H5-REPAIR-013 | 完成非分配给自己的工单 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/complete` 接口<br>2. 传入分配给其他维修人员的工单 ID | 1. 返回 403 Forbidden | 高 |

### 3.6 转派

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| H5-REPAIR-014 | 正常转派 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/transfer` 接口<br>2. 传入转派原因 | 1. 返回 200 OK<br>2. 工单状态变为 PENDING_DISPATCH | 高 |
| H5-REPAIR-015 | 转派非处理中工单 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/transfer` 接口<br>2. 传入非 IN_PROGRESS 状态的工单 ID | 1. 返回 400 Bad Request<br>2. 提示状态不允许变更 | 高 |
| H5-REPAIR-016 | 转派非分配给自己的工单 | 1. 调用 `/api/v1/workorder/h5/repair/orders/{id}/transfer` 接口<br>2. 传入分配给其他维修人员的工单 ID | 1. 返回 403 Forbidden | 高 |

---

## 四、对外 API 测试用例

### 4.1 WorkOrderApi

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| API-001 | 获取工单摘要信息 | 1. 调用 WorkOrderApi.getWorkOrderBrief 方法<br>2. 传入工单 ID | 1. 返回工单摘要信息 | 高 |
| API-002 | 统计楼栋待处理工单数 | 1. 调用 WorkOrderApi.countPendingByBuildingId 方法<br>2. 传入楼栋 ID | 1. 返回该楼栋的待处理工单数 | 高 |
| API-003 | 按小区统计工单数 | 1. 调用 WorkOrderApi.countByCommunityId 方法<br>2. 传入小区 ID | 1. 返回该小区的工单统计信息 | 高 |
| API-004 | 获取业主的工单列表 | 1. 调用 WorkOrderApi.listByOwnerId 方法<br>2. 传入业主 ID | 1. 返回该业主的工单列表 | 高 |

---

## 五、核心服务测试用例

### 5.1 工单编号生成

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| CORE-001 | 生成工单编号 | 1. 调用 OrderNoGenerator.generate() 方法 | 1. 返回格式为 WO+年月日+4位序号的工单编号<br>2. 编号唯一 | 高 |
| CORE-002 | 每日序号重置 | 1. 调用 OrderNoGenerator.generate() 方法<br>2. 模拟日期变更<br>3. 再次调用 generate() 方法 | 1. 日期变更后序号从 0001 开始 | 高 |

### 5.2 自动派发策略

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| CORE-003 | 楼栋优先策略 | 1. 调用 DispatchStrategyEngine.autoSelectRepairUser 方法<br>2. 传入关联楼栋的工单 | 1. 优先选择负责该楼栋的维修人员 | 高 |
| CORE-004 | 工作量均衡策略 | 1. 调用 DispatchStrategyEngine.autoSelectRepairUser 方法<br>2. 传入无关联楼栋的工单 | 1. 选择当前处理中工单数最少的维修人员 | 高 |

### 5.3 状态机

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| CORE-005 | 合法状态变更 | 1. 调用 WorkOrderStateMachine.checkTransition 方法<br>2. 传入 PENDING_DISPATCH 到 PENDING_ACCEPT | 1. 无异常抛出 | 高 |
| CORE-006 | 非法状态变更 | 1. 调用 WorkOrderStateMachine.checkTransition 方法<br>2. 传入 COMPLETED 到 IN_PROGRESS | 1. 抛出 BusinessException | 高 |

### 5.4 定时任务

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| CORE-007 | 接单超时检测 | 1. 调用 AcceptTimeoutTask.execute() 方法<br>2. 准备超过 2 小时未接单的工单 | 1. 发送提醒通知<br>2. 超过 3 小时未接单的工单退回待派发 | 高 |
| CORE-008 | 自动验收 | 1. 调用 AutoVerifyTask.execute() 方法<br>2. 准备超过 7 天未验收的工单 | 1. 工单自动完成<br>2. 生成默认 5 星评价 | 高 |

---

## 六、数据权限测试用例

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| PERM-001 | 物业管理员数据权限 | 1. 使用物业管理员账号登录<br>2. 查询工单列表 | 1. 只能看到自己管理的小区的工单 | 高 |
| PERM-002 | 片区经理数据权限 | 1. 使用片区经理账号登录<br>2. 查询工单列表 | 1. 能看到自己管理的片区内所有小区的工单 | 高 |
| PERM-003 | 维修人员数据权限 | 1. 使用维修人员账号登录<br>2. 查询待处理工单 | 1. 只能看到分配给自己的工单 | 高 |
| PERM-004 | 业主数据权限 | 1. 使用业主账号登录<br>2. 查询我的工单 | 1. 只能看到自己提交的工单 | 高 |

---

## 七、性能测试用例

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| PERF-001 | 工单列表查询性能 | 1. 使用 JMeter 模拟 100 QPS 的并发请求<br>2. 调用工单列表查询接口 | 1. 平均响应时间 < 500ms<br>2. 99% 响应时间 < 1000ms | 高 |
| PERF-002 | 工单统计性能 | 1. 使用 JMeter 模拟 50 QPS 的并发请求<br>2. 调用工单统计接口 | 1. 平均响应时间 < 1000ms | 高 |
| PERF-003 | 文件上传性能 | 1. 使用 JMeter 模拟 20 QPS 的并发请求<br>2. 上传 1MB 大小的图片 | 1. 平均响应时间 < 2000ms | 中 |

---

## 八、安全性测试用例

| 用例编号 | 用例名称 | 测试步骤 | 预期结果 | 优先级 |
|----------|----------|----------|----------|----------|
| SEC-001 | 文件上传大小限制 | 1. 尝试上传超过 5MB 的文件 | 1. 返回 400 Bad Request<br>2. 提示文件大小超过限制 | 高 |
| SEC-002 | 文件上传类型限制 | 1. 尝试上传非图片文件（如 .exe） | 1. 返回 400 Bad Request<br>2. 提示文件类型不允许 | 高 |
| SEC-003 | 手机号加密存储 | 1. 查看数据库中 wo_work_order 表的 reporter_phone 字段 | 1. 存储的是加密后的字符串 | 高 |
| SEC-004 | SQL 注入防护 | 1. 在关键词搜索中输入 SQL 注入语句 | 1. 系统正常处理，无 SQL 注入风险 | 高 |