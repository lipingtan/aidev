class_name SlotConfig
## 幸运轮盘·水果机（CR-8 修订3）：goslot 源项目 1:1 移植
## 格数/顺序/权重/赔付/跑灯节奏全部按 goslot ItemConfig.gd 原值
## 素材：board.png 底图 + rollpic01-08 + lucky01/02 + light1/2 + shine01-04 + wav 音效

## 可押注 8 符号（goslot betPads 倒序原值）：key/赔率（goslot scoreconfig big 分值）
const BET_SYMBOLS := [
	{"key": "bar",        "odd": 100, "label": "BAR"},
	{"key": "seven",      "odd": 40,  "label": "77"},
	{"key": "star",       "odd": 30,  "label": "星星"},
	{"key": "watermelon", "odd": 20,  "label": "西瓜"},
	{"key": "ring",       "odd": 20,  "label": "铃铛"},
	{"key": "lemon",      "odd": 15,  "label": "柠檬"},
	{"key": "orange",     "odd": 10,  "label": "橙子"},
	{"key": "apple",      "odd": 5,   "label": "苹果"},
]

## 24 格（goslot pointsconfig 顺序 1:1）：flag=押注映射 / big=大图标格 / score=押中赔率分值
const RING := [
	{"flag": "orange",     "big": true,  "score": 10},   # 0 bigorange
	{"flag": "ring",       "big": true,  "score": 20},   # 1 bigring
	{"flag": "bar",        "big": false, "score": 100},  # 2 smallbar
	{"flag": "bar",        "big": true,  "score": 500},  # 3 bigbar
	{"flag": "apple",      "big": true,  "score": 5},    # 4 bigapple
	{"flag": "apple",      "big": false, "score": 2},    # 5 smallapple
	{"flag": "lemon",      "big": true,  "score": 15},   # 6 biglemon
	{"flag": "watermelon", "big": true,  "score": 20},   # 7 bigwatermelon
	{"flag": "watermelon", "big": false, "score": 2},    # 8 smallwatermelon
	{"flag": "luckytry",   "big": true,  "score": 0},    # 9 lucky01 大Lucky
	{"flag": "apple",      "big": true,  "score": 5},    # 10 bigapple
	{"flag": "orange",     "big": false, "score": 2},    # 11 smallorange
	{"flag": "orange",     "big": true,  "score": 10},   # 12 bigorange
	{"flag": "ring",       "big": true,  "score": 20},   # 13 bigring
	{"flag": "seven",      "big": false, "score": 2},    # 14 smallseven
	{"flag": "seven",      "big": true,  "score": 40},   # 15 bigseven
	{"flag": "apple",      "big": true,  "score": 5},    # 16 bigapple
	{"flag": "lemon",      "big": false, "score": 2},    # 17 smalllemon
	{"flag": "lemon",      "big": true,  "score": 15},   # 18 biglemon
	{"flag": "star",       "big": true,  "score": 30},   # 19 bigstar
	{"flag": "star",       "big": false, "score": 2},    # 20 smallstar
	{"flag": "luckytry",   "big": true,  "score": 0},    # 21 lucky02 小Lucky
	{"flag": "apple",      "big": true,  "score": 5},    # 22 bigapple
	{"flag": "ring",       "big": false, "score": 2},    # 23 smallring
]

## 每格权重（goslot pointsconfig weight 原值）
const WEIGHTS := [2500, 125, 10, 25, 2500, 2500, 170, 125, 2500, 3000, 2500, 2500, 2500, 125, 2500, 60, 2500, 2500, 170, 85, 2500, 3000, 2500, 2500]

## 大/小 Lucky 格 index（散花触发格）
const LUCKY_IDX_BIG := 9
const LUCKY_IDX_SMALL := 21
## 散花灯珠数（大 5~7 / 小 3~5，用户 2026-09-02 确认）
const BIG_LUCKY_ORBS := [5, 7]
const SMALL_LUCKY_ORBS := [3, 5]

## 跑灯节奏（goslot RoundNormal.startbet 默认参数原值）
const RUN_OPTIONS := {
	"startindex": 0,
	"slowcount": 6,
	"slowwait1": 0.3,
	"slowwait2": 0.3,
	"fastwait1": 0.02,
	"fastwait2": 0.1,
	"slowendwait1": 0.3,
	"slowendwait2": 0.2,
	"slowendcount": 9,
	"delta": 0.04,
	"delta2": 0.1,
}

## 布局（goslot ItemConfig 原值）：startpos + 每边7格 + 步长56
const START_POS := Vector2(26, 110)
const STEPS := 7
const STEPLEN := 56
const ITEM_SCALE := 0.5            # BetItem.tscn bgitem scale
const SMALL_SCALE := 0.65          # small 图标缩放（pointsconfig imagescale）

## 押注：三档注额；初始余额/救济/上限
const BET_STEPS := [1, 5, 10]
const BASE_BALANCE := 100
const RELIEF_AMOUNT := 50
const MAX_BALANCE := 999999
const SAVE_KEY := "balance"

const IMG := "res://games/slot_machine/assets/images/"

static func icon_path(idx: int) -> String:
	var flag: String = RING[idx]["flag"]
	match flag:
		"luckytry": return IMG + ("lucky01.png" if idx == LUCKY_IDX_BIG else "lucky02.png")
	var roll_idx := 1
	match flag:
		"apple": roll_idx = 1
		"orange": roll_idx = 2
		"lemon": roll_idx = 3
		"watermelon": roll_idx = 4
		"ring": roll_idx = 5
		"seven": roll_idx = 6
		"star": roll_idx = 7
		"bar": roll_idx = 8
	return IMG + "rollpic0%d.png" % roll_idx

## 押中单格赔付：押额 × 格 score 分值（goslot getScoreVal 语义）
static func cell_payout(idx: int, bet_amount: int) -> int:
	return int(RING[idx]["score"]) * bet_amount

## 加权随机停位（goslot getRandomIndex 权重算法）
static func weighted_stop(rng: RandomNumberGenerator) -> int:
	var all_weight := 0
	for w in WEIGHTS:
		all_weight += w
	var rnd := rng.randi_range(0, all_weight - 1)
	var acc := 0
	for i in 24:
		acc += WEIGHTS[i]
		if rnd < acc:
			return i
	return 23
