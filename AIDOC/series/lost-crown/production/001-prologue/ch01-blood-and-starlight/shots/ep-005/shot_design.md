## Shot Design: ep-005

### 引擎
KLING AI（多镜头模式）

### Master Prompt 引用
```
Cinematic, photorealistic, dark fantasy, live-action style, shallow depth of field, dramatic lighting
```

### 角色视觉（从 library/characters/alan.md 提取）
- 艾伦(W3)：`A tall muscular young man in his early 20s, short messy black hair, sharp grey eyes, strong jawline, prominent dragon scale tattoo on his left arm, wearing a torn dark grey tunic, blood on his knuckles, disheveled, black sword on back`

### 场景视觉（从 library/scenes/alan-home.md 推断）
```
A modest stone house with a heavy wooden door, iron hinges,
warm amber light visible through frosted window panes,
snow piled on the doorstep, a small overhang above the entrance
```

**Spatial DNA（王都艾伦迪尔）**：
```
cold grey stone walls, iron torch brackets on stone walls,
cobblestone streets worn smooth, frosted window panes, snow-dusted surfaces
```

### 各镜头设计

| 镜头 | 时长 | 景别+运镜 | 动作描述（英文） | 光线 |
|------|------|-----------|-----------------|------|
| ① | 5s | Medium, static | He stops before a modest stone door, snow on his shoulders, looks at the warm light through the frosted window | cold moonlight on him, warm glow from window |
| ② | 5s | Close-up, static | He takes a deep breath, closes his eyes briefly, then opens them with a softer expression, wiping blood from his lip with his sleeve | cold ambient, faint warm reflection from window |
| ③ | 5s | Medium, slow push in | He pushes the heavy wooden door open, warm golden firelight floods out onto his face and the snowy ground, he steps inside | dramatic warm light flooding from interior |

### 情绪氛围（英文）
```
transition from cold isolation to warmth, composing himself, gentle anticipation
```

### 元素绑定

| 元素 | 标识 | 用途 |
|------|------|------|
| 艾伦·素体 | `@alan_body` | 角色面部/体型 |
| 艾伦·W3 斗殴后 | `@alan_w3` | 本段服装 |
| 艾伦居所（内部） | `@alan_home_int` | Shot ③ 门开后内部一致性 |

### 负面提示词
```
Negative: blur, distort, low quality, warping fingers, jittery eyes, character drift between shots, unnatural motion, lighting inconsistency
```

### 预期时长
15秒

### 对话音频
无
