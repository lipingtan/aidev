# 需求：知识库域管理 + Soul 传参（CR2）

## 1. 背景

本 CR 包含两个独立功能：
1. **知识库域管理**：租户下自定义知识库分类标签，文档按域归属，RAG 检索按策略过滤
2. **Soul 传参**：调用方通过请求传入自定义人格，覆盖 system prompt 的人格部分

---

## 2. CR2-1 知识库域管理

### 2.1 概念澄清

| 概念 | 说明 | 管理方式 |
|------|------|---------|
| Agent 工具域 | config/domains/*.yaml，决定工具和 prompt | 工程师配置文件 |
| 知识库域（KB域）| 文档分类标签，如 faq/policy/manual | 管理员在 Admin 页配置 |

两者完全独立，名字可以相同（如都叫 `mjarch`）但互不影响。

### 2.2 KB域数据模型

`agent_tenants` 表新增 `kb_domains` 字段（TEXT，存 JSON 数组）：

```json
["faq", "policy", "manual"]
```

默认值：`[]`（空数组）。

### 2.3 RAG 检索策略

每个租户配置 `kb_search_policy`（存在 `agent_tenants` 表）：

| 策略 | 行为 |
|------|------|
| `strict` | 只查请求指定的 KB域，没有则返回空 |
| `fallback` | 先查指定域，有效命中数 < `min_kb_hits`（默认2）时，补充查该租户其他域 |
| `cross` | 查该租户所有 KB域（忽略请求传的 domain） |

默认策略：`fallback`，`min_kb_hits=2`。

### 2.4 Admin 管理界面

**租户详情页新增「📚 知识库域」区块：**
- 显示当前租户的 KB域列表
- 可新增域名（自由输入，如 `faq`）
- 可删除域名（有文档时禁止删除并提示）
- 可设置 `kb_search_policy`（下拉选：strict/fallback/cross）

**文档管理上传表单改造：**
- 域下拉改为从 `/api/admin/tenants/{key}/kb-domains` 取（而非 `/api/admin/domains`）
- 选择租户后联动加载该租户的 KB域列表

### 2.5 API

```
GET    /api/admin/tenants/{key}/kb-domains          KB域列表 + 策略配置
POST   /api/admin/tenants/{key}/kb-domains          新增KB域
DELETE /api/admin/tenants/{key}/kb-domains/{domain} 删除KB域
PUT    /api/admin/tenants/{key}/kb-policy           更新检索策略
```

### 2.6 RAGFilter 扩展

```go
type RAGFilter struct {
    TenantID string
    Domain   string // 请求指定的 KB域
    Policy   string // "strict" | "fallback" | "cross"，空=fallback
    MinHits  int    // fallback 触发阈值，0=使用默认值2
}
```

### 2.7 验收条件

| # | 场景 | 预期 |
|---|------|------|
| AC-1 | 租户配置页新增 KB域 `faq` | 文档管理域下拉出现 `faq` 选项 |
| AC-2 | strict 策略，请求 domain=faq，只有 policy 文档 | 返回空，不跨域补充 |
| AC-3 | fallback 策略，请求 domain=faq，命中1条（< min_hits=2） | 自动补充其他域的文档 |
| AC-4 | cross 策略 | 返回该租户所有域的文档 |
| AC-5 | 文档上传选 tenant=mjarch，域下拉只显示 mjarch 的 KB域 | 与工具域无关 |
| AC-6 | 有文档的 KB域不能删除 | 返回 409，提示「该域下有 N 篇文档，请先删除」 |
| AC-13 | 租户无 KB域时打开文档上传 | 提示「请先在租户配置页添加知识库域」，禁止上传 |

---

## 3. CR2-2 Soul 管理与传参

### 3.1 Soul 数据模型

Soul 是命名的人格模板，存储在 `agent_souls` 表（SQLite，与 session DB 同文件）：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增 |
| name | TEXT UNIQUE | Soul 名称，调用方用此名引用，如 `friendly`、`formal` |
| display_name | TEXT | 显示名称，如「贴心小美」 |
| content | TEXT | 人格内容，替换 system_prompt_extra，≤ 2000 字符 |
| tenant_id | TEXT | 归属租户（空=全局可用） |
| created_at | DATETIME | 创建时间 |

### 3.2 chatRequest 新增字段

```go
type chatRequest struct {
    // ... 现有字段 ...
    SoulName string `json:"soul_name,omitempty"` // 指定 soul 名称，覆盖默认人格
}
```

调用方只需传 `soul_name`，内容由服务端查询，不暴露 prompt 细节。

### 3.2.1 默认 Soul Seed

系统启动时自动 seed 一个示例 soul（INSERT OR IGNORE）：
- name: `default`，tenant_id: `""`（全局可用）
- display_name: `通用客服`
- content: `你是一名专业、友善的智能客服助手，简洁回答用户问题，无法解决时主动建议转人工。`

管理员可在 Admin 页修改此默认 soul 的内容。

### 3.3 Soul 处理规则

- `soul_name` 非空：按以下优先级查询
  1. 当前租户的 soul：`WHERE name=? AND tenant_id=当前租户`
  2. 全局 soul：`WHERE name=? AND tenant_id=''`（空串，表示全局可用，非 "default" 字符串）
  3. 都没有：打 WARN 日志，fallback 使用 yaml 的 system_prompt_extra
- `soul_name` 为空/不传：直接使用 yaml 的 system_prompt_extra（现有行为不变）
- 找到的 soul content 替换 system_prompt_extra，其余 system prompt（工具说明、context 等）保留

### 3.4 Admin 管理界面

**新增「🎭 Soul 管理」section：**
- 列表：name、display_name、tenant_id（全局/指定租户）、内容预览
- 新增：name（英文唯一）、display_name、content（文本框）、归属租户（可选）
- 编辑：修改 display_name 和 content
- 删除：确认后删除

**API：**
```
GET    /api/admin/souls              Soul 列表（可按 tenant_id 过滤）
POST   /api/admin/souls              新增
PUT    /api/admin/souls/{name}       编辑
DELETE /api/admin/souls/{name}       删除
```

### 3.5 验收条件

| # | 场景 | 预期 |
|---|------|------|
| AC-7 | Admin 页新建 soul `friendly`，content 为「你是贴心客服小美…」 | 后续请求传 soul_name=friendly 生效 |
| AC-8 | 请求传 soul_name=friendly | AI 回复风格符合 soul 内容，不使用 yaml 默认人格 |
| AC-9 | 传不存在的 soul_name | WARN 日志，fallback 使用默认人格，SSE meta 事件含 `soul_used: default` |
| AC-10 | soul_name 为空/不传 | 使用 yaml 里的 system_prompt_extra（现有行为不变）|
| AC-11 | mjarch 无 friendly soul，default 租户有 | 使用 default 租户的 friendly soul，meta 含 `soul_used: friendly@default` |
| AC-12 | mjarch 和 default 都有 friendly soul | 使用 mjarch 的（优先级更高），meta 含 `soul_used: friendly@mjarch` |
| AC-14 | Soul 管理页无 soul 时 | 显示空态提示「暂无 Soul，新增后调用方可通过 soul_name 参数指定」 |
