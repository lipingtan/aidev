package service

import (
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DataScopeService 数据权限维度业务逻辑层
type DataScopeService struct {
	db         *gorm.DB
	configRepo repository.DataScopeConfigRepository
	scopeRepo  repository.DataScopeRepository
	logger     OperationLogger
}

// NewDataScopeService 创建 DataScopeService 实例
func NewDataScopeService(db *gorm.DB, configRepo repository.DataScopeConfigRepository, scopeRepo repository.DataScopeRepository) *DataScopeService {
	return &DataScopeService{
		db:         db,
		configRepo: configRepo,
		scopeRepo:  scopeRepo,
		logger:     &noopLogger{},
	}
}

// SetLogger 设置操作日志记录器
func (s *DataScopeService) SetLogger(logger OperationLogger) {
	s.logger = logger
}

// CreateDataScopeConfigRequest 注册维度请求
type CreateDataScopeConfigRequest struct {
	DimensionName string `json:"dimension_name" binding:"required"`
	DisplayName   string `json:"display_name" binding:"required"`
	TableColumn   string `json:"table_column"`
	ValueSource   string `json:"value_source"`
	HandlerName   string `json:"handler_name"`
	Status        *int   `json:"status"`
}

// UpdateDataScopeConfigRequest 更新维度请求
type UpdateDataScopeConfigRequest struct {
	DimensionName string `json:"dimension_name"`
	DisplayName   string `json:"display_name"`
	TableColumn   string `json:"table_column"`
	ValueSource   string `json:"value_source"`
	HandlerName   string `json:"handler_name"`
	Status        *int   `json:"status"`
}

// RoleDataScopeItem 角色数据权限绑定条目
type RoleDataScopeItem struct {
	DimensionName   string          `json:"dimension_name" binding:"required"`
	TargetEntity    string          `json:"target_entity" binding:"required"`
	DimensionValues datatypes.JSON  `json:"dimension_values" binding:"required"`
}

// SetRoleDataScopesRequest 角色配置数据权限请求
type SetRoleDataScopesRequest struct {
	Scopes []RoleDataScopeItem `json:"scopes" binding:"required"`
}

// CreateConfig 注册维度
func (s *DataScopeService) CreateConfig(req *CreateDataScopeConfigRequest) (*model.DataScopeConfig, error) {
	// 检查 dimension_name 唯一性
	existing, err := s.configRepo.FindByDimensionName(s.db, req.DimensionName)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "维度标识已存在")
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	cfg := &model.DataScopeConfig{
		DimensionName: req.DimensionName,
		DisplayName:   req.DisplayName,
		TableColumn:   req.TableColumn,
		ValueSource:   req.ValueSource,
		HandlerName:   req.HandlerName,
		Status:        status,
	}

	if err := s.configRepo.Create(s.db, cfg); err != nil {
		return nil, err
	}

	s.logger.Log(0, "create_data_scope_config", "data_scope_config", cfg.ID, "注册维度: "+cfg.DimensionName)
	return cfg, nil
}

// UpdateConfig 更新维度
func (s *DataScopeService) UpdateConfig(id int64, req *UpdateDataScopeConfigRequest) (*model.DataScopeConfig, error) {
	cfg, err := s.configRepo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}

	// 如果修改了 dimension_name，需要检查唯一性
	if req.DimensionName != "" && req.DimensionName != cfg.DimensionName {
		existing, err := s.configRepo.FindByDimensionName(s.db, req.DimensionName)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "维度标识已存在")
		}
		cfg.DimensionName = req.DimensionName
	}

	if req.DisplayName != "" {
		cfg.DisplayName = req.DisplayName
	}
	if req.TableColumn != "" {
		cfg.TableColumn = req.TableColumn
	}
	if req.ValueSource != "" {
		cfg.ValueSource = req.ValueSource
	}
	if req.HandlerName != "" {
		cfg.HandlerName = req.HandlerName
	}
	if req.Status != nil {
		cfg.Status = *req.Status
	}

	if err := s.configRepo.Update(s.db, cfg); err != nil {
		return nil, err
	}

	s.logger.Log(0, "update_data_scope_config", "data_scope_config", cfg.ID, "更新维度: "+cfg.DimensionName)
	return cfg, nil
}

// DeleteConfig 删除维度
func (s *DataScopeService) DeleteConfig(id int64) error {
	if err := s.configRepo.Delete(s.db, id); err != nil {
		return err
	}
	s.logger.Log(0, "delete_data_scope_config", "data_scope_config", id, "")
	return nil
}

// ListConfigs 查询维度注册列表
func (s *DataScopeService) ListConfigs() ([]model.DataScopeConfig, error) {
	return s.configRepo.List(s.db)
}

// SetRoleDataScopes 角色配置数据权限
func (s *DataScopeService) SetRoleDataScopes(roleID int64, req *SetRoleDataScopesRequest) error {
	scopes := make([]model.DataScope, 0, len(req.Scopes))
	for _, item := range req.Scopes {
		scopes = append(scopes, model.DataScope{
			RoleID:          roleID,
			DimensionName:   item.DimensionName,
			TargetEntity:    item.TargetEntity,
			DimensionValues: item.DimensionValues,
		})
	}

	if err := s.scopeRepo.ReplaceByRole(s.db, roleID, scopes); err != nil {
		return err
	}

	s.logger.Log(0, "set_role_data_scopes", "role", roleID, "配置角色数据权限")
	return nil
}

// GetRoleDataScopes 查询角色数据权限绑定
func (s *DataScopeService) GetRoleDataScopes(roleID int64) ([]model.DataScope, error) {
	return s.scopeRepo.FindByRole(s.db, roleID)
}
