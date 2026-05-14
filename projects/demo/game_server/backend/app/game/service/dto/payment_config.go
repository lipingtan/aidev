package dto

import (
	"go-admin/app/game/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

type PaymentConfigGetPageReq struct {
	dto.Pagination `search:"-"`
	GameId         int    `form:"gameId" search:"type:exact;column:game_id;table:game_payment_config" comment:"游戏ID"`
	Channel        string `form:"channel" search:"type:exact;column:channel;table:game_payment_config" comment:"渠道"`
}

func (m *PaymentConfigGetPageReq) GetNeedSearch() interface{} {
	return *m
}

type PaymentConfigSaveReq struct {
	Id          int    `json:"id"`
	GameId      int    `json:"gameId"`
	Channel     string `json:"channel" vd:"len($)>0;msg:'支付渠道不能为空'"`
	ChannelName string `json:"channelName"`
	Enabled     int    `json:"enabled"`
	Env         string `json:"env"`

	StripePublishableKey string `json:"stripePublishableKey"`
	StripeSecretKey      string `json:"stripeSecretKey"`
	StripeWebhookSecret  string `json:"stripeWebhookSecret"`

	AlipayAppId      string `json:"alipayAppId"`
	AlipayPrivateKey string `json:"alipayPrivateKey"`
	AlipayPublicKey  string `json:"alipayPublicKey"`
	AlipayNotifyUrl  string `json:"alipayNotifyUrl"`

	WechatAppId     string `json:"wechatAppId"`
	WechatMchId     string `json:"wechatMchId"`
	WechatApiKey    string `json:"wechatApiKey"`
	WechatNotifyUrl string `json:"wechatNotifyUrl"`

	common.ControlBy
}

func (s *PaymentConfigSaveReq) Generate(m *models.PaymentConfig) {
	m.GameId = s.GameId
	m.Channel = s.Channel
	m.ChannelName = s.ChannelName
	m.Enabled = s.Enabled
	m.Env = s.Env
	m.StripePublishableKey = s.StripePublishableKey
	m.StripeSecretKey = s.StripeSecretKey
	m.StripeWebhookSecret = s.StripeWebhookSecret
	m.AlipayAppId = s.AlipayAppId
	m.AlipayPrivateKey = s.AlipayPrivateKey
	m.AlipayPublicKey = s.AlipayPublicKey
	m.AlipayNotifyUrl = s.AlipayNotifyUrl
	m.WechatAppId = s.WechatAppId
	m.WechatMchId = s.WechatMchId
	m.WechatApiKey = s.WechatApiKey
	m.WechatNotifyUrl = s.WechatNotifyUrl
	m.CreateBy = s.CreateBy
	m.UpdateBy = s.UpdateBy
}

func (s *PaymentConfigSaveReq) GetId() interface{} {
	return s.Id
}

type PaymentConfigById struct {
	dto.ObjectById
}
