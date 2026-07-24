# E2E 测试报告：V2-CR5 双用户池 + C端认证

**关联 CR**: v2-cr5-dual-user-auth-strategy
**关联用例**: test_cases.md
**测试脚本**: 
- projects/demo/platform_admin/e2e/tests/v2-cr5-dual-user-auth.api.spec.ts
- projects/demo/platform_admin/e2e/tests/v2-cr5-dual-user-auth.ui.spec.ts
**项目**: platform_admin

---

## 测试覆盖矩阵

### API 测试覆盖

| 用例编号 | 用例名称 | 覆盖状态 | 备注 |
|---------|---------|---------|------|
| TC-001 | 管理端 grant_type=password 登录 | ✅ 已覆盖 | |
| TC-002 | 不传 grant_type 默认密码策略 | ✅ 已覆盖 | |
| TC-N01 | 不支持的 grant_type 返回 400 | ✅ 已覆盖 | |
| TC-N02 | captcha_key 缺少 code 返回 400 | ✅ 已覆盖 | |
| TC-A01 | 发送验证码成功 | ✅ 已覆盖 | |
| TC-A02 | 发送验证码未传 phone | ✅ 已覆盖 | |
| TC-A03 | 发送验证码限频返回 429 | ✅ 已覆盖 | |
| TC-A04 | C端短信登录成功 | ⏭️ 跳过 | 需手动输入验证码 |
| TC-A05 | 验证码错误返回 401 | ✅ 已覆盖 | |
| TC-A06 | 登录缺少 tenant_code 返回 400 | ✅ 已覆盖 | |
| TC-A07 | C端登出成功 | ⏭️ 跳过 | 需 C端 token |
| TC-A08 | C端登出未认证返回 401 | ✅ 已覆盖 | |
| TC-A09 | 获取菜单成功 | ⏭️ 跳过 | 需 C端 token |
| TC-A10 | 获取菜单未认证返回 401 | ✅ 已覆盖 | |
| TC-A11 | admin token 访问 C端菜单 | ✅ 已覆盖 | |
| TC-A12 | 获取 biz_user 列表 | ✅ 已覆盖 | |
| TC-A13 | 未认证访问 biz_user 列表 | ✅ 已覆盖 | |
| TC-A15 | 手机号模糊搜索 | ✅ 已覆盖 | |
| TC-A16 | biz_user 详情 | ✅ 已覆盖 | |
| TC-A17 | 不存在的 ID | ❌ 未覆盖 | 需补充 |
| TC-A18 | 创建 biz_user | ✅ 已覆盖 | |
| TC-A19 | 创建未认证 | ✅ 已覆盖 | (TC-A13 验证) |
| TC-A20 | 创建无权限 | ⏭️ 跳过 | 需无权限账号 |
| TC-A21 | 手机号重复返回 400 | ✅ 已覆盖 | |
| TC-A22 | 更新 biz_user | ✅ 已覆盖 | |
| TC-A23 | 更新未认证 | ✅ 已覆盖 | (TC-A13 验证) |
| TC-A24 | 删除 biz_user | ✅ 已覆盖 | |
| TC-A25 | 删除无权限 | ⏭️ 跳过 | 需无权限账号 |
| TC-A26 | 重置密码 | ✅ 已覆盖 | |
| TC-A27 | 重置密码未认证 | ✅ 已覆盖 | (TC-A13 验证) |
| TC-A28 | 强制登出 | ✅ 已覆盖 | |
| TC-A29 | 强制登出无权限 | ⏭️ 跳过 | 需无权限账号 |
| TC-A30 | 切换状态 | ✅ 已覆盖 | |
| TC-A31 | 切换状态未认证 | ✅ 已覆盖 | (TC-A13 验证) |
| TC-R01 | 不传 grant_type 登录正常 | ✅ 已覆盖 | |
| TC-R02 | grant_type=password 与不传一致 | ✅ 已覆盖 | |
| TC-R03 | SelectTenant 流程正常 | ✅ 已覆盖 | |
| TC-R05 | Logout 黑名单机制 | ✅ 已覆盖 | |
| TC-R09 | admin_user 表不被过滤 | ✅ 已覆盖 | |

### UI 测试覆盖

| 用例编号 | 用例名称 | 覆盖状态 | 备注 |
|---------|---------|---------|------|
| TC-F01 | 列表页表格正确展示数据 | ✅ 已覆盖 | |
| TC-F02 | 手机号搜索过滤正确 | ✅ 已覆盖 | |
| TC-F03 | 新增用户弹窗并提交成功 | ✅ 已覆盖 | |
| TC-F04 | 编辑用户弹窗回填数据 | ❌ 未覆盖 | 需补充 |
| TC-F05 | 重置密码流程 | ❌ 未覆盖 | 需补充 |
| TC-F06 | 启用/禁用开关切换 | ✅ 已覆盖 | |
| TC-F07 | 发送验证码按钮倒计时 | ✅ 已覆盖 | |
| TC-F08 | 登录成功跳转主页 | ⏭️ 跳过 | 需手动输入验证码 |
| TC-F09 | 登录失败显示错误提示 | ✅ 已覆盖 | |
| TC-F10 | 登录后菜单正确加载 | ⏭️ 跳过 | 需 C端登录状态 |
| TC-F11 | 点击登出跳转登录页 | ⏭️ 跳过 | 需 C端登录状态 |
| TC-F12 | 菜单折叠/展开功能 | ⏭️ 跳过 | 需 C端登录状态 |
| TC-UX01 | 布局合理性检查 | ✅ 已覆盖 | |
| TC-UX03 | 完整性检查 | ✅ 已覆盖 | |
| TC-UX06 | 删除操作二次确认 | ✅ 已覆盖 | |
| TC-UX07 | 登录卡片居中布局 | ✅ 已覆盖 | |
| TC-UX08 | placeholder 文案清晰 | ✅ 已覆盖 | |
| TC-UX09 | 加载状态显示 | ✅ 已覆盖 | |
| TC-UX10 | 未填手机号时发送验证码 | ✅ 已覆盖 | |

---

## 执行统计

| 脚本类型 | ✅ 已覆盖 | ❌ 未覆盖 | ⏭️ 跳过 | 小计 |
|---------|---------|---------|--------|------|
| API spec | 32 | 1 | 6 | 39 |
| UI spec | 12 | 2 | 5 | 19 |
| **总计** | **44** | **3** | **11** | **58** |

**覆盖率**: 44/58 = 75.9%

---

## 未覆盖用例清单

| 用例编号 | 用例名称 | 原因 | 处理建议 |
|---------|---------|------|---------|
| TC-A17 | 不存在的 ID 返回 400 | 未实现 | 补充到 API spec |
| TC-F04 | 编辑用户弹窗回填数据 | 未实现 | 补充到 UI spec |
| TC-F05 | 重置密码流程 | 未实现 | 补充到 UI spec |

---

## 跳过用例说明

| 用例编号 | 跳过原因 |
|---------|---------|
| TC-A04, TC-A07, TC-A09 | 需要真实验证码，mock 环境下需手动操作 |
| TC-A20, TC-A25, TC-A29 | 需要无权限测试账号 |
| TC-F08 | 需要手动输入验证码 |
| TC-F10, TC-F11, TC-F12 | 需要 C端用户登录状态 |

---

## 执行命令

```powershell
# 执行 V2-CR5 全部 API 测试
npx playwright test tests/v2-cr5-dual-user-auth.api.spec.ts --reporter=list

# 执行 V2-CR5 全部 UI 测试
npx playwright test tests/v2-cr5-dual-user-auth.ui.spec.ts --reporter=list

# 执行 V2-CR5 全部测试
npx playwright test tests/v2-cr5-dual-user-auth.*.spec.ts --reporter=list

# 生成 HTML 报告
npx playwright test tests/v2-cr5-dual-user-auth.*.spec.ts --reporter=html
```

---

## 备注

1. **验证码测试限制**：由于短信验证码使用内存 mock（控制台打印），自动化测试无法获取真实验证码，相关用例标记为 skip 或使用错误验证码验证失败场景。

2. **权限测试限制**：TC-A14, TC-A20, TC-A25, TC-A29 等需要无权限账号，当前测试环境未准备此类账号，建议后续补充。

3. **C端登录状态**：UI 测试中涉及 C端登录后操作的用例（TC-F10~TC-F12）需要先实现 C端登录流程的自动化（可能需要通过 API 注入 token 到 localStorage）。
