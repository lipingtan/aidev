# Bugfix Plan — Fix-001 前后端实现风险扫描（全栈）

## 基本信息

| 项目 | 内容 |
|------|------|
| Fix 编号 | Fix-001 |
| 简述 | 对已实现的 PC/H5 前端与 `spmp-backend` 后端模块做静态风险扫描，形成待解决问题清单与修复计划 |
| 文档类型 | Bugfix-Plan（含待办清单；后续可按条目拆 `bugfix.md` / `fix-design.md` / `fix-tasks.md`） |
| 主归属 | 前端 AI 交付文档（跨栈问题同步后端域 owner/notice/common 等） |
| 创建时间 | 2026-04-22 |
| 分析方法 | 仓库静态阅读 + 关键字检索；**非**生产日志/压测/全链路 E2E 结论 |

---

## 1. 分析范围

| 区域 | 路径 | 已实现模块（包/目录级） |
|------|------|---------------------------|
| 后端 | `src/spmp-backend/src/main/java/com/spmp/` | `base`、`billing`、`common`、`notice`、`owner`、`user`、`workorder` |
| PC 前端 | `src/spmp-web-pc/src/` | Vue3 + Pinia + Element Plus + Axios 封装 |
| H5 前端 | `src/spmp-web-h5/src/` | Vue3 + Pinia + Vant + Axios 封装 |
| E2E | `src/spmp-e2e/` | Playwright（本次未执行用例，仅作回归参考） |

---

## 2. 待解决问题清单

说明：**严重度**为静态评估（P0 安全/资金/越权，P1 功能错误/数据不一致，P2 体验/可维护/配置风险）。**状态**均为「待确认/待修复」，实施前请结合环境与发布阶段裁剪。

| ID | 严重度 | 域/端 | 问题摘要 | 证据与说明 | 建议方向 |
|----|--------|--------|----------|------------|----------|
| B-01 | P0 | common（安全） | 系统**未完成初始化**时，`SecurityConfig` 对 **任意请求 `permitAll()`**，等同于未认证可访问全部业务接口（仅依赖 `InitializationFilter` 返回 503 的时序与覆盖面）。 | `SecurityConfig.java`：`!initService.isInitialized()` 分支 | 收紧未初始化态：仅放行 `/api/v1/init/**`、健康检查等；其余仍走鉴权或统一 503/维护页策略；补充集成测试。 |
| B-02 | P0 | common（安全） | **JWT 密钥**存在代码/配置默认值，仓库 `application.yml` 含 `jwt.secret` 默认串，易在误配环境下被伪造 Token。 | `JwtTokenProvider.java`：`defaultSecretKeyForDevelopmentOnly1234`；`application.yml` `jwt.secret` | 生产强制环境变量注入、启动校验（长度/随机性）、文档禁止提交真实密钥。 |
| B-03 | P0 | common（安全） | **默认数据源账号**写入 `application.yml`（本地默认值），存在泄露与误连风险。 | `application.yml` `spring.datasource` 默认 `username`/`password` | 默认改为占位 + `local` profile 专用文件；CI 扫描敏感信息。 |
| B-04 | P1 | common（安全） | **Swagger / OpenAPI** 与 **Actuator health** 在白名单，生产若未关闭或网络未隔离，扩大攻击面。 | `SecurityConfig.java` `DEFAULT_WHITE_LIST` | 按 profile 控制暴露；生产关闭或加网关认证/IP 限制。 |
| B-05 | P1 | common（安全） | **CORS** 同时存在 `SecurityConfig` 中 `allowedOriginPatterns("*")` + `allowCredentials(true)`，以及 `WebMvcConfig` 中 `allowedOrigins`（默认 `*`）+ `allowCredentials(true)`；与规范组合可能导致浏览器拒绝或形成宽松跨域。 | `SecurityConfig.java` `corsConfigurationSource()`；`WebMvcConfig.java` `addCorsMappings` | 收敛为单一 CORS 来源配置；生产使用明确域名列表 + `allowedOriginPatterns`。 |
| B-06 | P1 | common（安全/正确性） | **文件下载** `GET /{category}/{filename}`：`filename` 未规范化，`category` 上传接口未白名单校验，存在 **路径穿越** 写入/读取风险（取决于 OS 路径解析）。 | `FileController.java`：`Paths.get(uploadBasePath, category, filename)`；`FileServiceImpl.upload` 直接使用 `category` 建目录 | `category` 枚举/正则白名单；`Path.normalize` + 校验结果路径必须以 `uploadBasePath` 为前缀；下载 `Content-Type` 按实际类型返回。 |
| B-07 | P1 | common（安全） | **文件下载** 响应头固定 `MediaType.IMAGE_JPEG`，与实际上传类型（png/gif/webp）不一致，可能影响客户端/缓存/CDN 行为。 | `FileController.java` `download` | 按扩展名或探测类型设置 `Content-Type`。 |
| B-08 | P1 | common（安全） | **数据权限拦截器** SELF 级别将 `username` 以 **字符串拼接** 注入 SQL 片段；若用户名含 `'` 等字符，存在 **SQL 注入/语法破坏** 风险。 | `DataPermissionInterceptor.java` `selfField + " = '" + username + "'"` | 使用参数化或严格转义/白名单字段名；MyBatis 层参数绑定。 |
| B-09 | P1 | common（正确性） | **数据权限** `addWhereCondition` 使用 `lastIndexOf("WHERE")` 拼接 AND，对含子查询/WHERE 的复杂 SQL 可能 **改错语义**。 | `DataPermissionInterceptor.java` | 使用解析器或约定仅对简单 Mapper SQL 生效；增加单测覆盖典型 SQL。 |
| B-10 | P1 | notice | **H5 公告不可见 / 重推无效**：快照生成 SQL 限制 `owner_status = 'CERTIFIED'`；`repush` 仅 `repushUnread`，不 **增量补快照**。与已存在后端 Bugfix 分析一致。 | `AnnouncementTargetSnapshotMapper.xml`；`NoticePublishServiceImpl.repush` | 与 `CUST-AIDEV-BACKEND-DOCS/domain/notice-center/vibe/Fix-001-h5-notice-invisible/bugfix-plan.md` 对齐后实施。 |
| B-11 | P2 | billing/workorder | 多个 **`@Scheduled`** 任务（账单逾期、支付超时、工单受理超时等）未见 **分布式锁/防重入** 证据（单机多实例或重叠执行时可能重复处理）。 | `OverdueCheckTask`、`PaymentTimeoutTask`、`AcceptTimeoutTask` 等 | 评估部署拓扑；必要时 ShedLock / DB 锁 / 幂等键。 |
| B-12 | P2 | base/common | 大量 **`catch (Exception)`**（Excel 导入、缓存预热、Redis 工具等）可能 **吞掉根因** 或仅打日志，排障困难。 | `BuildingExcelListener`、`BaseCacheWarmupRunner`、`RedisUtils` 等 | 区分可恢复/不可恢复；业务异常向上抛；关键路径 metrics。 |
| B-13 | P2 | common（演进） | `SecurityConfig` 继承 **`WebSecurityConfigurerAdapter`**（Spring Security 5.x 旧式），后续升级 Boot 3 需迁移 `SecurityFilterChain`。 | `SecurityConfig.java` | 纳入技术债清单，与框架升级任务绑定。 |
| F-01 | P1 | PC/H5（认证） | Axios **401** 仅移除 `access_token` 并跳转；**PC** 侧 `refresh_token` 仍可能残留于 localStorage（与 store 不一致）；**未实现静默刷新**流程。 | `spmp-web-pc/src/utils/request.ts`；`spmp-web-pc/src/store/modules/user.ts` | 401 时统一 `resetState`；或实现 refresh 队列 + 后端契约对齐。 |
| F-02 | P2 | H5（认证） | H5 登录 **`saveTokenAndRedirect`** 直接写 `localStorage`，与 Pinia 双写；若后续改为仅 store 持久化，易出现 **状态分裂**。 | `spmp-web-h5/src/views/login/LoginView.vue` | 收敛为单一入口（store action）写 token。 |
| F-03 | P2 | PC/H5（安全） | **Access Token 存 localStorage**，同源 XSS 可导致令牌窃取；需依赖 CSP/依赖审计与输入净化。 | `request.ts`、各 `store` | 评估 `httpOnly` Cookie + CSRF 或短期 Token + 严格 CSP；与后端会话策略一致。 |
| F-04 | P2 | PC/H5（体验） | 响应拦截器对 **HTTP 200 但 `code !== 200`** 一律 `ElMessage`/`showToast`，调用方 `catch` 空实现时用户 **只看到泛化错误**（如登录页 `loadCaptcha` `catch { // 忽略 }`）。 | `LoginView.vue` `loadCaptcha` | 区分静默错误与可提示错误；关键路径必须反馈。 |

---

## 3. Bugfix-Plan（基于清单的执行策略）

### 3.1 优先级与批次

1. **第一批次（安全基线，建议阻塞上生产）**：B-01～B-06、B-08、B-03、B-02。  
2. **第二批次（功能正确性/一致性）**：B-10（与已有 notice Fix-001 合并决策）、B-07、B-09、F-01。  
3. **第三批次（工程化与体验）**：B-11、B-12、B-04、B-05、B-13、F-02～F-04。

### 3.2 文档与任务拆分（与仓库 Bugfix 规范对齐）

- **已独立分析**：B-10 建议继续以 `notice-center/vibe/Fix-001-h5-notice-invisible/` 为主产出 `bugfix.md` / `fix-design.md` / `fix-tasks.md`；本 Fix-001 **引用**即可，避免重复叙述。  
- **本目录后续**：若你批准本 Plan，可将 **P0/P1** 条目分别拆为子目录 `Fix-00x-...`（按域 `common-center`、`notice-center` 等）或在本目录追加 `bugfix.md`（三段式）+ `fix-design.md` + `fix-tasks.md`。

### 3.3 回归测试策略（总纲）

- **不变行为**：登录、初始化向导、公告发布/审批、文件上传下载、带数据权限的分页查询。  
- **期望行为**：修复后未认证用户不能访问受保护资源；文件路径无法穿越；SELF 数据权限下特殊字符用户名查询正常；公告在目标业主侧可见且重推可覆盖新用户。  
- **建议自动化**：对 B-06/B-08/B-10 增加后端单测或集成测；PC/H5 对 401/刷新流程增加契约测试或 E2E（`spmp-e2e`）。

---

## 4. 澄清问题（`[Question]` / `[Answer]`）

> 按 Bugfix 规范：以下问题需要你 `[Answer]` 后，再进入 `bugfix.md` 定稿与开发执行，避免修复方向与发布策略冲突。

**[Question] Q1**：生产与预发环境是否 **永远** 在完成系统初始化后才对外暴露？若存在「未初始化即公网可访问」窗口，是否接受 **第一批次** 将 B-01 列为 P0 并立即改安全策略？  

**[Question] Q2**：公告业务上，**未认证业主**（`UNCERTIFIED`）是否 **必须** 收到已发布公告？（决定 B-10 是否去掉 `CERTIFIED` 条件或改为可配置。）  

**[Question] Q3**：文件上传的 `category` 是否只允许固定枚举（如 `workorder`）？若允许扩展，白名单由谁维护（配置表还是 `application.yml`）？  

**[Question] Q4**：PC 端是否已约定 **Refresh Token 轮换** 与 401 刷新流程？若无，是否接受 F-01 仅做「401 全清 token + 跳转」的一致性修复，静默刷新放到独立 CR？  

**[Question] Q5**：本清单中 **P2** 项是否纳入当前迭代，还是仅作技术债登记？  

---

## 5. 参考资料（代码定位示例）

以下路径便于评审时快速打开（行号随分支可能漂移，以文件内实际为准）：

- 后端安全：`src/spmp-backend/src/main/java/com/spmp/common/security/SecurityConfig.java`  
- JWT：`.../common/security/JwtTokenProvider.java`  
- 文件：`.../common/controller/FileController.java`、`.../common/service/impl/FileServiceImpl.java`  
- 数据权限：`.../common/security/DataPermissionInterceptor.java`  
- 公告快照 SQL：`src/spmp-backend/src/main/resources/mapper/notice/AnnouncementTargetSnapshotMapper.xml`  
- 公告重推：`.../notice/service/impl/NoticePublishServiceImpl.java`（`repush`）  
- 前端请求：`src/spmp-web-pc/src/utils/request.ts`、`src/spmp-web-h5/src/utils/request.ts`  

---

## 6. 状态

| 阶段 | 状态 |
|------|------|
| Bugfix-Plan | 已完成（待用户 `[Answer]` 澄清项） |
| bugfix.md / fix-design.md / fix-tasks.md | 未生成（待批准后按条目拆分） |
| 代码修复 | 未开始 |

---

*本清单为静态分析结论，实施前建议结合运行态日志、渗透测试结果更新严重度与范围。*
