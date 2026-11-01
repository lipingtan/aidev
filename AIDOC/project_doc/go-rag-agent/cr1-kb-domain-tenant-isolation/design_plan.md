# 设计计划：知识库业务域隔离 + 租户隔离（CR1）

## 模块划分

```
tenant/
  registry.go          TenantRegistry（SQLite CRUD + Resolve）
  registry_test.go

storage/
  qdrant_store.go      Search / Add 加 filter / payload 字段（改）
  qdrant_store_test.go

rag/
  rag.go               RAGFilter 参数透传（改）

agent/
  agent.go             session key 加前缀；RAG 调用传 filter（改）

main.go                请求进来后调 Resolve 注入 internal_key（改）
admin_api.go           租户 CRUD API + 文档管理 API 扩展（改）
frontend/admin.html    租户配置页 + 文档管理改造（改）

migration/
  patch_tenant_domain.go  一次性脚本（新增）
```

---

## 各模块设计要点

### tenant/registry.go

- **两张表**（AutoMigrate 到 session DB 同一 SQLite 文件）：
  - `agent_tenants`：internal_key（UNIQUE）、display_name、default_domain、allow_cross_domain
  - `agent_tenant_bindings`：tenant_id FK、source、external_id，UNIQUE(source, external_id)
- **Resolve(source, externalID) string**：先查 bindings，未命中则检查 externalID 是否等于某 internal_key，兜底返回 `"default"`
- **启动时 seed**：确保 `internal_key="default", default_domain="ecommerce"` 存在（INSERT OR IGNORE）
- **热加载**：增删改直接写 DB，Resolve 每次查 DB（行级加锁），无需重启；若有性能压力后续可加短 TTL 缓存

### storage/qdrant_store.go

**Add 改造**：
- `rag.Document` 增加 `TenantID`、`Domain` 两个字段（或通过 `Metadata` map 传入）
- 写入 payload 时加 `tenant_id`、`domain` 两个 key

**Search 改造**：
- 签名加 `filter RAGFilter` 参数
- 构造 Qdrant `Filter`：tenant_id must match（MatchKeyword）；domain 非空时加 domain must match
- collection 初始化（`ensureCollection`）时为这两个字段建 payload index

**接口变化**：`VectorStore` interface 的 `Search` 方法签名需同步更新；`memory_store.go` 的 mock 实现也需对应调整（加 filter 参数，忽略即可）。

### rag/rag.go

- `RAGFilter` 结构体：`TenantID string`、`Domain string`
- `Retrieve` / `RetrieveWithScores` 新增 `filter RAGFilter` 参数，透传给 `store.Search`
- 旧的无 filter 调用路径（如 admin 测试接口）传零值 `RAGFilter{}`，行为不变（无 filter = 全库搜索）

### agent/agent.go

- `ChatStream` / `Chat` 从 ctx 取 `internal_key`（新增 `WithInternalTenantKey` / `GetInternalTenantKey` callctx 辅助函数）
- session key 包装：`tenantKey + "/" + sessionID`（仅 AppendMessage / GetHistory / Clear 调用前包装，SessionManager 本身不动）
- RAG 调用构造 `RAGFilter{TenantID: tenantKey, Domain: domain}`；若租户 `allow_cross_domain=true` 则 `Domain=""` 

### main.go

- `/api/chat` handler 现有流程：提取 `X-Tenant-ID` + `X-Source` → 调 `TenantRegistry.Resolve` → 得到 `internal_key` → 注入 ctx（`WithInternalTenantKey`）
- `X-Source` 默认值 `"platform_admin"`
- strict_tenant_mode 由全局配置读取（现有 admin config 机制扩展）

### admin_api.go

新增租户 CRUD handler，复用现有 `jsonOK` / `jsonError` 辅助函数：
```
GET    /api/admin/tenants
POST   /api/admin/tenants
PUT    /api/admin/tenants/{key}
DELETE /api/admin/tenants/{key}         检查文档数，>0 返回 409
GET    /api/admin/tenants/{key}/bindings
POST   /api/admin/tenants/{key}/bindings
DELETE /api/admin/tenants/{key}/bindings/{id}
```

文档管理 API 扩展：
- `POST /api/admin/documents`：body 新增 `tenant_id`（必填）、`domain`（必填）校验
- `GET /api/admin/documents`：支持 `?tenant_id=&domain=` 过滤（传给 Qdrant List filter）

### frontend/admin.html

**「🏢 租户配置」section**（菜单项插在「文档管理」前）：
- 列表卡片：internal_key、display_name、default_domain、allow_cross_domain toggle、bindings 数
- 新增/编辑内联表单（无弹窗，卡片展开形式，符合现有 admin.html 风格）
- 展开 bindings：source + external_id 列表，可添加/删除行

**「📄 文档管理」section 改造**：
- 上传表单加两个 select：租户（fetch `/api/admin/tenants` 填充）、域（fetch `/api/admin/domains` 填充）
- 列表加「租户」「域」两列
- 顶部加筛选 select（租户、域），变化时重新 fetch 列表

---

## 模块依赖关系

```
main.go
  └── tenant/registry.go     （新）
  └── agent/agent.go
        └── rag/rag.go
              └── storage/qdrant_store.go  （改 Search/Add）
        └── session/          （不动）
  └── admin_api.go
        └── tenant/registry.go
        └── storage/qdrant_store.go
```

---

## 关键决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 租户注册表存储 | SQLite（复用 session DB） | 与 session 同文件，运维简单；数据量极小 |
| Qdrant 隔离方式 | payload filter（单 collection） | 新增 collection 需重新 embedding；payload filter 成本低 |
| Resolve 是否缓存 | 先不缓存，直接查 DB | 租户数量少，SQLite 查询 < 1ms；有压力再加缓存 |
| VectorStore 接口变更 | Search 加 filter 参数 | 破坏性变更，需同步修改 memory_store mock；影响范围明确可控 |
| Session key 迁移 | 不迁移，自然过期 | 会话短生命周期（TTL 已有），迁移收益低于成本 |

---

## 设计风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| VectorStore 接口签名变更影响调用方 | memory_store / 测试代码需同步改 | 改动范围明确（2 个文件），统一改 |
| 现有 ecommerce demo 文档迁移失败 | 检索不到文档，RAG 退化 | 迁移脚本执行前验证条数；保留回滚方案（清空 payload 字段） |
| Qdrant payload index 未建导致 filter 走全扫 | 检索变慢 | `ensureCollection` 里检查并建 index，启动时保证 |
| strict_tenant_mode=true 误配导致所有请求 400 | 服务不可用 | 默认 false；Admin 页改动有即时生效提示 |
