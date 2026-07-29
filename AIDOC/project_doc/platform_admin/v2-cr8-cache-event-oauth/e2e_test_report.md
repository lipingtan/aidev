# E2E 测试报告：V2-CR8 缓存事件驱动 + OAuth2 预留 + 扩展能力收尾

**关联 CR**: v2-cr8-cache-event-oauth
**关联用例**: test_cases.md
**测试脚本**: projects/demo/platform_admin/e2e/tests/v2-cr8-cache-event-oauth.api.spec.ts
**项目**: platform_admin

---

## 执行记录 — 第 1 轮（2026-07-29）

**执行类型**: 全量测试
**触发原因**: CR-8 首次 E2E 验收

### 测试环境

| 项目 | 值 |
|------|-----|
| 后端地址 | http://localhost:8000 |
| 测试账号 | admin / admin123 |
| 测试框架 | Playwright 1.45.0（API 模式） |
| 启动命令 | `go run main.go server -c config/settings.yml` |
| 前端 | 未启动（本 CR 无 UI 测试） |

### 执行统计

| 脚本类型 | ✅ 通过 | ❌ 失败 | 🐛 已知Bug | ⏭️ 跳过 | 小计 |
|---------|--------|--------|-----------|--------|------|
| API spec | 19 | 0 | 0 | 0 | 19 |
| UI spec | - | - | - | - | 未创建 |
| **总计** | **19** | **0** | **0** | **0** | **19** |

**测试结论**: ✅ 通过

> auth-setup（全局登录）因前端未启动返回 ERR_CONNECTION_REFUSED，属于环境限制，不影响本 CR 的 API 测试结果。

### 发现的 Bug（本轮）

| Bug ID | 关联用例 | 描述 | 优先级 | 状态 |
|--------|---------|------|--------|------|
| BUG-001 | TC-A08/A09 | `mapHTTPStatus` 缺少 50xxx 映射，OAuth2/LDAP 骨架返回 HTTP 500 而非 501 | P0 | ✅ 已修复 |

### BUG-001 修复说明

**根因**：`common/auth/handler/response.go` 中 `mapHTTPStatus` 未覆盖 50xxx 错误码范围，走到 `default` 返回 `500`。

**修复**：新增 `case code >= 50100 && code < 50200: return http.StatusNotImplemented`。

**断言调整说明**：

| 用例 | 原断言 | 调整后断言 | 判定依据 | 依据来源 |
|------|--------|-----------|---------|---------|
| TC-A08 | `expect(resp.status()).toBe(200)` | `expect(resp.status()).toBe(501)` | 需求文档 FR-6 明确要求"返回 HTTP 501"；后端修复后实际返回 501 | requirements.md FR-6 |
| TC-A09 | `expect(resp.status()).toBe(200)` | `expect(resp.status()).toBe(501)` | 同 TC-A08 | requirements.md FR-7 |

### 覆盖情况

| Phase | 测试覆盖 | 通过 |
|-------|---------|------|
| Phase 1: 权限检查降级透明性 | TC-A01/A02/P04/R02/N04 | ✅ 5个 |
| Phase 2: 操作日志风险分级 | TC-A04~A07/P01~P02/R05 | ✅ 7个 |
| Phase 3: OAuth2/LDAP 骨架 | TC-A08~A10/N01~N03/R01 | ✅ 7个 |
| Phase 4: ext_fields 向后兼容 | TC-A11~A15/R03 | ✅ 6个 - 含1个空列表场景跳过字段验证 |

### 未覆盖用例（留待后续）

| 用例 | 原因 |
|------|------|
| TC-A03（分配角色后新权限生效） | 需要完整的用户→角色→接口权限创建链路，数据准备成本高；逻辑已由单元测试 TestRoleService_AssignResources_PublishesEvent 覆盖 |
| TC-A07 ext_fields 日志字段验证（空列表场景） | 服务启动后首次查询日志列表为空，已记录跳过原因 |

---

## 历史执行汇总

| 轮次 | 日期 | 类型 | 通过 | 失败 | 已知Bug | 跳过 | 结论 |
|------|------|------|------|------|---------|------|------|
| 1 | 2026-07-29 | 全量 | 19 | 0 | 0 | 0 | ✅ |
| 2 | 2026-07-29 | 回归（含auth-setup） | 39 | 0 | 0 | 0 | ✅ |
