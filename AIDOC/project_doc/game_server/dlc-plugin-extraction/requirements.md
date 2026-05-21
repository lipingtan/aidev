# 需求：DLC 管理功能分离为独立插件

## 背景

当前 DLC 管理功能耦合在 Host 的 `app/game/` 模块中。为验证插件架构的可行性并实现业务解耦，将 DLC 管理作为第一个真实插件从 Host 中分离。

## 用户故事

- 作为系统管理员，我希望 DLC 管理以插件形式独立部署，以便按需启停而不影响核心系统
- 作为开发者，我希望 DLC 插件作为插件架构的参考实现，以便后续其他功能模块化

## 功能需求

### FR-1: DLC 插件实现完整 CRUD
**描述：** DLC 插件通过插件架构的 HandleRequest 接口处理所有 DLC 管理请求
**验收标准：**
- WHEN 请求 GET /api/v1/plugin/dlc/list THEN 系统 SHALL 返回 DLC 分页列表
- WHEN 请求 GET /api/v1/plugin/dlc/:id THEN 系统 SHALL 返回 DLC 详情
- WHEN 请求 POST /api/v1/plugin/dlc THEN 系统 SHALL 创建 DLC 记录
- WHEN 请求 PUT /api/v1/plugin/dlc/:id THEN 系统 SHALL 更新 DLC 记录
- WHEN 请求 DELETE /api/v1/plugin/dlc/:id THEN 系统 SHALL 删除 DLC 记录（有订单时拒绝）

### FR-2: DLC 插件支持文件上传/下载
**描述：** 插件处理 PCK 文件上传和下载 URL 生成
**验收标准：**
- WHEN 请求 POST /api/v1/plugin/dlc/:id/upload 携带文件 THEN 系统 SHALL 存储文件并更新哈希
- WHEN 请求 GET /api/v1/plugin/dlc/:id/download THEN 系统 SHALL 返回文件下载 URL

### FR-3: DLC 插件支持统计查询
**描述：** 插件通过 db_dsn 直连主库查询 game_order 表获取统计数据
**验收标准：**
- WHEN 请求 GET /api/v1/plugin/dlc/stats THEN 系统 SHALL 返回下载量、收入、趋势、排行

### FR-4: 插件自管理数据表
**描述：** DLC 插件启动时自动创建/迁移 `dlc_game_dlc` 表（插件名前缀）
**验收标准：**
- WHEN 插件首次启动 THEN 系统 SHALL 自动创建 dlc_game_dlc 表
- WHEN 插件版本升级 THEN 系统 SHALL 自动迁移表结构

### FR-5: Host 清理 DLC 代码
**描述：** 从 Host 的 app/game/ 中删除 DLC 相关代码
**验收标准：**
- WHEN DLC 插件分离完成 THEN Host 中 SHALL 不再包含 DLC 路由、API、Service、DTO 代码
- WHEN Host 编译 THEN go build ./... SHALL 零错误

## 非功能需求

- 插件通过 db_dsn 直连主库，无需额外数据库配置
- 文件存储复用 Host 的本地存储目录
- 所有接口通过 Host 代理，自动注入 JWT 认证上下文（user_id/tenant_id/roles/permissions）
