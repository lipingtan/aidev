package repository

import (
	"time"

	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// OperationLogListParams 操作日志列表查询参数
type OperationLogListParams struct {
	Page       int
	PageSize   int
	Module     string
	Action     string
	UserID     *int64
	TargetType string
	StartTime  *time.Time
	EndTime    *time.Time
}

// OperationLogRepository 操作日志数据访问接口
type OperationLogRepository interface {
	Create(db *gorm.DB, log *model.OperationLog) error
	List(db *gorm.DB, params OperationLogListParams) ([]model.OperationLog, int64, error)
}

// operationLogRepo OperationLogRepository 默认实现
type operationLogRepo struct{}

// NewOperationLogRepository 创建 OperationLogRepository 实例
func NewOperationLogRepository() OperationLogRepository {
	return &operationLogRepo{}
}

// Create 创建操作日志
func (r *operationLogRepo) Create(db *gorm.DB, log *model.OperationLog) error {
	return db.Create(log).Error
}

// List 分页查询操作日志
func (r *operationLogRepo) List(db *gorm.DB, params OperationLogListParams) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := db.Model(&model.OperationLog{})

	if params.Module != "" {
		query = query.Where("module = ?", params.Module)
	}
	if params.Action != "" {
		query = query.Where("action = ?", params.Action)
	}
	if params.UserID != nil {
		query = query.Where("user_id = ?", *params.UserID)
	}
	if params.TargetType != "" {
		query = query.Where("target_type = ?", params.TargetType)
	}
	if params.StartTime != nil {
		query = query.Where("created_at >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("created_at <= ?", *params.EndTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
