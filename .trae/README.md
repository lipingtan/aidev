# .trae 工作流配置目录

本目录为 Trae IDE 的工作流配置，内容从 `.kiro/` 转换而来。

## 目录结构

| 目录/文件 | 对应 .kiro | 说明 |
|-----------|-----------|------|
| `rules/core.md` | `.kiro/steering/core.md` | 核心规范（始终生效） |
| `rules/hooks-as-rules.md` | `.kiro/hooks/*.kiro.hook` | hooks 转换为内联规则 |
| `agents/project_manager_agent.json` | `.kiro/agents/project-manager-agent.md` | 项目管理专家代理 |
| `mcp.json` | `.kiro/settings/mcp.json` | MCP 服务器配置 |
| `skills/` | `.kiro/skills/` | 技能库（内容一致） |

## 转换说明

- **steering → rules**：Trae 使用 `rules/` 目录存放上下文规范，等价于 Kiro 的 `steering/`
- **hooks → rules**：Trae 无原生 hooks 机制，所有 hook 逻辑已内联到 `rules/hooks-as-rules.md` 作为 AI 行为约束
- **agents**：Trae 使用 JSON 格式定义 agent，已从 Kiro 的 Markdown frontmatter 格式转换
- **skills**：内容完全一致，直接复用
- **specs**：Kiro 专用工作流内部文件，无需迁移
