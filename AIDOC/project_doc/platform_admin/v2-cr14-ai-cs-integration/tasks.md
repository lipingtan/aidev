# CR14 任务清单

> 设计文档：[design.html](./design.html)  
> 最后更新：2026-10-31  
> 前置已完成：go-rag-agent 流式输出改造（ST-1~ST-4，2026-10-31 验证通过）

---

## 阶段说明

| 阶段 | 目标 | 交付标准 |
|------|------|---------|
| **P0 主链路** | platform_admin → Agent → MCP → platform_admin 完整调用链跑通 | 用户发消息，机器人自动回复，可转人工 |
| **P1 功能完整** | 禁言、Agent 开关、策略配置前端 | 管理端可配置并操作所有 CR14 功能 |
| **P2 配置化** | TokenManager 可配、Session 持久化可配 | Agent 凭证和会话存储均可在管理页调整 |

---

## 新增需求（2026-10-31）

### Agent 控制菜单

platform_admin 客服设置页需要独立的「AI 客服机器人」配置区块：

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `agent.enabled` | bool | **false** | 总开关，关闭时 platform_admin 完全不感知机器人存在 |
| `agent.base_url` | string | `http://127.0.0.1:8081` | go-rag-agent 服务地址 |

**关闭时行为（`agent.enabled = false`）：**
- 新建会话：`mode = "human"`，不绑定 bot CsAgent，直接进人工队列
- `UserSendCsMessage`：跳过 AgentService 触发，消息只存库 + WS 广播
- 会话列表：不显示「🤖 bot 处理中」状态，不显示 AI 相关入口
- 策略配置区块（max_bot_turns / keywords 等）在关闭时置灰不可编辑

**开启时行为（`agent.enabled = true`）：**
- 需要填写有效的 `agent.base_url`，保存时可选「测试连通性」
- 新建会话自动绑定 bot CsAgent，触发 AgentService

> 影响任务：PA-1、PA-7、PA-8、PA-10（扩展）

---

## P0 — 主链路任务

### 后端

#### PA-5（手工操作，无代码）
**内容**：在 platform_admin 管理端创建 `agent_bot` 账号，分配 `ai_agent` 角色，赋予以下 4 个权限点：
- `cs:bot:send_message`
- `cs:bot:handoff`
- `cs:query:member_info`
- `cs:query:orders`

**负责人**：—  
**依赖**：无  
**验收**：用 agent_bot 登录可获取 JWT，JWT 调用以上 4 个接口均返回 200

---

#### PA-6
**内容**：`InternalSendBotMessage` / `InternalHandoffToHuman` 移除 `checkInternalToken`，改为标准 JWT Bearer 鉴权，校验 `ai_agent` 角色对应权限点

**文件**：`app/license/apis/customer_service_api.go`  
**规模**：~20 行  
**依赖**：PA-5  
**验收**：用 agent_bot JWT 调用两个接口 200；无 token 或错误 token 返回 401

---

#### PA-1
**内容**：
1. `CsAgent` model 加 `AgentType varchar(20) not null default 'human'`，`AutoMigrate`
2. 新建会话时：若 `agent.enabled = true`，自动将 `session.agent_id` 设为 bot CsAgent ID，`session.mode = "bot"`；否则 `mode = "human"`，`agent_id = NULL`

**文件**：`app/license/models/models.go`，`app/license/service/customer_service.go`  
**规模**：~30 行  
**依赖**：PA-5（需要 bot CsAgent ID），Agent 控制菜单配置读取  
**验收**：开启 Agent 时新会话 mode=bot；关闭时 mode=human

---

#### PA-9
**内容**：独立进程 MCP Server（`:9001`），实现 4 个工具，每个工具调用 platform_admin 管理端 API 时携带 `Authorization: Bearer <token>`（token 由调用方注入，即 go-rag-agent TokenManager 提供）：

| 工具名 | 目标接口 | 入参 |
|--------|---------|------|
| `send_bot_message` | POST `/admin/cs/sessions/:id/messages` | session_id, content |
| `handoff_to_human` | POST `/admin/cs/sessions/:id/handoff` | session_id, reason |
| `query_member_info` | GET `/admin/cs/users/:id/member-info` | user_id |
| `query_orders` | GET `/admin/license/orders?user_id=xxx` | user_id, page, page_size |

**文件**：新建 `mcp-server-platform/main.go`（复用 `mcp-server/` 框架）  
**环境变量**：`PLATFORM_ADMIN_URL=http://127.0.0.1:8080`，`MCP_SERVER_PORT=9001`  
**规模**：~300 行  
**依赖**：PA-6  
**验收**：4 个工具均可通过 `POST /tools/:name/execute` 正常调用，工具结果符合预期

---

#### AG-1
**内容**：`TokenManager`，负责 go-rag-agent 登录 platform_admin 并维护 access_token：
- 懒初始化：第一次 MCP 工具调用时触发登录
- 后台 goroutine：提前 `advance_refresh_secs`（默认 300）秒自动刷新
- 刷新失败重试 3 次，全失败则重新登录
- callctx 注入：每次 MCP 工具调用前将 token 写入请求 header `Authorization: Bearer <token>`
- 配置来源：先读环境变量（`AGENT_LOGIN_URL` / `AGENT_USERNAME` / `AGENT_PASSWORD`），后读 AG-2 管理页面配置

**文件**：新建 `token/manager.go`  
**规模**：~200 行  
**依赖**：无（环境变量先行，AG-2 后续扩展）  
**验收**：首次工具调用自动登录；token 过期前自动刷新；手动令 token 过期后验证自动重新登录

---

#### PA-8
**内容**：`AgentService`：
- `agent.enabled = false` 时直接返回，完全跳过
- 策略 3（超轮次）：读 `cs_configs.handoff.max_bot_turns`，超过则调 `HandoffToHuman`，不触发 AI
- 策略 4（关键词）：读 `cs_configs.handoff.keywords`，命中则调 `HandoffToHuman`，不触发 AI
- 触发 AI：异步 goroutine 调 `POST {agent.base_url}/api/chat`（非流式），超时 `handoff.timeout_secs`（默认 30）
- 失败降级：写兜底消息（`handoff.fallback_msg`）+ 调 `HandoffToHuman`
- cs_configs 新增策略相关配置项（`handoff.*` 前缀）

**文件**：新建 `app/license/service/agent_service.go`  
**规模**：~150 行  
**依赖**：Agent 控制菜单配置读取  
**验收**：
- `enabled=false` 时用户发消息不触发 AI
- `enabled=true` 时正常触发，bot 回复写入会话
- 超时 30s 后写兜底消息并转人工
- 超轮次 / 关键词命中直接转人工不触发 AI

---

#### AG-5
**内容**：mjarch domain YAML，配置：
- `system_prompt_extra`：客服场景说明 + 平台背景
- `tools`：注册 PA-9 的 4 个 MCP 工具
- `knowledge_base`：占位（空列表，主链路跑通后补充知识库文档）

**文件**：新建 `config/domains/mjarch.yaml`  
**规模**：~100 行  
**依赖**：PA-9  
**验收**：go-rag-agent 加载 mjarch domain 无报错，工具列表可见，知识库为空时正常降级工具调用

---

## P1 — 功能完整任务

### 后端

#### PA-2
**内容**：新建 `CsMute` model，`AutoMigrate`

```go
type CsMute struct {
    ID         int64      `gorm:"primaryKey;autoIncrement"`
    TenantID   int64      `gorm:"not null;uniqueIndex:uk_cs_mute_user"`
    UserID     int64      `gorm:"not null;uniqueIndex:uk_cs_mute_user" json:"user_id,string"`
    MuteUntil  *time.Time `gorm:"index"`           // NULL = 永久禁言
    Reason     string     `gorm:"type:varchar(255)"`
    OperatorID int64
    CreatedAt  time.Time  `gorm:"autoCreateTime"`
    UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
}
```

**文件**：`app/license/models/models.go`  
**规模**：~30 行  
**依赖**：无  
**验收**：`AutoMigrate` 后 `cs_mutes` 表存在，唯一索引正确

---

#### PA-3
**内容**：service 方法：
- `CheckMuted(tenantID, userID) (bool, *time.Time, error)`：查询有效禁言，`mute_until IS NULL OR mute_until > now()`
- `MuteUser(tenantID, userID, durationHours float64, reason string, operatorID int64) error`：upsert（同一用户覆盖旧记录）
- `UnmuteUser(tenantID, userID) error`：删除禁言记录
- `ListMutes(tenantID int64, page, size int) ([]CsMute, int64, error)`

**文件**：`app/license/service/customer_service.go`  
**规模**：~80 行  
**依赖**：PA-2  
**验收**：单测覆盖 CheckMuted（有效/过期/永久/无记录四种情况）

---

#### PA-4
**内容**：3 个管理端接口 + 路由注册：
- `POST /admin/cs/sessions/:id/mute`，body: `{ duration_hours: float, reason: string }`（0=永久）
- `POST /admin/cs/sessions/:id/unmute`
- `GET /admin/cs/mutes`，query: `page, size, user_id（可选）`

**文件**：`app/license/apis/customer_service_api.go`  
**规模**：~60 行  
**依赖**：PA-3  
**验收**：三个接口 200；禁言后 CheckMuted 返回 true；解除后返回 false

---

#### PA-7
**内容**：`UserSendCsMessage` 开头插入两个检查：
1. `CheckMuted` → true 时返回 403 + `mute_until`
2. 消息存库 + WS 广播后，调用 `AgentService.TriggerIfEnabled(session, message)`

**文件**：`app/license/apis/customer_service_api.go`  
**规模**：~30 行  
**依赖**：PA-3，PA-8  
**验收**：被禁言用户发消息返回 403；正常用户发消息触发 AgentService

---

### 前端

#### PA-10（Vue）
**内容**：客服设置页新增两个区块：

**区块1「AI 客服机器人」**（新增需求，顶部显示）：
- 开关：启用 / 停用（默认停用），切换时有确认弹窗
- Base URL 输入框（enabled=false 时置灰）
- 「测试连通性」按钮（调 `GET {base_url}/api/admin/health`，显示 ok/失败）
- 保存按钮

**区块2「转人工策略」**（原有设计，enabled=false 时整体置灰）：
- max_bot_turns（数字输入，0=不限）
- keywords（标签输入，逗号分隔）
- timeout_secs（数字输入）
- fallback_msg（文本框）
- no_answer_msg（文本框）
- no_human_msg（文本框）

**规模**：~120 行 Vue  
**依赖**：PA-8（后端 cs_configs API）  
**验收**：开关关闭时所有策略项置灰；测试连通性功能正常

---

#### PA-11（Vue）
**内容**：工作台会话侧边栏禁言操作入口：
- 时长选择：10分钟 / 1小时 / 24小时 / 7天 / 永久（下拉）
- 原因输入（可选）
- 确认禁言 / 解除禁言按钮
- 会话列表卡片和会话详情顶部：被禁言用户显示 🔇 标识 + hover 显示到期时间

**规模**：~100 行 Vue  
**依赖**：PA-4  
**验收**：禁言操作成功后标识即时显示；解除后标识消失

---

## P2 — 配置化任务

### 后端

#### AG-2
**内容**：管理页面「外部鉴权配置」Admin API，支持读写以下配置（存 `auth_config` 表或 JSON 文件）：
`login_url / login_body / login_token_path / login_refresh_path / login_expire_path / refresh_url / refresh_body / username / password（加密存储）/ advance_refresh_secs`

TokenManager 监听配置变更，下次 token 操作时热更新生效（无需重启）

**文件**：`main.go`（新增 admin 路由），新建 `config/auth_config.go`  
**规模**：~150 行  
**依赖**：AG-1  
**验收**：管理页保存新 login_url 后，TokenManager 用新配置重新登录

---

#### AG-3
**内容**：`DBSessionStore` 替换现有内存 session 存储：
- `sync.Once` 懒初始化连接池
- `agent_sessions` 表：`session_id, messages_json, created_at, updated_at`
- SQLite 默认（`./data/sessions.db`），MySQL 可配
- `ttl_days > 0` 时启动定时清理 goroutine（默认 0 = 永久保存）

**文件**：新建 `session/db_store.go`  
**规模**：~200 行  
**依赖**：无  
**验收**：重启 go-rag-agent 后历史会话仍在；MySQL DSN 可切换

---

#### AG-4
**内容**：管理页面「Session 存储配置」区块 + Admin API（读写 `store_type / sqlite_path / mysql_dsn / pool_max_open / pool_max_idle / ttl_days`）

**文件**：`main.go`，`frontend/admin.html`  
**规模**：~80 行后端 + ~80 行前端  
**依赖**：AG-3  
**验收**：管理页面切换 SQLite ↔ MySQL 后重启生效

---

## 依赖图

```
PA-5（手工）
    └─→ PA-6
            └─→ PA-9
                    └─→ AG-5（mjarch domain YAML）
    
AG-1（TokenManager）
    └─→ PA-9 使用（token 注入）
    └─→ AG-2（P2，配置化）

PA-1（agent_type 字段）← Agent 控制菜单开关

PA-2（CsMute model）
    └─→ PA-3（service）
            └─→ PA-4（API）
                    └─→ PA-11（前端禁言入口）
            └─→ PA-7（UserSendCsMessage 禁言检查）

PA-8（AgentService）← Agent 控制菜单开关
    └─→ PA-7（UserSendCsMessage 触发）
    └─→ PA-10（前端策略配置）

AG-3（DB Session）
    └─→ AG-4（P2，管理页配置）
```

---

## 推荐开工顺序

```
第1步：PA-5（手工建账号）─────────────────────────────── 前置
第2步：PA-2 + PA-6 + AG-1（并行，互无依赖）─────────── 基础
第3步：PA-3 + PA-9（并行）──────────────────────────── 核心组件
第4步：PA-1 + PA-4 + PA-8（并行，依赖第2/3步）──────── 集成逻辑
第5步：PA-7 + AG-5（并行）──────────────────────────── 串联
第6步：联调主链路（P0 验收）────────────────────────── P0 完成
第7步：PA-10 + PA-11（前端，并行）──────────────────── P1 前端
第8步：PA-10 Agent 控制菜单区块─────────────────────── P1 完成
第9步：AG-2 + AG-3（并行）──────────────────────────── P2
第10步：AG-4（依赖 AG-3）───────────────────────────── P2 完成
```

---

## 工作量汇总

| 端 | P0 | P1 | P2 | 合计 |
|----|----|----|----|----|
| platform_admin 后端 | PA-1/5/6/7/8/9（~560行） | PA-2/3/4/7（~200行） | — | ~760行 |
| platform_admin 前端 | — | PA-10/11（~220行 Vue） | — | ~220行 |
| go-rag-agent 后端 | AG-1/5（~300行） | — | AG-2/3/4（~430行） | ~730行 |
| go-rag-agent 前端 | — | — | AG-2/4（~230行 HTML） | ~230行 |
| **合计** | **~860行** | **~420行** | **~660行** | **~1940行** |
