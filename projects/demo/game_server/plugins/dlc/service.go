package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"game-server/plugin-sdk/proto"

	"gorm.io/gorm"
)

// handleGetPage 分页查询 DLC 列表
func handleGetPage(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	params := parseQuery(req.Query)
	page, _ := strconv.Atoi(params["page"])
	pageSize, _ := strconv.Atoi(params["pageSize"])
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := db.Model(&DlcGameDlc{}).Where("tenant_id = ? AND deleted_at IS NULL", req.TenantId)

	// 过滤条件
	if v := params["gameId"]; v != "" {
		if gid, err := strconv.Atoi(v); err == nil && gid > 0 {
			query = query.Where("game_id = ?", gid)
		}
	}
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

	var list []DlcGameDlc
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list)

	return jsonResp(200, "ok", PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// handleGet 获取单个 DLC 详情
func handleGet(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var dlc DlcGameDlc
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&dlc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "DLC不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", dlc)
}

// handleInsert 创建 DLC
func handleInsert(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	var body DlcCreateReq
	if err := json.Unmarshal(req.Body, &body); err != nil {
		return jsonResp(400, "请求参数解析失败: "+err.Error(), nil)
	}
	if body.GameId <= 0 {
		return jsonResp(400, "游戏ID不能为空", nil)
	}
	if body.DlcKey == "" {
		return jsonResp(400, "DLC标识不能为空", nil)
	}
	if body.Name == "" {
		return jsonResp(400, "DLC名称不能为空", nil)
	}

	now := time.Now()
	dlc := DlcGameDlc{
		GameId:         body.GameId,
		DlcKey:         body.DlcKey,
		Name:           body.Name,
		Version:        body.Version,
		Description:    body.Description,
		Price:          body.Price,
		IsFree:         body.IsFree,
		Status:         body.Status,
		MinGameVersion: body.MinGameVersion,
		TenantId:       int(req.TenantId),
		CreateBy:       int(req.UserId),
		UpdateBy:       int(req.UserId),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if dlc.Version == "" {
		dlc.Version = "1.0.0"
	}
	if dlc.IsFree == 0 {
		dlc.IsFree = 1
	}
	if dlc.Status == 0 {
		dlc.Status = 2 // 默认下架
	}

	if err := db.Create(&dlc).Error; err != nil {
		return jsonResp(500, "创建失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", dlc)
}

// handleUpdate 更新 DLC
func handleUpdate(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing DlcGameDlc
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "DLC不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	var body DlcUpdateReq
	if err := json.Unmarshal(req.Body, &body); err != nil {
		return jsonResp(400, "请求参数解析失败: "+err.Error(), nil)
	}

	// 上架校验：无文件时拒绝上架
	if body.Status == 1 && existing.FilePath == "" {
		return jsonResp(400, "请先上传 PCK 文件后再上架", nil)
	}

	// 更新字段
	if body.Name != "" {
		existing.Name = body.Name
	}
	if body.Version != "" {
		existing.Version = body.Version
	}
	if body.Description != "" {
		existing.Description = body.Description
	}
	if body.Price > 0 {
		existing.Price = body.Price
	}
	if body.IsFree > 0 {
		existing.IsFree = body.IsFree
	}
	if body.Status > 0 {
		existing.Status = body.Status
	}
	if body.MinGameVersion != "" {
		existing.MinGameVersion = body.MinGameVersion
	}
	existing.UpdateBy = int(req.UserId)
	existing.UpdatedAt = time.Now()

	if err := db.Save(&existing).Error; err != nil {
		return jsonResp(500, "更新失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", existing)
}

// handleDelete 删除 DLC（有订单时拒绝删除）
func handleDelete(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing DlcGameDlc
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "DLC不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	// 删除保护：检查是否有关联订单
	var orderCount int64
	db.Table("game_order").
		Where("product_type = ? AND product_id = ? AND deleted_at IS NULL", "dlc", id).
		Count(&orderCount)
	if orderCount > 0 {
		return jsonResp(400, "该 DLC 已有玩家购买，无法删除", nil)
	}

	// 软删除
	now := time.Now()
	existing.DeletedAt = &now
	if err := db.Save(&existing).Error; err != nil {
		return jsonResp(500, "删除失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", nil)
}

// handleUpload 上传 PCK 文件
func handleUpload(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing DlcGameDlc
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "DLC不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	if len(req.Body) == 0 {
		return jsonResp(400, "文件内容为空", nil)
	}

	// 确保存储目录存在
	storageDir := "./static/dlc"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return jsonResp(500, "创建存储目录失败: "+err.Error(), nil)
	}

	// 存储文件
	filename := fmt.Sprintf("%d_%s.pck", id, existing.DlcKey)
	filePath := filepath.Join(storageDir, filename)
	if err := os.WriteFile(filePath, req.Body, 0644); err != nil {
		return jsonResp(500, "文件写入失败: "+err.Error(), nil)
	}

	// 计算 SHA256
	hasher := sha256.New()
	hasher.Write(req.Body)
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	// 更新数据库
	existing.FilePath = filePath
	existing.FileSize = int64(len(req.Body))
	existing.Sha256 = hash
	existing.UpdateBy = int(req.UserId)
	existing.UpdatedAt = time.Now()

	if err := db.Save(&existing).Error; err != nil {
		return jsonResp(500, "更新文件信息失败: "+err.Error(), nil)
	}

	return jsonResp(200, "ok", map[string]interface{}{
		"filePath": filePath,
		"fileSize": existing.FileSize,
		"sha256":   hash,
	})
}

// handleDownload 获取下载 URL
func handleDownload(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing DlcGameDlc
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "DLC不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	if existing.FilePath == "" {
		return jsonResp(400, "该DLC尚未上传PCK文件", nil)
	}

	// 更新下载次数
	db.Model(&existing).UpdateColumn("download_count", gorm.Expr("download_count + 1"))

	// 返回下载路径
	filename := fmt.Sprintf("%d_%s.pck", id, existing.DlcKey)
	downloadURL := "/static/dlc/" + filename

	return jsonResp(200, "ok", map[string]interface{}{
		"url":      downloadURL,
		"fileName": filename,
		"fileSize": existing.FileSize,
		"sha256":   existing.Sha256,
	})
}

// handleStats 获取 DLC 统计数据
func handleStats(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	params := parseQuery(req.Query)
	tenantId := req.TenantId

	// 统计结果
	type StatsResult struct {
		TotalDlc       int64       `json:"totalDlc"`
		TotalDownloads int64       `json:"totalDownloads"`
		TotalRevenue   int64       `json:"totalRevenue"`
		Trend          interface{} `json:"trend"`
		Ranking        interface{} `json:"ranking"`
	}

	result := StatsResult{}

	// DLC 总数
	db.Model(&DlcGameDlc{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantId).Count(&result.TotalDlc)

	// 从 game_order 表统计
	orderQuery := db.Table("game_order").
		Where("product_type = ? AND status = ? AND deleted_at IS NULL", "dlc", "paid")

	if v := params["gameId"]; v != "" {
		if gid, err := strconv.Atoi(v); err == nil && gid > 0 {
			orderQuery = orderQuery.Where("game_id = ?", gid)
		}
	}

	// 总下载/收入
	type TotalRow struct {
		TotalDownloads int64
		TotalRevenue   int64
	}
	var total TotalRow
	orderQuery.Select("COUNT(*) as total_downloads, COALESCE(SUM(amount),0) as total_revenue").Scan(&total)
	result.TotalDownloads = total.TotalDownloads
	result.TotalRevenue = total.TotalRevenue

	// 近30天趋势
	type TrendItem struct {
		Date    string `json:"date"`
		Count   int64  `json:"count"`
		Revenue int64  `json:"revenue"`
	}
	var trend []TrendItem
	trendQuery := db.Table("game_order").
		Select("DATE(paid_at) as date, COUNT(*) as count, COALESCE(SUM(amount),0) as revenue").
		Where("product_type = ? AND status = ? AND deleted_at IS NULL AND paid_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)", "dlc", "paid")
	if v := params["gameId"]; v != "" {
		if gid, err := strconv.Atoi(v); err == nil && gid > 0 {
			trendQuery = trendQuery.Where("game_id = ?", gid)
		}
	}
	trendQuery.Group("DATE(paid_at)").Order("date").Scan(&trend)
	result.Trend = trend

	// DLC 排行 TOP 10
	type RankingItem struct {
		ProductId   int    `json:"productId"`
		ProductName string `json:"productName"`
		Downloads   int64  `json:"downloads"`
		Revenue     int64  `json:"revenue"`
	}
	var ranking []RankingItem
	rankQuery := db.Table("game_order").
		Select("product_id, product_name, COUNT(*) as downloads, COALESCE(SUM(amount),0) as revenue").
		Where("product_type = ? AND status = ? AND deleted_at IS NULL", "dlc", "paid")
	if v := params["gameId"]; v != "" {
		if gid, err := strconv.Atoi(v); err == nil && gid > 0 {
			rankQuery = rankQuery.Where("game_id = ?", gid)
		}
	}
	rankQuery.Group("product_id, product_name").Order("revenue DESC").Limit(10).Scan(&ranking)
	result.Ranking = ranking

	return jsonResp(200, "ok", result)
}
