---
inclusion: always
---

# SPMP 项目上下文和规范

## 语言规范

**所有与此项目相关的沟通必须使用中文**

- 所有回复、说明、文档都必须用中文
- 代码注释使用中文
- 提交信息（commit message）使用中文
- 文档编写使用中文

## 重要说明

在处理此项目的任何任务时，你必须：

1. **始终参考项目文档规范**: 查看 `CUST-AIDEV-BACKEND-DOCS/` 目录了解文档组织结构
2. **理解项目全局信息**: 参考以下文件理解项目背景
   - 业务概览: `CUST-AIDEV-BACKEND-DOCS/global-info/business-brief.md`
   - 技术栈: `CUST-AIDEV-BACKEND-DOCS/global-info/backend-tech-stack-components.md`
   - 项目结构: `CUST-AIDEV-BACKEND-DOCS/global-info/backend-tech-stack-structure.md`
   - 服务间调用说明: `CUST-AIDEV-BACKEND-DOCS/global-info/backend-service-api-call.md`
   - 各域代码结构: 查看对应域的 `CUST-AIDEV-BACKEND-DOCS/domain/{域名}-center/domain-share/code-structure.md`（如有）
   - 前端路由跳转说明: `CUST-AIDEV-FRONTEND-DOCS/global-info/frontend-pages-route.md`
   - 前端API集成说明: `CUST-AIDEV-FRONTEND-DOCS/global-info/frontend-page-with-api.md`

## 项目概述

**系统名称：** 智慧物业管理平台（SPMP - Smart Property Management Platform）

**业务定位：** 为物业公司提供数字化管理工具，覆盖业主管理、报修工单、缴费账单、社区公告、门禁管理等核心业务场景，提升物业服务效率和业主满意度。

**用户群体：**
- 超级管理员：系统全局管理，角色权限配置
- 物业管理员：所管辖小区的日常物业管理
- 片区经理：所辖片区多个小区的业务监管
- 楼栋管家：具体楼栋的日常管理和业主沟通
- 维修人员：处理分配的报修工单
- 业主：在线报修、缴费、查看公告、访客预约

## 工作流程规范

### 接到新需求或任务时

1. **查阅相关文档**: 检查 `CUST-AIDEV-BACKEND-DOCS/domain/` 下是否有相关功能的规范文档
2. **遵循目录结构**: 按照文档指南中定义的标准目录结构组织文档
3. **更新变更记录**: 在相应的 `change-log/` 目录记录代码变更

### 代码开发规范

1. **遵循项目结构**: 严格按照 `backend-tech-stack-structure.md` 中定义的包结构和命名约定
2. **使用正确的技术栈**: 参考 `backend-tech-stack-components.md` 使用项目中已有的框架和库版本
3. **理解业务领域**: 根据业务模块划分理解不同中台服务的职责

### 数据库脚本规范

- 表名使用下划线分隔，如 `work_order`、`owner_info`
- 字段统一使用下划线命名，如 `create_time`、`update_time`
- 逻辑删除使用 `del_flag` 字段

### 构建和部署

- 遵循 `backend-tech-stack-components.md` 中定义的构建顺序
- 先构建 `spmp-common` 基础模块，再构建中台服务，最后构建 BFF 层
- 注意模块间的依赖关系

## 服务架构说明

### 分层架构

```
前端层 (Frontend)
    │
    ▼
网关层 (Gateway)
    │
    ▼
BFF 层 (业务聚合层)
    │
    ▼
中台服务层 (Domain Service)
    │
    ▼
基础服务层 (Infrastructure)
```

### 核心模块说明

| 层级 | 模块 | 职责 |
|------|------|------|
| BFF 层 | spmp-admin-be | 管理端 BFF |
| BFF 层 | spmp-owner-be | 业主端 BFF |
| 中台层 | spmp-user-center | 用户中心（账号、角色、权限） |
| 中台层 | spmp-owner-center | 业主中心（业主信息、房产绑定） |
| 中台层 | spmp-workorder-center | 工单中心（报修、投诉、工单流转） |
| 中台层 | spmp-billing-center | 缴费中心（账单、支付、催收） |
| 中台层 | spmp-notice-center | 公告中心（公告发布、通知推送） |
| 中台层 | spmp-access-center | 门禁中心（访客预约、门禁记录） |
| 中台层 | spmp-base-center | 基础中心（小区、楼栋、房屋数据） |

## 文档维护

- 任何代码变更都应该在相应的 `change-log/` 中记录
- 新功能开发应该在 `CUST-AIDEV-BACKEND-DOCS/domain/` 下创建完整的规范目录
- 保持文档与代码同步

## 关键原则

1. **文档先行**: 重要变更前先查阅或更新相关文档
2. **规范一致**: 遵循已建立的命名和结构规范
3. **上下文感知**: 理解 SPMP 系统的多模块特性，不同模块可能有不同的业务规则
4. **可追溯性**: 确保变更可以追溯到需求文档和设计文档

## 前端菜单分组规范

PC 管理端侧边栏菜单按路由路径前缀自动分组。新增业务域时，前端路由必须使用统一的域前缀，并在 `AppSidebar.vue` 中注册对应的分组。

| 路由前缀 | 菜单分组 | 图标 |
|----------|---------|------|
| `base/` | 基础数据 | OfficeBuilding |
| `owner/` | 业主管理 | UserFilled |
| `workorder/` | 工单管理 | Tickets |
| `billing/` | 缴费管理 | Wallet |
| `notice/` | 公告管理 | Bell |
| `access/` | 门禁管理 | Lock |
| `system/` | 系统管理 | Setting |
| `log/` | 日志管理 | Document |

详细规范见 `CUST-AIDEV-FRONTEND-DOCS/global-info/frontend-pages-route.md` 中的"PC 管理端侧边栏菜单分组规范"章节。
