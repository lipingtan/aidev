# 需求：知识库业务域隔离 + 租户隔离（CR1）

## 1. 背景

go-rag-agent 作为独立 AI 客服引擎，被 platform_admin 等多业务系统调用。各业务系统有自己的租户体系（外部租户 ID 格式不同），go-rag-agent 内部需维护一套统一的内部租户 key，所有隔离逻辑（RAG 检索、Session、知识库文档）均基于内部 key 执行。

本 CR 完成三项改造：
1. 租户映射层（TenantRegistry）
2. 知识库按租户+域隔离
3. Session 按租户隔离

并在 Admin 页面提供对应的管理界面。

---

## 2. 租户映射层

### 2.1 数据模型

租户注册表存储在 SQLite（复用现有 session DB 同一文件），表名 `agent_tenants`：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增 |
| internal_key | TEXT UNIQUE | 内部唯一标识，如 `mjarch` |
| display_name | TEXT | 显示名称，如「街机时代」 |
| default_domain | TEXT | 默认业务域，请求未传 domain 时使用 |
| allow_cross_domain | BOOLEAN | 是否允许同租户跨域 RAG 检索，默认 false。适用场景：同一租户下有多个业务域（如客服+销售）且知识库内容互通时开启；默认关闭保证最小权限，避免不相关域的知识污染检索结果 |
| created_at | DATETIME | 创建时间 |

租户外部绑定表 `agent_tenant_bindings`：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增 |
| tenant_id | INTEGER FK | 关联 agent_tenants.id |
| source | TEXT | 来源系统，如 `platform_admin` |
| external_id | TEXT | 业务系统的外部租户 ID |
| UNIQUE(source, external_id) | | 同一来源的外部 ID 唯一 |

### 2.2 映射解析逻辑

请求进来时，从请求头取：
- `X-Tenant-ID`：业务系统的外部租户 ID
- `X-Source`：来源系统标识（可选，默认 `platform_admin`）

解析优先级：
1. 在 `agent_tenant_bindings` 中按 `(source, external_id)` 查找 → 得到 `internal_key`
2. 若未命中，检查 `X-Tenant-ID` 是否直接等于某个 `internal_key` → 直接用
3. 若均未命中：
   - `strict_tenant_mode=false`（默认）→ 兜底 `internal_key="default"`，打 WARN 日志
   - `strict_tenant_mode=true` → 返回 400，body: `{"error":"unknown tenant"}`

### 2.3 内置 default 租户

系统启动时自动确保 `internal_key="default"` 的租户记录存在（AutoMigrate + seed），对应现有 ecommerce demo 场景，`default_domain="ecommerce"`。

**default 租户为调试专用**：仅在 `strict_tenant_mode=false` 时作为兜底使用，不应在生产环境中承载真实业务数据。业务系统上线前必须建立正式的 binding 映射，不得依赖 default 兜底运行。

### 2.4 strict_tenant_mode 配置

在现有系统配置（`data/agent_config.json` 或 admin config）中增加：
```json
{ "strict_tenant_mode": false }
```
Admin 页「系统配置」section 可修改，热更新，无需重启。

---

## 3. 知识库域隔离

### 3.1 Qdrant payload 字段

所有文档在写入 Qdrant 时，必须携带以下 payload 字段：

| 字段 | 值示例 | 说明 |
|------|--------|------|
| `tenant_id` | `"mjarch"` | 内部租户 key |
| `domain` | `"mjarch"` | 业务域名称 |
| `doc_id` | `"doc-xxx"` | 文档 ID（已有） |
| `content` | `"..."` | 文档内容（已有） |

Qdrant collection 初始化时为 `tenant_id` 和 `domain` 两个字段建立 payload index（keyword 类型），确保 filter 走索引。

### 3.2 RAG 检索 filter

`RAGPipeline.RetrieveWithScores` 和 `Retrieve` 方法新增 filter 参数：

```go
type RAGFilter struct {
    TenantID string
    Domain   string // 空=不过滤 domain（allow_cross_domain=true 时）
}
```

filter 规则：
- `TenantID` 必填，强制过滤（安全边界）
- `Domain`：当前请求的 domain；若租户配置了 `allow_cross_domain=true` 则传空（不过滤 domain）

### 3.3 现有文档迁移

提供一次性迁移命令（`go run migration/patch_tenant_domain.go`）：
- 将所有现有文档（无 `tenant_id` payload 字段的）补写 `tenant_id="default", domain="ecommerce"`
- 不重新 embedding（只更新 payload，调 Qdrant SetPayload API）
- 执行前打印预览条数，执行后打印成功/失败数

---

## 4. Session 隔离

### 4.1 Session key 格式

```
{internal_key}_{original_session_id}
```

例：`mjarch_web-mvxb1imt-4msuknz`

分隔符使用 `_` 而非 `/`，避免 Admin 会话管理页展示时泄露 internal_key（管理端展示时按首个 `_` split，只显示 session 部分）。

### 4.2 改造位置

`agent.go` 的 `ChatStream` 和 `Chat` 方法：在使用 `sessionID` 前，包装为 `tenantKey + "_" + sessionID`。

SessionManager 的 GetHistory / AppendMessage / Clear 等方法无需改动，key 透传即可。

### 4.3 现有 session 不迁移

现有 session 数据不迁移，过期后自然淘汰（TTL 已有）。

---

## 5. Admin 管理界面

### 5.1 新增「🏢 租户配置」section

位置：admin.html 左侧菜单，排在「文档管理」之前。

功能：
- **列表**：展示所有已注册租户（internal_key、display_name、default_domain、allow_cross_domain、绑定数）
- **新增**：填写 internal_key、display_name、default_domain、allow_cross_domain
- **编辑**：修改 display_name、default_domain、allow_cross_domain
- **删除**：删除租户及其所有 bindings（需确认弹窗；若该租户下有文档，禁止删除并提示先清理文档）
- **绑定管理**：点击某个租户展开，可增删外部 ID 绑定（source + external_id）

API：
```
GET    /api/admin/tenants              列表
POST   /api/admin/tenants              新增
PUT    /api/admin/tenants/{key}        编辑
DELETE /api/admin/tenants/{key}        删除
GET    /api/admin/tenants/{key}/bindings        绑定列表
POST   /api/admin/tenants/{key}/bindings        新增绑定
DELETE /api/admin/tenants/{key}/bindings/{id}   删除绑定
```

### 5.2 改造「📄 文档管理」section

**上传表单新增字段：**
- 租户（select，从 `/api/admin/tenants` 动态加载已注册租户列表，不允许手动输入）
- 业务域（select，从 `/api/admin/domains` 动态加载已启用域列表，不允许手动输入）

**列表新增列：** 租户、业务域

**列表新增筛选器：** 按租户筛选、按业务域筛选

**后端 `/api/admin/documents` 接口变更：**
- POST 上传：body 新增 `tenant_id`、`domain` 字段（必填）
- GET 列表：支持 `?tenant_id=mjarch&domain=mjarch` query 参数过滤

---

## 6. 验收条件

| # | 场景 | 预期结果 |
|---|------|----------|
| AC-1 | 环境中同时存在 mjarch 和 ecommerce 各1条文档；platform_admin 带外部租户 ID（已绑定 mjarch）发请求 | RAG 日志中可见 `[RAG] tenant=mjarch filter applied`；响应内容只涉及 mjarch 文档，不含 ecommerce 文档 |
| AC-2 | 请求不带 X-Tenant-ID，strict_tenant_mode=false | 兜底 `default` 正常返回；日志中打印 WARN `unknown tenant, fallback to default` |
| AC-3 | 请求不带 X-Tenant-ID，strict_tenant_mode=true | 返回 400 `{"error":"unknown tenant"}` |
| AC-4 | mjarch 租户上传一篇文档，以 ecommerce 租户身份发请求 | ecommerce 的 RAG 检索结果中不含 mjarch 的文档 |
| AC-5 | 两个租户使用相同 session_id 各发一条消息 | `GET /api/admin/sessions` 中可见两条独立 session（key 分别含各自 internal_key 前缀）；各自历史互不干扰 |
| AC-6 | Admin 页租户配置：新增租户 + 绑定外部 ID | 后续请求带该外部 ID 可正确映射到新租户 |
| AC-7 | Admin 页文档管理：上传文档时不填租户/域直接提交 | 前端提示「请选择租户和业务域」，文档不被上传 |
| AC-8 | 运行迁移脚本 | 现有文档全部补写 `tenant_id=default, domain=ecommerce` payload；Qdrant 中可验证 |
| AC-9 | 不带 X-Admin-Token 访问 `/api/admin/tenants` | 返回 401 |
