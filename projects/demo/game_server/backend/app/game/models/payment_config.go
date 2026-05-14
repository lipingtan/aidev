package models

import "go-admin/common/models"

// PaymentConfig 支付渠道配置
type PaymentConfig struct {
	models.Model
	GameId              int    `json:"gameId" gorm:"default:0;index;comment:所属游戏ID 0全局"`
	Channel             string `json:"channel" gorm:"size:32;not null;comment:支付渠道"`
	ChannelName         string `json:"channelName" gorm:"size:64;comment:渠道名称"`
	Enabled             int    `json:"enabled" gorm:"default:1;comment:是否启用 1是 2否"`
	Env                 string `json:"env" gorm:"size:32;default:'sandbox';comment:环境"`
	StripePublishableKey string `json:"stripePublishableKey" gorm:"size:256;comment:Stripe公钥"`
	StripeSecretKey     string `json:"stripeSecretKey" gorm:"size:256;comment:Stripe私钥"`
	StripeWebhookSecret string `json:"stripeWebhookSecret" gorm:"size:256;comment:Stripe Webhook密钥"`
	AlipayAppId         string `json:"alipayAppId" gorm:"size:64;comment:支付宝AppID"`
	AlipayPrivateKey    string `json:"alipayPrivateKey" gorm:"type:text;comment:支付宝私钥"`
	AlipayPublicKey     string `json:"alipayPublicKey" gorm:"type:text;comment:支付宝公钥"`
	AlipayNotifyUrl     string `json:"alipayNotifyUrl" gorm:"size:256;comment:支付宝回调地址"`
	WechatAppId         string `json:"wechatAppId" gorm:"size:64;comment:微信AppID"`
	WechatMchId         string `json:"wechatMchId" gorm:"size:64;comment:微信商户号"`
	WechatApiKey        string `json:"wechatApiKey" gorm:"size:256;comment:微信API密钥"`
	WechatNotifyUrl     string `json:"wechatNotifyUrl" gorm:"size:256;comment:微信回调地址"`

	models.ModelTime
	models.ControlBy
	models.TenantBy
}

func (PaymentConfig) TableName() string { return "game_payment_config" }
