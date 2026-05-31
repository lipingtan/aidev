## Shot Design: ep-003

### 引擎
KLING AI（多镜头模式）

### Master Prompt 引用
```
Cinematic, photorealistic, dark fantasy, live-action style, shallow depth of field, dramatic lighting
```

### 角色视觉（从 library/characters/alan.md 提取）
- 艾伦(W3)：`A tall muscular young man in his early 20s, short messy black hair, sharp grey eyes, strong jawline, prominent dragon scale tattoo on his left arm, wearing a torn dark grey tunic, blood on his knuckles, disheveled, black sword on back`

### 场景视觉
```
A narrow cobblestone alley in a medieval city at night, snow covering the ground,
iron torch brackets on stone walls casting flickering amber light,
frost on window panes, visible breath in cold air
```

**Spatial DNA（王都艾伦迪尔）**：
```
cold grey stone walls, iron torch brackets on stone walls,
cobblestone streets worn smooth, snow-dusted surfaces, winter mist between buildings
```

### 各镜头设计

| 镜头 | 时长 | 景别+运镜 | 动作描述（英文） | 光线 |
|------|------|-----------|-----------------|------|
| ① | 5s | Medium close-up, static | He grips his left arm in pain, leaning against a cold stone wall, dragon tattoo glowing faintly, teeth clenched, breath visible | gold-black tattoo glow, cold blue moonlight on wall |
| ② | 5s | Medium, slow pull back | The glow fades, he straightens up slowly, composing himself, releases his arm | fading gold glow, returning to cold ambient torch light |
| ③ | 5s | Wide, slow tracking follow | He walks alone down the long snowy street, large sword silhouette on his back, footprints trailing behind in fresh snow, small and solitary in the empty alley | cold blue moonlight from above, distant torch light |

### 情绪氛围（英文）
```
pain fading to numbness, forced composure, deep solitude
```

### 元素绑定

| 元素 | 标识 | 用途 |
|------|------|------|
| 艾伦·素体 | `@alan_body` | 角色面部/体型 |
| 艾伦·W3 斗殴后 | `@alan_w3` | 本段服装 |
| 龙鳞纹身发光 | `@fx_tattoo_glow` | Shot ①② 纹身光效 |
| 王都街巷（夜） | `@city_alley_night` | 场景一致性 |

### 负面提示词
```
Negative: blur, distort, low quality, warping fingers, jittery eyes, character drift between shots, unnatural motion, sliding feet
```

### 预期时长
15秒

### 对话音频
无
