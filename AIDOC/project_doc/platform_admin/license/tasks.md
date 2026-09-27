# License 插件 - 任务列表

## Phase 1: 后端开发（platform_admin/license 插件）

### Task 1: 插件工程搭建
**Scope**: 创建 license 插件工程结构，包含 main.go、handler.go、db.go、models.go、crypto.go
**Constraints**: 
- 遵循 platform_admin 插件开发规范（RegisterDirect 内嵌方式）
- 使用 GORM 作为 ORM
- 使用 JWT 进行认证
- 实现 RSA 加密解密（RSA/ECB/PKCS1Padding，2048位）
**Acceptance**:
- 插件可以成功通过 RegisterDirect 方式加载到 platform_admin
- RSA 加密解密功能正常工作

### Task 2: RSA 加密解密模块
**Scope**: 实现 RSA 加密解密模块（crypto.go），支持请求解密和响应加密
**Constraints**:
- 算法：RSA/ECB/PKCS1Padding
- 密钥长度：2048 位
- 私钥从插件配置中读取
- 请求解密：`rsa.DecryptPKCS1v15(rand.Reader, priv, encryptedData)`
- 响应加密：`rsa.EncryptPKCS1v15(rand.Reader, pub, plainData)`
**Acceptance**:
- 可以正确解密客户端加密的请求参数
- 可以正确加密响应数据供客户端解密
- 与客户端 RSAUtils 兼容

### Task 3: 数据模型和数据库初始化
**Scope**: 定义 activation_code_types、activation_codes、activation_records、user_memberships、user_devices、membership_levels 表模型，创建数据库初始化逻辑
**Constraints**:
- 使用 GORM AutoMigrate
- membership_levels 表插入初始数据（normal/vip/premium）
- activation_code_types 表插入初始数据（standard_yearly、trial）
- 激活码实例（activation_codes）和激活记录（activation_records）分开刻画
**Acceptance**:
- 启动插件时自动创建/迁移数据库表
- membership_levels 表有3条初始记录
- activation_code_types 表有2条初始记录

### Task 4: 设备管理 API
**Scope**: 实现 /devices、/devices/logout API，登录时检查设备数量限制
**Constraints**:
- 复用 user_auth 模块的登录流程
- 登录时检查设备数量限制（最多 N 台，由 license.max_devices 配置项控制，默认 2）
- 支持踢出指定设备
**Acceptance**:
- 同一用户超出 license.max_devices 限制时登录失败
- 可以查看用户设备列表
- 可以踢出指定设备

### Task 5: 激活码类型管理 API
**Scope**: 实现 /code-types（列表、创建）API
**Constraints**:
- 激活码类型包含：名称、代码、默认会员等级、默认时长、逻辑类型
- 初始数据包含标准年费和体验卡两种类型
**Acceptance**:
- 可以创建激活码类型
- 可以获取激活码类型列表

### Task 6: 激活码实例管理 API
**Scope**: 实现 /codes（创建）、/activate、/status、/restore、/activation-history、/cancel API
**Constraints**:
- 激活码一次性使用，重复使用提示"激活码已使用"
- 同一用户多次激活时顺延时间：new_expiry = max(now, old_expiry) + duration_days
- 恢复激活检查激活码归属
- 激活记录按时间倒序排列
- 会员身份可叠加，不同会员类型独立计算过期时间
- 取消激活码时使用 **Replay 重算**方式计算剩余会员时间（不是简单减法）：查询该用户该等级所有仍有效的激活记录（排除被取消的），按 activated_at 升序累加，得出正确的 expires_at
- 每个激活码恢复次数上限为 3 次（activation_codes.restore_count 字段控制），超过限制返回错误"激活码恢复次数已达上限"
- 创建激活码时生成随机字符串（16位，格式 XXXX-XXXX-XXXX）
- 激活码实例可以覆盖类型定义中的属性，未指定则继承
**Acceptance**:
- 可以创建激活码实例，指定类型和覆盖属性
- 有效激活码可以成功激活
- 已使用的激活码不能再次使用，提示"激活码已使用"
- 同一用户多次激活时间正确顺延
- 可以恢复之前用过的激活码
- 激活码恢复次数超过 3 次时返回错误"激活码恢复次数已达上限"
- 用户可以查看自己的激活记录列表
- 用户可同时拥有多个会员类型（叠加）
- /status API 返回用户当前所有有效会员身份列表
- 取消激活码后，用 Replay 算法正确计算会员剩余时间（叠加场景下验证正确性）

### Task 6b: 到期提醒
**Scope**: 实现到期提醒逻辑，在 /status API 响应中新增 `warning_days_left` 字段
**Constraints**:
- 基于用户所有有效会员中最近到期的那个计算剩余天数
- 剩余天数 ≤ 7 时返回具体剩余天数（整数），否则返回 `null`
- 客户端收到 `warning_days_left` 非 null 时，在启动时展示提醒 Toast
**Acceptance**:
- /status 接口在会员剩余 ≤ 7 天时返回 warning_days_left（具体天数）
- /status 接口在会员剩余 > 7 天时返回 warning_days_left: null
- 无有效会员时返回 warning_days_left: null

### Task 7: 会员等级 API
**Scope**: 实现 /levels API
**Constraints**:
- 返回所有会员等级列表
**Acceptance**:
- 可以获取会员等级列表

## Phase 2: 客户端开发（Retro/phoenixui/mjarch）

> **重要**：Phase 2 的所有改动**仅针对 mjarch flavor**。禁止修改其他 flavor（mjguoguan1、tcb1、jdgamespace 等）的任何代码或配置。

### Task 8: MembershipAccessControl 类
**Scope**: 创建 MembershipAccessControl 类，实现 IAccessControl 接口
**Constraints**:
- 检查本地缓存的会员状态
- 过期则弹出登录/激活对话框
- device_id 基于 Android ID 做 SHA-256 哈希生成（首次启动时计算并持久化到 SharedPreferences）
**Acceptance**:
- mjarch flavor 使用新的准入控制逻辑
- 其他 flavor 不受影响

### Task 9: 登录页面
**Scope**: 创建 LoginActivity（手机号+验证码登录）
**Constraints**:
- 调用 /api/v1/user/auth/send-code 发送验证码
- 调用 /api/v1/user/auth/login 登录
- 本地存储 token 和用户信息
**Acceptance**:
- 用户可以登录
- 登录状态持久化

### Task 10: 会员激活页面
**Scope**: 创建 MembershipActivationActivity
**Constraints**:
- 输入激活码
- 调用后端激活 API
- 显示激活结果和会员状态
**Acceptance**:
- 用户可以输入激活码激活会员
- 显示会员等级和过期时间

### Task 11: 会员状态和设备管理页面
**Scope**: 创建 MembershipStatusActivity
**Constraints**:
- 显示当前会员状态
- 显示激活记录列表（调用 /activation-history API）
- 显示已登录设备列表
- 支持踢出设备
**Acceptance**:
- 用户可以查看会员状态
- 用户可以查看激活记录
- 用户可以管理登录设备

### Task 12: mjarch flavor 配置
**Scope**: 修改 mjarch flavor 的 strings.xml，配置新的 access_control_class 和独立 RSA 密钥
**Constraints**:
- 只修改 mjarch flavor，其他 flavor 的 strings.xml 不得改动
- 新增 `license_rsa_key` 字段存放 license 插件专用 RSA 公钥（2048位）
- 不复用、不覆盖旧激活机制的 `key1` 字段
- 旧激活机制的 PHP 后端无需任何改动
**Acceptance**:
- mjarch flavor 使用 MembershipAccessControl
- mjarch flavor 的 `license_rsa_key` 与旧 `key1` 是不同的密钥
- 其他 flavor 使用原来的 ActivationAccessControl，激活功能回归测试通过
- 旧 PHP 后端不受影响，其接口和密钥与 license 插件完全独立

## Phase 3: 测试和部署

### Task 13: 后端单元测试
**Scope**: 编写核心业务逻辑的单元测试
**Constraints**:
- 覆盖激活、顺延、恢复、多设备登录等场景
**Acceptance**:
- 所有测试通过
- 核心逻辑覆盖率 > 80%

### Task 14: 集成测试
**Scope**: 编写端到端集成测试
**Constraints**:
- 覆盖注册 → 登录 → 激活 → 查询状态完整流程
**Acceptance**:
- 所有集成测试通过

### Task 15: 客户端测试
**Scope**: 测试 mjarch flavor 的新激活逻辑，并回归验证其他 flavor 不受影响
**Constraints**:
- 验证其他 flavor 激活逻辑不受影响
- 验证旧 PHP 后端无需改动且仍正常工作
- 验证 mjarch 使用独立 RSA 密钥，不影响旧激活逻辑的 key1
**Acceptance**:
- mjarch flavor 激活流程正常（MembershipAccessControl + license 插件）
- mjguoguan1 / tcb1 / jdgamespace 等其他 flavor 激活功能回归测试全部通过
- 旧 PHP 后端未做任何改动，接口功能正常

### Task 16: 部署
**Scope**: 构建和部署 license 插件
**Constraints**:
- 遵循 platform_admin 插件部署规范
**Acceptance**:
- 插件成功部署到 production
- API 可以正常访问
