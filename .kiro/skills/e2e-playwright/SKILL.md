# 技能：Playwright E2E 测试执行

## 触发条件

当用户消息包含以下词语时激活本技能：
"playwright"、"执行e2e"、"跑e2e"、"e2e测试"、"回归测试"、"跑测试"、"自动化测试"

---

## 角色

你是一位资深 QA 自动化工程师，精通 Playwright API 测试、测试报告生成与回归测试管理。

---

## 核心规则（强制）

1. **E2E 工程路径约定**：每个软件项目只创建一个 e2e 工程。路径约定：
   - 标准路径：`projects/{解决方案名}/{项目名}/e2e/`（已有工程优先复用，不重建）
   - 备选路径：`projects/{解决方案名}/{项目名}/e2e-tests/`（历史工程已存在时沿用）
   - **执行前必须先检查两个路径是否已存在**，存在哪个用哪个，不新建重复工程。
   - 多个 CR 的测试脚本在同一工程中累积。
2. **双脚本产出（强制）**：每个 CR 生成两份 spec 文件：
   - `{cr-name}.api.spec.ts` — **接口测试脚本**：使用 Playwright `request` API 直接发 HTTP 请求，验证状态码/响应体/错误码。覆盖 test_cases.md 中的"接口测试"和"正向/反向/边界"中可通过 API 验证的用例。
   - `{cr-name}.ui.spec.ts` — **端面测试脚本**：使用 Playwright `page` 操作浏览器，模拟真实用户操作流程（导航/点击/填写/断言页面元素）。覆盖 test_cases.md 中的"前端功能验证"和可通过 UI 操作验证的正向/反向场景。
3. **脚本命名规则**：放在 `e2e-tests/tests/` 目录下。同一 CR 后续补充测试时，修改对应 spec 文件，不新建。
4. **报告位置**：每个 CR 的测试报告写入 `AIDOC/project_doc/{项目名}/{cr目录}/e2e_test_report.md`；项目级汇总报告写入 `AIDOC/project_doc/{项目名}/domain_shared/e2e_test_report.md`。
5. **回归模式**：回归时不重新生成 test_cases.md，不重写 spec 文件（除非修复 Bug 标注）。在 `e2e_test_report.md` 末尾追加新轮次记录，同步更新"历史执行汇总"表。
6. **后端端口**：执行前先读 `config/settings.yml` 确认后端实际监听端口，不硬编码 8080。
7. **前端地址**：UI 测试需确认前端 dev server 地址（默认 `http://localhost:5173`），从 `vite.config.ts` 或用户确认获取。

---

## 测试失败处理原则（强制）

### 第一步：区分"脚本错误"还是"程序错误"

断言失败后，**不得直接修复代码，也不得直接调整断言**，必须先完成以下归因判定：

```
归因判定优先级（从高到低）：

1. 需求文档（requirements.md）
   └─ FR-N 的验收标准是否明确定义了此处预期行为？
   └─ 如果 FR 有明确规定 → 以 FR 为准，程序实现不符则是程序错误

2. 设计文档（design.md）
   └─ API 设计表中接口的响应格式/错误码是否有明确定义？
   └─ 状态转换图中该场景的转换结果是否有定义？
   └─ 如果 design 有明确定义 → 以 design 为准

3. 任务拆解（tasks.md）
   └─ Acceptance 标准中是否对该行为有约束？
   └─ 如果 tasks Acceptance 有描述 → 以 tasks 为准

4. 行业最佳实践（判据不足时的兜底）
   └─ 仅当以上三个文档均未明确定义时才使用
   └─ 使用行业标准（REST 规范/HTTP 语义/业务惯例）补充判定
   └─ 必须在测试报告中说明所采用的行业实践依据
```

### 归因结果与处置

| 归因结论 | 含义 | 处置方式 |
|---------|------|---------|
| **程序错误** | 实现与需求/设计不符 | 修复程序代码，保持断言不变 |
| **脚本错误** | 断言编写有误，与需求/设计不符 | 修正脚本断言，在报告中记录调整理由 |
| **需求/设计不明确** | 文档未覆盖此场景 | 以行业最佳实践补充判定，报告中注明依据 |
| **需求/设计有冲突** | 各文档定义不一致 | 停止自行判定，向用户提出澄清，不得假设 |

### 脚本调整记录规范

当归因结论为"脚本错误"或"需求/设计不明确（用行业实践补充）"时，**必须在测试报告中记录**：

```markdown
### 断言调整说明

| 用例 | 原断言 | 调整后断言 | 判定依据 | 依据来源 |
|------|--------|-----------|---------|---------|
| TC-N04 | `toMatch(/APPROVED\|无法撤销\|状态/)` | `toMatch(/APPROVED\|无法撤销\|状态\|原因/)` | 需求文档未定义具体错误信息格式，按行业实践 REST 错误信息应包含拒绝原因 | 行业最佳实践（REST 错误响应规范） |
```

**禁止无记录地修改断言。** 即使是"明显的脚本笔误"，也必须在报告中简短说明。

### 根因定位，不得绕过

确认是程序错误后，**必须从源头修复，严禁以下绕过行为**：

| 禁止行为 | 说明 |
|---------|------|
| 修改测试断言来适配错误的返回值 | 断言是需求规约的体现，不能降低验收标准 |
| 将失败用例标记为 `test.skip()` 掩盖问题 | skip 只用于"环境未就绪"或"功能计划外" |
| 改写测试逻辑绕过验证步骤 | 绕过等于放弃测试价值 |
| 接受"当前行为"为正确而不查根因 | 必须找出实现与需求不符的原因 |

### 程序错误根因分析流程

```
Step 1: 确认失败现象
  └─ 读取完整错误信息（状态码、响应体、错误堆栈）
  └─ 区分：环境问题 / 表不存在 / 接口未注册 / 响应格式不符

Step 2: 追踪调用链
  └─ 接口未找到(404) → 检查路由注册
  └─ 未认证(401)     → 检查中间件配置 / token 获取逻辑
  └─ 服务错误(500)   → 读取后端日志，找具体 error 消息
  └─ 响应格式不符    → 对比后端实际响应与测试预期，找代码差异

Step 3: 定位代码根因
  └─ 路由问题   → 检查 router.go / server.go 的路由注册
  └─ DB 问题    → 检查 AutoMigrate / 表是否存在 / 字段定义
  └─ 格式问题   → 对比响应格式，找 handler 的输出函数（e.OK vs c.JSON）

Step 4: 修复根因
  └─ 修改代码（路由/模型/handler）
  └─ 重新编译并验证编译通过
  └─ 重启服务，验证服务日志无异常
  └─ 重新执行测试
```

### 深度链路调试（一次分析未定位到根因时强制执行）

**触发条件**：按 Step 1-4 分析后，修复了一个假设的根因，重新编译部署测试后问题仍然复现，或者根本无法确定根因在哪个层次。

**原则：必须在代码里加链路日志，不得凭猜测反复修改代码**。

#### 链路日志添加规范

```
目标：在中间件链每个关键节点加 log.Printf，覆盖"请求从哪进来、在哪消失、带了什么状态出去"
```

**中间件链日志模板（Go）**：

```go
// 在每个中间件的关键分支点加日志
log.Printf("[DEBUG-MW-xxx] enter path=%s method=%s", c.Request.URL.Path, c.Request.Method)
log.Printf("[DEBUG-MW-xxx] aborted=%v status=%d body_len=%d", c.IsAborted(), c.Writer.Status(), bodyLen)
log.Printf("[DEBUG-MW-xxx] ABORT reason=%s", reason)
log.Printf("[DEBUG-MW-xxx] PASS calling c.Next()")
log.Printf("[DEBUG-MW-xxx] c.Next() returned aborted=%v status=%d", c.IsAborted(), c.Writer.Status())
```

**Handler 入口日志**（放在函数第一行）：
```go
log.Printf("[DEBUG-HANDLER-xxx] ENTER path=%s", c.FullPath())
// 每个关键步骤后也加日志
log.Printf("[DEBUG-HANDLER-xxx] step=db_query count=%d err=%v", count, err)
log.Printf("[DEBUG-HANDLER-xxx] step=response_written")
```

#### 链路日志分析要点

加完日志编译重启后，分析输出时注意：

| 现象 | 含义 | 下一步 |
|------|------|--------|
| 中间件日志有，handler 日志无 | handler 未被调用，某个中间件 Abort 了 | 找最后一条中间件日志，查该中间件的 ABORT 分支 |
| `c.Next() returned` 没有出现 | 该中间件的 `c.Next()` 没有返回 | 查 `c.Next()` 内部是否 panic（看 Recovery 日志） |
| gin 请求日志在中间件 `c.Next()` 日志之后打印 | 这是正常的，gin 请求日志在整个链完成后才写出 | 不要用 gin 日志的打印顺序判断执行顺序，改用 `log.Printf` |
| HTTP 200 空 body | Recovery 捕获了 panic，但 header 已提前被发出 | 查 handler 里是否有 nil pointer 访问（如 `nil DB`，`nil Orm`） |
| HTTP 200 空 body 且无 Recovery 日志 | 中间件链的 ResponseWriter 被拦截器（如 FieldFilterMiddleware）包装，`Abort` 的 body 被写入 buffer 但 buffer 写出时 header 已是 200 | 绕过中间件直接调用接口看原始状态码 |

#### 链路调试的典型陷阱（来自实战）

1. **`2>&1` 重定向下 PowerShell 缓冲区截断**：`get_process_output` 每次只返回有限行，后面的日志可能被截断。解决方式：增大 `lines` 参数，或在请求后 `Sleep` 一段时间再取日志。

2. **gin 请求日志时序误导**：gin 框架的请求日志（`info GET /api/... 200`）是在整个中间件链执行完后才写出，不代表 handler 执行完了。必须用业务代码里的 `log.Printf` 判断实际执行顺序。

3. **`go-admin SDK MakeOrm()` nil 陷阱**：`MakeOrm()` 用 `c.Request.Host` 作 key 查 DB，如果 key 不匹配会返回 nil DB。后续 `s.Orm.Model()` 触发 nil pointer panic，被 gin Recovery 捕获，导致 HTTP 200 空 body（header 已由框架设为 200）。修复方式：不用 `MakeOrm()`，改为直接从 `sdk.Runtime.GetDb()` 遍历取第一个可用 DB。

4. **FieldFilterMiddleware 包装 ResponseWriter 的副作用**：如果上游中间件 `AbortWithStatusJSON(403)` 的响应被 `FieldFilterMiddleware` 的 buffer 拦截，而 buffer 最终写出时 gin 状态码已被某处覆盖为 200，客户端会看到 200 空 body。这类问题需通过直接 HTTP 调用（不经过测试脚本）验证真实状态码来排查。

5. **新注册的路由未加入权限白名单**：通过 `RegisterExtraAdminRoutes()` 等扩展机制注册的路由，不会自动加入 `DynamicPermissionMiddleware` 的 `skipPaths` 或 `admin_api_permission` 表。必须显式加白名单，否则即使是 SUPER_ADMIN 也可能因为 `admin_role` 表里没有对应的 `SUPER_ADMIN` role_code 而被拒绝。

#### 链路日志清理（强制）

定位到根因并修复后，**必须清理所有临时调试日志**，不得提交：
- 删除所有 `[DEBUG-MW-xxx]` / `[DEBUG-HANDLER-xxx]` 格式的 `log.Printf`
- 删除调试用的临时 `log.Printf` import（如果原文件没有）
- 重新编译确认 0 错误后再跑最终 E2E

### 常见根因与正确修复方式

| 失败现象 | 归因判定 | 正确处置 |
|---------|---------|---------|
| 接口返回 404 | 程序错误（路由未注册） | 检查并修复路由注册，断言不变 |
| 接口返回 `code:200` 但测试期望 `code:0` | 程序错误（响应格式不符合 design.md 定义） | 修复 handler 响应格式，断言不变 |
| DB 表不存在导致 500 | 程序错误（AutoMigrate 未执行） | 修复服务启动时的表迁移逻辑 |
| GORM AutoMigrate 失败（invalid default value） | 程序错误（模型字段定义有误） | 检查并修复 GORM tag 中的冲突 |
| 错误信息文本与断言正则不匹配 | **先查 FR/design/tasks**：若文档未定义具体文本 → 脚本错误，按行业实践调整断言并记录；若文档有定义 → 程序错误，修改返回信息 |
| auth-setup 失败导致 chromium project 全不跑 | 脚本错误（测试依赖配置有误） | 为纯 API 测试新增不依赖 auth-setup 的 project |
| 使用旧进程缓存输出判断结果 | 环境问题 | 停止旧进程重新启动，等待完整输出 |

### 二进制/进程状态确认（部署后必须验证）

修改代码后执行测试前，**必须按以下顺序确认**，否则测试结果无效：

```
1. 调用 list_processes 确认旧后端进程的 terminalId
2. 调用 control_pwsh_process stop {terminalId} 停止旧进程
3. 等待 stop 返回成功（Success）后才能继续——必须确认旧进程已关闭
4. go build 编译（编译成功才继续，编译失败先修复）
5. 启动新编译的二进制（control_pwsh_process start）
6. 等待服务启动完成（get_process_output 确认日志无 Fatal 错误）
7. 停止旧的 E2E 进程（isReused=true 的进程输出可能是缓存）
8. 重新启动新的 E2E 进程
9. 等待完整输出后再读取结果
```

**严禁跳过步骤 2-3**：在 Windows 上，如果旧进程未关闭就重新编译，`go build` 会因为 `.exe` 文件被占用而失败；即使编译成功（输出到不同文件名），旧进程仍在监听原端口，新进程无法启动，测试打的是旧服务。

**判断进程是否为缓存**：`control_pwsh_process` 返回 `isReused: true` 时，该进程输出可能是上一次的缓存，必须先 stop 再重新 start。

---

## 工作流程

### Step 0: 检查 Playwright 环境

```powershell
# 检查 Node.js
node --version

# 检查 e2e 工程依赖是否已安装
# 进入 e2e-tests 目录检查 node_modules 是否存在
Test-Path "node_modules"
```

**情况处理：**

| 情况 | 处理 |
|------|------|
| node_modules 不存在 | 执行 `npm install` |
| playwright 浏览器未安装 | 执行 `npx playwright install chromium` |
| Node.js 未安装 | 提示用户先安装 Node.js，无法继续 |

安装完成后继续执行。

---

### Step 1: 确定任务类型

| 用户意图 | 判断依据 | 执行路径 |
|---------|---------|---------|
| 首次为某 CR 创建测试 | 无对应 spec 文件 | → Step 2（初始化） |
| 为已有 CR 补充用例 | 用户明确说"补充"/"追加" | → Step 2.4（仅创建/修改 spec） |
| 执行已有 CR 测试 | 有对应 spec 文件，无"补充"意图 | → Step 3（执行） |
| 跨 CR 全量回归 | 用户说"全量回归"/"全部跑一遍" | → Step 3（执行全部） |

#### 测试模式选择（API vs UI）

**每个用例都必须判断采用哪种模式，不能全部默认 API：**

| 场景特征 | 适用模式 | 原因 |
|---------|---------|------|
| 验证 HTTP 状态码/响应体字段/错误码 | API 模式 | 快，直接 |
| 验证 CRUD 接口基本功能 | API 模式 | 不依赖前端 |
| 验证用户登录后看到什么菜单/页面 | **UI 模式** | 需要完整浏览器会话 |
| 验证字段在页面上是否显示/隐藏 | **UI 模式** | 需要 DOM 断言 |
| 验证多步骤业务流（创建→分配→切换用户→验证效果） | **UI 模式** | API 模式需手动模拟 token 注入，脆弱 |
| 验证前端组件交互（弹窗/Tab切换/Tree 选择） | **UI 模式** | 只有 UI 能验证 |
| 验证认证守卫（无 token 请求被拒绝） | API 模式 | 不需要页面 |

**核心原则**：如果测试需要"以某个用户的视角看到什么"，用 UI 模式；如果只验证"接口返回什么"，用 API 模式。

#### UI 模式实现模式

```typescript
// 通过 API 准备数据 + 注入 localStorage token + page 操作验证
// 避免在 UI 上做复杂数据准备（慢且脆弱）

// 1. API 准备：创建用户/角色/配置权限
const adminToken = await apiLogin(api, 'admin', 'admin123');
await api.post('/users', { data: {...} });
await api.post('/users/:id/roles', { data: {...} });

// 2. 获取目标用户 token
const userAuth = await apiLogin(api, username, password);

// 3. 注入浏览器 localStorage
await page.goto(WEB_BASE);
await page.evaluate((data) => {
  localStorage.setItem('access_token', data.access_token);
  localStorage.setItem('platform_token', data.platform_token);
  localStorage.setItem('current_tenant_id', data.current_tenant_id);
  localStorage.setItem('tenant_list', JSON.stringify(data.tenants));
}, userAuth);
await page.goto(`${WEB_BASE}/#/home`);

// 4. 页面断言
await page.waitForSelector('.el-menu');
expect(await page.locator('.el-menu-item').count()).toBeGreaterThan(0);
```

---

### Step 2: 初始化（首次或补充）

#### 2.1 判断 e2e 工程是否已存在

按以下顺序检查：
1. `projects/{解决方案名}/{项目名}/e2e/package.json` → 存在则使用此路径
2. `projects/{解决方案名}/{项目名}/e2e-tests/package.json` → 存在则使用此路径
3. 两者都不存在 → 执行 2.2，在 `e2e/` 下新建

存在时跳过 2.2，直接到 2.3。

#### 2.2 创建 e2e-tests 工程骨架

目录结构：
```
e2e-tests/
├── tests/           # 所有 spec 文件
├── package.json
├── playwright.config.ts
└── tsconfig.json
```

`playwright.config.ts` 关键配置：
- `timeout: 60000`
- `reporter: [['html', { open: 'never' }], ['list']]`
- 使用 `projects` 区分 API 和 UI 测试：
  ```typescript
  export default defineConfig({
    timeout: 60000,
    fullyParallel: false,   // 顶层控制，API project 内部可覆盖
    workers: 1,
    reporter: [['html', { open: 'never' }], ['list']],
    projects: [
      {
        name: 'api',
        testMatch: '**/*.api.spec.ts',
        use: {},            // 纯 API 测试，无需浏览器配置
        // fullyParallel 在 defineConfig 顶层控制，不在 project 内设置
      },
      {
        name: 'ui',
        testMatch: '**/*.ui.spec.ts',
        use: { browserName: 'chromium', headless: true },
        // UI 测试串行（workers:1 在顶层已设置）
      },
    ],
  });
  ```
- 后端 `API_BASE` 在各 spec 文件中单独定义，不在 config 中统一

#### 2.3 定位测试用例文档（必须完成，不可跳过）

按以下顺序查找 `test_cases.md`：

1. 用户消息中是否直接提供了路径 → 直接使用
2. 当前打开的编辑器文件中是否有 `test_cases.md` → 读取并使用
3. 推断路径 `AIDOC/project_doc/{项目名}/{cr目录}/test_cases.md` → 检查文件是否存在

**找不到时，停止并询问：**

```
未找到该 CR 的测试用例文档（test_cases.md）。请选择：

1. 告诉我 test_cases.md 的具体路径
2. 先用 test-case-gen 技能生成测试用例，再执行 E2E 测试

请告知你的选择。
```

收到用户答复后再继续。**不得在缺少 test_cases.md 的情况下自行猜测测试用例并生成 spec 文件。**

#### 2.4 创建 / 修改 spec 文件

**双脚本生成规则：**

从 test_cases.md 中的用例按以下规则分配到两份 spec：

| 用例类型 | 归属脚本 | 验证方式 |
|---------|---------|---------|
| 接口测试（TC-A*） | `{cr}.api.spec.ts` | HTTP request → 断言状态码/body |
| 正向/反向/边界中「可通过 API 触发且断言 HTTP 响应」的 | `{cr}.api.spec.ts` | HTTP request |
| 回归测试（TC-R*） | `{cr}.api.spec.ts` | HTTP request 验证行为不变 |
| 前端功能（TC-F*） | `{cr}.ui.spec.ts` | page 操作 + 元素断言 |
| 正向/反向中「需通过页面操作才能完成」的 | `{cr}.ui.spec.ts` | page 操作 |
| 单元测试级（需 Mock） | 标注 `test.skip('单元测试级别')` 在 api spec 中 |
| 数据验证（需直连 DB） | 标注 `test.skip('需数据库直连验证')` 在 api spec 中 |

---

**API spec 文件结构（`{cr-name}.api.spec.ts`）：**

```typescript
/**
 * {CR名称} — 接口测试
 * 关联用例: AIDOC/project_doc/{项目名}/{cr目录}/test_cases.md
 */
import { test, expect, APIRequestContext, request } from '@playwright/test';

const API_BASE = 'http://localhost:{后端端口}';
// ... loginAdmin / auth / extractID 公共函数 ...

test.describe('接口测试 — {模块名}', () => {
  // TC-A01, TC-A02 ...
});

test.describe('正向测试 — API 验证', () => {
  // TC-001（通过 API 调用验证）...
});

test.describe('反向测试 — API 验证', () => {
  // TC-N01 ...
});

test.describe('回归测试', () => {
  // TC-R01 ...
});
```

---

**UI spec 文件结构（`{cr-name}.ui.spec.ts`）：**

```typescript
/**
 * {CR名称} — 端面测试（用户操作流程）
 * 关联用例: AIDOC/project_doc/{项目名}/{cr目录}/test_cases.md
 */
import { test, expect, Page } from '@playwright/test';

const APP_URL = 'http://localhost:5173';  // 前端 dev server
const API_BASE = 'http://localhost:{后端端口}';

/** 浏览器内登录（操作登录页面） */
async function uiLogin(page: Page, username = 'admin', password = 'admin123') {
  await page.goto(`${APP_URL}/login`);
  await page.getByPlaceholder('请输入用户名').fill(username);
  await page.getByPlaceholder('请输入密码').fill(password);
  await page.getByRole('button', { name: '登录' }).click();
  // 处理租户选择（如有）
  const tenantSelect = page.locator('[data-testid="tenant-select"]');
  if (await tenantSelect.isVisible({ timeout: 2000 }).catch(() => false)) {
    await tenantSelect.locator('[data-testid="tenant-item"]').first().click();
  }
  await page.waitForURL('**/home**', { timeout: 10000 });
}

/** 导航到指定菜单 */
async function navigateTo(page: Page, path: string) {
  await page.goto(`${APP_URL}${path}`);
  await page.waitForLoadState('networkidle');
}

test.describe('端面测试 — {功能模块}', () => {
  test.beforeEach(async ({ page }) => {
    await uiLogin(page);
  });

  test('TC-F01: {场景名} — 用户操作流', async ({ page }) => {
    await navigateTo(page, '/system/org');
    // 断言页面元素（使用显式等待，禁止 waitForTimeout）
    await expect(page.locator('[data-testid="org-tree"], .el-tree')).toBeVisible();
    // 执行操作
    await page.getByRole('button', { name: '新增顶级节点' }).click();
    // 断言弹窗出现
    await expect(page.locator('.el-dialog')).toBeVisible();
    // ...
  });
});
```

---

**UI spec 端面操作断言规范：**

| 断言对象 | 写法 |
|---------|------|
| 元素可见 | `await expect(locator).toBeVisible()` |
| 文本内容 | `await expect(locator).toContainText('xxx')` |
| 表格行数 | `await expect(page.locator('tr')).toHaveCount(N)` |
| Toast 成功提示 | `await expect(page.locator('.el-message--success')).toBeVisible()` |
| Toast 错误提示 | `await expect(page.locator('.el-message--error')).toContainText('xxx')` |
| 对话框消失 | `await expect(page.locator('.el-dialog')).not.toBeVisible()` |
| 下拉选项存在 | `await expect(page.locator('.el-select-dropdown__item:has-text("xxx")')).toBeVisible()` |

---

**选择器策略（优先级从高到低）：**

| 优先级 | 策略 | 示例 | 适用场景 |
|--------|------|------|---------|
| 1 | `data-testid` | `page.locator('[data-testid="org-tree"]')` | 前端已添加测试标记 |
| 2 | 语义化选择器 | `page.getByRole('button', { name: '新增' })` | 按钮/链接等有明确角色 |
| 3 | placeholder/label | `page.getByPlaceholder('请输入用户名')` | 表单输入框 |
| 4 | CSS class + text | `page.locator('.el-menu-item:has-text("系统管理")')` | 最后备选 |

**规则**：优先使用 `data-testid`（要求前端在关键交互元素上添加）。不可用时依次退回语义化选择器→placeholder→CSS+text。禁止使用 XPath 或层级过深的 CSS 选择器。

---

**导航策略：**

UI 测试导航到目标页面推荐使用**直接路由导航**，避免菜单折叠/动画等不确定性：

```typescript
/** 导航到指定页面（推荐：直接 URL 导航） */
async function navigateTo(page: Page, path: string) {
  await page.goto(`${APP_URL}${path}`);
  await page.waitForLoadState('networkidle');
}

// 用法
await navigateTo(page, '/system/org');  // 替代点击菜单层级
```

仅在测试「菜单可见性/菜单权限」时才使用点击菜单方式导航。

---

**测试数据隔离规范（强制）：**

| 规则 | 要求 |
|------|------|
| **幂等性** | 测试必须可重复执行，每次结果一致。用时间戳/随机后缀命名创建的数据，避免名称冲突 |
| **跑前清理** | `beforeAll` 里先删本次将要创建的同名残留数据，再创建新数据 |
| **describe 独立** | 每个 `describe` 只使用自己 `beforeAll` 创建的数据，不依赖其他 describe 的副作用 |
| **失败后留存** | 测试失败时不清理数据（便于排查），但要在测试名称中包含时间戳使下次跑不冲突 |
| **不共享可变状态** | token 可跨 describe 共享，但创建的实体 ID、业务数据不得跨 describe 共享 |

```typescript
test.beforeAll(async () => {
  api = await request.newContext()
  ;({ token, userId } = await loginAdmin(api))

  // 幂等性：先清理可能残留的同名数据
  const listResp = await api.get(
    `${API_BASE}/api/v1/admin/approval-flows?flow_code=${encodeURIComponent(flowCode)}`,
    authHeader(token)
  )
  const body = await listResp.json()
  const existing = body.data?.list ?? []
  for (const item of existing) {
    await api.delete(`${API_BASE}/api/v1/admin/approval-flows/${item.id}`, authHeader(token))
  }

  // 再创建本次测试需要的数据
  await createApprovalFlow(api, token, flowCode, userId)
})
```

**spec 文件必须包含的元素：**

```typescript
/**
 * {CR名称} E2E 测试
 * 关联用例: AIDOC/project_doc/{项目名}/{cr目录}/test_cases.md
 * 执行日期: YYYY-MM-DD
 */
import { test, expect, APIRequestContext, request } from '@playwright/test';

const API_BASE = 'http://localhost:{后端端口}';  // 从 settings.yml 读取
const ADMIN_USER = 'admin';
const ADMIN_PASS = 'admin123';

/** 登录并获取 access_token，自动处理多租户 tenant select */
async function loginAdmin(api: APIRequestContext): Promise<string> {
  const resp = await api.post(`${API_BASE}/auth/login`, {
    data: { username: ADMIN_USER, password: ADMIN_PASS },
  });
  const body = await resp.json();
  if (body.data?.tenants?.length > 0) {
    const tenantResp = await api.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(body.data.tenants[0].id) },
      headers: { Authorization: `Bearer ${body.data.token}` },
    });
    const tb = await tenantResp.json();
    return tb.data?.access_token ?? tb.data?.token;
  }
  return body.data?.access_token ?? body.data?.token;
}

/** 注入 Authorization header */
function auth(token: string) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } };
}

/** 从响应中提取 ID（字符串形式，避免雪花 ID 精度丢失） */
function extractID(body: any): string {
  const raw = body?.data?.id ?? body?.data?.role_id ?? body?.data?.share_id;
  return String(raw);
}
```

**雪花 ID 精度规范（强制）：**

本项目后端使用雪花算法（Snowflake）生成主键 ID，值域超过 JavaScript `Number.MAX_SAFE_INTEGER`（9007199254740991）。**禁止**在测试脚本中使用 `parseInt()` 或 `Number()` 转换 ID 值，否则会导致精度丢失、请求错误的资源 ID。

| 规则 | 正确写法 | 错误写法 |
|------|----------|----------|
| 提取响应中的 ID | `String(body.data.id)` | `parseInt(body.data.id, 10)` |
| ID 变量声明 | `let roleID: string` | `let roleID: number` |
| 列表中提取 ID | `.map((r: any) => String(r.id))` | `.map((r: any) => Number(r.id))` |
| ID 用于 URL 拼接 | `` `/roles/${roleID}/resources` `` | 同左（string 变量直接拼接） |
| ID 用于 JSON body | `{ resource_ids: [id1, id2] }` | 同左（已经是 string 数组） |
| ID 相等性比较 | `String(a.id) === targetID` | `Number(a.id) === targetID` |

---

## 登录会话复用规则（强制）

**核心原则**：避免每个测试用例都重复执行登录操作。登录应作为共享状态复用，除非测试场景本身与登录逻辑强相关。

### API 测试中的登录复用

```typescript
// ✅ 正确：在 describe 级别共享 token
test.describe('biz_user 管理接口', () => {
  let token: string;
  
  test.beforeAll(async ({ request }) => {
    token = await loginAdmin(request);
  });

  test('列表查询', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/v1/admin/biz-users`, auth(token));
    // ...
  });

  test('创建用户', async ({ request }) => {
    const resp = await request.post(`${API_BASE}/api/v1/admin/biz-users`, {
      ...auth(token),
      data: { phone: '13800138000' },
    });
    // ...
  });
});

// ❌ 错误：每个 test 都重新登录
test('列表查询', async ({ request }) => {
  const token = await loginAdmin(request);  // 浪费时间！
  // ...
});
```

### UI 测试中的登录复用

```typescript
// ✅ 正确：使用 storageState 复用浏览器会话
// playwright.config.ts 中配置 globalSetup 执行一次登录并保存状态
// 或在 beforeAll 中通过 API 获取 token 注入 localStorage

test.describe('C端用户管理页面', () => {
  test.beforeAll(async ({ browser }) => {
    const page = await browser.newPage();
    await uiLogin(page, 'admin', 'admin123');
    await page.context().storageState({ path: 'tests/.auth/admin.json' });
    await page.close();
  });

  test.use({ storageState: 'tests/.auth/admin.json' });

  test('列表页展示正确', async ({ page }) => {
    // 直接导航，无需再次登录
    await navigateTo(page, '/system/biz-user');
    await expect(page.locator('.el-table')).toBeVisible();
  });
});
```

### 必须重新登录的场景（例外）

仅当以下场景时才在 test 内部执行登录：

| 场景 | 原因 |
|------|------|
| 测试登录页面本身的功能（验证码/错误提示） | 登录是被测对象 |
| 测试登出后的行为 | 需要先登出再验证 |
| 测试"强制登出后 token 失效" | 需要重新登录获取新 token 对比 |
| 测试不同角色/不同租户的权限差异 | 需要以不同身份登录 |
| 测试"重新登录后菜单/权限已更新" | 验证登录触发的刷新逻辑 |

**规则总结**：如果用例的核心验证点不是「登录本身」或「登录触发的副作用」，必须复用已有会话，不得重复登录。

---

### Step 3: 执行测试

#### 3.0 执行优先级策略（强制）

**不得一次性跑完所有用例再统一排查**。采用分层递进策略，先确保核心链路通过，再扩展覆盖。

```
阶段1 — 核心链路验证（P0，必须 100% 通过才能继续）
  └─ 只跑 P0 用例：认证登录 + 核心正向流程 + 关键状态转换
  └─ 有任何失败 → 立即停止，定位根因并修复，不得继续跑后续用例
  └─ 全部通过 → 进入阶段2

阶段2 — 完整 API 测试（P0 + P1）
  └─ 跑全部接口测试（api spec）
  └─ 有 P0 失败 → 停止修复，优先级高于 P1
  └─ 仅 P1 失败 → 记录 Bug，可继续跑阶段3
  └─ 全部通过 → 进入阶段3

阶段3 — 端面/回归测试（P1 + P2）
  └─ 跑 UI spec 和回归用例
  └─ 记录所有失败，按优先级排队修复
```

**核心链路定义**（以下用例类型视为 P0 核心链路）：
- 认证：登录成功、token 认证通过
- 主实体 CRUD：创建正向、查询列表、删除正向
- 核心状态转换：业务实体的主路径状态流转（如审批的 PENDING→APPROVED）
- 回归：所有 RG-N 回归用例（防止已有功能被破坏）

**修复顺序原则**：
1. P0 失败 → 当前批次内必须修复，不得推迟
2. P1 失败且影响后续用例的前置条件 → 必须先修复（如 beforeAll 里的数据创建失败）
3. P1/P2 独立失败 → 记录为 Bug，标注 `test.fail()`，继续执行其他用例

```powershell
# PowerShell 原生写法（推荐，跨版本稳定）
netstat -ano | Select-String "LISTENING" | Select-String ":8000"

# CMD 备选写法
netstat -ano | findstr "LISTENING" | findstr ":8000"
```

**后端未启动时**：提示用户先启动后端服务，不自动启动（后端是长进程）。说明启动命令后等待用户确认。

#### 3.2 执行测试命令

> **超时与重试配置**：
> - 本地开发环境：`retries: 0`（失败立即停，便于调试）
> - CI 环境：`retries: 1`（偶发网络抖动自动重试一次，两次都失败才算真实 Bug）
> - 涉及异步任务/EventBus 的用例，在用例内用 `test.setTimeout(30000)` 单独延长
> - 不得用 `waitForTimeout` 做固定等待，应用 `waitForResponse` 或轮询断言

```typescript
// CI 环境 playwright.config.ts 建议配置
export default defineConfig({
  retries: process.env.CI ? 1 : 0,
  // ...
})

// 单个用例延长超时（涉及异步处理时）
test('TC-011 EventBus 异步联动', async () => {
  test.setTimeout(30000)
  // ...
})
```

```powershell
# 执行单个 CR 的接口测试
npx playwright test tests/{cr-name}.api.spec.ts --reporter=list

# 执行单个 CR 的端面测试
npx playwright test tests/{cr-name}.ui.spec.ts --reporter=list

# 执行单个 CR 的全部测试（接口 + 端面）
npx playwright test tests/{cr-name}.*.spec.ts --reporter=list

# 执行全部测试（全量回归）
npx playwright test --reporter=list

# 同时生成 HTML 报告
npx playwright test tests/{cr-name}.*.spec.ts --reporter=html
```

#### 3.3 解析执行结果

从命令输出中统计以下四类结果：

| 输出特征 | 含义 | 统计为 |
|---------|------|-------|
| `ok` | 用例通过 | ✅ 通过 |
| `x`（无 `test.fail`） | 未预期失败 | ❌ 失败 |
| `x`（有 `test.fail`，失败符合预期） | 已知 Bug | 🐛 已知 Bug |
| `Expected to fail, but passed` | `test.fail()` 的用例实际通过了 | ⚠️ Bug 已修复，需移除 `test.fail()` |
| `skipped` | 用例被跳过 | ⏭️ 跳过 |

**"Expected to fail, but passed" 的处理**：立即提示用户，该 Bug 可能已被修复，需要移除对应 spec 中的 `test.fail()`，并在报告中更新 Bug 状态为"已修复"。

---

### Step 4: 写入测试报告

#### 4.1 首次测试 → 创建 CR 专属报告

在 `AIDOC/project_doc/{项目名}/{cr目录}/e2e_test_report.md` 创建文件，使用下方报告模板（第 1 轮）。

#### 4.2 回归测试 → 追加轮次

读取已有报告，在文件末尾追加新的执行批次段落（参考模板），**同时更新"历史执行汇总"表中的对应行**。历史记录不删改。

#### 4.3 更新项目级汇总报告

在 `domain_shared/e2e_test_report.md` 中找到该 CR 对应章节，更新最新执行结果（无对应章节则新建）。

---

## 测试报告模板

```markdown
# E2E 测试报告：{CR名称}

**关联 CR**: {cr目录}
**关联用例**: test_cases.md
**测试脚本**: projects/{解决方案}/{项目}/e2e-tests/tests/{cr-name}.spec.ts
**项目**: {项目名}

---

## 执行记录 — 第 1 轮（YYYY-MM-DD）

**执行类型**: 全量测试 / 回归测试
**触发原因**: 首次 CR 验收 / CR-N 合入后回归 / Bugfix 后验证

### 测试环境

| 项目 | 值 |
|------|-----|
| 后端地址 | http://localhost:{端口} |
| 测试账号 | admin / admin123 |
| 测试框架 | Playwright {版本} |
| 浏览器 | Chromium |
| Git Commit | {执行 `git rev-parse --short HEAD` 获取，如 `a3f9c12`} |
| 后端版本 | {执行 `.\backend_test.exe version` 或记录构建时间} |
| Node 版本 | {执行 `node --version` 获取} |

### 执行统计

| 脚本类型 | ✅ 通过 | ❌ 失败 | 🐛 已知Bug | ⚠️ Bug已修复 | ⏭️ 跳过 | 小计 |
|---------|--------|--------|-----------|------------|--------|------|
| API spec | N | N | N | N | N | N |
| UI spec | N | N | N | N | N | N |
| **总计** | **N** | **N** | **N** | **N** | **N** | **N** |

**测试结论**: ✅ 通过 / ❌ 不通过 / ⚠️ 有风险

> 结论判定：P0 全部通过 + P1 通过率 ≥ 95% → ✅；P0 全部通过 + P1 < 95% → ⚠️；任意 P0 失败 → ❌

### 发现的 Bug（本轮新增）

| Bug ID | 关联用例 | 描述 | 优先级 | 状态 |
|--------|---------|------|--------|------|
| BUG-001 | TC-XXX | {描述} | P0/P1 | 待修复 / 已修复 |

### 失败用例分析（仅未预期失败）

| 用例 | 现象 | 预期 | 根因 | 处理建议 |
|------|------|------|------|---------|
| TC-XXX | ... | ... | ... | ... |

---

## 历史执行汇总

> 每次追加新轮次后同步更新此表

| 轮次 | 日期 | 类型 | 通过 | 失败 | 已知Bug | 跳过 | 结论 |
|------|------|------|------|------|---------|------|------|
| 1 | YYYY-MM-DD | 全量 | N | N | N | N | ✅/❌/⚠️ |
```

---

## 已知 Bug 处理规范

### 标注方式（spec 文件中）

```typescript
// BUG: {原因简述} — 后端未实现 / 参数校验缺失 / 错误码映射错误
test('TC-XXX {用例名称} [BUG: {简短说明}]', async () => {
  test.fail(); // 已知后端 bug，预期此用例失败
  // 测试逻辑照常写，便于 Bug 修复后直接移除 test.fail() 即可
});
```

### Bug 修复后的处理步骤

1. 移除 `test.fail()` 和测试名称中的 `[BUG: ...]`
2. 重新执行该 spec，确认用例通过
3. 在测试报告中将该 Bug 状态更新为"已修复"，并在对应轮次记录中注明

---

## 回归测试规则（与 test-case-gen 技能对齐）

| 触发场景 | 执行范围 |
|---------|---------|
| 单个 CR Bugfix 后验证 | 该 CR 的 spec 文件全部用例 |
| 新 CR 合入后回归 | 新 CR spec 全部用例 + 关联旧 CR spec 全部用例 |
| 大版本发布前 | `npx playwright test`（全部 spec） |

> 说明：自动化执行以 spec 文件为单位（即全部用例）。手工执行时可仅跑 test_cases.md 中"四、回归测试"章节的用例作为最小回归集合。

---

## 目录约定速查

```
projects/{解决方案}/{项目}/
└── e2e-tests/                          ← E2E 工程（每个项目唯一）
    ├── tests/
    │   ├── {cr1-name}.api.spec.ts      ← CR1 接口测试脚本
    │   ├── {cr1-name}.ui.spec.ts       ← CR1 端面测试脚本
    │   ├── {cr2-name}.api.spec.ts      ← CR2 接口测试脚本
    │   ├── {cr2-name}.ui.spec.ts       ← CR2 端面测试脚本
    │   └── ...
    ├── playwright.config.ts
    ├── package.json
    └── tsconfig.json

AIDOC/project_doc/{项目}/
├── {cr1目录}/
│   └── e2e_test_report.md              ← CR1 专属测试报告
├── {cr2目录}/
│   └── e2e_test_report.md              ← CR2 专属测试报告
└── domain_shared/
    └── e2e_test_report.md              ← 项目级汇总报告（跨 CR）
```

---

## 错误排查速查

> 排查前先完成归因判定（见"测试失败处理原则"），确认是程序错误还是脚本错误，再按下表定位根因。

| 现象 | 归因方向 | 正确处置 |
|------|---------|---------|
| `ECONNREFUSED` | 环境问题 | 提示用户启动后端，不自动处理 |
| `Cannot find module` | 环境问题 | 执行 `npm install` |
| 全部 401 | 程序/脚本错误 | 先查 FR/design：若规定需认证 → 检查 `loginAdmin()` + tenant select 逻辑；若脚本漏传 header → 修正脚本 |
| 全部 400 | 程序/脚本错误 | 检查请求 body 字段名/类型，int64 字段需传字符串；若 binding 标签过严 → 程序错误，修复 binding |
| 接口 404 | 程序错误（路由未注册） | 检查并修复路由注册，断言不变 |
| `code:200` 但期望 `code:0` | 程序错误（响应格式不符） | 修复 handler 响应函数，断言不变 |
| DB 表不存在导致 500 | 程序错误（AutoMigrate 未执行） | 修复服务启动时的表迁移逻辑 |
| GORM AutoMigrate 失败 | 程序错误（模型字段定义有误） | 检查并修复 GORM tag 冲突 |
| 错误信息文本与断言正则不匹配 | **先查文档**：FR/design 有定义 → 程序错误修代码；文档未定义 → 脚本错误按行业实践调整并记录 | 无论哪种，必须在报告中记录断言调整说明 |
| auth-setup 失败导致 project 全不跑 | 脚本/配置错误 | 为纯 API 测试新增不依赖 auth-setup 的 project |
| isReused: true 进程输出无变化 | 环境问题（缓存输出） | stop 旧进程，重新 start，等完整输出后再读取 |
| `Expected to fail, but passed` | Bug 已修复 | 移除 `test.fail()`，更新报告 Bug 状态为"已修复" |
