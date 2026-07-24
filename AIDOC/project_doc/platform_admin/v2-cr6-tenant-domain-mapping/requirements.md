# 需求：V2-CR6 域名-租户映射管理模块

## 背景

CR-5 实现了 C 端双用户池认证，C 端用户登录时需要提供 tenant_code 以确定所属租户。当前实现依赖前端静态配置，无法支持多租户按域名自动路由的场景（不同企业客户通过不同子域名访问同一 C 端服务）。

本 CR 在后端建立域名-租户映射关系表，并提供公开查询接口，使 C 端前端在用户访问时自动解析 tenant_code，彻底消除用户手动输入租户信息的需求。

---

## 用户故事

- 作为**平台超级管理员**，我希望能在管理后台配置域名与租户的绑定关系，以便为不同企业客户分配独立的访问入口
- 作为**C端普通用户**，我希望在访问企业专属域名时直接进入登录页，无需了解或填写任何租户信息，以便获得简洁无感知的登录体验
- 作为**系统运维人员**，我希望新增/修改域名绑定后缓存能及时失效，以便配置变更即刻生效

---

## 功能需求

### FR-1: 域名-租户映射 CRUD（管理端）

**描述**：超级管理员可以在管理端对域名-租户绑定关系进行增删改查操作

**验收标准**：
- WHEN 超级管理员提交有效的域名和租户信息 THEN 系统 SHALL 创建一条 `tenant_domain` 记录，domain 字段唯一（活跃记录中不允许重复）
- WHEN 超级管理员提交已存在的域名 THEN 系统 SHALL 返回 409 冲突错误，message 包含"域名已被绑定"
- WHEN 超级管理员提交 `localhost`、`127.0.0.1`、`0.0.0.0` THEN 系统 SHALL 返回 400，message 包含"保留域名不允许绑定"
- WHEN 超级管理员提交长度超过 255 字符的域名 THEN 系统 SHALL 返回 400，message 包含"域名长度超限"
- WHEN 超级管理员更新域名绑定 THEN 系统 SHALL 更新记录，同时**主动失效**该域名的查询缓存（写操作成功后执行失效）
- WHEN 超级管理员删除域名绑定 THEN 系统 SHALL 软删除记录，同时**主动失效**该域名的查询缓存
- WHEN 超级管理员查询域名列表 THEN 系统 SHALL 返回分页数据，支持按域名模糊搜索和按租户过滤
- WHEN 非 SUPER_ADMIN role_type 的角色调用 CRUD 接口 THEN 系统 SHALL 返回 403（Service 层显式校验 SUPER_ADMIN，不完全依赖 permission_code）

### FR-2: 域名查询公开接口

**描述**：C端前端在用户未登录时通过当前访问域名查询对应的 tenant_code

**验收标准**：
- WHEN 前端传入已配置的域名 THEN 系统 SHALL 返回对应的 tenant_code 和 tenant_name
- WHEN 前端传入未配置的域名 THEN 系统 SHALL 返回默认租户（tenant_code=default），不返回错误
- WHEN 该接口被请求 THEN 系统 SHALL 无需携带任何认证 token 即可访问（公开接口）
- WHERE 缓存存在 WHEN 相同域名被查询 THEN 系统 SHALL 命中内存缓存，不重复查库，响应时间 < 10ms

### FR-3: 域名查询缓存

**描述**：对域名-租户映射关系实施内存缓存以降低 DB 压力

**验收标准**：
- WHEN 域名首次被查询 THEN 系统 SHALL 查库并将结果写入内存缓存，TTL 为 5 分钟
- WHEN 管理员对某条域名记录执行增/改/删 THEN 系统 SHALL 立即失效该域名的缓存条目
- WHEN 缓存条目过期后 THEN 系统 SHALL 在下次请求时重新查库并回填缓存

### FR-4: 结构预留通配符扩展

**描述**：数据库结构和服务层预留通配符匹配的扩展点，本 CR 仅实现精确匹配

**验收标准**：
- WHEN 设计 `tenant_domain` 表 THEN 系统 SHALL 包含 `match_type` 字段（枚举：EXACT / WILDCARD），本 CR 所有记录均为 EXACT
- WHEN 查询服务接收到域名 THEN 系统 SHALL 优先精确匹配，未来可在此扩展通配符分支

### FR-5: C端登录页自动解析 tenant_code

**描述**：dev-web-user 登录页启动时调用域名查询接口，自动获取 tenant_code，用户不感知

**验收标准**：
- WHEN C端登录页加载 THEN 系统 SHALL 自动以当前 `window.location.hostname` 调用域名查询接口
- WHEN 域名查询**进行中** THEN 系统 SHALL 显示发送验证码按钮为禁用状态，按钮文案改为"加载中…"，防止竞态
- WHEN 查询成功（包括兜底返回 default）THEN 系统 SHALL 将 tenant_code 存入组件状态，恢复按钮可用，登录/发送验证码时自动携带，不展示给用户
- WHEN 查询超时（3 秒）或网络异常 THEN 系统 SHALL 使用 `VITE_TENANT_CODE` 环境变量作为兜底，恢复按钮可用，保证登录流程不中断
- WHEN 登录表单界面渲染 THEN 系统 SHALL 不显示任何租户编码相关输入字段

### FR-6: 管理端域名管理页面

**描述**：在 dev-web-admin 的系统管理菜单下新增"域名管理"页面

**验收标准**：
- WHEN 超级管理员进入域名管理页 THEN 系统 SHALL 展示域名列表（含：域名、绑定租户、创建时间、操作）
- WHEN 管理员点击"新增" THEN 系统 SHALL 弹出表单，包含域名输入框和租户下拉选择器
- WHEN 管理员点击"编辑" THEN 系统 SHALL 回填当前数据
- WHEN 管理员点击"删除" THEN 系统 SHALL 弹出二次确认框，确认后删除
- WHEN 操作成功 THEN 系统 SHALL 展示 ElMessage 成功提示，列表自动刷新

---

## 非功能需求

- **性能**：域名查询接口 P99 响应 < 50ms；缓存命中时 < 10ms
- **安全**：公开查询接口仅返回 tenant_code 和 tenant_name，不暴露 tenant_id 等内部字段；CRUD 接口需 JWT 认证 + SUPER_ADMIN 角色校验
- **多租户**：`tenant_domain` 表本身是平台级数据，不走 TenantIsolationCallback（无 tenant_id 字段需过滤）
- **可扩展**：`match_type` 字段预留通配符扩展，服务层查询路径已预留分支注释

---

## 数据约束

- 一个域名只能绑定一个租户（活跃记录中 domain 字段唯一，通过 Service 层业务唯一性校验保证）
- 一个租户可绑定多个域名
- `localhost`、`127.0.0.1`、`0.0.0.0` 不允许绑定到任何租户（保留为开发环境兜底）
- 域名长度不超过 255 字符
- 若 default 租户不存在，QueryByDomain 硬编码返回 `{tenant_code:"default", tenant_name:"默认租户"}`，不查库，确保接口永不失败
- 若某域名绑定的租户被软删除，该域名查询自动降级到 default 租户（属于运维问题，业务上可接受）

---

## 不涉及范围

- 本 CR 不实现通配符匹配逻辑（仅预留结构）
- 本 CR 不实现租户管理员自助绑定域名功能
- 本 CR 不实现域名所有权验证（CNAME/TXT 记录）
