# 负面提示词策略

> 跨引擎通用的负面提示词使用方法论。

## 核心原则

1. **少即是多**：5-8 个最有效，超过 20 个画面变平
2. **针对性**：根据场景类型选择，不要无脑堆叠
3. **引擎差异**：不同引擎的负面提示词位置和格式不同

## 引擎差异

| 引擎 | 位置 | 格式 |
|------|------|------|
| KLING 3.0 | Master Prompt 末尾 | `Negative: xxx, xxx, xxx` |
| WAN 2.5 | 整体描述末尾或独立字段 | 视平台而定 |

## 分层策略

### 第一层：基础集（每次必加）

**KLING 3.0（6 个）：**
```
blur, distort, low quality, warping fingers, jittery eyes, character drift between shots
```

**WAN 2.5（待验证）：**
```
blur, distort, low quality, warping fingers, jittery eyes, character inconsistency
```

### 第二层：场景追加（选 2-3 个）

| 场景类型 | 追加项 |
|----------|--------|
| 有对话 | audio desync, garbled speech, mouth not matching words |
| 多角色 | face swap, character merge, identity drift |
| 电影感 | unnatural motion, stuttered movement, flickering highlights |
| 多镜头 | tonal shift between cuts, lighting inconsistency |
| 手部动作 | extra fingers, deformed hands, fused fingers |
| 面部特写 | asymmetric face, crossed eyes, double iris |
| 动作场景 | frozen pose, sliding feet, physics defying |

### 第三层：风格防护（可选）

| 要避免的风格 | 负面提示词 |
|-------------|-----------|
| 避免卡通感 | cartoon, anime style, cel shading |
| 避免过度 HDR | oversaturated, HDR artifacts, tone mapping |
| 避免 AI 感 | artificial, plastic, uncanny valley |
| 避免模糊 | soft focus, out of focus, motion blur |

## 已废弃的负面提示词

以下在 KLING 3.0 中已不需要（引擎已修复）：

| 废弃项 | 原因 |
|--------|------|
| frozen lips | KLING 3.0 已改善 |
| plastic skin | KLING 3.0 已改善 |

## 使用决策流程

```
1. 加入基础集（6个）
2. 判断场景类型 → 选择 2-3 个场景追加
3. 是否有特殊风格需求 → 选择 0-2 个风格防护
4. 总数控制在 8-12 个以内
```
