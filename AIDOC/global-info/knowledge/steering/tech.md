# 视频引擎技术规格

## KLING AI 3.0

| 参数 | 规格 |
|------|------|
| 最大镜头数 | 6 镜头/次生成 |
| 最大时长 | 15 秒/次生成 |
| 分辨率 | 1080P（16:9 / 9:16 / 1:1） |
| 角色一致性 | 元素绑定功能（上传参考图） |
| 音频 | 原生对话 + 环境音 |
| 对话格式 | `[角色: 身份, 语气]: "台词"` |

### KLING 提示词公式

> KLING 多镜头模式中，每个镜头各自一个完整提示词（含场景、风格、负面提示词），没有独立的 Master Prompt。

```
Shot 1 (Duration: Xs): [完整镜头描述 + 风格锚定词 + Negative]
Shot 2 (Duration: Ys): [完整镜头描述 + 风格锚定词 + Negative]
...
Shot N (Duration: Zs): [完整镜头描述 + 风格锚定词 + Negative]
```

每个 Shot 提示词必须自包含：场景上下文、角色描述、动作、镜头语言、光线、氛围、风格锚定词、负面提示词。

### KLING 单镜头提示词要素（按顺序）

1. **Subject（主体）**：角色外貌、服装、表情、姿态
2. **Movement（动作）**：角色动作、物理运动
3. **Scene（场景）**：环境、背景、空间关系
4. **Cinematic Language（镜头语言）**：景别、运镜、视角
5. **Lighting（光线）**：光源、光质、光影效果
6. **Atmosphere（氛围）**：情绪基调、环境效果、色调

### KLING 对话格式

```
单人对话：
[角色A: 身份, 语气描述]: "台词内容"

多人对话：
[角色A: 身份, 语气描述]: "第一句台词"
Immediately, [角色B: 身份, 情绪语气]: "回应台词"
```

### KLING 对话时长限制

| 镜头时长 | 台词字数上限（英文） | 台词字数上限（中文） |
|----------|---------------------|---------------------|
| 5 秒 | 8~12 词 | 10~15 字 |
| 4 秒 | 6~9 词 | 8~12 字 |
| 3 秒 | 4~6 词 | 5~8 字 |

### KLING 6 轴摄像机控制

| 轴 | 正值效果 | 负值效果 |
|----|----------|----------|
| Horizontal | 镜头右移 | 镜头左移 |
| Vertical | 镜头上移 | 镜头下移 |
| Pan | 镜头右转 | 镜头左转 |
| Tilt | 镜头上仰 | 镜头下俯 |
| Roll | 顺时针旋转 | 逆时针旋转 |
| Zoom | 推进 | 拉远 |

---

## WAN 2.5 / 2.6

| 参数 | 规格 |
|------|------|
| 最大时长 | 15 秒/次生成 |
| 分辨率 | 480P / 720P / 1080P（16:9） |
| 角色一致性 | 角色卡 / 参考视频 |
| 音频 | 原生音频（对话 + 环境音） |
| 分镜格式 | 时间戳标注 `Shot N [Xs-Ys]` |

### WAN 提示词格式

```
[整体描述和风格设定]
Shot 1 [0-3s]: [镜头1内容描述]
Shot 2 [3-6s]: [镜头2内容描述]
Shot 3 [6-9s]: [镜头3内容描述]
Shot 4 [9-12s]: [镜头4内容描述]
Shot 5 [12-15s]: [镜头5内容描述]
```

### WAN 提示词要素

1. **主体**：角色/物体的视觉描述
2. **场景**：单一明确的场景设定
3. **光线锚点**：明确的光源描述
4. **具体动作**：清晰的动作指令
5. **风格**：色调、质感、画面风格

---

## 通用负面提示词（Negative Prompts）

> KLING 没有独立的 Negative Prompt 输入框，直接写在 Master Prompt 末尾，用 `Negative:` 前缀。
> **数量控制在 5~8 个**，超过 20 个会让画面变平。

### KLING 3.0 推荐基础集（6 个，每次必加）

```
Negative: blur, distort, low quality, warping fingers, jittery eyes, character drift between shots
```

> KLING 3.0 已改善 frozen lips 和 plastic skin 问题，可不加。但多镜头模式需要加 `character drift between shots`。

### 按场景追加（选 2~3 个）

| 场景类型 | 追加项 |
|----------|--------|
| 有对话 | audio desync, garbled speech, mouth not matching words |
| 多角色 | face swap, character merge, identity drift |
| 电影感 | unnatural motion, stuttered movement, flickering highlights |
| 多镜头 | tonal shift between cuts, lighting inconsistency |

---

## 常用镜头语言速查

| 中文 | 英文术语 | 说明 |
|------|----------|------|
| 特写 | Close-up / ECU | 面部或细节 |
| 中景 | Medium shot | 腰部以上 |
| 全景 | Wide shot / Full shot | 全身 + 环境 |
| 远景 | Establishing shot | 大环境 |
| 低角度 | Low angle | 仰拍，增强气势 |
| 高角度 | High angle | 俯拍，压迫感 |
| 跟拍 | Tracking shot | 跟随主体移动 |
| 推镜 | Push in / Dolly in | 逐渐靠近主体 |
| 拉镜 | Pull back / Dolly out | 逐渐远离主体 |
| 摇镜 | Pan | 水平旋转 |
| 升降 | Crane / Jib | 垂直移动 |
| 手持 | Handheld | 轻微晃动，纪实感 |
| 稳定器 | Stabilizer / Gimbal | 平滑移动 |
| 第一人称 | POV / First person | 主观视角 |
