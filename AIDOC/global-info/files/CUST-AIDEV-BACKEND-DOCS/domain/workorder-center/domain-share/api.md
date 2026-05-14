# 工单管理 API 文档

## 接口总览

### PC 管理端

| 接口 | 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|------|
| 工单列表 | GET | `/api/v1/workorder/orders` | 分页查询，受数据权限 | workorder:list |
| 工单详情 | GET | `/api/v1/workorder/orders/{id}` | 含图片、派发记录、材料、评价、操作日志 | workorder:detail |
| 派发工单 | PUT | `/api/v1/workorder/orders/{id}/dispatch` | 手动派发 | workorder:dispatch |
| 取消工单 | PUT | `/api/v1/workorder/orders/{id}/cancel` | 取消/强制关闭 | workorder:cancel |
| 统计看板 | GET | `/api/v1/workorder/statistics` | 统计数据 | workorder:statistics |
| 维修人员列表 | GET | `/api/v1/workorder/orders/staff` | REPAIR_STAFF 角色用户 | workorder:dispatch |

### H5 业主端

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 提交报修 | POST | `/api/v1/workorder/h5/orders` | @RequireCertified |
| 我的工单 | GET | `/api/v1/workorder/h5/orders/mine` | @RequireCertified |
| 工单详情 | GET | `/api/v1/workorder/h5/orders/{id}` | @RequireCertified |
| 验收 | PUT | `/api/v1/workorder/h5/orders/{id}/verify` | @RequireCertified |
| 评价 | POST | `/api/v1/workorder/h5/orders/{id}/evaluate` | @RequireCertified |
| 催单 | POST | `/api/v1/workorder/h5/orders/{id}/urge` | @RequireCertified |
| 取消 | PUT | `/api/v1/workorder/h5/orders/{id}/cancel` | @RequireCertified |

### H5 维修人员端

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 工作台 | GET | `/api/v1/workorder/h5/repair/dashboard` | @PreAuthorize REPAIR_STAFF |
| 待处理工单 | GET | `/api/v1/workorder/h5/repair/pending` | @PreAuthorize REPAIR_STAFF |
| 历史工单 | GET | `/api/v1/workorder/h5/repair/history` | @PreAuthorize REPAIR_STAFF |
| 接单 | PUT | `/api/v1/workorder/h5/repair/orders/{id}/accept` | @PreAuthorize REPAIR_STAFF |
| 完成维修 | PUT | `/api/v1/workorder/h5/repair/orders/{id}/complete` | @PreAuthorize REPAIR_STAFF |
| 转派 | PUT | `/api/v1/workorder/h5/repair/orders/{id}/transfer` | @PreAuthorize REPAIR_STAFF |

### 对外 API（WorkOrderApi）

| 方法 | 说明 | 调用方 |
|------|------|--------|
| `getWorkOrderBrief(Long id)` | 工单摘要 | notice、billing |
| `countPendingByBuildingId(Long buildingId)` | 楼栋待处理数 | base |
| `countByCommunityId(Long communityId)` | 小区工单数 | 首页仪表板 |
| `listByOwnerId(Long ownerId)` | 业主工单列表 | 其他模块 |
