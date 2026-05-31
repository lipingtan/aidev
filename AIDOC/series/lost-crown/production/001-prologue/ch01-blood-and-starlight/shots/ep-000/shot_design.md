## Shot Design: ep-000

### 引擎
KLING AI（多镜头模式）

### Master Prompt 引用
```
Cinematic, photorealistic, dark fantasy, live-action style, shallow depth of field, dramatic lighting
```

### 角色视觉（从 library/characters/alan.md 提取）
- 艾伦(W4)：`dragon scales covering entire body, both eyes glowing molten gold, no human expression visible, torn armor hanging in shreds, gold-black veins pulsing across all skin, feral predatory stance`

> Shot 3 为半怪物化（过渡态），调整为：龙鳞覆盖半脸，左眼金色右眼灰色，痛苦嘶吼。

### 场景视觉（从 library/scenes/dragon-mountains.md 提取）
```
The peak of the highest mountain, a massive crater-like arena of black bone and stone,
the sky above is a vortex of swirling black mist,
ancient dragon bones embedded in the ground, glowing purple cracks everywhere,
an altar-like formation at the center where the Shadow Dragon King manifests
```

**Spatial DNA（影龙山脉）**：
```
black volcanic rock and bone-white formations,
perpetual dark purple-black mist obscuring sky,
no living vegetation, petrified trees,
purple lightning in clouds above,
ground cracked with glowing purple fissures,
oppressive supernatural darkness
```

### 各镜头设计

| 镜头 | 时长 | 景别+运镜 | 动作描述（英文） | 光线 |
|------|------|-----------|-----------------|------|
| ① | 1s | Extreme wide, static | A colossal dragon with intertwined bone and black mist, bone wings over 50-meter wingspan blocking the sky, purple lightning tearing through the black mist sky behind it, crown-like bone horns on head | purple lightning backlight, no natural light |
| ② | 1s | Wide shot, static | Four intertwining pillars of light shooting into the stormy sky: gold-black pulsing, silver-white starlight, orange-red flames, emerald-green bioluminescence, four human silhouettes at the base | four-color self-illumination vs purple-black background |
| ③ | 1s | Close-up, static | Young man's face in agony, dragon scales spreading across half his face and neck, gold-black veins glowing beneath skin, left eye burning gold right eye grey, mouth open in painful scream, torn dark armor at shoulders | dragon scale gold-black glow + purple lightning from behind |

### 情绪氛围（英文）
```
overwhelming dread, epic confrontation, supernatural storm, agony and defiance
```

### 元素绑定

| 元素 | 标识 | 用途 |
|------|------|------|
| 艾伦·素体 | `@alan_body` | 角色面部/体型（Shot ③） |
| 艾伦·W4 怪物化 | `@alan_w4` | 龙鳞覆盖服装（Shot ③） |

> 影龙王无元素绑定，纯文字描述。四人剪影为远景轮廓，无需绑定。
> 影龙山脉场景未注册，用文字描述替代。

### 负面提示词
```
Negative: blur, distort, low quality, warping fingers, jittery eyes, character drift between shots, unnatural motion, cartoon style, anime style, tonal shift between cuts
```

### 预期时长
3秒（KLING 生成纯画面）+ 2秒（后期黑屏字幕"3 years earlier"）

### 对话音频
无

### 生成参数

| 参数 | 值 |
|------|-----|
| 模式 | 多镜头（Multi-shot） |
| 提交时长 | 5s |
| 画面比例 | 16:9 |
| 运动幅度 | 低 |
| 创意度 | 0.6 |

### 后期备注
- Shot ④（2s）为后期制作：硬切全黑 → 0.5s纯黑 → 白色字幕"3 years earlier"淡入(0.5s) → 保持(0.5s) → 淡出(0.5s) → 接 ep-001
- KLING 只需生成前3个镜头
