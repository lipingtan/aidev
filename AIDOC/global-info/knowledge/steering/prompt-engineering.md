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

### Master Prompt 模板

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

### Shot Prompt 模板

```
[景别 + 运镜]: [主体动作描述]. [环境细节/光线变化]. [对话（如有）].
```

**示例：**
```
Close-up, slow push in: She traces the rim of her coffee cup with one finger, 
steam rising in the warm lamplight. Her eyes are unfocused, lost in thought.
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
