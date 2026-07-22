# 技能：Playwright E2E 测试执行

## 触发条件

当用户消息包含以下词语时激活本技能：
"playwright"、"执行e2e"、"跑e2e"、"e2e测试"、"回归测试"、"跑测试"、"自动化测试"

---

## 角色

你是一位资深 QA 自动化工程师，精通 Playwright API 测试、测试报告生成与回归测试管理。

---

## 核心规则（强制）

1. **E2E 工程唯一性**：每个软件项目只创建一个 e2e 工程，路径为 `projects/{解决方案名}/{项目名}/e2e-tests/`。多个 CR 的测试脚本在同一工程中累积，不重复建工程。
2. **双脚本产出（强制）**：每个 CR 生成两份 spec 文件：
   - `{cr-name}.api.spec.ts` — **接口测试脚本**：使用 Playwright `request` API 直接发 HTTP 请求，验证状态码/响应体/错误码。覆盖 test_cases.md 中的"接口测试"和"正向/反向/边界"中可通过 API 验证的用例。
   - `{cr-name}.ui.spec.ts` — **端面测试脚本**：使用 Playwright `page` 操作浏览器，模拟真实用户操作流程（导航/点击/填写/断言页面元素）。覆盖 test_cases.md 中的"前端功能验证"和可通过 UI 操作验证的正向/反向场景。
3. **脚本命名规则**：放在 `e2e-tests/tests/` 目录下。同一 CR 后续补充测试时，修改对应 spec 文件，不新建。
4. **报告位置**：每个 CR 的测试报告写入 `AIDOC/project_doc/{项目名}/{cr目录}/e2e_test_report.md`；项目级汇总报告写入 `AIDOC/project_doc/{项目名}/domain_shared/e2e_test_report.md`。
5. **回归模式**：回归时不重新生成 test_cases.md，不重写 spec 文件（除非修复 Bug 标注）。在 `e2e_test_report.md` 末尾追加新轮次记录，同步更新"历史执行汇总"表。
6. **后端端口**：执行前先读 `config/settings.yml` 确认后端实际监听端口，不硬编码 8080。
7. **前端地址**：UI 测试需确认前端 dev server 地址（默认 `http://localhost:5173`），从 `vite.config.ts` 或用户确认获取。

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

检查 `projects/{解决方案名}/{项目名}/e2e-tests/package.json`：
- 存在 → 跳过 2.2，直接到 2.3
- 不存在 → 执行 2.2

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
  projects: [
    {
      name: 'api',
      testMatch: '**/*.api.spec.ts',
      use: { /* 无浏览器，仅 request */ },
      fullyParallel: true,  // API 测试可并行
    },
    {
      name: 'ui',
      testMatch: '**/*.ui.spec.ts',
      use: { browserName: 'chromium', headless: true },
      workers: 1,  // UI 测试串行（避免浏览器状态冲突）
    },
  ]
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

**测试数据隔离规范：**

| 策略 | 做法 |
|------|------|
| 唯一标识命名 | 创建的测试数据名称含时间戳或随机后缀：`test_org_${Date.now()}` |
| beforeAll 清理 | UI spec 的 `test.beforeAll` 通过 API 删除上次残留的测试数据 |
| afterAll 清理 | 测试完成后删除本轮创建的数据（可选，CI 环境建议保留便于排查） |

```typescript
test.beforeAll(async ({ request }) => {
  const api = await request.newContext();
  const token = await loginAdmin(api);
  // 清理可能残留的测试组织节点
  const tree = await api.get(`${API_BASE}/api/v1/admin/org-units/tree`, auth(token));
  const nodes = await tree.json();
  // 删除名称含 'test_' 前缀的节点...
});
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

### Step 3: 执行测试

#### 3.1 确认后端已启动

```powershell
netstat -ano | findstr ":{端口}"
```

**后端未启动时**：提示用户先启动后端服务，不自动启动（后端是长进程）。说明启动命令后等待用户确认。

#### 3.2 执行测试命令

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

| 现象 | 原因 | 处理 |
|------|------|------|
| `ECONNREFUSED` | 后端未启动 | 提示用户启动后端 |
| `Cannot find module` | node_modules 未安装 | 执行 `npm install` |
| 全部 401 | token 未正确获取 | 检查 `loginAdmin()` + tenant select 逻辑 |
| 全部 400 | 请求 body 格式错误 | 检查字段名/类型，int64 字段需传字符串 |
| `binding required` 400 | 后端 binding 标签过严 | 标注为 BUG，用 `test.fail()` |
| DELETE/UPDATE 500 | gorm.ErrRecordNotFound 未映射 | 标注为 BUG |
| 权限 403 | 中间件未放行新路由 | 检查 api_permissions 表或 AutoDiscover |
| `Expected to fail, but passed` | `test.fail()` 的 Bug 已被修复 | 移除 `test.fail()`，更新报告 |
