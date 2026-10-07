# 需求计划：客服会话路由策略

**状态：** 待规划（暂不实现）
**优先级：** 中
**背景：** 当前客服工作台为人工认领模式，随用户量增长需要自动路由提升效率

---

## 需求理解

当前 `cs_sessions.agent_id` 创建时为 NULL，客服在工作台手动点开未分配会话认领。需要逐步引入自动路由策略，减少人工操作，提升响应效率。

---

## TODO 列表

### Phase 1：数据模型准备

- [ ] `cs_agents` 表加 `max_concurrent` 字段（最大并发会话数，默认 10）
- [ ] `cs_agents` 表加 `skill_tags` 字段（JSON 数组，如 `["official_web","game_support"]`）
- [ ] `cs_sessions` 表加 `source_code` 索引（已有字段，加索引便于按来源路由）
- [ ] 新建 `cs_routing_rule` 表：`source_code → skill_tag` 映射关系

```sql
CREATE TABLE cs_routing_rule (
  id           BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id    BIGINT NOT NULL,
  source_code  VARCHAR(64) NOT NULL,   -- 触发规则的来源，如 official_web
  skill_tag    VARCHAR(64) NOT NULL,   -- 要求客服具备的技能标签
  priority     INT NOT NULL DEFAULT 0, -- 规则优先级
  is_enabled   TINYINT(1) NOT NULL DEFAULT 1,
  created_at   DATETIME(3),
  updated_at   DATETIME(3)
);
```

---

### Phase 2：路由引擎（后端）

- [ ] `AutoAssignAgent(tenantID, sessionID int64) (agentID int64, err error)`
  - 逻辑：
    1. 查会话 `source_code`，匹配 `cs_routing_rule` 得到 `skill_tag`
    2. 查在线（`is_online=true`）且启用（`status=1`）且有对应 `skill_tag` 的客服
    3. 按「当前活跃会话数 / max_concurrent」比例最低者优先（最少会话策略）
    4. 无匹配技能时 fallback 到全局最少会话客服
    5. 全部客服超载时返回 0（进队列等待）
  - UPDATE `cs_sessions SET agent_id=? WHERE id=?`

- [ ] `StickyRouting(tenantID, userID int64) (agentID int64)`
  - 查该用户最近一次会话的 `agent_id`，该客服在线则返回（粘性路由）
  - 结合 AutoAssignAgent 使用：粘性优先，粘性失效时走自动分配

- [ ] `GetOrCreateSession` 在创建新会话时调用路由引擎，自动填充 `agent_id`

---

### Phase 3：管理端路由规则配置

- [ ] 新增「路由规则」Tab（在 CsAgents 页面内或独立菜单）
- [ ] 功能：配置 source_code → skill_tag 映射
  - 列表展示所有规则
  - 新增/编辑/删除规则
  - 开关启用/禁用
- [ ] 客服人员配置加 `skill_tags` 多选（技能标签 + 最大并发数）
- [ ] 实时查看各客服当前负载（活跃会话数 / max_concurrent）

---

### Phase 4：队列与排队

- [ ] 会话分配失败（所有客服超载）时进排队状态
- [ ] `cs_sessions.queue_position` 字段记录排队位置
- [ ] 客服完成会话后触发队列中下一个会话的自动分配
- [ ] C 端显示「当前排队 X 人，预计等待 Y 分钟」

---

### Phase 5：优先级路由

- [ ] VIP/付费会员用户优先队列
- [ ] 联动 `user_membership` 表：有有效会员的用户插队
- [ ] SLA 配置（普通用户/VIP 分别配置最大等待时间）

---

## 技术约束

- 路由决策必须在事务内完成（防止多个请求同时分配给同一客服导致超载）
- 使用数据库行锁（`SELECT FOR UPDATE`）或 Redis 分布式锁
- 粘性路由的客服离线时必须降级到自动分配，不能让用户无限等待
- 所有路由决策记录日志（便于统计和调优）

---

## 待确认

1. 是否需要「人工干预」功能：客服主管可以手动重新分配会话
2. 客服超载时是否提示用户排队（C 端需要改造）
3. 技能标签是固定枚举还是可自定义
4. 跨时区/跨班次时路由策略（离线客服的处理）
