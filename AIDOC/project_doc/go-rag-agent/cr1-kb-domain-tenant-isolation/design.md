# 设计文档：知识库业务域隔离 + 租户隔离（CR1）

## 1. 数据模型

### 1.1 agent_tenants 表

```sql
CREATE TABLE IF NOT EXISTS agent_tenants (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    internal_key     TEXT NOT NULL UNIQUE,
    display_name     TEXT NOT NULL,
    default_domain   TEXT NOT NULL DEFAULT '',
    allow_cross_domain BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- seed: default 租户
INSERT OR IGNORE INTO agent_tenants (internal_key, display_name, default_domain)
VALUES ('default', 'Default', 'ecommerce');
```

### 1.2 agent_tenant_bindings 表

```sql
CREATE TABLE IF NOT EXISTS agent_tenant_bindings (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id   INTEGER NOT NULL REFERENCES agent_tenants(id) ON DELETE CASCADE,
    source      TEXT NOT NULL,
    external_id TEXT NOT NULL,
    UNIQUE(source, external_id)
);
```

### 1.3 rag.Document 扩展

```go
// rag/document.go（或 rag/rag.go 顶部）
type Document struct {
    ID       string
    Content  string
    Score    float64
    Metadata map[string]interface{}
    // 新增：隔离字段
    TenantID string // 内部租户 key
    Domain   string // 业务域
}
```

### 1.4 RAGFilter

```go
// rag/rag.go
type RAGFilter struct {
    TenantID string // 必填；空字符串=不过滤（仅用于 Admin 测试接口）
    Domain   string // 可选；空=不过滤 domain（allow_cross_domain=true 时）
}
```

---

## 2. tenant/registry.go

```go
package tenant

import (
    "database/sql"
    _ "github.com/glebarez/sqlite"
)

type Tenant struct {
    ID               int64
    InternalKey      string
    DisplayName      string
    DefaultDomain    string
    AllowCrossDomain bool
}

type Binding struct {
    ID         int64
    TenantID   int64
    Source     string
    ExternalID string
}

type Registry struct {
    db *sql.DB
}

func NewRegistry(db *sql.DB) (*Registry, error)   // AutoMigrate + seed default + 初始化 cache
    // 注意：不依赖 SQLite 外键自动级联（需要 PRAGMA foreign_keys=ON 且驱动支持）
    // 改为在 Delete 方法中手动级联删除 bindings，避免驱动差异问题
func (r *Registry) Resolve(source, externalID string) string
    // 缓存层（TTL=60s，sync.Map + 过期时间戳）：
    //   key = source+":"+externalID，命中且未过期直接返回
    // 未命中走 DB：
    // 1. SELECT internal_key FROM agent_tenants t
    //    JOIN agent_tenant_bindings b ON b.tenant_id=t.id
    //    WHERE b.source=? AND b.external_id=?
    // 2. SELECT internal_key FROM agent_tenants WHERE internal_key=? (externalID)
    // 3. 返回 "default"
    // 写入缓存，增删改 binding/tenant 时调 r.invalidateCache() 清除

func (r *Registry) List() ([]Tenant, error)
func (r *Registry) Get(internalKey string) (*Tenant, error)
func (r *Registry) Create(t Tenant) error              // INSERT; internal_key 唯一校验
func (r *Registry) Update(internalKey string, t Tenant) error
func (r *Registry) Delete(internalKey string) error
    // 1. 调 store.Count(filter{tenant_id=internalKey}) — 使用 Qdrant CountPoints API，O(1)
    //    count > 0 → 返回 ErrHasDocuments
    // 2. DELETE FROM agent_tenant_bindings WHERE tenant_id=... （手动级联，不依赖 SQLite 外键）
    // 3. DELETE FROM agent_tenants WHERE internal_key=?
    // 4. 清理 SQLite session 表中 key LIKE '{internalKey}_%' 的记录
    // 5. invalidateCache()
func (r *Registry) ListBindings(internalKey string) ([]Binding, error)
func (r *Registry) AddBinding(internalKey, source, externalID string) error
func (r *Registry) DeleteBinding(id int64) error

var ErrHasDocuments = errors.New("tenant has documents, delete them first")
var ErrNotFound     = errors.New("tenant not found")
var ErrDuplicate    = errors.New("internal_key already exists")
```

---

## 3. storage/qdrant_store.go 变更

### 3.1 VectorStore 接口

```go
// storage/store.go（接口文件）
type VectorStore interface {
    Add(ctx context.Context, docs []rag.Document) error
    Search(ctx context.Context, query string, topK int, filter rag.RAGFilter) ([]rag.ScoredDocument, error)
    Count(ctx context.Context, filter rag.RAGFilter) (int, error) // 新增，用于 Delete 时计数；调 Qdrant CountPoints API
    Delete(ctx context.Context, documentID string) error
    List(ctx context.Context, filter rag.RAGFilter) ([]string, error) // 新增 filter，用于按 tenant+domain 列文档
    // ... 其他方法不变
}
```

### 3.2 ensureCollection 新增 payload index

```go
func (s *QdrantStore) ensureCollection() error {
    // ... 现有建 collection 逻辑 ...

    // 建 payload index（幂等）
    for _, field := range []string{"tenant_id", "domain"} {
        _, err := s.client.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexPoints{
            CollectionName: s.collectionName,
            FieldName:      field,
            FieldType:      qdrant.FieldType_FieldTypeKeyword,
        })
        // 已存在时忽略错误
    }
}
```

### 3.3 Add 写 payload

```go
payload := map[string]*qdrant.Value{
    "content":   qdrant.NewValueString(doc.Content),
    "doc_id":    qdrant.NewValueString(doc.ID),
    "tenant_id": qdrant.NewValueString(doc.TenantID), // 新增
    "domain":    qdrant.NewValueString(doc.Domain),   // 新增
}
for k, v := range doc.Metadata { ... }
```

### 3.4 Search 构造 filter

```go
func (s *QdrantStore) Search(ctx context.Context, query string, topK int, filter rag.RAGFilter) ([]rag.ScoredDocument, error) {
    var conditions []*qdrant.Condition

    if filter.TenantID != "" {
        conditions = append(conditions, qdrant.NewCondition("tenant_id",
            &qdrant.Match{MatchValue: &qdrant.Match_Keyword{Keyword: filter.TenantID}}))
    }
    if filter.Domain != "" {
        conditions = append(conditions, qdrant.NewCondition("domain",
            &qdrant.Match{MatchValue: &qdrant.Match_Keyword{Keyword: filter.Domain}}))
    }

    var qdrantFilter *qdrant.Filter
    if len(conditions) > 0 {
        qdrantFilter = &qdrant.Filter{Must: conditions}
    }

    req := &qdrant.SearchPoints{
        CollectionName: s.collectionName,
        Vector:         vector,
        Limit:          uint64(topK),
        WithPayload:    qdrant.NewWithPayload(true),
        Filter:         qdrantFilter, // nil = 无过滤
    }
    // ...
}
```

### 3.5 memory_store.go 同步调整

Search / List 方法签名加 `filter rag.RAGFilter` 参数，函数体忽略 filter（测试用途）。

---

## 4. rag/rag.go 变更

```go
// RetrieveWithScores 加 filter 参数
func (p *RAGPipeline) RetrieveWithScores(ctx context.Context, query string, filter RAGFilter) ([]ScoredDocument, error) {
    // 透传 filter 给 store.Search
    return p.store.Search(ctx, query, p.topK, filter)
}

// Retrieve 同步更新
func (p *RAGPipeline) Retrieve(ctx context.Context, query string, filter RAGFilter) ([]Document, error) { ... }
```

旧的无 filter 调用（admin 测试）传 `RAGFilter{}`（零值），等价于不过滤。

---

## 5. callctx / main.go 变更

### 5.1 callctx 新增 InternalTenantKey

```go
// callctx/callctx.go
func WithInternalTenantKey(ctx context.Context, key string) context.Context
func GetInternalTenantKey(ctx context.Context) string
```

### 5.2 main.go /api/chat handler 流程

```
1. 提取 X-Tenant-ID、X-Source（默认 "platform_admin"）
2. internalKey := registry.Resolve(source, externalTenantID)
3. ctx = WithInternalTenantKey(ctx, internalKey)
4. ctx = WithTenantID(ctx, internalKey)   // 替换原有的 externalTenantID
5. 后续逻辑不变
```

---

## 6. agent/agent.go 变更

```go
func (a *Agent) ChatStream(ctx context.Context, sessionID, userMessage string) (<-chan StreamEvent, error) {
    tenantKey := callctx.GetInternalTenantKey(ctx)
    if tenantKey == "" { tenantKey = "default" }

    // session key 加前缀（使用 _ 分隔符，避免 / 在 Admin 展示时泄露 internal_key）
    scopedSessionID := tenantKey + "_" + sessionID

    // 构造 RAG filter
    domain := callctx.GetDomain(ctx)  // 已有
    ragFilter := rag.RAGFilter{TenantID: tenantKey, Domain: domain}
    // 若租户配置 allow_cross_domain=true，则 ragFilter.Domain = ""
    if t, _ := registry.Get(tenantKey); t != nil && t.AllowCrossDomain {
        ragFilter.Domain = ""
    }

    // RAG 检索
    scored, err := a.rag.RetrieveWithScores(ctx, userMessage, ragFilter)
    // ...

    // session 操作全部用 scopedSessionID
    a.sessions.GetHistory(scopedSessionID)
    a.sessions.AppendMessage(scopedSessionID, ...)
}
```

Agent 需持有 `*tenant.Registry` 引用（从 main.go 注入）。

---

## 7. Admin API 规范

所有接口需 X-Admin-Token 鉴权（现有 `adminAuthMiddleware` 复用）。

### 7.1 租户配置 API

**GET /api/admin/tenants**
```json
// Response 200
{
  "tenants": [
    {
      "internal_key": "mjarch",
      "display_name": "街机时代",
      "default_domain": "mjarch",
      "allow_cross_domain": false,
      "binding_count": 1
    }
  ]
}
```

**POST /api/admin/tenants**
```json
// Request
{ "internal_key": "mjarch", "display_name": "街机时代", "default_domain": "mjarch", "allow_cross_domain": false }
// Response 201 { "internal_key": "mjarch" }
// Error 409 { "error": "internal_key already exists" }
```

**PUT /api/admin/tenants/{key}**
```json
// Request（可部分更新）
{ "display_name": "街机时代v2", "default_domain": "mjarch", "allow_cross_domain": true }
// Response 200 { "ok": true }
// Error 404 { "error": "tenant not found" }
```

**DELETE /api/admin/tenants/{key}**
```
// Response 200 { "ok": true }
// Error 409 { "error": "tenant has documents, delete them first" }
```

**GET /api/admin/tenants/{key}/bindings**
```json
{ "bindings": [{ "id": 1, "source": "platform_admin", "external_id": "1234567890123456789" }] }
```

**POST /api/admin/tenants/{key}/bindings**
```json
// Request
{ "source": "platform_admin", "external_id": "1234567890123456789" }
// Response 201 { "id": 1 }
// Error 409 { "error": "binding already exists" }
```

**DELETE /api/admin/tenants/{key}/bindings/{id}**
```
// Response 200 { "ok": true }
```

### 7.2 文档管理 API 扩展

**POST /api/admin/documents**（已有，扩展 body）
```json
{
  "doc_id": "mjarch-faq-001",
  "content": "...",
  "tenant_id": "mjarch",   // 新增，必填
  "domain": "mjarch"       // 新增，必填
}
// Error 400 { "error": "tenant_id is required" }
// Error 400 { "error": "domain is required" }
// Error 400 { "error": "tenant not found: xxx" }
```

**GET /api/admin/documents**（已有，扩展 query）
```
GET /api/admin/documents?tenant_id=mjarch&domain=mjarch
// 按 tenant_id + domain 在 Qdrant 列出文档
// tenant_id 为空时列全部（Admin 调试用）
```

---

## 8. Admin UI 交互设计

### 8.1 「🏢 租户配置」section

```
┌─────────────────────────────────────────────┐
│ 🏢 租户配置                        [+ 新增] │
├─────────────────────────────────────────────┤
│ [mjarch]  街机时代  域:mjarch  跨域:否  绑定:1 [展开▼] [编辑] [删除] │
│   └─ 绑定列表：                              │
│      platform_admin | 1234567890123456789 [删除] │
│      [+ 添加绑定]                            │
├─────────────────────────────────────────────┤
│ [default] Default   域:ecommerce 跨域:否  绑定:0 [展开▼] [编辑]     │
└─────────────────────────────────────────────┘
```

- 新增/编辑：卡片内展开内联表单（不弹窗），保存后收起
- 删除：弹 confirm 对话框；有文档时按钮置灰并 tooltip「请先删除该租户下的所有文档」
- 绑定管理：行内添加（source + external_id 两个输入框 + 确认按钮）

### 8.2 「📄 文档管理」section 改造

```
┌────────────────────────────────────────────────────┐
│ 筛选：租户 [mjarch ▼]  域 [mjarch ▼]  [查询]       │
├────────────────────────────────────────────────────┤
│ 上传文档                                            │
│   文档 ID: [____________]                           │
│   租户:    [mjarch    ▼]  ← 从 /api/admin/tenants  │
│   业务域:  [mjarch    ▼]  ← 从 /api/admin/domains  │
│   内容:    [____________________]                   │
│   [上传]（提交时校验，未选租户/域时提示「请选择租户和业务域」）           │
├────────────────────────────────────────────────────┤
│ 文档列表                                            │
│ ID          | 租户   | 域     | 预览        | 操作  │
│ mjarch-faq-1| mjarch | mjarch | 会员有效期… | [删除]│
└────────────────────────────────────────────────────┘
```

---

## 9. 迁移脚本设计

`migration/patch_tenant_domain.go`（`go run` 独立执行）：

```
流程：
1. 连接 Qdrant，Scroll 全量拉取所有 point（无 filter）
2. 筛选出 payload 中没有 tenant_id 字段的 point
3. 打印：「发现 N 个待迁移文档，将补写 tenant_id=default, domain=ecommerce」
4. 提示用户确认（输入 yes 继续）
5. 分批调用 Qdrant SetPayload API 补写 tenant_id + domain
6. 打印：「迁移完成：成功 X，失败 Y」
```

---

## 10. 单测要点

| 文件 | 测试用例 |
|------|----------|
| tenant/registry_test.go | Resolve 命中 binding；Resolve 命中 internal_key；Resolve 兜底 default；Delete 有文档时报错；AddBinding 重复返回 ErrDuplicate；缓存 TTL 过期后重新查 DB |
| storage/qdrant_store_test.go | Search 带 tenant filter 只返回匹配文档；Add 写入 tenant_id payload；Count 带 filter 返回正确数量 |
| agent/agent_test.go | session key 含租户前缀（`_` 分隔）；RAG filter 正确传入 TenantID |

---

## 11. 不变行为清单（回归防护）

| # | 不变行为 | 验证方式 |
|---|---------|---------|
| RG-1 | `RAGFilter{}` 零值时 Search 返回全部文档，与改前一致 | memory_store 单测；现有 ecommerce demo 场景 |
| RG-2 | 现有 ecommerce demo 的 RAG 检索结果不变（迁移脚本执行后 tenant_id=default） | 迁移后手动发一条 ecommerce domain 请求验证 |
| RG-3 | `GET /api/admin/documents` 不带 filter 时返回全部文档 | 现有文档管理页加载测试 |
| RG-4 | 现有 session 过期机制（TTL）不受 session key 格式变更影响 | session 单测验证 TTL 仍生效 |
| RG-5 | strict_tenant_mode=false（默认）时，不带 X-Tenant-ID 的请求仍能正常处理 | AC-2 验收场景 |
