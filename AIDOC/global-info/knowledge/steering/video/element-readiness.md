# 元素就绪检查规范

> 在进入逐镜头制作（shot_plan/shot_design）之前，必须确认该章节涉及的所有关键元素已在 KLING 平台创建并注册了标识。

---

## 使用时机

| 步骤 | 动作 |
|------|------|
| **chapter_design 确认后** | 执行本文件的"元素就绪检查"，确认所有必需元素已就绪 |
| **shot_design 编写时** | 根据元素就绪状态决定用 `@标识` 还是文字描述 |
| **提示词组装时** | 按本文件的"提示词引用格式"输出 |

---

## 一、元素就绪检查流程

### 触发条件

chapter_design 确认后、开始第一个 shot_plan 之前。

### 检查步骤

```
1. 从 chapter_design 的"正式片段表"中提取所有涉及的：
   - 角色（含服装变体）
   - 场景
   - 武器/道具
   - 魔法效果（如高频出现）

2. 对照 elements-registry.md，逐项检查：
   - "KLING 平台标识"列是否已填入（非空 = 已就绪）
   - 未填入 = 未就绪

3. 输出就绪报告：
```

### 就绪报告模板

```markdown
## 元素就绪检查：{章节名}

### 已就绪（有平台标识）

| 元素 | 标识 | 类型 |
|------|------|------|
| 艾伦·素体 | @alan_body | 角色 |
| 艾伦·斗殴后 | @alan_torn | 服装 |
| 王都街巷（夜） | @city_alley | 场景 |

### 未就绪（需要创建或用文字描述）

| 元素 | 优先级 | 处理方式 |
|------|--------|----------|
| 酒馆外部 | 高（多次出现） | 建议创建 |
| 乌鸦 | 低（单次出现） | 文字描述即可 |

### 行动项

- [ ] 需要用户在 KLING 平台创建 {N} 个元素
- [ ] 创建后将标识填入 elements-registry.md
- [ ] 全部就绪后开始 shot_plan
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
