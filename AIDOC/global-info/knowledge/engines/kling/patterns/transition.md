# KLING 转场模式

> 验证有效的片段间转场提示词模式。

---

## 模式：动作衔接（Action Match）

**适用场景**：同一动作跨两个片段
**验证次数**：待积累

### 有效写法

**片段 A 结尾：**
```
Shot 6: [角色] turns toward the door, hand reaching for the handle.
(Duration: 3 seconds)
```

**片段 B 开头：**
```
Shot 1: [角色]'s hand grips the door handle, pulling it open. 
Cold wind rushes in from outside.
(Duration: 3 seconds)
```

### 关键要点
- 上一段结束状态 = 下一段开始状态
- 动作的"中间帧"作为衔接点
- 两段的 Master Prompt 光线/氛围保持一致

### 标签
#transition #action-match #continuity

---

## 模式：情绪衔接（Emotional Bridge）

**适用场景**：场景切换但情绪连续
**验证次数**：待积累

### 有效写法

**片段 A（室内，悲伤）结尾：**
```
Shot 6: Close-up: A tear rolls down her cheek, catching the lamplight.
(Duration: 3 seconds)
```

**片段 B（室外，悲伤延续）开头：**
```
Master Prompt: ... Melancholic, somber atmosphere. Desaturated cool tones.
Shot 1: Wide shot: Rain falls on empty cobblestone street, 
a lone figure walks away from camera, shoulders hunched.
(Duration: 4 seconds)
```

### 关键要点
- 两段 Master Prompt 使用相同的情绪关键词
- 色调保持一致（都是冷调/暖调）
- 用环境呼应情绪（泪水 → 雨水）

### 标签
#transition #emotional-bridge #mood-continuity

---

## 模式：对比切（Contrast Cut）

**适用场景**：刻意制造情绪/视觉反转
**验证次数**：待积累

### 有效写法

**片段 A（温暖回忆）结尾：**
```
Shot 6: Warm golden light, soft focus: Young [角色] laughing, 
sunlight in her hair.
(Duration: 3 seconds)
```

**片段 B（冰冷现实）开头：**
```
Master Prompt: ... Cold, sterile atmosphere. Harsh fluorescent lighting. 
Desaturated, clinical tones.
Shot 1: Wide shot: [角色] sits alone in a white hospital corridor, 
harsh overhead light casting sharp shadows.
(Duration: 4 seconds)
```

### 关键要点
- 刻意在色调、光线、氛围上制造反差
- 反差越大冲击力越强
- 适合回忆 vs 现实、希望 vs 绝望的叙事

### 标签
#transition #contrast-cut #dramatic

---

## 模式：时间跳跃（Time Jump）

**适用场景**：跨越较长时间段
**验证次数**：待积累

### 有效写法

**新片段开头：**
```
Shot 1: Establishing shot, slow crane down: The city skyline at dawn, 
first light breaking over the rooftops. [时间标记性元素：如季节变化、
建筑变化等].
(Duration: 4 seconds)
```

### 关键要点
- 用 Establishing shot 重新建立环境
- 加入时间标记元素（季节、光线、环境变化）
- 给观众"重新定位"的时间（4-5秒）

### 标签
#transition #time-jump #establishing-shot

---

## 转场方式选择指南

| 叙事需求 | 推荐转场 | 注意事项 |
|----------|----------|----------|
| 同场景连续动作 | 动作衔接 | 保持光线一致 |
| 换场景但情绪连续 | 情绪衔接 | 色调和氛围词一致 |
| 情绪/场景反转 | 对比切 | 反差要明显 |
| 跨时间段 | 时间跳跃 | 用 Establishing shot |
| 平行叙事 | 交叉剪辑 | 两条线的节奏要匹配 |
