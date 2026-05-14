# CR04 - 工单管理 需求规格说明书

| 字段 | 内容 |
|------|------|
| 文档名称 | 工单管理模块需求文档 |
| CR 编号 | CR04 |
| 版本 | 2.1 |
| 创建日期 | 2026-04-17 |
| 更新日期 | 2026-04-19 |
| 负责人 | 技术团队 |
| 状态 | 已确认 |
| 依赖 | CR01（基础数据）、CR02（用户权限）、CR03（业主管理） |

---

## 简介

本文档定义 SPMP 智慧物业管理平台工单管理模块（workorder）的需求。工单管理模块是平台的核心业务模块，负责报修工单的全生命周期管理，包括业主在线提交报修、工单派发（手动 + 自动）、维修人员处理、业主验收和满意度评价。该模块采用 Spring Boot 2.4.4 + Java 1.8 技术栈，位于 `com.spmp.workorder` 包下，数据库表前缀为 `wo_`，API 路径前缀为 `/api/v1/workorder/`。

### 已有基础设施（其他模块已完成）

| 组件 | 来源 | 说明 |
|------|------|------|
| Result / PageResult | common 模块 | 统一响应包装 |
| ErrorCode / BusinessException | common 模块 | 错误码和业务异常 |
| GlobalExceptionHandler | common 模块 | 全局异常处理 |
| BaseEntity（公共字段 + 自动填充） | common 模块 | create_by、create_time、update_by、update_time 自动填充 |
| RedisUtils | common 模块 | Redis 缓存工具 |
| DataPermissionInterceptor（五级数据权限） | common 模块 | SQL 层面数据权限过滤 |
| @OperationLog（操作日志注解） | user 模块 | AOP 切面自动记录操作日志 |
| RBAC 权限体系（@PreAuthorize） | user 模块 | 后端接口权限标识校验 |
| UserApi / PermissionApi | user 模块 | 用户信息和权限查询接口 |
| BaseApi | base 模块 | 小区/楼栋/单元/房屋查询接口 |
| OwnerApi | owner 模块 | 业主信息查询接口 |
| Excel 导出工具 | common 模块 | EasyExcel 导出 |

---

## 一、用户故事与验收标准

### US-01：业主提交报修

**作为** 已认证业主（H5端），
**我希望** 能够在线提交报修申请（包括自有房屋和公共区域），
**以便** 快速反馈房屋或公共区域的问题。

验收标准：
- SHALL 支持选择报修类型，报修类型通过 base-center 字典管理（base_dict）维护，字典分类编码为 `WORKORDER_TYPE`
- SHALL 必填字段：报修类型、问题描述、报修地址
- SHALL 报修地址支持两种模式：
  - 房屋报修：默认绑定业主已认证的房屋，通过级联选择（小区→楼栋→单元→房屋）
  - 公共区域报修：地址粒度为楼栋级或单元级，绑定小区下对应楼栋/单元
- SHALL 支持上传报修图片（最多 5 张，单张不超过 5MB）
- SHALL 提交后自动生成工单编号（`WO` + 年月日 + 4位序号），状态为"待派发"
- SHALL 业主可查看自己提交的工单列表和进度
- WHEN 业主未认证时，SHALL 不允许提交报修（复用 @RequireCertified 注解）

### US-02：工单派发

**作为** 楼栋管家/物业管理员（PC端），
**我希望** 能够将待派发的工单分配给维修人员（手动或自动），
**以便** 安排维修工作。

验收标准：
- SHALL 待派发工单列表受数据权限控制（方案C：通过报修房屋关联 base 五级层级 + 通过报修人关联业主房产绑定）
- SHALL 支持手动选择维修人员进行派发，维修人员通过 UserApi 查询角色为 `REPAIR_STAFF` 的用户
- SHALL 支持自动派发规则：
  - 按楼栋分配：优先派发给负责该楼栋的维修人员
  - 按工作量均衡分配：当前处理中工单数最少的维修人员优先
  - 按报修类型匹配技能：维修人员技能与报修类型匹配的优先
- SHALL 派发时可填写备注和预计完成时间
- SHALL 派发后工单状态变为"待接单"
- SHALL 派发后通过 Spring Event 发送站内消息和短信通知维修人员

### US-03：维修人员处理工单

**作为** 维修人员（H5端），
**我希望** 能够查看分配给我的工单并反馈处理结果，
**以便** 完成维修任务。

验收标准：
- SHALL 维修人员只能看到分配给自己的工单
- SHALL 支持"接单"操作，状态变为"处理中"，接单后通过 Spring Event 通知业主
- SHALL 支持"完成"操作，需填写：
  - 维修处理说明（必填）
  - 维修过程照片（最多 5 张）
  - 维修使用的材料（材料名称、数量、费用，支持多条）
  - 实际维修时长（分钟）
- SHALL 完成后状态变为"待验收"，通过 Spring Event 通知业主验收
- SHALL 支持"转派"操作，转派次数无限制，转派后工单状态变为"待派发"
- SHALL 维修人员可查看自己的历史工单（已完成的）
- SHALL 维修人员有"工作台"首页：今日待处理数、本月完成数等概览

### US-04：业主验收评价

**作为** 业主（H5端），
**我希望** 在维修完成后进行验收和评价，
**以便** 反馈维修质量。

验收标准：
- SHALL 工单状态为"待验收"时，业主可操作
- SHALL 验收通过：工单状态变为"已完成"
- SHALL 验收不通过：
  - 工单状态回退为"处理中"，指定同一维修人员继续处理
  - 必须填写不通过原因
  - 必须附带图片证据
  - 通过 Spring Event 通知维修人员（站内消息 + 短信）
  - 验收不通过次数最多 3 次，超过 3 次自动升级处理（通知楼栋管家）
- SHALL 支持满意度评分（1-5 星）和文字评价
- SHALL 超过 7 天未验收的工单自动完成（定时任务，默认 5 星评价）

### US-05：业主催单

**作为** 业主（H5端），
**我希望** 在工单处理过程中可以催促，
**以便** 加快工单处理进度。

验收标准：
- SHALL 业主可在"待派发"、"待接单"、"处理中"状态下发起催单
- SHALL 催单后通过 Spring Event 通知楼栋管家和维修人员（站内消息 + 短信）
- SHALL 记录催单时间
- SHALL 记录催单次数
- SHALL 记录催单状态（已催单/未催单）

### US-06：工单取消与强制关闭

**作为** 业主/物业管理员，
**我希望** 能够取消或强制关闭工单，
**以便** 处理不再需要的工单。

验收标准：
- SHALL 业主可在"待派发"、"待接单"状态下取消自己的工单
- SHALL 物业管理员可在"待派发"、"待接单"状态下取消工单
- SHALL 物业管理员可在"处理中"、"待验收"状态下强制关闭工单
- SHALL 取消/关闭工单必须填写原因
- SHALL 已派发的工单取消后，通过 Spring Event 通知维修人员
- SHALL 工单取消后不可恢复

### US-07：维修人员接单超时

**作为** 系统定时任务，
**我希望** 自动检测超时未接单的工单，
**以便** 及时提醒或退回。

验收标准：
- SHALL 维修人员超过 2 小时未接单，自动发送提醒通知（站内消息 + 短信）给维修人员
- SHALL 提醒后再过 1 小时仍未接单，工单自动退回"待派发"状态
- SHALL 退回时通过 Spring Event 通知楼栋管家（站内消息 + 短信）

### US-08：工单查询与统计

**作为** 物业管理员/片区经理（PC端），
**我希望** 能够查询和统计工单数据，
**以便** 了解维修服务质量。

验收标准：
- SHALL 支持按状态、类型、时间范围、小区、楼栋筛选
- SHALL 列表查询受数据权限控制（方案C）
- SHALL 支持导出工单列表（Excel），导出字段包括：工单编号、工单类型、状态、报修人、维修人员、创建时间、完成时间等
- SHALL 导出数量限制 1000 条
- SHALL 导出数据受数据权限控制
- SHALL 工单列表查询使用 Redis 缓存（5 分钟过期）
- SHALL 统计看板实时计算（不使用缓存），包含：
  - 时间范围选项：今日、本周、本月、自定义
  - 统计指标：待处理数、处理中数、本月完成数、平均处理时长、满意度
  - 按小区/楼栋维度统计
  - 趋势图：每日工单量趋势、月度满意度趋势
- SHALL 统计数据受数据权限控制（片区经理看片区数据，物业管理员看小区数据）

---

## 二、工单状态流转

### 2.1 状态定义

| 状态 | 编码 | 说明 |
|------|------|------|
| 待派发 | PENDING_DISPATCH | 业主提交后的初始状态 |
| 待接单 | PENDING_ACCEPT | 已派发给维修人员，等待接单 |
| 处理中 | IN_PROGRESS | 维修人员已接单，正在处理 |
| 待验收 | PENDING_VERIFY | 维修完成，等待业主验收 |
| 已完成 | COMPLETED | 验收通过或自动完成 |
| 已取消 | CANCELLED | 业主或管理员取消 |
| 已关闭 | FORCE_CLOSED | 管理员强制关闭 |

### 2.2 状态流转图

```
业主提交 → [待派发] ──派发──→ [待接单] ──接单──→ [处理中] ──完成──→ [待验收] ──验收通过──→ [已完成]
               │                  │                   │                   │
               ├─业主取消         ├─维修超时退回       ├─管理员强制关闭     ├─验收不通过→[处理中]
               ├─管理员取消       ├─业主取消           ├─转派→[待派发]     ├─超时7天→[已完成]
               └─自动派发         └─管理员取消         └─管理员强制关闭     └─管理员强制关闭
                                  └─接单超时退回
```

### 2.3 流转规则

| 操作 | 当前状态 | 目标状态 | 操作人 | 前置条件 |
|------|---------|---------|--------|---------|
| 派发 | PENDING_DISPATCH | PENDING_ACCEPT | 管理员/管家 | 选择维修人员 |
| 自动派发 | PENDING_DISPATCH | PENDING_ACCEPT | 系统 | 匹配自动派发规则 |
| 接单 | PENDING_ACCEPT | IN_PROGRESS | 维修人员 | 已派发给当前维修人员 |
| 接单超时退回 | PENDING_ACCEPT | PENDING_DISPATCH | 系统（定时） | 超过 2+1 小时未接单 |
| 完成 | IN_PROGRESS | PENDING_VERIFY | 维修人员 | 填写处理说明+材料+时长 |
| 转派 | IN_PROGRESS | PENDING_DISPATCH | 维修人员发起 | 无次数限制 |
| 验收通过 | PENDING_VERIFY | COMPLETED | 业主 | 填写评分+评价 |
| 验收不通过 | PENDING_VERIFY | IN_PROGRESS | 业主 | 填写原因+图片，同一维修人员 |
| 验收超时 | PENDING_VERIFY | COMPLETED | 系统（定时） | 超过 7 天未验收 |
| 取消 | PENDING_DISPATCH | CANCELLED | 业主/管理员 | 填写取消原因 |
| 取消 | PENDING_ACCEPT | CANCELLED | 业主/管理员 | 填写取消原因，通知维修人员 |
| 强制关闭 | IN_PROGRESS | FORCE_CLOSED | 管理员 | 填写关闭原因 |
| 强制关闭 | PENDING_VERIFY | FORCE_CLOSED | 管理员 | 填写关闭原因 |

---

## 三、数据模型

### 3.1 实体关系概要

| 实体 | 表名 | 说明 | 主要字段 |
|------|------|------|----------|
| 工单 | wo_work_order | 工单主体 | 工单编号、类型、描述、状态、地址类型、房屋ID/楼栋ID/单元ID、报修人（业主ID）、维修人员ID |
| 工单图片 | wo_work_order_image | 报修/维修/验收图片 | 工单ID、图片URL、图片类型（REPORT/REPAIR/REJECT） |
| 派发记录 | wo_dispatch_record | 派发/转派历史 | 工单ID、维修人员ID、派发人、派发类型（MANUAL/AUTO）、派发时间 |
| 维修材料 | wo_repair_material | 维修使用材料 | 工单ID、材料名称、数量、费用 |
| 评价记录 | wo_evaluation | 满意度评价 | 工单ID、评分、评价内容 |
| 催单记录 | wo_urge_record | 催单历史 | 工单ID、催单人、催单时间 |
| 工单操作日志 | wo_work_order_log | 状态流转记录 | 工单ID、操作类型、操作人、操作时间、备注 |

### 3.2 关键字段说明

**wo_work_order 核心字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_no | VARCHAR(32) | 工单编号，格式 WO202604190001 |
| order_type | VARCHAR(32) | 报修类型，字典编码 WORKORDER_TYPE |
| address_type | VARCHAR(20) | 地址类型：HOUSE-房屋、PUBLIC-公共区域 |
| house_id | BIGINT | 房屋报修时关联 bs_house.id |
| building_id | BIGINT | 公共区域报修时关联 bs_building.id |
| unit_id | BIGINT | 公共区域报修时关联 bs_unit.id（可选） |
| community_id | BIGINT | 小区ID，关联 bs_community.id（冗余，用于数据权限） |
| reporter_id | BIGINT | 报修人（业主ID），关联 ow_owner.id |
| reporter_name | VARCHAR(64) | 报修人姓名（冗余） |
| reporter_phone | VARCHAR(256) | 报修人手机号（AES 加密，冗余） |
| repair_user_id | BIGINT | 当前维修人员ID，关联 sys_user.id |
| status | VARCHAR(20) | 工单状态 |
| reject_count | INT | 验收不通过次数（累计） |
| urge_count | INT | 催单次数（累计） |
| last_urge_time | DATETIME | 最近催单时间 |
| expected_complete_time | DATETIME | 预计完成时间 |
| actual_start_time | DATETIME | 实际开始时间（接单时间） |
| actual_complete_time | DATETIME | 实际完成时间（维修完成时间） |
| repair_duration | INT | 实际维修时长（分钟） |
| auto_verify_time | DATETIME | 自动验收时间（定时任务用） |

---

## 四、通知机制

### 4.1 通知场景矩阵

| 场景 | 触发事件 | 通知对象 | 通知方式 |
|------|---------|---------|---------|
| 工单派发 | 派发/自动派发 | 维修人员 | 站内消息 + 短信 |
| 维修人员接单 | 接单 | 业主 | 站内消息 + 短信 |
| 维修完成 | 完成 | 业主 | 站内消息 + 短信 |
| 验收不通过 | 验收不通过 | 维修人员 | 站内消息 + 短信 |
| 验收不通过超3次 | 第3次不通过 | 楼栋管家（升级） | 站内消息 + 短信 |
| 接单超时提醒 | 2小时未接单 | 维修人员 | 站内消息 + 短信 |
| 接单超时退回 | 3小时未接单 | 楼栋管家 | 站内消息 + 短信 |
| 业主催单 | 催单 | 楼栋管家 + 维修人员 | 站内消息 + 短信 |
| 工单取消（已派发） | 取消 | 维修人员 | 站内消息 |

### 4.2 技术实现

- 所有通知通过 Spring ApplicationEvent 异步发布
- 事件监听器统一处理站内消息和短信发送
- 短信发送通过短信平台接口（预留，当前仅记录日志）

---

## 五、接口需求

### 5.1 PC 管理端接口

| 接口 | 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|------|
| 工单列表 | GET | `/api/v1/workorder/orders` | 分页查询，受数据权限 | workorder:list |
| 工单详情 | GET | `/api/v1/workorder/orders/{id}` | 含图片、派发记录、材料、评价、操作日志 | workorder:detail |
| 派发工单 | PUT | `/api/v1/workorder/orders/{id}/dispatch` | 手动派发，选择维修人员 | workorder:dispatch |
| 取消工单 | PUT | `/api/v1/workorder/orders/{id}/cancel` | 取消/强制关闭 | workorder:cancel |
| 工单统计 | GET | `/api/v1/workorder/orders/statistics` | 统计看板数据 | workorder:statistics |
| 工单导出 | GET | `/api/v1/workorder/orders/export` | Excel 导出，限 1000 条 | workorder:export |
| 维修人员列表 | GET | `/api/v1/workorder/staff` | 查询 REPAIR_STAFF 角色用户 | workorder:dispatch |

### 5.2 H5 业主端接口

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 提交报修 | POST | `/api/v1/workorder/h5/orders` | 业主提交报修 |
| 我的工单 | GET | `/api/v1/workorder/h5/orders/mine` | 业主的工单列表 |
| 工单详情 | GET | `/api/v1/workorder/h5/orders/{id}` | 含进度、图片、评价 |
| 验收 | PUT | `/api/v1/workorder/h5/orders/{id}/verify` | 验收通过/不通过 |
| 评价 | POST | `/api/v1/workorder/h5/orders/{id}/evaluate` | 满意度评价 |
| 催单 | POST | `/api/v1/workorder/h5/orders/{id}/urge` | 催单 |
| 取消工单 | PUT | `/api/v1/workorder/h5/orders/{id}/cancel` | 业主取消 |

### 5.3 H5 维修人员端接口

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 工作台概览 | GET | `/api/v1/workorder/h5/repair/dashboard` | 今日待处理、本月完成等 |
| 待处理工单 | GET | `/api/v1/workorder/h5/repair/pending` | 待接单+处理中列表 |
| 历史工单 | GET | `/api/v1/workorder/h5/repair/history` | 已完成工单列表 |
| 接单 | PUT | `/api/v1/workorder/h5/repair/orders/{id}/accept` | 接单 |
| 完成维修 | PUT | `/api/v1/workorder/h5/repair/orders/{id}/complete` | 提交维修结果+材料 |
| 转派 | PUT | `/api/v1/workorder/h5/repair/orders/{id}/transfer` | 转派（退回待派发） |

### 5.4 对外 API（WorkOrderApi）

| 方法 | 说明 | 调用方 |
|------|------|--------|
| `getWorkOrderBrief(Long id)` | 工单摘要信息 | notice、billing |
| `countPendingByBuildingId(Long buildingId)` | 楼栋待处理工单数 | base（统计展示） |
| `countByCommunityId(Long communityId)` | 按小区统计工单数 | 首页仪表板 |
| `listByOwnerId(Long ownerId)` | 业主的工单列表 | 其他模块展示 |

---

## 六、前端需求

### 6.1 PC 管理端

| 页面 | 路由 | 功能 |
|------|------|------|
| 工单列表 | `/workorder/list` | 多条件筛选（状态、类型、时间、小区、楼栋）、派发操作、查看详情 |
| 工单详情 | `/workorder/detail/:id` | 完整信息、操作按钮（派发/取消/关闭）、流转时间线、图片查看、材料清单、评价信息 |
| 工单统计 | `/workorder/statistics` | 统计看板（待处理/处理中/本月完成/平均时长/满意度）、按小区楼栋维度、趋势图、时间范围选择 |

### 6.2 H5 业主端

| 页面 | 路由 | 功能 |
|------|------|------|
| 提交报修 | `/workorder/create` | 选择报修类型（字典）、问题描述、地址选择（房屋/公共区域）、上传图片 |
| 我的工单 | `/workorder/mine` | 工单列表、状态筛选 Tab（全部/待派发/处理中/待验收/已完成） |
| 工单详情 | `/workorder/detail/:id` | 进度步骤条、图片、维修材料、验收操作、评价、催单 |

### 6.3 H5 维修人员端（共用业主端 H5 应用，角色判断显示）

| 页面 | 路由 | 功能 |
|------|------|------|
| 工作台 | `/repair/dashboard` | 今日待处理数、本月完成数、待接单/处理中快捷入口 |
| 待处理工单 | `/repair/pending` | 待接单/处理中列表 |
| 历史工单 | `/repair/history` | 已完成工单列表 |
| 工单处理 | `/repair/handle/:id` | 接单、上传照片、填写材料、完成、转派 |

---

## 七、定时任务

| 任务 | 执行频率 | 说明 |
|------|---------|------|
| 接单超时检测 | 每 30 分钟 | 超过 2 小时未接单→提醒，超过 3 小时→退回待派发 |
| 自动验收 | 每 1 小时 | 超过 7 天未验收→自动完成，默认 5 星 |

---

## 八、非功能性需求

### NFR-01：工单编号生成

- 格式：`WO` + 年月日 + 4 位序号（如 WO202604190001）
- 全局唯一，不可重复
- 工单创建时自动生成，后续不可修改
- 内部通过数据库序列或 Redis 自增实现，需保证并发安全

### NFR-02：性能要求

- 每日工单量级约 1000 条
- 工单列表查询响应 < 500ms
- 工单列表查询使用 Redis 缓存，过期时间 5 分钟
- 统计看板数据实时计算，不使用缓存
- 图片上传单张不超过 5MB

### NFR-03：图片存储

- 工单图片存储到本地文件系统目录
- 复用项目已有的统一文件上传服务/接口（如有）
- 如无统一文件上传服务，在 common 模块中新建一个 Task 实现文件上传逻辑

**文件上传安全性要求：**
- SHALL 后端进行三重校验：文件大小（≤5MB）、MIME Type 白名单（JPEG/PNG/GIF/WEBP）、文件头魔数校验（防止伪造扩展名）
- SHALL 后端配置化管理上传参数（`file.upload.max-size`、`file.upload.allowed-types`），不硬编码
- SHALL 前端上传前进行本地校验（文件类型 + 大小），减少无效请求
- SHALL 前端上传组件限制最大上传数量（最多 5 张），超出后禁止继续上传
- SHALL 上传过程中显示上传状态（上传中/成功/失败），失败时允许重新上传

**文件上传集成场景：**

| 场景 | 端 | 上传时机 | 图片用途 | 对应图片类型 |
|------|------|---------|---------|-------------|
| 业主提交报修 | H5 业主端 | 提交报修弹窗内 | 现场照片 | REPORT |
| 维修完成 | H5 维修人员端 | 完成维修弹窗内 | 维修过程照片 | REPAIR |
| 验收不通过 | H5 业主端 | 验收弹窗内（不通过时） | 不通过证据照片 | REJECT |
| 工单详情查看 | PC 管理端 + H5 业主端 | 只读展示 | 所有类型图片预览 | - |

### NFR-04：数据权限

- 采用方案C：双重路径实现
  - 路径1：通过报修房屋关联 base 五级层级（wo_work_order.community_id → bs_community → bs_district），复用 DataPermissionInterceptor
  - 路径2：通过报修人关联业主房产绑定（wo_work_order.reporter_id → ow_property_binding → bs_house → bs_community → bs_district）
- PC 端列表查询、导出、统计均受数据权限控制
- H5 端业主只能看自己的工单，维修人员只能看分配给自己的工单

### NFR-05：技术规范

- 表前缀：`wo_`
- 包名：`com.spmp.workorder`
- API 路径前缀：`/api/v1/workorder/`
- 报修类型：复用 base-center 字典管理（base_dict），分类编码 `WORKORDER_TYPE`
- 维修人员：通过 UserApi 查询角色 `REPAIR_STAFF` 的用户

---

## 九、关键决策记录

| 决策编号 | 决策内容 | 决策结果 | 原因 |
|---------|---------|---------|------|
| D-01 | 报修类型管理方式 | 使用 base_dict 字典管理 | 便于后续扩展，避免硬编码 |
| D-02 | 公共区域报修地址粒度 | 楼栋级或单元级 | 满足公共设施报修需求 |
| D-03 | 派发方式 | 手动 + 自动 | 自动派发提升效率，手动派发保留灵活性 |
| D-04 | 转派次数 | 无限制 | 实际场景中可能需要多次转派 |
| D-05 | 验收不通过处理 | 指定同一维修人员继续处理 | 保持处理连续性 |
| D-06 | 验收不通过上限 | 3 次后自动升级通知楼栋管家 | 避免无限循环 |
| D-07 | 维修材料记录 | 在工单模块中记录材料+费用 | 完整记录维修过程 |
| D-08 | 数据权限方案 | 方案C（双重路径） | 同时支持房屋维度和业主维度查询 |
| D-09 | 维修人员 H5 入口 | 共用业主端 H5，角色判断 | 降低维护成本 |
| D-10 | 接单超时 | 2 小时提醒 + 1 小时退回 | 平衡等待时间和处理效率 |
