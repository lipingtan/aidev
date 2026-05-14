# 域索引

快速定位各域对应的代码模块和文档。

---

## 中台服务层

| 域 | 代码模块 | 主要职责 | 文档链接 | 文档状态 |
|---|---------|---------|---------|---------|
| 用户中心 | spmp-user-center | 用户账号、角色、权限管理 | [user-center/](./user-center/) | ⏳ 待生成 |
| 业主中心 | spmp-owner-center | 业主信息、房产绑定、家庭成员 | [owner-center/](./owner-center/) | ⏳ 待生成 |
| 工单中心 | spmp-workorder-center | 报修工单、投诉、工单流转 | [workorder-center/](./workorder-center/) | ⏳ 待生成 |
| 缴费中心 | spmp-billing-center | 账单、支付、催收、统计 | [billing-center/](./billing-center/) | ⏳ 待生成 |
| 公告中心 | spmp-notice-center | 公告发布、通知推送 | [notice-center/](./notice-center/) | ⏳ 待生成 |
| 门禁中心 | spmp-access-center | 访客预约、门禁记录 | [access-center/](./access-center/) | ⏳ 待生成 |
| 基础中心 | spmp-base-center | 小区、楼栋、房屋基础数据 | [base-center/](./base-center/) | ⏳ 待生成 |

---

## BFF 层

| BFF | 代码模块 | 聚合的中台 | 文档状态 |
|-----|---------|-----------|---------|
| 管理端 BFF | spmp-admin-be | 全部中台服务 | ⏳ 待生成 |
| 业主端 BFF | spmp-owner-be | 用户、业主、工单、缴费、公告、门禁 | ⏳ 待生成 |

---

## 基础服务

| 服务 | 代码模块 | 主要职责 | 文档链接 | 文档状态 |
|-----|---------|---------|---------|---------|
| 公共组件 | com.spmp.common | 公共配置、异常处理、统一返回、安全组件、工具类 | [common-center/](./common-center/) | 📝 需求已完成 |
| 网关 | spmp-gateway | API 网关、路由转发、认证鉴权 | — | ⏳ 待生成 |

---

## 域编号对照表

| 编号 | 域 |
|-----|---|
| 00 | common-center |
| 01 | user-center |
| 02 | owner-center |
| 03 | workorder-center |
| 04 | billing-center |
| 05 | notice-center |
| 06 | access-center |
| 07 | base-center |
| 10-19 | BFF 层 |
| 99 | cross-domain |

---

## 快速导航

- [命名规范](./naming-convention.md)
- [需求模板](./templates/)
