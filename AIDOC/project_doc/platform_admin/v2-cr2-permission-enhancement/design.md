# 设计：V2-CR2 权限体系增强

## 技术方案

### API 设计

#### Phase 1: 角色继承

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/roles/:id/assignable-resources | 获取角色可分配给子角色的资源范围 | JWT + 租户 |
| GET | /api/v1/admin/roles/:id/assignable-apis | 获取角色可分配给子角色的 API 范围 | JWT + 租户 |

> AssignResources (`PUT /roles/:id/resources`) 和 AssignApis (`PUT /roles/:id/apis`) 为已有接口，增加内部校验逻辑，API 签名不变。响应增加 `affected_children` 字段。

#### Phase 2: 权限集

无新增接口。角色 CRUD 已有接口通过 `role_type=PERMISSION_SET` 参数区分。角色列表接口增加 `role_type` 过滤参数。

| 方法 | 路径 | 描述 | 变更 |
|------|------|------|------|
| GET | /api/v1/admin/roles?role_type=PERMISSION_SET | 按类型过滤角色列表 | 新增 query param |

#### Phase 3: 字段权限

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/admin/field-objects | 获取已注册的字段对象列表 | JWT + 租户 |
| GET | /api/v1/admin/field-objects/:objectCode/fields | 获取对象的字段定义列表 | JWT + 租户 |
| PUT | /api/v1/admin/field-objects/:objectCode/fields/:fieldName | 修改字段描述 | JWT + 租户 |
| POST | /api/v1/admin/field-objects | 手动注册字段对象 | JWT + 租户 |
| POST | /api/v1/admin/field-objects/:objectCode/fields | 手动注册字段 | JWT + 租户 |
| GET | /api/v1/admin/field-permissions?role_id=X&object_code=Y | 查询角色字段权限配置 | JWT + 租户 |
| PUT | /api/v1/admin/field-permissions | 批量设置角色字段权限 | JWT + 租户 |
| DELETE | /api/v1/admin/field-permissions/:id | 删除字段权限配置 | JWT + 租户 |

#### Phase 4: 记录共享

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /api/v1/admin/record-shares | 创建共享规则 | JWT + 租户 |
| GET | /api/v1/admin/record-shares?object_code=X&record_id=Y | 查询记录的共享规则 | JWT + 租户 |
| DELETE | /api/v1/admin/record-shares/:id | 删除共享规则 | JWT + 租户 |

### 数据库设计

```sql
-- Phase 3: 字段对象元数据
CREATE TABLE admin_field_object (
    id          BIGINT PRIMARY KEY,
    object_code VARCHAR(64) NOT NULL COMMENT '业务对象标识',
    object_name VARCHAR(128) NOT NULL COMMENT '对象显示名（如"发票"）',
    source      VARCHAR(16) NOT NULL DEFAULT 'AUTO' COMMENT '来源: AUTO/MANUAL',
    app_code    VARCHAR(64) NULL COMMENT '所属应用（插件业务表时填写）',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_object_code (object_code)
) COMMENT '字段权限对象注册表';

CREATE TABLE admin_field_definition (
    id            BIGINT PRIMARY KEY,
    object_code   VARCHAR(64) NOT NULL COMMENT '业务对象标识',
    field_name    VARCHAR(64) NOT NULL COMMENT '字段名（对应 JSON key）',
    description   VARCHAR(128) NOT NULL COMMENT '字段描述（中文）',
    source        VARCHAR(16) NOT NULL DEFAULT 'AUTO' COMMENT '来源: AUTO/MANUAL',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_object_field (object_code, field_name)
) COMMENT '字段权限字段定义表';

-- Phase 3: 字段权限配置
CREATE TABLE admin_field_permission (
    id          BIGINT PRIMARY KEY,
    role_id     BIGINT NOT NULL,
    object_code VARCHAR(64) NOT NULL COMMENT '业务对象标识',
    field_name  VARCHAR(64) NOT NULL COMMENT '字段名（对应 JSON key）',
    access      VARCHAR(16) NOT NULL DEFAULT 'VISIBLE' COMMENT 'VISIBLE/EDITABLE/HIDDEN',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_object_field (role_id, object_code, field_name)
) COMMENT '角色字段权限配置';

-- Phase 4: 记录共享
CREATE TABLE admin_record_share (
    id            BIGINT PRIMARY KEY,
    tenant_id     BIGINT NOT NULL,
    object_code   VARCHAR(64) NOT NULL COMMENT '业务对象',
    record_id     BIGINT NOT NULL COMMENT '被共享的记录 ID',
    share_to_type VARCHAR(16) NOT NULL COMMENT 'USER/ROLE/DEPT',
    share_to_id   BIGINT NOT NULL COMMENT '共享目标 ID',
    access_level  VARCHAR(16) NOT NULL DEFAULT 'READ' COMMENT 'READ/EDIT',
    expire_at     DATETIME NULL COMMENT '过期时间（NULL=永久）',
    created_by    BIGINT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_record (tenant_id, object_code, record_id),
    INDEX idx_share_to (share_to_type, share_to_id, tenant_id)
) COMMENT '记录共享规则';
```

### 核心逻辑

#### 1. 角色继承 — AssignResources 子集校验 + 级联裁剪

```go
func (s *RoleService) AssignResources(roleID int64, req *AssignResourcesRequest) (*AssignResult, error) {
    role, _ := s.roleRepo.FindByID(s.db, roleID)

    // PERMISSION_SET 不做子集校验
    if role.RoleType != "PERMISSION_SET" && role.ParentID != nil {
        parentResIDs := s.getParentResourceIDs(*role.ParentID)
        if !isSubset(req.ResourceIDs, parentResIDs) {
            return nil, ErrExceedsParentPermission
        }
    }

    // 执行全量替换（现有逻辑）
    s.replaceRoleResources(roleID, req.ResourceIDs)

    // 级联裁剪子角色
    affected := s.cascadeTrimChildren(roleID, req.ResourceIDs)

    // 操作日志
    for _, child := range affected {
        s.logger.Log(operatorID, "cascade_trim_permissions", "role", child.RoleID,
            fmt.Sprintf("级联裁剪: 移除 %d 项资源", child.RemovedCount))
    }

    // 缓存失效：收集受影响角色关联的用户，批量清除权限缓存
    if len(affected) > 0 {
        affectedRoleIDs := make([]int64, 0, len(affected))
        for _, child := range affected {
            affectedRoleIDs = append(affectedRoleIDs, child.RoleID)
        }
        s.invalidatePermCacheByRoles(affectedRoleIDs)
    }

    return &AssignResult{AffectedChildren: affected}, nil
}
```

#### 2. 权限计算合并（含 Permission Set）

```go
func (s *PermissionService) GetUserPermCodes(userID, tenantID int64) []string {
    roles := s.getUserRoles(userID, tenantID) // 含 PERMISSION_SET

    var allCodes []string
    for _, ur := range roles {
        codes := s.getRolePermCodes(ur.RoleID)
        allCodes = append(allCodes, codes...)
    }
    return deduplicate(allCodes)
}

// GetUserMenu 同样合并 PERMISSION_SET 的 resourceIDs
func (s *ResourceService) GetUserMenu(...) {
    // 查用户所有角色（含 PERMISSION_SET）的 resourceIDs 合集
    var allResourceIDs []int64
    for _, roleID := range roleIDs {
        resIDs := s.getRoleResourceIDs(roleID)
        allResourceIDs = append(allResourceIDs, resIDs...)
    }
    allResourceIDs = uniqueInt64(allResourceIDs)
    // ... 后续过滤逻辑不变
}
```

#### 3. 字段过滤 — Gin Middleware 自动拦截 + 手动排除

```go
// FieldFilterMiddleware 自动拦截 JSON 响应，按路由注册的 objectCode 过滤字段
func FieldFilterMiddleware(registry *FieldObjectRegistry, permRepo FieldPermissionRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 检查是否被手动标记为跳过
        if c.GetBool("skip_field_filter") {
            return
        }

        // 根据路由路径匹配 objectCode（从 registry 的路由映射表中查找）
        objectCode := registry.MatchRoute(c.Request.Method, c.FullPath())
        if objectCode == "" {
            return
        }

        // 获取用户角色
        authCtx := GetAuthContext(c)
        if authCtx == nil {
            return
        }

        // 查询字段权限配置
        perms := permRepo.GetByRolesAndObject(authCtx.Roles, objectCode)
        if len(perms) == 0 {
            return // 未配置则不过滤
        }

        // 拦截响应体，移除 HIDDEN 字段
        // 冲突解决：多角色取最高权限（EDITABLE > VISIBLE > HIDDEN）
        filterResponseBody(c, perms)
    }
}

// 手动跳过：在 Handler 中调用
func (h *SomeHandler) TreeEndpoint(c *gin.Context) {
    c.Set("skip_field_filter", true) // 树形接口跳过字段过滤
    // ...
}
```

**路由-对象映射注册：**
```go
// 启动时注册路由与 objectCode 的映射
registry.RegisterRoute("GET", "/api/v1/admin/users", "user")
registry.RegisterRoute("GET", "/api/v1/admin/users/:id", "user")
registry.RegisterRoute("GET", "/api/v1/admin/tenants", "tenant")
// ... 从 admin_field_object 表中加载
```

#### 4. 记录共享 — DataScopeCallback 扩展

```go
// Service 层注入 object_code
func (s *InvoiceService) List(ctx context.Context) ([]Invoice, error) {
    ctx = WithObjectCode(ctx, "invoice")
    db := s.db.WithContext(ctx)
    var list []Invoice
    return list, db.Find(&list).Error
}

// DataScopeCallback 扩展 OR 共享规则
func DataScopeCallback(db *gorm.DB) {
    ctx := db.Statement.Context
    // 现有数据权限逻辑...

    // 记录共享扩展
    objectCode := GetObjectCode(ctx)
    if objectCode == "" {
        return
    }

    authCtx := GetAuthContextFromCtx(ctx)
    if authCtx == nil {
        return
    }

    // 构建共享规则子查询
    shareSubQuery := db.Session(&gorm.Session{NewDB: true}).
        Model(&model.RecordShare{}).
        Select("record_id").
        Where("tenant_id = ? AND object_code = ?", authCtx.TenantID, objectCode).
        Where("(share_to_type = 'USER' AND share_to_id = ?) OR "+
              "(share_to_type = 'ROLE' AND share_to_id IN ?) OR "+
              "(share_to_type = 'DEPT' AND share_to_id IN ?)",
              authCtx.UserID, authCtx.Roles, authCtx.DeptIDs).
        Where("expire_at IS NULL OR expire_at > NOW()")

    // 将原有 WHERE 条件包裹为 OR 共享命中
    db.Where("id IN (?)", shareSubQuery)
}
```

#### 5. 字段对象自动注册

```go
// struct tag 示例
type User struct {
    Phone    string `json:"phone" fieldperm:"手机号"`
    Email    string `json:"email" fieldperm:"邮箱"`
    RealName string `json:"real_name" fieldperm:"真实姓名"`
    Password string `json:"password"` // 无 fieldperm tag，不注册
}

// 启动时扫描注册
func (r *FieldObjectRegistry) AutoRegister(objectCode string, model interface{}) {
    t := reflect.TypeOf(model)
    if t.Kind() == reflect.Ptr { t = t.Elem() }

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        desc := field.Tag.Get("fieldperm")
        if desc == "" || desc == "-" {
            continue
        }
        jsonTag := parseJSONFieldName(field.Tag.Get("json"))
        if jsonTag == "" || jsonTag == "-" {
            continue
        }
        // 写入数据库（INSERT IGNORE，不覆盖已有的自定义描述）
        r.upsertFieldDefinition(objectCode, jsonTag, desc)
    }
}
```

#### 6. 级联裁剪响应 + 前端确认流程

```
前端 PUT /roles/:id/resources {resource_ids: [...]}
  │
  ├── 后端校验子集关系
  ├── 执行替换 + 级联裁剪
  ├── 响应: {
  │     "affected_children": [
  │       {"role_id": "123", "role_name": "销售组长", "removed_count": 3},
  │       {"role_id": "456", "role_name": "销售员", "removed_count": 5}
  │     ]
  │   }
  │
  └── 前端收到响应后，如 affected_children 非空，
      Toast 提示: "已影响 2 个子角色，共移除 8 项权限"
```

> 注：当前设计为"先执行后告知"（非二次确认），因为父角色缩减权限是明确的管理意图。如需二次确认可改为两步接口（preview + confirm），但增加复杂度，CR-2 暂不实现。

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有 AssignResources 对顶级角色（parent_id=NULL）的行为不变 | 顶级角色分配资源正常执行 |
| RG-2 | 现有 AssignApis 对顶级角色的行为不变 | 顶级角色分配 API 正常执行 |
| RG-3 | 现有 GetUserMenu 在未配置字段权限时响应不变 | 无 FieldFilter 配置时 JSON 字段完整 |
| RG-4 | 现有 DataScopeCallback 在无记录共享时行为不变 | 无 object_code context 时不注入共享子查询 |
| RG-5 | 现有角色 CRUD 不受 PERMISSION_SET 类型影响 | 创建/编辑/删除普通角色正常 |
| RG-6 | 现有用户-角色分配流程不变 | 分配普通角色时无额外校验 |
| RG-7 | 权限缓存在无 PERMISSION_SET 时计算结果不变 | 用户无权限集时权限集合与 CR-1 一致 |

## 正确性属性

- 子角色权限 ⊆ 父角色权限（任何时刻，对非 PERMISSION_SET 角色成立）
- 级联裁剪后：所有子角色权限 ⊆ 新的父角色权限
- 用户最终权限 = ∪(普通角色权限) ∪ ∪(权限集权限)
- HIDDEN 字段在 API 响应中不可见（服务端过滤，前端无法绕过）
- 共享规则过期后立即失效（基于数据库 NOW() 比较）
- 字段权限冲突时取最高权限：EDITABLE > VISIBLE > HIDDEN
- 未注册的 object_code 不参与字段过滤和记录共享（零配置无感知）
