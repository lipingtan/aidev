# EP-000: 闪前开场（Flash Forward） — Shot Design

## 参考文件

> 本设计基于以下文件：

- `AIDOC/lost-crown/library/characters/alan.md` — 艾伦 W4 怪物化服装视觉关键词
- `AIDOC/lost-crown/library/characters/karasa.md` — 影龙王形态视觉关键词
- `AIDOC/lost-crown/library/style-bible.md` — 风格锚定词、负面提示词
- `AIDOC/lost-crown/library/visual-direction.md` — 闪前开场规范、影龙之眼、四色光柱

---

## 基本信息

| 属性 | 值 |
|------|-----|
| 片段编号 | ep-000 |
| 总时长 | 5s（前3s画面 + 后2s黑屏字幕） |
| KLING 实际生成 | 3s（纯画面部分），字幕后期叠加 |
| 镜头数 | 3（KLING 生成）+ 1（后期） |
| 生成模式 | KLING 多镜头模式 |
| 画面比例 | 16:9 |
| 对话 | 无 |

---

## 元素绑定

| 元素名称 | 类型 | 用途 |
|----------|------|------|
| 艾伦·素体 | 角色 | 基础面部/体型锁定 |
| 艾伦·怪物化(W4) | 服装 | 龙鳞覆盖、金黑脉络 |

> 影龙王为纯画面描述（无元素绑定），四人剪影为远景轮廓（无需绑定）。

---

## 逐镜头设计

### Shot 1（1s）— 影龙王骨翼遮天

| 属性 | 值 |
|------|-----|
| 景别 | 远景（Extreme Wide Shot） |
| 运镜 | 静止 |
| 光源 | 紫色闪电为主光源，逆光 |
| 色调 | 紫黑为主，闪电瞬间照亮骨翼轮廓 |

**画面描述**：

```
A colossal dragon with intertwined bone and black mist, 
bone wings over 50-meter wingspan blocking the sky like dark clouds,
purple lightning tearing through the black mist sky behind it,
silhouetted against the storm, crown-like bone horns on head,
massive pulsing dark purple heart faintly visible in ribcage
```

**视觉要点**：
- 影龙王占据画面上方 2/3，压迫感极强
- 紫色闪电在骨翼缝隙间劈下，照亮骨骼纹理
- 画面底部是黑暗的山脉轮廓（影龙山脉）
- 黑雾从龙身周围翻涌

---

### Shot 2（1s）— 四色光柱冲天

| 属性 | 值 |
|------|-----|
| 景别 | 全景（Wide Shot） |
| 运镜 | 静止 |
| 光源 | 四色光柱自发光 |
| 色调 | 四色（金黑+银白+橙红+翠绿）vs 紫黑背景 |

**画面描述**：

```
Four intertwining pillars of light shooting into the stormy sky:
gold-black pulsing light, silver-white starlight, orange-red flames, 
emerald-green bioluminescence, spiraling together,
four human silhouettes standing at the base of the light pillars,
dark purple mist swirling around them, epic scale confrontation
```

**视觉要点**：
- 四人为黑色剪影，站在山顶岩石上，面朝影龙王方向
- 四色光柱从四人位置向天空射出，在高处交汇
- 光柱与周围紫黑雾气形成强烈冷暖对比
- 画面构图：四人在下方 1/3，光柱贯穿中间，天空占上方 1/3

---

### Shot 3（1s）— 艾伦半怪物化特写

| 属性 | 值 |
|------|-----|
| 景别 | 特写（Close-up） |
| 运镜 | 静止 |
| 光源 | 龙鳞金黑自发光 + 环境紫色闪电 |
| 色调 | 金黑（龙鳞）+ 紫色（环境）+ 肤色 |

**画面描述**：

```
Extreme close-up of a young man's face in agony,
dragon scales spreading across half his face and neck,
gold-black veins glowing beneath skin,
left eye burning gold, right eye remaining grey,
mouth open in a painful scream,
torn dark armor visible at shoulders,
purple lightning illuminating from behind,
rain and mist particles in the air
```

**视觉要点**：
- 面部占满画面，龙鳞从左侧蔓延到鼻梁附近
- 双眼异色是关键视觉锚点（左金右灰）
- 金黑脉络在皮肤下发光，如岩浆裂纹
- 表情是痛苦而非愤怒——嘶吼中带有挣扎
- 背景虚化为紫色闪电和黑雾

---

### Shot 4（2s）— 黑屏 + 字幕（后期制作）

| 属性 | 值 |
|------|-----|
| 景别 | — |
| 运镜 | — |
| 制作方式 | 后期剪辑叠加，非 KLING 生成 |

**效果**：
- Shot 3 结束后硬切全黑
- 0.5s 纯黑
- 白色字幕 "3 years earlier" 淡入（0.5s）
- 保持 0.5s
- 淡出（0.5s）→ 接 ep-001

---

## 生成参数

| 参数 | 值 |
|------|-----|
| 模式 | 多镜头（Multi-shot） |
| 时长 | 5s（含字幕过渡）或 3s（纯画面） |
| 画面比例 | 16:9 |
| 运动幅度 | 低（静止镜头为主） |
| 创意度 | 0.6（保持可控） |

---

## Master Prompt

```
Cinematic, photorealistic, dark fantasy, live-action style, shallow depth of field, dramatic lighting,
epic final battle scene on a dark mountain peak during a supernatural storm,
purple-black mist and lightning filling the sky, rain and particles in the air
```

## 负面提示词

```
blur, distort, low quality, warping fingers, jittery eyes, character drift between shots,
unnatural motion, cartoon style, anime style, 3D render, oversaturated
```

---

## 状态

- [x] Shot Design 已生成
- [x] 用户已确认 ✅
- [x] 提示词已生成 → `output/lost-crown/kling/001-prologue/ch01-blood-and-starlight/ep-000.md`
