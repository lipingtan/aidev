# 需求：SaaS 底座插件架构

## 背景

将当前 platform_admin 工程抽象为通用 SaaS 服务脚手架底座。基础能力（用户/角色/菜单/字典/租户/权限）作为 Host 核心，业务功能通过插件机制扩展。插件基于 hashicorp/go-plugin（gRPC over stdio）实现进程隔离，同时提供前后端能力。

## 用户故事

- 作为平台管理员，我希望在管理界面安装/卸载/启停插件，以便灵活扩展系统功能
- 作为插件开发者，我希望通过脚手架工具快速创建插件项目，以便专注业务逻辑开发
- 作为插件开发者，我希望插件能同时提供后端 API 和前端 UI，以便交付完整功能模块
- 作为系统，我希望插件崩溃不影响 Host 运行，以便保证平台稳定性
- 作为系统，我希望插件之间可以通信，以便实现跨模块协作

## 功能需求

### FR-1: 插件管理器（Host 后端）

**描述：** Host 内置插件管理器，负责插件的发现、启动、停止、健康检查。

**验收标准：**
- WHEN Host 启动 THEN 系统 SHALL 扫描插件目录，自动启动已启用的插件进程
- WHEN 管理员通过 API 启动插件 THEN 系统 SHALL 通过 go-plugin 启动插件子进程并建立 gRPC 连接
- WHEN 管理员通过 API 停止插件 THEN 系统 SHALL 优雅关闭插件进程
- WHEN 插件进程崩溃 THEN 系统 SHALL 记录错误日志，不影响 Host 和其他插件运行
- WHEN 插件启动 THEN 系统 SHALL 调用插件的 Register 方法获取路由/菜单/权限声明

### FR-2: 插件 gRPC 接口协议

**描述：** 定义 Host 与插件之间的 gRPC 通信协议。

**验收标准：**
- WHEN 插件启动 THEN 系统 SHALL 通过 Register RPC 返回：插件名称、版本、路由前缀、菜单声明、权限声明、前端 bundle 路径
- WHEN Host 收到匹配插件路由前缀的 HTTP 请求 THEN 系统 SHALL 通过 HandleRequest RPC 转发到插件，携带用户ID/租户ID/请求体
- WHEN Host 调用 Healthcheck RPC THEN 插件 SHALL 返回健康状态
- WHEN 插件需要调用其他插件 THEN 系统 SHALL 通过 Host 提供的 CallPlugin RPC 中转

### FR-3: HTTP 请求代理

**描述：** Host 的 Gin 中间件拦截插件路由前缀的请求，通过 gRPC 转发到对应插件。

**验收标准：**
- WHEN 请求路径匹配 `/api/v1/plugin/{pluginName}/...` THEN 系统 SHALL 转发到对应插件的 HandleRequest
- WHEN 转发请求 THEN 系统 SHALL 注入上下文：当前用户ID、租户ID、角色、权限列表
- WHEN 插件返回响应 THEN 系统 SHALL 将 gRPC 响应转换为 HTTP 响应返回客户端
- WHEN 插件未运行 THEN 系统 SHALL 返回 503 错误

### FR-4: 插件前端加载（独立 JS bundle）

**描述：** 插件前端打包为独立 JS bundle，Host 前端运行时动态加载。

**验收标准：**
- WHEN 插件注册时声明前端 bundle 路径 THEN Host 前端 SHALL 通过 dynamic import 加载该 bundle
- WHEN bundle 加载成功 THEN 系统 SHALL 将插件导出的路由注册到 Vue Router
- WHEN 插件声明菜单 THEN 系统 SHALL 将菜单项注入侧边栏
- WHEN 插件停止 THEN 系统 SHALL 从路由和菜单中移除该插件的条目

### FR-5: 插件数据库隔离

**描述：** 插件共享 Host 数据库，表名通过插件名前缀隔离。

**验收标准：**
- WHEN 插件需要创建表 THEN 系统 SHALL 使用 `{pluginName}_{tableName}` 格式命名
- WHEN Host 提供 DB 连接给插件 THEN 系统 SHALL 通过 gRPC 传递 DSN 或提供 DB 代理 RPC
- WHEN 插件卸载 THEN 系统 SHALL 提供选项：保留数据 / 清除插件表

### FR-6: 插件管理界面

**描述：** Host 前端提供插件管理页面。

**验收标准：**
- WHEN 管理员访问插件管理页 THEN 系统 SHALL 展示已安装插件列表（名称、版本、状态、描述）
- WHEN 管理员点击"安装" THEN 系统 SHALL 支持本地上传（zip/tar.gz）或输入远程下载 URL
- WHEN 管理员点击"启动/停止" THEN 系统 SHALL 调用对应 API 启停插件
- WHEN 管理员点击"卸载" THEN 系统 SHALL 停止插件并删除文件（可选清除数据）

### FR-7: 插件开发脚手架 CLI

**描述：** 提供 CLI 工具快速创建插件项目骨架。

**验收标准：**
- WHEN 开发者执行 `scaffold new-plugin {name}` THEN 系统 SHALL 生成插件项目目录结构（backend + frontend）
- WHEN 生成完成 THEN 项目 SHALL 包含：main.go（插件入口）、proto 定义、前端 Vue 组件骨架、构建脚本
- WHEN 开发者执行 `scaffold build` THEN 系统 SHALL 编译后端二进制 + 打包前端 bundle

## 非功能需求

- 性能：插件 gRPC 调用延迟 < 5ms（本地 stdio 通信）
- 安全：插件进程独立运行，按最佳实践限制权限，保持灵活性
- 稳定性：插件崩溃不影响 Host，go-plugin 进程隔离保证
- 兼容性：插件接口版本化，Host 升级时向后兼容旧版插件
- 开发体验：提供脚手架 CLI + 插件开发文档
