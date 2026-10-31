# CR14 任务清单

> 设计文档：[design.html](./design.html)  
> 规范参考：`go-task-representation.md`、`go-development-workflow.md`  
> 最后更新：2026-10-31  
> 前置已完成：go-rag-agent 流式输出改造（ST-1~ST-4，2026-10-31 验证通过）

---

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 18 |
| 已完成 | 0 |
| 进行中 | 0 |
| 未开始 | 18 |
| 完成率 | 0/18 (0%) |
| 当前阶段 | P0: 主链路 |

---

## 阶段说明

| 阶段 | 目标 | 交付标准 |
|------|------|---------|
| **P0 主链路** | platform_admin → Agent → MCP → platform_admin 完整调用链跑通 | 用户发消息，机器人自动回复，可转人工 |
| **P1 功能完整** | 禁言、Agent 开关前端、策略配置 | 管理端可配置并操作所有 CR14 功能 |
| **P2 配置化** | TokenManager 可配、Session 持久化可配 | Agent 凭证和会话存储均可在管理页调整 |

---

## 新增需求说明（2026-10-31）

### Agent 控制菜单（`agent.enabled` + `agent.base_url`）

- 默认关闭，关闭时 platform_admin 完全不感知机器人存在
- 每个租户**独立配置**（cs_configs 表按 tenant_id 隔离）
- 内嵌对话测试面板，直连 Agent，不影响生产会话
- 测试面板顶部显示「🔧 调试工具，不受启用开关控制」说明

### 关键设计决策

| 决策点 | 方案 |
|--------|------|
| session_id 格式 | `cs-{cs_session_id}`；context 传 `cs_session_id` 原始 ID |
| MCP token 注入 | go-rag-agent 在请求头带 `X-Agent-Token`，MCP Server 透传为 `Authorization: Bearer` |
| 关键词匹配 | 包含匹配（`strings.Contains`），OR 关系 |
| 策略3通知文案 | 单独 `handoff.max_turns_msg`，与策略2的 `no_answer_msg` 分开 |
| cs_configs 缓存 | TTL=30s 进程内缓存，多 Pod 最大不一致窗口=30s |
| TokenManager 多 Pod | 每 Pod 独立维护，刷新失败 fallback 重登录，无需分布式锁 |
| goroutine 并发 | per-tenant 信号量，上限=10（可配 `agent.max_concurrent`） |
| 测试 session 清理 | prefix=`admin-test-` 固定 TTL=1h，独立于 ttl_days |

---

## P0 — 主链路任务

### Task 1: 新建 cs_configs 键值表及配置服务 ⬜

**复杂度**: 高

**设计来源:** design.html §5 平台管理端「AI 客服机器人」区块；§5 转人工策略配置表

**Scope（边界）:**
- 涉及文件:
  - `app/license/models/models.go`（修改，新增 CsConfig model）
  - `app/license/service/cs_config_service.go`（新增）
  - `app/license/plugin.go`（修改，AutoMigrate 加入 CsConfig，注入默认值）
- 涉及模块: license
- 不触碰: 其他已有 service、现有 model 字段

**Constraints（约束）:**
- `cs_configs` 表结构：`(id, tenant_id, key VARCHAR(64), value TEXT, updated_at)`，唯一索引 `uk_cs_configs(tenant_id, key)`
- 读取带进程内缓存，**TTL=30s**（多 Pod 最大不一致窗口），写入时失效本 Pod 缓存（其他 Pod 靠 TTL 自动刷新）
- 提供三个类型安全读方法：`GetString / GetBool / GetInt`；写入时做格式校验（bool 只允许 "true"/"false"，int 只允许数字）
- `plugin.go` 启动时用 UPSERT 写入 8 个默认配置项（不覆盖已有值）
- `value` 列类型为 TEXT，MySQL 8.x 禁止 TEXT 列有 DEFAULT，`gorm` tag 不加 `default`

**Acceptance（验证标准）:**
- AC: `cs_configs` 表自动创建，唯一索引存在
- AC: `GetBool(tenantID, "agent.enabled", false)` 返回 false（默认值）
- AC: `SetConfig` 写入后，缓存失效，下次 `GetString` 返回新值
- AC: 写入 bool 时传 "True" 应被规范化为 "true"
- AC: `go build ./...` 零错误
- AC: `go vet ./...` 无警告

**自测:**
- 测试文件: `app/license/service/cs_config_service_test.go`
- ST: GetBool 键不存在 → 返回 defaultVal
- ST: SetConfig 写入 "true" → GetBool 返回 true
- ST: SetConfig 写入非法 bool 值 → 返回 error
- ST: GetInt 写入 "10" → 返回 10
- ST: GetString TTL 超时后 → 重新读 DB 返回最新值

---

### Task 2: agent_bot 账号、ai_agent 角色及权限点迁移 ⬜

**复杂度**: 中

**设计来源:** design.html §3 Agent 鉴权配置化机制；§7 MCP Server 工具清单

**Scope（边界）:**
- 涉及文件:
  - `app/license/migrations/`（新增迁移脚本 `cr14_ai_agent_permissions.sql`）
  - 管理端后台（手工操作，创建账号和角色绑定）
- 涉及模块: license/migrations
- 不触碰: 现有权限点、现有角色、其他账号

**Constraints（约束）:**
- 迁移脚本使用 `INSERT IGNORE` 写入权限点，防重复执行
- 权限点标识符：`cs:bot:send_message`、`cs:bot:handoff`、`cs:query:member_info`、`cs:query:orders`
- 管理端手工步骤：创建 `ai_agent` 角色并绑定以上 4 个权限点；创建 `agent_bot` 账号并分配该角色

**Acceptance（验证标准）:**
- AC: 执行迁移脚本后，`permissions` 表中存在 4 个权限点
- AC: 用 `agent_bot` 账号登录，JWT 中含 `ai_agent` 角色
- AC: 用该 JWT 调用 `POST /admin/cs/sessions/:id/messages` → 200（不报 401/403）

---

### Task 3: InternalSendBotMessage / InternalHandoffToHuman 改 JWT 鉴权 ⬜

**复杂度**: 中

**设计来源:** design.html §7 MCP Server；§3 Agent 鉴权配置化机制

**Scope（边界）:**
- 涉及文件:
  - `app/license/apis/customer_service_api.go`（修改，移除 checkInternalToken，改为 JWT 权限点校验）
- 涉及模块: license/apis
- 不触碰: 其他接口、现有 JWT 中间件逻辑

**Constraints（约束）:**
- 使用项目现有的权限中间件校验 `cs:bot:send_message`（send_message 接口）和 `cs:bot:handoff`（handoff 接口）
- 移除 `checkInternalToken` 函数调用，不保留旧逻辑
- 错误返回标准格式：401（无 token）/ 403（权限不足）

**Acceptance（验证标准）:**
- AC: 无 Authorization header → 401
- AC: 无效 JWT → 401
- AC: JWT 有效但无对应权限点 → 403
- AC: agent_bot JWT 调用 → 200
- AC: `go build ./...` 零错误
- AC: 【回归】现有人工客服接口不受影响（RG-1）

**自测:**
- 测试文件: `app/license/apis/customer_service_api_test.go`
- ST: 无 token 调用 InternalSendBotMessage → 401
- ST: agent_bot token 调用 → 200 + 消息写入 DB
- ST: 【回归】人工客服发消息接口仍正常（RG-1）

---

### Task 4: CsAgent 新增 agent_type 字段，新会话绑定 bot ⬜

**复杂度**: 高

**设计来源:** design.html §2 CsAgent agent_type 字段；§2 CsSession mode 状态流转

**Scope（边界）:**
- 涉及文件:
  - `app/license/models/models.go`（修改，CsAgent 加 AgentType 字段）
  - `app/license/service/customer_service.go`（修改，新建会话逻辑读 agent.enabled 决定 mode）
  - `app/license/plugin.go`（修改，AutoMigrate 含新字段）
- 涉及模块: license/models, license/service
- 不触碰: 现有会话查询逻辑、转人工逻辑

**Constraints（约束）:**
- `AgentType` 字段：`gorm:"type:varchar(20);not null;default:'human'"` 
- 新建会话时读 `cs_config_service.GetBool(tenantID, "agent.enabled", false)`：true → `mode="bot"`，`agent_id=<bot CsAgent ID>`；false → `mode="human"`，`agent_id=NULL`
- bot CsAgent ID 从 `cs_agents` 表查询 `agent_type='bot'` 的第一条，查不到时降级为 `mode="human"`（不 panic）
- MySQL 8.x：新增字段有 default，AutoMigrate 时使用 `ALTER TABLE ADD COLUMN IF NOT EXISTS`

**Acceptance（验证标准）:**
- AC: `cs_agents` 表新增 `agent_type` 列，默认值 'human'
- AC: `agent.enabled=true` 时新建会话 → `mode="bot"`，`agent_id=<bot id>`
- AC: `agent.enabled=false` 时新建会话 → `mode="human"`，`agent_id=NULL`
- AC: 查不到 bot CsAgent 时 → 降级 `mode="human"`，不报错
- AC: `go build ./...` 零错误
- AC: 【回归】现有会话创建流程其余字段不变（RG-2）

**自测:**
- 测试文件: `app/license/service/customer_service_test.go`
- ST: enabled=true，有 bot agent → 新会话 mode=bot
- ST: enabled=false → 新会话 mode=human
- ST: enabled=true 但无 bot agent → 新会话 mode=human（降级）
- ST: 【回归】现有会话字段 tenant_id/user_id 赋值正确（RG-2）

---

### Task 5: platform_admin MCP Server（独立进程 :9001） ⬜

**复杂度**: 高

**设计来源:** design.html §7 platform_admin MCP Server；Token 注入机制

**Scope（边界）:**
- 涉及文件:
  - `mcp-server-platform/main.go`（新增，独立 Go 模块）
  - `mcp-server-platform/go.mod`（新增）
  - `deploy/start-mcp-server.sh`（新增，启动脚本）
- 涉及模块: 独立进程，不依赖 platform_admin 代码包
- 不触碰: 现有 `mcp-server/` 目录

**Constraints（约束）:**
- 从请求 header `X-Agent-Token` 取 token，转发 platform_admin 时换为 `Authorization: Bearer <token>`；`X-Agent-Token` 缺失时返回 401
- 4 个工具路由：`POST /tools/send_bot_message/execute`，`/tools/handoff_to_human/execute`，`/tools/query_member_info/execute`，`/tools/query_orders/execute`
- `query_orders` 默认 `page_size=5`（防止大量数据撑爆 LLM context）
- 环境变量：`PLATFORM_ADMIN_URL`（默认 `http://127.0.0.1:8080`），`MCP_SERVER_PORT`（默认 `9001`）
- `GET /` 健康检查，返回 `{"status":"ok","name":"platform-admin-mcp"}`
- `GET /tools` 列出工具定义（含参数 JSON Schema）

**Acceptance（验证标准）:**
- AC: `GET /` → `{"status":"ok"}`
- AC: `GET /tools` → 返回 4 个工具定义
- AC: 无 `X-Agent-Token` → 401
- AC: `POST /tools/send_bot_message/execute` 携带有效 token → 调用 platform_admin 接口并返回结果
- AC: `POST /tools/query_orders/execute` → 返回最多 5 条订单
- AC: `go build ./...` 零错误

**自测:**
- 测试文件: `mcp-server-platform/main_test.go`
- ST: GET / → 200 + status=ok
- ST: 无 X-Agent-Token → 401
- ST: POST /tools/send_bot_message/execute，mock platform_admin → 200 + 消息写入
- ST: POST /tools/query_orders/execute，mock 返回 10 条 → 只返回 5 条

---

### Task 6: TokenManager（懒初始化 + 后台自动刷新） ⬜

**复杂度**: 高

**设计来源:** design.html §3 Agent 鉴权配置化机制；TokenManager 行为

**Scope（边界）:**
- 涉及文件:
  - `token/manager.go`（新增）
  - `token/manager_test.go`（新增）
  - `callctx/callctx.go`（修改，新增 WithAgentToken / GetAgentToken）
  - `main.go`（修改，初始化 TokenManager，注入到 MCP 工具调用链）
- 涉及模块: token, callctx
- 不触碰: 现有 llm/client.go、agent/agent.go 逻辑

**Constraints（约束）:**
- 懒初始化：第一次 `GetToken()` 调用时触发登录，不在进程启动时执行
- 后台 goroutine 提前 `advance_refresh_secs`（默认 300）秒刷新；刷新失败重试 3 次，全失败则重新登录
- 多 Pod 场景：每 Pod 独立维护 token，refresh_token 冲突时 fallback 重登录即可自愈，无需分布式锁（文档注明）
- 配置来源：环境变量 `AGENT_LOGIN_URL`、`AGENT_USERNAME`、`AGENT_PASSWORD`（P2 阶段升级为 DB 可配）
- `callctx` 注入：MCP 工具调用前调 `callctx.WithAgentToken(ctx, token)` 注入 token，工具内用 `callctx.GetAgentToken(ctx)` 取值，写入 `X-Agent-Token` header

**Acceptance（验收标准）:**
- AC: 首次 `GetToken()` 触发登录，返回有效 token
- AC: 模拟 token 过期前 300s，后台 goroutine 自动刷新
- AC: 刷新失败 3 次后触发重登录
- AC: `callctx.WithAgentToken` / `GetAgentToken` 正确传递 token
- AC: `go build ./...` 零错误

**自测:**
- 测试文件: `token/manager_test.go`
- ST: 初始化后首次 GetToken → 触发 login，返回 token
- ST: 模拟 token 即将过期 → 后台触发 refresh
- ST: mock refresh 失败 3 次 → 触发 re-login
- ST: callctx WithAgentToken → GetAgentToken 返回同值

---

### Task 7: AgentService（AI 触发 + 策略检测 + 降级） ⬜

**复杂度**: 高

**设计来源:** design.html §5 触发流程；Token 注入机制；goroutine 并发限制

**Scope（边界）:**
- 涉及文件:
  - `app/license/service/agent_service.go`（新增）
  - `app/license/service/agent_service_test.go`（新增）
  - `app/license/service/customer_service.go`（修改，新增 HandoffToHuman 调用点）
- 涉及模块: license/service
- 不触碰: customer_service.go 现有方法签名、现有 WS 广播逻辑

**Constraints（约束）:**
- `TriggerIfEnabled(ctx, session, message)` 入口：先读 `agent.enabled`，false 直接返回
- 策略 3（超轮次）：`GetInt(tenantID, "handoff.max_bot_turns", 10)`，超过则先调 `InternalWriteSystemMsg`（取 `max_turns_msg`）再 `HandoffToHuman`，**不触发 AI**
- 策略 4（关键词）：`GetString(tenantID, "handoff.keywords", "...")`，`strings.Contains` 包含匹配，命中则先写 `no_answer_msg` 再 `HandoffToHuman`，**不触发 AI**
- 触发 AI：异步 goroutine，调 `POST {agent.base_url}/api/chat`（非流式），request body 含 `session_id="cs-{id}"`、`context.cs_session_id / user_id / tenant_id`
- **per-tenant 信号量**：带缓冲 channel 限制并发 goroutine 数，默认上限=10（可配 `agent.max_concurrent`）；超限时直接写 `fallback_msg` + `HandoffToHuman`，不启动新 goroutine
- 超时：`GetInt(tenantID, "handoff.timeout_secs", 30)` 秒；超时写 `fallback_msg` + `HandoffToHuman`
- `InternalSendBotMessage`（bot 回复写入）走独立路由，不触发 `TriggerIfEnabled`（天然防循环）

**Acceptance（验证标准）:**
- AC: `enabled=false` 时调用 `TriggerIfEnabled` → 立即返回，无任何 DB/网络操作
- AC: `enabled=true` 时触发 AI，bot 回复写入会话
- AC: 超轮次 → 写 max_turns_msg + handoff，不触发 AI
- AC: 关键词命中 → 写 no_answer_msg + handoff，不触发 AI
- AC: Agent 超时 30s → 写 fallback_msg + handoff
- AC: 并发 >10 → 第 11 个请求直接降级，不阻塞
- AC: bot 调 send_bot_message 后不触发 TriggerIfEnabled（无循环）
- AC: `go build ./...` 零错误

**自测:**
- 测试文件: `app/license/service/agent_service_test.go`
- ST: enabled=false → 无副作用返回
- ST: 超轮次（mock GetInt=2，当前3轮）→ 写通知消息 + HandoffToHuman
- ST: 关键词「投诉」命中 → 写通知消息 + HandoffToHuman
- ST: mock Agent 超时 → 写 fallback_msg + HandoffToHuman
- ST: mock 并发 11 goroutine → 第 11 个走降级路径
- ST: bot 写消息不触发 TriggerIfEnabled（验证无循环）

---

### Task 8: mjarch domain YAML 配置 ⬜

**复杂度**: 低

**设计来源:** design.html §7 MCP Server 工具清单

**Scope（边界）:**
- 涉及文件:
  - `config/domains/mjarch.yaml`（新增）
- 涉及模块: go-rag-agent/config
- 不触碰: 现有 `ecommerce.yaml`

**Acceptance（验证标准）:**
- AC: go-rag-agent 启动后 `GET /api/admin/domains` 返回列表中含 `mjarch`
- AC: mjarch domain 工具列表含 send_bot_message / handoff_to_human / query_member_info / query_orders
- AC: query_orders 工具描述注明默认 `page_size=5`，session_id/user_id 从 `__context` 取值
- AC: `go build ./...` 零错误

---


## P1 — 功能完整任务

### Task 9: go-rag-agent 开放 CORS ⬜

**复杂度**: 低

**设计来源:** design.html §5 测试面板接入方式

**Scope（边界）:**
- 涉及文件:
  - `main.go`（修改，HTTP handler 前加 CORS middleware）
- 不触碰: 现有路由、现有 handler

**Constraints（约束）:**
- 支持环境变量 `CORS_ALLOW_ORIGINS`（逗号分隔，默认 `*`，生产必须配具体 origin）
- 只需处理 `OPTIONS` preflight 和响应头 `Access-Control-Allow-*`

**Acceptance（验证标准）:**
- AC: platform_admin 管理端页面 fetch `{base_url}/api/chat` 无 CORS 报错
- AC: `go build ./...` 零错误

---

### Task 10: CsMute model 及 AutoMigrate ⬜

**复杂度**: 高

**设计来源:** design.html §6 禁言功能；CsMute model

**Scope（边界）:**
- 涉及文件:
  - `app/license/models/models.go`（修改，新增 CsMute struct）
  - `app/license/plugin.go`（修改，AutoMigrate 加入 CsMute）
- 涉及模块: license/models
- 不触碰: 现有 model 定义

**Constraints（约束）:**
- 字段定义严格按设计：`ID int64`、`TenantID int64`、`UserID int64 json:",string"`、`MuteUntil *time.Time`、`Reason varchar(255)`、`OperatorID int64`
- 唯一索引：`gorm:"uniqueIndex:uk_cs_mute_user"` 同时标在 TenantID 和 UserID
- `MuteUntil` 为指针（`*time.Time`），NULL 表示永久禁言
- MySQL 8.x：`MuteUntil` 字段不加 `default` tag

**Acceptance（验证标准）:**
- AC: AutoMigrate 后 `cs_mutes` 表存在
- AC: 唯一索引 `uk_cs_mute_user(tenant_id, user_id)` 存在
- AC: `MuteUntil` 列可为 NULL
- AC: `go build ./...` 零错误

**自测:**
- 测试文件: `app/license/models/models_test.go`
- ST: 插入一条 CsMute，MuteUntil=nil → 可写入，读出 nil
- ST: 同一 tenant_id+user_id 插入两条 → 第二条触发唯一索引冲突

---

### Task 11: 禁言 service 方法 ⬜

**复杂度**: 中

**设计来源:** design.html §6 禁言功能；禁言检查逻辑

**Scope（边界）:**
- 涉及文件:
  - `app/license/service/customer_service.go`（修改，新增 CheckMuted / MuteUser / UnmuteUser / ListMutes）
- 涉及模块: license/service
- 不触碰: 现有 service 方法签名

**Constraints（约束）:**
- `CheckMuted(tenantID, userID int64) (bool, *time.Time, error)`：查询 `mute_until IS NULL OR mute_until > NOW()`
- `MuteUser`：用 GORM `Save`（upsert），同一用户覆盖旧记录；禁言成功后向用户最近一条活跃 `cs_session` 写入系统消息（`sender_type=system`），内容为禁言提示
- `UnmuteUser`：硬删除该用户禁言记录（非软删除）
- `ListMutes`：分页，支持按 `user_id` 过滤，返回 `([]CsMute, int64, error)`
- 所有方法必须带 `tenant_id` 过滤（多租户隔离）

**Acceptance（验证标准）:**
- AC: 永久禁言用户 → `CheckMuted` 返回 `(true, nil, nil)`
- AC: 有时限禁言且未过期 → 返回 `(true, &muteUntil, nil)`
- AC: 禁言已过期 → 返回 `(false, nil, nil)`
- AC: 无禁言记录 → 返回 `(false, nil, nil)`
- AC: `MuteUser` 成功后，该用户活跃会话中写入系统消息
- AC: `UnmuteUser` 后 `CheckMuted` 返回 false
- AC: `go build ./...` 零错误

**自测:**
- 测试文件: `app/license/service/customer_service_test.go`
- ST: CheckMuted 永久禁言 → (true, nil, nil)
- ST: CheckMuted 未过期时限禁言 → (true, &t, nil)
- ST: CheckMuted 过期禁言 → (false, nil, nil)
- ST: CheckMuted 无记录 → (false, nil, nil)
- ST: MuteUser 两次（覆盖）→ 只有一条记录，以新的为准
- ST: MuteUser 后检查系统消息写入

---

### Task 12: 禁言管理端接口 ⬜

**复杂度**: 中

**设计来源:** design.html §6 接口设计

**Scope（边界）:**
- 涉及文件:
  - `app/license/apis/customer_service_api.go`（修改，新增 3 个接口）
  - `app/license/plugin.go`（修改，注册路由）
- 涉及模块: license/apis
- 不触碰: 现有接口、现有路由

**Constraints（约束）:**
- `POST /admin/cs/sessions/:id/mute`：body `{ duration_hours float64, reason string }`，0=永久禁言；从 session 取 user_id，调 `MuteUser`；成功后通过 WS 广播 `cs_mute_changed` 事件
- `POST /admin/cs/sessions/:id/unmute`：调 `UnmuteUser`；成功后 WS 广播 `cs_mute_changed`
- `GET /admin/cs/mutes`：query params `page, page_size, user_id（可选）`，调 `ListMutes`，返回分页结果
- `duration_hours` 换算：`mute_until = now + duration_hours * 3600s`，存 UTC
- 雪花 ID `user_id` 返回时加 `json:",string"` tag（已在 CsMute model 定义）

**Acceptance（验证标准）:**
- AC: `POST mute duration_hours=1` → DB 写入，`mute_until ≈ now+1h`
- AC: `POST mute duration_hours=0` → `mute_until=NULL`（永久）
- AC: `POST unmute` → DB 记录删除
- AC: `GET mutes` → 分页返回，含 user_id（string 类型）
- AC: mute/unmute 成功后 WS 广播 `cs_mute_changed`
- AC: `go build ./...` 零错误

**自测:**
- 测试文件: `app/license/apis/customer_service_api_test.go`
- ST: POST mute duration_hours=24 → DB mute_until ≈ now+24h
- ST: POST mute duration_hours=0 → mute_until=nil
- ST: POST unmute → CheckMuted 返回 false
- ST: GET mutes → 返回列表，user_id 为 string

---

### Task 13: UserSendCsMessage 禁言检查 + AgentService 触发 ⬜

**复杂度**: 中

**设计来源:** design.html §5 触发流程步骤1-5；禁言检查逻辑代码

**Scope（边界）:**
- 涉及文件:
  - `app/license/apis/customer_service_api.go`（修改，UserSendCsMessage 开头插入检查）
- 涉及模块: license/apis
- 不触碰: 现有消息存储逻辑、WS 广播逻辑

**Constraints（约束）:**
- **禁言检查优先**（在存消息之前）：`CheckMuted` → true 时返回 403 + `{"code":403,"msg":"...","mute_until":...}`
- 永久禁言 msg：`"您已被永久禁言，如有疑问请联系平台"`
- 有时限禁言 msg：`"您已被禁言，禁言将于 2006-01-02 15:04 结束"`（本地时区格式）
- 存消息 + WS 广播后，调 `AgentService.TriggerIfEnabled(session, message)`（异步，不阻塞响应）
- **`InternalSendBotMessage` 走独立路由，不经过 `UserSendCsMessage`，天然防止循环**

**Acceptance（验证标准）:**
- AC: 永久禁言用户发消息 → 403 + 永久禁言文案，消息不入库
- AC: 时限禁言用户发消息 → 403 + 含到期时间文案
- AC: 正常用户发消息 → 200，AgentService.TriggerIfEnabled 被调用
- AC: bot 通过 InternalSendBotMessage 写消息 → TriggerIfEnabled 不被调用
- AC: `go build ./...` 零错误
- AC: 【回归】现有消息存储和 WS 广播逻辑不变（RG-3）

**自测:**
- 测试文件: `app/license/apis/customer_service_api_test.go`
- ST: 被禁言用户 POST → 403，DB 无新消息
- ST: 正常用户 POST → 200，TriggerIfEnabled 被调用（mock 验证）
- ST: 【回归】消息写入 cs_messages + WS 广播仍发生（RG-3）

---

### Task 14: 管理端前端 PA-10（AI 客服机器人配置 + 测试面板 + 策略配置） ⬜

**复杂度**: 高

**设计来源:** design.html §5「AI 客服机器人」区块；测试机器人功能；「转人工策略配置」区块

**Scope（边界）:**
- 涉及文件:
  - `frontend/cs-settings.vue`（修改或新增，待确认路径）— 客服设置页
  - `src/api/cs-config.ts`（新增，cs_configs 读写 API）
- 涉及模块: platform_admin 前端（Vue）
- 不触碰: 其他设置页、现有工作台组件

**Constraints（约束）:**
- 区块1「AI 客服机器人」：enabled 开关（默认 off，切换二次确认）、base_url 输入框（enabled=false 时置灰）、测试连通性按钮（调 `GET {base_url}/api/admin/health`）
- 测试面板（可折叠，默认折叠）：
  - 顶部固定说明文案：`🔧 以下为调试工具，直连 Agent 服务，不受上方启用开关控制，不影响生产会话`
  - 域选择下拉（调 `GET {base_url}/api/admin/domains`，取 enabled=true 的域）
  - 多轮对话（session_id=`admin-test-{uuid}`，重置时重新生成）
  - 前端直接 fetch `{base_url}/api/chat`（stream:true），SSE 实时渲染 token/thinking/meta/tool_start/tool_end
  - 「重置」清空对话和 session_id
- 区块2「转人工策略」：7 个配置项表单（`handoff.*`），enabled=false 时整体置灰
- API 调用：`GET /api/v1/cs/configs`（读，返回键值 map）、`POST /api/v1/cs/configs`（批量写）
- int64 字段无需传，使用 string 类型 API key

**Acceptance（验证标准）:**
- AC: 开关关闭 → base_url 置灰，策略配置区块置灰，测试面板仍可展开
- AC: 测试面板说明文案正确显示
- AC: 测试面板多轮对话正常，流式渲染，thinking/meta/工具状态均显示
- AC: 重置后 session_id 重新生成，历史消息清空
- AC: 测试连通性成功 → 显示绿色 ✓；失败 → 显示红色 ✗ + 错误信息
- AC: 策略配置保存后下次进入页面显示新值

---

### Task 15: 管理端前端 PA-11（工作台禁言操作 + 🔇 标识） ⬜

**复杂度**: 中

**设计来源:** design.html §6 管理端操作位置

**Scope（边界）:**
- 涉及文件:
  - `frontend/cs-workbench-sidebar.vue`（修改，待确认路径）— 会话侧边栏
  - `frontend/cs-session-list.vue`（修改，待确认路径）— 会话列表卡片
  - `src/api/cs-mute.ts`（新增，禁言 API）
- 涉及模块: platform_admin 前端（Vue）
- 不触碰: 侧边栏其他区块、现有发消息逻辑

**Constraints（约束）:**
- 时长下拉：10分钟(0.167h) / 1小时 / 24小时 / 7天 / 永久(0h) + 自定义输入（精度 0.1h）
- 确认禁言后调 `POST /admin/cs/sessions/:id/mute`；解除调 `POST /admin/cs/sessions/:id/unmute`
- WS 监听 `cs_mute_changed` 事件：收到后刷新当前会话的禁言状态，🔇 标识在 **5 秒内**消失/出现
- 会话列表卡片和详情顶部：被禁言用户显示 🔇；hover tooltip 显示到期时间（永久显示「永久禁言」）

**Acceptance（验证标准）:**
- AC: 禁言操作成功后，当前页面 🔇 立即显示
- AC: 解除禁言后，WS 事件到达，🔇 在 5 秒内消失
- AC: hover 🔇 显示到期时间
- AC: 永久禁言 hover 显示「永久禁言」

---


## P2 — 配置化任务

### Task 16: TokenManager 鉴权配置化（Admin API + 热更新） ⬜

**复杂度**: 高

**设计来源:** design.html §3 管理页面「外部鉴权配置」区块

**Scope（边界）:**
- 涉及文件:
  - `main.go`（修改，新增 `GET/PUT /api/admin/auth-config` 路由）
  - `config/config.go`（修改，新增 AuthConfig 结构体）
  - `token/manager.go`（修改，支持热更新配置）
  - `frontend/admin.html`（修改，新增「外部鉴权配置」区块）
- 涉及模块: token, config, main
- 不触碰: 现有域配置逻辑、现有 admin API

**Constraints（约束）:**
- 10 个配置项持久化到 JSON 文件（`data/auth_config.json`）或 DB，优先 JSON（轻量）
- password 字段加密存储，管理页面显示 `****`，不明文回传
- TokenManager `UpdateConfig(cfg AuthConfig)` 方法：下次 token 操作时（不重启）用新配置重登录
- `advance_refresh_secs` 默认 300，支持修改

**Acceptance（验证标准）:**
- AC: `GET /api/admin/auth-config` 返回配置（password 返回 `****`）
- AC: `PUT /api/admin/auth-config` 保存后，TokenManager 下次操作使用新配置
- AC: 修改 login_url 为测试地址后，TokenManager 向新地址发起登录
- AC: `go build ./...` 零错误

**自测:**
- 测试文件: `token/manager_test.go`
- ST: UpdateConfig 新 login_url → 下次 GetToken 向新 URL 登录
- ST: password 字段存储加密，读取时不明文

---

### Task 17: Session 数据库持久化（DBSessionStore） ⬜

**复杂度**: 高

**设计来源:** design.html §4 Session 持久化

**Scope（边界）:**
- 涉及文件:
  - `session/db_store.go`（新增）
  - `session/db_store_test.go`（新增）
  - `agent/session_store.go`（修改，新增 DBSessionStore 实现 SessionStore 接口）
  - `main.go`（修改，根据配置选择 SessionStore 实现）
- 涉及模块: session, agent
- 不触碰: 现有内存 SessionStore 实现（保留作默认）

**Constraints（约束）:**
- `agent_sessions` 表：`(session_id VARCHAR(128) PK, messages_json TEXT, is_test BOOL, created_at, updated_at)`
- `sync.Once` 懒初始化连接池，不在启动时连接
- SQLite 默认（`./data/sessions.db`），MySQL 可配（`SESSION_STORE_DSN`）
- 清理策略：
  - `is_test=true`（session_id 前缀 `admin-test-`）：**TTL=1h**，定时清理 goroutine 每小时执行
  - 普通会话：`SESSION_TTL_DAYS > 0` 时清理，默认 0（永久）
- MySQL 8.x：`messages_json` 类型 TEXT，不加 `default` tag

**Acceptance（验证标准）:**
- AC: 重启 go-rag-agent 后，历史会话消息仍在（SQLite 持久化）
- AC: `admin-test-` 前缀 session 超过 1h 后被清理
- AC: `SESSION_TTL_DAYS=7` 时，7 天前的普通会话被清理
- AC: `go build ./...` 零错误

**自测:**
- 测试文件: `session/db_store_test.go`
- ST: Save + Load → 内容一致
- ST: Load 不存在的 session_id → 返回 nil, nil
- ST: admin-test- session 写入后模拟超 1h → 清理 goroutine 删除
- ST: 普通 session TTL=0 → 不被清理

---

### Task 18: Session 存储管理页配置 ⬜

**复杂度**: 中

**设计来源:** design.html §4 管理页面「Session 存储配置」区块

**Scope（边界）:**
- 涉及文件:
  - `main.go`（修改，新增 `GET/PUT /api/admin/session-config` 路由）
  - `frontend/admin.html`（修改，新增「Session 存储配置」区块）
- 涉及模块: main
- 不触碰: DBSessionStore 内部逻辑

**Constraints（约束）:**
- 6 个配置项：`store_type(sqlite|mysql)`, `sqlite_path`, `mysql_dsn`（加密存储）, `pool_max_open`, `pool_max_idle`, `ttl_days`
- 配置变更后重启生效（不支持热切换，管理页面提示「修改后需重启服务」）
- `mysql_dsn` 不明文回传

**Acceptance（验证标准）:**
- AC: `GET /api/admin/session-config` 返回当前配置
- AC: `PUT` 保存配置持久化，重启后读取新配置
- AC: `go build ./...` 零错误

**自测:**
- 测试文件: `main_test.go` 或集成测试
- ST: PUT session-config store_type=sqlite → GET 返回 sqlite

---

## 依赖图

```
Task 1（cs_configs）
    └─→ Task 4（agent_type 读 enabled）
    └─→ Task 7（AgentService 读所有 handoff.*）
    └─→ Task 14（前端写配置）

Task 2（权限迁移）
    └─→ Task 3（JWT 鉴权改造）
            └─→ Task 5（MCP Server 依赖权限点已存在）

Task 5（MCP Server）+ Task 6（TokenManager）
    └─→ Task 8（mjarch YAML 注册 MCP 工具）

Task 7（AgentService）
    └─→ Task 13（UserSendCsMessage 触发）

Task 10（CsMute model）
    └─→ Task 11（禁言 service）
            └─→ Task 12（禁言接口）
                    └─→ Task 15（前端禁言入口）
            └─→ Task 13（禁言检查）

Task 9（CORS）← 独立，无依赖

Task 16（鉴权配置化）← 依赖 Task 6（TokenManager）
Task 17（DB Session）← 独立
Task 18（Session 管理页）← 依赖 Task 17
```

---

## 推荐开工顺序

```
第1步：Task 1 + Task 2 + Task 9 + Task 17（并行，互无依赖）
第2步：Task 3 + Task 10 + Task 6（并行，依赖第1步部分）
第3步：Task 4 + Task 11 + Task 5（并行）
第4步：Task 7 + Task 12（并行，依赖第3步）
第5步：Task 8 + Task 13（并行）
第6步：联调主链路（P0 Task 1-8 验收）——————P0 完成
第7步：Task 14 + Task 15 + Task 18（并行）
第8步：Task 16（依赖 Task 6）
第9步：P1/P2 集成验收————————————————P1/P2 完成
```

---

## 工作量汇总

| 端 | P0（Task1-8） | P1（Task9-15） | P2（Task16-18） | 合计 |
|----|--------------|----------------|-----------------|------|
| platform_admin 后端 | Task1/2/3/4/7（~560行） | Task10/11/12/13（~200行） | — | ~760行 |
| platform_admin 前端 | — | Task14/15（~300行 Vue） | — | ~300行 |
| go-rag-agent 后端 | Task5/6/8（~620行） | Task9（~10行） | Task16/17（~400行） | ~1030行 |
| go-rag-agent 前端 | — | — | Task16/18（~230行 HTML） | ~230行 |
| **合计** | **~1180行** | **~510行** | **~630行** | **~2320行** |
