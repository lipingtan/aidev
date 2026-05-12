# 叙事工作流规范

## 概述

本文档定义了 RPG 游戏叙事策划的完整工作流，从世界观构建到对话脚本落地，确保叙事内容与玩法系统深度集成。

---

## 一、RPG 叙事策划流程

### 1.1 五阶段流程

```
世界观构建 → 剧情大纲 → 支线设计 → 对话编写 → 触发条件配置
    │            │           │          │           │
    ▼            ▼           ▼          ▼           ▼
 worldview.md  outline.md  quests/   dialogues/  triggers配置
```

### 1.2 阶段详细定义

#### 阶段一：世界观构建

**输入**：游戏类型、美术风格、核心体验
**产出**：`AIDOC/projects/{游戏名}/design/worldview.md`

| 章节 | 内容要求 | 完成标准 |
|------|----------|----------|
| 时代背景 | 历史纪元、当前时代特征、科技/魔法水平 | 能回答"这个世界处于什么发展阶段" |
| 地理环境 | 大陆/区域划分、气候、地标 | 有可绘制地图的信息量 |
| 势力阵营 | 主要势力、关系网（同盟/敌对/中立）| 至少 3 个势力，关系有冲突张力 |
| 魔法/科技体系 | 能力来源、使用规则、限制条件 | 规则自洽，无逻辑漏洞 |
| 社会结构 | 阶层、经济、文化习俗 | 能支撑 NPC 行为逻辑 |
| 历史事件 | 影响当前局势的关键历史 | 至少 3 个历史事件，与主线相关 |

#### 阶段二：剧情大纲

**输入**：世界观文档、主角设定
**产出**：`AIDOC/projects/{游戏名}/narrative/outline.md`

| 章节 | 内容要求 | 完成标准 |
|------|----------|----------|
| 主题 | 故事核心主题（1-2个词） | 贯穿全篇的情感/哲学主题 |
| 三幕结构 | 起承转合的大框架 | 每幕有明确的开始/结束事件 |
| 主线节点 | 关键剧情点（8-15个） | 每个节点有：事件、地点、参与角色、玩家选择 |
| 转折点 | 重大反转（2-3个） | 有伏笔铺垫，不突兀 |
| 结局 | 可能的结局（1-3个） | 与玩家选择关联，情感满足 |
| 角色弧光 | 主要角色的成长变化 | 每个重要角色有起点→终点的变化 |

#### 阶段三：支线设计

**输入**：剧情大纲、世界观、区域设计
**产出**：`AIDOC/projects/{游戏名}/narrative/quests/side/`

| 支线类型 | 设计要求 | 与主线关系 |
|----------|----------|-----------|
| 角色支线 | 深化同伴/NPC 背景故事 | 可解锁主线中的额外选项 |
| 区域支线 | 探索该区域的历史/秘密 | 丰富世界观，提供背景信息 |
| 势力支线 | 影响势力关系和声望 | 可能改变主线中的势力态度 |
| 收集支线 | 收集物品/信息 | 提供资源奖励，可选 |
| 挑战支线 | 战斗/解谜挑战 | 提供装备/技能奖励 |

#### 阶段四：对话编写

**输入**：支线设计、角色设定、场景上下文
**产出**：`AIDOC/projects/{游戏名}/narrative/dialogues/`

对话编写规则：
1. 每段对话有明确的叙事目的（推进剧情/揭示信息/建立关系）
2. 对话风格与角色性格一致
3. 分支选项有实质性差异（不是换个说法表达同一意思）
4. 重要选择有后果标记

#### 阶段五：触发条件配置

**输入**：对话脚本、任务设计、关卡布局
**产出**：触发器配置数据（JSON/tres）

触发条件类型：
- 位置触发（进入区域）
- 交互触发（与 NPC/物品交互）
- 条件触发（完成前置任务/达到属性阈值）
- 时间触发（游戏内时间/现实时间）
- 战斗触发（击败特定敌人/血量阈值）

---

## 二、对话数据 JSON 格式

### 2.1 对话节点格式

```json
{
  "dialogue_id": "npc_elder_first_meet",
  "metadata": {
    "speaker": "village_elder",
    "location": "village_square",
    "chapter": 1,
    "priority": "main_quest",
    "tags": ["introduction", "world_building"]
  },
  "preconditions": {
    "quest_state": { "main_quest_01": "active" },
    "player_level": { "min": 1 },
    "flags": ["entered_village_first_time"],
    "not_flags": ["elder_already_spoken"]
  },
  "nodes": [
    {
      "id": "node_01",
      "type": "dialogue",
      "speaker": "village_elder",
      "text": "年轻人，你终于来了。我等这一天已经很久了。",
      "emotion": "relieved",
      "animation": "gesture_welcome",
      "voice_id": "elder_line_001",
      "next": "node_02"
    },
    {
      "id": "node_02",
      "type": "choice",
      "speaker": "player",
      "prompt": "你要如何回应？",
      "choices": [
        {
          "text": "你认识我？",
          "emotion": "curious",
          "next": "node_03a",
          "effects": {
            "relationship": { "village_elder": +5 },
            "set_flag": "player_curious"
          }
        },
        {
          "text": "我只是路过的旅人。",
          "emotion": "neutral",
          "next": "node_03b",
          "effects": {
            "relationship": { "village_elder": -2 }
          }
        },
        {
          "text": "[力量≥10] 少废话，有什么事快说。",
          "emotion": "aggressive",
          "condition": { "attribute": { "strength": { "min": 10 } } },
          "next": "node_03c",
          "effects": {
            "relationship": { "village_elder": -10 },
            "set_flag": "player_aggressive"
          }
        }
      ]
    },
    {
      "id": "node_03a",
      "type": "dialogue",
      "speaker": "village_elder",
      "text": "预言中提到过你的到来。来，让我告诉你这片土地的故事。",
      "emotion": "warm",
      "next": "node_04",
      "effects": {
        "unlock_lore": "prophecy_of_hero",
        "set_flag": "elder_told_prophecy"
      }
    },
    {
      "id": "node_04",
      "type": "narration",
      "text": "长老带你走向村庄中央的古老石碑...",
      "camera": "pan_to_monument",
      "next": "end",
      "effects": {
        "quest_update": { "main_quest_01": "spoke_to_elder" },
        "set_flag": "elder_already_spoken"
      }
    }
  ],
  "on_complete": {
    "set_flags": ["elder_conversation_done"],
    "quest_progress": { "main_quest_01": "next_step" },
    "unlock_area": "elder_house"
  }
}
```

### 2.2 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| dialogue_id | string | 唯一标识符，格式：{npc}_{场景}_{序号} |
| emotion | enum | 情感标记：neutral/happy/sad/angry/surprised/fearful/relieved/curious/aggressive/warm |
| animation | string | 对应的动画名称 |
| condition | object | 显示该选项的前置条件 |
| effects | object | 选择后的即时效果 |
| next | string | 下一个节点 ID，"end" 表示对话结束 |
| type | enum | 节点类型：dialogue/choice/narration/event |

---

## 三、任务数据格式

### 3.1 任务 JSON 格式示例

```json
{
  "quest_id": "side_quest_blacksmith_sword",
  "metadata": {
    "name": "失落的传承",
    "type": "side_quest",
    "category": "character",
    "chapter": 1,
    "region": "starter_village",
    "estimated_time": "15min",
    "difficulty": "easy",
    "repeatable": false
  },
  "description": {
    "brief": "帮助铁匠找回祖传的锻造图纸",
    "full": "村里的铁匠告诉你，他祖传的锻造图纸被盗贼偷走了。那些盗贼藏身在村外的废弃矿洞中。",
    "completion": "你找回了图纸，铁匠感激地为你打造了一把特殊的武器。"
  },
  "preconditions": {
    "player_level": { "min": 3 },
    "quests_completed": ["main_quest_01"],
    "flags": ["entered_village", "talked_to_blacksmith"],
    "not_flags": ["blacksmith_quest_failed"],
    "relationship": { "blacksmith": { "min": 0 } }
  },
  "objectives": [
    {
      "id": "obj_01",
      "type": "talk",
      "target": "blacksmith_npc",
      "description": "与铁匠交谈了解情况",
      "optional": false,
      "triggers_dialogue": "blacksmith_quest_start"
    },
    {
      "id": "obj_02",
      "type": "go_to",
      "target": "abandoned_mine_entrance",
      "description": "前往废弃矿洞",
      "optional": false,
      "waypoint": true
    },
    {
      "id": "obj_03",
      "type": "kill",
      "target": "bandit_thief",
      "count": 3,
      "description": "击败矿洞中的盗贼 (0/3)",
      "optional": false
    },
    {
      "id": "obj_04_a",
      "type": "collect",
      "target": "forging_blueprint",
      "count": 1,
      "description": "找到锻造图纸",
      "optional": false,
      "location_hint": "盗贼首领身上"
    },
    {
      "id": "obj_04_b",
      "type": "collect",
      "target": "rare_ore",
      "count": 5,
      "description": "[可选] 收集稀有矿石 (0/5)",
      "optional": true,
      "bonus": true
    },
    {
      "id": "obj_05",
      "type": "talk",
      "target": "blacksmith_npc",
      "description": "将图纸交还铁匠",
      "optional": false,
      "triggers_dialogue": "blacksmith_quest_complete"
    }
  ],
  "branches": {
    "branch_merciful": {
      "condition": { "flag": "spared_bandit_leader" },
      "description": "你饶恕了盗贼首领",
      "effects": {
        "relationship": { "bandit_faction": +20 },
        "unlock_quest": "side_quest_bandit_redemption"
      }
    },
    "branch_ruthless": {
      "condition": { "flag": "killed_bandit_leader" },
      "description": "你消灭了盗贼首领",
      "effects": {
        "relationship": { "village": +10, "bandit_faction": -30 },
        "loot": "bandit_leader_drops"
      }
    }
  },
  "rewards": {
    "base": {
      "experience": 150,
      "gold": 50,
      "items": [
        { "id": "blacksmith_sword", "type": "weapon", "quality": "uncommon" }
      ],
      "relationship": { "blacksmith": +15 }
    },
    "bonus": {
      "condition": "obj_04_b_completed",
      "experience": 50,
      "items": [
        { "id": "reinforced_blacksmith_sword", "type": "weapon", "quality": "rare" }
      ]
    }
  },
  "failure_conditions": {
    "blacksmith_dies": {
      "trigger": "npc_death:blacksmith_npc",
      "result": "quest_failed",
      "set_flag": "blacksmith_quest_failed"
    },
    "timeout": null
  },
  "on_complete": {
    "set_flags": ["blacksmith_quest_done", "has_blacksmith_sword"],
    "unlock_shop": "blacksmith_advanced_items",
    "unlock_quest": "side_quest_blacksmith_masterwork"
  }
}
```

### 3.2 任务字段说明

| 字段 | 说明 |
|------|------|
| objectives[].type | 目标类型：talk/go_to/kill/collect/interact/escort/defend/craft |
| objectives[].optional | 是否可选目标（影响奖励等级） |
| branches | 任务分支（基于玩家选择产生不同结果） |
| rewards.base | 基础奖励（完成必做目标） |
| rewards.bonus | 额外奖励（完成可选目标） |
| failure_conditions | 任务失败条件 |

---

## 四、叙事与玩法集成点

### 4.1 集成点定义

| 集成类型 | 触发方式 | 叙事表现 | 技术实现 |
|----------|----------|----------|----------|
| **触发器叙事** | 玩家进入区域/交互物品 | 对话、过场、旁白 | Area3D + signal → DialogueManager |
| **环境叙事** | 玩家观察环境 | 场景细节、可检查物品、环境音 | InteractableObject + inspect_text |
| **物品叙事** | 获得/使用物品 | 物品描述、日记、信件 | ItemData.lore_text + LoreUI |
| **NPC 行为叙事** | NPC 日常行为 | NPC 按时间表活动、对话变化 | AISchedule + context_dialogue |
| **战斗叙事** | 战斗中/战斗后 | Boss 台词、战斗评价、剧情触发 | CombatEvent + narrative_trigger |
| **成就叙事** | 达成特定条件 | 解锁背景故事、隐藏剧情 | AchievementSystem + lore_unlock |

### 4.2 触发器设计规范

```json
{
  "trigger_id": "env_story_ancient_mural",
  "type": "interaction",
  "target_node": "ancient_mural_01",
  "activation": {
    "method": "interact_button",
    "range": 2.0,
    "prompt_text": "观察壁画"
  },
  "conditions": {
    "first_time_only": false,
    "required_item": null,
    "required_skill": { "perception": 5 }
  },
  "narrative_action": {
    "type": "inspect_text",
    "content": "壁画描绘了一场远古战争...",
    "unlock_lore": "ancient_war_01",
    "camera_focus": "mural_closeup"
  }
}
```

### 4.3 环境叙事层级

| 层级 | 发现难度 | 示例 | 奖励 |
|------|----------|------|------|
| 表层 | 路径上必经 | 路边的告示牌、NPC 闲聊 | 基础信息 |
| 中层 | 需要探索 | 隐藏房间的日记、可破坏墙壁后的壁画 | 背景故事 + 少量经验 |
| 深层 | 需要条件 | 特定技能才能解读的符文、组合多个线索 | 重要剧情 + 稀有奖励 |

### 4.4 NPC 行为与叙事联动

```
NPC 日程表 → 不同时间在不同地点
    → 不同地点有不同对话内容
    → 特定事件后行为模式改变
    → 关系值影响对话态度和可用选项
```

---

## 五、角色一致性管理规则

### 5.1 角色档案模板

每个重要角色必须有完整档案：

```markdown
## 角色档案：{角色名}

### 基础信息
- 全名：
- 年龄：
- 种族/职业：
- 所属势力：

### 性格特征
- 核心特质（3个词）：
- MBTI 参考：
- 说话风格：（正式/随意/粗鲁/文雅/...）
- 口头禅/语癖：
- 禁忌话题：

### 动机与目标
- 表面目标：
- 深层动机：
- 恐惧/弱点：

### 关系网
- 与主角关系：
- 与其他角色关系：
- 态度变化条件：

### 对话规则
- 称呼主角方式：
- 情绪表达方式：
- 知识范围（知道什么/不知道什么）：
- 绝对不会说的话：
```

### 5.2 一致性检查规则

| 检查项 | 规则 | 违规示例 |
|--------|------|----------|
| 语言风格 | 同一角色全程保持一致的说话方式 | 粗犷战士突然说文言文 |
| 知识边界 | 角色不能知道其不应知道的信息 | 村民知道远方王国的宫廷秘密 |
| 动机一致 | 行为必须符合角色动机 | 贪婪商人无缘无故免费赠送 |
| 情感连续 | 情绪变化需要合理触发 | 刚失去亲人的角色下一句开玩笑 |
| 能力边界 | 角色行为不超出其能力设定 | 普通农民突然展示高超剑术 |
| 关系逻辑 | 角色间互动符合关系设定 | 死敌突然亲密无间 |

### 5.3 对话生成时的一致性约束

生成对话前必须确认：
1. 读取该角色的完整档案
2. 确认当前剧情阶段（角色此时知道什么）
3. 确认与对话对象的关系状态
4. 确认角色当前情绪状态（受前序事件影响）
5. 检查是否有伏笔需要在此处铺设/回收

---

## 六、伏笔追踪表

### 6.1 伏笔追踪格式

```markdown
## 伏笔追踪表

| 伏笔ID | 铺设位置 | 铺设方式 | 回收位置 | 回收方式 | 状态 | 重要度 |
|---------|----------|----------|----------|----------|------|--------|
| FS-001 | 第1章-村庄 | 长老提到"预言" | 第3章-神殿 | 预言石碑完整内容 | 已铺设 | 主线 |
| FS-002 | 第1章-矿洞 | 墙壁上的奇怪符号 | 第4章-遗迹 | 符号是古代封印的一部分 | 已铺设 | 支线 |
| FS-003 | 第2章-酒馆 | 旅人提到南方的异变 | 第5章-南方 | 南方是最终Boss的据点 | 计划中 | 主线 |
```

### 6.2 伏笔管理规则

| 规则 | 说明 |
|------|------|
| 铺设密度 | 每章至少 2 个伏笔铺设点，避免集中爆发 |
| 回收时机 | 铺设后 1-3 章内必须回收，避免玩家遗忘 |
| 多次暗示 | 重要伏笔至少暗示 2 次再正式揭示 |
| 可发现性 | 主线伏笔必须在主线路径上，支线伏笔可隐藏 |
| 自洽性 | 回收时的解释必须与铺设时的细节完全吻合 |
| 层次感 | 表层伏笔（1章回收）+ 中层（2-3章）+ 深层（跨幕） |

### 6.3 伏笔状态枚举

| 状态 | 说明 |
|------|------|
| 计划中 | 已设计但尚未写入游戏内容 |
| 已铺设 | 已在对话/环境/物品中植入 |
| 已暗示 | 已有第二次或更多次暗示 |
| 已回收 | 已正式揭示/解释 |
| 已废弃 | 设计变更导致不再使用（需确保无残留线索） |

### 6.4 伏笔与技术实现的对应

| 伏笔载体 | 技术实现 | 数据存储 |
|----------|----------|----------|
| 对话中的暗示 | dialogue_node.foreshadow_id | 对话 JSON |
| 环境中的线索 | InspectableObject.lore_id | 场景节点属性 |
| 物品描述 | ItemData.lore_text | .tres 资源 |
| NPC 行为异常 | AISchedule.special_behavior | AI 配置 |
| 场景变化 | LevelState.visual_change | 存档数据 |
