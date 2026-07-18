# 设计计划：SaaS 底座插件架构

## 设计范围

基于 requirements.md 中的 FR-1 ~ FR-7，设计插件架构的技术方案。

## 澄清问题

[Question] Q1: go-plugin 的 gRPC 通信中，插件需要访问数据库——是通过 Host 传递 DSN 让插件自己连接，还是 Host 提供 DB 代理 RPC（插件通过 gRPC 调用 Host 执行 SQL）？前者性能好但插件有直接 DB 访问权限，后者安全但性能差。
[Answer]
前者
[Question] Q2: 插件前端 bundle 的存储位置——是放在 Host 的 static 目录下由 Host 直接 serve，还是插件自带 HTTP 服务提供静态文件？
[Answer]
host 直接serve
[Question] Q3: 插件间通信的 CallPlugin RPC——是同步调用（等待响应），还是需要支持异步事件（发布/订阅模式）？
[Answer]
两者都需要
[Question] Q4: 插件的权限模型——插件声明的权限是自动注册到 Host 的 RBAC 系统（casbin），还是插件自己管理权限？
[Answer]
复用host的RBAC
[Question] Q5: 第一阶段的交付范围——是否只实现 FR-1（插件管理器）+ FR-2（gRPC 协议）+ FR-3（请求代理）作为 MVP，FR-4~FR-7 后续迭代？
[Answer]
全部