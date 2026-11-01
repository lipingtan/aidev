# 设计文档：知识库域管理 + Soul 管理（CR2）

## 1. 数据模型

### 1.1 agent_tenants 表扩展

```go
// migrate() 中检查列是否存在再 ALTER（SQLite 不支持 ADD COLUMN IF NOT EXISTS）
func (r *Registry) migrateKBColumns() error {
    // 查现有列
    rows, _ := r.db.Query("PRAGMA table_info(agent_tenants)")
    existing := map[string]bool{}
    for rows.Next() {
        var cid int; var name, typ, notnull, dflt, pk string
        rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk)
        existing[name] = true
    }
    rows.Close()
    if !existing["kb_domains"] {
        r.db.Exec(`ALTER TABLE agent_tenants ADD COLUMN kb_domains TEXT NOT NULL DEFAULT '[]'`)
    }
    if !existing["kb_search_policy"] {
        r.db.Exec(`ALTER TABLE agent_tenants ADD COLUMN kb_search_policy TEXT NOT NULL DEFAULT 'fallback'`)
    }
    if !existing["min_kb_hits"] {
        r.db.Exec(`ALTER TABLE agent_tenants ADD COLUMN min_kb_hits INTEGER NOT NULL DEFAULT 2`)
    }
    return nil
}
```

### 1.2 agent_souls 表

```sql
CREATE TABLE IF NOT EXISTS agent_souls (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    content      TEXT NOT NULL,
    tenant_id    TEXT NOT NULL DEFAULT '',   -- 空串=全局；具体值=租户专属
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, tenant_id)
);
```

---

## 2. CR2-1 知识库域管理

### 2.1 tenant/registry.go 扩展

```go
// KB域管理方法
func (r *Registry) GetKBDomains(internalKey string) ([]string, error)
func (r *Registry) AddKBDomain(internalKey, domain string) error
    // 校验 domain 不重复；若 KB域列表为空且 domain 合法则直接 append
func (r *Registry) RemoveKBDomain(internalKey, domain string) error
    // 先调 counter.CountDocuments 检查文档数（带 domain filter），>0 返回 ErrHasDocuments
    // 检查超时（3s）降级：允许删除但打 WARN

// KB检索策略方法
func (r *Registry) GetKBPolicy(internalKey string) (policy string, minHits int, err error)
func (r *Registry) SetKBPolicy(internalKey, policy string, minHits int) error
    // policy 合法值：strict / fallback / cross
```

`Tenant` 结构体扩展：
```go
type Tenant struct {
    // 现有字段...
    KBDomains       []string // 反序列化自 kb_domains JSON
    KBSearchPolicy  string   // "strict" | "fallback" | "cross"
    MinKBHits       int
}
```

### 2.2 RAGFilter 扩展

```go
type RAGFilter struct {
    TenantID string
    Domain   string  // 请求指定的 KB域
    Policy   string  // "strict" | "fallback" | "cross"；空=fallback
    MinHits  int     // fallback 阈值；0=使用默认值2
}
```

### 2.3 rag/rag.go RetrieveWithScores 策略实现

```go
func (p *RAGPipeline) RetrieveWithScores(ctx context.Context, query string, filter RAGFilter) ([]ScoredDocument, error) {
    policy := filter.Policy
    if policy == "" { policy = "fallback" }
    minHits := filter.MinHits
    if minHits <= 0 { minHits = 2 }

    switch policy {
    case "strict":
        // 一次查询，锁定指定域
        return p.store.Search(ctx, query, p.topK, filter)

    case "cross":
        // 只传 TenantID，不传 Domain
        return p.store.Search(ctx, query, p.topK, RAGFilter{TenantID: filter.TenantID})

    case "fallback":
        // 第一次：查指定域
        first, err := p.store.Search(ctx, query, p.topK, filter)
        if err != nil { return nil, err }
        // 有效命中数（通过 score threshold 的）
        valid := countValid(first, p.scoreThreshold)
        if valid >= minHits {
            return first, nil // 够用，不补充
        }
        // 第二次：查其余域（不指定 domain，排除已有结果）
        all, err := p.store.Search(ctx, query, p.topK, RAGFilter{TenantID: filter.TenantID})
        if err != nil { return first, nil } // 降级：返回第一次结果
        // 合并去重，第一次结果优先，按 score 排序，截取 topK
        return mergeDedup(first, all, p.topK), nil
    }
    return p.store.Search(ctx, query, p.topK, filter)
}

func countValid(docs []ScoredDocument, threshold float64) int {
    count := 0
    for _, d := range docs {
        // 只计 Scored=true 且 score 达标的文档；keyword fallback(Scored=false) 不计入
        if d.Scored && d.Score >= threshold { count++ }
    }
    return count
}

func mergeDedup(primary, secondary []ScoredDocument, topK int) []ScoredDocument {
    seen := make(map[string]bool)
    combined := make([]ScoredDocument, 0, len(primary)+len(secondary))
    for _, d := range append(primary, secondary...) {
        if !seen[d.ID] {
            seen[d.ID] = true
            combined = append(combined, d)
        }
    }
    // 按 score 降序排序（Scored=false 的排最后）
    sort.Slice(combined, func(i, j int) bool {
        si := combined[i].Score
        sj := combined[j].Score
        if !combined[i].Scored { si = -1 }
        if !combined[j].Scored { sj = -1 }
        return si > sj
    })
    if len(combined) > topK { combined = combined[:topK] }
    return combined
}
```

### 2.4 agent/agent.go 变更

`ChatStream` 构造 RAGFilter 时读取租户的 KB策略：

```go
tenantKey := callctx.GetInternalTenantKey(ctx)
domain := domainFromContext(ctx)

policy := "fallback"
minHits := 2
if globalTenantRegistry2 != nil {
    if t, err := globalTenantRegistry2.(interface{ Get(string) (*tenantpkg.Tenant, error) }).Get(tenantKey); err == nil && t != nil {
        policy = t.KBSearchPolicy
        minHits = t.MinKBHits
    }
}
ragFilter := rag.RAGFilter{
    TenantID: tenantKey,
    Domain:   domain,
    Policy:   policy,
    MinHits:  minHits,
}
```

### 2.5 Admin API

```
GET    /api/admin/tenants/{key}/kb-domains
    Response: { "kb_domains": ["faq","policy"], "policy": "fallback", "min_hits": 2 }

POST   /api/admin/tenants/{key}/kb-domains
    Body: { "domain": "faq" }
    Error 409: domain already exists
    Error 400: invalid domain name

DELETE /api/admin/tenants/{key}/kb-domains/{domain}
    Error 409: domain has documents (+ count)
    Error 404: domain not found

PUT    /api/admin/tenants/{key}/kb-policy
    Body: { "policy": "fallback", "min_hits": 2 }
    Error 400: invalid policy value
```

---

## 3. CR2-2 Soul 管理

### 3.1 soul/store.go

```go
package soul

type Soul struct {
    ID          int64
    Name        string
    DisplayName string
    Content     string
    TenantID    string // "" = default 租户可用
}

type Store struct { db *sql.DB }

func NewStore(db *sql.DB) (*Store, error)  // AutoMigrate

// Resolve 按优先级查询：当前租户 → "default" 租户 → (nil, nil)
func (s *Store) Resolve(soulName, tenantID string) (*Soul, error)
    // SELECT id,name,display_name,content,tenant_id
    // FROM agent_souls
    // WHERE name=? AND tenant_id IN (?, '')
    // ORDER BY CASE WHEN tenant_id=? THEN 0 ELSE 1 END ASC
    // LIMIT 1
    // 参数：soulName, tenantID, tenantID
    // 无结果时返回 (nil, nil)——不是 error

func (s *Store) List(tenantID string) ([]Soul, error)   // tenantID="" 返回全部
func (s *Store) Create(soul Soul) error                 // UNIQUE(name, tenant_id) 校验
func (s *Store) Update(name, tenantID string, soul Soul) error
func (s *Store) Delete(name, tenantID string) error

var ErrDuplicate = errors.New("soul name already exists for this tenant")
var ErrNotFound  = errors.New("soul not found")
```

### 3.2 callctx 扩展

```go
func WithSoulName(ctx context.Context, name string) context.Context
func GetSoulName(ctx context.Context) string  // 空字符串=未设置
```

### 3.3 main.go 变更

```go
type chatRequest struct {
    // 现有字段...
    SoulName string `json:"soul_name,omitempty"`
}

// handler 中：
if req.SoulName != "" {
    chatCtx = callctx.WithSoulName(chatCtx, req.SoulName)
}
```

启动时初始化 SoulStore（与 TenantRegistry 同一 DB 连接）：
```go
globalSoulStore, _ = soulspkg.NewStore(tenantDB)
```

### 3.4 agent/agent.go system prompt 构建

Agent 结构体新增 `soulStore` 字段，通过 `SetSoulStore` 注入（避免直接依赖全局变量）：

```go
// agent.go
type Agent struct {
    // 现有字段...
    soulStore soulResolver // 接口，避免循环依赖
}

type soulResolver interface {
    Resolve(name, tenantID string) (*soulEntry, error)
}

func (a *Agent) SetSoulStore(store soulResolver) { a.soulStore = store }

// buildSystemPrompt 时
soulName := callctx.GetSoulName(ctx)
systemPromptExtra := domainSystemPromptExtra  // yaml 默认值

if soulName != "" && a.soulStore != nil {
    tenantKey := callctx.GetInternalTenantKey(ctx)
    if s, err := a.soulStore.Resolve(soulName, tenantKey); err == nil && s != nil {
        systemPromptExtra = s.Content
        log.Printf("[Soul] using soul=%s@%s", soulName, s.TenantID)
    } else {
        log.Printf("[WARN][Soul] soul_name=%s not found for tenant=%s, using default", soulName, tenantKey)
    }
}
```

main.go 启动时注入：`ag.SetSoulStore(globalSoulStore)`

### 3.5 Admin API

```
GET    /api/admin/souls?tenant_id=mjarch   Soul 列表（空 tenant_id 返回全部）
POST   /api/admin/souls                    新增
PUT    /api/admin/souls/{name}             编辑（需传 tenant_id 确定唯一性）
DELETE /api/admin/souls/{name}             删除（需传 tenant_id query param）
```

**POST /api/admin/souls Request:**
```json
{
    "name": "friendly",
    "display_name": "贴心小美",
    "content": "你是贴心客服小美，说话活泼可爱，多用表情符号...",
    "tenant_id": "mjarch"   // 空=""=default租户可用
}
```

---

## 4. Admin UI 设计

### 4.1 租户详情页新增「📚 知识库域」区块

```
📚 知识库域
当前域: [faq] [policy] [manual]  [+ 添加域]
检索策略: [fallback ▼]  当命中数 < [2] 时补充其他域

添加新域: [输入域名]  [添加]
```

点击域名旁的 × 删除，有文档时禁用并 tooltip 提示。

### 4.2「🎭 Soul 管理」section（新菜单项）

列表卡片：name、display_name、归属租户、内容前50字预览

新增/编辑表单：
```
名称（英文）: [___________]  归属租户: [全局(default) ▼]
显示名称:    [___________]
人格内容:    [大文本框，多行]
```

---

## 5. 不变行为清单

| # | 不变行为 | 验证方式 |
|---|---------|---------|
| RG-1 | soul_name 为空时，system prompt 与改前完全一致 | AC-10 |
| RG-2 | RAGFilter 零值（Policy=""）时，检索行为与 CR1 改后一致（fallback） | 单测 |
| RG-3 | 现有 ecommerce/mjarch 请求不传 domain，RAG 检索不受 KB域列表影响 | 回归测试 |
| RG-4 | `tenant/registry.go` 其余方法（Resolve/Create/Delete等）不受 KB域字段影响 | tenant 单测 |
