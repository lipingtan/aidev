# 产出文件位置规范

## 核心规则

| 文件类型 | 存放位置 | 生成时机 |
|----------|----------|----------|
| 过程文档（plan、design、迭代记录） | `AIDOC/series/{系列}/production/` | 制作各阶段 |
| **元素参考图提示词**（素体/服装/场景/道具） | `output/{系列}/assets/{类型}/` | **元素就绪检查时**（早于视频片段） |
| 视频片段提示词 | `output/{系列}/kling/{NNN}-{单集}/{章节}/` | shot_design 确认后 |
| 生成的参考图片 | `AIDOC/series/{系列}/production/references/` | 用提示词在 KLING 生成后 |
| 生成的视频文件 | `output_video/{系列}/kling/{NNN}-{单集}/{章节}/` | 视频生成后 |

**禁止**将任何产出文件生成到 `.kiro/specs/` 目录下。

---

## output/ 完整结构

```
output/{系列}/
│
├── assets/                         # ① 元素参考图提示词（先生成）
│   ├── characters/                 # 角色素体 + 服装
│   │   ├── alan_body.md            # 素体提示词
│   │   ├── alan_w1.md              # 服装提示词（W1）
│   │   ├── alan_w3.md
│   │   ├── avira_16yr_body.md
│   │   └── ...
│   ├── scenes/                     # 场景参考图提示词
│   │   ├── royal_city_night.md
│   │   ├── palace_chamber.md
│   │   └── ...
│   ├── props/                      # 武器/道具提示词
│   │   ├── alan_sword.md
│   │   ├── lost_crown.md
│   │   └── ...
│   ├── fx/                         # 魔法效果提示词
│   │   ├── fx_dragon_tattoo_glow.md
│   │   └── ...
│   └── enemies/                    # 敌方单位提示词
│       ├── enemy_mist_beast.md
│       └── ...
│
├── kling/                          # ② 视频片段提示词（后生成）
│   └── {NNN}-{单集名}/
│       └── {章节名}/
│           ├── ep-001.md
│           └── ep-002.md
│
└── wan25/                          # 需用户要求时生成
    └── {NNN}-{单集名}/
        └── {章节名}/
            └── ep-001.md
```

**两类提示词的区别：**

| | assets/ | kling/ |
|--|---------|--------|
| 用途 | 在 KLING 图片生成/Midjourney 生成参考图 | 在 KLING 视频生成 |
| 生成时机 | 元素就绪检查时（chapter_design 确认后） | shot_design 确认后 |
| 前置条件 | 无（只需 library/ 中的角色/场景描述） | 需要元素已注册（有 @标识） |
| 文件命名 | `{标识}.md`（与 elements-registry 对应） | `ep-{NNN}.md` |

---

## 过程文档路径

```
AIDOC/series/{系列名}/production/{NNN}-{单集名}/
├── story_plan.md
├── story_design.md
├── synopsis.md
└── {章节名}/
    ├── chapter_plan.md
    ├── chapter_design.md
    └── shots/ep-{NNN}/
        ├── shot_plan.md
        ├── shot_design.md
        └── iterations/
```

## 视频文件路径

```
output_video/{系列}/kling/{NNN}-{单集名}/{章节名}/
├── ep-001.mp4
└── ep-002.mp4
```

> `output_video/` 已被 `.gitignore` 排除，不纳入版本控制。

---

## 规则

1. `assets/` 提示词在元素就绪检查时生成，**早于任何视频片段提示词**
2. 同一元素的提示词文件只生成一次，已存在则不重复生成
3. `kling/` 提示词只有 `shot_design.md` 被用户确认后才能生成
4. 生成 `kling/` 子目录时，必须同步创建 `output_video/` 对应目录结构
5. 迭代优化后更新 `kling/` 中的文件，旧版本记录在 `iterations/` 中
