package main

import (
	"encoding/json"
	"strconv"

	"game-server/plugin-sdk/proto"

	"gorm.io/gorm"
)

// ===== 玩家管理 =====

// handlePlayerGetPage 玩家分页列表
func handlePlayerGetPage(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	params := parseQuery(req.Query)
	page, pageSize := parsePage(params)

	query := db.Model(&Player{}).Where("tenant_id = ? AND deleted_at IS NULL", req.TenantId)

	if v := params["gameId"]; v != "" {
		if gid, err := strconv.Atoi(v); err == nil && gid > 0 {
			query = query.Where("game_id = ?", gid)
		}
	}
	if v := params["uid"]; v != "" {
		query = query.Where("uid LIKE ?", "%"+v+"%")
	}
	if v := params["nickname"]; v != "" {
		query = query.Where("nickname LIKE ?", "%"+v+"%")
	}
	if v := params["status"]; v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 {
			query = query.Where("status = ?", s)
		}
	}

	var total int64
	query.Count(&total)

	var list []Player
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list)

	return jsonResp(200, "ok", PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// handlePlayerGet 获取单个玩家详情
func handlePlayerGet(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var player Player
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&player).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "玩家不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", player)
}

// PlayerBanReq 封禁/解封请求
type PlayerBanReq struct {
	Status    int    `json:"status"`
	BanReason string `json:"banReason"`
}

// handlePlayerBan 封禁/解封玩家
func handlePlayerBan(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var player Player
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&player).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "玩家不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	var body PlayerBanReq
	if err := json.Unmarshal(req.Body, &body); err != nil {
		return jsonResp(400, "请求参数解析失败: "+err.Error(), nil)
	}

	if body.Status != 1 && body.Status != 2 {
		return jsonResp(400, "状态值无效，1=正常 2=封禁", nil)
	}

	player.Status = body.Status
	player.BanReason = body.BanReason
	player.UpdateBy = int(req.UserId)

	if err := db.Save(&player).Error; err != nil {
		return jsonResp(500, "操作失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", player)
}
