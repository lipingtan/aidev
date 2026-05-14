# sql-001-示例表名-ddl

| 项目       | 内容                                      |
|------------|-------------------------------------------|
| 文档名称   | _[表名]-DDL 脚本_                         |
| CR 编号    | _[CRxx]_                                  |
| 版本       | v1.0                                      |
| 执行环境   | DEV / UAT / PROD                          |
| 执行顺序   | 1（DDL 优先于 DML 执行）                  |
| 执行人     | _[执行人姓名]_                            |
| 审核人     | _[审核人姓名]_                            |
| 创建日期   | _[YYYY-MM-DD]_                            |
| 最后更新   | _[YYYY-MM-DD]_                            |

---

## 执行环境说明

| 环境 | 数据库地址 | 注意事项 |
|------|-----------|----------|
| DEV  | _[dev-db-host:port/db_name]_ | 开发环境，可直接执行，执行后通知团队 |
| UAT  | _[uat-db-host:port/db_name]_ | 测试环境，需在发版前 1 个工作日执行，执行后通知测试人员 |
| PROD | _[prod-db-host:port/db_name]_ | 生产环境，需在发版窗口期内执行，执行前必须完成审批流程 |

> **重要提示**：生产环境执行前必须在 UAT 环境验证通过，并获得 DBA 和项目负责人审批。

---

## DDL 脚本

> 说明：以下为付款申请表（`pay_order`）的建表脚本示例，实际使用时请替换为对应业务表结构。

```sql
-- ============================================================
-- 表名：pay_order（付款申请表）
-- CR编号：CRxx
-- 创建日期：YYYY-MM-DD
-- 说明：记录付款申请的主要信息，包含申请人、金额、审批状态等
-- ============================================================

CREATE TABLE IF NOT EXISTS `pay_order` (
  `id`            bigint        NOT NULL AUTO_INCREMENT                                          COMMENT '主键ID',
  `order_no`      varchar(32)   NOT NULL                                                         COMMENT '付款申请编号，格式：PAY+yyyyMMdd+6位序号',
  `applicant_id`  bigint        NOT NULL                                                         COMMENT '申请人ID，关联 sys_user.id',
  `dept_id`       bigint        NOT NULL                                                         COMMENT '申请部门ID，关联 sys_dept.id',
  `amount`        decimal(18,2) NOT NULL                                                         COMMENT '付款金额，单位：元，精确到分',
  `currency`      varchar(10)   NOT NULL DEFAULT 'CNY'                                           COMMENT '币种，默认人民币（CNY）',
  `pay_type`      tinyint       NOT NULL DEFAULT 0                                               COMMENT '付款类型：0-普通付款，1-预付款，2-报销',
  `status`        tinyint       NOT NULL DEFAULT 0                                               COMMENT '状态：0-草稿，1-审批中，2-已通过，3-已拒绝，4-已撤回',
  `approve_id`    bigint                 DEFAULT NULL                                            COMMENT '审批人ID，关联 sys_user.id',
  `approve_time`  datetime               DEFAULT NULL                                            COMMENT '审批时间',
  `approve_remark` varchar(500)          DEFAULT NULL                                            COMMENT '审批意见',
  `remark`        varchar(500)           DEFAULT NULL                                            COMMENT '申请备注',
  `create_by`     bigint        NOT NULL                                                         COMMENT '创建人ID',
  `update_by`     bigint        NOT NULL                                                         COMMENT '最后更新人ID',
  `create_time`   datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP                               COMMENT '创建时间',
  `update_time`   datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP   COMMENT '更新时间',
  `del_flag`      tinyint       NOT NULL DEFAULT 0                                               COMMENT '删除标志：0-未删除，1-已删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pay_order_order_no` (`order_no`),
  KEY `idx_pay_order_applicant_id` (`applicant_id`),
  KEY `idx_pay_order_dept_id` (`dept_id`),
  KEY `idx_pay_order_status` (`status`),
  KEY `idx_pay_order_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='付款申请表';
```

---

## 回滚脚本

> **警告**：回滚脚本将删除整张表及其所有数据，执行前请确认已备份数据。

```sql
-- ============================================================
-- 回滚脚本：删除 pay_order 表
-- 执行条件：仅在 DDL 执行出错或需要回退时使用
-- ============================================================

DROP TABLE IF EXISTS `pay_order`;
```

---

## 注意事项

### 执行前检查

1. **确认数据库版本**：本脚本适用于 MySQL 5.7+ 或 MySQL 8.0+，执行前请确认数据库版本。
2. **确认表不存在**：脚本使用 `CREATE TABLE IF NOT EXISTS`，若表已存在则跳过，不会报错，但需人工确认表结构是否一致。
3. **确认字符集**：数据库默认字符集应为 `utf8mb4`，否则需调整建表语句中的 `CHARSET` 参数。
4. **权限确认**：执行账号需具备目标数据库的 `CREATE` 权限。

### 影响范围

| 影响项 | 说明 |
|--------|------|
| 新增表 | `pay_order`（付款申请表） |
| 影响服务 | _[列出依赖此表的微服务名称]_ |
| 影响接口 | _[列出相关 API 接口路径]_ |
| 数据迁移 | 无（新建表，无历史数据迁移） |

### 索引说明

| 索引名 | 类型 | 字段 | 说明 |
|--------|------|------|------|
| `PRIMARY` | 主键索引 | `id` | 自增主键，唯一标识每条记录 |
| `uk_pay_order_order_no` | 唯一索引 | `order_no` | 保证付款申请编号全局唯一 |
| `idx_pay_order_applicant_id` | 普通索引 | `applicant_id` | 按申请人查询时使用 |
| `idx_pay_order_dept_id` | 普通索引 | `dept_id` | 按部门查询时使用 |
| `idx_pay_order_status` | 普通索引 | `status` | 按状态筛选时使用 |
| `idx_pay_order_create_time` | 普通索引 | `create_time` | 按时间范围查询时使用 |

### 审计字段规范

本表包含以下标准审计字段，所有业务表必须包含：

| 字段 | 类型 | 说明 |
|------|------|------|
| `create_by` | bigint | 创建人用户ID |
| `update_by` | bigint | 最后更新人用户ID |
| `create_time` | datetime | 记录创建时间，自动填充 |
| `update_time` | datetime | 记录最后更新时间，自动更新 |
| `del_flag` | tinyint | 逻辑删除标志，0=未删除，1=已删除 |
