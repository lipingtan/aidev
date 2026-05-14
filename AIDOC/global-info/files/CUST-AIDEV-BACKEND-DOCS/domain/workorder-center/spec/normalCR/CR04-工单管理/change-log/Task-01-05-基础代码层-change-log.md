# Task-01~05 数据库 + 基础代码层 Change Log

## 基本信息

| 字段 | 内容 |
|------|------|
| 任务编号 | Task-01 ~ Task-05 |
| 任务名称 | 数据库 DDL/DML + 常量/枚举/错误码 + DO/Mapper + DTO/VO + XML 映射 |
| 开发人员 | AI |
| 完成时间 | 2026-04-19 |
| 关联需求 | CR04-工单管理 |

## 变更概述

完成工单管理模块基础代码层，包括数据库脚本、常量枚举、实体类、Mapper 接口和 XML 映射文件、DTO 和 VO 类。

## 变更详情

### 新增

- `src/spmp-backend/src/main/java/com/spmp/workorder/constant/WorkOrderErrorCode.java`：工单错误码枚举（5000-5999）
- `src/spmp-backend/src/main/java/com/spmp/workorder/constant/WorkOrderConstants.java`：工单常量类 + 6 个枚举（Status/AddressType/ImageType/DispatchType/Action/EvaluateType）
- `src/spmp-backend/src/main/java/com/spmp/workorder/domain/entity/WorkOrderDO.java`：工单主实体
- `src/spmp-backend/src/main/java/com/spmp/workorder/domain/entity/WorkOrderImageDO.java`：工单图片实体
- `src/spmp-backend/src/main/java/com/spmp/workorder/domain/entity/DispatchRecordDO.java`：派发记录实体
- `src/spmp-backend/src/main/java/com/spmp/workorder/domain/entity/RepairMaterialDO.java`：维修材料实体
- `src/spmp-backend/src/main/java/com/spmp/workorder/domain/entity/EvaluationDO.java`：评价记录实体
- `src/spmp-backend/src/main/java/com/spmp/workorder/domain/entity/UrgeRecordDO.java`：催单记录实体
- `src/spmp-backend/src/main/java/com/spmp/workorder/domain/entity/WorkOrderLogDO.java`：工单操作日志实体
- `src/spmp-backend/src/main/java/com/spmp/workorder/repository/WorkOrderMapper.java`：工单 Mapper 接口
- `src/spmp-backend/src/main/java/com/spmp/workorder/repository/WorkOrderImageMapper.java`：图片 Mapper
- `src/spmp-backend/src/main/java/com/spmp/workorder/repository/DispatchRecordMapper.java`：派发记录 Mapper
- `src/spmp-backend/src/main/java/com/spmp/workorder/repository/RepairMaterialMapper.java`：材料 Mapper
- `src/spmp-backend/src/main/java/com/spmp/workorder/repository/EvaluationMapper.java`：评价 Mapper
- `src/spmp-backend/src/main/java/com/spmp/workorder/repository/UrgeRecordMapper.java`：催单 Mapper
- `src/spmp-backend/src/main/java/com/spmp/workorder/repository/WorkOrderLogMapper.java`：操作日志 Mapper
- `src/spmp-backend/src/main/resources/mapper/workorder/WorkOrderMapper.xml`：工单 XML 映射（含分页查询、统计查询、数据权限 JOIN）
- `src/spmp-backend/src/main/resources/mapper/workorder/WorkOrderImageMapper.xml`：图片 XML 映射
- `src/spmp-backend/src/main/resources/mapper/workorder/DispatchRecordMapper.xml`：派发记录 XML 映射
- `src/spmp-backend/src/main/resources/mapper/workorder/RepairMaterialMapper.xml`：材料 XML 映射
- `src/spmp-backend/src/main/resources/mapper/workorder/EvaluationMapper.xml`：评价 XML 映射
- `src/spmp-backend/src/main/resources/mapper/workorder/UrgeRecordMapper.xml`：催单 XML 映射
- `src/spmp-backend/src/main/resources/mapper/workorder/WorkOrderLogMapper.xml`：操作日志 XML 映射
- `src/spmp-backend/src/main/java/com/spmp/workorder/domain/dto/`：10 个 DTO 类
- `src/spmp-backend/src/main/java/com/spmp/workorder/domain/vo/`：12 个 VO 类
- `src/spmp-backend/src/main/java/com/spmp/workorder/api/dto/`：3 个对外 API DTO 类

## 数据库变更

### DDL 变更

- 设计文档中已定义 7 张表的 DDL（wo_work_order、wo_work_order_image、wo_dispatch_record、wo_repair_material、wo_evaluation、wo_urge_record、wo_work_order_log）

### DML 变更

- 报修类型字典数据、菜单初始化数据（设计文档中已定义）

## 影响范围

- 模块：workorder-center
- 接口变更：无（本层仅定义数据结构和映射）
- 配置变更：无
