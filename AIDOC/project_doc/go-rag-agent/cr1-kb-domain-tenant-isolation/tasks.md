# 任务拆解：知识库业务域隔离 + 租户隔离（CR1）

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 9 |
| 已完成 | 9 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 9/9 (100%) |
| 当前阶段 | 完成 ✅ |

## 并行批次规划

```
批次1（无依赖，并行）：  T1  T2  T3
批次2（依赖T1+T2+T3）：  T4  T7（并行）→ T5（依赖T4）
批次3（依赖T1+T5）：     T6
批次4（依赖T2+T6+T7）：  T8  T9（并行）
```

---

## Task1：租户注册表（TenantRegistry）✅

**复杂度：** 高

**设计来源：**
- design.md §1.1 agent_tenants 表
- design.md §1.2 agent_tenant_bindings 表
- design.md §2 tenant/registry.go

**Scope（边界）：**
- 涉及文件：`tenant/registry.go`、`tenant/registry_test.go`（新增）
- 不触碰：`session/`、`storage/`、`agent/`

**Constraints（约束）：**
- 使用 `github.com/glebarez/sqlite` 驱动（与 session DB 同一驱动）
- 不依赖 SQLite 外键自动级联（PRAGMA foreign_keys），Delete 方法手动级联删 bindings
- Resolve 结果写入 TTL=60s 内存缓存（sync.Map + 时间戳），增删改时 invalidateCache()
- Delete 调 `store.Count(filter)` 计数，不调 List
- Delete 成功时清理 session 表中 `{key}_%` 前缀的记录

**Acceptance（验证标准）：**
- AC: Resolve 按 binding 命中返回 internal_key
- AC: Resolve 按 external_id == internal_key 直接命中
- AC: Resolve 找不到时返回 "default" 并打 WARN 日志
- AC: Create 重复 internal_key 返回 ErrDuplicate
- AC: AddBinding 重复(source, external_id) 返回 ErrDuplicate
- AC: Delete 有文档时返回 ErrHasDocuments；成功时 session 前缀记录被清理
- AC: Resolve 缓存命中时不触发 SQL（单测验证）
- AC: `go test ./tenant/...` 全部 PASS

**估时：** 2h

---

## Task2：VectorStore 接口 + qdrant_store 扩展 ✅

**复杂度：** 高

**设计来源：**
- design.md §3.1 VectorStore 接口
- design.md §3.2 ensureCollection payload index
- design.md §3.3 Add 写 payload
- design.md §3.4 Search 构造 filter
- design.md §3.5 memory_store 同步

**Scope（边界）：**
- 涉及文件：`storage/store.go`（接口）、`storage/qdrant_store.go`、`storage/memory_store.go`
- 不触碰：`rag/`、`agent/`

**Constraints（约束）：**
- VectorStore 接口新增 `Count(ctx, filter) (int, error)` 方法
- Search / List 签名加 `filter rag.RAGFilter` 参数
- memory_store 的 Count / filter 实现体忽略参数（测试用途，返回零值）
- payload index 建立幂等（已存在时忽略错误）
- TenantID 为空时不加 filter 条件（零值兼容）

**Acceptance（验证标准）：**
- AC: Search 带 tenant_id filter 只返回匹配租户文档
- AC: Search filter 零值时返回全部（RG-1 回归）
- AC: Add 写入文档后 payload 包含 tenant_id 和 domain 字段
- AC: Count 带 filter 返回正确数量
- AC: `go build ./...` 零错误

**估时：** 2h

---

## Task3：rag.Document + RAGFilter + RAGPipeline 更新 ✅

**复杂度：** 中

**设计来源：**
- design.md §1.3 rag.Document 扩展
- design.md §1.4 RAGFilter
- design.md §4 rag/rag.go 变更

**Scope（边界）：**
- 涉及文件：`rag/rag.go`
- 不触碰：`storage/`、`agent/`

**Constraints（约束）：**
- Retrieve / RetrieveWithScores 加 `filter RAGFilter` 参数
- 全局搜索所有调用处，补传 `RAGFilter{}`（零值兼容）
- RAGFilter 零值时行为与改前完全一致（RG-1）

**Acceptance（验证标准）：**
- AC: 所有调用方编译通过
- AC: RAGFilter 零值时行为与改前一致（RG-1 回归）
- AC: `go build ./...` 零错误

**估时：** 1h

---

## Task4：callctx 扩展 + main.go Resolve 注入 ✅

**复杂度：** 中

**依赖：** Task1、Task2、Task3

**设计来源：**
- design.md §5.1 callctx 新增 InternalTenantKey
- design.md §5.2 main.go handler 流程

**Scope（边界）：**
- 涉及文件：`callctx/callctx.go`、`main.go`
- 不触碰：`agent/`

**Constraints（约束）：**
- X-Source 默认值 `"platform_admin"`
- Resolve 结果用 internalKey 替换 ctx 中原有的 externalTenantID
- TenantRegistry 在 main.go 启动时初始化（传入 session DB 连接）

**Acceptance（验证标准）：**
- AC: 请求带已绑定外部 ID 时，ctx 中 internal_key 正确
- AC: 请求不带 X-Tenant-ID 时，internal_key 为 "default"
- AC: `go build ./...` 零错误

**估时：** 1h

---

## Task5：agent.go session key 前缀 + RAGFilter 传入 ✅

**复杂度：** 中

**依赖：** Task4

**设计来源：**
- design.md §6 agent/agent.go 变更

**Scope（边界）：**
- 涉及文件：`agent/agent.go`
- 不触碰：`session/`、`main.go`

**Constraints（约束）：**
- session key 格式：`{internal_key}_{session_id}`（`_` 分隔）
- allow_cross_domain=true 时 ragFilter.Domain = ""
- Agent 结构体新增 `registry *tenant.Registry`，通过 `SetRegistry` 注入

**Acceptance（验证标准）：**
- AC: 两个租户相同 sessionID，GetHistory 各自独立
- AC: RAGFilter 正确带 TenantID
- AC: allow_cross_domain=true 时 Domain 为空
- AC: `go build ./...` 零错误

**估时：** 1.5h

---

## Task6：Admin 租户配置 API ✅

**复杂度：** 高

**依赖：** Task1、Task5

**设计来源：**
- design.md §7.1 租户配置 API

**Scope（边界）：**
- 涉及文件：`admin_api.go`、`main.go`（路由注册）
- 不触碰：`frontend/`

**Constraints（约束）：**
- 所有接口复用 `adminAuthMiddleware`
- Delete 调 `store.Count(filter)` 计数，>0 返回 409
- 路由注册：`/api/admin/tenants`、`/api/admin/tenants/`

**Acceptance（验证标准）：**
- AC: GET /api/admin/tenants 返回租户列表
- AC: POST 重复 key 返回 409
- AC: DELETE 有文档时返回 409
- AC: POST /bindings 重复返回 409
- AC: 不带 X-Admin-Token 返回 401（RG 回归）
- AC: `go build ./...` 零错误

**估时：** 2h

---

## Task7：文档管理 API 扩展 ✅

**复杂度：** 中

**依赖：** Task2、Task3

**设计来源：**
- design.md §7.2 文档管理 API 扩展

**Scope（边界）：**
- 涉及文件：`admin_api.go`（文档相关 handler）
- 不触碰：`frontend/`

**Constraints（约束）：**
- POST 上传：tenant_id、domain 必填，校验 tenant 在注册表中存在
- GET 列表：`?tenant_id=&domain=` 过滤，为空时列全部（RG-3 回归）

**Acceptance（验证标准）：**
- AC: 不带 tenant_id 返回 400
- AC: tenant_id 不在注册表中返回 400
- AC: 上传后 Qdrant payload 含 tenant_id、domain
- AC: GET 带 filter 只返回匹配文档
- AC: GET 不带 filter 返回全部（RG-3 回归）
- AC: `go build ./...` 零错误

**估时：** 1h

---

## Task8：Admin UI — 租户配置页 + 文档管理改造 ✅

**复杂度：** 高

**依赖：** Task6、Task7

**设计来源：**
- design.md §8.1 租户配置 section
- design.md §8.2 文档管理 section 改造

**Scope（边界）：**
- 涉及文件：`frontend/admin.html`
- 不触碰：后端代码

**Constraints（约束）：**
- 复用现有 admin.html 卡片/表单 CSS 风格
- 租户配置页菜单排在「文档管理」之前
- 文档上传提交时校验，未选租户/域时提示「请选择租户和业务域」

**Acceptance（验证标准）：**
- AC: 租户配置页可完成新增租户 + 添加 binding 操作
- AC: 文档上传未选租户/域提交时，提示错误，文档不被上传
- AC: 上传文档后列表刷新，可见 tenant/domain 列
- AC: 筛选后列表只显示匹配文档

**估时：** 3h

---

## Task9：迁移脚本 + 单测补全 ✅

**复杂度：** 中

**依赖：** Task2

**设计来源：**
- design.md §9 迁移脚本设计
- design.md §10 单测要点

**Scope（边界）：**
- 涉及文件：`migration/patch_tenant_domain.go`（新增）、`tenant/registry_test.go`（补全）、`storage/qdrant_store_test.go`（补 filter 用例）
- 不触碰：主程序代码

**Constraints（约束）：**
- 迁移脚本需交互式确认（输入 yes 继续），支持 dry-run
- 分批 SetPayload，每批 ≤100 条

**Acceptance（验证标准）：**
- AC: 脚本 dry-run 打印待迁移条数后等待确认
- AC: 执行后 Qdrant point payload 含 tenant_id=default、domain=ecommerce
- AC: `go test ./tenant/... ./storage/...` 全部 PASS

**估时：** 1.5h

---

## 总估时

| Task | 描述 | 估时 | 依赖 | 并行批次 |
|------|------|------|------|---------|
| Task1 | TenantRegistry | 2h | 无 | 批次1 |
| Task2 | VectorStore 扩展 | 2h | 无 | 批次1 |
| Task3 | RAGFilter | 1h | 无 | 批次1 |
| Task4 | callctx + main Resolve | 1h | T1+T2+T3 | 批次2 |
| Task7 | 文档管理 API | 1h | T2+T3 | 批次2（与T4并行） |
| Task5 | agent session+filter | 1.5h | T4 | 批次2（T4后） |
| Task6 | 租户配置 API | 2h | T1+T5 | 批次3 |
| Task8 | Admin UI | 3h | T6+T7 | 批次4 |
| Task9 | 迁移脚本+单测 | 1.5h | T2 | 批次4（与T8并行） |
| **关键路径** | T1→T4→T5→T6→T8 | **10.5h** | | |
