# 元素就绪检查规范

> 在进入逐镜头制作（shot_plan/shot_design）之前，必须确认该章节涉及的所有关键元素已在 KLING 平台创建并注册了标识。

---

## 使用时机

| 步骤 | 动作 |
|------|------|
| **chapter_design 确认后** | 自动生成本章元素记录（填入标识），输出就绪报告 |
| **shot_design 编写时** | 根据注册状态决定用 `@标识` 还是文字描述 |
| **提示词组装时** | 按本文件的"提示词引用格式"输出 |

---

## 零、元素标识命名规范

所有标识使用 snake_case 英文，规则如下：

### 角色素体

```
{角色英文简写}_body              # 单一年龄版本
{角色英文简写}_{年龄}yr_body     # 多年龄版本
```

示例：`alan_body`、`avira_16yr_body`、`avira_19yr_body`、`tom_10yr_body`

### 服装

```
{角色英文简写}_w{N}              # 按衣橱编号 W1~W5
```

示例：`alan_w1`、`alan_w2`、`avira_w1`

### 武器/道具

```
{角色英文简写}_{武器描述}        # 主武器用描述性名称
{角色英文简写}_item{N}           # 次要道具用编号
```

示例：`alan_sword`、`gray_holy_sword`、`gray_holy_sword_awakened`

### 场景

```
{场景描述性名称}                 # 全小写 snake_case，不带角色前缀
```

示例：`royal_city_night`、`palace_chamber`、`seal_tower_ext`

### 魔法效果/特效

```
fx_{效果描述}
```

示例：`fx_dragon_tattoo_glow`、`fx_starlight_hands`、`fx_black_mist`

### 元素注册表列格式约定

所有 elements-registry.md 中的表格统一格式：

- **第一列为注册状态列**，列名 `注册`
- 空白 = 尚未在 KLING 平台创建
- 填 `Y` = 已在平台创建，可在提示词中用 `@标识` 引用
- 用户在 KLING 平台完成创建后，手动在对应行第一列填入 `Y`

```
| 注册 | 元素名称 | 标识 | 参考图 | 使用范围 |
|------|----------|------|--------|----------|
| Y    | 艾伦·素体 | `alan_body` | alan-body.png | 全系列 |
|      | 艾薇儿·素体（16岁） | `avira_16yr_body` | avira-16-body.png | 序幕~第三幕 |
```

龙纹纹身是角色身体的一部分，**不单独创建元素**，分两种状态处理：

| 状态 | 处理方式 |
|------|----------|
| 平时（隐藏/不发光） | 内嵌到素体参考图中，素体描述包含纹身外观 |
| 激活发光 | 用 `fx_dragon_tattoo_glow` 特效元素 + 提示词强化描述 |

提示词写法：
```
# 激活状态
@alan_body @alan_w3, dragon scale tattoo on left arm glowing gold-black,
veins of molten light spreading from the tattoo up to shoulder,
@fx_dragon_tattoo_glow

# 平时状态（素体参考图本身带纹身，无需额外描述）
@alan_body @alan_w1
```

---

## 一、元素就绪检查流程

### 触发条件

chapter_design 确认后、开始第一个 shot_plan 之前。

### 执行步骤

```
步骤 1：从 chapter_design 的"正式片段表"中提取本章所有涉及的：
        - 角色（含服装变体编号）
        - 场景
        - 武器/道具
        - 魔法效果（高频出现时）

步骤 2：对照 elements-registry.md，逐项检查第一列注册状态：
        - 标记为 Y = 已注册，可直接用 @标识
        - 空白 = 未注册，shot_design 中用文字描述替代

步骤 3：对于注册列为空白的元素，检查 output/{系列}/assets/ 下是否已有对应提示词文件：
        - 已有 → 跳过（提示用户该文件已存在，可直接使用）
        - 没有 → 按 reference-generation-guide.md 的模板生成提示词，
                  写入 output/{系列}/assets/{类型}/{标识}.md

步骤 4：对于注册表中尚无记录的元素（新元素），
        按命名规范自动生成标识并追加到 elements-registry.md，
        第一列留空（待用户手动填 Y）

步骤 5：输出就绪报告
```

> **步骤 3 的输出位置**：
> - 角色素体/服装 → `output/{系列}/assets/characters/{标识}.md`
> - 场景 → `output/{系列}/assets/scenes/{标识}.md`
> - 武器/道具 → `output/{系列}/assets/props/{标识}.md`
> - 魔法效果 → `output/{系列}/assets/fx/{标识}.md`
> - 敌方单位 → `output/{系列}/assets/enemies/{标识}.md`

### 就绪报告模板

```markdown
## 元素就绪检查：{章节名}

### 已注册（可直接用 @标识）

| 注册 | 元素 | 标识 | 类型 |
|------|------|------|------|
| Y | 艾伦·素体 | @alan_body | 角色 |
| Y | 艾伦·W3 斗殴后 | @alan_w3 | 服装 |
| Y | 王都街巷（夜） | @royal_city_night | 场景 |

### 未注册（需在 KLING 平台创建后填 Y）

| 元素 | 标识（已预填） | 优先级 | 当前处理方式 |
|------|--------------|--------|------------|
| 酒馆外部 | @tavern_ext | 高（出现3次） | 文字描述，建议尽快创建 |
| 乌鸦 | — | 低（单次） | 文字描述即可，无需创建 |

### 行动项

- [ ] 在 KLING 平台创建以上"未注册"中优先级为高/必须的元素
- [ ] 创建完成后在 elements-registry.md 对应行第一列填入 `Y`
- [ ] 标记完成后即可开始 shot_plan（不必等所有元素就绪）
```

---

## 二、元素优先级判定

| 优先级 | 条件 | 处理方式 |
|--------|------|----------|
| **必须创建** | 主角素体 + 当前章节服装 | 必须有平台标识才能开始制作 |
| **强烈建议** | 出现 ≥3 次的场景 | 创建后一致性显著提升 |
| **建议创建** | 重要配角素体 + 服装 | 有标识更好，无标识可用详细文字 |
| **可选** | 出现 1-2 次的场景/道具 | 文字描述即可 |
| **不需要** | 单次出现的背景元素、天气、光效 | 纯文字描述 |

---

## 三、提示词中的元素引用格式

### 有平台标识的元素

在提示词中使用 `@标识` 引用，KLING 会自动绑定参考图保持一致性：

```
@alan_body @alan_torn_outfit, standing in @city_alley_night,
blood on the corner of his mouth, snow falling on shoulders...
```

### 无平台标识的元素

使用详细文字描述替代，需要包含足够的视觉特征：

```
A black crow with glossy feathers flies across the full moon,
wings spread wide against the pale moonlight...
```

### 混合使用规则

同一段提示词中可以混合使用：

```
@alan_body @alan_torn_outfit walks through @city_alley_night,
a stray cat (orange tabby, thin, wet fur) watches from a windowsill...
```

- `@标识`：角色、服装、主要场景（有参考图保证一致性）
- 文字描述：临时出现的动物、天气效果、背景路人等

---

## 四、shot_design 中的元素绑定字段写法

### 有标识时

```markdown
### 元素绑定
| 元素 | 标识 | 用途 |
|------|------|------|
| 艾伦·素体 | @alan_body | 角色面部/体型 |
| 艾伦·斗殴后 | @alan_torn | 本段服装 |
| 王都街巷（夜） | @city_alley | 场景一致性 |
```

### 无标识时

```markdown
### 元素绑定
| 元素 | 标识 | 用途 |
|------|------|------|
| 艾伦·素体 | @alan_body | 角色面部/体型 |
| 酒馆外部 | （文字描述） | 场景 |

### 文字描述补充（无标识元素）
- 酒馆外部：A weathered wooden tavern with warm amber light spilling from the doorway, hanging iron lantern, snow-covered roof tiles, cobblestone street
```

---

## 五、与其他规范的关系

| 规范 | 关系 |
|------|------|
| `elements-registry.md`（系列 library 下） | 标识的唯一真相源，本规范检查其"KLING 平台标识"列 |
| `element-priority-guide.md`（系列 library 下） | 提供更细粒度的优先级排序 |
| `shot-production-workflow.md` | 本规范在 chapter_design 确认后、shot_plan 之前触发 |
| `prompt-engineering.md` | 提示词组装时按本规范的格式引用元素 |
| `character-consistency.md` | 有标识时一致性由平台保证；无标识时需严格遵循文字描述规则 |
