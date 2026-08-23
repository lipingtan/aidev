# NOVA ARCADE 服务端设计（M3+ 启用）

> 定位：客户端本地优先——**没有服务端时盒子功能完整可用**。服务端只负责
> 云评价、个性化推荐、支付验签、排行榜、云存档、目录下发。
> 启用前客户端零改动接口预留（EventBus/DB 已隔离数据出口）。

---

## 1. 技术选型与部署形态

| 项 | 方案A（快起步） | 方案B（自控，推荐正式期） |
|---|---|---|
| 形态 | Supabase / Firebase | Go(Gin) + PostgreSQL + Redis |
| 认证 | 自带 | JWT：设备游客 token → 绑定手机/Google 升级 |
| 存储 | 自带 | 对象存储(S3/OSS) + CDN（DLC/截图） |
| 部署 | 托管 | docker-compose: api + pg + redis, 前置 LB, 水平扩 api |

初期量级（<10w DAU）单机 2C4G 足够；推荐/榜单读多写少，Redis 缓存扛住。

---

## 2. API 一览（REST /v1, JSON, Bearer JWT）

| Method | Path | 说明 | 鉴权 |
|---|---|---|---|
| POST | /v1/auth/device | 设备注册，换游客 JWT | 无 |
| POST | /v1/auth/bind | 绑定手机/Google，升级账号 | JWT |
| GET | /v1/catalog?since={ver} | **增量**游戏目录+编辑配置(带ETag) | 无 |
| GET | /v1/games/:gid | 详情+评分聚合(全部/近30天) | 无 |
| GET | /v1/games/:gid/reviews?sort&cursor | 评价分页(sort=useful/new/good/bad) | 无 |
| POST | /v1/games/:gid/reviews | 提交评价(服务端二次校验时长) | JWT |
| PUT/DELETE | /v1/reviews/:rid | 修改/删除自己的评价 | JWT |
| POST | /v1/reviews/:rid/like \| unlike | 点赞（一设备一评一赞） | JWT |
| POST | /v1/reports/reviews/:rid | 举报评价 | JWT |
| POST | /v1/payments/orders | 创建订单(返回渠道预支付单) | JWT |
| POST | /v1/payments/verify | 渠道票据验签（幂等） | JWT |
| POST | /v1/payments/restore | 恢复购买：按账号/设备重放权益 | JWT |
| GET | /v1/recommend?scene=home/after_game&gid= | 推荐列表（可降级客户端本地） | JWT |
| GET/POST | /v1/leaderboards/:gid?period=all/week | 榜单读/成绩提交 | GET无/POST JWT |
| GET/PUT | /v1/sync/save/:gid | 云存档（≤64KB blob，LWW；**仅 PCK/HTML 游戏**，Arcade 游戏 savestate 不走此接口） | JWT |
| POST | /v1/events | 埋点批量上报（≤100条/批；**需用户同意隐私政策后才上报**，未同意时客户端本地缓存，不发送） | JWT |

通用约定：cursor 分页；错误体 `{code, msg}`；版本号自增整数。

---

## 3. 数据库模型（PostgreSQL）

```sql
users(id, kind[guest|bound], phone, provider_id, created_at)
devices(id, user_id, fp, platform, last_seen)         -- 设备指纹
games(gid PK, meta jsonb, version int, updated_at)    -- 目录源数据
game_stats(gid PK, players int, rating_avg numeric, rating_cnt int,
           rating_30d_avg numeric, rating_30d_cnt int, charts_score numeric)
reviews(id, gid, user_id, stars smallint, text, playtime_sec int,
        status[visible|pending|folded|deleted], likes int,
        created_at, edited_at)                        -- 一人一游戏一条 active
review_likes(rid, user_id, PRIMARY KEY(rid,user_id))
orders(order_id PK, gid, user_id, amount_cents, channel, status,
       channel_receipt jsonb, receipt_hash varchar, created_at, paid_at)
       -- receipt_hash: 首次验签时写入 hash(channel_receipt)，幂等校验用
entitlements(user_id, gid, source[purchase|restore|gift], granted_at, PRIMARY KEY(user_id,gid))
records(user_id, gid, playtime_sec, sessions, best, updated_at)
achievements(user_id, aid, unlocked_at, PRIMARY KEY(user_id,aid))
leaderboard_entries(gid, period, user_id, score, submitted_at, PRIMARY KEY(gid,period,user_id))
saves(user_id, gid, data bytea, rev int, updated_at)
       -- 仅用于 PCK/HTML 游戏（≤64KB）；Arcade 游戏 savestate 不在服务端存储
events(id bigserial, user_id, name, props jsonb, ts)  -- 分区表按月
```

索引要点：reviews(gid,status,created_at)、reviews(gid,status,likes desc)、
orders(user_id,status)、leaderboard(gid,period,score desc)。

---

## 4. 关键流程时序

### 4.1 支付验签（幂等）

```
客户端: create_order → 渠道SDK支付成功(拿到 receipt)
客户端: POST /payments/verify {order_id, receipt}
服务端:
  1. order 存在且属于该 user? 否则 403
  2. status==paid?
     → 校验 receipt_hash 是否与首次一致（防御：同 order_id 不同 receipt 的重放攻击）
     → 一致则直接返回成功(幂等)；不一致返回 409（已支付但票据不匹配，告警）
  3. 调渠道服务端验签 API(Google Play Developer API / 华为验签接口)
     失败 → order.status=verify_failed, 返回可重试错误
  4. 通过 → 事务: order.status=paid, receipt_hash=hash(receipt) + upsert entitlements + 发货事件
  5. 返回 {granted:[gid]}
崩溃恢复: 客户端 pending 订单启动时重放 verify；服务端幂等保证不重复发货
退款: 渠道回调(RTDN/webhook) → 撤销 entitlement + status=refunded
```

### 4.2 评价提交（服务端二次防线）

```
POST /reviews
  1. 限频: 同设备当日 ≤5 条; 同 user 同 gid 仅 1 条 active(覆盖更新)
  2. 时长校验: records.playtime_sec ≥ 600（防本地伪造, 偏差容忍）
  3. 文本: 敏感词表 + 长度 ≤500 → 命中 → status=pending 人工审核队列
  4. 落库 status=visible(默认) → 异步更新 game_stats:
     增量重算 avg/cnt(全量) + 近30天窗口; 写 Redis 缓存失效
  5. 客户端本地评价(status=local)此后拉取对齐 → 转 published
```

### 4.3 目录增量下发

```
GET /catalog?since=43
  version 是服务端全局单调递增整数（类 Lamport clock）：
    - 每次任意 gid 元数据更新时 +1（全局一个序列，非 per-gid）
    - 客户端携带本地最大已知 version（如 43）
    - 服务端返回 {version:47, changed:[所有 version>43 的 meta...], removed:[gids]}
    - 无变化 → 304(ETag)
  客户端只需一个数字记住同步位置，简单可靠；代价是单 gid 更新会触发所有客户端拉取 diff
  （目录规模 <1k 游戏时 diff payload 通常 <10KB，可接受）
用途: Banner/精选/新游戏/价格调整/DLC地址 全走这张表, 不发版可运营
DLC 下载: meta.dlc_url 指向 CDN; 响应头带 sha256, 客户端下载后校验
```

### 4.4 街机内容包下发（M5+，见 runtime-design §3.4）

```
资源类型新增:
  core   : myosd-0.288 核心包(libmain.so + gdextension, 80~150MB, 版本锁)【待按 v2 调整：libMAME4droid.so 按 ABI 分发，见 mame-godot-plugin/DESIGN.md】
           /catalog 返回 {type:"core", version:"0.288", cdn_url, sha256, required:true}
           客户端每次启动 CoreManager.check_installed("0.288") 做版本比对
  arcade : 街机内容包
           /catalog 返回 gid 级 manifest:
             {
               gid, version,
               resources: {
                 core: {...},
                 arcade: {
                   type: "arcade",
                   files: [
                     {path: "pacman.zip", sha256: "...", cdn: "https://...", sig: "..."},
                     {path: "pac-man2.zip", sha256: "...", cdn: "...", sig: "..."}
                   ],
                   meta_json: {path: "meta.json", sha256: "...", cdn: "...", sig: "..."}
                 }
               }
             }
           每个 zip 独立签名；客户端逐文件校验 → 落盘
           一个 gid 的 files[] 可能包含：
             - 每 ROM 一个 zip（多机合一，如 84 合 1）
             - Parent zip + Clone zip（clone 的 parent 声明在 meta.json 中）
             - BIOS zip（CPS1/NeoGeo 等系统级 BIOS）
           下载/校验/安装逻辑见 runtime-design §3.4 "下载与校验流程"
下发: 客户端按 gid 落盘到 user://games/<gid>/roms/<path>, 已存在文件跳过(同 mjarch3 copyGameFiles)
校验: 每文件 sha256 → 签名校验 → MAME 原生 romset 校验(CRC/SHA1) → 驱动存在性
用户自导入 ROM 不经过服务端(客户端本地扫描 user://roms/arom*/)
```

### 4.5 推荐（服务端版）

```
GET /recommend?scene=home
召回: ItemCF(玩了A也玩B, 离线算) ∪ 标签向量近邻 ∪ 热度兜底 → 200条
排序: 轻量GBDT/逻辑回归(特征: tag匹配数/价格档/历史CTR/新鲜度)
重排: 多样性(同category≤2) / 已拥有过滤 / 曝光频控(7日) / 新游扶持
降级: 服务端异常 → 客户端继续用本地 Recommender(接口已对齐)
冷启动用户 → 编辑精选 + 热榜
```

---

## 5. 聚合与缓存策略

| 数据 | 计算 | 缓存 |
|---|---|---|
| 评分聚合 | 新评价触发增量重算(精确) + 每日全量校准 job | Redis `game:{gid}:agg` TTL 1h |
| 榜单(热门榜) | 每小时 job 重算 charts_score | Redis list, 首页直接读 |
| 热搜词 | 输入→点击日志日聚合 | Redis zset |
| 评价列表 | 实时查(带索引) | 首页高频位可缓存 60s |

---

## 6. 防刷与限流

```
网关层 : IP 限流(令牌桶, nginx/redis) / JWT 必检
业务层 : 评价——设备+账号双维度限频; 时长门槛双端校验
         点赞——需登录, 一人一评一赞, 日上限200
         榜单——成绩提交会话签名 + 区间合法性(防999999) + 异常值抽样回放
内容层 : 敏感词(政治/色情/广告) → pending 队列; 举报≥3 自动折叠待审
设备层 : 指纹聚类识别工作室刷量 → 折叠其全部评价
```

---

## 7. 云存档与账号体系

```
游客: 首启 POST /auth/device {fp} → guest JWT + uid（免注册, 转化无损）
绑定: 手机号验证码 / Google → 账号合并(游客记录并入), 可多设备

云存档范围（分游戏类型）:
  PCK/HTML 游戏: 游戏退出时若登录 → 后台 PUT /sync/save（整包 blob, ≤64KB, LWW）
  Arcade 游戏:   savestate 文件体积可达数MB，不走云端存档；
                 savestate / nvram / cfg 只保存在客户端本地（user://games/<gid>/）；
                 云端仅同步轻量数据：best_score / total_playtime（由 DB.records 随
                 leaderboard 接口上报，不走 /sync/save）。
  多端冲突(PCK/HTML): rev 高者胜(LWW), rev 相同保留 updated_at 新者

隐私合规（国内上架）:
  - 首次启动展示隐私政策弹窗，用户同意后才可上报埋点（POST /events）；
    未同意前 Analytics 仅本地缓存 JSONL，不发送；用户撤销同意后清空缓存且停止上报。
  - 埋点上报走 WiFi+充电时优先（M3+），每批 ≤100 条；
    未同意隐私前 app_open 等基础事件仅记录本地，不纳入上报队列。
  - 隐私政策页 + 未成年模式开关（游玩时长提醒/付费限额，国内上架必需）
```

---

## 8. 观测与运维

```
日志  : 结构化 JSON log(request_id 贯穿) → ELK/Loki
指标  : QPS / P99延迟 / 错误率 / 支付成功率 / 验签失败率(告警)
追踪  : Sentry(api + 未来客户端)
备份  : pg 每日全量+WAL; 对象存储版本化
灰度  : catalog 加 gray_uids 字段 → 新游戏/价格灰度放量
```

---

## 9. 与客户端的对接点回顾（已在客户端设计中预留）

| 客户端模块 | 服务端对应 | 切换开关 |
|---|---|---|
| Registry.reload | /catalog 增量 | 配置 `use_cloud_catalog` |
| ReviewStore(local) | /reviews 系列 | status: local→published |
| PayService(Mock) | /payments 系列 | 配置 `pay_channel` |
| Recommender(本地) | /recommend | 接口同构，异常自动降级 |
| DB.records | /leaderboards /sync/save | 登录后后台静默 |
| ArcadeRunner（核心+内容包） | core/arcade 资源类型 + 签名清单（§4.4） | 配置 `arcade_enabled` |
