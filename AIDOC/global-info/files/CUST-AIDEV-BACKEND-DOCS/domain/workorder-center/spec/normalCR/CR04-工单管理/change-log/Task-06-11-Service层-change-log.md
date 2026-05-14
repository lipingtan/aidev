# Task-06~11 Service 层 + Config Change Log

## 基本信息

| 字段 | 内容 |
|------|------|
| 任务编号 | Task-06 ~ Task-11 |
| 任务名称 | Service 层实现 + Config 配置层 |
| 开发人员 | AI |
| 完成时间 | 2026-04-19 |
| 关联需求 | CR04-工单管理 |

## 变更概述

完成工单管理模块全部 Service 层实现，包括核心工单服务、H5 业主端/维修人员端服务、派发策略、编号生成、状态机、评价/催单/统计服务、事件发布、定时任务、缓存和文件上传。

## 变更详情

### 新增

- `service/WorkOrderService.java` + `impl/WorkOrderServiceImpl.java`：核心工单操作（派发/取消/接单/完成/转派/验收）
- `service/WorkOrderQueryService.java` + `impl/WorkOrderQueryServiceImpl.java`：工单查询（分页列表+详情聚合）
- `service/H5OwnerWorkOrderService.java` + `impl/H5OwnerWorkOrderServiceImpl.java`：H5 业主端（提交报修/我的工单/验收/评价/催单/取消）
- `service/H5RepairWorkOrderService.java` + `impl/H5RepairWorkOrderServiceImpl.java`：H5 维修人员端（工作台/待处理/历史/接单/完成/转派）
- `service/DispatchService.java` + `impl/DispatchServiceImpl.java`：派发服务（手动+自动策略）
- `service/OrderNoGenerator.java`：工单编号生成（Redis 自增 + 日期前缀）
- `service/WorkOrderStateMachine.java`：状态流转校验（EnumMap）
- `service/EvaluationService.java` + `impl/EvaluationServiceImpl.java`：评价服务
- `service/UrgeService.java` + `impl/UrgeServiceImpl.java`：催单服务
- `service/WorkOrderStatisticsService.java` + `impl/WorkOrderStatisticsServiceImpl.java`：统计服务
- `config/WorkOrderEventPublisher.java` + `config/WorkOrderEventListener.java`：Spring Event 通知
- `config/AcceptTimeoutTask.java`：接单超时检测定时任务
- `config/AutoVerifyTask.java`：自动验收定时任务
- `service/impl/WorkOrderCacheServiceImpl.java`：Redis 缓存服务
- `common/service/impl/FileServiceImpl.java`：文件上传服务（三重安全校验）
- `common/controller/FileController.java`：文件上传/下载 Controller

### 修改

- `application.yml`：新增文件上传配置（file.upload.path/max-size/allowed-types）
- `common/service/FileService.java`：接口已存在，实现增强

## 数据库变更

### DDL 变更

- 无

### DML 变更

- 无

## 影响范围

- 模块：workorder-center + common
- 接口变更：新增 FileController 上传/下载接口
- 配置变更：application.yml 新增 file.upload 配置段
