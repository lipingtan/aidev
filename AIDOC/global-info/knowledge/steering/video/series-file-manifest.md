# 系列文件清单与生成规范对照表

> 本文件记录一个完整系列所有文件的生成来源、依赖关系和对应规范。
> 用于验证任何系列的文件完整性，以及确认每个文件"由谁在什么时候按什么规范生成"。
> 以 lost-crown 为实例。

---

## 一、foundation/（策划层，Phase 0~3 + 审查产出）

| # | 文件 | 生成步骤 | 对应规范 | 输入依赖 | 说明 |
|---|------|----------|----------|----------|------|
| F1 | `worldview/worldview.md` | Phase 1 | `series-bootstrap-workflow.md` Phase 1 节 | README(类型) + `AIDOC/global-info/knowledge/steering/video/templates/worldview/{类型}.md` | 从模板继承并定制，包含政治/地理/历史/经济/信仰/种族/魔法/核心冲突 |
| F2 | `worldview/timeline.md` | Phase 2 | `series-bootstrap-workflow.md` Phase 2 节 | F1(世界历史) + README(集数/时间跨度) | 故事内时间轴，按月/阶段标注主要事件 |
| F3 | `scripts/大纲.md` | Phase 2 | `series-bootstrap-workflow.md` Phase 2 节 | F1(世界规则) + README(集数) | 整体结构+每幕概述+情绪弧线+伏笔规划 |
| F4 | `scripts/故事细化.md` | Phase 2 | `series-bootstrap-workflow.md` Phase 2 节 | F1 + F3(大纲) | 每幕每章10个节点+心理剖析+关键矛盾+伏笔标注 |
| F5 | `characters/角色设定.md` | Phase 3 | `series-bootstrap-workflow.md` Phase 3 节 | F1(社会位置/种族/能力边界) + F3(叙事功能) + F4(具体出场) | 全角色表格+关系网+敌方单位 |
| F5+ | `characters/角色设定.md` 中的"叙事深化补充"章节 | 叙事审查 | `narrative-quality-review.md` + `climax-design-philosophy.md` + `unpredictability-design.md` | F1~F5全部 | 审查后追加：血脉真相、创伤事件、动机深化、代价重设计、行为约束等 |

---

## 二、production/library/（制作层，Phase 4~5 产出）

### Phase 4 产出（视觉转化，严格按序号顺序生成）

| # | 文件 | 生成步骤 | 对应规范 | 输入依赖 | 说明 |
|---|------|----------|----------|----------|------|
| P1 | `style-bible.md` | Phase 4-1 | `series-bootstrap-workflow.md` Phase 4 步骤A | README(风格偏好) + F1(色调氛围) | 风格锚定词、色彩方案、负面提示词、对话语言 |
| P2 | `characters/*.md` (21个) | Phase 4-2 | `series-bootstrap-workflow.md` Phase 4 步骤A(角色转化) | P1(风格锚定词) + F5(中文设定) + F4(出场场景) | 素体描述(英文)+服装衣橱W1~W5+配饰+语气关键词 |
| P3 | `scenes/*.md` (18个) | Phase 4-3 | `series-bootstrap-workflow.md` Phase 4 步骤A(场景转化) | P1(色彩方案) + F1(地理气候) + F4(场景出现时机) | 场景英文描述+光线设定+环境锚点 |
| P4 | `props-registry.md` | Phase 4-4 | `series-bootstrap-workflow.md` Phase 4 步骤A(道具转化) | P2(角色武器) + P3(环境道具) + F1(魔法体系) + F4(道具出现时机) | 武器/护符/剧情道具/魔法效果视觉描述+出现时间线 |
| P5 | `visual-direction.md` | Phase 4-5 | `series-bootstrap-workflow.md` Phase 4 | P1~P4全部 + F3(伏笔的视觉符号) | 标志性视觉符号+规模感镜头+色彩策略+时间流逝视觉化 |
| P6 | `cinematography-guide.md` | Phase 4-6 | `series-bootstrap-workflow.md` Phase 4 | P5 + F4(场景类型) + F5+审查补充(情感约束) | 按场景类型的镜头模板+浪漫线专属风格+魔法处理+禁忌清单 |
| P7 | `audio-design.md` | Phase 4-7 | `series-bootstrap-workflow.md` Phase 4 | P6 + F5(角色语气) + F4(情绪节奏) | 音乐主题+音效清单+对话语气规范+静默策略 |
| P8 | `production-constraints.md` | Phase 4-8 | `series-bootstrap-workflow.md` Phase 4 | foundation/全部 + P1~P7全部 | 世界观约束+角色行为约束+叙事约束+代价时间线+红鲱鱼清单 |

### Phase 5 产出（制作准备）

| # | 文件 | 生成步骤 | 对应规范 | 输入依赖 | 说明 |
|---|------|----------|----------|----------|------|
| P9 | `series-bible.md` | Phase 5-1 | `series-bootstrap-workflow.md` Phase 5 | F3+F4(伏笔表) + F5+审查补充(叙事规则/代价/称呼) | 叙事铁律+伏笔追踪+称呼规范+一致性检查清单 |
| P10 | `elements-registry.md` | Phase 5-2 | `series-bootstrap-workflow.md` Phase 5 | P2+P3+P4(所有视觉元素) | 角色素体+服装+武器+道具+场景+敌方+魔法效果的KLING平台注册 |
| P11 | `element-priority-guide.md` | Phase 5-3 | `series-bootstrap-workflow.md` Phase 5 | P10(元素清单) + F4(出场频率) | P0~P3优先级标记+创建批次建议 |
| P12 | `reference-generation-guide.md` | Phase 5-4 | `series-bootstrap-workflow.md` Phase 5 | P1(风格要求) + P10(需要生成哪些) | 参考图生成工具+模板+规范+质量检查+批次计划 |
| P13 | `pacing-map-prologue.md` | Phase 5-5 | `series-bootstrap-workflow.md` Phase 5 | F4(故事细化/序幕章节) + P6(时长分配规则) | 情绪曲线+时长分配+段落拆分+关键情绪节点 |

### Phase 5 条件产出

| # | 文件 | 生成步骤 | 触发条件 | 输入依赖 | 说明 |
|---|------|----------|----------|----------|------|
| P14 | `romance-dynamics.md` | Phase 5-7 | 有2+浪漫支线 | F5(角色设定) + F4(情感节点) + 审查补充(差异化根源) | 每条线的独特互动模式+视觉区分+三女互动设计 |
| P15 | `supporting-arcs.md` | Phase 5-8 | 有3+重要配角独立弧线 | F5(配角设定) + F4(配角出场节点) | 配角独立弧线+关键场景+与主线交织规则 |
| P16 | `daily-life-texture.md` | Phase 5-9 | 时间跨度>3月且有日常场景 | F1(日常生活) + F5(角色习惯) + 审查补充(脆弱习惯) | 角色日常习惯+团队非战斗互动+世界活人感+天气季节 |

---

## 三、production/ 其他文件

| # | 文件 | 生成步骤 | 对应规范 | 说明 |
|---|------|----------|----------|------|
| P17 | `production-tracker.md` | Phase 5-6 | `series-bootstrap-workflow.md` Phase 5 | 全局进度看板+资产就绪状态+阻塞项 |
| P18 | `references/README.md` | Phase 0 | `series-bootstrap-workflow.md` Phase 0 | 参考图存放说明 |
| P19 | `001-prologue/story_plan.md` | 制作阶段 | `development-workflow.md` 第1节 | 分集故事策划（进入制作后生成） |
| P20 | `001-prologue/synopsis.md` | 制作阶段 | `development-workflow.md` 第1节 | 分集大纲（进入制作后生成） |
| P21 | `001-prologue/ch01-.../shot_plan.md` | 制作阶段 | `development-workflow.md` 第3节 | 片段方案（进入制作后生成） |

---

## 四、output/assets/（提示词层，Phase 4 同步产出）

| # | 目录/文件 | 生成步骤 | 对应规范 | 输入依赖 | 说明 |
|---|----------|----------|----------|----------|------|
| O1 | `characters/{角色名}-body.md` | Phase 4-2（与P2同步） | `series-bootstrap-workflow.md` Phase 4 步骤B + `reference-generation-guide.md` 角色素体模板 | P2对应角色的素体描述 | 可直接粘贴到图片生成工具的提示词 |
| O2 | `characters/{角色名}-outfit-{场景}.md` | Phase 4-2（与P2同步） | 同上，服装模板 | P2对应角色的Wardrobe条目 | 每套服装一个文件 |
| O3 | `scenes/{场景名}.md` | Phase 4-3（与P3同步） | `series-bootstrap-workflow.md` Phase 4 步骤B + `reference-generation-guide.md` 场景模板 | P3对应场景描述 | 16:9无人场景提示词 |
| O4 | `props/{道具名}.md` | Phase 4-4（与P4同步） | 同上，道具模板 | P4对应道具描述 | 产品摄影式提示词 |
| O5 | `enemies/{单位名}.md` | Phase 4-4（与P4同步） | 同上，敌方单位模板 | P4中enemies部分 | 威胁姿态+暗色环境 |

---

## 五、output/kling/（视频提示词，制作阶段产出）

| # | 文件 | 生成步骤 | 对应规范 | 输入依赖 |
|---|------|----------|----------|----------|
| V1 | `{NNN}-{集}/{章}/ep-{NNN}.md` | 制作阶段 | `development-workflow.md` 第3节 + `prompt-engineering.md` + `tech.md` | P1(风格) + P2(角色) + P3(场景) + P4(道具) + P8(约束) + shot_design |

---

## 六、规范文件索引（生成上述文件时需要参考的规范）

| 规范文件 | 位置 | 影响哪些文件的生成 |
|----------|------|-------------------|
| `series-bootstrap-workflow.md` | `global-info/knowledge/steering/` | 所有 foundation/ 和 production/library/ 文件 |
| `narrative-quality-review.md` | 同上 | F5+审查补充 → 影响 P6/P7/P8/P9/P14/P16 |
| `climax-design-philosophy.md` | 同上 | F4故事细化中的高潮幕 → 影响 P5/P6/P8 |
| `unpredictability-design.md` | 同上 | F3~F5中的不可预测性设计 → 影响 P8/P9 |
| `development-workflow.md` | 同上 | P19~P21 + V1（制作阶段所有文件） |
| `prompt-engineering.md` | 同上 | O1~O5 + V1（所有提示词文件） |
| `tech.md` | 同上 | V1（视频提示词格式） |
| `structure.md` | 同上 | 所有文件的命名和存放位置 |

---

## 七、当前 lost-crown 状态

| 层 | 应有文件数 | 实际文件数 | 状态 |
|----|-----------|-----------|------|
| foundation/ | 5 | 5 | ✅ 完整 |
| production/library/ 独立文件 | 16 | 16 | ✅ 完整 |
| production/library/characters/ | 21 | 21 | ✅ 完整 |
| production/library/scenes/ | 18 | 18 | ✅ 完整 |
| production/ 其他 | 2(tracker+refs README) | 2 | ✅ 完整 |
| output/assets/ | ~100+(角色×服装+场景+道具+敌方) | 0 | ⚠️ 待生成（开拍前按P12指南批量生成） |
| output/kling/ | 按需 | 1(README) | ⚠️ 制作阶段逐步生成 |

---

## 八、文件生成的完整时序图

```
Phase 0 → README.md + 目录骨架
    ↓
Phase 1 → F1 worldview.md
    ↓
Phase 2 → F2 timeline + F3 大纲 + F4 故事细化
    ↓
Phase 3 → F5 角色设定
    ↓
★ 叙事审查 → F5 追加"叙事深化补充"章节
    ↓
Phase 4-1 → P1 style-bible
Phase 4-2 → P2 characters/*.md + O1/O2 assets/characters/*.md
Phase 4-3 → P3 scenes/*.md + O3 assets/scenes/*.md
Phase 4-4 → P4 props-registry + O4/O5 assets/props+enemies/*.md
Phase 4-5 → P5 visual-direction
Phase 4-6 → P6 cinematography-guide
Phase 4-7 → P7 audio-design
Phase 4-8 → P8 production-constraints
    ↓
Phase 5-1 → P9 series-bible
Phase 5-2 → P10 elements-registry
Phase 5-3 → P11 element-priority-guide
Phase 5-4 → P12 reference-generation-guide
Phase 5-5 → P13 pacing-map
Phase 5-6 → P17 production-tracker
Phase 5-7~9 → P14/P15/P16 (条件文件)
    ↓
★ 启动完成检查 → 全部通过
    ↓
制作阶段 → P19~P21 + V1（逐集逐章逐段）
```
