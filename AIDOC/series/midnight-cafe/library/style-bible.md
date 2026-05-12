# 午夜咖啡馆系列 — 视觉风格指南

## 风格锚定词
```
Cinematic 35mm film, warm color grading, shallow depth of field, intimate atmosphere
```

## 对话语言

| 配置项 | 设定 |
|--------|------|
| 对话语言 | 中文 |
| 提示词描述语言 | 英文 |

> 如某段需要使用其他语言（如英文对话），在该段 `shot_plan.md` 中单独标注覆盖。

## 色彩方案
- 主色调：warm amber, deep brown, soft cream
- 点缀色：golden light
- 对比色：cool blue night（仅窗外）

## 全局负面提示词

> 写在 Master Prompt 末尾，用 `Negative:` 前缀。控制在 5~8 个。

基础集（每段必加）：
```
Negative: blur, distort, low quality, warping fingers, jittery eyes, character drift between shots
```

有对话时追加：`audio desync, mouth not matching words`
电影感追加：`unnatural motion, flickering highlights`
