# 新系列启动工作流（Series Bootstrap）

> 当从零开始一个全新系列时，按本流程依次生成所有前置基础资产。
> 确保每个新系列启动时都有完整、一致的基础设定，不依赖个人经验。

---

## 一、流程总览

```
用户输入：粗略创意/灵感/参考
    │
    ▼
Phase 0: 系列初始化（创建目录 + 确认基本参数）
    │
    ▼
Phase 1: 世界观构建（worldview + 魔法/科技体系）
    │
    ▼
Phase 2: 故事架构（大纲 + 时间线 + 故事细化）
    │
    ▼
Phase 3: 角色设计（全角色设定 + 关系网）
    │
    ▼
★ 叙事质量审查（narrative-quality-review.md）★
    │
    ▼
Phase 4: 视觉转化（foundation → production/library → output/assets）
    │
    ▼
Phase 5: 制作准备（元素注册 + 节奏图 + 系列圣经）
    │
    ▼
基础资产就绪，可进入制作流程（development-workflow.md）
```

**核心原则**：
- 每个 Phase 完成后必须用户确认才能进入下一个
- 后一个 Phase 依赖前一个 Phase 的产出
- 每个 Phase 内部遵循 plan → 澄清 → 确认 → 产出 的门控机制
- **每个 Phase 执行时遵循执行协议**（`execution-protocol.md`）：PRE-CHECK → EXECUTE → POST-CHECK

---

## 二、Phase 0: 系列初始化

### 目标

确认基本参数，创建目录骨架。

### 澄清问题（必问）

| # | 问题 | 选项/说明 |
|---|------|----------|
| 1 | 系列名称？ | kebab-case |
| 2 | 类型/题材？ | 中世纪奇幻/科幻/现代都市/其他 |
| 3 | 预计集数和每集时长？ | |
| 4 | 目标平台？ | 横屏/竖屏/方形 |
| 5 | 画面风格偏好？ | 写实/动漫/混合 |
| 6 | 是否有已有素材？ | |
| 7 | 默认视频引擎？ | KLING AI / WAN 2.5 |

### 产出

创建目录骨架：
```
AIDOC/series/{系列名}/
├── README.md
├── foundation/
│   ├── worldview/
│   ├── characters/
│   └── scripts/
└── production/
    ├── library/
    │   ├── characters/
    │   └── scenes/
    └── references/
        ├── characters/
        ├── scenes/
        ├── style/
        └── props/

output/{系列名}/
├── assets/
│   ├── characters/
│   ├── scenes/
│   ├── props/
│   └── enemies/
└── kling/
```

---

## 三、Phase 1: 世界观构建

### 目标

建立故事发生的世界规则。

### 输入依赖

- Phase 0 确认的类型/题材（README.md）
- 用户提供的素材（如有）
- 对应的世界观模板（`global-info/templates/worldview/{类型}.md`）

### 产出

`foundation/worldview/worldview.md`

### 必须包含

| 模块 | 是否必须 |
|------|----------|
| 政治/社会结构 | 是 |
| 地理格局 | 是 |
| 历史纪年 | 是 |
| 经济与日常 | 是 |
| 信仰/宗教体系 | 如有 |
| 种族关系 | 如有 |
| 魔法/科技体系 | 如有 |
| 核心冲突 | 是 |

---

## 四、Phase 2: 故事架构

### 目标

确定完整的故事骨架。

### 输入依赖

- `foundation/worldview/worldview.md`（世界规则决定故事的可能性边界）
- Phase 0 确认的集数/时长

### 产出

| 文件 | 路径 |
|------|------|
| 故事大纲 | `foundation/scripts/大纲.md` |
| 时间线 | `foundation/worldview/timeline.md` |
| 故事细化 | `foundation/scripts/故事细化.md` |

### 大纲必须包含

- 整体结构（几集/几幕）
- 每幕概述
- 情绪弧线
- 冲突起伏节奏
- 伏笔规划

### 故事细化必须包含

- 每章10个节点
- 每章心理剖析
- 每章关键矛盾
- 伏笔标注

---

## 五、Phase 3: 角色设计

### 目标

设计所有角色的完整设定。

### 输入依赖

- `foundation/worldview/worldview.md`（角色的社会位置、种族、魔法能力边界）
- `foundation/scripts/大纲.md`（角色的叙事功能）
- `foundation/scripts/故事细化.md`（角色的具体出场和行为）

### 产出

`foundation/characters/角色设定.md`

### 每个角色必须包含

体貌特征、个性、信仰/派系、成长历程、首次出场、主要故事线、关键使命、最后归属、角色关系、核心技能（如有）、核心矛盾（主角必须）。

---

## 六、Phase 4: 视觉转化

### 前置条件

**必须先通过叙事质量审查**（参见 `narrative-quality-review.md`）。
审查在 Phase 3 完成后自动触发，通过后才能进入本阶段。

**审查产出归属**：
- 审查中发现的问题修复直接更新到 foundation/ 对应文件中（不产生独立文件）
- 修复内容以"补充章节"形式追加到原文件末尾（如"叙事深化补充"、"不可预测性补充"）
- 修复完成后在 foundation/ 对应文件中标注"已通过叙事质量审查"

### 目标

将策划层设定（foundation/）转化为制作资产（production/library/），并生成资产提示词（output/assets/）。

### Phase 4 产出顺序（严格按此顺序）

| 序号 | 文件 | 输入依赖（必须读取） |
|------|------|---------------------|
| 1 | style-bible.md | Phase 0 README（风格偏好）+ worldview（色调氛围） |
| 2 | characters/*.md | style-bible（风格锚定词）+ foundation/characters（中文设定）+ foundation/scripts（出场场景） |
| 3 | scenes/*.md | style-bible（色彩方案）+ worldview（地理气候）+ foundation/scripts（场景出现时机） |
| 4 | props-registry.md | characters/*.md（角色武器）+ scenes/*.md（环境道具）+ worldview（魔法体系）+ foundation/scripts（道具出现时机） |
| 5 | visual-direction.md | 以上全部 + foundation/scripts/大纲（伏笔的视觉符号）|
| 6 | cinematography-guide.md | visual-direction + foundation/scripts/故事细化（场景类型决定镜头模板）+ 叙事审查补充（情感表达约束） |
| 7 | audio-design.md | cinematography-guide + foundation/characters（角色语气）+ foundation/scripts（情绪节奏） |
| 8 | production-constraints.md | foundation/ 全部文件 + 以上所有 library 文件（提取所有影响制作的约束） |

> 每完成一个文件，立即生成对应的 output/assets/ 提示词（如完成 characters/alan.md 后立即生成 output/assets/characters/alan-body.md）。

### 核心链路

```
foundation/（策划层，中文）
    │
    ▼ 步骤A：提取视觉要素，翻译为英文
    │
production/library/（设计层，中英混合）
    │
    ▼ 步骤B：按模板组装为可直接使用的提示词
    │
output/assets/（提示词层，纯英文）
```

### 步骤A：策划层 → 设计层

**角色转化**：

| foundation/ 字段 | → production/library/ 位置 |
|-----------------|---------------------------|
| 体貌特征（中文） | library/characters/{名}.md → 素体描述（英文） |
| 服装描述 | library/characters/{名}.md → Wardrobe W1~W5 |
| 技能/魔法 | library/characters/{名}.md → 配饰/道具 |
| 个性/语气 | library/characters/{名}.md → 语气关键词 |

**场景转化**：

| foundation/ 来源 | → production/library/ 位置 |
|-----------------|---------------------------|
| worldview 地理描述 | library/scenes/{名}.md → 场景描述（英文） |
| worldview 气候 | library/scenes/{名}.md → 光线设定 |
| 故事中的建筑描述 | library/scenes/{名}.md → 环境锚点 |

**道具转化**：

| foundation/ 来源 | → production/library/ 位置 |
|-----------------|---------------------------|
| 角色武器/物品 | library/props-registry.md → 武器类 |
| 魔法效果 | library/props-registry.md → 魔法效果 |
| 关键剧情道具 | library/props-registry.md → 剧情道具 |

### 步骤B：设计层 → 提示词层

按 `reference-generation-guide.md` 的模板，将设计文档组装为 output/assets/ 中的提示词文件。

**角色素体提示词模板**：
```
[素体描述]
neutral expression, looking directly at camera,
simple grey background, soft even lighting,
photorealistic, high detail, cinematic quality, upper body shot

Negative: blur, distort, low quality, anime style, cartoon,
oversaturated, busy background, text, watermark
```

**场景提示词模板**：
```
[场景描述]
no people, empty scene, establishing shot,
cinematic, photorealistic, dramatic lighting, 16:9 aspect ratio

Negative: people, characters, figures, anime style, low quality
```

**道具提示词模板**：
```
[道具描述]
product photography style, centered composition,
dark background, dramatic lighting highlighting details,
photorealistic, high detail, no hands, isolated object

Negative: blur, low quality, hands, people, busy background
```

### 产出文件清单

**设计层（production/library/）**：
1. style-bible.md
2. characters/*.md（每个角色一个文件）
3. scenes/*.md（每个场景一个文件）
4. props-registry.md
5. visual-direction.md
6. cinematography-guide.md
7. audio-design.md
8. **production-constraints.md**（从 foundation 提取的制作约束摘要）

**提示词层（output/assets/）**：
9. characters/{角色名}-body.md
10. characters/{角色名}-outfit-{场景}.md
11. scenes/{场景名}.md
12. props/{道具名}.md
13. enemies/{单位名}.md

### 转化质量检查

- [ ] foundation/ 中每个角色都有对应的 library/characters/*.md
- [ ] 每个 library/characters/*.md 都有对应的 output/assets/characters/*.md
- [ ] 故事中所有场景都有 library/scenes/*.md + output/assets/scenes/*.md
- [ ] 所有道具都在 props-registry 中且有 output/assets/props/*.md
- [ ] 英文描述与中文原始设定语义一致
- [ ] 提示词符合 style-bible 风格锚定词

---

## 七、Phase 5: 制作准备

### 目标

生成制作管理和约束文件。

### 输入依赖

所有 Phase 5 文件都依赖 Phase 4 的完整产出 + foundation/ 全部内容。具体：

| 文件 | 特别依赖 |
|------|----------|
| series-bible | foundation/scripts（伏笔表）+ foundation/characters（叙事规则、代价时间线、称呼规范）+ 叙事审查补充 |
| elements-registry | production/library/characters + scenes + props-registry（所有需要视觉一致性的元素） |
| element-priority-guide | elements-registry（标记优先级）+ foundation/scripts（出场频率） |
| reference-generation-guide | style-bible（风格要求）+ elements-registry（需要生成哪些参考图） |
| pacing-map | foundation/scripts/故事细化（每章节点）+ cinematography-guide（时长分配规则） |
| romance-dynamics | foundation/characters（三女设定）+ foundation/scripts（情感场景节点）+ 叙事审查补充（差异化根源） |
| supporting-arcs | foundation/characters（配角设定）+ foundation/scripts（配角出场节点） |
| daily-life-texture | foundation/worldview（日常生活）+ foundation/characters（独处习惯）+ 叙事审查补充（脆弱习惯） |

### 产出（按顺序生成）

**必须生成**：

| 序号 | 文件 | 路径 | 触发条件 |
|------|------|------|----------|
| 1 | 系列圣经 | `production/library/series-bible.md` | 始终 |
| 2 | 元素注册表 | `production/library/elements-registry.md` | 始终 |
| 3 | 元素优先级 | `production/library/element-priority-guide.md` | 始终 |
| 4 | 参考图生成指南 | `production/library/reference-generation-guide.md` | 始终 |
| 5 | 第一集节奏图 | `production/library/pacing-map-{集名}.md` | 始终 |
| 6 | 制作进度追踪 | `production/production-tracker.md` | 始终 |

**条件生成（根据故事类型自动判断）**：

| 序号 | 文件 | 路径 | 触发条件 |
|------|------|------|----------|
| 7 | 浪漫线设计 | `production/library/romance-dynamics.md` | 故事中有2条以上浪漫支线 |
| 8 | 配角弧线 | `production/library/supporting-arcs.md` | 有3个以上重要配角且有独立弧线 |
| 9 | 日常质感 | `production/library/daily-life-texture.md` | 故事时间跨度 > 3个月且有日常场景 |

**条件判断规则**：在 Phase 5 开始时，根据以下问题自动判断哪些条件文件需要生成：

- 故事中是否有多条浪漫/情感支线？→ 是 → 生成 romance-dynamics.md
- 是否有3个以上有独立成长弧线的配角？→ 是 → 生成 supporting-arcs.md
- 故事时间跨度是否超过3个月？是否有日常/非战斗场景？→ 是 → 生成 daily-life-texture.md

> 不需要用户判断——AI 根据 foundation/ 中的故事内容自动判断并生成。

---

## 八、启动完成检查清单

### 文件完整性

**foundation/（必须全部存在）**：
- [ ] `foundation/worldview/worldview.md` 存在且标注"已通过叙事质量审查"
- [ ] `foundation/worldview/timeline.md` 存在
- [ ] `foundation/scripts/` 下有大纲和故事细化
- [ ] `foundation/characters/` 下有完整角色设定（含叙事深化补充）

**production/library/（必须全部存在）**：
- [ ] `style-bible.md`
- [ ] `characters/` 下每个角色都有视觉档案
- [ ] `scenes/` 下每个主要场景都有设定
- [ ] `props-registry.md`
- [ ] `visual-direction.md`
- [ ] `cinematography-guide.md`
- [ ] `audio-design.md`
- [ ] `production-constraints.md`
- [ ] `series-bible.md`
- [ ] `elements-registry.md`
- [ ] `element-priority-guide.md`
- [ ] `reference-generation-guide.md`
- [ ] `pacing-map-{第一集}.md`

**production/（必须存在）**：
- [ ] `production-tracker.md`

**条件文件（根据触发条件判断）**：
- [ ] 如有多条浪漫线 → `romance-dynamics.md` 存在
- [ ] 如有重要配角独立弧线 → `supporting-arcs.md` 存在
- [ ] 如有日常场景 → `daily-life-texture.md` 存在

**output/assets/（必须与 library 对应）**：
- [ ] `characters/` 下每个角色有素体+服装提示词
- [ ] `scenes/` 下每个场景有提示词
- [ ] `props/` 下每个道具有提示词
- [ ] `enemies/` 下每个敌方单位有提示词（如有）
- [ ] `production/library/style-bible.md` 存在
- [ ] `production/library/characters/` 下每个角色有视觉档案
- [ ] `production/library/scenes/` 下每个场景有设定
- [ ] `production/library/props-registry.md` 存在
- [ ] `production/library/series-bible.md` 存在
- [ ] `production/library/elements-registry.md` 存在
- [ ] `output/{系列}/assets/` 下有对应的提示词文件
- [ ] `production/production-tracker.md` 存在

### 一致性检查

- [ ] 角色数量：foundation/ = library/characters/ = output/assets/characters/
- [ ] 场景数量：故事提及 = library/scenes/ = output/assets/scenes/
- [ ] 道具数量：props-registry = output/assets/props/
- [ ] 时间线与大纲一致
- [ ] 魔法体系规则在各文件中一致

---

## 九、快捷模式

| 用户已有 | 可跳过 | 从哪里开始 |
|----------|--------|-----------|
| 一句话创意 | 无 | Phase 0 |
| 完整剧本 | Phase 2 大部分 | Phase 0 → 1 → 提取2 → 3 |
| 剧本+角色 | Phase 2+3 大部分 | Phase 0 → 1 → 验证 → 4 |
| 剧本+角色+视觉参考 | Phase 2~4 大部分 | Phase 0 → 1 → 验证 → 5 |

**即使跳过步骤，最终检查清单必须全部通过。**

---

## 十、Phase 依赖关系

```
Phase 0 → Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5
                                  ↑可部分并行↑
```

---

## 十一、时间预估

| Phase | 轮次 |
|-------|------|
| 0 初始化 | 1 |
| 1 世界观 | 2~3 |
| 2 故事 | 3~5 |
| 3 角色 | 2~3 |
| 4 视觉转化 | 2~3 |
| 5 制作准备 | 1~2 |
| **总计** | **12~17** |
