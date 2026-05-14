# Task-17~26 前端页面（PC + H5）+ 文件上传集成 Change Log

## 基本信息

| 字段 | 内容 |
|------|------|
| 任务编号 | Task-17 ~ Task-26 |
| 任务名称 | PC 前端 + H5 前端 + 文件上传集成 |
| 开发人员 | AI |
| 完成时间 | 2026-04-19 |
| 关联需求 | CR04-工单管理 |

## 变更概述

完成工单管理模块全部前端页面开发，包括 PC 管理端（列表/详情/统计）、H5 业主端（报修/列表/详情）、H5 维修人员端（工作台/待处理/历史/处理），以及文件上传 API 封装和各页面图片上传集成。

## 变更详情

### 新增 — PC 前端（spmp-web-pc）

- `src/api/workorder/workOrder.ts`：工单管理 API 封装（12 个 DTO/VO 类型 + 6 个 API 方法）
- `src/api/common/upload.ts`：文件上传 API 封装（uploadFile/validateFile/uploadFiles + UPLOAD_CONSTANTS）
- `src/views/workorder/order/index.vue`：工单列表页（多条件筛选 + 派发弹窗 + 取消弹窗）
- `src/views/workorder/order/detail.vue`：工单详情页（基本信息 + 图片预览 + 派发记录 + 材料 + 评价 + 操作日志时间线）
- `src/views/workorder/statistics/index.vue`：统计看板（5 个统计卡片 + ECharts 趋势图）

### 新增 — H5 前端（spmp-web-h5）

- `src/api/workorder.ts`：业主端 API 封装（7 个方法）
- `src/api/repair.ts`：维修人员端 API 封装（6 个方法）
- `src/api/common/upload.ts`：文件上传 API 封装
- `src/views/workorder/WorkorderView.vue`：从占位页改为完整的工单列表 + 提交报修弹窗（含图片上传）
- `src/views/workorder/detail.vue`：工单详情页（步骤条 + 图片预览 + 验收弹窗含图片上传 + 评价 + 催单 + 取消）
- `src/views/repair/dashboard.vue`：维修人员工作台
- `src/views/repair/pending.vue`：待处理工单列表
- `src/views/repair/history.vue`：历史工单列表
- `src/views/repair/handle.vue`：工单处理页（接单 + 完成维修含图片上传 + 转派）

### 修改 — 路由和菜单

- `spmp-web-pc/src/router/static-routes.ts`：注册 workorder/list、workorder/list/:id、workorder/statistics 路由
- `spmp-web-pc/src/layout/AppSidebar.vue`：新增工单管理菜单分组（Tickets 图标 + DataAnalysis 图标）
- `spmp-web-h5/src/router/static-routes.ts`：注册工单详情 + 维修端 4 个路由

## 数据库变更

### DDL 变更

- 无

### DML 变更

- 无

## 影响范围

- 模块：spmp-web-pc、spmp-web-h5
- 接口变更：无后端接口变更
- 配置变更：无
