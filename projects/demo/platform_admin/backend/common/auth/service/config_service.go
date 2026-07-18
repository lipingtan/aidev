package service

import (
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// ConfigListParams 配置列表查询参数
type ConfigListParams struct {
	Page       int
	PageSize   int
	ConfigName string
	ConfigKey  string
}

// ConfigListResult 配置分页结果
type ConfigListResult struct {
	List     []model.SysConfig `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// ConfigService 系统配置服务
type ConfigService struct {
	db *gorm.DB
}

// NewConfigService 创建配置服务
func NewConfigService(db *gorm.DB) *ConfigService {
	return &ConfigService{db: db}
}

// List 分页查询配置
func (s *ConfigService) List(params ConfigListParams) (*ConfigListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	query := s.db.Model(&model.SysConfig{})
	if params.ConfigName != "" {
		query = query.Where("config_name LIKE ?", "%"+params.ConfigName+"%")
	}
	if params.ConfigKey != "" {
		query = query.Where("config_key LIKE ?", "%"+params.ConfigKey+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var list []model.SysConfig
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}

	return &ConfigListResult{
		List:     list,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// GetByID 按 ID 查询配置
func (s *ConfigService) GetByID(id int64) (*model.SysConfig, error) {
	var cfg model.SysConfig
	if err := s.db.First(&cfg, id).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetByKey 按 key 查询配置
func (s *ConfigService) GetByKey(key string) (*model.SysConfig, error) {
	var cfg model.SysConfig
	if err := s.db.Where("config_key = ?", key).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Create 创建配置
func (s *ConfigService) Create(cfg *model.SysConfig) error {
	return s.db.Create(cfg).Error
}

// Update 更新配置
func (s *ConfigService) Update(id int64, updates map[string]interface{}) error {
	return s.db.Model(&model.SysConfig{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除配置
func (s *ConfigService) Delete(id int64) error {
	return s.db.Delete(&model.SysConfig{}, id).Error
}
