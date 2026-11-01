# 需求计划：知识库业务域隔离 + 租户隔离

## 需求理解

**目标：** go-rag-agent 改造为支持多租户隔离和知识库按业务域分域管理，并增加业务系统租户 ID 到 go-rag-agent 内部租户的映射层。

**背景与现状：**
- go-rag-agent 作为独立 AI 客服引擎，可被多个业务系统（如 platform_admin）调用，每个业务系统有自己的租户体系（如 platform_admin 的租户 ID 是雪花 ID）。
- 当前 Qdrant 所有文档混存在同一个 collection，无 tenant_id / domain 过滤，RAG 检索会跨租户污染。
- 当前 session key 不含租户前缀，不同租户的相同 session_id 会冲突。
- 本体论（Ontology）描述企业级业务实体定义（Order / Customer / Member 等），是全局共享的，不按租户或业务域拆分——本 CR 不改变这一层。

**三层职责澄清：**

| 层 | 职责 | 隔离维度 |
|----|------|----------|
| 本体论（Ontology） | 企业级业务实体定义（Order、Member…） | 全局共享，不隔离 |
| 知识库（Knowledge Base） | FAQ、政策、说明文档 | 按 tenant（内部 key）+ domain 双维度隔离 |
| 工具/Domain 配置 | 业务域工具绑定（query_orders 等） | 按 domain 隔离（已有，不变） |

**改造范围：**
1. **租户映射层**：维护业务系统外部租户 ID → go-rag-agent 内部租户 key 的映射注册表，Admin 页面可增删改，支持多个业务系统的不同租户 ID 格式。
2. **知识库域隔离**：文档挂载到指定租户 + 业务域下；RAG 检索按当前请求的内部租户 key + domain 过滤。
3. **租户隔离**：Session key 加租户前缀；RAG 检索强制过滤当前租户。
4. **Admin 管理界面**：新增「租户配置」页（映射管理）；「文档管理」页改造支持选择租户+域上传/筛选。

**预期效果：**
- 业务系统传来外部租户 ID（如 `1234567890123456789`），go-rag-agent 映射为内部 key（如 `mjarch`）后全程使用内部 key 隔离
- mjarch 的会员政策文档不会出现在 ecommerce 的 RAG 检索结果中
- 不同租户的相同 session_id 不会互相干扰
- 管理员在 Admin 页面可为每个租户/域单独维护知识文档

---

## 租户映射层设计

```
请求进来：
  X-Tenant-ID: "1234567890123456789"   ← 业务系统的外部租户 ID
  X-Source: "platform_admin"            ← 来源系统（可选，默认 platform_admin）

          ↓ TenantRegistry.Resolve(source, externalID)

内部租户记录：
  internal_key: "mjarch"
  display_name: "街机时代"
  default_domain: "mjarch"
  bindings:
    - source: "platform_admin"
      external_id: "1234567890123456789"

          ↓

后续全部用 internal_key="mjarch"：
  RAG filter:   tenant_id = "mjarch"
  session key:  "mjarch/{session_id}"
  文档 payload: tenant_id = "mjarch", domain = "mjarch"
```

**找不到映射时的降级策略：**
- 外部 ID 直接匹配某个 internal_key（即业务系统直接传内部 key）→ 直接用
- 都找不到 → 兜底 `"default"`（调试模式）或 400（严格模式，由配置控制）

---

## 假设列表

- [假设-1] Qdrant 向量库继续使用，通过 payload filter 实现隔离（而非为每个租户建独立 collection）；collection 维持现有结构，新增 `tenant_id`（内部 key）、`domain` 两个 payload 字段。
- [假设-2] 现有已存储的文档迁移时补 `tenant_id="default", domain="ecommerce"` payload；"default" 租户需在租户注册表中有对应记录（含 external_id 映射），确保请求进来能映射上，与知识库文档的 tenant_id 保持一致。
- [假设-3] 租户注册表（TenantRegistry）存储在 SQLite（复用现有 session DB 同一数据库文件），Admin API 写后热加载，无需重启。
- [假设-4] 知识库文档上传时，管理员在 Admin 页面选择归属的 tenant（内部 key）和 domain，不由业务系统自动推断。
- [假设-5] domain 仍由请求参数传入（现有机制不变），租户级 default_domain 作为请求未传 domain 时的兜底。
- [假设-6] 本体论（Ontology classes）全局共享，不按租户/域拆分，本 CR 不触碰 ontology 层代码。
- [假设-7] 工具调用的租户隔离（行级数据权限）由 mcp-server-platform 的 JWT 鉴权保证（已有），go-rag-agent 侧只需正确透传 tenant_id 到工具 context。

---

## 澄清问题

- [Question-1] 知识库隔离粒度：tenant 强制过滤 + domain 强制过滤，还是 tenant 强制、domain 可选（同一租户内允许跨 domain 检索）？
  推荐：tenant_id 强制过滤（安全边界），domain 可选过滤（默认用当前请求的 domain；每个租户可配置 allow_cross_domain=true 允许跨域检索）。
  [Answer-1]
ok
- [Question-2] 现有 ecommerce demo 知识文档（5 条）是否迁移补 tenant_id 字段，还是直接清空重建？
  推荐：写一次性迁移脚本补 `tenant_id="default", domain="ecommerce"` payload，不重新 embedding。
  [Answer-2]
ok, 注意和租户配置那里保持一致，go rag agent可以使用自己的租户id，但要提供业务域和租户id的映射关系，以便请求进来时能够映射上租户
- [Question-3] Admin 文档管理页（现有 `/api/admin/documents`）：上传时租户/域选择器是自由输入，还是只能从已注册的租户+域列表选择？
  推荐：只能从已注册租户列表选择（防止误建游离文档）；domain 从该租户的 default_domain + 已有域配置中选。
  [Answer-3]
OK
- [Question-4] Session key 租户前缀方案：`{internal_key}/{original_key}` 还是其他格式？现有 session 数据是否迁移？
  推荐：`{internal_key}/{original_key}`；现有数据不迁移（会话短生命周期，过期自然淘汰）。
  [Answer-4] OK，用 `{internal_key}/{original_key}` 格式；现有 session 不迁移。

- [Question-5] 请求 tenant_id 为空（如直接访问 index.html 调试）时：兜底 `"default"` 还是 400 拒绝？由哪个配置控制？
  推荐：兜底 `"default"`，由 `strict_tenant_mode=false`（默认）控制；设为 true 时空 tenant_id 返回 400。
  [Answer-5]
ok
- [Question-6] 租户注册表的存储格式：SQLite 表（与 session DB 共用）还是独立 JSON 文件（热加载更简单）？
  推荐：独立 JSON 文件 `data/tenants.json`，Admin API 写文件 + 热加载；数量少、变更不频繁，不值得上 DB。
  [Answer-6]
sqlite
---

## 非功能需求建议

- **性能：** Qdrant payload filter 走索引，检索延迟增加 < 5ms；tenant_id、domain 字段需在 Qdrant collection 建立 payload index。
- **向后兼容：** 现有 ecommerce demo 场景（tenant=default）不受影响。
- **安全：** 租户注册表写操作（增删改）、文档写操作（上传/删除）均需 X-Admin-Token 鉴权；RAG 检索由 context 的 tenant_id 自动隔离，无需额外校验。
- **可观测性：** RAG 检索日志增加 `[RAG] tenant=mjarch domain=mjarch` 信息；租户映射失败打 WARN 日志。

---

## 影响范围预判

**新增：**
- `tenant/registry.go`：TenantRegistry（注册表 CRUD、Resolve 方法、SQLite 存储、热加载）
- `tenant/registry_test.go`：单测
- `storage/qdrant_store.go`：Search 加 tenant_id + domain payload filter 参数；Add 写 tenant_id + domain payload；初始化时建 payload index
- `admin_api.go`：租户配置 CRUD API（`/api/admin/tenants`）；文档管理 API 加 tenant/domain filter 和上传参数
- `migration/patch_tenant_domain.go`：一次性脚本补全现有文档的 tenant_id/domain payload（"default" 租户需在注册表中预先存在）

**修改：**
- `main.go`：请求进来后调 TenantRegistry.Resolve 得到内部 key，替换 context 里的 tenant_id
- `agent/agent.go`：session key 加 `{internal_key}/` 前缀；RAG 调用传 tenant_id + domain
- `rag/rag.go`：Retrieve / RetrieveWithScores 接收 tenant_id + domain filter 参数
- `frontend/admin.html`：新增「🏢 租户配置」section（注册表增删改）；「📄 文档管理」section 加租户/域选择器和列表列

**不触碰：**
- `ontology/`（本体论全局共享）
- `framework/tool_executor.go`（工具调用租户隔离已由 JWT 保证）
- `config/domains/*.yaml`（domain 配置不变）
- mcp-server-platform（不变）
