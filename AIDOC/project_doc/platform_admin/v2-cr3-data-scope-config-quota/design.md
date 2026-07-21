# 设计：数据权限增强 + 三级配置 + 配额管理（V2-CR3）

## 技术方案

### 数据库设计

```sql
-- ==================== 组织架构节点表 ====================
CREATE TABLE admin_org_unit (
    id          BIGINT PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    parent_id   BIGINT NULL COMMENT 'NULL=顶级节点',
    node_type   VARCHAR(32) NOT NULL DEFAULT 'DEPARTMENT' COMMENT 'COMPANY/BRANCH/DEPARTMENT/GROUP/TEAM',
    name        VARCHAR(128) NOT NULL,
    code        VARCHAR(64) NULL COMMENT '组织编码（租户内唯一）',
    sort_order  INT NOT NULL DEFAULT 0,
    status      TINYINT NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用',
    version     INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_tenant_code (tenant_id, code),
    INDEX idx_tenant_parent (tenant_id, parent_id)
);

-- ==================== 用户-组织关联表 ====================
CREATE TABLE admin_user_org (
    id          BIGINT PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    org_unit_id BIGINT NOT NULL,
    tenant_id   BIGINT NOT NULL,
    is_primary  TINYINT NOT NULL DEFAULT 0 COMMENT '1=主归属',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_org (user_id, org_unit_id, tenant_id),
    INDEX idx_org_unit (org_unit_id, tenant_id)
);

-- ==================== admin_data_scope 增加 scope_type ====================
ALTER TABLE admin_data_scope
    ADD COLUMN scope_type VARCHAR(16) NOT NULL DEFAULT 'CUSTOM' COMMENT 'ALL/SELF/DEPT/DEPT_TREE/CUSTOM'
    AFTER dimension_name;

-- ==================== admin_data_scope_config 增加 supported_scope_types ====================
ALTER TABLE admin_data_scope_config
    ADD COLUMN supported_scope_types JSON NOT NULL DEFAULT '["ALL","SELF","CUSTOM"]' COMMENT '该维度支持的 scope_type 列表'
    AFTER table_column;

-- ==================== 三级配置表 ====================
CREATE TABLE admin_config (
    id              BIGINT PRIMARY KEY,
    config_key      VARCHAR(128) NOT NULL,
    config_value    TEXT NULL,
    config_type     VARCHAR(32) NOT NULL DEFAULT 'string' COMMENT 'string/number/boolean/json',
    scope           VARCHAR(16) NOT NULL DEFAULT 'SYSTEM' COMMENT 'SYSTEM/TENANT/USER',
    scope_id        BIGINT NOT NULL DEFAULT 0 COMMENT 'SYSTEM=0, TENANT=tenantId, USER=userId',
    tenant_id       BIGINT NOT NULL DEFAULT 0 COMMENT '所属租户(SYSTEM=0)',
    display_name    VARCHAR(128) NULL,
    description     VARCHAR(512) NULL,
    is_feature_flag TINYINT NOT NULL DEFAULT 0 COMMENT '1=功能开关',
    status          TINYINT NOT NULL DEFAULT 1,
    deleted_at      DATETIME NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_scope_key (scope, scope_id, tenant_id, config_key)
);
```

### API 设计

| 方法 | 路径 | 描述 | 认证 | permission_code |
|------|------|------|------|-----------------|
| GET | /api/v1/admin/org-units/tree | 获取组织架构树 | JWT+租户 | org:tree |
| POST | /api/v1/admin/org-units | 创建组织节点 | JWT+租户 | org:create |
| PUT | /api/v1/admin/org-units/:id | 更新组织节点 | JWT+租户 | org:update |
| DELETE | /api/v1/admin/org-units/:id | 删除组织节点 | JWT+租户 | org:delete |
| GET | /api/v1/admin/org-units/:id/users | 获取节点下用户 | JWT+租户 | org:user:list |
| PUT | /api/v1/admin/org-units/:id/users | 设置节点用户 | JWT+租户 | org:user:assign |
| GET | /api/v1/admin/configs | 配置列表（支持 scope/key/is_feature_flag 过滤） | JWT+租户 | config:list |
| POST | /api/v1/admin/configs | 创建配置 | JWT+租户 | config:create |
| PUT | /api/v1/admin/configs/:id | 更新配置 | JWT+租户 | config:update |
| DELETE | /api/v1/admin/configs/:id | 删除配置 | JWT+租户 | config:delete |
| GET | /api/v1/admin/configs/resolve/:key | 按三级逻辑解析有效值 | JWT+租户 | config:resolve |
| GET | /api/v1/admin/configs/feature-flags | 当前租户功能开关列表 | JWT+租户 | config:feature-flags |

### 核心逻辑

#### 1. DataScopeCallback scope_type 分支

```go
// DataScopeDimension 扩展
type DataScopeDimension struct {
    DimensionName string
    ScopeType     string   // 新增: ALL/SELF/DEPT/DEPT_TREE/CUSTOM
    TargetEntity  string
    ColumnName    string
    Values        []string // scope_type=CUSTOM 时使用
    UserID        int64    // 新增: 供 SELF 使用
    TenantID      int64    // 新增: 供 DEPT/DEPT_TREE 使用
}

// dataScopeQueryCallback 增强逻辑
func dataScopeQueryCallback(db *gorm.DB) {
    // ... 现有 SystemOp 检查 ...
    dsc := GetDataScopeContext(ctx)
    targetTable := resolveTableName(db)

    hasDimensionScope := false
    if dsc != nil && len(dsc.Dimensions) > 0 && targetTable != "" {
        hasDimensionScope = injectScopeTypeWhere(db, dsc, targetTable)
    }
    // 记录共享逻辑保持不变
    injectRecordShareScope(db, hasDimensionScope)
}

func injectScopeTypeWhere(db *gorm.DB, dsc *DataScopeContext, targetTable string) bool {
    injected := false
    hasAll := false
    // 先检查是否有 ALL 短路
    for _, dim := range dsc.Dimensions {
        if dim.TargetEntity == targetTable && dim.ScopeType == "ALL" {
            hasAll = true
            break
        }
    }
    if hasAll { return false } // ALL 短路：不注入任何条件

    for _, dim := range dsc.Dimensions {
        if dim.TargetEntity != targetTable { continue }
        switch dim.ScopeType {
        case "ALL":
            continue // 不注入
        case "SELF":
            db.Where("create_by = ?", dim.UserID)
            injected = true
        case "DEPT":
            orgIDs, _ := orgProvider.GetOrgIds(dim.UserID, dim.TenantID)
            if len(orgIDs) > 0 {
                db.Where(dim.ColumnName+" IN ?", orgIDs)
            } else {
                db.Where("1 = 0") // 无归属 → 无数据
            }
            injected = true
        case "DEPT_TREE":
            // 先获取用户主归属 org_unit_id，再获取子树
            orgIDs, _ := orgProvider.GetOrgIds(dim.UserID, dim.TenantID)
            allIDs := []int64{}
            for _, oid := range orgIDs {
                subIDs, _ := orgProvider.GetSubOrgIds(oid, dim.TenantID)
                allIDs = append(allIDs, subIDs...)
            }
            allIDs = append(allIDs, orgIDs...)
            if len(allIDs) > 0 {
                db.Where(dim.ColumnName+" IN ?", unique(allIDs))
            } else {
                db.Where("1 = 0")
            }
            injected = true
        case "CUSTOM":
            // 现有逻辑保持不变
            if len(dim.Values) > 0 {
                db.Where(dim.ColumnName+" IN ?", dim.Values)
                injected = true
            }
        }
    }
    return injected
}
```

#### 2. DefaultOrganizationProvider

```go
type DefaultOrganizationProvider struct {
    db *gorm.DB
}

// GetOrgIds 获取用户在该租户下的主归属组织节点 ID
func (p *DefaultOrganizationProvider) GetOrgIds(userID, tenantID int64) ([]int64, error) {
    var orgIDs []int64
    err := p.db.Table("admin_user_org").
        Where("user_id = ? AND tenant_id = ? AND is_primary = 1", userID, tenantID).
        Pluck("org_unit_id", &orgIDs).Error
    return orgIDs, err
}

// GetSubOrgIds 获取指定节点及所有子节点 ID（单次查询 + 内存构建树）
func (p *DefaultOrganizationProvider) GetSubOrgIds(orgID, tenantID int64) ([]int64, error) {
    // 单次查询该租户所有启用节点，避免 N+1
    var allNodes []struct{ ID, ParentID int64 }
    p.db.Table("admin_org_unit").
        Select("id, COALESCE(parent_id, 0) as parent_id").
        Where("tenant_id = ? AND status = 1", tenantID).
        Find(&allNodes)
    // 内存构建父子映射
    childMap := make(map[int64][]int64)
    for _, n := range allNodes {
        childMap[n.ParentID] = append(childMap[n.ParentID], n.ID)
    }
    // BFS 从 orgID 开始收集子树
    var result []int64
    queue := []int64{orgID}
    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        result = append(result, current)
        queue = append(queue, childMap[current]...)
    }
    return result, nil
}

// GetOrgPath 获取从指定节点到根的路径
func (p *DefaultOrganizationProvider) GetOrgPath(orgID, tenantID int64) ([]int64, error) {
    var path []int64
    current := orgID
    for current != 0 {
        path = append(path, current)
        var parentID *int64
        p.db.Table("admin_org_unit").
            Where("id = ? AND tenant_id = ?", current, tenantID).
            Pluck("parent_id", &parentID)
        if parentID == nil { break }
        current = *parentID
    }
    return path, nil
}
```

#### 3. AdminConfigService 三级配置

```go
type AdminConfigService struct {
    db    *gorm.DB
    cache *configCache // sync.Map + TTL
}

// Resolve 三级合并查询（USER > TENANT > SYSTEM）
func (s *AdminConfigService) Resolve(key string, tenantID, userID int64) (string, bool) {
    // 1. 检查缓存
    cacheKey := fmt.Sprintf("%s:%d:%d", key, tenantID, userID)
    if val, ok := s.cache.Get(cacheKey); ok {
        return val, true
    }
    // 2. 查数据库（单次查询取三级，按优先级取第一条）
    var cfgs []model.AdminConfig
    err := s.db.Where("config_key = ? AND deleted_at IS NULL AND status = 1", key).
        Where(`(
            (scope = 'USER' AND scope_id = ? AND tenant_id = ?) OR
            (scope = 'TENANT' AND scope_id = ? AND tenant_id = ?) OR
            (scope = 'SYSTEM' AND scope_id = 0 AND tenant_id = 0)
        )`, userID, tenantID, tenantID, tenantID).
        Find(&cfgs).Error
    if err != nil || len(cfgs) == 0 {
        return "", false
    }
    // 应用层按优先级排序（避免 MySQL FIELD() 不可移植）
    priority := map[string]int{"USER": 0, "TENANT": 1, "SYSTEM": 2}
    best := cfgs[0]
    for _, c := range cfgs[1:] {
        if priority[c.Scope] < priority[best.Scope] {
            best = c
        }
    }
    // 3. 写入缓存
    s.cache.Set(cacheKey, best.ConfigValue, 60*time.Second)
    return best.ConfigValue, true
}

// ResolveInt 获取 int 类型配置
func (s *AdminConfigService) ResolveInt(tenantID int64, key string, defaultVal int) int {
    val, ok := s.Resolve(key, tenantID, 0)
    if !ok { return defaultVal }
    n, err := strconv.Atoi(val)
    if err != nil { return defaultVal }
    return n
}

// IsFeatureEnabled 检查功能开关
func (s *AdminConfigService) IsFeatureEnabled(tenantID int64, featureKey string) bool {
    val, ok := s.Resolve(featureKey, tenantID, 0)
    if !ok { return true } // 未配置默认开启
    return val == "true" || val == "1"
}

// InvalidateCache 清除指定 key 相关缓存
func (s *AdminConfigService) InvalidateCache(key string) {
    s.cache.DeleteByPrefix(key)
}
```

#### 4. DynamicPermissionMiddleware 功能开关集成

```go
// 在权限码校验前增加 feature_flag 检查
func DynamicPermissionMiddleware(db *gorm.DB, cfg *config.Config, configSvc *AdminConfigService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... 白名单/认证/SUPER_ADMIN 检查同现有逻辑 ...

        // 功能开关检查：按 module_code 查找对应 feature_flag
        moduleCode := getModuleCode(c) // 从 AppResolveMiddleware 注入的上下文获取
        if moduleCode != "" {
            featureKey := "feature." + moduleCode + ".enabled"
            if !configSvc.IsFeatureEnabled(authCtx.TenantID, featureKey) {
                c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                    "code": 40302, "data": nil, "message": "该功能未开启",
                })
                return
            }
        }

        // ... 现有权限码校验逻辑 ...
    }
}
```

#### 5. 配额校验（Service 层植入示例）

```go
// UserService.CreateUser 增加配额检查
func (s *UserService) CreateUser(tenantID int64, req *CreateUserRequest) error {
    // SUPER_ADMIN 跳过配额
    if !isSuperAdmin {
        var currentCount int64
        s.db.Model(&model.User{}).Where("tenant_id = ?", tenantID).Count(&currentCount)
        maxUsers := s.configSvc.ResolveInt(tenantID, "quota.max_admin_users", 999999)
        if int(currentCount) >= maxUsers {
            return errors.NewAuthError(errors.ErrQuotaExceeded, "管理用户数已达上限")
        }
    }
    // ... 正常创建 ...
}
```

#### 6. sys_config 迁移逻辑

```go
// MigrateConfigs 一次性将 sys_config 迁移到 admin_config（幂等）
func MigrateConfigs(db *gorm.DB) error {
    var sysConfigs []model.SysConfig
    db.Find(&sysConfigs)
    for _, sc := range sysConfigs {
        var count int64
        db.Model(&model.AdminConfig{}).
            Where("config_key = ? AND scope = 'SYSTEM' AND scope_id = 0", sc.ConfigKey).
            Count(&count)
        if count > 0 { continue } // 幂等：已迁移跳过
        ac := &model.AdminConfig{
            ConfigKey:   sc.ConfigKey,
            ConfigValue: sc.ConfigValue,
            ConfigType:  "string",
            Scope:       "SYSTEM",
            ScopeID:     0,
            TenantID:    0,
            DisplayName: sc.ConfigName,
            Description: sc.Remark,
            Status:      1,
        }
        db.Create(ac)
    }
    return nil
}
```

### 中间件链变更

```
Request
  │
  ├── AuthMiddleware（JWT 解析）
  ├── TenantContextMiddleware（租户上下文）
  ├── AppResolveMiddleware（应用识别 + enabled_modules）
  ├── DynamicPermissionMiddleware（功能开关检查 + 权限码校验）  ← 增加 feature_flag 检查
  ├── DataScopeMiddleware（加载用户数据权限到 context）       ← scope_type 加载
  ├── FieldFilterMiddleware
  └── Handler
```

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有 DataScopeCallback CUSTOM 维度行为不变 | scope_type=CUSTOM 的查询结果与改动前一致 |
| RG-2 | 现有记录共享逻辑不变 | injectRecordShareScope 输出不受 scope_type 影响 |
| RG-3 | 现有 ConfigService List/Get/Create/Update/Delete 行为兼容 | 迁移后接口返回相同结构 |
| RG-4 | SUPER_ADMIN 不受配额限制 | SUPER_ADMIN 创建用户/角色不触发配额拒绝 |
| RG-5 | 登录/选租户/刷新 token 不受功能开关影响 | 免检接口白名单不变 |
| RG-6 | 现有 DynamicPermissionMiddleware 权限码校验逻辑不变 | 功能开关仅在校验前增加一步检查 |

## 正确性属性

- scope_type=ALL 的维度不得产生任何 WHERE 条件（零开销）
- 同一用户多角色数据权限取并集（最宽松范围生效）
- 三级配置 Resolve 严格遵循 USER > TENANT > SYSTEM 优先级
- 配额校验仅在 Create 操作前执行，不影响 Read/Update/Delete
- admin_config 的 uk_scope_key 唯一约束保证同一 scope+scope_id+tenant_id 下 key 不重复
- 功能开关未配置时默认开启（不影响现有功能）
- admin_org_unit 树形结构无循环引用（删除前检查子节点）
