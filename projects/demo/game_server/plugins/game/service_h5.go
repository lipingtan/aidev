package main

import (
	"encoding/json"
	"strconv"
	"time"

	"game-server/plugin-sdk/proto"

	"gorm.io/gorm"
)

// ===== H5 页面管理 =====

// handleH5GetPage H5 页面分页列表
func handleH5GetPage(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	params := parseQuery(req.Query)
	page, pageSize := parsePage(params)

	query := db.Model(&H5Page{}).Where("tenant_id = ? AND deleted_at IS NULL", req.TenantId)

	if v := params["gameId"]; v != "" {
		if gid, err := strconv.Atoi(v); err == nil && gid > 0 {
			query = query.Where("game_id = ?", gid)
		}
	}
	if v := params["name"]; v != "" {
		query = query.Where("name LIKE ?", "%"+v+"%")
	}
	if v := params["pageType"]; v != "" {
		query = query.Where("page_type = ?", v)
	}
	if v := params["status"]; v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 {
			query = query.Where("status = ?", s)
		}
	}

	var total int64
	query.Count(&total)

	var list []H5Page
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list)

	return jsonResp(200, "ok", PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// handleH5Get 获取单个 H5 页面详情
func handleH5Get(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var h5 H5Page
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&h5).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "H5页面不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", h5)
}

// H5PageInsertReq 创建 H5 页面请求
type H5PageInsertReq struct {
	GameId      int    `json:"gameId"`
	PageKey     string `json:"pageKey"`
	Name        string `json:"name"`
	PageType    string `json:"pageType"`
	Content     string `json:"content"`
	ExternalUrl string `json:"externalUrl"`
	UseExternal int    `json:"useExternal"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
}

// handleH5Insert 创建 H5 页面
func handleH5Insert(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	var body H5PageInsertReq
	if err := json.Unmarshal(req.Body, &body); err != nil {
		return jsonResp(400, "请求参数解析失败: "+err.Error(), nil)
	}
	if body.GameId <= 0 {
		return jsonResp(400, "游戏ID不能为空", nil)
	}
	if body.PageKey == "" {
		return jsonResp(400, "页面标识不能为空", nil)
	}
	if body.Name == "" {
		return jsonResp(400, "页面名称不能为空", nil)
	}

	now := time.Now()
	h5 := H5Page{
		GameId:      body.GameId,
		PageKey:     body.PageKey,
		Name:        body.Name,
		PageType:    body.PageType,
		Content:     body.Content,
		ExternalUrl: body.ExternalUrl,
		UseExternal: body.UseExternal,
		Status:      body.Status,
		Remark:      body.Remark,
		ModelTime:   ModelTime{CreatedAt: now, UpdatedAt: now},
		ControlBy:   ControlBy{CreateBy: int(req.UserId), UpdateBy: int(req.UserId)},
		TenantBy:    TenantBy{TenantId: int(req.TenantId)},
	}
	if h5.PageType == "" {
		h5.PageType = "custom"
	}
	if h5.UseExternal == 0 {
		h5.UseExternal = 2
	}
	if h5.Status == 0 {
		h5.Status = 1
	}

	if err := db.Create(&h5).Error; err != nil {
		return jsonResp(500, "创建失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", h5)
}

// H5PageUpdateReq 更新 H5 页面请求
type H5PageUpdateReq struct {
	Name        string `json:"name"`
	PageType    string `json:"pageType"`
	Content     string `json:"content"`
	ExternalUrl string `json:"externalUrl"`
	UseExternal int    `json:"useExternal"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
}

// handleH5Update 更新 H5 页面
func handleH5Update(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing H5Page
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "H5页面不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	var body H5PageUpdateReq
	if err := json.Unmarshal(req.Body, &body); err != nil {
		return jsonResp(400, "请求参数解析失败: "+err.Error(), nil)
	}

	if body.Name != "" {
		existing.Name = body.Name
	}
	if body.PageType != "" {
		existing.PageType = body.PageType
	}
	// Content 允许设置为空字符串（清空内容），所以不做非空判断
	existing.Content = body.Content
	if body.ExternalUrl != "" {
		existing.ExternalUrl = body.ExternalUrl
	}
	if body.UseExternal > 0 {
		existing.UseExternal = body.UseExternal
	}
	if body.Status > 0 {
		existing.Status = body.Status
	}
	if body.Remark != "" {
		existing.Remark = body.Remark
	}
	existing.UpdateBy = int(req.UserId)
	existing.UpdatedAt = time.Now()

	if err := db.Save(&existing).Error; err != nil {
		return jsonResp(500, "更新失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", existing)
}

// handleH5Delete 删除 H5 页面（软删除）
func handleH5Delete(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing H5Page
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "H5页面不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	now := time.Now()
	existing.DeletedAt = &now
	if err := db.Save(&existing).Error; err != nil {
		return jsonResp(500, "删除失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", nil)
}
