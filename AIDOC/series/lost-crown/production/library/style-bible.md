# 《铁血龙魂：失落的王冠》视觉风格指南

## 整体风格

| 属性 | 设定 |
|------|------|
| 画面风格 | 偏写实电影感（参考《权力的游戏》《巫师》） |
| 画面质感 | Cinematic, live-action, photorealistic, shallow depth of field |
| 色彩基调 | 冷峻暗调为主，战斗时有火焰/星辰的暖色对比 |
| 光影风格 | 自然光+戏剧性侧光，夜间场景以月光和火把为主光源 |
| 参考作品 | 《权力的游戏》的写实感 + 《巫师》的黑暗奇幻氛围 |

## 色彩方案

| 场景类型 | 主色调 | 说明 |
|----------|--------|------|
| 王都（序幕） | 冷灰蓝 + 暖金（室内火把） | 表面繁华下的阴冷 |
| 东方废墟 | 苔绿 + 石灰白 + 暗棕 | 古老、荒废、神秘 |
| 北方冰原 | 冰蓝 + 纯白 + 钢灰 | 严酷、孤寂 |
| 战斗场景 | 黑雾紫 + 火焰橙红 + 星辰银白 | 魔法对抗的视觉冲击 |
| 尾声 | 暖金 + 新绿 + 朝阳橙 | 新生与希望 |

## 风格锚定词（每段 Master Prompt 必须包含）

```
Cinematic, photorealistic, dark fantasy, live-action style, shallow depth of field, dramatic lighting
```

## 全局负面提示词

```
Negative: blur, distort, low quality, warping fingers, jittery eyes, character drift between shots, unnatural motion, cartoon style, anime style, 3D render, oversaturated
```

有对话时追加：`audio desync, mouth not matching words`
多角色场景追加：`face swap, character merge, identity drift`

## 对话语言

| 配置项 | 设定 |
|--------|------|
| 对话语言 | **英文**（电影中所有角色对话使用英文） |
| 提示词描述语言 | 英文 |
