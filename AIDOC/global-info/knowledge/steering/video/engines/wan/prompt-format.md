# WAN 2.5 提示词格式规范

## 整体结构

```
[整体描述和风格设定]
Shot 1 [0-3s]: [镜头1内容描述]
Shot 2 [3-6s]: [镜头2内容描述]
Shot 3 [6-9s]: [镜头3内容描述]
Shot 4 [9-12s]: [镜头4内容描述]
Shot 5 [12-15s]: [镜头5内容描述]
```

## 整体描述模板

```
[叙事概述，说明这是什么故事/场景]. [风格设定].
Shot 1 [0s-Xs]: [景别]: [具体画面描述，包含主体、动作、环境细节].
Shot 2 [Xs-Ys]: [景别]: [具体画面描述].
...
```

## 示例

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

## 提示词要素

1. **主体**：角色/物体的视觉描述
2. **场景**：单一明确的场景设定
3. **光线锚点**：明确的光源描述
4. **具体动作**：清晰的动作指令
5. **风格**：色调、质感、画面风格

## 关键规则

1. **时间戳必须连续**：`[0-3s]` → `[3-6s]` → `[6-9s]`，不能有间隔
2. **单一场景**：WAN 对单一场景的理解更好，避免场景跳转
3. **光线锚点**：明确指定一个主光源
4. **动作简洁**：每个 Shot 只描述一个主要动作
5. **总时长 ≤ 15s**：所有 Shot 时间戳之和不超过 15 秒
