# Task-001-示例任务名 变更日志

| 字段     | 内容                                      |
|----------|-------------------------------------------|
| 文档标题 | _[任务名称]_ 变更日志                     |
| 任务编号 | Task-001                                  |
| CR 编号  | _[CRxx]_                                  |
| 变更时间 | _[YYYY-MM-DD HH:mm]_                      |
| 变更人   | _[开发人员姓名]_                          |
| 版本     | v1.0                                      |
| 关联任务 | task/Task-001-示例任务名.md               |
| 状态     | 草稿 / 已提交 / 已审核                    |

---

## 变更概述

本次变更完成了付款申请模块的数据库表结构创建，新增 `pay_order`（付款申请主表）和 `pay_order_item`（付款明细表），并初始化了付款类型、审批状态等基础字典数据，为后续业务接口开发提供数据层支撑。

---

## 变更文件清单

| 文件路径 | 变更类型 | 变更说明 |
|----------|----------|----------|
| `src/main/java/com/example/payment/entity/PayOrder.java` | 新增 | 付款申请主表实体类，含 MyBatis-Plus 注解 |
| `src/main/java/com/example/payment/entity/PayOrderItem.java` | 新增 | 付款申请明细表实体类 |
| `src/main/java/com/example/payment/mapper/PayOrderMapper.java` | 新增 | 付款申请 Mapper 接口，继承 BaseMapper |
| `src/main/java/com/example/payment/mapper/PayOrderItemMapper.java` | 新增 | 付款明细 Mapper 接口 |
| `src/main/resources/mapper/PayOrderMapper.xml` | 新增 | 付款申请自定义 SQL 映射文件 |
| `src/main/java/com/example/payment/enums/PayOrderStatus.java` | 新增 | 付款申请状态枚举（DRAFT/PENDING/APPROVED/REJECTED/WITHDRAWN） |
| `src/test/java/com/example/payment/mapper/PayOrderMapperTest.java` | 新增 | Mapper 层集成测试，使用 Testcontainers |
| `sql/V1.0.0__create_pay_order.sql` | 新增 | Flyway 迁移脚本，创建 pay_order 和 pay_order_item 表 |
| `sql/V1.0.1__init_pay_dict_data.sql` | 新增 | 初始化付款类型、状态字典数据 |

> **说明**：_[如有其他变更文件，在此补充。删除的文件需说明删除原因及影响范围。]_

---

## SQL 变更记录

### 执行脚本清单

| 脚本文件 | 执行环境 | 执行时间 | 执行人 | 执行结果 |
|----------|----------|----------|--------|----------|
| `V1.0.0__create_pay_order.sql` | DEV | _[YYYY-MM-DD HH:mm]_ | _[执行人]_ | ✅ 成功 |
| `V1.0.1__init_pay_dict_data.sql` | DEV | _[YYYY-MM-DD HH:mm]_ | _[执行人]_ | ✅ 成功 |
| `V1.0.0__create_pay_order.sql` | SIT | _[YYYY-MM-DD HH:mm]_ | _[执行人]_ | _[✅ 成功 / ❌ 失败]_ |
| `V1.0.1__init_pay_dict_data.sql` | SIT | _[YYYY-MM-DD HH:mm]_ | _[执行人]_ | _[✅ 成功 / ❌ 失败]_ |

### DDL 变更说明

```sql
-- 新增表：pay_order（付款申请主表）
-- 字段数：18 个，含 id、order_no、applicant_id、amount、status 等核心字段
-- 索引：主键索引 + order_no 唯一索引 + applicant_id 普通索引 + status 普通索引
CREATE TABLE pay_order ( ... );  -- 详见 sql/sql-001-pay-order-ddl.md

-- 新增表：pay_order_item（付款申请明细表）
-- 字段数：10 个，含 id、order_id、item_name、amount 等字段
-- 索引：主键索引 + order_id 普通索引
CREATE TABLE pay_order_item ( ... );  -- 详见 sql/sql-001-pay-order-ddl.md
```

### DML 变更说明

```sql
-- 初始化字典数据：付款类型（dict_type = 'pay_type'）
-- 新增 4 条记录：日常办公(01)、差旅费用(02)、采购付款(03)、其他(99)
INSERT INTO sys_dict_data ...;  -- 详见 sql/sql-001-pay-order-dml.md

-- 初始化字典数据：付款申请状态（dict_type = 'pay_order_status'）
-- 新增 5 条记录：草稿(0)、审批中(1)、已通过(2)、已拒绝(3)、已撤回(4)
INSERT INTO sys_dict_data ...;  -- 详见 sql/sql-001-pay-order-dml.md
```

### 回滚脚本

如需回滚，执行以下脚本（按逆序执行）：

```sql
-- 回滚 DML：删除初始化字典数据
DELETE FROM sys_dict_data WHERE dict_type IN ('pay_type', 'pay_order_status');

-- 回滚 DDL：删除新增表（注意：会丢失所有数据，生产环境谨慎操作）
DROP TABLE IF EXISTS pay_order_item;
DROP TABLE IF EXISTS pay_order;
```

---

## API 变更记录

> **本任务说明**：Task-001 为数据库表结构创建任务，本次变更不涉及对外暴露的 API 接口。API 接口将在 Task-002（付款申请 CRUD 接口）中实现。

### 新增接口（预告）

以下接口将在后续任务中实现，本次变更完成了其数据层基础：

| 接口路径 | HTTP 方法 | 说明 | 计划任务 |
|----------|-----------|------|----------|
| `/api/v1/payments` | POST | 创建付款申请 | Task-002 |
| `/api/v1/payments` | GET | 分页查询付款申请列表 | Task-002 |
| `/api/v1/payments/{id}` | GET | 查询付款申请详情 | Task-002 |
| `/api/v1/payments/{id}` | PUT | 修改付款申请 | Task-002 |
| `/api/v1/payments/{id}/submit` | POST | 提交审批 | Task-003 |
| `/api/v1/payments/{id}/approve` | POST | 审批通过/拒绝 | Task-003 |

_[如本任务涉及 API 变更，在此填写实际新增/修改的接口信息，并链接到 api/ 目录下的详细文档。]_

---

## 测试结果

### 单元测试

| 测试类 | 测试方法数 | 通过 | 失败 | 跳过 | 覆盖率 |
|--------|-----------|------|------|------|--------|
| `PayOrderMapperTest` | 8 | 8 | 0 | 0 | - |
| `PayOrderStatusTest` | 5 | 5 | 0 | 0 | 100% |

**测试执行命令：**

```bash
mvn test -pl payment-service -Dtest=PayOrderMapperTest,PayOrderStatusTest
```

**测试输出摘要：**

```
Tests run: 13, Failures: 0, Errors: 0, Skipped: 0
BUILD SUCCESS
```

### 手动测试结果

| 测试项 | 测试步骤 | 预期结果 | 实际结果 | 状态 |
|--------|----------|----------|----------|------|
| 数据库表创建验证 | 连接 DEV 数据库，执行 `SHOW TABLES LIKE 'pay_%'` | 返回 `pay_order` 和 `pay_order_item` 两张表 | 符合预期 | ✅ 通过 |
| 表结构验证 | 执行 `DESC pay_order` | 字段与 DDL 文档一致，含 18 个字段 | 符合预期 | ✅ 通过 |
| 索引验证 | 执行 `SHOW INDEX FROM pay_order` | 存在主键、唯一索引（order_no）、普通索引（applicant_id、status） | 符合预期 | ✅ 通过 |
| 字典数据验证 | 查询 `SELECT * FROM sys_dict_data WHERE dict_type IN ('pay_type', 'pay_order_status')` | 返回 9 条初始化数据 | 符合预期 | ✅ 通过 |
| 实体类映射验证 | 启动应用，观察 MyBatis-Plus 日志 | 无映射错误，实体类与表结构匹配 | 符合预期 | ✅ 通过 |

---

## 关联 Git Commit

| Commit Hash | 提交时间 | 提交人 | 提交说明 |
|-------------|----------|--------|----------|
| `a1b2c3d` | _[YYYY-MM-DD HH:mm]_ | _[开发人员]_ | feat(payment): 新增 pay_order 和 pay_order_item 实体类及 Mapper |
| `e4f5g6h` | _[YYYY-MM-DD HH:mm]_ | _[开发人员]_ | feat(payment): 新增付款申请状态枚举 PayOrderStatus |
| `i7j8k9l` | _[YYYY-MM-DD HH:mm]_ | _[开发人员]_ | test(payment): 新增 PayOrderMapper 集成测试 |
| `m1n2o3p` | _[YYYY-MM-DD HH:mm]_ | _[开发人员]_ | sql: 新增 pay_order 建表脚本和字典数据初始化脚本 |

> **分支信息**：`feature/CRxx-payment-module`，已合并至 `develop` 分支（PR #_[编号]_）
