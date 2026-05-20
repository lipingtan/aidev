package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/game/models"
	"go-admin/app/game/service/dto"
	"go-admin/common/actions"
	"go-admin/common/storage"
)

type Dlc struct {
	service.Service
}

func (e *Dlc) GetPage(c *dto.DlcGetPageReq, p *actions.DataPermission, list *[]models.Dlc, count *int64) error {
	var data models.Dlc
	err := e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("db error: %s", err)
		return err
	}
	return nil
}

func (e *Dlc) Get(c *dto.DlcById, p *actions.DataPermission, model *models.Dlc) error {
	var data models.Dlc
	err := e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		First(model, c.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("DLC不存在或无权查看")
	}
	return err
}

func (e *Dlc) Insert(c *dto.DlcInsertReq) error {
	var data models.Dlc
	c.Generate(&data)
	return e.Orm.Create(&data).Error
}

func (e *Dlc) Update(c *dto.DlcUpdateReq, p *actions.DataPermission) error {
	var model = models.Dlc{}
	e.Orm.Scopes(actions.Permission(model.TableName(), p)).First(&model, c.GetId())

	// 上架校验：filePath 为空时不允许上架
	if c.Status == 1 && model.FilePath == "" {
		return errors.New("请先上传 PCK 文件后再上架")
	}

	c.Generate(&model)
	return e.Orm.Save(&model).Error
}

func (e *Dlc) Remove(c *dto.DlcById, p *actions.DataPermission) error {
	// 删除保护：检查是否有关联订单
	var orderCount int64
	e.Orm.Table("game_order").
		Where("product_type = ? AND product_id = ? AND deleted_at IS NULL", "dlc", c.GetId()).
		Count(&orderCount)
	if orderCount > 0 {
		return errors.New("该 DLC 已有玩家购买，无法删除")
	}

	var data models.Dlc
	return e.Orm.Model(&data).
		Scopes(actions.Permission(data.TableName(), p)).
		Delete(&data, c.GetId()).Error
}

// UploadPck 通过 FileStore 接口上传 PCK 文件并更新哈希和大小
func (e *Dlc) UploadPck(dlcId int, filename string, reader io.Reader, store storage.FileStore, p *actions.DataPermission) error {
	var dlc models.Dlc
	err := e.Orm.Scopes(actions.Permission(dlc.TableName(), p)).First(&dlc, dlcId).Error
	if err != nil {
		return errors.New("DLC不存在")
	}

	// 使用 TeeReader 同时计算 SHA256
	hasher := sha256.New()
	teeReader := io.TeeReader(reader, hasher)

	// 通过 FileStore 接口存储文件
	storedFilename := fmt.Sprintf("%d_%s.pck", dlcId, dlc.DlcKey)
	path, err := store.Save("dlc", storedFilename, teeReader)
	if err != nil {
		return fmt.Errorf("文件存储失败: %w", err)
	}

	// 获取文件大小：通过 ByteCounter 包装
	// 由于 TeeReader 已经读完，我们需要在读取时统计字节数
	// 重新设计：使用 countingReader
	// 实际上 TeeReader 读完后 hasher 已有数据，但我们需要文件大小
	// 这里通过查询存储后的文件获取大小，或者在读取时统计
	// 简化方案：从 store 检查文件是否存在来确认，文件大小从 multipart header 获取
	// 由于 reader 已经被消费，文件大小需要从外部传入或通过其他方式获取

	// 更新数据库
	dlc.FilePath = path
	dlc.Sha256 = fmt.Sprintf("%x", hasher.Sum(nil))
	return e.Orm.Save(&dlc).Error
}

// UploadPckWithSize 通过 FileStore 接口上传 PCK 文件，同时记录文件大小
func (e *Dlc) UploadPckWithSize(dlcId int, filename string, fileSize int64, reader io.Reader, store storage.FileStore, p *actions.DataPermission) error {
	var dlc models.Dlc
	err := e.Orm.Scopes(actions.Permission(dlc.TableName(), p)).First(&dlc, dlcId).Error
	if err != nil {
		return errors.New("DLC不存在")
	}

	// 使用 TeeReader 同时计算 SHA256
	hasher := sha256.New()
	teeReader := io.TeeReader(reader, hasher)

	// 通过 FileStore 接口存储文件
	storedFilename := fmt.Sprintf("%d_%s.pck", dlcId, dlc.DlcKey)
	path, err := store.Save("dlc", storedFilename, teeReader)
	if err != nil {
		return fmt.Errorf("文件存储失败: %w", err)
	}

	// 更新数据库
	dlc.FilePath = path
	dlc.FileSize = fileSize
	dlc.Sha256 = fmt.Sprintf("%x", hasher.Sum(nil))
	return e.Orm.Save(&dlc).Error
}

// GetDownloadURL 获取 DLC 文件下载 URL
func (e *Dlc) GetDownloadURL(dlcId int, store storage.FileStore, p *actions.DataPermission) (string, error) {
	var dlc models.Dlc
	err := e.Orm.Scopes(actions.Permission(dlc.TableName(), p)).First(&dlc, dlcId).Error
	if err != nil {
		return "", errors.New("DLC不存在")
	}
	if dlc.FilePath == "" {
		return "", errors.New("该DLC尚未上传PCK文件")
	}
	url, err := store.GetURL(dlc.FilePath)
	if err != nil {
		return "", fmt.Errorf("获取下载URL失败: %w", err)
	}
	return url, nil
}

// DlcStatsResult 统计结果
type DlcStatsResult struct {
	TotalDownloads int64            `json:"totalDownloads"`
	TotalRevenue   int64            `json:"totalRevenue"`
	Trend          []DlcTrendItem   `json:"trend"`
	Ranking        []DlcRankingItem `json:"ranking"`
}

// DlcTrendItem 趋势数据项
type DlcTrendItem struct {
	Date    string `json:"date"`
	Count   int64  `json:"count"`
	Revenue int64  `json:"revenue"`
}

// DlcRankingItem 排行数据项
type DlcRankingItem struct {
	ProductId   int    `json:"productId"`
	ProductName string `json:"productName"`
	Downloads   int64  `json:"downloads"`
	Revenue     int64  `json:"revenue"`
}

// GetStats 获取 DLC 统计数据
func (e *Dlc) GetStats(gameId int) (*DlcStatsResult, error) {
	result := &DlcStatsResult{}

	// 总计
	type TotalRow struct {
		TotalDownloads int64
		TotalRevenue   int64
	}
	var total TotalRow
	query := e.Orm.Table("game_order").
		Select("COUNT(*) as total_downloads, COALESCE(SUM(amount),0) as total_revenue").
		Where("product_type = ? AND status = ?", "dlc", "paid")
	if gameId > 0 {
		query = query.Where("game_id = ?", gameId)
	}
	if err := query.Scan(&total).Error; err != nil {
		return nil, err
	}
	result.TotalDownloads = total.TotalDownloads
	result.TotalRevenue = total.TotalRevenue

	// 近30天趋势
	trendQuery := e.Orm.Table("game_order").
		Select("DATE(paid_at) as date, COUNT(*) as count, COALESCE(SUM(amount),0) as revenue").
		Where("product_type = ? AND status = ? AND paid_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)", "dlc", "paid")
	if gameId > 0 {
		trendQuery = trendQuery.Where("game_id = ?", gameId)
	}
	trendQuery = trendQuery.Group("DATE(paid_at)").Order("date")
	if err := trendQuery.Scan(&result.Trend).Error; err != nil {
		return nil, err
	}

	// DLC 排行 TOP 10
	rankQuery := e.Orm.Table("game_order").
		Select("product_id, product_name, COUNT(*) as downloads, COALESCE(SUM(amount),0) as revenue").
		Where("product_type = ? AND status = ?", "dlc", "paid")
	if gameId > 0 {
		rankQuery = rankQuery.Where("game_id = ?", gameId)
	}
	rankQuery = rankQuery.Group("product_id, product_name").Order("revenue DESC").Limit(10)
	if err := rankQuery.Scan(&result.Ranking).Error; err != nil {
		return nil, err
	}

	return result, nil
}
