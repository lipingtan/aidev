## Shot Design: ep-001

### 引擎
KLING AI（多镜头模式）

### Master Prompt 引用
```
Cinematic, photorealistic, dark fantasy, live-action style, shallow depth of field, dramatic lighting
```

### 角色视觉
无（纯环境建立镜头）

### 场景视觉（从 library/scenes/royal-city.md 提取）

全景：
```
Aerial view of a medieval fantasy city at night, snow falling gently,
stone walls and towers silhouetted against a cold blue moonlit sky,
warm amber torch lights dotting the streets below like scattered embers,
a large castle at the city center, breath of winter mist over rooftops
```

街巷：
```
A narrow cobblestone alley in a medieval city at night, snow covering the ground,
iron torch brackets on stone walls casting flickering amber light,
frost on window panes, visible breath in cold air,
empty and quiet, footprints in fresh snow, a sense of solitude
```

**Spatial DNA（王都艾伦迪尔）**：
```
cold grey stone walls, dark slate-grey rooftops,
gothic spires and towers, iron torch brackets on stone walls,
cobblestone streets worn smooth, frosted window panes,
snow-dusted surfaces, winter mist between buildings
```

### 各镜头设计

| 镜头 | 时长 | 景别+运镜 | 动作描述（英文） | 光线 |
|------|------|-----------|-----------------|------|
| ① | 5s | Extreme wide, slow crane down | Aerial view of medieval city at night, snow falling, gothic spires silhouetted against moonlit sky, warm torch lights scattered below | cold blue moonlight from above, warm amber torch lights below |
| ② | 5s | Wide to medium, slow push in | Camera descends to street level, iron torch brackets on stone walls casting flickering amber light, snow on cobblestone, winter mist drifting | flickering torch light on stone walls, deep shadows between buildings |
| ③ | 5s | Medium, slow pan right | A tavern door in the distance with warm golden light spilling through cracks, snow-covered street leading toward it, footprints in fresh snow | warm interior glow from tavern vs cold exterior moonlight |

### 情绪氛围（英文）
```
solitude, cold winter night, quiet before the storm, faint warmth in the distance
```

### 元素绑定

| 元素 | 标识 | 用途 |
|------|------|------|
| 王都全景（夜） | `@royal_city_night` | Shot ① 城市俯瞰一致性 |
| 王都街巷（夜） | `@city_alley_night` | Shot ②③ 街道一致性 |

### 负面提示词
```
Negative: blur, distort, low quality, warping fingers, jittery eyes, character drift between shots, tonal shift between cuts, flickering highlights
```

### 预期时长
15秒

### 对话音频
无
