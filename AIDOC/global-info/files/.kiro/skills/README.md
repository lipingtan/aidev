---
文档类型: Skills 目录说明
适用范围: CUST-AIDEV-BACKEND-DOCS
维护团队: AI 开发团队
最后更新: 2025-01
---

# Skills 目录

## Skills 目录用途

Skills 是为 AI 智能体注入领域专业能力的知识库。每个 skill 封装了特定领域的规则、检查清单和最佳实践，使 AI 在执行相关任务时能够自动应用这些专业知识，而无需在每次对话中重复说明。

**核心价值：**
- 将团队积累的最佳实践固化为可复用的 AI 能力
- 确保 AI 在执行代码审查、架构设计等任务时遵循统一标准
- 减少人工干预，提升 AI 辅助开发的质量和一致性

## 标准目录结构

每个 skill 以独立目录存放，结构如下：

```
skills/
├── README.md                    # 本文件，skills 目录总览
├── code-review/                 # 代码审查 skill
│   ├── SKILL.md                 # skill 主文件（必须）
│   ├── scripts/                 # 辅助脚本（可选）
│   ├── references/              # 参考资料、规范文档（可选）
│   └── assets/                  # 图片、附件等资源（可选）
└── [other-skill]/
    └── SKILL.md
```

**SKILL.md 是每个 skill 的核心文件**，包含：
- YAML frontmatter（name、description、version）
- 使用场景说明
- 核心规则和检查清单
- 代码示例

## 如何创建新 Skill

**步骤 1：创建目录**

在 `skills/` 下创建以 skill 名称命名的目录，使用小写字母和连字符：

```bash
mkdir skills/your-skill-name
```

**步骤 2：编写 SKILL.md**

在新目录下创建 `SKILL.md`，包含 YAML frontmatter 和正文内容：

```markdown
---
name: your-skill-name
description: 简要描述此 skill 的用途
version: 1.0.0
---

# Skill 标题

## 使用场景
...

## 核心规则
...

## 检查清单
...
```

**步骤 3：添加引用文件（可选）**

- `references/`：放置相关规范文档、标准链接
- `scripts/`：放置辅助检查脚本
- `assets/`：放置图表、截图等资源

## 已有 Skills 清单

| Skill 名称 | 用途 | 触发方式 |
|------------|------|----------|
| code-review | Java 后端代码审查，检查代码质量、安全性和规范符合度 | 提交 PR 前、完成开发任务后，手动调用或 AI 自动触发 |

## GDC 内网 Skills 资产库

团队可从 GDC 内网获取更多预置 skill，涵盖架构审查、性能分析、安全扫描等场景。

内网地址：_[GDC 内网 skills 资产库地址]_

> 从资产库获取的 skill 可直接放入本目录使用，建议根据项目实际情况进行适当调整。
