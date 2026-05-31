# 提示词工程规范

## 核心原则

### 1. 场景先行（Scene-First）

KLING 和 WAN 都推荐先建立环境上下文，再描述主体和动作。这给模型提供了空间和光线的计算基础。

```
✅ 好的顺序：环境 → 主体 → 动作 → 镜头 → 光线 → 氛围
❌ 差的顺序：动作 → 主体 → 环境（模型缺乏空间上下文）
```

### 2. 具体化（Specificity）

用具体的视觉描述替代抽象概念：

| ❌ 抽象 | ✅ 具体 |
|---------|---------|
| beautiful lighting | golden hour sunlight streaming through venetian blinds |
| sad expression | eyes downcast, slight furrow between brows, lips pressed thin |
| walking slowly | measured deliberate steps, shoulders slightly hunched |
| nice outfit | tailored navy wool coat, white silk scarf, leather gloves |

### 3. 物理真实（Physical Realism）

描述动作时考虑物理规律：

```
✅ "gravity-affected smoke drifting upward from the cup"
✅ "wind-blown hair sweeping across her face"
❌ "hair floating magically"（除非是奇幻风格）
```

### 4. 简洁有力（Concise Power）

- 每个镜头提示词控制在 2~4 句话
- 避免重复 Master Prompt 中已有的信息
- 对话台词越短越好

---

## KLING AI 提示词编写指南

> **重要**：KLING 多镜头模式中，每个镜头各自一个完整提示词，没有独立的 Master Prompt。
> 下方的"Master Prompt"概念指的是**每个镜头提示词中都要包含的风格/场景/角色基础描述**，
> 不是一个单独的输入框。

### KLING 提示词要素顺序（官方推荐）

> 严格按以下顺序组织，这是 KLING 物理引擎的计算逻辑。

```
Subject（主体细节）→ Movement（运动物理）→ Scene（场景背景）
→ Cinematic Language（镜头语言）→ Lighting（光线）→ Atmosphere（氛围）
```

| 要素 | 作用 | 示例术语 |
|------|------|----------|
| **Subject** | 定义主体的物理属性和身份 | @alan_body @alan_w3, dragon scale tattoo on left arm |
| **Movement** | 定义运动的物理规律 | gravity-affected smoke, wind-blown flames |
| **Scene** | 建立空间上下文 | dark medieval alley, cobblestone wet with rain |
| **Cinematic Language** | 控制景别、视角、运镜 | close-up, slow push in |
| **Lighting** | 定义光线交互方式 | volumetric moonlight, warm lantern glow |
| **Atmosphere** | 建立情绪基调 | tense, foreboding silence |

### 每镜头自包含提示词模板

```
[Subject: 角色视觉特征（@元素ID 或文字）].
[Movement: 主体动作描述].
[Scene: 环境描述].
[Cinematic Language: 景别 + 运镜].
[Lighting: 光线].
[Atmosphere: 情绪/氛围].
Negative: [负面提示词]
(Duration: Xs)
```

**示例：**
```
Cinematic 35mm film, warm color grading. A dimly lit vintage café at night, 
rain streaking down the windows. @cafe_woman_body @cream_sweater sits alone 
at a corner table.
Close-up, slow push in: She traces the rim of her coffee cup with one finger, 
steam rising in the warm lamplight. Her eyes are unfocused, lost in thought.
Negative: blur, distort, low quality, warping fingers, jittery eyes, character drift between shots
(Duration: 4 seconds)
```

### 多镜头衔接技巧

1. **动作衔接**：上一镜头的结束动作 = 下一镜头的开始状态
2. **视线衔接**：保持视线方向一致（180度法则）
3. **情绪衔接**：情绪变化要有过渡，不要突变
4. **光线衔接**：同一场景内光线保持一致

---

## WAN 2.5 提示词编写指南

### 整体描述模板

```
[叙事概述，说明这是什么故事/场景]. [风格设定].
Shot 1 [0s-Xs]: [景别]: [具体画面描述，包含主体、动作、环境细节].
Shot 2 [Xs-Ys]: [景别]: [具体画面描述].
...
```

**示例：**
```
A tense emotional confrontation in a rain-soaked café at night. 
Cinematic warm tones, shallow depth of field, 35mm film grain.
Shot 1 [0-3s]: Wide shot: Rain-streaked café window, warm amber light inside, 
a woman sits alone at corner table, steam rising from untouched coffee.
Shot 2 [3-6s]: Medium shot: She looks up as the door opens, rain sound intensifies, 
her expression shifts from distant to alert.
Shot 3 [6-9s]: Close-up: Her eyes, reflecting the neon signs outside, 
a single tear forming but not falling.
```

### WAN 特有注意事项

1. **时间戳必须连续**：`[0-3s]` → `[3-6s]` → `[6-9s]`，不能有间隔
2. **单一场景**：WAN 对单一场景的理解更好，避免场景跳转
3. **光线锚点**：明确指定一个主光源
4. **动作简洁**：每个 Shot 只描述一个主要动作

---

## 风格关键词库

### 画面风格

| 风格 | 关键词 |
|------|--------|
| 电影感 | cinematic, 35mm film, anamorphic lens, shallow depth of field |
| 纪录片 | documentary style, handheld, natural light, raw footage |
| 复古 | vintage, film grain, faded colors, retro color grading |
| 赛博朋克 | neon noir, cyberpunk, holographic, rain-soaked streets |
| 温暖治愈 | warm tones, soft focus, golden hour, cozy atmosphere |
| 冷峻悬疑 | desaturated, cold blue tones, harsh shadows, noir lighting |
| 动漫风 | anime style, cel shading, vibrant colors, dynamic angles |

### 光线类型

| 类型 | 关键词 |
|------|--------|
| 自然光 | natural sunlight, overcast diffused light, golden hour |
| 人工光 | neon lights, fluorescent, warm lamplight, candlelight |
| 戏剧光 | volumetric light, rim lighting, chiaroscuro, Tyndall effect |
| 环境光 | ambient glow, reflected light, bounced light |

### 情绪氛围

| 情绪 | 关键词 |
|------|--------|
| 温馨 | warm, intimate, cozy, gentle, tender |
| 紧张 | tense, suspenseful, uneasy, claustrophobic |
| 忧伤 | melancholic, somber, wistful, bittersweet |
| 欢快 | joyful, vibrant, energetic, playful |
| 神秘 | mysterious, ethereal, dreamlike, surreal |
| 史诗 | epic, grand, majestic, awe-inspiring |

---

## 常见错误和修正

| 错误 | 问题 | 修正 |
|------|------|------|
| 描述太长 | 模型抓不住重点 | 每镜头 2~4 句，突出核心动作 |
| 多个动作堆叠 | 动作混乱或只执行第一个 | 每镜头只描述一个主要动作 |
| 抽象情绪词 | 模型无法视觉化 | 用具体的面部/肢体表现替代 |
| 忽略物理 | 动作不自然 | 加入重力、惯性等物理描述 |
| 风格不一致 | 各镜头画面割裂 | Master Prompt 统一风格锚定 |
| 角色描述变化 | 角色外貌不一致 | 固定使用 characters.md 中的关键词 |

---

## shot_design → 最终提示词 组装流程

> 当 shot_design 被用户确认后，按以下步骤组装最终提示词写入 output/。

### KLING AI 组装步骤

> KLING 多镜头模式：每个镜头一个完整自包含提示词，无独立 Master Prompt。

```
对于 chapter_design 中的每个镜头：

步骤 1：组装镜头提示词
    ├── 风格锚定词（从 style-bible.md）
    ├── 场景描述（从 shot_design 场景视觉）
    ├── 角色描述（@元素ID 或文字，从 shot_design 元素绑定）
    ├── 景别+运镜+动作（从 shot_design 各镜头设计表）
    ├── 光线+氛围（从 shot_design）
    ├── 负面提示词（基础集 + 场景追加）
    └── Duration

步骤 2：附加元数据
    ├── 元素绑定 ID 列表
    └── 对话音频指令（如有）

步骤 3：写入 output/
    └── output/{系列}/kling/{NNN}-{单集}/{章节}/ep-{NNN}.md
```

### 最终提示词文件格式（KLING 多镜头）

```markdown
Shot 1 (Duration: Xs):
{风格锚定}, {场景描述}. {角色: @ID 或文字描述}.
{景别, 运镜}: {动作描述}. {光线}. {对话（如有）}.
Negative: {负面提示词}
Elements: {@id1, @id2, ...}

Shot 2 (Duration: Ys):
{风格锚定}, {场景描述}. {角色: @ID 或文字描述}.
{景别, 运镜}: {动作描述}. {光线}.
Negative: {负面提示词}
Elements: {@id1, @id2, ...}

...
```

### WAN 2.5 组装步骤

```
步骤 1：组装整体描述
    ├── 从 shot_design 提取叙事概述（一句话）
    ├── 从 style-bible.md 提取风格设定
    └── 组合为：[叙事概述]. [风格设定].

步骤 2：组装 Shot 时间戳
    ├── 从 shot_design 提取时长
    ├── 计算时间戳区间 [0s-Xs]
    ├── 从 shot_design 提取景别+动作+环境
    └── 组合为：Shot 1 [0s-Xs]: [景别]: [画面描述].

步骤 3：写入 output/
    └── output/{系列}/wan25/{NNN}-{单集}/{章节}/ep-{NNN}.md
```

### 最终提示词文件格式（KLING）

> 见上方"KLING AI 组装步骤"中的格式。每个镜头自包含所有信息。

### 质量检查（写入前）

- [ ] 风格锚定词与 style-bible 一致
- [ ] 角色描述与 characters/*.md 完全一致（逐词对照）
- [ ] 无中文残留（output/ 中纯英文）
- [ ] 无注释/说明文字（纯提示词）
- [ ] 时长与 pacing-map 一致
- [ ] 元素绑定 ID 与 elements-registry 一致
