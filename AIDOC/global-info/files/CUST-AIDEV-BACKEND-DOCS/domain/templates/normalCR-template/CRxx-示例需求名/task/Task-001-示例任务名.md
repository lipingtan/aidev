# Task-001-示例任务名

| 字段     | 内容                                      |
|----------|-------------------------------------------|
| 文档标题 | Task-001 _[任务名称]_                     |
| CR 编号  | _[CRxx]_                                  |
| 任务编号 | Task-001                                  |
| 负责人   | _[开发人员姓名]_                          |
| 审核人   | _[技术负责人姓名]_                        |
| 并行组   | A                                         |
| 依赖任务 | -（无前置依赖）                           |
| 预估工时 | _[X]_ 小时                               |
| 创建日期 | _[YYYY-MM-DD]_                            |
| 最后更新 | _[YYYY-MM-DD]_                            |
| 状态     | 待开始 / 进行中 / 已完成 / 已阻塞        |

---

## 任务目标

创建 `pay_order`（付款申请表）及相关字典数据的数据库脚本，为后续业务接口开发提供数据存储基础。

_[替换为本任务的一句话目标描述，说明要实现什么功能、产出什么结果。]_

---

## 输入文档引用

开发本任务前，请仔细阅读以下文档：

| 文档 | 路径 | 关注重点 |
|------|------|----------|
| 正式需求文档 | [requirements/requirements.md](../requirements/requirements.md) | 功能需求 1、2，了解业务字段含义和约束条件 |
| 技术设计文档 | [design/design.md](../design/design.md) | 数据库设计引用、状态机设计、风险与约束章节 |
| DDL 脚本文档 | [sql/sql-001-示例表名-ddl.md](../sql/sql-001-示例表名-ddl.md) | 建表语句、索引设计、回滚脚本 |
| DML 脚本文档 | [sql/sql-001-示例表名-dml.md](../sql/sql-001-示例表名-dml.md) | 初始化字典数据、枚举值说明 |

---

## 实现范围

### Controller 层

本任务为纯数据库层任务，**不涉及 Controller 层改动**。

_[若任务涉及 Controller 层，在此列出需要新增或修改的接口，格式如下：]_

| 操作类型 | 接口路径 | HTTP 方法 | 说明 |
|----------|----------|-----------|------|
| _[新增]_ | _[/api/v1/xxx]_ | _[POST]_ | _[接口功能说明]_ |
| _[修改]_ | _[/api/v1/xxx/{id}]_ | _[PUT]_ | _[修改内容说明]_ |

---

### Service 层

本任务为纯数据库层任务，**不涉及 Service 层改动**。

_[若任务涉及 Service 层，在此描述需要实现的业务逻辑，格式如下：]_

| 方法名 | 说明 | 关键逻辑 |
|--------|------|----------|
| _[createPayment]_ | _[创建付款申请]_ | _[参数校验 → 生成编号 → 插入记录]_ |
| _[submitForApproval]_ | _[提交审批]_ | _[状态校验 → 创建审批任务 → 更新状态]_ |

---

### Mapper 层

本任务的核心工作在数据库层，具体实现内容如下：

#### 1. 执行 DDL 脚本

在各环境数据库中执行 `sql/sql-001-示例表名-ddl.md` 中的建表语句，创建以下表：

| 表名 | 说明 | 执行顺序 |
|------|------|----------|
| `pay_order` | 付款申请主表 | 1 |
| _[table_name_2]_ | _[表说明]_ | _[2]_ |

#### 2. 执行 DML 脚本

在各环境数据库中执行 `sql/sql-001-示例表名-dml.md` 中的初始化数据脚本，插入以下字典数据：

| 字典类型 | 字典项 | 说明 |
|----------|--------|------|
| `pay_type`（付款类型） | 0-普通付款、1-预付款、2-报销 | 付款申请类型枚举 |
| `pay_status`（申请状态） | 0-草稿、1-审批中、2-已通过、3-已拒绝、4-已撤回 | 付款申请状态枚举 |

#### 3. 创建 MyBatis-Plus 实体类和 Mapper

在项目代码中创建以下文件：

```
src/main/java/com/example/
├── entity/
│   └── PayOrder.java          # 付款申请实体类（对应 pay_order 表）
├── mapper/
│   └── PayOrderMapper.java    # MyBatis-Plus Mapper 接口
└── mapper/xml/
    └── PayOrderMapper.xml     # 自定义 SQL（如有复杂查询）
```

实体类示例：

```java
@Data
@TableName("pay_order")
public class PayOrder implements Serializable {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String orderNo;

    private Long applicantId;

    private Long deptId;

    private BigDecimal amount;

    private String currency;

    private Integer payType;

    private Integer status;

    private Long approveId;

    private LocalDateTime approveTime;

    private String approveRemark;

    private String remark;

    @TableField(fill = FieldFill.INSERT)
    private Long createBy;

    @TableField(fill = FieldFill.INSERT_UPDATE)
    private Long updateBy;

    @TableField(fill = FieldFill.INSERT)
    private LocalDateTime createTime;

    @TableField(fill = FieldFill.INSERT_UPDATE)
    private LocalDateTime updateTime;

    @TableLogic
    private Integer delFlag;
}
```

---

## 验收标准

以下验收条件均需通过，方可标记本任务为已完成：

1. **DDL 执行成功**：在 DEV 环境执行建表脚本后，`SHOW CREATE TABLE pay_order` 输出与设计文档一致，包含所有字段、索引和注释。
2. **DML 执行成功**：字典数据初始化完成，查询 `SELECT * FROM sys_dict_data WHERE dict_type IN ('pay_type', 'pay_status')` 返回预期的枚举值记录。
3. **唯一索引生效**：尝试插入重复 `order_no` 时，数据库抛出唯一键冲突错误（`Duplicate entry`）。
4. **逻辑删除生效**：将 `del_flag` 设为 1 后，通过 MyBatis-Plus 的 `selectById` 查询不到该记录（自动过滤已删除数据）。
5. **实体类映射正确**：通过 MyBatis-Plus 的 `insert` 方法插入一条测试记录，`create_time` 和 `update_time` 自动填充，`del_flag` 默认为 0。
6. **回滚脚本可用**：在 DEV 环境验证回滚脚本（`DROP TABLE IF EXISTS pay_order`）可正常执行，执行后表不存在。

---

## 注意事项

1. **执行顺序**：DDL 脚本必须先于 DML 脚本执行，且必须先于任何业务代码部署。
2. **字符集**：建表时必须指定 `CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`，避免中文乱码和 emoji 存储问题。
3. **生产环境执行**：生产环境执行 DDL 前，必须获得 DBA 审批，并在业务低峰期（如凌晨 2:00-4:00）执行，执行前通知相关人员。
4. **`IF NOT EXISTS` 保护**：建表语句已包含 `CREATE TABLE IF NOT EXISTS`，重复执行不会报错，但需人工确认表结构是否与预期一致。
5. **索引命名规范**：唯一索引以 `uk_` 前缀命名，普通索引以 `idx_` 前缀命名，命名格式为 `前缀_表名_字段名`。
6. **实体类注意事项**：`delFlag` 字段必须添加 `@TableLogic` 注解，确保 MyBatis-Plus 自动处理逻辑删除；`createTime` 和 `updateTime` 通过 `MetaObjectHandler` 自动填充，不需要在业务代码中手动赋值。

---

## 并行性说明

| 项目 | 说明 |
|------|------|
| 并行组 | A（基础设施层） |
| 是否可并行 | 否，本任务是所有其他任务的前置依赖，必须最先完成 |
| 前置依赖 | 无 |
| 后置任务 | Task-002（付款申请 CRUD 接口）、Task-003（审批流程接口）、Task-004（权限控制） |
| 阻塞风险 | 本任务延期将导致并行组 B 的所有任务无法启动，需优先保障本任务按时完成 |

> **建议**：本任务预计工时较短（_[X]_ 小时），建议在 CR 启动第一天完成，以便尽早解锁后续并行任务，最大化团队并行效率。
