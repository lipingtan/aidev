package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"platform-admin/plugin-sdk/proto"

	"gorm.io/gorm"
)

// ===== DLC CRUD + 上传/下载/统计 =====

// handleDlcGetPage DLC 分页列表
func handleDlcGetPage(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	params := parseQuery(req.Query)
	page, pageSize := parsePage(params)

	query := db.Model(&Dlc{}).Where("tenant_id = ? AND deleted_at IS NULL", req.TenantId)

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

	var list []Dlc
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list)

	return jsonResp(200, "ok", PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// handleDlcGet 获取单个 DLC 详情
func handleDlcGet(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var dlc Dlc
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&dlc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "DLC不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", dlc)
}

// DlcCreateReq 创建 DLC 请求
type DlcCreateReq struct {
	GameId         int    `json:"gameId"`
	DlcKey         string `json:"dlcKey"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	Price          int    `json:"price"`
	IsFree         int    `json:"isFree"`
	Status         int    `json:"status"`
	MinGameVersion string `json:"minGameVersion"`
}

// handleDlcInsert 创建 DLC
func handleDlcInsert(req *proto.HttpRequest) (*proto.HttpResponse, error) {
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
	dlc := Dlc{
		GameId:         body.GameId,
		DlcKey:         body.DlcKey,
		Name:           body.Name,
		Version:        body.Version,
		Description:    body.Description,
		Price:          body.Price,
		IsFree:         body.IsFree,
		Status:         body.Status,
		MinGameVersion: body.MinGameVersion,
		ModelTime:      ModelTime{CreatedAt: now, UpdatedAt: now},
		ControlBy:      ControlBy{CreateBy: int(req.UserId), UpdateBy: int(req.UserId)},
		TenantBy:       TenantBy{TenantId: int(req.TenantId)},
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
	if dlc.MinGameVersion == "" {
		dlc.MinGameVersion = "0.0.0"
	}

	if err := db.Create(&dlc).Error; err != nil {
		return jsonResp(500, "创建失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", dlc)
}

// DlcUpdateReq 更新 DLC 请求
type DlcUpdateReq struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	Price          int    `json:"price"`
	IsFree         int    `json:"isFree"`
	Status         int    `json:"status"`
	MinGameVersion string `json:"minGameVersion"`
}

// handleDlcUpdate 更新 DLC
func handleDlcUpdate(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing Dlc
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

	if body.Name != "" {
		existing.Name = body.Name
	}
	if body.Version != "" {
		existing.Version = body.Version
	}
	if body.Description != "" {
		existing.Description = body.Description
	}
	// isFree 优先：免费时强制 price=0，否则按传入值更新
	if body.IsFree > 0 {
		existing.IsFree = body.IsFree
	}
	if existing.IsFree == 1 {
		existing.Price = 0
	} else if body.Price >= 0 {
		existing.Price = body.Price
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

// handleDlcDelete 删除 DLC（有订单时拒绝删除）
func handleDlcDelete(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing Dlc
	err := db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", id, req.TenantId).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return jsonResp(404, "DLC不存在", nil)
		}
		return jsonResp(500, "查询失败: "+err.Error(), nil)
	}

	// 删除保护：检查是否有关联订单
	var orderCount int64
	db.Model(&Order{}).
		Where("product_type = ? AND product_id = ? AND deleted_at IS NULL", "dlc", id).
		Count(&orderCount)
	if orderCount > 0 {
		return jsonResp(400, "该 DLC 已有玩家购买，无法删除", nil)
	}

	now := time.Now()
	existing.DeletedAt = &now
	if err := db.Save(&existing).Error; err != nil {
		return jsonResp(500, "删除失败: "+err.Error(), nil)
	}
	return jsonResp(200, "ok", nil)
}

// handleDlcUpload 上传 PCK 文件
func handleDlcUpload(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing Dlc
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

// handleDlcDownload 获取下载 URL
func handleDlcDownload(req *proto.HttpRequest, id int) (*proto.HttpResponse, error) {
	var existing Dlc
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

	filename := fmt.Sprintf("%d_%s.pck", id, existing.DlcKey)
	downloadURL := "/static/dlc/" + filename

	return jsonResp(200, "ok", map[string]interface{}{
		"url":      downloadURL,
		"fileName": filename,
		"fileSize": existing.FileSize,
		"sha256":   existing.Sha256,
	})
}

// handleDlcStats DLC 统计数据
func handleDlcStats(req *proto.HttpRequest) (*proto.HttpResponse, error) {
	params := parseQuery(req.Query)
	tenantId := req.TenantId

	type StatsResult struct {
		TotalDlc       int64       `json:"totalDlc"`
		TotalDownloads int64       `json:"totalDownloads"`
		TotalRevenue   int64       `json:"totalRevenue"`
		Trend          interface{} `json:"trend"`
		Ranking        interface{} `json:"ranking"`
	}

	result := StatsResult{}

	// DLC 总数
	db.Model(&Dlc{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantId).Count(&result.TotalDlc)

	// 从订单表统计
	orderQuery := db.Model(&Order{}).
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
	trendQuery := db.Model(&Order{}).
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
	rankQuery := db.Model(&Order{}).
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
