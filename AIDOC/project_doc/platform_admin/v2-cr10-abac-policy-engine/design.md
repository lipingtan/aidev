# 设计：ABAC 策略引擎（行权限 + 列权限）

## 一、整体架构

```
common/auth/abac/
├── model/          # 数据模型（abac_policy / abac_row_policy / abac_col_policy）
├── repository/     # 数据访问层
├── service/        # 业务逻辑（策略 CRUD、缓存管理）
├── handler/        # HTTP Handler（策略管理 + 评估 API）
├── engine/         # 核心引擎
│   ├── registry.go     # 资源属性 SPI 注册表
│   ├── translator.go   # 条件树 → SQL 翻译器
│   ├── merger.go       # 多策略合并（行/列）
│   ├── masker.go       # 列脱敏处理器
│   └── evaluator.go    # 策略评估入口
└── callback.go     # GORM Callback 注册（auth:abac）
```

**请求链路：**
```
HTTP 请求
  → AuthMiddleware（JWT 解析，写入 user_id/role_ids/dept_ids）
  → [业务 Handler]
      → GORM Query
          → auth:tenant_query Callback
          → auth:data_scope Callback
          → auth:abac Callback（行过滤注入）
          → gorm:query
      → AbacSerializer.Apply(ctx, responseData)（列权限处理）
  → HTTP 响应
```

---

## 二、数据库设计

```sql
-- 策略主表
CREATE TABLE abac_policy (
    id            BIGINT       PRIMARY KEY COMMENT '雪花 ID',
    tenant_id     BIGINT       NOT NULL    COMMENT '租户 ID，0=平台级',
    name          VARCHAR(128) NOT NULL    COMMENT '策略名称',
    resource_type VARCHAR(64)  NOT NULL    COMMENT '资源对象标识，如 order/user',
    subject_type  VARCHAR(16)  NOT NULL    COMMENT 'ROLE|PERMISSION_SET|USER|DEPT',
    subject_id    VARCHAR(64)  NOT NULL    COMMENT '主体标识（角色code/用户id/部门id）',
    effect        VARCHAR(8)   NOT NULL    COMMENT 'ALLOW|DENY',
    priority      INT          NOT NULL DEFAULT 100 COMMENT '优先级，越小越优先',
    status        INT          NOT NULL DEFAULT 1   COMMENT '1=启用 0=禁用',
    version       INT          NOT NULL DEFAULT 1   COMMENT '乐观锁版本号',
    description   VARCHAR(512)            COMMENT '策略描述',
    created_at    DATETIME,
    updated_at    DATETIME,
    deleted_at    DATETIME,
    UNIQUE KEY uk_tenant_name (tenant_id, name, deleted_at),
    INDEX idx_resource (tenant_id, resource_type, status)
);

-- 行权限（每条策略可有多个 action 的行权限）
CREATE TABLE abac_row_policy (
    id             BIGINT      PRIMARY KEY COMMENT '雪花 ID',
    policy_id      BIGINT      NOT NULL    COMMENT '关联 abac_policy.id',
    action         VARCHAR(16) NOT NULL    COMMENT 'read|create|update|delete',
    condition_expr JSON                    COMMENT '条件表达式树（JSON）',
    created_at     DATETIME,
    INDEX idx_policy (policy_id)
);

-- 列权限（每条策略可有多个字段配置）
CREATE TABLE abac_col_policy (
    id           BIGINT       PRIMARY KEY COMMENT '雪花 ID',
    policy_id    BIGINT       NOT NULL    COMMENT '关联 abac_policy.id',
    field_name   VARCHAR(64)  NOT NULL    COMMENT '字段逻辑名（对应响应 JSON 的 key）',
    effect       VARCHAR(8)   NOT NULL    COMMENT 'SHOW|HIDE|MASK',
    mask_type    VARCHAR(16)              COMMENT 'phone|email|id_card|custom',
    mask_pattern VARCHAR(256)             COMMENT 'custom 类型的正则规则',
    created_at   DATETIME,
    INDEX idx_policy (policy_id)
);
```

---

## 三、条件表达式树 Schema

叶子节点（`type = "expr"`）：
```json
{
  "type": "expr",
  "left":  {"source": "resource", "attr": "dept_id"},
  "op":    "eq",
  "right": {"source": "subject",  "attr": "dept_ids"}
}
```

组合节点（`type = "group"`）：
```json
{
  "type": "group",
  "operator": "AND",
  "children": [
    {"type": "expr", "left": {...}, "op": "eq", "right": {...}},
    {"type": "group", "operator": "OR", "children": [...]}
  ]
}
```

**source 取值：**
- `resource`：资源属性，`attr` 为注册的逻辑属性名，翻译时替换为 `QualifiedCol`
- `subject`：主体属性，`attr` 为 `user_id / role_ids / dept_ids / tenant_id` 或自定义注册的属性
- `const`：常量，`value` 字段存储字面量值

**两列互比：** 左右值均为 `source: "resource"` 时，翻译为 `table1.col1 = table2.col2`（无参数绑定 `?`）

---

## 四、SPI 接口设计

```go
// common/auth/abac/engine/registry.go

// ResourceAttrDef 资源属性定义
type ResourceAttrDef struct {
    AttrName     string   // 逻辑属性名，在条件表达式中使用
    Display      string   // 显示名，前端编辑器展示
    QualifiedCol string   // 带表限定符的物理列，如 "orders.dept_id"
    DataType     string   // string|int|float|bool
    JoinPath     string   // EXISTS 子查询模板，含 {main_table}/{col}/{op}/{param} 占位符
                          // 为空时直接用 QualifiedCol 生成 WHERE 条件
}

// ResourceDef 资源对象定义
type ResourceDef struct {
    Type        string             // 资源类型标识，如 "order"
    DisplayName string             // 显示名
    MainTable   string             // 主表名（用于 resolveTableName 匹配）
    Attributes  []ResourceAttrDef
}

// RegisterResource 注册资源对象（插件/业务模块启动时调用）
func RegisterResource(def *ResourceDef)

// GetResource 查询已注册的资源定义
func GetResource(resourceType string) (*ResourceDef, bool)

// ListResources 列出所有已注册资源（供前端编辑器加载属性列表）
func ListResources() []*ResourceDef

// SubjectAttrDef 主体属性定义
type SubjectAttrDef struct {
    AttrName string   // 属性名
    Display  string   // 显示名
    DataType string   // string|int|[]int 等
    Resolver func(authInfo *middleware.AuthInfo) interface{}  // 运行时取值函数
}

// RegisterSubjectAttr 注册自定义主体属性（支持扩展）
func RegisterSubjectAttr(def SubjectAttrDef)

// 内置主体属性（默认注册）
// user_id / role_ids / dept_ids / tenant_id
```

---

## 五、API 设计

| 方法 | 路径 | 描述 | 认证 | 权限码 |
|------|------|------|------|--------|
| GET | `/api/v1/admin/abac/policies` | 分页查询策略列表 | admin JWT | `system:abac:list` |
| POST | `/api/v1/admin/abac/policies` | 创建策略 | admin JWT | `system:abac:create` |
| PUT | `/api/v1/admin/abac/policies/:id` | 更新策略 | admin JWT | `system:abac:update` |
| DELETE | `/api/v1/admin/abac/policies/:id` | 删除策略 | admin JWT | `system:abac:delete` |
| GET | `/api/v1/admin/abac/policies/:id` | 查询策略详情 | admin JWT | `system:abac:list` |
| GET | `/api/v1/admin/abac/resources` | 查询已注册资源及其属性列表 | admin JWT | `system:abac:list` |
| POST | `/api/v1/admin/abac/evaluate` | 策略评估（外部调用） | admin JWT（有效 token 即可，无需特定权限码） | — |

**创建策略请求体：**
```json
{
  "name": "销售只能读自己区域订单",
  "resource_type": "order",
  "subject_type": "ROLE",
  "subject_id": "sales",
  "effect": "ALLOW",
  "priority": 100,
  "row_policies": [
    {
      "action": "read",
      "condition_expr": {
        "type": "expr",
        "left":  {"source": "resource", "attr": "region_id"},
        "op":    "eq",
        "right": {"source": "subject",  "attr": "dept_ids"}
      }
    }
  ],
  "col_policies": [
    {"field_name": "phone",    "effect": "MASK", "mask_type": "phone"},
    {"field_name": "id_card",  "effect": "HIDE"}
  ]
}
```

**评估 API 请求体：**
```json
{
  "resource_type": "order",
  "action": "read",
  "subject": {
    "user_id": "123456",
    "role_ids": ["sales", "manager"],
    "dept_ids": ["10", "11"],
    "tenant_id": "1"
  }
}
```

**评估 API 响应体：**
```json
{
  "allowed": true,
  "row_condition": {
    "type": "group",
    "operator": "OR",
    "children": [...]
  },
  "col_effects": {
    "phone":   "MASK",
    "id_card": "HIDE"
  }
}
```

---

## 六、核心逻辑

### 6.1 策略匹配

```
给定请求：(tenant_id, user_id, role_ids, dept_ids, resource_type, action)

1. 从缓存或 DB 加载该 (tenant_id, resource_type) 下的所有启用策略
2. 过滤主体匹配的策略：
   - subject_type=ROLE 且 subject_id 在 role_ids 中
   - subject_type=PERMISSION_SET 且 subject_id 在 role_ids 中（PERMISSION_SET 也存在 role_ids）
   - subject_type=USER 且 subject_id == user_id
   - subject_type=DEPT 且 subject_id 在 dept_ids 中
3. 过滤命中该 action 的行权限（action 匹配或 all）
4. 进入合并流程
```

### 6.2 行权限合并

```
收集所有命中策略的行权限：

DENY 优先：
  如果任一策略 effect=DENY → 注入 WHERE 1=0，终止

ALLOW 合并（OR 语义）：
  将所有 ALLOW 策略的 condition_expr 翻译为 SQL 片段
  多个片段用 OR 拼接：WHERE (cond_A) OR (cond_B) OR (cond_C)

无任何 ALLOW 但有 ABAC 策略存在：
  注入 WHERE 1=0（默认拒绝）

无任何 ABAC 策略：
  不注入（不影响现有查询）
```

### 6.3 条件树翻译

```go
// translator.go 核心逻辑（伪代码）
func Translate(node CondNode, authInfo *AuthInfo, resDef *ResourceDef) (sql string, args []interface{}, err error) {
    switch node.Type {
    case "group":
        parts := []string{}
        for _, child := range node.Children {
            s, a, _ := Translate(child, authInfo, resDef)
            parts = append(parts, "("+s+")")
            args = append(args, a...)
        }
        return strings.Join(parts, " "+node.Operator+" "), args, nil

    case "expr":
        leftCol := resolveLeft(node.Left, resDef)   // 返回 QualifiedCol 或 JoinPath 模板
        rightVal := resolveRight(node.Right, authInfo, resDef)

        // 属性名校验：未注册时安全降级为 WHERE 1=0，记录日志，不暴露 500
        if node.Left.Source == "resource" {
            attr, ok := resDef.GetAttr(node.Left.Attr)
            if !ok {
                log.Printf("[abac] unknown resource attr: %s on resource %s", node.Left.Attr, resDef.Type)
                return "1 = 0", nil, nil  // 安全降级
            }
            leftCol = attr.QualifiedCol
        }

        if node.Left.Source == "resource" && node.Right.Source == "resource" {
            // 两列互比：无参数绑定
            rightCol := resolveLeft(node.Right, resDef)
            return buildColCompare(leftCol, node.Op, rightCol), nil, nil
        }

        if resDef.GetAttr(node.Left.Attr).JoinPath != "" {
            // 关联表属性：生成 EXISTS 子查询
            return buildExistsSubquery(leftCol, node.Op, rightVal, resDef.MainTable), []interface{}{rightVal}, nil
        }

        return buildWhere(leftCol, node.Op), []interface{}{rightVal}, nil
    }
}
```

### 6.4 列权限合并

```
收集所有命中策略的列权限：
对于每个字段，取最严格效果：HIDE > MASK > SHOW

合并结果：map[fieldName]ColEffect
```

### 6.5 列脱敏处理

Handler 主动调用（业务代码一行）：
```go
// Handler 示例
func (h *OrderHandler) List(c *gin.Context) {
    orders := h.svc.ListOrders(c.Request.Context(), ...)
    result := abac.ApplyColPolicy(c.Request.Context(), orders)  // 列权限处理
    handler.Success(c, result)
}
```

`ApplyColPolicy` 内部：
1. 从 context 取 `AbacColContext`（由 GORM Callback 阶段写入，复用同一次策略查询结果）
2. 将 `result` 序列化为 `map[string]interface{}`
3. 按 `col_effects` 遍历字段，执行 HIDE（置 null）或 MASK（调用对应脱敏函数）
4. 返回处理后的数据

---

## 七、缓存设计

**Key 格式：** `abac:{tenant_id}:{user_id}:{resource_type}`

**Value：** 该用户命中的原始策略列表（JSON 序列化，含 row_policies 和 col_policies）

**TTL：** 10 分钟

**主动失效：**
```go
// 策略 CRUD 成功后发布事件
// 注：event.EventAbacPolicyChanged 为新增常量，需在 common/event/bus.go 中添加
event.DefaultBus.Publish(event.EventAbacPolicyChanged, &event.AbacPolicyChangedEvent{
    TenantID:     policy.TenantID,
    ResourceType: policy.ResourceType,
})

// 订阅方批量删缓存
event.DefaultBus.Subscribe(event.EventAbacPolicyChanged, func(payload interface{}) {
    ev := payload.(*event.AbacPolicyChangedEvent)
    prefix := fmt.Sprintf("abac:%d:*:%s", ev.TenantID, ev.ResourceType)
    cache.DeleteByPrefix(prefix)
})
```

**降级：** CacheAdapter 不可用时直接查 DB，不影响功能。

---

## 八、GORM Callback 注册

```go
// callback.go
func RegisterAbacCallback(db *gorm.DB, enabled bool) {
    if !enabled { return }
    _ = db.Callback().Query().
        After("auth:data_scope").
        Before("gorm:query").
        Register("auth:abac", abacQueryCallback)
}

func abacQueryCallback(db *gorm.DB) {
    if db.Statement == nil || db.Statement.Context == nil { return }
    ctx := db.Statement.Context
    if middleware.IsSystemOp(ctx) { return }

    authInfo := middleware.GetAuthInfo(ctx)
    if authInfo == nil { return }

    targetTable := resolveTableName(db)
    if targetTable == "" { return }

    // 通过主表名反查资源类型
    resDef := engine.GetResourceByTable(targetTable)
    if resDef == nil { return }  // 未注册 → 不注入

    // 加载策略（含缓存）
    policies := loadPolicies(ctx, authInfo, resDef.Type)

    // 当前请求 action（从 context 读取，由 Handler 写入）
    action := GetAbacAction(ctx)

    // 合并行权限
    result := engine.MergeRowPolicies(policies, authInfo, action)
    if result.Deny {
        db.Where("1 = 0")
        return
    }
    for _, cond := range result.Conditions {
        db.Where(cond.SQL, cond.Args...)
    }

    // 合并列权限，写入 context 供 ApplyColPolicy 使用
    colEffects := engine.MergeColPolicies(policies)
    SetAbacColContext(ctx, colEffects)
}
```

---

## 九、前端页面设计

**策略列表页（`src/views/system/abac/PolicyList.vue`）：**
- 左侧筛选：resource_type 下拉、subject_type 筛选
- 表格：name / resource_type / subject_type / subject_display_name（可读名称）/ effect / priority / status
- 操作：新增、编辑、删除（二次确认）、启用/禁用

> 列表接口响应需包含 `subject_display_name`（Service 层根据 subject_type 二次查询角色名/用户名/部门名补充），不得只返回裸 subject_id。

**策略编辑器（`src/views/system/abac/PolicyForm.vue`）：**
- 基本信息：名称、资源对象（下拉，从 `/abac/resources` 加载）、主体类型+主体（级联选择）、效果、优先级
- 行权限 Tab：action 多选 + 条件可视化编辑器（递归树形组件，支持添加叶子/组合节点）
- 列权限 Tab：字段名（文本输入）+ 效果下拉（SHOW/HIDE/MASK）+ 脱敏类型（MASK 时出现）

**条件编辑器右值选择逻辑：**
- 左值 source 选 `resource` → 属性名从该资源的注册列表加载
- 右值 source 选 `subject` → 属性名从主体属性列表加载（内置 + SPI 扩展）
- 右值 source 选 `resource` → 同左值属性列表（支持两列互比）
- 右值 source 选 `const` → 自由输入文本框

---

## 十、不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 无 ABAC 策略的资源，查询结果不受任何影响 | 查询未注册资源的接口，返回全量数据 |
| RG-2 | 现有 DataScope 行过滤逻辑不受 ABAC Callback 干扰 | DataScope 单测通过；两者叠加时 AND 语义正确 |
| RG-3 | 现有字段权限（FieldPermission）不受影响 | 字段权限相关接口返回正常 |
| RG-4 | 现有登录/租户选择流程不受影响 | `/auth/login` 和 `/auth/tenant/select` 正常返回 |
| RG-5 | `SystemOpContext` 标记的操作跳过 ABAC 过滤 | 种子数据初始化、系统级查询不被 ABAC 拦截 |
| RG-6 | 策略 CRUD 失败时缓存不失效 | 事务回滚后缓存 key 仍有效 |

---

## 十一、正确性属性

- DENY 效果对任意 action 均优先于 ALLOW，无论 priority 大小
- 同一资源有 ABAC 策略但无命中 ALLOW 时，默认注入 `WHERE 1=0`（Closed World 假设）
- 无 ABAC 策略的资源，Callback 不注入任何 WHERE 条件（不影响现有权限体系）
- 列脱敏原始值不出现在任何响应体中（服务端处理，非前端标记）
- 缓存失效在事务提交成功后触发（写操作成功后才失效，失败不失效）
- 策略按 `tenant_id` 隔离，`tenant_id=0` 平台级策略覆盖所有租户
- `SystemOpContext` 标记的 context 跳过所有 ABAC 注入
