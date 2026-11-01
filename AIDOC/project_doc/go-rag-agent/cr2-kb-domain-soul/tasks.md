# 任务拆解：知识库域管理 + Soul 管理（CR2）

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 7 |
| 已完成 | 7 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 7/7 (100%) |
| 当前阶段 | 完成 ✅ |

## 并行批次规划

```
批次1（无依赖，并行）：  T1  T2  T3
批次2（依赖T1+T3）：     T4  T5（并行）
批次3（依赖T2+T4+T5）：  T6  T7（并行）
```

---

## Task1：KB域管理扩展 tenant/registry.go ⬜

**复杂度：** 高

**设计来源：**
- design.md §1.1 agent_tenants 表扩展
- design.md §2.1 tenant/registry.go 扩展

**Scope（边界）：**
- 涉及文件：`tenant/registry.go`、`tenant/registry_test.go`
- 不触碰：`soul/`、`rag/`、`agent/`

**Constraints（约束）：**
- migrate() 加 `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`（SQLite 不支持 IF NOT EXISTS，需先查列是否存在）
- `kb_domains` 存 JSON 字符串，Get/Set 时序列化/反序列化
- `kb_search_policy` 合法值：strict/fallback/cross
- RemoveKBDomain 先调 counter.CountDocuments 带 domain filter，>0 返回 ErrHasDocuments；Count 超时3s降级允许删除

**Acceptance（验证标准）：**
- AC: GetKBDomains 返回空列表（新租户）
- AC: AddKBDomain 添加后 GetKBDomains 包含新域
- AC: AddKBDomain 重复域名返回 error
- AC: RemoveKBDomain 有文档时返回 ErrHasDocuments
- AC: SetKBPolicy 设置非法值返回 error
- AC: `go test ./tenant/...` 全部 PASS

**估时：** 1.5h

---

## Task2：RAG 检索策略（strict/fallback/cross）⬜

**复杂度：** 高

**设计来源：**
- design.md §2.2 RAGFilter 扩展
- design.md §2.3 RetrieveWithScores 策略实现

**Scope（边界）：**
- 涉及文件：`rag/rag.go`
- 不触碰：`storage/`、`agent/`

**Constraints（约束）：**
- Policy 空字符串等同 "fallback"
- MinHits 0 等同默认值 2
- `countValid`：只计 `Scored=true` 且 `Score >= threshold` 的文档（Scored=false 的 keyword fallback 文档不阻止 fallback 触发）
- `mergeDedup`：合并后按 score 降序排序（Scored=false 排最后），去重截取 topK
- rag.go 需要 `import "sort"` 供 mergeDedup 使用
- fallback 第二次查询：TenantID 保留，Domain 置空

**Acceptance（验证标准）：**
- AC: strict 策略只查指定域，不补充
- AC: cross 策略忽略 domain，查全租户
- AC: fallback 命中数 >= MinHits 不触发第二次查询
- AC: fallback 命中数 < MinHits 触发补充，结果去重
- AC: `go build ./...` 零错误

**估时：** 1.5h

---

## Task3：Soul Store ⬜

**复杂度：** 高

**设计来源：**
- design.md §1.2 agent_souls 表
- design.md §3.1 soul/store.go

**Scope（边界）：**
- 涉及文件：`soul/store.go`（新增）、`soul/store_test.go`（新增）
- 不触碰：`tenant/`、`agent/`

**Constraints（约束）：**
- 使用 `modernc.org/sqlite` 驱动（与 session/tenant 同驱动）
- Resolve 查询：`WHERE name=? AND tenant_id IN (?, 'default') ORDER BY (tenant_id=当前租户) DESC LIMIT 1`（SQLite 用 CASE 排序）
- content ≤ 2000 字符校验在 Create/Update 里
- Resolve 精确 SQL：`SELECT id,name,display_name,content,tenant_id FROM agent_souls WHERE name=? AND tenant_id IN (?,  '') ORDER BY CASE WHEN tenant_id=? THEN 0 ELSE 1 END ASC LIMIT 1`；无结果返回 `(nil, nil)`
- UNIQUE(name, tenant_id) 冲突返回 ErrDuplicate

**Acceptance（验证标准）：**
- AC: Resolve 当前租户有 soul → 返回租户 soul
- AC: Resolve 当前租户无 soul，default 有 → 返回 default soul
- AC: Resolve 都没有 → 返回 nil, nil（非 error）
- AC: Create 重复返回 ErrDuplicate
- AC: content > 2000 字符返回 error
- AC: `go test ./soul/...` 全部 PASS

**估时：** 1.5h

---

## Task4：Admin API 扩展（KB域 + Soul CRUD）⬜

**复杂度：** 高

**依赖：** Task1、Task3

**设计来源：**
- design.md §2.5 Admin API（KB域）
- design.md §3.5 Admin API（Soul）

**Scope（边界）：**
- 涉及文件：`admin_api.go`、`main.go`（路由注册）
- 不触碰：`frontend/`

**Constraints（约束）：**
- KB域 API 全部需要 adminAuthMiddleware
- Soul API 同样需要鉴权
- Soul 的唯一性由 (name, tenant_id) 联合确定
- Soul content sanitize：去除 `---` prompt 分隔注入

**Acceptance（验证标准）：**
- AC: GET /kb-domains 返回域列表+策略
- AC: POST /kb-domains 重复返回 409
- AC: DELETE /kb-domains/{domain} 有文档返回 409
- AC: PUT /kb-policy 非法值返回 400
- AC: GET /souls 返回 soul 列表
- AC: POST /souls 重复返回 409
- AC: `go build ./...` 零错误

**估时：** 2h

---

## Task5：main.go soul_name 注入 + callctx 扩展 ⬜

**复杂度：** 中

**依赖：** Task3

**设计来源：**
- design.md §3.2 callctx 扩展
- design.md §3.3 main.go 变更

**Scope（边界）：**
- 涉及文件：`callctx/callctx.go`、`main.go`
- 不触碰：`agent/`

**Constraints（约束）：**
- chatRequest 加 `SoulName string json:"soul_name,omitempty"`
- 启动时初始化 globalSoulStore（与 TenantRegistry 同一 DB 连接）
- soul_name 注入 ctx，不在 main.go 做查询（查询在 agent.go 做）

**Acceptance（验证标准）：**
- AC: 请求带 soul_name 时 callctx.GetSoulName(ctx) 返回对应值
- AC: 请求不带 soul_name 时 GetSoulName 返回空字符串
- AC: `go build ./...` 零错误

**估时：** 0.5h

---

## Task6：agent.go soul 覆盖 system_prompt_extra ⬜

**复杂度：** 中

**依赖：** Task5

**设计来源：**
- design.md §3.4 agent.go system prompt 构建
- design.md §2.4 agent.go KB策略读取

**Scope（边界）：**
- 涉及文件：`agent/agent.go`
- 不触碰：`main.go`、`soul/`

**Constraints（约束）：**
- Agent 结构体加 `soulStore soulResolver` 接口字段，通过 `SetSoulStore` 注入（main.go 注入 globalSoulStore）
- soul_name 非空时调 `a.soulStore.Resolve(name, tenantKey)`，找到则替换 system_prompt_extra
- 找不到 soul 打 WARN 日志，不报错，继续用 yaml 默认
- RAGFilter 构造加入 Policy/MinHits（从 TenantRegistry.Get 读取）
- 不改变 soul_name="" 时的现有行为（RG-1）

**Acceptance（验证标准）：**
- AC: soul_name="" 时 system prompt 与改前一致（RG-1）
- AC: soul_name="friendly" 且 soul 存在时，system prompt 包含 soul content
- AC: soul_name 不存在时，WARN 日志，使用 yaml 默认
- AC: RAGFilter.Policy 从租户配置读取
- AC: `go build ./...` 零错误

**估时：** 1h

---

## Task7：Admin UI（KB域管理 + Soul管理）⬜

**复杂度：** 高

**依赖：** Task4

**设计来源：**
- design.md §4.1 租户详情页 KB域区块
- design.md §4.2 Soul管理 section

**Scope（边界）：**
- 涉及文件：`frontend/admin.html`
- 不触碰：后端代码

**Constraints（约束）：**
- 租户详情页 KB域区块插入现有租户详情（renderTenantDetail）
- 文档管理 `onDocTenantChange` 改为调 `/api/admin/tenants/{key}/kb-domains` 取域列表（而非 `/api/admin/domains`）；租户无 KB域时域下拉显示「该租户暂无知识库域，请先在租户配置页添加」并禁止上传
- Soul 管理新增菜单项「🎭 Soul 管理」；列表空态提示「暂无 Soul，新增后调用方可通过 soul_name 参数指定」

**Acceptance（验证标准）：**
- AC: 租户详情页可添加/删除 KB域
- AC: 文档上传域下拉只显示当前租户的 KB域
- AC: 租户无 KB域时域下拉提示「请先添加知识库域」，上传按钮不可用（AC-13）
- AC: Soul 管理页可新增/删除 soul；无 soul 时显示空态提示（AC-14）
- AC: Soul 管理页菜单排在「租户配置」之后、「文档管理」之前
- AC: script 标签 1开1闭（无泄漏）

**估时：** 2.5h

---

## 总估时

| Task | 描述 | 估时 | 批次 |
|------|------|------|------|
| T1 | KB域 Registry 扩展 | 1.5h | 1 |
| T2 | RAG 检索策略 | 1.5h | 1 |
| T3 | Soul Store | 1.5h | 1 |
| T4 | Admin API | 2h | 2 |
| T5 | callctx + main.go | 0.5h | 2 |
| T6 | agent.go soul+策略 | 1h | 3 |
| T7 | Admin UI | 2.5h | 3 |
| **合计** | | **10.5h** | |
