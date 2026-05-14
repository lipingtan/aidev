---
inclusion: always
---

# Knowledge Index

所有沟通、文档、代码注释使用中文。

## 文档输出位置

| 类型 | 路径 |
|------|------|
| CR 需求 | `CUST-AIDEV-BACKEND-DOCS/domain/{域}-center/spec/normalCR/CRxx-{名}/` |
| 小型 CR | `CUST-AIDEV-BACKEND-DOCS/domain/{域}-center/spec/minorCR/MinorCR-{N}-{名}/` |
| Bugfix | `CUST-AIDEV-BACKEND-DOCS/domain/{域}-center/vibe/Fix-{NNN}-{简述}/` |
| 变更日志 | `domain/{域}-center/change-log/{YYYY-MM}/CL-YYYY-NNN.md` |

禁止输出到 `.kiro/specs/`。

## 知识库

所有项目规范和上下文信息位于 `CUST-AIDEV-BACKEND-DOCS/knowledge/steering/`，接到任务时按需读取相关文件：

| 分类 | 文件 | 何时读取 |
|------|------|----------|
| 项目上下文 | `project-context.md` | 首次接触项目或需要了解域/架构时 |
| 项目结构 | `structure.md` | 需要了解目录结构或文件位置时 |
| 开发流程 | `development-workflow.md` | 执行 CR 需求/设计/任务流程时 |
| 任务表示 | `task-representation.md` | 拆解或执行 Task 时 |
| Bugfix 流程 | `development-workflow.md` §5 | 处理 Bug 修复时 |
| 回归防护 | `regression-guard.md` | 设计阶段编写不变行为清单时 |
| Java 规范 | `java-conventions.md`, `java-standards.md` | 编写/审查 Java 代码时 |
| Spring Boot | `spring-boot-patterns.md` | 编写/审查后端代码时 |
| 安全规范 | `security-best-practices.md` | 涉及安全相关代码时 |
| 代码审查 | `code-review.md` | 执行代码审查时 |
| 变更日志 | `change-log.md` | 代码变更完成后记录时 |
| Git 规范 | `git-best-practices.md` | 提交/推送代码时 |
| SonarQube | `sonarqube-quality.md` | 关注代码质量时 |
| 调试方法 | `debugging-methodology.md` | 排查问题时 |
| 日志分析 | `log-analyzer.md` | 分析日志时 |

## 全局信息

项目详细信息位于 `CUST-AIDEV-BACKEND-DOCS/global-info/`，按需读取。

## 原则

1. 文档先行：重要变更前先查阅或更新相关文档
2. 变更必记录：代码变更完成后必须创建变更日志
3. 不自行决策：关键决策用 [Question]/[Answer] 向用户澄清
