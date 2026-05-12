# KLING AI 3.0 提示词格式规范

## 多镜头提示词结构

```
Master Prompt: [整体场景描述, 角色, 情绪, 视觉风格]
Multi shot Prompt 1: [镜头1内容] (0s-Xs, Duration: X seconds)
Multi shot Prompt 2: [镜头2内容] (Xs-Ys, Duration: Y-X seconds)
...
Multi shot Prompt 6: [镜头6内容] (As-Bs, Duration: B-A seconds)
```

## 单镜头提示词要素（按顺序）

1. **Subject（主体）**：角色外貌、服装、表情、姿态
2. **Movement（动作）**：角色动作、物理运动
3. **Scene（场景）**：环境、背景、空间关系
4. **Cinematic Language（镜头语言）**：景别、运镜、视角
5. **Lighting（光线）**：光源、光质、光影效果
6. **Atmosphere（氛围）**：情绪基调、环境效果、色调

## Master Prompt 模板

```
[风格锚定], [环境描述], [角色A: 2~3个视觉特征], [角色B: 2~3个视觉特征（如有）]. [情绪/氛围]. [色调].

Negative prompt: [负面提示词]
```

**示例：**
```
Cinematic 35mm film, warm color grading. A dimly lit vintage café at night, 
rain streaking down the windows. A woman in her late 20s with short black hair, 
wearing an oversized cream knit sweater, sits alone at a corner table. 
Melancholic, intimate atmosphere. Palette: warm amber, deep brown, soft cream.

Negative prompt: blurry, low quality, watermark, jittery eyes, warping fingers, 
character drift, frozen lips.
```

## Shot Prompt 模板

```
[景别 + 运镜]: [主体动作描述]. [环境细节/光线变化]. [对话（如有）].
(Duration: X seconds)
```

**示例：**
```
Close-up, slow push in: She traces the rim of her coffee cup with one finger, 
steam rising in the warm lamplight. Her eyes are unfocused, lost in thought.
(Duration: 4 seconds)
```

## 对话格式

### 单人对话
```
[角色A: 身份, 语气描述]: "台词内容"
```

### 多人对话
```
[角色A: 身份, 语气描述]: "第一句台词"
Immediately, [角色B: 身份, 情绪语气]: "回应台词"
```

## 负面提示词规范

### 位置
直接写在 Master Prompt 末尾，用 `Negative:` 或 `Negative prompt:` 前缀。

### 数量控制
5~8 个，超过 20 个会让画面变平。

### 基础集（每次必加，6 个）
```
blur, distort, low quality, warping fingers, jittery eyes, character drift between shots
```

### 按场景追加（选 2~3 个）

| 场景类型 | 追加项 |
|----------|--------|
| 有对话 | audio desync, garbled speech, mouth not matching words |
| 多角色 | face swap, character merge, identity drift |
| 电影感 | unnatural motion, stuttered movement, flickering highlights |
| 多镜头 | tonal shift between cuts, lighting inconsistency |
