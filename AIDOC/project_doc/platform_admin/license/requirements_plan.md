# License 插件 - 需求计划

## 文档说明

本文档记录需求讨论过程中的待确认问题和决策结果。

---

## 已确认决策（第一轮）

### Q1: 游戏列表数据来源 ✅

**决策**：混合模式
- 客户端本地 gamelist 作为基础数据
- 服务端返回 `{ "rom_name": "xxx", "required_levels": ["normal", "vip"] }` 列表
- 客户端按 rom_name 匹配覆盖
- 刷新时机：启动时同步一次，之后每 6 小时自动同步

---

### Q2: 淘宝订单号作为激活码 ✅

**决策**：
- 淘宝订单号 = 激活码
- 复用现有 taobao_order 推送逻辑
- 一个订单可多次激活，`max_activations = 购买数量`
- 每次激活 `used_activations += 1`，达到上限后拒绝

---

### Q3: 套餐价格模块 ✅

**决策**：本次 Phase 2a 实现定价模块

**数据结构**：
```sql
CREATE TABLE price_configs (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  code_type_id BIGINT NOT NULL,
  price INT NOT NULL,
  original_price INT,
  status TINYINT DEFAULT 1,
  sort_order INT DEFAULT 0,
  created_at DATETIME DEFAULT NOW(),
  updated_at DATETIME DEFAULT NOW() ON UPDATE NOW(),
  FOREIGN KEY (code_type_id) REFERENCES activation_code_types(id)
);
```

---

### Q4: 现有 biz_user 数据处理 ✅

**决策**：
- email/password 字段允许 NULL
- 注册接口强制必填
- 不做数据迁移，现有手机号用户数据保留

---

### Q5: 邮箱登录与手机号登录的关系 ✅

**决策**：
- mjarch 只用邮箱登录，phone 字段为空
- 两套登录体系（邮箱/手机号）完全独立

---

## 已确认决策（第二轮）

### Q6: 激活码多次激活 vs 状态字段 ✅

**决策**：
- 两个字段都保留（status 和 used_activations）
- 状态保持为 0（未使用），直到 `used_activations == max_activations` 才变为 1（已用完）
- 判断是否可激活：`status != 3（已取消） && used_activations < max_activations`

---

### Q7: 一个激活码被多个用户使用 ✅

**决策**：激活码可以被不同用户使用
- 每次激活记录一次 `used_activations += 1`
- 不限制必须是同一个用户
- 激活记录（activation_records）中记录是哪个用户激活的

---

### Q8: 激活码恢复 与 多次激活 的关系 ✅

**决策**：
- 恢复是针对**激活记录**，不是针对激活码
- 激活记录中记录：激活用户、生效开始时间、结束时间
- 每个用户能恢复自己的激活记录

---

### Q9: 会员等级叠加逻辑 ✅

**决策**：
- 用户的会员等级是一个**集合**，可以有多个标签等级
- 各等级相互独立，没有高低之分，是不同的权益身份
- 每个等级各自有独立的有效期
- 游戏访问控制：检查用户是否拥有该游戏要求的等级标签之一，且该标签未过期

---

### Q10: 游戏列表服务端数据结构 ✅

**决策**：
- games 表只存储有配置的游戏（需要会员的游戏）
- 不存在于 games 表的游戏默认为免费

---

## 已确认决策（第三轮）

### Q11: 设备登录模式 ✅

**决策**：单设备模式
- 用户同一账号只能在一台设备上登录
- 未登出时，再次在另一设备登录要提示"需要先登出原设备"
- 原设备无法使用时，提示"联系客服处理"

---

### Q12: 会员等级集合的数据结构 ✅ (已更新)

**决策**：新建 user_memberships 表，使用 level_code
```sql
CREATE TABLE user_memberships (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  level_code VARCHAR(20) NOT NULL,     -- 使用 level code 而非 level_id
  expires_at DATETIME NOT NULL,
  created_at DATETIME DEFAULT NOW(),
  updated_at DATETIME DEFAULT NOW() ON UPDATE NOW(),
  FOREIGN KEY (user_id) REFERENCES biz_user(id),
  UNIQUE KEY uk_user_level (user_id, level_code)
);
```

---

### Q13: 游戏等级要求改为多值 ✅

**决策**：
- 游戏的等级要求改为 `required_levels`（code 字符串数组）
- 使用 `MembershipLevel.Code` 字段值（如 "normal", "vip", "premium"）
- 游戏可以配置多个等级 code，用户拥有其中任意一个未过期等级即可玩

---

## 已确认决策（第四轮）

### Q14: games 表存储多等级的方式 ✅

**决策**：方案 A - JSON 字段
```sql
CREATE TABLE games (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  rom_name VARCHAR(50) NOT NULL,
  display_name VARCHAR(100),
  required_levels JSON,                -- 存储 ["normal", "vip"]
  sort_order INT DEFAULT 0,
  status TINYINT DEFAULT 1,
  created_at DATETIME DEFAULT NOW(),
  updated_at DATETIME DEFAULT NOW() ON UPDATE NOW(),
  UNIQUE KEY uk_tenant_rom (tenant_id, rom_name)
);
```

---

### Q15: user_memberships 使用 level_code ✅

**决策**：使用 level_code（字符串）而非 level_id
- 与游戏的 required_levels 保持一致
- 查询时无需 JOIN membership_levels 表

---

### Q16: 登出操作 ✅

**决策**：调用 API 显式登出
- 客户端调用登出 API
- 服务端清除登录状态（Redis 中的设备登录记录）
- 客户端清除本地 token

---

### Q17: 激活码并发控制 ✅

**决策**：使用乐观锁
- activation_codes 表已有 version 字段
- 更新时 `UPDATE ... SET used_activations = used_activations + 1, version = version + 1 WHERE id = ? AND version = ?`
- 更新失败时重试或返回错误

---

## 决策汇总表

| 日期 | 问题 | 决策 |
|------|------|------|
| 2026-10-01 | 密码复杂度 | 不做复杂度要求，长度 ≥ 6 即可 |
| 2026-10-01 | 忘记密码功能 | 不做，让用户联系客服 |
| 2026-10-01 | 用户登录凭证 | 邮箱 + 密码（不用手机号） |
| 2026-10-01 | 淘宝激活码 | 激活码 = 淘宝订单号，可多次激活（按数量） |
| 2026-10-01 | 游戏列表 | 混合模式，客户端本地 + 服务端覆盖 |
| 2026-10-01 | 游戏列表刷新 | 启动时 + 每 6 小时同步 |
| 2026-10-01 | 邮箱查重 | 已存在提示"该邮箱已注册，请更换邮箱或直接登录" |
| 2026-10-01 | 密码必填 | 是，注册时必填，数据库允许 NULL |
| 2026-10-01 | API 路径 | 优先复用 user_auth，否则放 /license/ 下 |
| 2026-10-01 | 多 Pod 缓存 | 使用 Redis |
| 2026-10-01 | 套餐价格 | 本次实现定价模块（关联激活码类型） |
| 2026-10-01 | biz_user 迁移 | 不迁移，字段允许 NULL |
| 2026-10-01 | 登录体系 | mjarch 只用邮箱，两套体系独立 |
| 2026-10-01 | 激活码状态 | 保留 status 和 used_activations，用完才变状态 |
| 2026-10-01 | 激活码多用户 | 可被不同用户使用，激活记录中记录用户 |
| 2026-10-01 | 激活码恢复 | 针对激活记录，每个用户可恢复自己的记录 |
| 2026-10-01 | 会员等级 | 集合模式，多个独立权益标签，各自有效期 |
| 2026-10-01 | games 表 | 只存有配置的游戏，JSON 字段存 required_levels |
| 2026-10-01 | 设备登录 | 单设备模式，需 API 显式登出 |
| 2026-10-01 | 会员数据结构 | user_memberships 表用 level_code |
| 2026-10-01 | 游戏等级要求 | required_levels（code 数组），任一满足即可玩 |
| 2026-10-01 | 激活码并发 | 乐观锁（version 字段） |

---

## 变更记录

| 日期 | 变更内容 |
|------|---------|
| 2026-10-01 | 创建需求计划文档 |
| 2026-10-01 | 三角色 review 后更新 requirements.md / design.md / tasks.md |
| 2026-10-01 | 整理 Q1-Q5 决策 |
| 2026-10-01 | 整理 Q6-Q10 决策 |
| 2026-10-01 | 整理 Q11-Q13 决策 |
| 2026-10-01 | 整理 Q14-Q17 决策，需求计划确认完成 |
