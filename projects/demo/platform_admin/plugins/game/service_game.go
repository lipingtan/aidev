package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"platform-admin/plugin-sdk/proto"

	"gorm.io/gorm"
)

// ===== 游戏 CRUD =====

// handleGameGetPage 游戏列表分页查询
func handleGameGetPage(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	params := parseQuery(req.Query)
	page, pageSize := parsePage(params)

	query := db.Model(&Game{}).Where("tenant_id = ? AND deleted_at IS NULL", req.TenantId)

	if v := params["name"]; v != "" {
		query = query.Where("name LIKE ?", "%"+v+"%")
	}
	if v := params["status"]; v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 {
			query = query.Where("status = ?", s)
		}
	}

	var total int64
	query.Count(&total)

	var list []Game
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list)

	// 隐藏 AppSecret
	for i := range list {
		list[i].AppSecret = "***"
	}

	return jsonResp(200, "ok", PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// handleGameGet 获取单个游戏详情
func handleGameGet(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var game Game
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&game).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "游戏不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}
	game.AppSecret = "***"
	return jsonResp(200, "ok", game)
}

// GameInsertReq 创建游戏请求
type GameInsertReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Version     string `json:"version"`
	Status      int    `json:"status"`
}

// handleGameInsert 创建游戏
func handleGameInsert(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	var body GameInsertReq
	if err := json.Unmarshal(req.Body, &body); err != nil {
		return jsonResp(400, "请求参数解析失败: "+err.Error(), nil)
	}
	if body.Name == "" {
		return jsonResp(400, "游戏名称不能为空", nil)
	}

	// 自动生成 AppKey（8字符随机 hex）
	appKeyBytes := make([]byte, 4)
	rand.Read(appKeyBytes)
	appKey := hex.EncodeToString(appKeyBytes)

	// 生成 AppSecret
	secret, err := generateSecret()
	if err != nil {
		return jsonResp(500, "生成密钥失败: "+err.Error(), nil)
	}

	now := time.Now()
	game := Game{
		Name:        body.Name,
		AppKey:      appKey,
		AppSecret:   secret,
		Description: body.Description,
		Icon:        body.Icon,
		Version:     body.Version,
		Status:      body.Status,
		ModelTime:   ModelTime{CreatedAt: now, UpdatedAt: now},
		ControlBy:   ControlBy{CreateBy: int(req.UserId), UpdateBy: int(req.UserId)},
		TenantBy:    TenantBy{TenantId: int(req.TenantId)},
	}
	if game.Version == "" {
		game.Version = "1.0.0"
	}
	if game.Status == 0 {
		game.Status = 1
	}

	if err := db.Create(&game).Error; err != nil {
		return jsonResp(500, "创建失败: "+err.Error(), nil)
	}
	game.AppSecret = "***"
	return jsonResp(200, "ok", game)
}

// GameUpdateReq 更新游戏请求
type GameUpdateReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Version     string `json:"version"`
	Status      int    `json:"status"`
}

// handleGameUpdate 更新游戏
func handleGameUpdate(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing Game
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "游戏不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	var body GameUpdateReq
	if err := json.Unmarshal(req.Body, &body); err != nil {
		return jsonResp(400, "请求参数解析失败: "+err.Error(), nil)
	}

	if body.Name != "" {
		existing.Name = body.Name
	}
	if body.Description != "" {
		existing.Description = body.Description
	}
	if body.Icon != "" {
		existing.Icon = body.Icon
	}
	if body.Version != "" {
		existing.Version = body.Version
	}
	if body.Status > 0 {
		existing.Status = body.Status
	}
	existing.UpdateBy = int(req.UserId)
	existing.UpdatedAt = time.Now()

	if err := db.Save(&existing).Error; err != nil {
		return jsonResp(500, "更新失败: "+err.Error(), nil)
	}
	existing.AppSecret = "***"
	return jsonResp(200, "ok", existing)
}

// handleGameDelete 删除游戏（软删除）
func handleGameDelete(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing Game
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "游戏不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	// 删除保护：检查是否有关联 DLC
	var dlcCount int64
	db.Model(&Dlc{}).Where("game_id = ? AND deleted_at IS NULL", id).Count(&dlcCount)
	if dlcCount > 0 {
		return jsonResp(400, "该游戏下存在 DLC，无法删除", nil)
	}

	now := time.Now()
	existing.DeletedAt = &now
	if err := db.Save(&existing).Error; err != nil {
		return jsonResp(500, "删除失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", nil)
}

// generateSecret 生成随机 32 字节 hex 密钥
func generateSecret() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// handleGameRegenSecret 重置游戏 AppSecret
func handleGameRegenSecret(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing Game
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "游戏不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	secret, err := generateSecret()
	if err != nil {
		return jsonResp(500, "生成密钥失败: "+err.Error(), nil)
	}

	existing.AppSecret = secret
	existing.UpdateBy = int(req.UserId)
	existing.UpdatedAt = time.Now()

	if err := db.Save(&existing).Error; err != nil {
		return jsonResp(500, "重置密钥失败: "+err.Error(), nil)
	}
	// 返回新密钥（仅此一次明文返回）
	return jsonResp(200, "ok", map[string]string{"appSecret": secret})
}
