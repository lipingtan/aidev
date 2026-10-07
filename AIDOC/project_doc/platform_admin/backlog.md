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
