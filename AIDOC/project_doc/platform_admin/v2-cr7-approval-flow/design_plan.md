# 设计计划：CR-7 审批流引擎

## 设计方向

基于已确认的需求，拟采用以下技术方案：

1. **数据模型**：新增 3 张表（admin_approval_flow / admin_approval / admin_approval_node），审批节点的 flow_config 以 JSON 存储节点列表，保持灵活性
2. **状态机**：审批实例和节点各自维护独立状态，节点状态驱动实例状态流转
3. **定时任务**：用 Go 的 `jobs` 包（项目已有 `app/jobs/`）实现超时扫描，不引入新依赖
4. **EventBus**：在 `common/event/bus.go` 新建进程内业务事件总线，接口与插件 EventBus 分离
5. **模块位置**：新建 `app/admin/` 下的 approval 相关文件，复用现有路由注册模式（`init()` + `routerCheckRole`）

---

## 技术选型

### 超时扫描方案

| 选项 | 说明 | 优缺点 |
|------|------|--------|
| **A. Go cron job（项目内 jobs 包）** | 定时扫描 admin_approval_node 中超时节点 | 无额外依赖，项目已有基础，推荐 |
| B. Redis 延迟队列 | 精确到秒，支持分布式 | 需引入 Redis 依赖，本项目未使用 |
| C. DB 轮询 + 行锁 | 扫描超时节点，加行锁防并发 | 最简单，适合低频超时场景 |

**选 A**：项目 `app/jobs/` 已有 cron 框架，扫描间隔 5 分钟即可满足需求，不引入新依赖。

### EventBus 实现方案

| 选项 | 说明 |
|------|------|
| **进程内 sync.Map + handler slice** | 轻量，无依赖，足够审批场景 |
| 第三方 eventbus 库 | 功能更丰富但引入依赖 |

**选进程内实现**：`common/event/bus.go`，接口简单（Publish / Subscribe），后续可无缝替换为外部 MQ。

---

## 澄清问题

- [Question-1] **cron 扫描间隔**：超时节点扫描间隔设 5 分钟是否可接受？还是需要更精确（如 1 分钟）？
  [Answer-1]
可以
- [Question-2] **审批流定义是否需要版本管理**？即：修改了流程定义后，进行中的实例是否继续走旧版本流程还是立即切换新版本？
  [Answer-2]
继续旧版本直到完成
- [Question-3] **admin_tenant_app 表是否已存在**？subscription_mode 字段需要 ALTER TABLE 还是在新建时包含？
  [Answer-3]
参考现有实现确认是否按增量处理
- [Question-4] **前端审批管理页面的路由**：是作为独立顶级菜单（如"审批管理"）还是作为某个现有模块的子菜单（如"系统管理 → 审批流"）？
  [Answer-4]
顶级菜单
---

## 风险点

- [Risk-1] **cron 扫描并发**：多 pod 下同一超时节点可能被多个 pod 同时扫到，需要用 SELECT FOR UPDATE 行锁防止重复执行
- [Risk-2] **流程定义修改影响进行中实例**：Question-2 未回答前，默认进行中实例快照 flow_config（发起时复制到实例），不受后续定义修改影响
- [Risk-3] **超时 ESCALATE 的 escalate_to 用户已离职/角色被删除**：需在执行前验证 escalate_to 是否仍有效，无效时降级 AUTO_APPROVE 并记录日志
- [Risk-4] **审批与 EventBus 事务边界**：DB 事务提交后才能发布事件，否则事务回滚但事件已发出（订阅方已执行）；需确保 Publish 在事务 Commit 之后调用
