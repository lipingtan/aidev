package service

import (
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// LoginLogListParams 登录日志查询参数
type LoginLogListParams struct {
	Page     int
	PageSize int
	Username string
	IP       string
	Status   *int
}

// LoginLogListResult 登录日志分页结果
type LoginLogListResult struct {
	List     []model.LoginLog `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// LoginLogService 登录日志服务
type LoginLogService struct {
	db *gorm.DB
}

// NewLoginLogService 创建登录日志服务
func NewLoginLogService(db *gorm.DB) *LoginLogService {
	return &LoginLogService{db: db}
}

// List 分页查询登录日志
func (s *LoginLogService) List(params LoginLogListParams) (*LoginLogListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	query := s.db.Model(&model.LoginLog{})
	if params.Username != "" {
		query = query.Where("username LIKE ?", "%"+params.Username+"%")
	}
	if params.IP != "" {
		query = query.Where("ip LIKE ?", "%"+params.IP+"%")
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var list []model.LoginLog
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}

	return &LoginLogListResult{
		List:     list,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// Delete 删除登录日志
func (s *LoginLogService) Delete(id int64) error {
	return s.db.Delete(&model.LoginLog{}, id).Error
}
