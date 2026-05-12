# Token 费用节省最佳实践（人工参考）

> 本文档仅供团队成员参考，不作为 AI steering 规则加载。
> 基于 Claude Prompt Caching 机制整理，适用于 Kiro IDE + Claude API 的开发场景。

---

## 核心机制速览

### Prompt Caching 原理

Claude 的 prompt caching 基于**严格前缀匹配**：

- 每次 API 请求会发送完整的 `[tools 定义] → [system prompt] → [对话历史]`
- 如果本次请求的前缀与上次完全一致（逐字节匹配），则命中缓存
- 前缀中**任何一个字节**变化，该位置之后的所有缓存全部失效（级联失效）

### 费率对比

| 操作 | 费率（相对正常 input） |
|------|----------------------|
| 缓存命中读取 | **0.1x**（节省 90%） |
| 首次写入缓存 | **1.25x**（多付 25%） |
| 未缓存的新内容 | **1x** |
| Output token | **1x**（无缓存，始终全价） |

### 缓存 TTL

- 默认 5 分钟不活跃即过期
- 可选 1 小时（需 beta header）
- 每次请求命中缓存会刷新 TTL

### 最小可缓存大小

- Claude Sonnet / Opus: 1024 tokens
- Claude Haiku: 2048 tokens

---

## 一、Steering 文件管理

### ✅ 应该做

| 实践 | 原因 |
|------|------|
| `.kiro/steering/` 下只放极简、确定不再修改的核心文件 | system prompt 前缀不变 = 缓存持续命中 |
| 可变规范放到 `AIDOC/global-info/knowledge/steering/` 下 | AI 按需读取，修改不影响 system prompt 缓存 |
| 新增 steering 文件优先用 `fileMatch` 或 `manual` 模式 | 只在需要时注入 system prompt，不污染每次请求 |
| steering 文件内容确定后做一次性定稿 | 减少后续修改导致的缓存失效 |

### ❌ 避免做

| 反模式 | 后果 |
|--------|------|
| 频繁修改 `inclusion: always` 的 steering 文件 | 每次修改 → system prompt 变化 → 缓存失效 → 重新付 1.25x 写入 |
| 在 steering 文件中放时间戳、版本号等动态内容 | 每次请求前缀都不同，缓存永远命中不了 |
| 堆积大量 `inclusion: always` 文件 | 任何一个改动都导致全部失效；增大每次请求的 token 基数 |
| 在 steering 中用 `#[[file:...]]` 引用频繁变动的文件 | 被引用文件变化 = steering 内容变化 = 缓存失效 |

---

## 二、会话管理

### ✅ 应该做

| 实践 | 原因 |
|------|------|
| 同一任务在同一会话内完成 | 对话历史被缓存，后续轮次只付 0.1x |
| 相关任务集中处理（如连续写多段提示词） | knowledge 文件读取一次，整个会话内缓存复用 |
| 会话开头明确任务范围 | AI 一次性读取所需文件，后续不再重复读取 |
| 适时开新会话（对话超过 ~50 轮或上下文接近上限时） | 避免触发 context compaction 导致缓存重组失效 |

### ❌ 避免做

| 反模式 | 后果 |
|--------|------|
| 每个小问题都开新会话 | 每次新会话 system prompt 重新写入缓存（1.25x），knowledge 重新读取 |
| 一个会话里跳跃处理完全不相关的任务 | 对话历史膨胀，每轮发送的 token 总量越来越大 |
| 长时间不发消息（>5 分钟）后继续 | 缓存可能已过期，下一轮需要重新写入 |
| 会话过长不收尾 | 触发 context compaction，缓存前缀被截断/重组 |

---

## 三、文件读取策略

### ✅ 应该做

| 实践 | 原因 |
|------|------|
| 会话开头让 AI 集中读取本次需要的所有 knowledge 文件 | 一次读取，后续所有轮次享受缓存 |
| 用 `#File` 精确引用文件 | 避免 AI 猜测读取不需要的文件 |
| knowledge 文件保持合理大小（1000~3000 tokens/个） | 太大浪费，太小不值得缓存 |
| 每个 knowledge 文件职责单一 | 按需读取时不会引入无关内容 |

### ❌ 避免做

| 反模式 | 后果 |
|--------|------|
| 每轮都让 AI 重新读取同一个文件 | 虽然历史中已有内容会被缓存，但新的 tool call/result 增加对话长度 |
| 把所有 knowledge 合并成一个巨大文件 | 即使只需要一部分也要全部读入 |
| knowledge 文件之间大量重复内容 | 浪费 token 且增加维护成本 |

---

## 四、交互方式

### ✅ 应该做

| 实践 | 原因 |
|------|------|
| 指令清晰简洁，一次说清需求 | 减少来回澄清轮次 = 减少对话历史膨胀 |
| 批量任务一次性提出 | 比分多次会话处理省很多 |
| 给反馈时具体指出修改点 | 避免 AI 重新读取文件或重新生成整个内容 |
| 用"修改 XX 部分的 YY"代替"重新来" | 减少 output token（全价）消耗 |

### ❌ 避免做

| 反模式 | 后果 |
|--------|------|
| 碎片化交互（每句话一条消息） | 每条消息都是完整 API 调用，对话历史快速膨胀 |
| 粘贴大段文件内容到聊天中 | 用 `#File` 引用更高效，避免 messages 中重复存储 |
| 反复说"重新来"而不指出具体问题 | 大量 output token 浪费 + 可能重新读取文件 |
| 在消息中重复描述已在 steering/knowledge 中定义的规范 | 冗余 token，AI 已经知道这些信息 |

---

## 五、工具和 MCP 配置

### ✅ 应该做

| 实践 | 原因 |
|------|------|
| MCP 工具配置确定后保持稳定 | tool 定义在 prompt 最前面，变化导致后面所有缓存级联失效 |
| 不用的 MCP server 设置 `disabled: true` | 减少 tool 定义的 token 数量 |
| 只安装实际需要的 MCP server | 每个 server 的 tool schema 都占 token |

### ❌ 避免做

| 反模式 | 后果 |
|--------|------|
| 会话中途添加/删除 MCP 工具 | tool 定义变化 → 整个 prompt 前缀失效 → 级联缓存失效 |
| 安装大量不常用的 MCP server | 增大 token 基数 + 增大缓存失效风险面 |
| 频繁切换 model（如 Sonnet ↔ Opus） | 缓存按 model 隔离，切换 = 缓存全部失效 |

---

## 六、项目结构对缓存的影响

### 当前架构的缓存效果

```
.kiro/steering/core.md（~40 行，inclusion: always）
    ↓ 注入 system prompt
    ↓ 极度稳定 → 缓存持续命中 ✅

AIDOC/global-info/knowledge/steering/（10 个文件）
    ↓ AI 按需 read_file
    ↓ 出现在 messages 中
    ↓ 同一会话内后续轮次被缓存 ✅
    ↓ 修改这些文件不影响 system prompt 缓存 ✅
```

### 缓存失效的级联关系

```
tools 定义变化
    → system prompt 缓存失效
        → 对话历史缓存失效
            → 一切从头重新计算（全价）

system prompt 变化（steering 文件修改）
    → 对话历史缓存失效
        → 一切从 system prompt 之后重新计算

对话历史中间某轮变化
    → 该轮之后的缓存失效
        → 但之前的缓存仍然有效
```

---

## 七、费用估算参考

### 典型会话费用构成

假设一次 10 轮的提示词编写会话：

| 组成部分 | Token 数（估） | 费率 | 说明 |
|----------|---------------|------|------|
| System prompt（含 steering） | ~2000 | 第 1 轮 1.25x，后续 0.1x | 只付一次写入费 |
| Knowledge 文件读取 | ~5000 | 第 1 轮 1x，后续 0.1x | 读取后进入对话历史被缓存 |
| 对话历史（累积） | 逐轮增长 | 已缓存部分 0.1x | 越往后缓存比例越高 |
| 每轮新增内容 | ~500 | 1x | 你的新消息 + AI 的新 tool call |
| AI 输出 | ~2000/轮 | 1x | 始终全价，无法缓存 |

### 节省对比

| 场景 | 10 轮会话估算成本（相对值） |
|------|---------------------------|
| 无缓存（每轮全价处理） | 100% |
| 有缓存 + 稳定 steering | ~35%（节省约 65%） |
| 有缓存但 steering 每轮都变 | ~90%（几乎没省） |

---

## 八、检查清单

在开始一个新的工作周期前，检查以下项目：

- [ ] `.kiro/steering/` 下的文件是否稳定，近期是否有不必要的修改？
- [ ] 是否有新的 `inclusion: always` 文件被添加？是否可以改为 `fileMatch` 或 `manual`？
- [ ] MCP 配置是否稳定？是否有不用的 server 可以 disable？
- [ ] knowledge 文件是否有大量重复内容可以精简？
- [ ] 是否有可以合并处理的相关任务（减少开新会话的次数）？

---

## 参考来源

- [Anthropic Prompt Caching 官方文档](https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching)
- [Kiro Steering 官方文档](https://kiro.dev/docs/steering/)
- [Cache Breakpoints 技术解析](https://kindatechnical.com/claude-ai/cache-breakpoints-how-claudes-caching-actually-works.html)

> Content was rephrased for compliance with licensing restrictions.
