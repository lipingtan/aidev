package main

import "time"

// ===== 基础结构（替代 go-admin/common/models） =====

// Model 基础模型（主键）
type Model struct {
	Id int `json:"id" gorm:"primaryKey;autoIncrement"`
}

// ModelTime 时间字段
type ModelTime struct {
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt" gorm:"index"`
}

// ControlBy 操作人字段
type ControlBy struct {
	CreateBy int `json:"createBy" gorm:"default:0"`
	UpdateBy int `json:"updateBy" gorm:"default:0"`
}

// TenantBy 租户字段
type TenantBy struct {
	TenantId int `json:"tenantId" gorm:"default:0;index;comment:租户ID"`
}

// ===== 业务模型 =====

// Game 游戏
type Game struct {
	Model
	Name        string `json:"name" gorm:"size:128;not null;comment:游戏名称"`
	AppKey      string `json:"appKey" gorm:"size:64;uniqueIndex;not null;comment:游戏标识"`
	AppSecret   string `json:"appSecret" gorm:"size:128;not null;comment:API密钥"`
	Description string `json:"description" gorm:"size:512;comment:游戏描述"`
	Icon        string `json:"icon" gorm:"size:256;comment:游戏图标"`
	Status      int    `json:"status" gorm:"default:1;comment:状态 1启用 2禁用"`
	Version     string `json:"version" gorm:"size:32;default:'1.0.0';comment:游戏版本"`
	ModelTime
	ControlBy
	TenantBy
}

func (Game) TableName() string { return "game_game" }

// Dlc DLC 数据模型
type Dlc struct {
	Model
	GameId         int    `json:"gameId" gorm:"not null;index;comment:所属游戏ID"`
	DlcKey         string `json:"dlcKey" gorm:"size:64;not null;comment:DLC标识"`
	Name           string `json:"name" gorm:"size:128;not null;comment:DLC名称"`
	Version        string `json:"version" gorm:"size:32;default:'1.0.0';comment:版本号"`
	Description    string `json:"description" gorm:"size:512;comment:描述"`
	FilePath       string `json:"filePath" gorm:"size:256;comment:PCK文件路径"`
	FileSize       int64  `json:"fileSize" gorm:"default:0;comment:文件大小字节"`
	Sha256         string `json:"sha256" gorm:"size:64;comment:SHA256哈希"`
	Price          int    `json:"price" gorm:"default:0;comment:价格分"`
	IsFree         int    `json:"isFree" gorm:"default:1;comment:是否免费 1是 2否"`
	Status         int    `json:"status" gorm:"default:1;comment:状态 1上架 2下架"`
	DownloadCount  int    `json:"downloadCount" gorm:"default:0;comment:下载次数"`
	MinGameVersion string `json:"minGameVersion" gorm:"size:32;default:'0.0.0';comment:最低游戏版本"`
	ModelTime
	ControlBy
	TenantBy
}

func (Dlc) TableName() string { return "game_dlc" }

// Player 游戏玩家
type Player struct {
	Model
	GameId      int        `json:"gameId" gorm:"not null;index;comment:所属游戏ID"`
	Uid         string     `json:"uid" gorm:"size:64;not null;index;comment:玩家UID"`
	Nickname    string     `json:"nickname" gorm:"size:64;comment:玩家昵称"`
	Email       string     `json:"email" gorm:"size:128;comment:邮箱"`
	Platform    string     `json:"platform" gorm:"size:32;comment:登录平台"`
	PlatformId  string     `json:"platformId" gorm:"size:128;comment:平台ID"`
	DeviceId    string     `json:"deviceId" gorm:"size:128;comment:设备ID"`
	Status      int        `json:"status" gorm:"default:1;comment:状态 1正常 2封禁"`
	BanReason   string     `json:"banReason" gorm:"size:256;comment:封禁原因"`
	LastLoginAt *time.Time `json:"lastLoginAt" gorm:"comment:最后登录时间"`
	LastLoginIp string     `json:"lastLoginIp" gorm:"size:64;comment:最后登录IP"`
	ModelTime
	ControlBy
	TenantBy
}

func (Player) TableName() string { return "game_player" }

// Save 玩家存档
type Save struct {
	Model
	GameId      int    `json:"gameId" gorm:"not null;index;comment:所属游戏ID"`
	PlayerId    int    `json:"playerId" gorm:"not null;index;comment:玩家ID"`
	Slot        int    `json:"slot" gorm:"default:0;comment:存档槽位"`
	SaveName    string `json:"saveName" gorm:"size:64;comment:存档名称"`
	Data        string `json:"data" gorm:"type:longtext;comment:存档数据JSON"`
	DataVersion string `json:"dataVersion" gorm:"size:32;comment:存档数据版本"`
	PlayTime    int64  `json:"playTime" gorm:"default:0;comment:游戏内时间秒"`
	Screenshot  string `json:"screenshot" gorm:"size:256;comment:存档截图"`
	ModelTime
	ControlBy
	TenantBy
}

func (Save) TableName() string { return "game_save" }

// Order 支付订单
type Order struct {
	Model
	OrderNo      string     `json:"orderNo" gorm:"size:64;uniqueIndex;not null;comment:订单号"`
	GameId       int        `json:"gameId" gorm:"not null;index;comment:所属游戏ID"`
	PlayerId     int        `json:"playerId" gorm:"not null;index;comment:玩家ID"`
	ProductType  string     `json:"productType" gorm:"size:32;not null;comment:商品类型"`
	ProductId    int        `json:"productId" gorm:"not null;comment:商品ID"`
	ProductName  string     `json:"productName" gorm:"size:128;comment:商品名称"`
	Amount       int        `json:"amount" gorm:"not null;comment:金额分"`
	Currency     string     `json:"currency" gorm:"size:8;default:'CNY';comment:货币"`
	Channel      string     `json:"channel" gorm:"size:32;comment:支付渠道"`
	ThirdOrderNo string     `json:"thirdOrderNo" gorm:"size:128;comment:第三方订单号"`
	Status       string     `json:"status" gorm:"size:32;default:'pending';comment:状态"`
	PaidAt       *time.Time `json:"paidAt" gorm:"comment:支付时间"`
	RefundedAt   *time.Time `json:"refundedAt" gorm:"comment:退款时间"`
	RefundReason string     `json:"refundReason" gorm:"size:256;comment:退款原因"`
	CallbackData string     `json:"callbackData" gorm:"type:text;comment:回调原始数据"`
	ModelTime
	ControlBy
	TenantBy
}

func (Order) TableName() string { return "game_order" }

// PaymentConfig 支付渠道配置
type PaymentConfig struct {
	Model
	GameId               int    `json:"gameId" gorm:"default:0;index;comment:所属游戏ID 0全局"`
	Channel              string `json:"channel" gorm:"size:32;not null;comment:支付渠道"`
	ChannelName          string `json:"channelName" gorm:"size:64;comment:渠道名称"`
	Enabled              int    `json:"enabled" gorm:"default:1;comment:是否启用 1是 2否"`
	Env                  string `json:"env" gorm:"size:32;default:'sandbox';comment:环境"`
	StripePublishableKey string `json:"stripePublishableKey" gorm:"size:256;comment:Stripe公钥"`
	StripeSecretKey      string `json:"stripeSecretKey" gorm:"size:256;comment:Stripe私钥"`
	StripeWebhookSecret  string `json:"stripeWebhookSecret" gorm:"size:256;comment:Stripe Webhook密钥"`
	AlipayAppId          string `json:"alipayAppId" gorm:"size:64;comment:支付宝AppID"`
	AlipayPrivateKey     string `json:"alipayPrivateKey" gorm:"type:text;comment:支付宝私钥"`
	AlipayPublicKey      string `json:"alipayPublicKey" gorm:"type:text;comment:支付宝公钥"`
	AlipayNotifyUrl      string `json:"alipayNotifyUrl" gorm:"size:256;comment:支付宝回调地址"`
	WechatAppId          string `json:"wechatAppId" gorm:"size:64;comment:微信AppID"`
	WechatMchId          string `json:"wechatMchId" gorm:"size:64;comment:微信商户号"`
	WechatApiKey         string `json:"wechatApiKey" gorm:"size:256;comment:微信API密钥"`
	WechatNotifyUrl      string `json:"wechatNotifyUrl" gorm:"size:256;comment:微信回调地址"`
	ModelTime
	ControlBy
	TenantBy
}

func (PaymentConfig) TableName() string { return "game_payment_config" }

// H5Page H5 嵌入页面
type H5Page struct {
	Model
	GameId      int    `json:"gameId" gorm:"not null;index;comment:所属游戏ID"`
	PageKey     string `json:"pageKey" gorm:"size:64;not null;comment:页面标识"`
	Name        string `json:"name" gorm:"size:128;not null;comment:页面名称"`
	PageType    string `json:"pageType" gorm:"size:32;default:'custom';comment:页面类型"`
	Content     string `json:"content" gorm:"type:longtext;comment:页面内容"`
	ExternalUrl string `json:"externalUrl" gorm:"size:256;comment:外部URL"`
	UseExternal int    `json:"useExternal" gorm:"default:2;comment:是否使用外链 1是 2否"`
	Status      int    `json:"status" gorm:"default:1;comment:状态 1启用 2禁用"`
	Remark      string `json:"remark" gorm:"size:256;comment:备注"`
	ModelTime
	ControlBy
	TenantBy
}

func (H5Page) TableName() string { return "game_h5_page" }
