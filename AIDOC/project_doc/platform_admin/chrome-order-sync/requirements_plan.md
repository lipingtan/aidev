# 需求计划：Chrome 插件订单同步

## 需求理解

- 目标：改造 Chrome 插件的订单推送逻辑，使其将订单数据同步到新的 platform_admin（game 插件），替代旧的 PHP 服务（html/addons/tradeorder）
- 范围：在 game 插件中新增订单同步 API 端点，接收 Chrome 插件推送的订单数据（包含订单和订单项），解析并保存到数据库
- 预期效果：Chrome 插件推送订单后，platform_admin 能够正确接收、解析、存储订单和订单项数据，并在管理后台展示

## 假设列表

- [假设-1] Chrome 插件推送的订单 JSON 结构与旧 PHP 服务一致（包含 id、buyer、payInfo、orderInfo、statusInfo、subOrders 等字段）
- [假设-2] game 插件中现有的 Order 模型需要扩展以支持订单项（subOrders），或新增 OrderItem 模型
- [假设-3] Chrome 插件推送的订单属于特定租户，tenant_id 从请求头中获取（与现有 API 一致）
- [假设-4] 订单同步 API 需要幂等性，根据订单号（orderid）去重，已存在的订单更新状态，不存在的创建新订单

## 澄清问题

- [Question-1] 数据模型方案：旧的 PHP 服务中订单（tradeorder）和订单项（tradeorder_item）是分开存储的。game 插件中现有的 Order 模型是简化的单一订单模型。是否需要：
  - A. 保持 Order 模型不变，新增 OrderItem 模型（订单项）
  - B. 在 Order 模型中内嵌订单项数组（JSON 字段）
  - 业界最佳实践：主流电商系统（如 Shopify、Stripe）都将订单和订单项分离存储，便于查询和扩展。订单项是订单的一对多关系，分离存储更符合数据库范式。
  - 推荐答案及理由：方案 A（新增 OrderItem 模型），与旧 PHP 服务的数据结构一致，便于数据迁移和后续扩展。
  [Answer-1]
方案A
- [Question-2] API 端点路径：Chrome 插件推送订单的 API 端点路径是什么？
  - A. `/api/v1/plugin/game/order/sync`（推荐，与现有 order 路由一致）
  - B. 其他路径
  - 业界最佳实践：API 端点路径应与资源类型一致，使用动词表示操作。`/order/sync` 表示订单同步操作，符合 RESTful 风格。
  - 推荐答案及理由：方案 A（`/api/v1/plugin/game/order/sync`），与现有 game 插件的 order 路由前缀一致，便于维护和权限管理。
  [Answer-2]
方案A
- [Question-3] 幂等性策略：Chrome 插件可能重复推送同一订单，是否需要去重？
  - A. 根据订单号（order_no）去重，已存在的订单更新状态，不存在的创建新订单
  - B. 每次推送都创建新订单（不去重）
  - 业界最佳实践：主流支付系统（如 Stripe、支付宝）都采用幂等性设计，根据订单号去重，避免重复扣款或重复发货。
  - 推荐答案及理由：方案 A（根据订单号去重），确保数据一致性，避免重复订单。
  [Answer-3]
A
- [Question-4] 租户隔离：Chrome 插件推送的订单是否属于特定租户？如何确定 tenant_id？
  - A. 从请求头中获取（与现有 API 一致）
  - B. 订单数据中包含 tenant_id
  - C. 全局订单，tenant_id = 0
  - 业界最佳实践：多租户 SaaS 系统中，租户 ID 通常从认证上下文（请求头或 Token）中获取，而非从请求体中获取，以防止租户 ID 被篡改。
  - 推荐答案及理由：方案 A（从请求头中获取），与现有 game 插件 API 的租户隔离机制一致。
  [Answer-4]
方案A
- [Question-5] 响应格式：Chrome 插件期望的响应格式是什么？
  - A. 与现有 game 插件 API 保持一致（`{code, msg, data}`）
  - B. 与旧 PHP 服务保持一致
  - 业界最佳实践：同一系统的 API 应使用统一的响应格式，便于前端和客户端处理。
  - 推荐答案及理由：方案 A（与现有 game 插件 API 保持一致），减少 Chrome 插件的适配成本。
  [Answer-5]
A

- [Question-6] 认证机制：Chrome 插件推送订单时，使用什么认证机制？
  - A. JWT Token（与现有 API 一致，Chrome 插件需先获取 Admin 用户的 Token）
  - B. API Key（为 Chrome 插件生成专用 API Key）
  - C. 无认证（不推荐）
  - 业界最佳实践：浏览器端发起的 API 请求通常使用 JWT Token，与现有 API 认证机制保持一致，便于统一管理和权限控制。
  - Token 有效期：Admin Access Token 2 小时，Platform Token 7 天（用于刷新）
  - Token 刷新：`POST /auth/refresh` 接口，使用 Platform Token 刷新 Access Token
  - Platform Token 过期后需要重新登录（无刷新机制）
  - Chrome 插件需要实现：Token 存储、Access Token 刷新、Platform Token 过期后重新登录
  - 推荐答案及理由：方案 A（JWT Token），与现有 game 插件 API 的认证机制一致，Chrome 插件在推送订单前需要先获取一个 Admin 用户的 Token。Chrome 插件需要实现 Token 存储、刷新和重新登录逻辑。
  [Answer-6]

## 非功能需求建议

- 性能：订单同步接口响应时间 < 500ms（P99），支持批量推送（一次推送多个订单）
- 安全：订单同步接口需要 JWT 认证和租户隔离
- 多租户：订单数据按 tenant_id 隔离，不同租户的订单互不可见
- 幂等性：根据订单号去重，重复推送不产生重复订单
- 数据一致性：订单和订单项的保存必须在同一事务中，确保数据一致性

## 影响范围预判

### platform_admin（game 插件）
- 涉及模块：game 插件（plugins/game/）
- 涉及文件（预估）：
  - plugins/game/models.go（新增 OrderItem 模型）
  - plugins/game/handler.go（新增订单同步路由）
  - plugins/game/service_order.go（新增订单同步处理逻辑）
  - plugins/game/db.go（自动迁移新增表）
- 可能的副作用：无，新增功能不影响现有功能

### Chrome 插件（另一个任务）
- Chrome 插件代码路径：`C:\data\developer\projects\tbsend\thinkfansxiaoer1-edge-mv3`
- Chrome 插件改造是另一个独立任务，需要改造 Chrome 插件让它同时也推送到新的 platform_admin 服务
- **注意：Chrome 插件改造任务开始前需要与用户确认**
