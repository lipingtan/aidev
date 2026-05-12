# KLING 动作场景模式

> 验证有效的动作场景提示词模式。

---

## 模式：单一明确动作

**适用场景**：角色执行一个清晰的物理动作
**验证次数**：待积累

### 有效写法
```
Medium shot: [角色] reaches for the cup on the table, fingers wrapping 
around the warm ceramic. Steam curls upward as she lifts it to her lips.
(Duration: 4 seconds)
```

### 无效写法
```
❌ She picks up the cup, drinks, puts it down, and looks out the window.
（多个动作堆叠，模型只会执行第一个或混乱）
```

### 为什么有效
- 一个镜头只描述一个主要动作
- 加入物理细节（fingers wrapping, steam curls）增加真实感
- 动作有起始和结束状态

### 标签
#action #single-action #physical-detail

---

## 模式：行走/移动

**适用场景**：角色在空间中移动
**验证次数**：待积累

### 有效写法
```
Tracking shot, following from behind: [角色] walks down the rain-soaked 
cobblestone street, shoulders slightly hunched against the cold. 
Her coat billows gently with each step.
(Duration: 5 seconds)
```

### 关键要点
- 用跟拍（tracking）配合移动
- 描述移动的物理表现（衣物摆动、步态特征）
- 加入环境互动（雨水、风）

### 标签
#action #walking #tracking-shot

---

## 模式：战斗/激烈动作

**适用场景**：打斗、追逐等高强度动作
**验证次数**：待积累

### 有效写法
```
Dynamic low angle, handheld: [角色] swings the sword in a wide arc, 
blade catching the firelight. Sparks fly as metal meets metal.
(Duration: 3 seconds)
```

### 关键要点
- 时长要短（2-3秒），动作越激烈镜头越短
- 用动态角度（low angle, handheld）增强冲击力
- 加入物理反馈（火花、尘土、碎片）
- 每镜头只有一个动作节拍

### 标签
#action #combat #dynamic #short-duration

---

## 模式：细微动作/小动作

**适用场景**：手部动作、面部微表情等细节
**验证次数**：待积累

### 有效写法
```
Extreme close-up: Her finger traces the edge of the old photograph, 
nail catching slightly on the worn corner. A barely visible tremor in her hand.
(Duration: 3 seconds)
```

### 关键要点
- 用特写/极特写
- 描述触觉细节
- 微动作暗示情绪

### 标签
#action #micro-action #close-up #detail
