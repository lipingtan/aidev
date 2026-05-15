# 项目概览

本工作区是 AI 短剧/电影制作工作台，核心产出物是适配不同视频引擎的结构化提示词。

## 工作区结构

```
/
├── AIDOC/                      ← 📋 所有过程文档（策划、设计、确认、参考图）
│   ├── global-info/            ← 全局信息（跨系列共享）
│   │   └── knowledge/          #   制作知识库 + 规范
│   │       └── steering/video/ #   视频制作规范
│   │
│   └── series/                 ← 所有系列的容器
│       └── {系列名}/           ← 按系列组织
│           ├── foundation/     #   策划层（世界观、角色、剧本）
│           ├── production/     #   制作层（library + 分集目录）
│           └── README.md
│
├── output/                     ← 🎬 所有最终提示词（干净、即用）
│   └── {系列名}/               #   按系列组织
│       ├── assets/             #   资产生成提示词（参考图用）
│       ├── kling/              #   KLING AI（默认）
│       │   └── {NNN}-{单集}/{章节}/ep-{NNN}.md
│       └── wan25/              #   WAN 2.5（按需）
│
├── output_video/               ← 生成的视频文件（git ignored）
│   └── {系列名}/
│
├── .kiro/                      ← ⚙️ 工作流配置
└── .gitignore
```

## 核心分离原则

| 目录 | 内容 | 特点 |
|------|------|------|
| `AIDOC/` | 过程文档 + 参考图 | 完整保留历史，可回溯，AI 协作工作台 |
| `output/` | 最终提示词 | 干净、最新版、可直接粘贴到引擎使用 |

## 全局约定

- 过程文档全部在 `AIDOC/series/` 下，结果文档全部在 `output/` 下
- 项目按 `AIDOC/series/{系列名}/production/{NNN}-{单集名}/` 组织，单集编号从 001 开始
- `output/` 目录结构：`output/{系列}/{引擎}/{NNN}-{单集}/{章节}/ep-{NNN}.md`
- 每个 5~15 秒片段有完整的 plan → 确认 → design → 确认 → 生成提示词流程
- 资产查找优先级：单集 > 系列 library/ > 全局 global-info/
- 系列名和单集名使用 kebab-case 格式
- 默认视频引擎为 KLING AI，**仅生成 KLING 提示词**，WAN 2.5 需用户明确要求时才生成
- 文档使用中文，提示词内容使用英文
- 参考图片放在 `AIDOC/series/{系列名}/production/references/` 下
- KLING 元素绑定标识记录在 `AIDOC/series/{系列名}/production/library/elements-registry.md`
- 有元素标识的用 `@标识` 引用，无标识的用文字描述（见 `element-readiness.md`）
