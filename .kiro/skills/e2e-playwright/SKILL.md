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
2. **脚本命名规则**：每个 CR 对应一个脚本文件，命名为 `{cr-name}.spec.ts`（全小写 kebab-case），放在 `e2e-tests/tests/` 目录下。同一 CR 后续补充测试时，修改同一 spec 文件，不新建。
3. **报告位置**：每个 CR 的测试报告写入 `AIDOC/project_doc/{项目名}/{cr目录}/e2e_test_report.md`；项目级汇总报告写入 `AIDOC/project_doc/{项目名}/domain_shared/e2e_test_report.md`。
4. **回归模式**：回归时不重新生成 test_cases.md，不重写 spec 文件（除非修复 Bug 标注）。在 `e2e_test_report.md` 末尾追加新轮次记录，同步更新"历史执行汇总"表。
5. **后端端口**：执行前先读 `config/settings.yml` 确认后端实际监听端口，不硬编码 8080。

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
- `workers: 1`（避免并发写数据库冲突）
- `timeout: 60000`
- `reporter: [['html', { open: 'never' }], ['list']]`
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

**如何将 test_cases.md 转为 spec 代码：**

1. 读取 test_cases.md，识别所有章节（正向/反向/边界/回归/接口/数据验证）
2. 按章节创建 `test.describe` 分组，describe 名称与章节名称对应
3. 每条用例 → 一个 `test(...)` 函数，用例编号和名称作为 test 名称
4. 前置条件中的"创建数据"逻辑放入 `test.beforeAll`
5. 无法自动执行的数据验证用例（需要直连 DB）→ 用 `test.skip()` 并注释原因
6. 已知后端缺陷（接口行为与预期不符）→ 用 `test.fail()` 并在名称中标注 `[BUG: 简短说明]`

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
# 执行单个 CR 的测试
npx playwright test tests/{cr-name}.spec.ts --reporter=list

# 执行全部测试（全量回归）
npx playwright test --reporter=list

# 同时生成 HTML 报告
npx playwright test tests/{cr-name}.spec.ts --reporter=html
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

| 状态 | 数量 |
|------|------|
| ✅ 通过 | N |
| ❌ 失败（未预期） | N |
| 🐛 已知 Bug（test.fail 符合预期） | N |
| ⚠️ Bug 已修复（test.fail 但实际通过） | N |
| ⏭️ 跳过 | N |
| **总计** | **N** |

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
    │   ├── {cr1-name}.spec.ts          ← CR1 测试脚本
    │   ├── {cr2-name}.spec.ts          ← CR2 测试脚本（追加）
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
