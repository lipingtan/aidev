# 需求计划：对象 API 显示名称增强

## 需求理解

- 目标：为对象（GROUP）和接口（ENDPOINT）增加中文显示名称（display_name），在应用管理和角色 API 权限配置界面中展示可读名称而非路径
- 范围：后端 model/AutoDiscover + 前端应用管理对象API树 + 角色权限 Drawer API Tab
- 预期效果：
  - 对象分组显示中文名（如 "用户管理(users)" 而非 "users"）
  - 接口节点保留 HTTP 方法 + 路径展示，display_name 用于补充描述
  - 角色 API 权限配置和应用管理对象API树保持一致的展示风格
  - 编辑 API 节点时可修改 display_name

## 假设列表

- [假设-1] admin_api_permission 表新增 `display_name` 字段（varchar(128)），`name` 保留原始路径值不变
- [假设-2] AutoDiscover 根据已知路径映射表自动生成中文 display_name
- [假设-3] 无匹配映射时 ENDPOINT fallback 显示 "endpoint"，GROUP fallback 显示原始路径段

## 澄清问题（已回答）

- [Question-1] 显示名称是新增一个 `display_name` 字段，还是直接改 `name` 字段？
  [Answer-1] name 保留但暂不使用，增加 display_name 用于这个意图

- [Question-2] AutoDiscover 自动生成的中文名格式？
  [Answer-2] GROUP: "用户管理(users)"；ENDPOINT: 保留 HTTP 方法标签 + 路径的树结构展示，display_name 作为补充描述

- [Question-3] 角色 API 权限 Drawer 中的展示？
  [Answer-3] 树结构与应用管理一致，显示/隐藏开关只读，无编辑/删除/新增操作按钮

## 补充需求

- **编辑 display_name**：应用管理中编辑 API 节点的弹窗需新增 display_name 输入框，允许管理员自定义显示名称
- **Fallback 逻辑**：AutoDiscover 无匹配映射时，ENDPOINT 的 display_name 设为空字符串，前端展示时如果 display_name 为空则显示 URL 路径
- **已知映射表**：后端维护一个 GROUP 路径段 → 中文名 的 map，以及 ENDPOINT HTTP方法+路径 → 中文描述的 map，初始化时自动匹配

## 影响范围预判

**后端：**
- model/api_permission.go — 新增 display_name 字段
- discovery/auto_discover.go — GROUP/ENDPOINT 创建时根据映射表自动填充 display_name
- service/api_permission_service.go — Update 方法支持更新 display_name
- handler/api_permission_handler.go — 编辑接口接受 display_name 参数

**前端：**
- api/api-permission.ts — 类型定义添加 display_name 字段
- api/role.ts — ApiPermTreeNode 添加 display_name
- ApiPermissionTree.vue — GROUP 节点显示 display_name（若有）；编辑弹窗增加 display_name 输入框
- ApiPermTab.vue（角色配置）— 树结构同应用管理，开关只读，无操作按钮
