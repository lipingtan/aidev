# CR13 任务拆解

## 状态：✅ 全部完成并真机验收（2026-10-05）

---

## 后端任务

### T1 models.go ✅
- games 表加 description / download_count / avg_score / score_count / updated_at
- 新增 GameScore / GameReview / GameReviewLike struct
- errors.go 加 ErrCodeReviewAlreadyLiked(40112) / ErrCodeScoreOutOfRange(40110) / ErrCodeReviewEmpty(40111)

### T2 game_content_service.go ✅
- GetDownloadURLs 异步 download_count +1（goroutine，写失败不阻塞主流程）
- ScoreGame：验证→事务upsert→事务外重算avg写回
- CreateReview：验证→插入
- ListReviews：分页查询+is_liked字段（currentUserID=0时全false）
- LikeReview：事务insert+like_count+1，唯一索引防重复

### T3 game_content_api.go ✅
- POST /games/:rom/score（userAuth）
- GET  /games/:rom/reviews（public，JWT可选解析）
- POST /games/:rom/reviews（userAuth）
- POST /reviews/:id/like（userAuth）
- PublicGameItem 补充 description/download_count/avg_score/score_count/updated_at
- 新增 queryInt / paramInt64 helper函数

### T4 license_service.go ✅
- AutoMigrate 加 GameScore / GameReview / GameReviewLike

### T5 plugin.go ✅
- userAuthGroup 注册评分/评价/点赞路由
- publicGroup 注册评价列表路由

### T6 game_price_service.go ✅
- GameInput 加 description 字段
- CreateGame/UpdateGame 保存 description
- PublicGameItem 加 description/download_count/avg_score/score_count/updated_at
- ListPublicGames 返回新字段

### T7 管理端 Games.vue + license.ts ✅
- GameInput/Game 接口加 description
- 表单加「游戏简介」文本域（textarea 4行）
- handleAdd/handleEdit/form 初始化/赋值 description 字段

---

## App 端任务

### T8 新增 Model ✅
- ScoreResult.java / ReviewItem.java / ReviewListResult.java（含完整 @SerializedName 注解）
- GameItem 加 description / downloadCount / avgScore / scoreCount / updatedAt / screenshots

### T9 GameCenterApi.java ✅
- scoreGame / listReviews / createReview / likeReview 四个方法

### T10 GameDetailActivity 完整重写 ✅

**布局：**
- 沉浸式头图（全宽 CENTER_CROP，230dp，淡紫渐变背景，返回+分享占位按钮，VIP badge）
- 标题区（游戏名+副标题+收藏★+分享↗占位）
- 三列统计行（评分/大小/下载量，formatCount: ≥10000→万）
- 分类标签行（tags pill 样式）
- 首屏主按钮行（删除+主按钮，7态状态机）
- 预览区（来自服务端 screenshots 字段，无截图显示「暂无预览」）
- Tab栏（简介/评分/评价，吸顶实现，评价数量异步更新）
- 简介面板（文本3行折叠展开，信息表格5行含更新时间）
- 评分面板（大字评分+5→1星分布条+交互打分）
- 评价面板（列表+随机渐变头像+点赞+写评价弹窗）
- 吸底主按钮（动态阈值，280ms动画）

**状态机（7态，权限优先原则）：**
```
1. 未登录（任何游戏）→ STATE_NOT_LOGGED → 「登录下载」→ 跳登录页
2. 已登录 + VIP游戏 + 无会员 → STATE_VIP_NO_MEMBER → 「开通会员」→ 跳购买页
3. 已登录 + 有权限 + ROM已就绪 → STATE_READY → 「开始游戏」→ 启动 MAME4droid
4. 已登录 + 有权限 + ROM不在本地 → STATE_FREE_UNDOWNLOADED → 「下载」
5. 下载中 → STATE_DOWNLOADING → 「暂停」
6. 已暂停 → STATE_PAUSED → 「继续」
7. 校验中 → STATE_VERIFYING → 「校验中」
```

**权限检测顺序（每次 onResume 重新判断）：**
- 先检查登录 → 再检查会员 → 再检查ROM是否就绪
- ROM就绪检查：DB已完成 OR 内置路径 OR HBMJ1/roms下载路径

**游戏启动：**
- 通过 `Intent.setClassName(ctx, "com.seleuco.mame4droid.MAME4droid")`
- 传入 `com.btcola.baseuilib.common.GameConfig`（实现 Serializable）
- 与 GameListFragment / GameListActivity 的启动路径完全一致

### T11 GameHallFragment 游戏大厅 ✅

**布局重构：**
- 顶栏（固定）+ 进度条（固定）+ ScrollView（Banner+公告+Tab+卡片整体上下滚动）
- GridView 改为嵌套双列 LinearLayout（兼容 ScrollView）

**功能：**
- 免费Tab计数：requiredLevels 为空才算免费（不是"非VIP"）
- 卡片移除下载按钮，统一点击进详情页（「详情 ›」提示）
- 标签显示：优先服务端 tags，无数据时按 requiredLevels 推断（免费/麻将+VIP badge）
- 分页：每页20个，底部「加载更多」
- 收藏夹从底部 Tab 移到「我的」页面内（FavoritesActivity 包装 FavoritesFragment）
- 底部改为3个Tab：游戏大厅 / 我的 / 设置

### T12 本地数据（files.zip / gamelist）✅
- `files.zip/playlist/gamelist` 和 `res/raw/gamelist` 均加入 `tags: ["majiang"]`
- `lhb` 的 `requiredLevels: []`（免费），其余24个 `requiredLevels: ["vip"]`
- `loadBuiltinGameList` 补全解析 tags / requiredLevels 字段
- `getGameList` 检测 SQLite 旧缓存 tags 为空时自动 patch 写回

### T13 AndroidManifest.xml ✅
- 注册 GameDetailActivity（NoActionBar）
- 注册 NoticeListActivity
- 注册 FavoritesActivity（parentActivity = MjarchGameListActivity）

---

## 测试任务 ✅
- TestScoreGame：upsert+avg重算 PASS
- TestCreateAndListReviews：空内容报错+正常发评 PASS
- TestLikeReview：点赞成功+重复点赞报错+is_liked PASS
- TestPublicGameItem_Description：description字段正确返回 PASS

---

## 真机验收结果（2026-10-05）

| 验证项 | 结果 |
|---|---|
| go build | ✅ PASS |
| go test -p 1（全部 PASS） | ✅ PASS |
| vite build | ✅ PASS |
| APK 编译 versionCode=5 | ✅ PASS |
| 真机安装（R5CXC2E6CAR） | ✅ Success |
| Banner 点击跳详情 | ✅ 修复（AndroidManifest 注册缺失） |
| 卡片点击跳详情 | ✅ 修复（同上） |
| 开始游戏启动模拟器 | ✅ 修复（改用 MAME4droid 路径） |
| 未登录拦截 | ✅ 修复（权限检测前置，ROM就绪也拦截） |
| 有会员正确进游戏 | ✅ 修复（状态机单一判断源） |
| 返回后按钮状态正确 | ✅ 修复（isRomReady 三路径检查） |
| 标签筛选计数正确 | ✅ 修复（免费=requiredLevels空） |
| 卡片标签显示 | ✅ 修复（files.zip加tags+兜底逻辑） |
| 底部3个Tab（收藏夹移走） | ✅ PASS |
| 整体上下滚动 | ✅ PASS |
| 分页每页20个 | ✅ PASS |

---

## 已知待优化（下一 CR）
- 卡片图片无 LruCache（滑动重复下载）
- 切「评价」Tab 每次重新请求（应缓存首次结果）
- 评分/评价无防重复提交保护
- renderCards 每次全量重建 View（无 diff）
- isRomReady 在主线程做磁盘 IO（ANR 风险低但应异步）
- 吸顶 Tab 初始化时序（tabThreshold 首次可能为0）
- 截图区等服务端部署 COS 素材管理后接入（见 backlog.md BACKLOG-GC-01）

---

## 代码提交记录

| 端 | commit | 说明 |
|---|---|---|
| 后端 | f29dfc0 | feat(cr13) 评分/评价/点赞/简介 |
| 后端 | 6926daf | test(cr13) 4个单元测试 |
| App | 77264e84 | feat(cr13) GameDetailActivity 完整重写 |
| App | e7faf9bd | fix(cr13) API21兼容 |
| App | ce65630d | fix(cr13) isFinishing保护 |
| App | e70c02fa | fix(cr13) isDestroyed保护 |
| App | 2757b541 | fix(cr13/gamehall) 真机验收全量修复 |

分支：后端 `feature/license-phase2`，App `feature/license-phase2b`


---

## 2026-10-05 后续优化（游戏大厅 UI 改版 + 搜索功能）

### 游戏大厅 UI 改版

**后端变更：**
- `ListPublicGames` 支持 `page/pageSize` 分页参数，返回 `PublicGameListResult{total, page, page_size, items}`
- 支持 `?q=关键词` 模糊搜索（display_name / rom_name / tag name 子查询）
- `SeedDefaultTags` 改为只 seed `majiang` 标签，删除 free/vip/fight 等默认标签
- commit: `dbd7e04`（分页）、`1b90ff7`（搜索 q 参数）、`ee7c3b2`（seed 清理）

**App 端变更：**
- 顶栏：独立白底 topbar（去掉 Banner 内品牌浮层），「街机时代社区」SpannableString（时代粉红），右侧三态 chip
- Banner：16:9 圆角（`setClipToOutline` + `GradientDrawable` `dp(12)` 圆角），左右 margin `dp(10)`
- 公告栏：点击跳消息 Tab（`switchToTab(1)`）
- 标签区：FlowLayout 折叠容器（默认2行），底部操作行（更多▾/收起▴ + 🔍 搜索图标），搜索栏默认 GONE
- 搜索：点 🔍 展开，支持本地/联网双模式，联网 5 秒超时 fallback 本地 + Toast 提示，历史浮层（SP 最多15条）
- 游戏详情页标签：异步加载 tagNameMap 显示中文名
- 分页：数字分页控件（1/2/3…），卡片上下各一排
- commit: `15a78e16`（topbar/分页/标签中文名）、`58bc7bc8`（搜索）、后续若干修复

**原型文档：**
- `C:\data\run\hermes\proto\mjarch_home_v2.html`（首页，含搜索交互）
- `C:\data\run\hermes\proto\mjarch_game_detail_ui.html`（详情页，7态状态机）
