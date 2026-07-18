# 设计计划：插件化架构规范（Plugin Architecture）

## 设计方向

基于需求文档确认的"最小框架 + 插件扩展"架构，设计三端（admin 前端、user 前端、Go 后端）的插件化技术方案。

核心技术选型：
- **前端插件加载**：Vite 动态 import + ES Module（不引入微前端框架）
- **后端插件机制**：Go plugin（.so）+ 路由注册 + DB 元数据（沿用现有 plugin-sdk 方向）
- **插件间通信**：Pinia 只读共享 Store + mitt EventBus + 扩展点注册
- **集群同步**：DB 真相源 + Redis Pub/Sub（预留接口）
- **用户端响应式**：Vant 4 + CSS Grid/Flex media query

## 技术选型理由

| 选项 | 优点 | 缺点 | 采用 |
|------|------|------|------|
| Vite 动态 import 加载插件 | 零额外依赖、原生 ESM、开发时 HMR 直接支持 | 不支持沙箱隔离 | ✓ |
| qiankun/wujie 微前端 | 强隔离 | 过重、调试困难、HMR 失效 | ✗ |
| Module Federation | 跨应用共享依赖 | 配置复杂、Vite 生态不成熟 | ✗ |
| mitt 作为 EventBus | 轻量（<1KB）、TypeScript 友好 | 功能少（够用） | ✓ |
| Redis Pub/Sub 集群广播 | 实时、已有 Redis 基础设施 | 依赖 Redis 可用性 | ✓（预留） |

## 澄清问题

- [Question-1] 后端插件目前是 Go plugin（.so 动态库）方式还是其他方式（如独立进程/gRPC）？需要确认当前 plugin-sdk 的实现方向。
  [Answer-1] 采用统一接口 + 双适配器设计。定义 Plugin Interface，实现 LocalAdapter（内嵌注册，当前使用）和 RemoteAdapter（gRPC/HTTP，预留接口不实现）。框架面向接口编程，后续新增远程插件只需实现 RemoteAdapter，框架代码零改动。

- [Question-2] 插件前端资产（独立构建产物）的分发方式：是跟后端一起打包到容器镜像中，还是上传到 CDN/对象存储？这影响安装/升级流程设计。
  [Answer-2] 和后端一起打包到容器镜像中。

- [Question-3] 插件回滚的"快照"指什么？是保留旧版本的二进制文件/JS bundle，还是需要数据库迁移回滚能力？
  [Answer-3] 都需要。保留旧版本文件（二进制 + 前端 bundle）+ 数据库迁移回滚（插件需提供 down migration）。

## 风险点

- [Risk-1] Go plugin 机制要求主程序和插件使用完全相同的 Go 版本和依赖版本编译，版本不匹配会导致加载失败。如果后续 Go 版本升级频繁，可能需要改用 gRPC/HTTP 插件方式。
- [Risk-2] 插件动态 import 在生产环境需要正确配置 CORS 和缓存策略，否则可能出现加载失败或旧版本缓存问题。
- [Risk-3] 插件自动注册权限可能与手动配置的权限产生冲突（同名覆盖），需要明确冲突解决策略。
