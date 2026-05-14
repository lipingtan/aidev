# Task-12~15 Controller 层 + 对外 API Change Log

## 基本信息

| 字段 | 内容 |
|------|------|
| 任务编号 | Task-12 ~ Task-15 |
| 任务名称 | Controller 层 + WorkOrderApi 对外接口 |
| 开发人员 | AI |
| 完成时间 | 2026-04-19 |
| 关联需求 | CR04-工单管理 |

## 变更概述

完成工单管理模块全部 Controller 层（PC 端、H5 业主端、H5 维修人员端）和对外 API 接口实现。

## 变更详情

### 新增

- `controller/WorkOrderController.java`：PC 端工单管理（列表/详情/派发/取消/维修人员列表）
- `controller/WorkOrderStatisticsController.java`：PC 端统计看板
- `controller/h5/H5WorkOrderController.java`：H5 业主端（提交报修/我的工单/详情/验收/评价/催单/取消）
- `controller/h5/repair/H5RepairController.java`：H5 维修人员端（工作台/待处理/历史/接单/完成/转派）
- `api/WorkOrderApi.java`：对外 API 接口定义（4 个方法）
- `api/WorkOrderApiImpl.java`：对外 API 实现

### 修改

- `service/H5OwnerWorkOrderService.java`：新增 evaluateWorkOrder 方法
- `service/impl/H5OwnerWorkOrderServiceImpl.java`：实现 evaluateWorkOrder + 新增 EvaluationMapper 依赖

## 影响范围

- 模块：workorder-center
- 接口变更：新增 20 个 REST API 接口
- 配置变更：无
