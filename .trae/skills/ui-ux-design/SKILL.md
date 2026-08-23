---
name: ui-ux-pro-max
description: >-
  UI/UX 设计智能。当用户需要构建页面、设计界面、选择配色/字体/风格、审查 UI 代码时激活。
  触发词："做一个页面" "landing page" "dashboard" "配色" "UI风格" "设计系统"
  "build a page" "design" "create UI" "review UI"
---

# UI/UX Pro Max - 设计智能

完整的 Web 和移动端设计指南。包含 67+ 种风格、161 套配色、57 组字体搭配、161 种产品类型推理规则、99 条 UX 准则、25 种图表类型，覆盖 15 个技术栈。

## 核心文件引用

- **完整 Skill 内容**: `.claude/skills/ui-ux-pro-max/SKILL.md`
- **搜索脚本**: `src/ui-ux-pro-max/scripts/search.py`
- **数据文件**: `src/ui-ux-pro-max/data/` (CSV 数据库)
- **模板**: `src/ui-ux-pro-max/templates/`
- **Quick Reference**: `src/ui-ux-pro-max/templates/base/quick-reference.md`
- **Skill 主体内容**: `src/ui-ux-pro-max/templates/base/skill-content.md`

## 搜索命令

```bash
python3 .kiro/skills/ui-ux-design/src/ui-ux-pro-max/scripts/search.py "<query>" --domain <domain> [-n <max_results>]
```

**生成设计系统（必须先执行）:**
```bash
python3 .kiro/skills/ui-ux-design/src/ui-ux-pro-max/scripts/search.py "<product_type> <industry> <keywords>" --design-system [-p "Project Name"]
```

**可用 Domain:**
| Domain | 用途 |
|--------|------|
| `product` | 产品类型推荐（SaaS、电商、作品集等） |
| `style` | UI 风格（glassmorphism、minimalism 等）+ AI 提示词和 CSS 关键词 |
| `typography` | 字体搭配 + Google Fonts 导入 |
| `color` | 按产品类型的配色方案 |
| `landing` | 页面结构和 CTA 策略 |
| `chart` | 图表类型和库推荐 |
| `ux` | 最佳实践和反模式 |

**可用 Stack:**
```bash
python3 .kiro/skills/ui-ux-design/src/ui-ux-pro-max/scripts/search.py "<query>" --stack <stack>
```
支持: `html-tailwind`(默认), `react`, `nextjs`, `astro`, `vue`, `nuxtjs`, `nuxt-ui`, `svelte`, `swiftui`, `react-native`, `flutter`, `shadcn`, `jetpack-compose`, `angular`, `laravel`

## 工作流程

详见 `src/ui-ux-pro-max/templates/base/skill-content.md`，核心步骤：

1. **分析需求** — 提取产品类型、目标受众、风格关键词、技术栈
2. **生成设计系统**（必须） — 运行 `--design-system` 获取完整推荐
3. **补充细节搜索** — 按需用 `--domain` 深入查询
4. **技术栈指南** — 用 `--stack` 获取实现细节
5. **实现代码** — 综合设计系统 + 搜索结果输出代码
6. **交付前检查** — 执行 Quick Reference 检查清单

## Quick Reference 规则优先级

详见 `src/ui-ux-pro-max/templates/base/quick-reference.md`，按优先级：

| 优先级 | 类别 | 影响 |
|--------|------|------|
| 1 | Accessibility | CRITICAL |
| 2 | Touch & Interaction | CRITICAL |
| 3 | Performance | HIGH |
| 4 | Style Selection | HIGH |
| 5 | Layout & Responsive | HIGH |
| 6 | Typography & Color | MEDIUM |
| 7 | Animation | MEDIUM |
| 8 | Forms & Feedback | MEDIUM |
| 9 | Navigation Patterns | HIGH |
| 10 | Charts & Data | LOW |

## 子 Skill（高级）

本仓库还包含多个专项子 Skill，位于 `.claude/skills/`：

| 子 Skill | 用途 |
|----------|------|
| `banner-design` | Banner/海报设计 |
| `brand` | 品牌设计系统 |
| `design` | 通用设计 |
| `design-system` | 设计系统生成 |
| `slides` | 幻灯片设计 |
| `ui-styling` | UI 样式化 |
