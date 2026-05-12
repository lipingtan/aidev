# AI 视频制作知识库

> 系统化积累 AI 视频生成的经验、模式和方法论。
> 这是未来产品化的核心资产之一。

## 目录结构

```
knowledge/
├── README.md                       # 本文件
├── engines/                        # 按引擎组织的技术知识
│   ├── kling/                      # KLING AI 3.0
│   │   ├── spec.md                 # 技术规格
│   │   ├── prompt-format.md        # 提示词格式规范
│   │   ├── best-practices.md       # 最佳实践
│   │   ├── limitations.md          # 已知限制和坑
│   │   ├── changelog.md            # 引擎版本更新记录
│   │   └── patterns/               # 验证有效的模式
│   │       ├── dialogue.md         # 对话场景模式
│   │       ├── action.md           # 动作场景模式
│   │       ├── emotion.md          # 情绪表达模式
│   │       ├── transition.md       # 转场模式
│   │       └── camera.md           # 运镜模式
│   └── wan/                        # WAN 2.5 / 2.6
│       ├── spec.md                 # 技术规格
│       ├── prompt-format.md        # 提示词格式规范
│       ├── best-practices.md       # 最佳实践
│       ├── limitations.md          # 已知限制和坑
│       ├── changelog.md            # 引擎版本更新记录
│       └── patterns/               # 验证有效的模式
│           ├── dialogue.md         # 对话场景模式
│           ├── action.md           # 动作场景模式
│           ├── emotion.md          # 情绪表达模式
│           ├── transition.md       # 转场模式
│           └── camera.md           # 运镜模式
│
├── common/                         # 跨引擎通用知识
│   ├── cinematography.md           # 镜头语言通用规则
│   ├── character-consistency.md    # 角色一致性方法论
│   ├── pacing.md                   # 节奏控制方法论
│   ├── style-keywords.md           # 风格关键词库
│   └── negative-prompts.md         # 负面提示词策略
│
└── cases/                          # 案例库
    ├── README.md                   # 案例记录规范
    ├── success/                    # 成功案例
    └── failure/                    # 失败案例
```

## 知识条目规范

### 模式条目格式（patterns/ 下的每条记录）

```markdown
## 模式名称：[简短描述]

**适用场景**：什么时候用这个模式
**引擎**：KLING 3.0 / WAN 2.5
**验证次数**：X/Y 次效果稳定

### 有效写法
（具体的提示词片段）

### 无效写法
（对比：不好的写法）

### 为什么有效
（原理分析）

### 标签
#tag1 #tag2 #tag3
```

### 案例条目格式（cases/ 下的每条记录）

```markdown
# 案例 XXX：[简短标题]

## 基本信息
| 项目 | 值 |
|------|-----|
| 引擎 | KLING / WAN |
| 场景类型 | 对话/动作/情绪/转场 |
| 日期 | YYYY-MM-DD |
| 结果 | 成功/失败 |

## 使用的提示词
（完整提示词）

## 效果描述
（生成结果的描述，好的和不好的方面）

## 关键发现
（从这次生成中学到了什么）

## 标签
#tag1 #tag2
```

## 使用方式

1. **制作过程中**：每次生成视频后，记录经验到对应的 patterns/ 或 cases/
2. **写提示词前**：查阅对应引擎的 best-practices 和 patterns
3. **遇到问题时**：查阅 limitations 和 cases/failure/
4. **产品化时**：这些知识作为 RAG 数据源驱动智能推荐
