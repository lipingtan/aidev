# 产出文件位置规范

## 核心规则

| 文件类型 | 存放位置 |
|----------|----------|
| 过程文档（plan、design、迭代记录） | `AIDOC/{系列}/{NNN}-{单集}/` |
| 参考图片 | `AIDOC/{系列}/references/` |
| 最终提示词 | `output/{系列}/{引擎}/{NNN}-{单集}/{章节}/` |
| 生成的视频文件 | `output_video/{系列}/{引擎}/{NNN}-{单集}/{章节}/` |

**禁止**将任何产出文件生成到 `.kiro/specs/` 目录下。

## 过程文档路径

```
AIDOC/{系列}/{NNN}-{单集名}/
├── story_plan.md
├── synopsis.md
├── pacing-map.md
└── {章节名}/
    ├── chapter_plan.md
    └── shots/ep-{NNN}/
        ├── shot_plan.md
        ├── shot_design.md
        └── iterations/
```

## 结果文档路径

```
output/{系列}/
├── kling/                          # 默认只生成 KLING
│   └── {NNN}-{单集名}/
│       └── {章节名}/
│           ├── ep-001.md
│           └── ep-002.md
└── wan25/                          # 需用户要求时生成
    └── {NNN}-{单集名}/
        └── {章节名}/
            └── ep-001.md
```

## 视频文件路径

```
output_video/{系列}/
├── kling/
│   └── {NNN}-{单集名}/
│       └── {章节名}/
│           ├── ep-001.mp4
│           └── ep-002.mp4
└── wan25/
    └── {NNN}-{单集名}/
        └── {章节名}/
            └── ep-001.mp4
```

> `output_video/` 目录及其所有内容已被 `.gitignore` 排除，不纳入版本控制。

## 规则

1. `output/` 下的文件始终是最新可用版本
2. 只有 `shot_design.md` 被用户确认后，才能在 `output/` 中生成对应提示词
3. **生成 `output/` 子目录时，必须同步创建 `output_video/` 对应的子目录结构**
4. 迭代优化后更新 `output/` 中的文件，旧版本记录在 `iterations/` 中
5. 引擎目录（kling/wan25）在系列目录下，便于按系列管理所有引擎输出
6. 视频文件命名与提示词文件对应（如 `ep-001.md` → `ep-001.mp4`）
