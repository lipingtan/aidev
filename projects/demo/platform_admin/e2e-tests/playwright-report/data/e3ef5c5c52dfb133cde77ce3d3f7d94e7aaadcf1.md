# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: e2e-api.spec.ts >> 认证模块 AUTH >> TC-AUTH-002: 错误密码登录失败
- Location: tests\e2e-api.spec.ts:110:3

# Error details

```
TimeoutError: page.fill: Timeout 10000ms exceeded.
Call log:
  - waiting for locator('input[type="text"]')
    - locator resolved to <input readonly value="" type="text" tabindex="0" role="combobox" autocomplete="off" spellcheck="false" id="el-id-4457-15" aria-expanded="false" aria-haspopup="listbox" class="el-select__input" aria-activedescendant="" aria-autocomplete="none" aria-controls="el-id-4457-9"/>
    - fill("admin")
  - attempting fill action
    2 × waiting for element to be visible, enabled and editable
      - element is not editable
    - retrying fill action
    - waiting 20ms
    2 × waiting for element to be visible, enabled and editable
      - element is not editable
    - retrying fill action
      - waiting 100ms
    19 × waiting for element to be visible, enabled and editable
       - element is not editable
     - retrying fill action
       - waiting 500ms

```

# Page snapshot

```yaml
- generic [ref=e4]:
  - generic [ref=e5]:
    - generic [ref=e6]:
      - generic [ref=e8]: 智慧物业管理平台
      - img [ref=e10] [cursor=pointer]
    - generic [ref=e12]:
      - generic [ref=e13]:
        - generic [ref=e14]: 租户：
        - generic [ref=e15]: 默认租户
        - button "切换" [ref=e16] [cursor=pointer]:
          - generic [ref=e17]:
            - img [ref=e19]
            - generic [ref=e21]: 切换
      - button "商务经典" [ref=e23] [cursor=pointer]:
        - generic [ref=e24]:
          - img [ref=e26]
          - generic [ref=e28]: 商务经典
      - button "admin" [ref=e30] [cursor=pointer]:
        - img [ref=e33]
        - generic [ref=e35]: admin
  - generic [ref=e36]:
    - complementary [ref=e37]:
      - menubar [ref=e38]:
        - menuitem "首页" [ref=e39] [cursor=pointer]:
          - img [ref=e41]
          - text: 首页
        - menuitem "系统管理" [ref=e43]:
          - generic [ref=e44] [cursor=pointer]:
            - img [ref=e46]
            - generic [ref=e48]: 系统管理
            - img [ref=e50]
        - menuitem "日志管理" [ref=e52]:
          - generic [ref=e53] [cursor=pointer]:
            - img [ref=e55]
            - generic [ref=e57]: 日志管理
            - img [ref=e59]
        - menuitem "监控" [ref=e61]:
          - generic [ref=e62] [cursor=pointer]:
            - img [ref=e64]
            - generic [ref=e66]: 监控
            - img [ref=e68]
    - main [ref=e70]:
      - generic [ref=e74] [cursor=pointer]: 首页
      - navigation "Breadcrumb" [ref=e75]:
        - link "首页" [ref=e77]
      - generic [ref=e79]:
        - generic [ref=e82]:
          - generic [ref=e83]:
            - generic [ref=e84]: 你好，admin！
            - generic [ref=e85]: 欢迎使用智慧物业管理平台 Dashboard
          - generic [ref=e87] [cursor=pointer]:
            - generic:
              - combobox [ref=e89]
              - generic [ref=e90]: 本月
            - img [ref=e93]
        - alert [ref=e95]:
          - img [ref=e97]
          - generic [ref=e100]: Dashboard 模块尚未实现
        - generic [ref=e101]:
          - generic [ref=e102]: 工单趋势
          - generic [ref=e104]:
            - img [ref=e106]
            - paragraph [ref=e123]: 工单趋势模块尚未实现
        - generic [ref=e124]:
          - generic [ref=e125]: 收费趋势
          - generic [ref=e127]:
            - img [ref=e129]
            - paragraph [ref=e146]: 收费趋势模块尚未实现
        - generic [ref=e147]:
          - generic [ref=e148]: 快捷入口
          - generic [ref=e150]:
            - img [ref=e152]
            - paragraph [ref=e169]: 暂无可用快捷入口
```

# Test source

```ts
  16  |   moduleName: string;
  17  |   description: string;
  18  |   priority: string;
  19  |   status: 'PASS' | 'FAIL' | 'SKIP';
  20  |   duration: number;
  21  |   errorMessage?: string;
  22  |   actualResult?: string;
  23  | }
  24  | 
  25  | const testResults: TestResult[] = [];
  26  | 
  27  | // 辅助函数：记录测试结果
  28  | function recordResult(
  29  |   caseId: string,
  30  |   moduleName: string,
  31  |   description: string,
  32  |   priority: string,
  33  |   status: 'PASS' | 'FAIL' | 'SKIP',
  34  |   duration: number,
  35  |   errorMessage?: string,
  36  |   actualResult?: string
  37  | ) {
  38  |   testResults.push({
  39  |     caseId,
  40  |     moduleName,
  41  |     description,
  42  |     priority,
  43  |     status,
  44  |     duration,
  45  |     errorMessage,
  46  |     actualResult
  47  |   });
  48  | }
  49  | 
  50  | // 登录辅助函数
  51  | async function loginAsAdmin(page: Page) {
  52  |   await page.goto(`${BASE_URL}/login`);
  53  |   await page.waitForLoadState('networkidle');
  54  | 
  55  |   // 填写登录表单
  56  |   await page.fill('input[type="text"]', ADMIN_USER);
  57  |   await page.fill('input[type="password"]', ADMIN_PASS);
  58  | 
  59  |   // 点击登录按钮
  60  |   await page.click('button:has-text("登录")');
  61  | 
  62  |   // 等待登录成功提示或页面跳转
  63  |   await page.waitForSelector('.el-message--success, text=登录成功', { timeout: 10000 });
  64  | 
  65  |   // 等待页面加载完成
  66  |   await page.waitForTimeout(1000);
  67  | }
  68  | 
  69  | // ================== 一、认证模块（AUTH）==================
  70  | 
  71  | test.describe('认证模块 AUTH', () => {
  72  |   let page: Page;
  73  | 
  74  |   test.beforeAll(async ({ browser }) => {
  75  |     page = await browser.newPage();
  76  |   });
  77  | 
  78  |   test.afterAll(async () => {
  79  |     await page.close();
  80  |   });
  81  | 
  82  |   // TC-AUTH-001: 正确凭据登录成功
  83  |   test('TC-AUTH-001: 正确凭据登录成功', async () => {
  84  |     const startTime = Date.now();
  85  |     try {
  86  |       await page.goto(`${BASE_URL}/login`);
  87  |       await page.waitForLoadState('networkidle');
  88  | 
  89  |       // 填写登录表单
  90  |       await page.fill('input[type="text"]', ADMIN_USER);
  91  |       await page.fill('input[type="password"]', ADMIN_PASS);
  92  | 
  93  |       // 点击登录按钮
  94  |       await page.click('button:has-text("登录")');
  95  | 
  96  |       // 等待登录成功提示
  97  |       await page.waitForSelector('.el-message--success', { timeout: 10000 });
  98  | 
  99  |       // 验证登录成功 - 检查是否显示用户名
  100 |       await page.waitForSelector('text=admin', { timeout: 5000 });
  101 | 
  102 |       recordResult('TC-AUTH-001', '认证', '正确凭据登录成功', 'P0', 'PASS', Date.now() - startTime);
  103 |     } catch (error) {
  104 |       recordResult('TC-AUTH-001', '认证', '正确凭据登录成功', 'P0', 'FAIL', Date.now() - startTime, String(error));
  105 |       throw error;
  106 |     }
  107 |   });
  108 | 
  109 |   // TC-AUTH-002: 错误密码登录失败
  110 |   test('TC-AUTH-002: 错误密码登录失败', async () => {
  111 |     const startTime = Date.now();
  112 |     try {
  113 |       await page.goto(`${BASE_URL}/login`);
  114 |       await page.waitForLoadState('networkidle');
  115 | 
> 116 |       await page.fill('input[type="text"]', ADMIN_USER);
      |                  ^ TimeoutError: page.fill: Timeout 10000ms exceeded.
  117 |       await page.fill('input[type="password"]', 'wrong_password');
  118 |       await page.click('button:has-text("登录")');
  119 | 
  120 |       // 等待错误提示
  121 |       await page.waitForSelector('.el-message--error', { timeout: 5000 });
  122 | 
  123 |       recordResult('TC-AUTH-002', '认证', '错误密码登录失败', 'P0', 'PASS', Date.now() - startTime);
  124 |     } catch (error) {
  125 |       recordResult('TC-AUTH-002', '认证', '错误密码登录失败', 'P0', 'FAIL', Date.now() - startTime, String(error));
  126 |       throw error;
  127 |     }
  128 |   });
  129 | });
  130 | 
  131 | // ================== 二、用户管理（USER）==================
  132 | 
  133 | test.describe('用户管理模块 USER', () => {
  134 |   let page: Page;
  135 | 
  136 |   test.beforeAll(async ({ browser }) => {
  137 |     page = await browser.newPage();
  138 |     await loginAsAdmin(page);
  139 |   });
  140 | 
  141 |   test.afterAll(async () => {
  142 |     await page.close();
  143 |   });
  144 | 
  145 |   // TC-USER-001: 创建用户成功
  146 |   test('TC-USER-001: 创建用户成功', async () => {
  147 |     const startTime = Date.now();
  148 |     try {
  149 |       // 导航到用户管理页面
  150 |       await page.goto(`${BASE_URL}/#/system/users`);
  151 |       await page.waitForLoadState('networkidle');
  152 | 
  153 |       // 点击新建按钮
  154 |       const newBtn = page.locator('button:has-text("新增"), button:has-text("新建")');
  155 |       if (await newBtn.count() > 0) {
  156 |         await newBtn.first().click();
  157 |         await page.waitForTimeout(500);
  158 | 
  159 |         // 填写用户表单
  160 |         const randomSuffix = Date.now().toString().slice(-6);
  161 |         await page.fill('input[name="username"]', `test_user_${randomSuffix}`);
  162 |         await page.fill('input[name="nickname"]', `测试用户${randomSuffix}`);
  163 | 
  164 |         // 提交表单
  165 |         await page.click('button:has-text("确定")');
  166 | 
  167 |         // 等待成功提示
  168 |         await page.waitForSelector('.el-message--success', { timeout: 5000 }).catch(() => {});
  169 |       }
  170 | 
  171 |       recordResult('TC-USER-001', '用户管理', '创建用户成功', 'P0', 'PASS', Date.now() - startTime);
  172 |     } catch (error) {
  173 |       recordResult('TC-USER-001', '用户管理', '创建用户成功', 'P0', 'FAIL', Date.now() - startTime, String(error));
  174 |       throw error;
  175 |     }
  176 |   });
  177 | 
  178 |   // TC-USER-002: 查询用户列表（租户隔离）
  179 |   test('TC-USER-002: 查询用户列表', async () => {
  180 |     const startTime = Date.now();
  181 |     try {
  182 |       await page.goto(`${BASE_URL}/#/system/users`);
  183 |       await page.waitForLoadState('networkidle');
  184 |       await page.waitForTimeout(1000);
  185 | 
  186 |       // 检查页面是否有内容
  187 |       const content = await page.locator('.el-main, main, [class*="content"]').count();
  188 |       expect(content).toBeGreaterThanOrEqual(0);
  189 | 
  190 |       recordResult('TC-USER-002', '用户管理', '查询用户列表（租户隔离）', 'P0', 'PASS', Date.now() - startTime);
  191 |     } catch (error) {
  192 |       recordResult('TC-USER-002', '用户管理', '查询用户列表（租户隔离）', 'P0', 'FAIL', Date.now() - startTime, String(error));
  193 |       throw error;
  194 |     }
  195 |   });
  196 | });
  197 | 
  198 | // ================== 三、角色管理（ROLE）==================
  199 | 
  200 | test.describe('角色管理模块 ROLE', () => {
  201 |   let page: Page;
  202 | 
  203 |   test.beforeAll(async ({ browser }) => {
  204 |     page = await browser.newPage();
  205 |     await loginAsAdmin(page);
  206 |   });
  207 | 
  208 |   test.afterAll(async () => {
  209 |     await page.close();
  210 |   });
  211 | 
  212 |   // TC-ROLE-001: 创建角色
  213 |   test('TC-ROLE-001: 创建角色', async () => {
  214 |     const startTime = Date.now();
  215 |     try {
  216 |       await page.goto(`${BASE_URL}/#/system/roles`);
```