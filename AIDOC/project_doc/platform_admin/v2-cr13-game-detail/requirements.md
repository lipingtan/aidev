# CR13 — 游戏详情页增强（评分 / 评价 / 简介）

## 背景

CR12 已完成游戏大厅基础功能（游戏列表、标签筛选、Banner 轮播、下载/播放状态机）。
CR13 在此基础上补充游戏详情页的内容层：评分、评价、游戏简介，并将 App 端详情页 UI
按原型 `mjarch_game_detail_ui.html` 重写为完整沉浸式布局。

原型文件：`C:/data/run/hermes/proto/mjarch_game_detail_ui.html`

---

## 功能需求

### FR-1 游戏简介字段

- `games` 表新增 `description TEXT` 字段
- 管理端 `Games.vue` 表单加「简介」文本域（可选填）
- `PublicGameItem` 接口返回 `description` 字段（omitempty，空时不返回）
- App 详情页「简介」Tab 展示，超过 3 行折叠，点击展开

### FR-2 下载次数统计

- `games` 表新增 `download_count BIGINT DEFAULT 0` 字段
- 下载计数触发时机：**用户调用下载 URL 接口（`GET /api/v1/user/license/games/:rom/download-urls`）时后端 +1**，即「请求下载」即计入，不依赖 App 端回调（App 端无法保证回调可靠性）
- App 详情页「评分行」展示下载次数（格式化：≥10000 显示「x.x万」，≥1000 显示「x.xk」，否则直接显示数字）
- 注意：`download_count` 是冗余写回字段，在 `GetDownloadURLs` 服务方法内 `UPDATE games SET download_count = download_count + 1` 即可

### FR-3 游戏评分（`game_scores` 子表）

**数据结构：**
```
game_scores(id, tenant_id, game_id, user_id, score INT 1-5, created_at, updated_at)
唯一索引：(tenant_id, game_id, user_id)  ← 一人一分，upsert
```

`games` 表新增冗余字段（避免实时聚合）：
- `avg_score DECIMAL(3,1)` — 平均分（每次 upsert 评分后重算写回）
- `score_count INT DEFAULT 0` — 评分人数

**接口：**
- `POST /api/v1/user/license/games/:rom/score` — 需登录，body `{score: 1-5}`；后端先按 `rom_name` 查 `game_id`，再 upsert `game_scores`，最后重算并写回 `games.avg_score/score_count`
- 返回：`{avg_score: 4.7, score_count: 1285}`

**权限：** 未登录 → 返回 `40100 ErrCodeLoginRequired`

### FR-4 游戏评价（`game_reviews` 子表）

**数据结构：**
```
game_reviews(id, tenant_id, game_id, user_id, nickname VARCHAR(50),
             content TEXT, score INT 1-5, like_count INT DEFAULT 0,
             created_at, updated_at, deleted_at)
索引：(tenant_id, game_id)
```

防重复点赞子表：
```
game_review_likes(id, review_id, user_id)
唯一索引：(review_id, user_id)
```

**接口：**
- `GET  /api/v1/user/public/license/games/:rom/reviews?page=1&page_size=10` — 公开接口；`is_liked` 字段：需登录时根据 JWT token 中的 user_id 查点赞表，未登录时全部返回 `false`（`current_user_id=0` 直接跳过点赞查询）
- `POST /api/v1/user/license/games/:rom/reviews` — 需登录，body `{content, score, nickname}`；`nickname` 字段由 App 从本地登录态读取后传入
- `POST /api/v1/user/license/reviews/:id/like` — 需登录，防重复（唯一索引冲突返回「已点赞过」）

**排序：** 点赞数 DESC，创建时间 DESC

**权限：** 发评价/点赞未登录 → `40100 ErrCodeLoginRequired`

### FR-5 App 端详情页 UI 重写（`GameDetailActivity`）

按原型 `mjarch_game_detail_ui.html` 实现，程序化布局：

**布局结构（从上到下）：**
1. 沉浸式头图区（高 230dp）
   - 封面图居中（118×158dp，圆角14dp，阴影）
   - 渐变背景（深紫→背景色）
   - 左上返回按钮，右上分享按钮（圆形半透明）
   - VIP badge（右上角，仅 VIP 游戏显示）

2. 标题区
   - 游戏名（21sp，粗体）
   - 副标题行：`romName · 分类 · 需要VIP会员`（12.5sp，灰色）
   - 右侧：收藏按钮（★/☆）；**分享功能本 CR 不实现，按钮保留占位但点击无操作**

3. 评分/大小/下载量 三列统计行（App Store 式）
   - 评分：`avg_score ★`，副标题「N人评分」
   - 大小：`xx.x MB`，副标题「ROM体积」
   - 下载量：`x.x万`，副标题「下载次数」

4. 首屏通栏主按钮（高 46dp）+ 删除按钮（左侧 46×46dp）
   - 按状态机显示文案（下载/暂停/继续/校验中/开始游戏/开通会员/登录下载）
   - **所有操作先检测登录状态，未登录跳 `LoginActivity`**
   - **VIP 游戏未开通会员跳 `PurchaseActivity`**

5. 截图横滑区（暂用占位图或封面图重复展示）

6. Tab 区（简介 / 评分 / 评价 N）
   - **简介 Tab**：description 文本（3行折叠），标签行（tags），信息表格（ROM文件/依赖/MD5）
   - **评分 Tab**：大评分数字 + 星级分布条 + 「为这个游戏打分」交互星星（需登录）
   - **评价 Tab**：评价列表（头像缩写/昵称/星级/时间/内容/点赞），底部「写评价」按钮（需登录）；Tab 标题上的 N 来自第一次请求 `ListReviews` 时的 `total` 字段，初始显示「评价」

7. 吸底主按钮（ScrollView 滚动超过首屏后浮现，与首屏按钮状态同步）

**权限检测统一逻辑（所有需要登录的操作）：**
```java
if (!MembershipStorage.isLoggedIn()) {
    startActivity(new Intent(ctx, LoginActivity.class));
    return;
}
```

---

## 技术要点

- `avg_score` 冗余写回：每次 upsert 评分后在事务内 `AVG(score)` 重算，避免实时聚合影响性能
- `like_count` 冗余写回：点赞时 `UPDATE game_reviews SET like_count = like_count + 1`，防重复用 `game_review_likes` 唯一索引
- App 端评分星星交互：5个 TextView，点击第 N 个 → 前 N 个高亮，调用评分接口
- 吸底按钮：使用 `NestedScrollView` 包裹内容区域，`setOnScrollChangeListener` 监听；吸底栏触发阈值动态计算（头图 230dp + 标题区约 80dp + 统计行约 60dp + 标签行约 50dp + 按钮行 60dp = 约 480dp，取 `headerHeight` 运行时实际测量值），不硬编码
- 截图区：当前版本用封面图或占位色块，后续 CR 再补截图上传管理

---

## 不在本 CR 范围

- 评价回复功能
- 管理端评价审核/删除
- 截图上传与管理（后续 CR）
- 举报评价

---

## 状态

**待开发**（CR12 完成后接续）


---


---

## 附录：UI 原型精读（基于 mjarch_game_detail_ui.html）

原型文件：`C:/data/run/hermes/proto/mjarch_game_detail_ui.html`（副本：`prototype.html`）

---

### 整体风格

**亮色主题**，参考 App Store / TapTap / Google Play 游戏详情页业界最佳实践。

| 设计变量 | 色值 | 用途 |
|---------|------|------|
| `--bg` | `#f7f8fc` | 页面背景（浅灰白） |
| `--card` | `#ffffff` | 卡片背景（纯白） |
| `--card2` | `#eef1f8` | 次要卡片背景 |
| `--txt` | `#2a3045` | 主文字（深蓝灰） |
| `--sub` | `#98a0b8` | 副文字（灰色） |
| `--line` | `#e3e7f2` | 分割线（浅蓝灰） |
| `--gold` | `#ffd75e` | 金色（下划线/填充） |
| `--gold-deep` | `#b8860b` | 深金（评分数字/选中态） |
| `--blue` | `#4da3ff` | 蓝色（链接/操作） |
| `--green` | `#4ade80` | 绿色（完成/开始游戏） |

字体：系统默认（-apple-system / PingFang SC / Microsoft YaHei）

---

### 页面布局（从上到下完整描述）

**① 沉浸式头图区（高 230dp）**

- 背景：淡紫渐变（`#e8def8 0% → #f0eafc 55% → 页面背景色`），覆盖整个区域
- 封面图：水平居中，距顶 36dp，尺寸 118×158dp，圆角 14dp，深紫渐变占位背景，外有阴影（`rgba(74,42,106,.35) 0 8dp 24dp`）+ 白色半透明细边框
- 返回按钮（左上）：38×38dp 白色圆形，轻阴影，「←」字符，字色 `#5a6280`
- 更多/分享按钮（右上）：同规格，内含「⋯」字符
- VIP badge：仅 VIP 游戏显示，定位 top:56dp right:20dp，金色渐变圆角标签（`#ffd75e → #ffb84d`，20dp 圆角），深棕色文字「VIP」，11sp 粗体，字间距 0.5

**② 标题区（padding: 14dp 20dp 10dp）**

- 左：游戏名（21sp 粗体，`#2a3045`）
- 左下：副标题行（12.5sp，`#98a0b8`），格式 `romName · 分类名 · 需要VIP会员`，「·」点分隔（各6dp 间距）
- 右：竖向排列两个小按钮（gap 14dp）
  - 「★ 已收藏」：选中态星星 `#b8860b`，未收藏显示「☆ 收藏」；点击切换状态
  - 「↗ 分享」：灰色

**③ 三列统计行（App Store 式，margin: 12dp 20dp）**

- 上下各一条 `#e3e7f2` 分割线，内 padding 12dp
- 三列等宽（flex:1），相邻列间有竖向分割线（高度 70%，居中）
- 第1列 - 评分：数值金棕色 `#b8860b`，16sp 粗体，后跟小 ★；副标题「N人评分」11sp 灰
- 第2列 - 大小：数值深色，16sp 粗体，「xx.x MB」；副标题「ROM体积」
- 第3列 - 下载量：数值深色，16sp 粗体（万/千格式化）；副标题「下载次数」

**④ 分类标签行（统计行下方，padding: 10dp 20dp 2dp）**

- 来自游戏 `tags` 字段，横向自动换行（flex-wrap）
- 每个 tag：白底（`#ffffff`），细边框（`#e3e7f2`），圆角 14dp，字色 `#6a7290`，11.5sp，padding 4dp 10dp

**⑤ 首屏主按钮行（padding: 4dp 20dp 0）**

- 左：46×46dp 方圆角（14dp）白色删除按钮（🗑），灰色边框，`#98a0b8` 字色；**已下载才可用，未下载置灰**
- 右：通栏主按钮（flex:1，高 46dp，圆角 14dp），主标题 + 副标题（10.5sp，opacity 0.85）两行

按钮状态：

| 状态 | 主文字 | 副文字 | 背景 | 文字色 |
|------|--------|--------|------|--------|
| 未登录 | 登录下载 | 登录后可下载 | 紫色渐变 | 白 |
| 免费·已登录 | 下载 | 免费 · xx MB | 蓝色渐变 `#4da3ff→#7dc4ff` | 深蓝 |
| VIP·未开通 | 开通会员 | VIP 专享 | 金色渐变 `#ffd75e→#ffb84d` | 深棕 `#33302a` |
| 下载中 | 暂停 | 下载中 xx% · xx/xx MB | 蓝色渐变 + 白色半透明进度条叠加 | 深蓝 |
| 已暂停 | 继续 | 已暂停 · xx% | 蓝色淡化渐变 | 深蓝 |
| 校验中 | 校验中 | MD5 校验… | 蓝色 + 满进度条 | 深蓝 |
| 已就绪 | 开始游戏 | 已就绪 | 绿色渐变 `#4ade80→#22b365` | 深绿 `#0a2a18` |

进度条实现：白色半透明宽条（`rgba(255,255,255,.35)`）绝对定位叠加在按钮内左侧，右侧有 3dp 白色亮边

**⑥ 截图横滑区（margin-top: 18dp）**

- 标题行：「预览」（16sp 粗体）+ 右侧「查看全部」（12.5sp 蓝色），左右 padding 20dp
- 横向滚动，snap 对齐，间距 10dp，左右 padding 20dp，隐藏滚动条
- 每张截图：218×128dp，圆角 10dp，白到浅蓝灰渐变背景，`#e3e7f2` 细边框
- 占位内容：大 emoji 居中（26sp 灰色）+ 底部半透明白色标题条（白底 `#ffffffdd`，细顶边框）

**⑦ Tab 栏（sticky top:0，background: `#f7f8fc`，z-index:5）**

- 三个等宽 Tab：「简介」/ 「评分」/ 「评价 N」
- 未选中：14sp，`#98a0b8` 灰色，无下划线
- 选中：14sp，`#b8860b` 金棕色，粗体，2dp `#ffd75e` 金色下划线
- Tab 栏下方有圆角 `border-radius: 0 0 8px 8px`

**⑧ 简介面板（Tab 1）**

- 正文：14sp，行高 1.75，`#4a5170`；超 3 行折叠，末尾蓝色「展开」文字
- 信息表格（margin-top: 16dp，13sp）：
  - 标签列：92dp 固定宽，`#98a0b8` 灰色
  - 值列：`#2a3045` 深色
  - 每行底部分割线，padding 9dp 上下
  - 字段：ROM 文件（名+大小）/ 依赖 ROM（无则「无」）/ MD5 校验（截断）/ 操作方式 / 更新时间

**⑨ 评分面板（Tab 2）**

- 顶部横向布局（gap 22dp）：
  - 左：44sp 粗体金棕色评分数字（`avg_score`）
  - 中：5颗金星（`#b8860b`，16sp，字间距 2dp）+ 「共 N 人评分」（12sp 灰）
  - 右：5→1 星分布条（flex列排列，gap 5dp，每行：星数标签 + 金黄渐变填充条（`#ffd75e→#ffb84d`），高 6dp，圆角 3dp，灰底 `#e8ecf5`）
- 下方白色交互卡片（圆角 12dp，`#e3e7f2` 边框，padding 14dp）：
  - 「为这个游戏打个分吧」（13.5sp）
  - 5颗点击星星（26sp，间距 6dp）：未选中 `#dfe4f0`，选中 `#b8860b`
  - **点击需登录**，未登录弹跳转登录

**⑩ 评价面板（Tab 3）**

- 评价卡片，底部 `#e3e7f2` 分割线（最后一条无），padding 14dp 上下
- 每条结构：
  - 头部行（flex，gap 10dp）：圆形头像（34×34dp，昵称首字，随机渐变背景，白色 14sp 粗体）+ 昵称（13.5sp 粗体 `#2a3045`）+ 星级（`#b8860b` 11sp）+ 右对齐日期（11sp `#98a0b8`，格式 MM-DD）
  - 正文（13.5sp，行高 1.7，`#4a5170`，margin 9dp 上下）
  - 底部操作：「👍 有用 N」（已点赞 `#4da3ff` 蓝色，未点赞灰色）、「💬 回复」（灰色，本 CR 不实现）
- 列表底部：「✍️ 写评价」通栏按钮（白底，圆角 12dp，`#e3e7f2` 边框，13.5sp，**需登录**）

**⑪ 吸底主按钮区（scrollY > 420dp 后滑入）**

- 白色半透明背景（`rgba(255,255,255,.97)`）+ backdrop-filter blur + 顶部细分割线 + 轻阴影（`0 -2dp 12dp rgba(40,60,120,.08)`）
- padding 12dp，底部兼容 safe-area-inset-bottom（刘海屏）
- 布局与首屏主按钮行完全一致：🗑 删除（左）+ 通栏主按钮（右）
- **与首屏主按钮实时同步**（同一状态机驱动，两处同时更新文字/颜色/进度条）
- 动画：`transform: translateY(110%)` → `translateY(0)`，过渡 0.28s ease

---

### 交互行为说明

- **收藏**：点击切换 ★/☆，调用收藏接口，Toast 提示
- **评分**：点击第 N 颗星 → 前 N 颗高亮，调用评分接口；**需登录**
- **写评价**：点击「✍️ 写评价」→ 跳转评价撰写页（或弹底部 Sheet）；**需登录**
- **点赞**：点击「👍 有用」→ 切换蓝色高亮，计数+1；**需登录**；防重复点赞
- **主按钮**：按状态机分发（下载/暂停/继续/开始/开通会员/登录）；**所有操作需登录**
- **Tab 切换**：点击 Tab → 对应面板 display:block，其余 display:none
- **吸底栏**：ScrollView 监听滚动，超过 420dp 后滑入，回顶后滑出

// ─── CR13 测试：评分/评价/描述 ────────────────────────────────────────────────

// TestScoreGame 评分接口：upsert + avg 重算
func TestScoreGame(t *testing.T) {
	db := setupGameTestDB(t)
	db.AutoMigrate(&licenseModels.GameScore{})
	gameSvc := service.NewGameService(db)
	contentSvc := service.NewGameContentService(db)

	// 先建一个游戏
	g, err := gameSvc.CreateGame(1, service.GameInput{RomName: "test_score", DisplayName: "评分测试"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	// 评分超范围
	_, err = contentSvc.ScoreGame(1, g.ID, 100, service.ScoreInput{Score: 6})
	if err == nil {
		t.Fatal("期望评分超范围返回错误")
	}

	// 正常评分
	result, err := contentSvc.ScoreGame(1, g.ID, 100, service.ScoreInput{Score: 4})
	if err != nil {
		t.Fatalf("ScoreGame: %v", err)
	}
	if result.ScoreCount != 1 {
		t.Errorf("期望 score_count=1，实际 %d", result.ScoreCount)
	}

	// upsert：同一用户改分
	result2, err := contentSvc.ScoreGame(1, g.ID, 100, service.ScoreInput{Score: 5})
	if err != nil {
		t.Fatalf("ScoreGame upsert: %v", err)
	}
	if result2.ScoreCount != 1 {
		t.Errorf("upsert 后 score_count 应仍为 1，实际 %d", result2.ScoreCount)
	}
	if result2.AvgScore != 5.0 {
		t.Errorf("upsert 后 avg_score 应为 5.0，实际 %.1f", result2.AvgScore)
	}

	// 两个用户评分
	contentSvc.ScoreGame(1, g.ID, 101, service.ScoreInput{Score: 3})
	result3, _ := contentSvc.ScoreGame(1, g.ID, 102, service.ScoreInput{Score: 4})
	if result3.ScoreCount < 2 {
		t.Errorf("多用户评分后 score_count 应>=2，实际 %d", result3.ScoreCount)
	}
}

// TestCreateAndListReviews 发表+查询评价
func TestCreateAndListReviews(t *testing.T) {
	db := setupGameTestDB(t)
	db.AutoMigrate(&licenseModels.GameReview{}, &licenseModels.GameReviewLike{})
	gameSvc := service.NewGameService(db)
	contentSvc := service.NewGameContentService(db)

	g, _ := gameSvc.CreateGame(1, service.GameInput{RomName: "test_review", DisplayName: "评价测试"})

	// 空内容应报错
	err := contentSvc.CreateReview(1, g.ID, 200, service.ReviewInput{Content: "", Score: 5})
	if err == nil {
		t.Fatal("期望空评价内容返回错误")
	}

	// 正常发评价
	err = contentSvc.CreateReview(1, g.ID, 200, service.ReviewInput{
		Content: "非常好玩！", Score: 5, Nickname: "测试用户",
	})
	if err != nil {
		t.Fatalf("CreateReview: %v", err)
	}

	// 查询评价
	result, err := contentSvc.ListReviews(1, g.ID, 200, 1, 10)
	if err != nil {
		t.Fatalf("ListReviews: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("期望 1 条评价，实际 %d", result.Total)
	}
	if result.Items[0].Content != "非常好玩！" {
		t.Errorf("评价内容不匹配: %s", result.Items[0].Content)
	}
}

// TestLikeReview 点赞防重复
func TestLikeReview(t *testing.T) {
	db := setupGameTestDB(t)
	db.AutoMigrate(&licenseModels.GameReview{}, &licenseModels.GameReviewLike{})
	gameSvc := service.NewGameService(db)
	contentSvc := service.NewGameContentService(db)

	g, _ := gameSvc.CreateGame(1, service.GameInput{RomName: "test_like", DisplayName: "点赞测试"})
	contentSvc.CreateReview(1, g.ID, 300, service.ReviewInput{Content: "好游戏", Score: 5, Nickname: "用户A"})

	result, _ := contentSvc.ListReviews(1, g.ID, 0, 1, 10)
	if len(result.Items) == 0 {
		t.Fatal("评价列表为空")
	}
	reviewID := result.Items[0].ID

	// 点赞
	err := contentSvc.LikeReview(1, reviewID, 300)
	if err != nil {
		t.Fatalf("LikeReview: %v", err)
	}

	// 重复点赞应报错
	err2 := contentSvc.LikeReview(1, reviewID, 300)
	if err2 == nil {
		t.Fatal("期望重复点赞返回错误")
	}

	// 验证 like_count+1
	result2, _ := contentSvc.ListReviews(1, g.ID, 300, 1, 10)
	if result2.Items[0].LikeCount != 1 {
		t.Errorf("期望 like_count=1，实际 %d", result2.Items[0].LikeCount)
	}
	if !result2.Items[0].IsLiked {
		t.Error("期望 is_liked=true（用户300已点赞）")
	}
}

// TestPublicGameItem_NewFields 公开游戏接口包含 description/download_count 字段
func TestPublicGameItem_NewFields(t *testing.T) {
	db := setupGameTestDB(t)
	db.AutoMigrate(&licenseModels.GameTagRelation{})
	svc := service.NewGameService(db)

	desc := "这是一款经典麻将游戏"
	svc.CreateGame(1, service.GameInput{
		RomName:     "test_pub",
		DisplayName: "公开接口测试",
		Description: &desc,
	})

	items := svc.ListPublicGames(1)
	if len(items) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(items))
	}
	if items[0].Description == nil || *items[0].Description != desc {
		t.Errorf("Description 期望 %q，实际 %v", desc, items[0].Description)
	}
}


// ─── CR13 测试：评分/评价 ──────────────────────────────────────────────────────

// TestScoreGame 评分接口：upsert + avg 重算
func TestScoreGame(t *testing.T) {
	db := setupGameTestDB(t)
	db.AutoMigrate(&licenseModels.GameScore{})
	gameSvc := service.NewGameService(db)
	contentSvc := service.NewGameContentService(db)

	g, err := gameSvc.CreateGame(1, service.GameInput{RomName: "test_score", DisplayName: "评分测试"})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	// 评分超范围
	_, err = contentSvc.ScoreGame(1, g.ID, 100, service.ScoreInput{Score: 6})
	if err == nil {
		t.Fatal("期望评分超范围返回错误")
	}

	// 正常评分
	result, err := contentSvc.ScoreGame(1, g.ID, 100, service.ScoreInput{Score: 4})
	if err != nil {
		t.Fatalf("ScoreGame: %v", err)
	}
	if result.ScoreCount != 1 {
		t.Errorf("期望 score_count=1，实际 %d", result.ScoreCount)
	}

	// upsert：同一用户改分
	result2, err := contentSvc.ScoreGame(1, g.ID, 100, service.ScoreInput{Score: 5})
	if err != nil {
		t.Fatalf("ScoreGame upsert: %v", err)
	}
	if result2.ScoreCount != 1 {
		t.Errorf("upsert 后 score_count 应仍为 1，实际 %d", result2.ScoreCount)
	}
	if result2.AvgScore != 5.0 {
		t.Errorf("upsert 后 avg_score 应为 5.0，实际 %.1f", result2.AvgScore)
	}
}

// TestCreateAndListReviews 发表+查询评价
func TestCreateAndListReviews(t *testing.T) {
	db := setupGameTestDB(t)
	db.AutoMigrate(&licenseModels.GameReview{}, &licenseModels.GameReviewLike{})
	gameSvc := service.NewGameService(db)
	contentSvc := service.NewGameContentService(db)

	g, _ := gameSvc.CreateGame(1, service.GameInput{RomName: "test_review", DisplayName: "评价测试"})

	// 空内容应报错
	err := contentSvc.CreateReview(1, g.ID, 200, service.ReviewInput{Content: "", Score: 5})
	if err == nil {
		t.Fatal("期望空评价内容返回错误")
	}

	// 正常发评价
	err = contentSvc.CreateReview(1, g.ID, 200, service.ReviewInput{
		Content: "非常好玩！", Score: 5, Nickname: "测试用户",
	})
	if err != nil {
		t.Fatalf("CreateReview: %v", err)
	}

	// 查询评价
	result, err := contentSvc.ListReviews(1, g.ID, 200, 1, 10)
	if err != nil {
		t.Fatalf("ListReviews: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("期望 1 条评价，实际 %d", result.Total)
	}
	if len(result.Items) == 0 || result.Items[0].Content != "非常好玩！" {
		t.Errorf("评价内容不匹配")
	}
}

// TestLikeReview 点赞防重复
func TestLikeReview(t *testing.T) {
	db := setupGameTestDB(t)
	db.AutoMigrate(&licenseModels.GameReview{}, &licenseModels.GameReviewLike{})
	gameSvc := service.NewGameService(db)
	contentSvc := service.NewGameContentService(db)

	g, _ := gameSvc.CreateGame(1, service.GameInput{RomName: "test_like", DisplayName: "点赞测试"})
	contentSvc.CreateReview(1, g.ID, 300, service.ReviewInput{Content: "好游戏", Score: 5, Nickname: "用户A"})

	result, _ := contentSvc.ListReviews(1, g.ID, 0, 1, 10)
	if len(result.Items) == 0 {
		t.Fatal("评价列表为空")
	}
	reviewID := result.Items[0].ID

	// 点赞
	err := contentSvc.LikeReview(1, reviewID, 300)
	if err != nil {
		t.Fatalf("LikeReview: %v", err)
	}

	// 重复点赞应报错
	err2 := contentSvc.LikeReview(1, reviewID, 300)
	if err2 == nil {
		t.Fatal("期望重复点赞返回错误")
	}

	// 验证 like_count=1 且 is_liked=true
	result2, _ := contentSvc.ListReviews(1, g.ID, 300, 1, 10)
	if result2.Items[0].LikeCount != 1 {
		t.Errorf("期望 like_count=1，实际 %d", result2.Items[0].LikeCount)
	}
	if !result2.Items[0].IsLiked {
		t.Error("期望 is_liked=true（用户300已点赞）")
	}
}

// TestPublicGameItem_Description description 字段在公开接口中正确返回
func TestPublicGameItem_Description(t *testing.T) {
	db := setupGameTestDB(t)
	db.AutoMigrate(&licenseModels.GameTagRelation{})
	svc := service.NewGameService(db)

	desc := "这是一款经典麻将游戏"
	svc.CreateGame(1, service.GameInput{
		RomName:     "test_pub",
		DisplayName: "公开接口测试",
		Description: &desc,
	})

	items := svc.ListPublicGames(1)
	if len(items) != 1 {
		t.Fatalf("期望 1 条，实际 %d", len(items))
	}
	if items[0].Description == nil || *items[0].Description != desc {
		t.Errorf("Description 期望 %q，实际 %v", desc, items[0].Description)
	}
}
