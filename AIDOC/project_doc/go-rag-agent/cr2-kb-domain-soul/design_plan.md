# 设计计划：知识库域管理 + Soul 管理（CR2）

## 设计方向

### CR2-1 知识库域管理

`agent_tenants` 表加两个字段：
- `kb_domains`：JSON 数组，如 `["faq","policy","manual"]`
- `kb_search_policy`：枚举字符串 strict/fallback/cross，默认 fallback
- `min_kb_hits`：整数，fallback 触发阈值，默认 2

RAG 检索策略在 `rag/rag.go` 的 `RetrieveWithScores` 里实现，根据 `RAGFilter.Policy` 决定：
- strict：一次查询，filter 锁定指定域
- fallback：先查指定域，命中数 < MinHits 时再查其余域补充
- cross：filter 只传 TenantID，不传 Domain

文档管理域下拉从 `/api/admin/tenants/{key}/kb-domains` 取，完全独立于工具域。

### CR2-2 Soul 管理

独立 `soul/` 包，SQLite 存储（与 session DB 同文件），查询顺序：当前租户 → default 租户 → yaml fallback。

---

## 澄清问题

- [Question-1] fallback 策略补充查询时，补充的文档和第一次的文档如何合并排序？
  推荐：先按 score 降序放第一次结果，再追加第二次结果（score 降序），去重后截取 topK。
  [Answer-1] ok，按推荐方案

- [Question-2] `kb_domains` 列表变更（删除某个域）时，已有文档的 Qdrant payload 不会自动更新——是否需要校验？
  推荐：删除 KB域前调 `store.Count(filter{tenant+domain})` 检查，有文档则 409 禁止删除。
  [Answer-2] ok

- [Question-3] Soul 表是否需要版本控制（修改 content 后旧请求不受影响）？
  推荐：本期不做版本控制，修改即生效；后续可加 version 字段。
  [Answer-3] ok，不做版本控制

---

## 风险点

- fallback 策略需要两次 Qdrant 查询，P99 延迟会增加（约 +50ms）；可以并行查询缓解
- KB域删除保护依赖 Qdrant Count API，如果 Qdrant 不可达会阻塞删除操作——需要加超时降级
