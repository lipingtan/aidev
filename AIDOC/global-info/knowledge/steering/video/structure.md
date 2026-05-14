# 项目结构规范

## 根目录结构

```
/
├── AIDOC/                          # 所有过程文档 + 参考图
│   ├── global-info/                # 全局信息（跨系列共享）
│   └── series/                     # 所有系列的容器
│       └── {系列名}/               # 单个系列
│
├── output/                         # 所有最终提示词（干净、即用）
│   └── {系列名}/
│       ├── assets/                 # 资产生成提示词（角色/场景/道具参考图）
│       └── kling/                  # 视频片段提示词
│
├── output_video/                   # 生成的视频文件（git ignored）
│   └── {系列名}/
│
├── .kiro/                          # 工作流配置
└── .gitignore
```

> `AIDOC/` 顶层只有 `global-info/` 和 `series/` 两个目录。

---

## AIDOC 目录详细结构

### 全局信息（跨系列共享 — 纯参考材料库）

```
AIDOC/global-info/
├── templates/                      # 通用模板库（可被系列继承）
│   ├── worldview/                  # 世界观模板（按类型分）
│   ├── character-archetypes/       # 角色原型模板
│   └── story-structures/           # 故事结构模板
│
├── references/                     # 原始参考素材（灵感、截图、文章）
│   ├── visual-styles/
│   ├── cinematography/
│   └── audio/
│
├── universe/                       # 跨系列世界观索引
├── characters/                     # 角色原型索引
├── scripts/                        # 未归属系列的原始创意素材
│
└── knowledge/                      # 制作知识库
    ├── engines/                    # 引擎技术知识
    ├── common/                     # 通用提示词技巧
    ├── cases/                      # 案例库
    └── steering/                   # 工作流规范
```

### 系列目录

```
AIDOC/series/{系列名}/
├── README.md                       # 系列概述（基本参数）
│
├── foundation/                     # 基础资产（策划层，Phase 0~3 产出）
│   ├── worldview/                  # 世界观设定
│   │   ├── worldview.md
│   │   └── timeline.md
│   ├── characters/                 # 角色策划设定（中文，完整背景）
│   │   └── {角色设定文件}.md
│   └── scripts/                    # 剧本/故事素材
│       ├── 大纲.md
│       └── 故事细化.md
│
└── production/                     # 制作资产 + 分集（Phase 4~5 + 制作阶段）
    ├── library/                    # 制作级视觉资产（英文提示词层）
    │   ├── style-bible.md
    │   ├── series-bible.md
    │   ├── production-constraints.md   # foundation→production 桥梁文件
    │   ├── elements-registry.md
    │   ├── props-registry.md
    │   ├── element-priority-guide.md
    │   ├── reference-generation-guide.md
    │   ├── cinematography-guide.md
    │   ├── audio-design.md
    │   ├── visual-direction.md
    │   ├── pacing-map-{集名}.md
    │   ├── characters/             # 角色视觉档案（英文，提示词专用）
    │   │   └── {角色名}.md
    │   └── scenes/                 # 场景视觉设定（英文，提示词专用）
    │       └── {场景名}.md
    │
    ├── references/                 # 生成的参考图片
    │   ├── characters/
    │   ├── scenes/
    │   ├── style/
    │   └── props/
    │
    ├── production-tracker.md       # 制作进度追踪
    │
    └── {NNN}-{单集名}/             # 分集制作目录
        ├── story_plan.md               # 集级策划（草案+澄清问题）
        ├── story_design.md             # 集级定稿（确认后的正式执行版）
        ├── synopsis.md                 # 本集概要（可选）
        ├── assembly.md                 # 组装顺序（后期用）
        └── {章节名}/
            ├── chapter_plan.md         # 章级策划（含澄清问题）
            ├── chapter_design.md       # 章级定稿（确认后的正式执行版）
            └── shots/
                └── ep-{NNN}/
                    ├── shot_plan.md
                    ├── shot_design.md
                    └── iterations/
```

---

## output 目录详细结构

```
output/
└── {系列名}/
    ├── assets/                     # 资产生成提示词（参考图用）
    │   ├── characters/             # 角色素体 + 服装参考图提示词
    │   │   ├── {角色名}-body.md
    │   │   ├── {角色名}-outfit-{场景}.md
    │   │   └── ...
    │   ├── scenes/                 # 场景参考图提示词
    │   │   └── {场景名}.md
    │   ├── props/                  # 道具参考图提示词
    │   │   └── {道具名}.md
    │   └── enemies/                # 敌方单位参考图提示词
    │       └── {单位名}.md
    │
    ├── kling/                      # KLING 视频片段提示词
    │   └── {NNN}-{单集}/{章节}/ep-{NNN}.md
    │
    └── wan25/                      # WAN 2.5（需用户要求时才创建）
        └── ...
```

**output 核心原则**：
- 所有文件都是"干净、最新版、可直接复制粘贴到生成工具使用"
- `assets/` = 给图片生成工具的提示词（KLING 图片 / Midjourney / Flux）
- `kling/` = 给 KLING 视频生成的提示词
- 不含任何设计说明或注释，纯提示词

---

## 资产生命周期

```
foundation/（策划层，中文）
    │
    ▼ 转化流程（Phase 4）
    │
production/library/（设计层，中英混合）
    │
    ▼ 提示词提取
    │
output/assets/（提示词层，纯英文，可直接使用）
    │
    ▼ 图片生成工具
    │
production/references/（生成的参考图 PNG）
    │
    ▼ 上传 KLING 平台
    │
production/library/elements-registry.md（填入平台标识）
```

---

## 片段（Shot）生命周期

```
production/library/（设计资产）
    │
    ▼ shot_plan → shot_design
    │
output/kling/{NNN}-{单集}/{章节}/ep-{NNN}.md（视频提示词）
    │
    ▼ KLING 生成视频
    │
output_video/（生成的视频文件）
    │
    ▼ 效果评估
    │
production/{NNN}-{单集}/{章节}/shots/ep-{NNN}/iterations/（迭代记录）
```

---

## 命名规范

| 类型 | 格式 | 示例 |
|------|------|------|
| 系列名 | kebab-case | `lost-crown`、`midnight-cafe` |
| 单集名 | `{NNN}-{名称}` | `001-prologue`、`002-act1` |
| 章节名 | `ch{NN}-{名称}` | `ch01-blood-and-starlight` |
| 片段编号 | `ep-{NNN}` | `ep-000`、`ep-001` |
| 迭代版本 | `v{N}` | `v1`、`v2` |
| 资产提示词 | `{对象名}.md` | `alan-body.md`、`royal-city-night.md` |

---

## 资产查找优先级

分集 > production/library/ > foundation/ > global-info/
