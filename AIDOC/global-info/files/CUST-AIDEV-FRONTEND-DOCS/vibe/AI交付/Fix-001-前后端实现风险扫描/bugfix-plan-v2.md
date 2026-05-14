# Bugfix Plan V2 — Fix-001 前后端实现风险扫描（Claude 4.7 复扫版）

## 基本信息

| 项目 | 内容 |
|------|------|
| Fix 编号 | Fix-001（V2） |
| 文档目标 | 基于 Claude 4.7 复扫结果，统一问题编号、标记新增/复核/升级，输出可执行修复批次 |
| 主归属 | `CUST-AIDEV-FRONTEND-DOCS/vibe/AI交付/Fix-001-前后端实现风险扫描/` |
| 关联文档（v1） | `bugfix-plan.md` |
| 关联文档（notice 主计划） | `CUST-AIDEV-BACKEND-DOCS/domain/notice-center/vibe/Fix-001-h5-notice-invisible/bugfix-plan.md` |
| 创建时间 | 2026-04-22 |
| 分析边界 | 静态代码审查（非压测/非生产日志） |

---

## 1. 差异总览（v1 vs Claude 4.7 复扫）

| 维度 | v1 | v2（本次） |
|------|----|-----------|
| 后端问题数 | 13 | 17 |
| 前端问题数 | 4 | 15 |
| 新增高危点 | 有限 | 明确新增支付、公告详情越权、认证/路由权限链路问题 |
| 已有问题复核 | 部分 | 全量复核并标识“复核一致” |
| 执行性 | 中 | 高（统一编号 + 批次 + 依赖 +门禁问题） |

---

## 2. 统一问题清单（含状态标签）

状态标签说明：
- `新增`：v1 未覆盖，v2 首次纳入；
- `复核一致`：v1 已有，v2 再次确认；
- `升级`：v1 有提及，但 v2 判断严重度或影响面应上调。

### 2.1 后端问题（B-*）

| ID | 严重度 | 状态 | 问题摘要 | 核心证据路径 | 建议方向 |
|----|--------|------|----------|--------------|----------|
| B-01 | P0 | 复核一致 | 未初始化态全放行（`permitAll`） | `src/spmp-backend/src/main/java/com/spmp/common/security/SecurityConfig.java` | 未初始化仅放行 init/health，其余继续鉴权或统一 503 |
| B-02 | P0 | 升级 | JWT 默认密钥与弱配置风险（含 local 兜底） | `.../common/security/JwtTokenProvider.java`, `src/spmp-backend/src/main/resources/application.yml` | 强制环境注入 + 启动失败校验 |
| B-03 | P0 | 升级 | `local` profile 默认启用且包含本地账号/敏感参数 | `src/spmp-backend/src/main/resources/application.yml`, `application-local.yml` | 默认 profile 改非 local；local 配置去敏感化 |
| B-04 | P1 | 复核一致 | Swagger/Actuator 暴露风险 | `.../common/security/SecurityConfig.java` | 生产关闭或网关保护 |
| B-05 | P1 | 复核一致 | CORS 双配置且通配策略过宽 | `.../common/security/SecurityConfig.java`, `.../common/config/WebMvcConfig.java` | 单点配置 + 白名单域名 |
| B-06 | P0 | 升级 | 文件上传/下载路径穿越风险 | `.../common/controller/FileController.java`, `.../common/service/impl/FileServiceImpl.java` | 路径归一+前缀校验+分类白名单 |
| B-07 | P1 | 复核一致 | 文件下载 Content-Type 固定 JPEG | `.../common/controller/FileController.java` | 按实际类型返回 |
| B-08 | P0 | 升级 | 数据权限 SELF SQL 字符串拼接 | `.../common/security/DataPermissionInterceptor.java` | 参数化注入或 AST 改写 |
| B-09 | P1 | 复核一致 | `addWhereCondition` 复杂 SQL 语义误改风险 | `.../common/security/DataPermissionInterceptor.java` | 解析器注入或限定适用范围 |
| B-10 | P1 | 复核一致 | 公告快照/重推问题（已知 notice Fix-001） | `.../mapper/notice/AnnouncementTargetSnapshotMapper.xml`, `.../notice/service/impl/NoticePublishServiceImpl.java` | 以 notice-center 既有 Fix-001 为主计划 |
| B-11 | P1 | 升级 | 定时任务缺分布式锁/幂等 | `.../billing/config/*Task.java`, `.../workorder/config/*Task.java` | 加锁与幂等键 |
| B-12 | P2 | 复核一致 | 大量 broad catch 降低可观测性 | 多处 `catch (Exception)` | 异常分层处理与指标化 |
| B-13 | P2 | 复核一致 | `WebSecurityConfigurerAdapter` 技术债 | `.../common/security/SecurityConfig.java` | 迁移 `SecurityFilterChain` |
| B-14 | P0 | 新增 | H5 支付回调缺鉴权/签名/金额校验（资金风险） | `.../billing/controller/h5/H5BillController.java`, `.../billing/service/impl/PaymentServiceImpl.java` | 回调内网化/签名校验/金额对账 |
| B-15 | P0 | 新增 | H5 公告详情 IDOR（可读非目标公告） | `.../notice/service/impl/NoticeQueryServiceImpl.java` | 按用户目标范围过滤详情 |
| B-16 | P1 | 新增 | 可能缺 `@EnableScheduling` 导致任务不执行 | `src/spmp-backend/src/main/java/com/spmp/SpmpApplication.java` | 显式启用并加启动自检 |
| B-17 | P1 | 新增 | 支付并发/幂等不足（回调与创建竞争） | `.../billing/service/impl/PaymentServiceImpl.java` | 乐观锁/状态条件更新/幂等键 |
| B-18 | P1 | 新增 | H5 支付详情归属校验缺失（IDOR） | `.../billing/service/impl/H5BillServiceImpl.java`, `PaymentServiceImpl.java` | 加 ownerId 归属校验 |
| B-19 | P1 | 新增 | 退款金额上限/累计校验不足 | `.../billing/service/impl/BillServiceImpl.java` | 强校验退款额度与状态机 |
| B-20 | P1 | 新增 | JWT 黑名单 Redis 失败降级放行 | `.../common/security/JwtAuthenticationFilter.java` | 敏感接口 fail-close 策略 |

### 2.2 前端问题（F-*）

| ID | 严重度 | 状态 | 问题摘要 | 核心证据路径 | 建议方向 |
|----|--------|------|----------|--------------|----------|
| F-01 | P1 | 复核一致 | 401 清理不一致，refresh 残留风险 | `src/spmp-web-pc/src/utils/request.ts`, `src/spmp-web-pc/src/store/modules/user.ts` | 401 统一 `resetState` + 守卫状态重置 |
| F-02 | P2 | 复核一致 | H5 登录双写 localStorage 与 store | `src/spmp-web-h5/src/views/login/LoginView.vue` | 收敛到 store action |
| F-03 | P1 | 升级 | Token 存 localStorage（XSS 面） | `src/spmp-web-pc/src/utils/request.ts`, `src/spmp-web-h5/src/utils/request.ts` | 评估 Cookie 会话或缩短凭据寿命 |
| F-04 | P2 | 复核一致 | 错误处理泛化与静默吞错 | 多处 `catch {}` | 区分 silent 与可视错误 |
| F-05 | P0 | 新增 | H5 前端直接触发支付 callback（伪造支付成功链路） | `src/spmp-web-h5/src/views/billing/BillDetailView.vue`, `src/spmp-web-h5/src/api/billing.ts` | 前端禁调 callback；由后端/网关回调闭环 |
| F-06 | P0 | 新增 | PC 侧边栏未按权限过滤（RBAC 展示越权） | `src/spmp-web-pc/src/layout/AppSidebar.vue`, `src/spmp-web-pc/src/router/index.ts` | 菜单改为后端权限树驱动 |
| F-07 | P1 | 新增 | 短信/refresh 敏感参数经 query 传输 | `src/spmp-web-pc/src/api/auth.ts`, `src/spmp-web-h5/src/api/auth.ts` | 改 POST body 参数 |
| F-08 | P1 | 新增 | 初始化守卫 fail-open + 缓存不失效 | `src/spmp-web-pc/src/router/guard.ts` | fail-close 与状态失效机制 |
| F-09 | P1 | 新增 | H5 守卫缺少认证状态门禁（仅 token） | `src/spmp-web-h5/src/router/guard.ts` | 增加 `requiresCertified` 等策略 |
| F-10 | P1 | 新增 | 响应解包风格混乱导致 `res.data` 误读 | `src/spmp-web-pc/src/utils/request.ts`, `src/spmp-web-pc/src/views/notice/detail/index.vue` | 请求层泛型化，统一分页/普通响应 |
| F-11 | P1 | 新增 | 上传分类参数前端可任意传递 | `src/spmp-web-pc/src/api/common/upload.ts`, `src/spmp-web-h5/src/api/common/upload.ts` | category 枚举化 |
| F-12 | P1 | 新增 | 首页关键数据硬编码（误导业务状态） | `src/spmp-web-h5/src/views/home/HomeView.vue` | 切真实接口并区分空态 |
| F-13 | P2 | 新增 | base 缓存无自动失效 | `src/spmp-web-pc/src/store/modules/base.ts` | CRUD 后自动失效策略 |
| F-14 | P2 | 新增 | PC/H5 请求层重复代码导致漂移 | `src/spmp-web-pc/src/utils/request.ts`, `src/spmp-web-h5/src/utils/request.ts` | 抽 shared 请求层 |
| F-15 | P2 | 新增 | 环境/代理配置硬编码弹性不足 | `src/spmp-web-pc/vite.config.ts`, `src/spmp-web-h5/vite.config.ts` | 环境变量分层治理 |

---

## 2.3 业务逻辑错误清单（第0批次优先处理）

以下问题以“业务规则实现错误 / 业务流程不闭环 / 业务状态不一致”为判定标准，要求纳入第0批次按 Bugfix 流程推进（`bugfix.md -> fix-design.md -> fix-tasks.md`）：

- 后端：`B-10`（公告快照与重推）、`B-17`（支付并发幂等）、`B-18`（支付详情归属校验）、`B-19`（退款额度校验）
- 前端：`F-01`（401 状态一致性）、`F-05`（支付链路前端误触发回调）、`F-08`（初始化守卫 fail-open）、`F-09`（认证态门禁缺失）、`F-10`（响应解包误读）、`F-12`（首页关键数据硬编码）

> 说明：`B-10` 继续沿用 `notice-center` 既有 Fix-001 主计划，本目录只负责批次优先级与联动约束。

---

## 3. 修复批次与里程碑

### 第0批次（业务逻辑错误，最优先处理）
- 后端：B-10、B-17、B-18、B-19
- 前端：F-01、F-05、F-08、F-09、F-10、F-12
- 里程碑：先恢复关键业务闭环与状态一致性，确保核心流程“可正确使用”

### 第一批（阻塞上线，安全高危）
- 后端（安全高危）：B-01、B-02、B-03、B-06、B-08、B-14、B-15
- 前端（安全高危）：F-03、F-06
- 里程碑：关闭高危安全窗口，形成可上线安全基线

### 第二批（功能正确性与一致性）
- 后端：B-07、B-09、B-11、B-16、B-20
- 前端：F-07、F-11
- 里程碑：闭环关键业务链路，减少线上功能性回归

### 第三批（工程化与技术债）
- 后端：B-04、B-05、B-12、B-13
- 前端：F-02、F-04、F-13、F-14、F-15
- 里程碑：提升长期可维护性与演进稳定性

---

## 4. 依赖与联动关系

- `F-05` 与 `B-14` 必须同批落地：仅改前端或仅改后端都会留下资金漏洞窗口。
- `F-06` 依赖后端权限树接口契约稳定（`menus/permissions`）；与 `F-01` 一起做登录态一致性整改。
- `B-10` 作为 notice 专项问题，主修复文档保持在：
  - `CUST-AIDEV-BACKEND-DOCS/domain/notice-center/vibe/Fix-001-h5-notice-invisible/bugfix-plan.md`
  - 本文仅做引用，不重复拆任务。
- `F-11` 与 `B-06` 建议并行：前后端同时白名单，避免单侧绕过。

---

## 4.1 公告可见性专项（B-10 口径更新）

### 目标口径（已确认）

- 任何 **H5 已登录用户** 都可见公告（不依赖认证状态、不依赖房产绑定）。

### 可选实现方案（与 notice 主文档保持一致）

#### 方案 A（推荐）：快照口径放宽 + 重推前补快照

- 快照生成改为面向 H5 登录用户集合（不再限定 `CERTIFIED` / 房产绑定）。
- `repush` 前先增量补快照，再执行未读重推。
- 优点：沿用既有快照分发架构，改动集中、上线风险可控。
- 风险：快照量上升，需要关注批处理性能。

#### 方案 B（可选）：查询动态可见 + 弱化快照依赖

- H5 查询链路按“发布态 + 登录态”动态判定可见，不再强依赖快照。
- 快照主要用于推送与统计，不再作为可见性主约束。
- 优点：新登录用户即时可见。
- 风险：改动面更大，统计/推送一致性回归成本更高。

### 主文档引用

- 以 `CUST-AIDEV-BACKEND-DOCS/domain/notice-center/vibe/Fix-001-h5-notice-invisible/bugfix-plan.md` 为主。
- 本文仅同步口径与批次影响，不重复展开实现细节。

### 回归范围

- H5 公告列表与详情可见性
- 已读/未读统计口径
- 重推覆盖与幂等
- PC 审批/发布/撤回行为不回归

### 确认项（请二选一）

- [x] 采用方案 A（快照口径放宽 + 重推前补快照）
- [ ] 采用方案 B（查询动态可见 + 弱化快照依赖）

---

## 5. 进入 bugfix.md 前的门禁问题（Question/Answer）

**[Question] Q1**：支付链路是否接受“短期禁用前端 callback + 后端仅内网回调 + 管理端手工核销兜底”的过渡方案，还是必须一步到位接真实第三方支付签名回调？  

**[Question] Q2**：认证凭据策略是否批准为“两步走”：
1) 当前迭代先做最小安全修复（统一 401 全清、减少 localStorage 落地）；
2) 下个 CR 再迁移到 HttpOnly Cookie 会话。  

**[Question] Q3**：未初始化态安全策略是否统一为 fail-close（除 `/init` 与健康检查外一律不放行业务请求）？  

**[Question] Q4**：notice 问题（B-10）是否确认完全并入 `notice-center` 既有 Fix-001，不在本目录重复产出 `bugfix.md`？  

**[Question] Q5**：第三批技术债（B-13、F-14、F-15）本迭代是否仅登记，不进入本次开发窗口？  

---

## 6. 状态

| 阶段 | 状态 |
|------|------|
| Bugfix-Plan V2 | 已完成（门禁问题已澄清） |
| bugfix.md / fix-design.md / fix-tasks.md | 已产出（第0批次） |
| 代码修复 | 未开始 |

