# platform_admin 项目 Backlog

> 未排期的待办事项，按模块分类记录。正式开发前需要拆解为 CR 文档。

---

## 游戏中心（game-center）

### BACKLOG-GC-01：腾讯 COS 素材管理 + 游戏详情图片关联配置

**优先级：** 中  
**来源：** 2026-10-05 用户需求  
**状态：** 待规划

**背景：**
当前游戏封面图、截图等媒体资源直接用外链 URL 硬填，无统一素材管理。

**需求描述：**
1. **素材管理模块**（管理端）
   - 集成腾讯云 COS（对象存储）作为媒体存储后端
   - 管理端提供素材库页面：支持上传图片/视频、分类管理（游戏封面/截图/Banner/公告图）、预览、删除
   - 素材以租户隔离（按 tenant_id 划分 COS 路径前缀）
   - 后端提供签名上传 URL（避免前端直接持有 COS 密钥）

2. **游戏详情图片关联**
   - `games` 表增加 `screenshots` 字段（JSON 数组，存 COS 图片 URL 列表）
   - 游戏编辑页（Games.vue）支持从素材库选择截图，最多可关联 N 张
   - 游戏封面图（`cover_url`）也改为从素材库选取，而非手填 URL
   - `game_banners` 的 `image_url` 同理支持素材库关联

3. **App 端消费**
   - `PublicGameItem` 的 `screenshots` 字段由后端返回 COS URL 列表
   - `GameDetailActivity` 的预览横滑区展示真实截图（当前已预留 `gameItem.screenshots` 字段和 `populateShots()` 逻辑，服务端就绪后自动生效）

**技术要点：**
- 腾讯 COS Go SDK：`github.com/tencentyun/cos-go-sdk-v5`
- 签名上传：后端生成预签名 PUT URL，前端直传 COS，再把 URL 写库（避免后端转发大文件）
- CDN 加速：COS 桶绑定 CDN 域名，App 端访问走 CDN
- 素材去重：上传时按文件 MD5 查重，同一租户内相同文件复用
- 权限：管理端素材管理需要 `game:media:upload` 权限点

**依赖：**
- 需要申请腾讯云 COS Bucket 及访问密钥
- 管理端需要新增「素材管理」菜单项和路由

**预估规模：** 大型 CR，建议拆为两个子 CR：
- CR-A：后端 COS 集成 + 签名上传接口 + 素材表
- CR-B：管理端素材库 UI + 游戏编辑页图片选择器 + App 端 screenshots 字段生效

---


## 游戏中心（game-center）

### BACKLOG-GC-02：同名 ROM 不同修改版本支持

**优先级：** 低  
**来源：** 2026-10-23 架构讨论  
**状态：** 待规划

**背景：**
当前 `games` 表以 `(tenant_id, rom_name)` 为联合唯一键，同一租户内同名 ROM 只能有一条记录。
实际存在同一驱动名但内容不同的修改版本（如不同补丁版、不同地区版），当前架构无法区分。

**问题描述：**
1. **数据层**：`(tenant_id, rom_name)` 唯一约束导致同名 ROM 只能存一条
2. **文件层**：不同版本 ROM zip 文件名相同，下载到同一目录会互相覆盖
3. **MAME 层**：驱动名（rom_name）是 MAME 内核识别游戏的唯一标识，不能修改

**方案方向（待细化）：**
- **数据层**：唯一键改为 `(tenant_id, rom_name, variant)`，`variant` 默认空字符串代表主版本
- **文件层**：下载目录按 `game_id`（雪花 ID）隔离，路径从 `roms/{rom_name}.zip` 改为 `roms/{game_id}/{rom_name}.zip`；parent ROM 放公共 `roms/shared/` 目录避免重复下载
- **App 层**：`launchGame` 启动前将 `PREF_ROMsDIR_2` 设为该 game_id 对应的隔离目录

**注意：**
- `Emulator.ROM_PATH` 是单路径字符串，MAME native 层是否支持 `;` 分隔多路径**未验证**
- 方案B（game_id 目录隔离）不依赖多路径，每次启动只设一个目录，规避了该风险
- 改动涉及：后端 DDL + service + CDN 路径生成 + App DownloadManager + launchGame

**依赖：**
- 需明确实际有几个同名 ROM 不同版本的游戏，评估优先级
- 建议有具体案例时再立 CR



---

## CR14-B1：App 端客服 bot 回复等待态 UI

**来源**：CR14 三角色 review（2026-10-31）  
**优先级**：中  
**背景**：CR14 P0 主链路中，用户在 App 发消息后，platform_admin 异步触发 Agent，整个链路耗时可达 5~30s（Agent 调用 + LLM 生成 + MCP 工具 + send_bot_message 回写）。当前 App 端 CS 聊天界面没有设计「bot 正在回复」的等待态，用户可能误认为没有响应。

**需求描述：**
- 用户发送消息后，App 端聊天界面立即显示一个「AI 正在分析您的问题...」的 typing indicator（三点跳动气泡，bot 头像）
- 触发条件：`session.mode == "bot"` 时，用户消息发送成功（HTTP 200）后即显示
- 消失条件：收到 bot 新消息（轮询或 WS 推送）时自动消失；或超过 35s（略大于后端 30s 超时）后自动消失并提示「智能客服处理中，请稍候...」
- 转人工时：typing indicator 消失，不需要展示

**依赖：**
- CR14 P0 主链路完成后评估实现
- App 端需要知道 `session.mode`（接口已有此字段）

**注意：**
- 超出 CR14 范围，不列入 CR14 任务清单，CR14 交付后单独排期

