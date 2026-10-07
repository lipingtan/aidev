# 后端统一错误码文档

> 版本：v1.0（2026-10-04，v2-cr12 建立）
> 来源文件：`backend/common/errors/errors.go`

---

## 设计原则

- 错误码为 5 位整数，前两位表示模块，后三位表示具体错误
- 后端响应格式：`{ "code": <错误码>, "msg": <描述>, "error_code": <同 code> }`
- 成功响应格式：`{ "code": 0, "msg": "ok", "data": ... }`
- HTTP 状态码由 `apperrors.HTTPStatus(code)` 推导（401/403/400/500），前端不要依赖 HTTP status 做业务逻辑，只看 `code` / `error_code` 字段
- 新模块从 6xxxx 开始顺延，不复用已废弃错误码

---

## 错误码段规划

| 段 | 范围 | 模块 |
|----|------|------|
| `1xxxx` | 10001–10199 | 通用错误（跨模块） |
| `2xxxx` | 20001–20499 | 平台基础服务（认证/租户/用户/角色/菜单） |
| `3xxxx` | 30001–30299 | license 插件（激活码/会员/订单/设备） |
| `4xxxx` | 40001–40199 | 游戏中心（CR12，游戏/分类/公告/Banner/CDN/VIP） |
| `5xxxx` | 50001–50004 | 系统/基础设施 |
| `6xxxx+` | 预留 | 新插件顺延 |

---

## 通用错误码 1xxxx

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 10001 | ErrCodeInvalidParam | 请求参数非法（格式/类型/范围） |
| 10002 | ErrCodeParamTooLong | 参数超过长度限制 |
| 10003 | ErrCodeParamEmpty | 必填参数为空 |
| 10004 | ErrCodeParamFormat | 参数格式不合法（字符集/正则等） |
| 10005 | ErrCodePageSizeExceeded | page_size 超过上限（最大100） |
| 10101 | ErrCodeNotFound | 资源不存在 |
| 10102 | ErrCodeDuplicate | 资源已存在（唯一约束冲突） |
| 10103 | ErrCodeConflict | 资源状态冲突 |
| 10104 | ErrCodeQuotaExceeded | 数量配额超限 |

---

## 平台基础服务 2xxxx

### 认证/授权 20001–20010

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 20001 | ErrCodeInvalidCredentials | 用户名或密码错误 |
| 20002 | ErrCodeTokenExpired | Token 已过期 |
| 20003 | ErrCodeTokenBlacklisted | Token 已失效（登出/刷新后旧 Token） |
| 20004 | ErrCodeAccountDisabled | 账号已禁用 |
| 20005 | ErrCodeTenantDisabled | 租户已禁用 |
| 20006 | ErrCodeNotAssociatedTenant | 用户未关联该租户 |
| 20007 | ErrCodeIPBlacklisted | IP 因连续失败被封禁 |
| 20008 | ErrCodePermissionDenied | 权限不足 |
| 20009 | ErrCodeDataScopeViolation | 数据范围越权 |
| 20010 | ErrCodeForbiddenTenant | 租户访问被禁止 |

### 租户管理 20100–20199

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 20101 | ErrCodeTenantNotFound | 租户不存在 |
| 20102 | ErrCodeTenantDuplicate | 租户已存在 |
| 20103 | ErrCodeTenantAppQuota | 租户应用配额超限 |

### 用户管理 20200–20299

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 20201 | ErrCodeUserNotFound | 用户不存在 |
| 20202 | ErrCodeUserDuplicate | 用户名已存在 |
| 20203 | ErrCodeUserNotInTenant | 用户未关联该租户 |
| 20204 | ErrCodeWrongOldPassword | 旧密码错误 |
| 20205 | ErrCodeLastUserProtected | 不能删除最后一个用户 |

### 角色/权限管理 20300–20399

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 20301 | ErrCodeRoleNotFound | 角色不存在 |
| 20302 | ErrCodeRoleDuplicate | 角色编码已存在 |
| 20303 | ErrCodeRoleHasUsers | 角色下存在用户，无法删除 |
| 20304 | ErrCodeCyclicHierarchy | 角色层级形成循环 |
| 20305 | ErrCodeMaxHierarchyDepth | 超过最大层级深度 |
| 20306 | ErrCodeExceedsParentPerm | 超出父角色权限范围 |

---

## license 插件 3xxxx

### 激活码/会员 30001–30099

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 30001 | ErrCodeCodeNotFound | 激活码不存在或已失效 |
| 30002 | ErrCodeCodeUsed | 激活码已使用 |
| 30003 | ErrCodeCodeExpired | 激活码已过期 |
| 30004 | ErrCodeCodeTypeNotFound | 激活码类型不存在 |
| 30005 | ErrCodeLevelNotFound | 会员等级不存在 |
| 30006 | ErrCodeActivateDeviceLimit | 超过设备绑定数量上限 |
| 30007 | ErrCodeUnbindLimitExceeded | 超过解绑次数上限 |
| 30008 | ErrCodeMembershipNotFound | 会员记录不存在 |

### 订单/支付 30100–30199

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 30101 | ErrCodeOrderNotFound | 订单不存在 |
| 30102 | ErrCodeOrderAlreadyPaid | 有未完成的支付订单 |
| 30103 | ErrCodeOrderClosed | 订单已关闭 |
| 30104 | ErrCodePaymentFailed | 支付失败 |
| 30105 | ErrCodeRefundFailed | 退款失败 |
| 30106 | ErrCodePriceCodeNotFound | 套餐不存在 |
| 30107 | ErrCodePriceCodeDuplicate | 套餐编码已存在 |

---

## 游戏中心 4xxxx（CR12）

### 游戏管理 40001–40049

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 40001 | ErrCodeGameNotFound | 游戏不存在或已下架 |
| 40002 | ErrCodeGameRomExists | ROM 名称已存在（同租户内唯一约束冲突） |
| 40003 | ErrCodeGameRomNameEmpty | rom_name 不能为空 |
| 40004 | ErrCodeGameRomTooLong | rom_name 超过 50 字符 |
| 40005 | ErrCodeGameRomFormat | rom_name 包含非法字符（只允许小写字母/数字/下划线） |

### 分类管理 40050–40069

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 40050 | ErrCodeCategoryNotFound | 游戏分类不存在 |
| 40051 | ErrCodeCategoryNameEmpty | 分类名称不能为空 |

### 公告管理 40070–40089

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 40070 | ErrCodeNoticeNotFound | 公告不存在 |
| 40071 | ErrCodeNoticeTitleEmpty | 公告标题不能为空 |

### Banner 管理 40090–40099

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 40090 | ErrCodeBannerNotFound | Banner 不存在 |
| 40091 | ErrCodeBannerImageURLEmpty | Banner 图片 URL 不能为空 |
| 40092 | ErrCodeBannerLimitExceeded | 已达同时启用 Banner 上限（最多 5 条） |

### ROM 下载/CDN/VIP 40100–40199

| 错误码 | 常量名 | App 端处理 | 描述 |
|--------|--------|-----------|------|
| 40101 | ErrCodeVIPRequired | 跳转会员购买页 | 该游戏需要 VIP 会员权限 |
| 40102 | ErrCodeCDNDomainNotConfigured | 展示"服务不可用" | CDN 域名环境变量未配置 |
| 40103 | ErrCodeCDNSecretNotConfigured | 展示"服务不可用" | CDN 密钥环境变量未配置 |
| 40104 | ErrCodeRomNameInvalid | 展示"游戏标识有误" | rom_name 格式非法 |

---

## 系统/基础设施 5xxxx

| 错误码 | 常量名 | 描述 |
|--------|--------|------|
| 50001 | ErrCodeInternalError | 系统内部错误（兜底，不暴露 DB 细节） |
| 50002 | ErrCodeDBError | 数据库操作失败 |
| 50003 | ErrCodeExternalService | 外部服务调用失败（CDN/微信等） |
| 50004 | ErrCodeNotImplemented | 功能未实现 |

---

## 前端接入指南

### web-admin / web-user

```typescript
// request.ts 响应拦截器已自动提取 error_code 并附到 Error 对象
// catch 块中使用 handleApiError 获取友好提示

import { handleApiError } from '@/utils/error-handler'

try {
  await someApi()
} catch (e: any) {
  const { errorCode, message } = handleApiError(e)
  ElMessage.error(message)
  if (errorCode === 40101) {
    // VIP 限制，跳转购买页
  }
}
```

### App 端（mjarch Java）

```java
// DownloadManager.enqueueJobFromServer 回调的 error 字符串：
//   null          → 成功
//   "VIP_REQUIRED" → error_code=40101，跳转 PurchaseActivity
//   其他字符串      → 网络/CDN错误，展示 Toast

// ROMDownloadResult 提供便捷判断：
//   result.isVipRequired       → 40101
//   result.isCdnNotConfigured  → 40102/40103
//   result.isInvalidRomName    → 40104
//   result.isError()           → 以上任一
```

---

## 扩展规则

1. 新 license 插件子功能：在 3xxxx 段内追加，不复用已有编号
2. 新应用插件（如会员积分、抽奖等）：从 60001 开始，按 6xxxx / 7xxxx 顺延
3. 废弃的错误码：注释说明废弃原因，不删除，不复用
4. 所有错误码统一在 `common/errors/errors.go` 维护，禁止在各 service 中自定义局部错误码常量
