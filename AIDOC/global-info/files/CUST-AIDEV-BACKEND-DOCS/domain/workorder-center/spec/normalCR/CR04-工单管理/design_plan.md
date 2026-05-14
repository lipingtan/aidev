# CR04 - 工单管理 设计计划

| 项目 | 内容 |
|------|------|
| CR 编号 | CR04 |
| 需求名称 | 工单管理 |
| 所属模块 | workorder（工单中心） |
| 版本 | 1.0 |
| 创建日期 | 2026-04-19 |
| 状态 | 已完成 |
| 需求文档 | requirements.md（已确认，版本 2.0） |
| 依赖 | CR01（基础数据管理，已完成）、CR02（用户权限管理，已完成）、CR03（业主管理，已完成）、common 模块（已完成） |

---

## 一、设计范围

基于已确认的 requirements.md，本设计文档需覆盖以下模块：

| 序号 | 模块 | 设计内容 |
|------|------|----------|
| 1 | 架构概述 | 模块定位、依赖关系、请求处理流程 |
| 2 | 数据库设计 | 7 张表（wo_work_order、wo_work_order_image、wo_dispatch_record、wo_repair_material、wo_evaluation、wo_urge_record、wo_work_order_log）完整 DDL + 索引 + DML |
| 3 | 代码分层与包结构 | controller / service / repository / domain / api / constant / config 包结构 |
| 4 | 组件与接口设计 | PC 端 7 个接口 + H5 业主端 7 个接口 + H5 维修人员端 6 个接口 + WorkOrderApi 4 个方法，完整 DTO/VO 定义 |
| 5 | 工单状态机设计 | 状态流转枚举 + 状态校验 + 流转规则实现 |
| 6 | 工单编号生成方案 | Redis 自增 or 数据库序列 |
| 7 | 自动派发规则引擎 | 派发策略设计（楼栋优先、工作量均衡、类型匹配） |
| 8 | 文件上传方案 | 本地存储 + 统一上传接口（common 模块） |
| 9 | Spring Event 通知方案 | 事件定义、监听器、站内消息 + 短信预留 |
| 10 | 定时任务方案 | 接单超时检测 + 自动验收 |
| 11 | 数据权限方案 | 方案C 双重路径（房屋维度 + 业主维度） |
| 12 | 缓存方案 | 工单列表 Redis 缓存 Key 设计、TTL、清除策略 |
| 13 | 操作日志方案 | @OperationLog 使用 + wo_work_order_log 工单流转日志 |
| 14 | 跨模块调用 | BaseApi、OwnerApi、UserApi 调用方案 |
| 15 | 错误处理 | 错误码定义（5000-5999） |
| 16 | 菜单与权限 DML | sys_menu + sys_role_menu 初始化数据 |

---

## 二、设计步骤

- [x] 步骤 1：概述与架构设计（模块定位、分层架构、依赖关系、请求处理流程）
- [x] 步骤 2：数据库详细设计（7 张表完整 DDL + 索引 + DML 初始化数据）
- [x] 步骤 3：代码分层与包结构设计
- [x] 步骤 4：组件与接口设计（Controller/Service 接口签名 + DTO/VO 定义）
- [x] 步骤 5：工单状态机设计
- [x] 步骤 6：工单编号生成方案
- [x] 步骤 7：自动派发规则引擎设计
- [x] 步骤 8：文件上传方案设计
- [x] 步骤 9：Spring Event 通知方案设计
- [x] 步骤 10：定时任务方案设计
- [x] 步骤 11：数据权限方案设计（方案C 双重路径）
- [x] 步骤 12：缓存方案设计
- [x] 步骤 13：操作日志方案设计
- [x] 步骤 14：跨模块调用方案
- [x] 步骤 15：错误处理设计（错误码定义 5000-5999）
- [x] 步骤 16：菜单与权限 DML

---

## 三、参考资料

| 文档 | 路径 | 用途 |
|------|------|------|
| 需求文档 | requirements.md | 设计的输入，所有设计必须覆盖需求 |
| CR01 设计文档 | base-center/spec/normalCR/CR01-基础数据管理/design.md | 参考格式、缓存方案、BaseApi 接口定义 |
| CR02 设计文档 | user-center/spec/normalCR/CR02-用户权限管理/design.md | 参考 RBAC 权限、@OperationLog、DataPermissionInterceptor |
| CR03 设计文档 | owner-center/spec/normalCR/CR03-业主管理/design.md | 参考 @RequireCertified、AES 加密、OwnerApi 接口定义 |
| 技术栈结构 | global-info/backend-tech-stack-structure.md | 参考包结构和模块间通信规范 |
| 模块间通信 | global-info/backend-service-api-call.md | 参考 api 包通信模式 |

---

## 四、设计澄清问题

### DQ1：文件上传服务

[Question] 需求要求图片存储到本地目录。请确认：
1. 项目中是否已有统一的文件上传接口？如有，直接复用。
2. 如没有，是否同意在 common 模块中新建 `FileService` + `FileController`，提供统一的上传/下载接口，工单模块调用？
3. 文件存储根目录是否通过配置文件指定（如 `file.upload.path`）？

[Answer] 确认。

---

### DQ2：自动派发规则的实现优先级

[Question] 需求要求支持三种自动派发规则（楼栋优先、工作量均衡、类型匹配）。实现建议：
- **方案 A**：三种规则全部实现，按优先级依次匹配（先楼栋→再工作量→最后类型），V1 版本全功能
- **方案 B**：V1 版本仅实现"工作量均衡"（最简单实用），其余两种规则预留接口后续迭代

考虑到首次实现的复杂度，建议采用方案 A（全部实现但规则可配置）。是否同意？

[Answer] 确认。

---

### DQ3：工单编号生成方案

[Question] 需求要求 `WO` + 年月日 + 4 位序号。实现建议：
- **方案 A**：Redis 自增（`INCR workorder:seq:{yyyyMMdd}`），每日自动重置，高性能，需处理 Redis 不可用的降级
- **方案 B**：数据库查询 MAX(order_no) + 1，简单但并发性能差
- **方案 C**：base_sequence_config 序列号服务（如果 base 模块已实现序列号功能）

建议采用方案 A（Redis 自增），与项目已有的 Redis 基础设施一致。是否同意？

[Answer] 确认。

---

### DQ4：短信通知

[Question] 需求要求多种场景发送短信通知。请确认：
1. 当前是否已有短信平台对接？还是仅做接口预留（先记录日志，不实际发送）？
2. 如果需要对接，短信平台的 API 信息由谁提供？

[Answer] 确认。暂时接口预留后续对接。

---

### DQ5：统计看板趋势图数据

[Question] 需求要求"每日工单量趋势"和"月度满意度趋势"图表。数据查询方案建议：
- 趋势数据通过 SQL 聚合查询实时计算（按日期 GROUP BY），不额外建汇总表
- 前端传入时间范围，后端返回日期+数值的数组

是否同意此方案？还是需要建独立的统计汇总表来提升查询性能？

[Answer] 确认。

---

### DQ6：维修人员技能标签

[Question] 自动派发的"按报修类型匹配技能"规则需要维修人员有技能标签。实现建议：
- **方案 A**：在 sys_user 表增加 `skills` 字段（JSON 格式，如 `["PLUMBING","ELECTRICAL"]`），user 模块提供维护接口
- **方案 B**：新建 `wo_repair_staff_skill` 表，在 workorder 模块内管理维修人员技能
- **方案 C**：V1 版本不做技能匹配，仅按楼栋和工作量派发，技能匹配后续迭代

建议采用方案 C（V1 不做技能匹配），降低首次实现复杂度。是否同意？

[Answer] 确认。

---

## 等待用户回答

请逐一填写上述 **[Answer]** 部分，完成后回复我，我将基于你的回答按步骤生成正式的 `design.md`。
