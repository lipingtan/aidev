# 需求计划：知识库域管理 + Soul 传参（CR2）

## 需求理解

### CR2-1 知识库域管理
**目标：** 在租户下管理自定义知识库域标签，文档按域归属，RAG 检索按策略决定查哪些域。

**背景：** go-rag-agent 中「业务域」有两个独立概念：
- **Agent 工具域**（config/domains/*.yaml）：决定用哪些工具、system prompt，工程师配置，不在本 CR 范围
- **知识库域**（本 CR）：文档分类标签，管理员自定义，用于知识检索隔离

前端请求传 `domain` 参数 → RAG 检索时按策略决定范围：
- `strict`：只查指定域
- `fallback`：先查指定域，不够（低于阈值或条数不足）再补其他域
- `cross`：查该租户所有域

**范围：**
- 租户配置：新增「知识库域」列表（自由命名，如 faq/policy/manual）
- 文档管理：上传时从租户的知识库域列表选择（而非 config/domains 的工具域）
- RAG 检索：按请求的 domain 和租户的 kb_search_policy 策略过滤

### CR2-2 Soul 传参
**目标：** chatRequest 加 `soul` 字段，调用方可传入自定义 system prompt 人格部分，覆盖默认的 system_prompt_extra。

**背景：** 目前人格固定在 config/domains/*.yaml 的 system_prompt_extra 里，不同调用场景（如不同语气的客服）需要动态切换。

---

## 假设列表

- [假设-1] 知识库域（KB域）完全独立于 config/domains/*.yaml 的 Agent 工具域，两者无耦合。KB域只是字符串标签，存在租户记录里。
- [假设-2] 默认 kb_search_policy = `fallback`（先查指定域，条数不足时补其他域），行为对现有功能无破坏。
- [假设-3] soul 字段是纯文本，插入 system prompt 的人格位置，不替换整个 system prompt（工具说明、context 等保留）。
- [假设-4] KB域列表存储在 agent_tenants 表新增的 `kb_domains` 字段（JSON 数组），无需单独建表。
- [假设-5] 现有文档的 domain payload 字段继续有效，KB域列表只影响新上传文档的可选项，不做历史数据迁移。

---

## 澄清问题

- [Question-1] RAG 检索 fallback 策略的触发条件：文档条数不足（如 < topK）还是分数不足（< threshold）？
  业界做法：通常两者都判断（有效命中 < min_hits 时补充）。
  推荐：命中数 < min_hits（默认 2）时触发 fallback，min_hits 可配置。
  [Answer-1]
  命中数 < min_hits 触发，ok

- [Question-2] KB域列表是租户级别（所有 Agent 工具域共享同一套 KB域）还是每个 Agent 工具域独立一套 KB域？
  推荐：租户级别共享，够用且简单；如果需要隔离，后续再拆分。
  [Answer-2]
  租户级别共享，ok

- [Question-3] soul 字段的插入位置：完全替换 system_prompt_extra，还是在其前面/后面追加？
  推荐：替换 system_prompt_extra 部分（人格），工具说明和 context 保留，这样调用方只需关心人格。
  [Answer-3]
  替换 system_prompt_extra 部分，ok

- [Question-4] soul 字段是否需要鉴权或长度限制，防止注入？
  推荐：长度限制 ≤ 2000 字符，内容做基础 sanitize（去除 `\n---\n` 等 prompt 分隔注入）。
  [Answer-4]
  ok

- [Question-5] 文档管理域下拉：改为从租户的 KB域列表取，还是同时兼容工具域？
  推荐：仅从 KB域列表取，完全解耦；工具域名可以和 KB域名相同（如都叫 mjarch）但独立管理。
  [Answer-5]
  仅从 KB域列表取，ok

---

## 影响范围预判

### CR2-1 知识库域管理

**新增：**
- `tenant/registry.go`：`KBDomains` 字段（[]string，JSON存储）；CRUD 方法 `GetKBDomains / AddKBDomain / RemoveKBDomain`
- `admin_api.go`：GET/POST/DELETE `/api/admin/tenants/{key}/kb-domains`
- `frontend/admin.html`：租户详情页加 KB域管理区块；文档管理域下拉改为从 KB域 API 取

**修改：**
- `agent_tenants` 表：新增 `kb_domains` TEXT 字段（JSON 数组，默认 `[]`）
- `rag/rag.go`：`RAGFilter` 加 `Policy` 字段；`RetrieveWithScores` 按策略处理 fallback/cross/strict
- `agent/agent.go`：ChatStream 构造 RAGFilter 时传入策略（从租户配置读）

### CR2-2 Soul 管理与传参

**新增：**
- `soul/store.go`：SoulStore（SQLite CRUD，agent_souls 表）
- `admin_api.go`：Soul CRUD API（`/api/admin/souls`）
- `frontend/admin.html`：新增「🎭 Soul 管理」section

**修改：**
- `main.go`：chatRequest 加 `soul_name` 字段；启动时初始化 SoulStore
- `callctx/callctx.go`：新增 `WithSoulName / GetSoulName`
- `agent/agent.go`：buildSystemPrompt 时若 ctx 有 soul_name，查 SoulStore 替换 system_prompt_extra
