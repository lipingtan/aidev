---
inclusion: always
---

# Spec 文档输出位置规范

## 核心规则

**禁止**将需求文档（requirements.md）、设计文档（design.md）、任务文档（tasks.md）等 spec 相关文件生成到 `.kiro/specs/` 目录下。

所有 spec 文档必须生成到项目的 CR 目录中，路径格式为：

```
CUST-AIDEV-BACKEND-DOCS/domain/{域名}-center/spec/normalCR/CRxx-{需求名}/
```

## 具体要求

1. 当用户提出新需求时，先确认归属的域（如 common、user、owner、workorder 等）
2. 查看该域的 `spec/normalCR/` 目录下已有的 CR 编号，确定新 CR 的序号
3. 在正确的 CR 目录下创建 spec 文档，而不是 `.kiro/specs/`
4. CR 命名格式：`CRxx-{需求名}`（xx 为两位数序号）
5. 小型需求使用 `spec/minorCR/MinorCR-{N}-{需求名}/` 路径

## 域目录映射

| 域 | 目录路径 |
|---|---|
| 公共 | `CUST-AIDEV-BACKEND-DOCS/domain/common-center/spec/` |
| 用户中心 | `CUST-AIDEV-BACKEND-DOCS/domain/user-center/spec/` |
| 业主中心 | `CUST-AIDEV-BACKEND-DOCS/domain/owner-center/spec/` |
| 工单中心 | `CUST-AIDEV-BACKEND-DOCS/domain/workorder-center/spec/` |
| 缴费中心 | `CUST-AIDEV-BACKEND-DOCS/domain/billing-center/spec/` |
| 公告中心 | `CUST-AIDEV-BACKEND-DOCS/domain/notice-center/spec/` |
| 门禁中心 | `CUST-AIDEV-BACKEND-DOCS/domain/access-center/spec/` |
| 基础中心 | `CUST-AIDEV-BACKEND-DOCS/domain/base-center/spec/` |

## 文件结构

每个 CR 目录下的标准文件结构：

```
CRxx-{需求名}/
├── requirements_plan.md      # 需求计划
├── requirements.md           # 需求文档
├── design_plan.md            # 设计计划
├── design.md                 # 设计文档
├── tasks.md                  # 任务列表
├── sql/                      # SQL 脚本（如有）
└── api/                      # API 文档（如有）
```
