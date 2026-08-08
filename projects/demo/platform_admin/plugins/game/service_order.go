package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"platform-admin/plugin-sdk/proto"

	"gorm.io/gorm"
)

// ===== 订单管理 =====

// handleOrderGetPage 订单分页列表
func handleOrderGetPage(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	params := parseQuery(req.Query)
	page, pageSize := parsePage(params)

	query := db.Model(&Order{}).Where("tenant_id = ? AND deleted_at IS NULL", req.TenantId)

	if v := params["gameId"]; v != "" {
		if gid, err := strconv.Atoi(v); err == nil && gid > 0 {
			query = query.Where("game_id = ?", gid)
		}
	}
	if v := params["playerId"]; v != "" {
		if pid, err := strconv.Atoi(v); err == nil && pid > 0 {
			query = query.Where("player_id = ?", pid)
		}
	}
	if v := params["orderNo"]; v != "" {
		query = query.Where("order_no LIKE ?", "%"+v+"%")
	}
	if v := params["status"]; v != "" {
		query = query.Where("status = ?", v)
	}
	if v := params["channel"]; v != "" {
		query = query.Where("channel = ?", v)
	}

	var total int64
	query.Count(&total)

	var list []Order
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list)

	return jsonResp(200, "ok", PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// handleOrderGet 获取单个订单详情
func handleOrderGet(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var order Order
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "订单不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", order)
}

// OrderRefundReq 退款请求
type OrderRefundReq struct {
	RefundReason string `json:"refundReason"`
}

// handleOrderRefund 订单退款
func handleOrderRefund(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var order Order
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "订单不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	if order.Status != "paid" {
		return jsonResp(400, "只有已支付的订单才能退款", nil)
	}

	var body OrderRefundReq
	if err := json.Unmarshal(req.Body, &body); err != nil {
		return jsonResp(400, "请求参数解析失败: "+err.Error(), nil)
	}
	if strings.TrimSpace(body.RefundReason) == "" {
		return jsonResp(400, "退款原因不能为空", nil)
	}
	body.RefundReason = strings.TrimSpace(body.RefundReason)

	now := time.Now()
	order.Status = "refunded"
	order.RefundedAt = &now
	order.RefundReason = body.RefundReason
	order.UpdateBy = int(req.UserId)

	if err := db.Save(&order).Error; err != nil {
		return jsonResp(500, "退款失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", order)
}
