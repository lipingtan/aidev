# CR13 — 游戏详情页增强 设计文档

## 1. 数据库变更

### 1.1 `games` 表新增字段

```sql
ALTER TABLE games
  ADD COLUMN description    TEXT            COMMENT '游戏简介',
  ADD COLUMN download_count BIGINT NOT NULL DEFAULT 0 COMMENT '下载次数（请求下载URL时+1）',
  ADD COLUMN avg_score      DECIMAL(3,1)    COMMENT '平均评分（冗余缓存，评分时重算）',
  ADD COLUMN score_count    INT NOT NULL    DEFAULT 0 COMMENT '评分人数';
```

由 GORM AutoMigrate 自动执行，无需手动 DDL。

### 1.2 新增 `game_scores` 表

```sql
CREATE TABLE game_scores (
  id         BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id  BIGINT NOT NULL,
  game_id    BIGINT NOT NULL,
  user_id    BIGINT NOT NULL,
  score      INT NOT NULL,              -- 1-5
  created_at DATETIME,
  updated_at DATETIME,
  UNIQUE KEY uk_score (tenant_id, game_id, user_id),
  KEY idx_score_game (game_id)
);
```

### 1.3 新增 `game_reviews` 表

```sql
CREATE TABLE game_reviews (
  id         BIGINT AUTO_INCREMENT PRIMARY KEY,
  tenant_id  BIGINT NOT NULL,
  game_id    BIGINT NOT NULL,
  user_id    BIGINT NOT NULL,
  nickname   VARCHAR(50),
  content    TEXT NOT NULL,
  score      INT NOT NULL DEFAULT 5,    -- 1-5
  like_count INT NOT NULL DEFAULT 0,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME,
  KEY idx_review_game (tenant_id, game_id),
  KEY idx_review_user (user_id)
);
```

### 1.4 新增 `game_review_likes` 表（防重复点赞）

```sql
CREATE TABLE game_review_likes (
  id         BIGINT AUTO_INCREMENT PRIMARY KEY,
  review_id  BIGINT NOT NULL,
  user_id    BIGINT NOT NULL,
  created_at DATETIME,
  UNIQUE KEY uk_review_like (review_id, user_id)
);
```

---

## 2. 后端接口设计

### 2.1 路由汇总

| Method | Path | Auth | 说明 |
|--------|------|------|------|
| POST | `/api/v1/user/license/games/:rom/score` | 需登录 | 评分（upsert） |
| GET | `/api/v1/user/public/license/games/:rom/reviews` | 公开 | 评价列表 |
| POST | `/api/v1/user/license/games/:rom/reviews` | 需登录 | 发表评价 |
| POST | `/api/v1/user/license/reviews/:id/like` | 需登录 | 点赞评价 |

注：登录态接口在 `plugin.go` 注册到 `userAuthGroup`（需 JWT），公开接口注册到 `publicGroup`。

### 2.2 下载次数写回（修改现有接口）

`GetDownloadURLs`（`game_content_service.go`）成功返回 URL 后，异步执行：
```go
go func() {
    s.db.Model(&licenseModels.Game{}).
        Where("tenant_id = ? AND rom_name = ?", tenantID, romName).
        UpdateColumn("download_count", gorm.Expr("download_count + 1"))
}()
```
异步执行避免阻塞下载 URL 响应，写失败不影响主流程。

### 2.3 评分接口（POST /games/:rom/score）

**Request:**
```json
{ "score": 4 }
```

**Flow:**
1. 验证 score ∈ [1,5]
2. 按 rom_name + tenant_id 查 game_id（不存在返回 40001）
3. 事务内 upsert `game_scores`
4. 事务内重算：`SELECT AVG(score), COUNT(*) FROM game_scores WHERE game_id=?`
5. 写回 `games.avg_score` 和 `games.score_count`

**Response:**
```json
{ "code": 0, "data": { "avg_score": 4.7, "score_count": 1285 } }
```

### 2.4 评价列表（GET /games/:rom/reviews）

**Query params:** `page=1&page_size=10`

**Flow:**
1. 按 rom_name + tenant_id 查 game_id
2. 查 `game_reviews`，ORDER BY like_count DESC, created_at DESC，分页
3. JWT 可选解析：handler 中尝试从 `Authorization` header 解析 token，成功则得到 user_id，失败/不存在则 user_id=0
4. user_id > 0 时查 `game_review_likes` 获取当前用户点赞集合；user_id=0 时 `is_liked` 全为 false
5. 组装 `ReviewListResult`

**Response:**
```json
{
  "code": 0,
  "data": {
    "total": 326,
    "page": 1,
    "page_size": 10,
    "items": [
      {
        "id": 1,
        "nickname": "连庄小能手",
        "content": "手感还原度很高...",
        "score": 5,
        "like_count": 42,
        "created_at": "10-02",
        "is_liked": false
      }
    ]
  }
}
```

### 2.5 发表评价（POST /games/:rom/reviews）

**Request:**
```json
{ "content": "手感很好", "score": 5, "nickname": "连庄小能手" }
```

**Validation:** content 非空；score ∈ [1,5]；nickname 空则默认「匿名用户」

### 2.6 点赞评价（POST /reviews/:id/like）

- 向 `game_review_likes` 插记录，唯一索引冲突 → 返回「已点赞」错误码
- 成功则 `game_reviews.like_count + 1`
- 事务内执行

### 2.7 PublicGameItem 接口补充字段

`ListPublicGames` 返回的 `PublicGameItem` 补充：
```go
Description   *string  `json:"description,omitempty"`
DownloadCount int64    `json:"download_count"`
AvgScore      *float64 `json:"avg_score,omitempty"`
ScoreCount    int      `json:"score_count"`
```

---

## 3. App 端详情页架构

### 3.1 布局层次（NestedScrollView）

```
Activity (WindowSoftInputMode = adjustResize)
└── CoordinatorLayout (全屏)
    ├── NestedScrollView (占满，内边距底部 = 吸底栏高度)
    │   └── LinearLayout (vertical)
    │       ├── HeroView (230dp，封面/渐变/navbar)
    │       ├── TitleBlock (游戏名+收藏)
    │       ├── StatRow (三列统计)
    │       ├── TagRow (分类标签)
    │       ├── InlineActionRow (删除+主按钮)
    │       ├── ShotsRow (截图横滑)
    │       ├── TabStrip (吸顶)
    │       └── TabContent (简介/评分/评价，高度 WRAP_CONTENT)
    └── BottomActionBar (MATCH_PARENT × 70dp，初始 translationY=+height)
```

### 3.2 吸底按钮触发逻辑

```java
// onCreateView 末尾，等 layout 完成后测量
inlineActionRow.post(() -> {
    int threshold = inlineActionRow.getBottom(); // 首屏主按钮底部 Y
    scrollView.setOnScrollChangeListener((v, scrollX, scrollY, oldX, oldY) -> {
        boolean show = scrollY > threshold;
        bottomBar.animate().translationY(show ? 0 : bottomBar.getHeight()).setDuration(280).start();
    });
});
```

### 3.3 状态机

```
enum State {
    NOT_LOGGED_IN,      // 未登录 → 「登录下载」紫色
    FREE_UNDOWNLOADED,  // 免费·未下载 → 「下载」蓝色
    VIP_NO_MEMBER,      // VIP·未开通会员 → 「开通会员」金色
    DOWNLOADING,        // 下载中 → 「暂停」蓝色+进度条
    PAUSED,             // 已暂停 → 「继续」蓝色淡化
    VERIFYING,          // 校验中 → 「校验中」蓝色+满进度
    READY               // 已就绪 → 「开始游戏」绿色
}
```

状态机在 `determineState()` 中计算，按钮点击统一调用 `onMainButtonClick()`，内部 switch 分发：
- `NOT_LOGGED_IN` → `startActivity(LoginActivity)`
- `VIP_NO_MEMBER` → `startActivity(PurchaseActivity)`
- `FREE_UNDOWNLOADED` → 调用下载接口
- `DOWNLOADING` → 暂停下载
- `PAUSED` → 继续下载
- `READY` → 启动模拟器

### 3.4 API 调用

`GameCenterApi` 新增方法：
```java
ScoreResult scoreGame(String romName, int score)        // POST /games/:rom/score（需 token）
ReviewListResult listReviews(String romName, int page)  // GET /games/:rom/reviews
void createReview(String romName, ReviewInput input)    // POST /games/:rom/reviews（需 token）
void likeReview(long reviewId)                          // POST /reviews/:id/like（需 token）
```

### 3.5 新增 Model

```java
// ScoreResult.java
public class ScoreResult {
    @SerializedName("avg_score")   public double avgScore;
    @SerializedName("score_count") public int scoreCount;
}

// ReviewItem.java
public class ReviewItem {
    @SerializedName("id")          public long id;
    @SerializedName("nickname")    public String nickname;
    @SerializedName("content")     public String content;
    @SerializedName("score")       public int score;
    @SerializedName("like_count")  public int likeCount;
    @SerializedName("created_at")  public String createdAt;
    @SerializedName("is_liked")    public boolean isLiked;
}

// ReviewListResult.java
public class ReviewListResult {
    @SerializedName("total")     public long total;
    @SerializedName("page")      public int page;
    @SerializedName("page_size") public int pageSize;
    @SerializedName("items")     public List<ReviewItem> items;
}
```

---

## 4. 管理端变更（web-admin）

`Games.vue` 表单新增「游戏简介」文本域：
```html
<el-form-item label="游戏简介">
  <el-input type="textarea" :rows="4" v-model="form.description"
    placeholder="游戏介绍文字，App 详情页展示（可选）" />
</el-form-item>
```

`GameInput` / `Game` 接口类型加 `description?: string`。
`PublicGameItem` 接口类型加 `description`, `download_count`, `avg_score`, `score_count`。

---

## 5. 错误码

复用现有 `errors.go`，新增：

| 常量 | code | 说明 |
|------|------|------|
| `ErrCodeReviewAlreadyLiked` | 40301 | 已点赞过该评价 |
| `ErrCodeScoreOutOfRange`    | 40302 | 评分需在1-5之间 |
| `ErrCodeReviewEmpty`        | 40303 | 评价内容不能为空 |
