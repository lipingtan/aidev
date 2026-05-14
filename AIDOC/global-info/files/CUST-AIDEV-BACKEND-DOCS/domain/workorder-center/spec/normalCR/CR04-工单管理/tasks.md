# CR04 - 工单管理 实施任务列表

| 项目 | 内容 |
|------|------|
| CR 编号 | CR04 |
| 需求名称 | 工单管理 |
| 版本 | 1.1 |
| 创建日期 | 2026-04-19 |
| 需求文档 | requirements.md（版本 2.0，已确认） |
| 设计文档 | design.md（版本 1.0，已确认） |
| 后端语言 | Java 1.8 + Spring Boot 2.4.4 + MyBatis Plus |
| 前端语言 | TypeScript + Vue 3 + Element Plus（PC）/ Vant（H5） |

---

## 概述

本任务列表将 CR04 工单管理的设计方案拆解为可执行的编码任务。任务按依赖关系排列，标注可并行执行的任务组。前后端任务分开但标注依赖关系。

**并行组说明：**
- 并行组 A：数据库 + 基础代码（无依赖，最先执行）
- 并行组 B：后端 Service 层 + Config（依赖并行组 A）
- 并行组 C：后端 Controller 层 + 对外 API（依赖并行组 B）
- 并行组 D：PC 前端（依赖并行组 C）
- 并行组 E：H5 前端（依赖并行组 C）
- 并行组 D 和 E 可同时执行

---

## 任务列表

### 并行组 A：数据库 + 基础代码层

- [ ] 1. 数据库 DDL/DML 脚本
  - [ ] 1.1 创建 7 张业务表的 DDL 脚本
    - wo_work_order、wo_work_order_image、wo_dispatch_record、wo_repair_material、wo_evaluation、wo_urge_record、wo_work_order_log
    - 按 design.md 第三章 DDL 定义，含所有索引和约束
    - _Requirements: R1-R8 数据模型_
  - [ ] 1.2 创建 DML 初始化脚本
    - 报修类型字典数据（base_dict_category + base_dict，WORKORDER_TYPE）
    - 菜单初始化数据（sys_menu 表，ID 从 12000 开始）+ 角色关联
    - _Requirements: US-01, NFR-05_

- [ ] 2. 常量、枚举、错误码
  - [ ] 2.1 创建 WorkOrderErrorCode 错误码枚举（5000-5999）
    - 按 design.md 第十六章定义 20 个错误码
    - _Requirements: 全模块错误处理_
  - [ ] 2.2 创建 WorkOrderConstants 常量类 + 6 个枚举类
    - WorkOrderStatus（7 种状态）
    - AddressType（HOUSE/PUBLIC）
    - ImageType（REPORT/REPAIR/REJECT）
    - DispatchType（MANUAL/AUTO）
    - WorkOrderAction（10 种操作类型）
    - EvaluateType（OWNER/AUTO）
    - _Requirements: US-01~US-08_

- [ ] 3. DO 实体 + Mapper 接口 + XML 映射文件
  - [ ] 3.1 创建 7 个 DO 实体类
    - WorkOrderDO（继承 BaseEntity，核心字段含 community_id 冗余）
    - WorkOrderImageDO、DispatchRecordDO、RepairMaterialDO
    - EvaluationDO、UrgeRecordDO、WorkOrderLogDO
    - _Requirements: 数据模型_
  - [ ] 3.2 创建 7 个 Mapper 接口 + 7 个 XML 映射文件
    - WorkOrderMapper（含分页查询、数据权限 JOIN、统计查询）
    - 其他 6 个 Mapper（基础 CRUD + 按 order_id 查询）
    - _Requirements: NFR-04 数据权限_

- [ ] 4. DTO 和 VO 类
  - [ ] 4.1 创建 10 个 DTO 类
    - WorkOrderCreateDTO、WorkOrderQueryDTO、WorkOrderDispatchDTO
    - WorkOrderCancelDTO、WorkOrderVerifyDTO、WorkOrderEvaluateDTO
    - WorkOrderCompleteDTO、RepairMaterialDTO、StatisticsQueryDTO、H5WorkOrderQueryDTO
    - 含 JSR-303 校验注解
    - _Requirements: US-01~US-08_
  - [ ] 4.2 创建 12 个 VO 类
    - WorkOrderListVO、WorkOrderDetailVO、WorkOrderSimpleVO、WorkOrderProgressVO
    - DispatchRecordVO、RepairMaterialVO、EvaluationVO、WorkOrderLogVO
    - StatisticsVO、TrendDataVO、RepairDashboardVO、RepairStaffVO
    - _Requirements: 全模块接口返回_

- [ ] 5. 检查点 — 基础代码层编译通过

---

### 并行组 B：后端 Service 层 + Config（依赖并行组 A）

- [ ] 6. 工单核心 Service（WorkOrderService + WorkOrderQueryService）
  - [ ] 6.1 实现 WorkOrderService 接口 + WorkOrderServiceImpl
    - dispatchWorkOrder：状态校验 + 派发记录 + 通知事件
    - cancelWorkOrder：状态校验 + 取消原因 + 通知事件
    - acceptWorkOrder：状态校验 + 更新实际开始时间 + 通知事件
    - completeWorkOrder：状态校验 + 材料保存 + 时长记录 + 通知事件
    - transferWorkOrder：状态校验 + 退回待派发
    - verifyWorkOrder：通过/不通过分支 + 不通过次数校验 + 升级通知
    - _Requirements: US-02~US-06_
  - [ ] 6.2 实现 WorkOrderQueryService 接口 + WorkOrderQueryServiceImpl
    - listWorkOrders：分页查询 + 数据权限过滤
    - getWorkOrderDetail：聚合查询（图片 + 派发记录 + 材料 + 评价 + 日志）
    - _Requirements: US-08_

- [ ] 7. H5 业主端 Service + H5 维修人员端 Service
  - [ ] 7.1 业主端：createWorkOrder（编号生成 + 图片保存 + 字典校验 + OwnerApi 校验）
    - _Requirements: US-01_
  - [ ] 7.2 业主端：listMyWorkOrders、cancelWorkOrder（业主只能取消自己的）
    - _Requirements: US-01, US-06_
  - [ ] 7.3 维修人员端：getDashboard（今日待处理/本月完成统计）
    - _Requirements: US-03_
  - [ ] 7.4 维修人员端：listPendingOrders、listHistoryOrders
    - _Requirements: US-03_

- [ ] 8. 派发服务 + 编号生成 + 状态机
  - [ ] 8.1 实现 OrderNoGenerator（Redis 自增 + 降级方案）
    - _Requirements: NFR-01_
  - [ ] 8.2 实现 WorkOrderStateMachine（状态流转校验 EnumMap）
    - _Requirements: 状态流转规则_
  - [ ] 8.3 实现 DispatchService + DispatchStrategyEngine（楼栋优先 + 工作量均衡）
    - _Requirements: US-02 自动派发_

- [ ] 9. 评价 + 催单 + 统计 Service
  - [ ] 9.1 实现 EvaluationService（验收评价 + 自动验收默认 5 星）
    - _Requirements: US-04_
  - [ ] 9.2 实现 UrgeService（催单记录 + 次数累计 + 通知）
    - _Requirements: US-05_
  - [ ] 9.3 实现 WorkOrderStatisticsService（实时计算 + 趋势图 SQL 聚合）
    - _Requirements: US-08_

- [ ] 10. Config 层（定时任务 + 事件 + 缓存 + 文件上传）
  - [ ] 10.1 实现 WorkOrderEventPublisher + WorkOrderEventListener（9 种事件）
    - _Requirements: 通知机制_
  - [ ] 10.2 实现 AcceptTimeoutTask（30 分钟检测，2h 提醒 + 3h 退回）
    - _Requirements: US-07_
  - [ ] 10.3 实现 AutoVerifyTask（1 小时检测，7 天自动完成）
    - _Requirements: US-04_
  - [ ] 10.4 实现 WorkOrderCacheService（列表缓存 5 分钟 + 清除策略）
    - _Requirements: NFR-02_
  - [x] 10.5 common 模块新建 FileService + FileController（统一文件上传）
    - FileServiceImpl 实现三重安全校验（文件大小 + MIME Type 白名单 + 文件头魔数校验）
    - 配置化管理（`file.upload.max-size`、`file.upload.allowed-types`、`file.upload.path`）
    - FileController 提供 `/api/v1/common/files/upload` 和 `/{category}/{filename}` 接口
    - _Requirements: NFR-03, DQ1_
    - _Design: 第九章 9.1~9.4_

- [ ] 11. 检查点 — Service 层 + Config 编译通过

---

### 并行组 C：后端 Controller 层 + 对外 API（依赖并行组 B）

- [x] 12. PC 端 Controller（WorkOrderController + WorkOrderStatisticsController）
  - [x] 12.1 WorkOrderController：列表/详情/派发/取消/维修人员列表
    - @PreAuthorize + @OperationLog 注解
    - _Requirements: US-02, US-06, US-08_
  - [x] 12.2 WorkOrderStatisticsController：统计看板 + 趋势数据
    - _Requirements: US-08_
  - [ ] 12.3 Excel 导出（EasyExcel，限 1000 条，数据权限过滤）
    - _Requirements: US-08_

- [x] 13. H5 业主端 Controller（H5WorkOrderController）
  - [x] 13.1 提交报修/我的工单/详情/验收/评价/催单/取消
    - @RequireCertified + @OperationLog 注解
    - _Requirements: US-01, US-04, US-05, US-06_

- [x] 14. H5 维修人员端 Controller（H5RepairController）
  - [x] 14.1 工作台/待处理/历史/接单/完成/转派
    - @PreAuthorize REPAIR_STAFF 角色 + @OperationLog 注解
    - _Requirements: US-03_

- [x] 15. 对外 API（WorkOrderApi + WorkOrderApiImpl）
  - [x] 15.1 WorkOrderApi 接口定义 + 3 个 DTO
    - getWorkOrderBrief、countPendingByBuildingId、countByCommunityId、listByOwnerId
    - _Requirements: 对外 API_

- [ ] 16. 检查点 — 后端全模块编译通过 + 启动验证

---

### 并行组 D：PC 管理端前端（依赖并行组 C）

- [ ] 17. PC 前端 API 封装 + Store
  - [x] 17.1 创建 workorder API 模块（`src/api/workorder/`）
    - workOrder API（列表/详情/派发/取消/导出/统计/维修人员列表）
    - _Requirements: US-02, US-06, US-08_
  - [x] 17.2 创建 common upload API 模块（`src/api/common/upload.ts`）
    - uploadFile、validateFile、uploadFiles 方法 + UPLOAD_CONSTANTS 常量
    - _Requirements: NFR-03_
  - [ ] 17.3 创建 workorder Pinia Store
    - _Requirements: 全模块状态管理_

- [x] 18. PC 前端页面 — 工单列表
  - [x] 18.1 工单列表页（`src/views/workorder/order/index.vue`）
    - 多条件筛选（状态/类型/时间/小区/楼栋/关键词）
    - 派发弹窗（选择维修人员 + 备注 + 预计完成时间）
    - 取消/关闭弹窗
    - _Requirements: US-02, US-06, US-08_

- [x] 19. PC 前端页面 — 工单详情
  - [x] 19.1 工单详情页（`src/views/workorder/order/detail.vue`）
    - 基本信息 + 状态步骤条 + 图片查看（el-image + preview-src-list 支持点击放大） + 派发记录 + 维修材料 + 评价信息 + 操作日志时间线
    - _Requirements: US-08_
    - _Design: 9.5.4 PC 工单详情图片预览_

- [x] 20. PC 前端页面 — 工单统计
  - [x] 20.1 统计看板页（`src/views/workorder/statistics/index.vue`）
    - 统计卡片（待处理/处理中/本月完成/平均时长/满意度）
    - 按小区/楼栋维度筛选
    - 趋势图（ECharts：每日工单量 + 月度满意度）
    - _Requirements: US-08_

- [x] 21. PC 前端路由 + 菜单集成
  - [x] 21.1 注册 workorder 路由（`/workorder/list`、`/workorder/list/:id`、`/workorder/statistics`）
  - [x] 21.2 侧边栏菜单分组（Tickets 图标，AppSidebar.vue 新增工单管理分组）

---

### 并行组 E：H5 前端（依赖并行组 C，可与并行组 D 同时执行）

- [x] 22. H5 业主端 API 封装
  - [x] 22.1 创建 H5 workorder API 模块（`src/api/workorder.ts`）
    - 提交报修/我的工单/详情/验收/评价/催单/取消
    - _Requirements: US-01, US-04, US-05, US-06_
  - [x] 22.2 创建 H5 common upload API 模块（`src/api/common/upload.ts`）
    - uploadFile、validateFile、uploadFiles 方法 + UPLOAD_CONSTANTS 常量
    - _Requirements: NFR-03_

- [x] 23. H5 业主端页面
  - [x] 23.1 提交报修页（`src/views/workorder/WorkorderView.vue`，集成在我的工单页底部弹窗）
    - 报修类型选择（字典）+ 问题描述 + 地址选择（房屋/公共区域）+ 图片上传（van-uploader，最多5张）
    - 图片上传：after-read 回调调用 uploadFile，上传状态展示，提交时传 imageUrls
    - _Requirements: US-01_
    - _Design: 9.5.4 H5 提交报修_
  - [x] 23.2 我的工单列表页（集成在 WorkorderView.vue 中）
    - 状态筛选 Tab + 工单卡片列表 + 下拉刷新 + 无限加载
    - _Requirements: US-01_
  - [x] 23.3 工单详情页（`src/views/workorder/detail.vue`）
    - 进度步骤条 + 报修图片预览（van-image + showImagePreview） + 维修材料 + 验收操作（验收不通过时 van-uploader 上传证据照片，传 rejectImageUrls） + 评价 + 催单按钮
    - _Requirements: US-04, US-05_
    - _Design: 9.5.4 H5 工单详情_

- [x] 24. H5 维修人员端 API 封装
  - [x] 24.1 创建 H5 repair API 模块（`src/api/repair.ts`）
    - 工作台/待处理/历史/接单/完成/转派
    - _Requirements: US-03_

- [x] 25. H5 维修人员端页面
  - [x] 25.1 工作台首页（`src/views/repair/dashboard.vue`）
    - 今日待处理数 + 本月完成数 + 快捷入口
    - _Requirements: US-03_
  - [x] 25.2 待处理工单列表页（`src/views/repair/pending.vue`）
    - 待接单 + 处理中列表 + 下拉刷新 + 无限加载
    - _Requirements: US-03_
  - [x] 25.3 历史工单列表页（`src/views/repair/history.vue`）
    - 已完成工单
    - _Requirements: US-03_
  - [x] 25.4 工单处理页（`src/views/repair/handle.vue`）
    - 接单 + 上传照片（van-uploader，完成维修时上传，传 imageUrls） + 填写材料 + 完成 + 转派
    - _Requirements: US-03_
    - _Design: 9.5.4 H5 维修处理_

- [x] 26. H5 路由集成 + 角色判断
  - [x] 26.1 注册 workorder 和 repair 路由
  - [x] 26.2 角色判断逻辑（业主显示 workorder 入口，维修人员显示 repair 入口）

---

### 最终阶段

- [ ] 27. 前后端联调
  - [ ] 27.1 PC 端联调（列表/详情/派发/统计/导出）
  - [ ] 27.2 H5 业主端联调（提交/查看/验收/评价/催单）
  - [ ] 27.3 H5 维修人员端联调（工作台/接单/完成/转派）

- [ ] 28. 最终检查点
  - 编译通过 + 启动验证 + 核心流程走通
