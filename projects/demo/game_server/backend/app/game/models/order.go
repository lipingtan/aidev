package models

import (
	"time"
	"go-admin/common/models"
)

// Order 支付订单
type Order struct {
	models.Model
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

	models.ModelTime
	models.ControlBy
	models.TenantBy
}

func (Order) TableName() string { return "game_order" }
