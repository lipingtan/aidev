# 参考图生成策略指南

> 本文件定义如何为 elements-registry 中的注册条目生成 KLING 平台所需的参考图提示词。

---

## 零、触发时机与输出位置

### 触发时机

**元素就绪检查时**（chapter_design 确认后、shot_plan 开始前）自动触发。
具体流程见 `element-readiness.md` 步骤 3。

### 输出位置

所有参考图提示词输出到 `output/{系列}/assets/`，按类型分子目录：

```
output/{系列}/assets/
├── characters/     # 角色素体（{标识}.md）+ 服装（{标识}.md）
├── scenes/         # 场景（{标识}.md）
├── props/          # 武器/道具（{标识}.md）
├── fx/             # 魔法效果（{标识}.md）
└── enemies/        # 敌方单位（{标识}.md）
```

文件命名与 elements-registry.md 中的标识完全对应：
- `alan_body` → `output/{系列}/assets/characters/alan_body.md`
- `alan_sword` → `output/{系列}/assets/props/alan_sword.md`
- `royal_city_night` → `output/{系列}/assets/scenes/royal_city_night.md`

### 文件格式（每个 .md 文件）

```markdown
# {标识}

{提示词正文，纯英文，可直接复制粘贴到 KLING 图片生成}

Negative: {负面提示词}
```

> 文件内容只有提示词，无注释无说明，可直接复制使用。

### 生成规则

- 已存在的文件不重复生成
- 优先级为"不需要"的元素（单次出现的背景元素）不生成提示词文件
- 生成后在就绪报告中列出新生成的文件路径

---

## 一、工具选择

| 用途 | 推荐工具 | 理由 |
|------|----------|------|
| 角色素体 | KLING 图片生成 / Flux | 与视频引擎风格一致性最好 |
| 服装展示 | KLING 图片生成 | 同上 |
| 场景参考 | KLING 图片生成 | 确保视频生成时风格匹配 |
| 道具特写 | Midjourney / DALL-E | 细节控制更好 |

> **核心原则**：参考图的风格必须与 style-bible.md 中定义的"偏写实电影感"一致。避免使用动漫风格或过度风格化的参考图。

---

## 二、角色素体参考图规范

### 生成要求

| 属性 | 规范 |
|------|------|
| 构图 | 正面半身（头顶到腰部），或正面全身 |
| 背景 | 纯色/简洁背景（白色、浅灰、深灰） |
| 服装 | **不含具体服装**，仅穿简单基础衣物以展示体型 |
| 表情 | 中性表情，面部特征清晰可辨 |
| 光线 | 均匀柔光，无强烈阴影 |
| 分辨率 | ≥ 1024×1024 |
| 比例 | 1:1 或 3:4（竖版） |

### 提示词模板

```
Portrait of [角色素体描述 from characters/*.md],
neutral expression, looking directly at camera,
simple grey background, soft even lighting,
photorealistic, high detail, cinematic quality,
upper body shot

Negative: blur, distort, low quality, anime style, cartoon,
oversaturated, busy background, text, watermark
```

### 示例：艾伦素体

```
Portrait of a tall muscular young man in his early 20s,
short messy black hair, sharp grey eyes, strong jawline,
prominent dragon scale tattoo on his left arm,
wearing a simple dark undershirt,
neutral expression, looking directly at camera,
simple grey background, soft even lighting,
photorealistic, high detail, cinematic quality,
upper body shot

Negative: blur, distort, low quality, anime style, cartoon,
oversaturated, busy background, text, watermark
```

---

## 三、服装参考图规范

### 生成要求

| 属性 | 规范 |
|------|------|
| 构图 | 全身站立，展示完整服装 |
| 背景 | 纯色或极简环境 |
| 姿势 | 自然站立，双手可见 |
| 角度 | 正面或3/4侧面 |
| 重点 | 服装细节清晰（材质、颜色、配饰位置） |
| 分辨率 | ≥ 1024×1024 |

### 提示词模板

```
Full body shot of [角色素体描述],
[服装描述 from characters/*.md Wardrobe 对应条目],
natural standing pose, hands visible,
simple background, soft even lighting,
photorealistic, cinematic quality, full body visible head to toe

Negative: blur, distort, low quality, cropped body,
anime style, busy background
```

---

## 四、场景参考图规范

### 生成要求

| 属性 | 规范 |
|------|------|
| 构图 | 宽幅横版（16:9） |
| 内容 | **无角色**，纯环境 |
| 光线 | 与 scenes/*.md 中定义的光线一致 |
| 氛围 | 与 style-bible.md 色彩方案一致 |
| 分辨率 | ≥ 1920×1080 |
| 比例 | 16:9 |

### 提示词模板

```
[场景描述 from scenes/*.md 对应段落],
no people, empty scene, establishing shot,
cinematic, photorealistic, dark fantasy,
dramatic lighting, 16:9 aspect ratio

Negative: people, characters, figures, anime style,
low quality, blur, text
```

---

## 五、道具/武器参考图规范

### 生成要求

| 属性 | 规范 |
|------|------|
| 构图 | 物品居中，占画面 60~80% |
| 背景 | 纯黑或纯灰 |
| 光线 | 产品摄影式打光（突出材质和细节） |
| 角度 | 最能展示特征的角度 |
| 分辨率 | ≥ 1024×1024 |

### 提示词模板

```
[道具描述 from props-registry.md],
product photography style, centered composition,
dark background, dramatic lighting highlighting details,
photorealistic, high detail, no hands, isolated object

Negative: blur, low quality, hands, people, busy background
```

---

## 六、敌方单位参考图规范

### 生成要求

| 属性 | 规范 |
|------|------|
| 构图 | 全身，展示完整形态 |
| 背景 | 深色/黑雾环境（符合其出现场景） |
| 姿势 | 威胁性姿态（攻击准备/移动中） |
| 分辨率 | ≥ 1024×1024 |

### 提示词模板

```
[敌方单位描述 from characters/enemies.md],
threatening pose, dark misty environment,
cinematic, photorealistic, dark fantasy,
dramatic purple-black lighting

Negative: cute, friendly, anime style, low quality, blur
```

---

## 七、参考图命名与存放

### 存放路径

```
AIDOC/lost-crown/references/
├── characters/          # 角色素体 + 服装
│   ├── alan-body.png
│   ├── alan-outfit-casual.png
│   ├── alan-outfit-combat.png
│   └── ...
├── scenes/              # 场景参考
│   ├── royal-city-night.png
│   ├── royal-city-alley.png
│   └── ...
├── props/               # 道具/武器
│   ├── alan-sword.png
│   ├── lost-crown.png
│   └── ...
└── enemies/             # 敌方单位
    ├── enemy-mist-beast.png
    ├── enemy-bone-knight.png
    └── ...
```

### 命名规范

- 文件名与 elements-registry.md 中的"参考图"列完全一致
- 全部小写，单词用连字符分隔
- 格式：PNG（优先）或 JPG

---

## 八、质量检查清单

生成参考图后，逐项检查：

- [ ] 风格是否为写实电影感（非动漫/卡通）？
- [ ] 角色面部特征是否与 characters/*.md 描述一致？
- [ ] 服装细节是否完整可辨？
- [ ] 背景是否足够简洁（不干扰主体）？
- [ ] 分辨率是否达标？
- [ ] 是否有明显的 AI 瑕疵（多余手指、扭曲面部）？
- [ ] 色调是否与 style-bible 一致？

> 如果某张参考图不满意，重新生成时调整提示词，不要勉强使用低质量参考图——它会影响后续所有视频的角色一致性。

---

## 九、生成批次计划

按 element-priority-guide.md 的优先级分批生成：

| 批次 | 内容 | 数量 | 时机 |
|------|------|------|------|
| 第一批 | P0 角色素体（5个）+ P0 服装（8个）+ P0 武器（1个） | ~14 | 开拍前 |
| 第二批 | P0 场景（2个）+ P1 场景（8个） | ~10 | 序幕开拍前 |
| 第三批 | P1 角色素体 + P1 服装 + P1 武器 | ~15 | 第一幕开拍前 |
| 后续 | P2 按需 | — | 对应章节前 |
