package dto

import (
	"go-admin/common/dto"
)

type OrderGetPageReq struct {
	dto.Pagination `search:"-"`
	GameId         int    `form:"gameId" search:"type:exact;column:game_id;table:game_order" comment:"游戏ID"`
	PlayerId       int    `form:"playerId" search:"type:exact;column:player_id;table:game_order" comment:"玩家ID"`
	OrderNo        string `form:"orderNo" search:"type:contains;column:order_no;table:game_order" comment:"订单号"`
	Status         string `form:"status" search:"type:exact;column:status;table:game_order" comment:"状态"`
	Channel        string `form:"channel" search:"type:exact;column:channel;table:game_order" comment:"支付渠道"`
}

func (m *OrderGetPageReq) GetNeedSearch() interface{} {
	return *m
}

// OrderRefundReq 退款请求
type OrderRefundReq struct {
	Id           int    `uri:"id" json:"-"`
	RefundReason string `json:"refundReason" vd:"len($)>0;msg:'退款原因不能为空'"`
}

func (s *OrderRefundReq) GetId() interface{} {
	return s.Id
}

type OrderById struct {
	dto.ObjectById
}
