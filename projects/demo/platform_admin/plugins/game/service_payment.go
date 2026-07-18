package main

import (
	"encoding/json"
	"strconv"

	"platform-admin/plugin-sdk/proto"
)

// ===== 支付配置管理 =====

// handlePaymentGetPage 支付配置列表
func handlePaymentGetPage(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	params := parseQuery(req.Query)
	page, pageSize := parsePage(params)

	query := db.Model(&PaymentConfig{}).Where("tenant_id = ? AND deleted_at IS NULL", req.TenantId)

	if v := params["gameId"]; v != "" {
		if gid, err := strconv.Atoi(v); err == nil && gid > 0 {
			query = query.Where("game_id = ?", gid)
		}
	}
	if v := params["channel"]; v != "" {
		query = query.Where("channel = ?", v)
	}

	var total int64
	query.Count(&total)

	var list []PaymentConfig
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list)

	// 隐藏敏感密钥
	for i := range list {
		list[i].StripeSecretKey = "***"
		list[i].AlipayPrivateKey = "***"
		list[i].WechatApiKey = "***"
	}

	return jsonResp(200, "ok", PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// PaymentConfigSaveReq 保存支付配置请求
type PaymentConfigSaveReq struct {
	Id                   int    `json:"id"`
	GameId               int    `json:"gameId"`
	Channel              string `json:"channel"`
	ChannelName          string `json:"channelName"`
	Enabled              int    `json:"enabled"`
	Env                  string `json:"env"`
	StripePublishableKey string `json:"stripePublishableKey"`
	StripeSecretKey      string `json:"stripeSecretKey"`
	StripeWebhookSecret  string `json:"stripeWebhookSecret"`
	AlipayAppId          string `json:"alipayAppId"`
	AlipayPrivateKey     string `json:"alipayPrivateKey"`
	AlipayPublicKey      string `json:"alipayPublicKey"`
	AlipayNotifyUrl      string `json:"alipayNotifyUrl"`
	WechatAppId          string `json:"wechatAppId"`
	WechatMchId          string `json:"wechatMchId"`
	WechatApiKey         string `json:"wechatApiKey"`
	WechatNotifyUrl      string `json:"wechatNotifyUrl"`
}

// handlePaymentSave 创建或更新支付配置（upsert）
func handlePaymentSave(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	var body PaymentConfigSaveReq
	if err := json.Unmarshal(req.Body, &body); err != nil {
		return jsonResp(400, "请求参数解析失败: "+err.Error(), nil)
	}
	if body.Channel == "" {
		return jsonResp(400, "支付渠道不能为空", nil)
	}

	var model PaymentConfig
	// 查找是否已存在同游戏同渠道的配置
	db.Where("game_id = ? AND channel = ? AND tenant_id = ? AND deleted_at IS NULL",
		body.GameId, body.Channel, req.TenantId).First(&model)

	// 填充字段
	model.GameId = body.GameId
	model.Channel = body.Channel
	model.ChannelName = body.ChannelName
	model.Enabled = body.Enabled
	model.Env = body.Env
	model.StripePublishableKey = body.StripePublishableKey
	model.StripeSecretKey = body.StripeSecretKey
	model.StripeWebhookSecret = body.StripeWebhookSecret
	model.AlipayAppId = body.AlipayAppId
	model.AlipayPrivateKey = body.AlipayPrivateKey
	model.AlipayPublicKey = body.AlipayPublicKey
	model.AlipayNotifyUrl = body.AlipayNotifyUrl
	model.WechatAppId = body.WechatAppId
	model.WechatMchId = body.WechatMchId
	model.WechatApiKey = body.WechatApiKey
	model.WechatNotifyUrl = body.WechatNotifyUrl
	model.TenantId = int(req.TenantId)
	model.CreateBy = int(req.UserId)
	model.UpdateBy = int(req.UserId)

	var err error
	if model.Id == 0 {
		err = db.Create(&model).Error
	} else {
		err = db.Save(&model).Error
	}
	if err != nil {
		return jsonResp(500, "保存失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", model)
}
