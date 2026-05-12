# KLING AI 3.0 最佳实践

## 提示词编写

### 1. 场景先行（Scene-First）

先建立环境上下文，再描述主体和动作。给模型提供空间和光线的计算基础。

```
✅ 好的顺序：环境 → 主体 → 动作 → 镜头 → 光线 → 氛围
❌ 差的顺序：动作 → 主体 → 环境（模型缺乏空间上下文）
```

### 2. 具体化（Specificity）

用具体的视觉描述替代抽象概念：

| ❌ 抽象 | ✅ 具体 |
|---------|---------|
| beautiful lighting | golden hour sunlight streaming through venetian blinds |
| sad expression | eyes downcast, slight furrow between brows, lips pressed thin |
| walking slowly | measured deliberate steps, shoulders slightly hunched |
| nice outfit | tailored navy wool coat, white silk scarf, leather gloves |

### 3. 物理真实（Physical Realism）

描述动作时考虑物理规律：

```
✅ "gravity-affected smoke drifting upward from the cup"
✅ "wind-blown hair sweeping across her face"
❌ "hair floating magically"（除非是奇幻风格）
```

### 4. 简洁有力（Concise Power）

- 每个镜头提示词控制在 2~4 句话
- 避免重复 Master Prompt 中已有的信息
- 对话台词越短越好（中文 ≤ 15 字/5秒镜头）

## 多镜头衔接

1. **动作衔接**：上一镜头的结束动作 = 下一镜头的开始状态
2. **视线衔接**：保持视线方向一致（180度法则）
3. **情绪衔接**：情绪变化要有过渡，不要突变
4. **光线衔接**：同一场景内光线保持一致

## 角色一致性

1. 使用元素绑定功能上传参考图
2. Master Prompt 中固定角色的 2~3 个核心视觉特征
3. 每个 Shot 中不要重新描述角色外貌（依赖 Master Prompt）
4. 多镜头模式必须加 `character drift between shots` 负面提示词

## 对话处理

1. 台词使用中文，KLING 3.0 原生支持中文对话效果好
2. 严格遵守时长-字数限制（5秒 ≤ 15字，4秒 ≤ 12字，3秒 ≤ 8字）
3. 语气描述要具体（"低沉沙哑" 比 "伤心" 好）
4. 多人对话用 `Immediately,` 连接

## 运镜建议

1. 情绪镜头：慢推（slow push in）配合特写
2. 动作镜头：跟拍（tracking）或手持（handheld）
3. 建立镜头：稳定器（stabilizer）配合远景/全景
4. 紧张感：低角度（low angle）+ 推镜
5. 压迫感：高角度（high angle）+ 缓慢下降
