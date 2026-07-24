package service

import (
	"fmt"
	"strings"
	"time"

	"go-admin/common/auth/cache"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

// 缓存配置
const (
	tenantDomainCachePrefix = "tenant_domain:"
	tenantDomainCacheTTL    = 5 * time.Minute
)

// 保留域名列表
var reservedDomains = []string{"localhost", "127.0.0.1", "0.0.0.0"}

// QueryDomainResult 域名查询结果 DTO
type QueryDomainResult struct {
	TenantCode string `json:"tenant_code"`
	TenantName string `json:"tenant_name"`
	Matched    bool   `json:"matched"`
}

// CreateTenantDomainRequest 创建域名映射请求
type CreateTenantDomainRequest struct {
	Domain   string `json:"domain" binding:"required"`
	TenantID int64  `json:"tenant_id,string" binding:"required"`
	Remark   string `json:"remark"`
}

// UpdateTenantDomainRequest 更新域名映射请求
type UpdateTenantDomainRequest struct {
	Domain   *string `json:"domain"`
	TenantID *int64  `json:"tenant_id,string"`
	Remark   *string `json:"remark"`
	Version  int     `json:"version" binding:"required"`
}

// TenantDomainListItem 列表项 DTO
type TenantDomainListItem struct {
	ID         int64  `json:"id,string"`
	Domain     string `json:"domain"`
	MatchType  string `json:"match_type"`
	TenantID   int64  `json:"tenant_id,string"`
	TenantCode string `json:"tenant_code"`
	TenantName string `json:"tenant_name"`
	Remark     string `json:"remark"`
	Version    int    `json:"version"`
	CreatedAt  string `json:"created_at"`
}

// TenantDomainService 域名-租户映射业务逻辑
type TenantDomainService struct {
	db         *gorm.DB
	domainRepo repository.TenantDomainRepository
	tenantRepo repository.TenantRepository
	cache      *cache.LocalCache
}

// NewTenantDomainService 创建 TenantDomainService 实例
func NewTenantDomainService(db *gorm.DB, domainRepo repository.TenantDomainRepository, tenantRepo repository.TenantRepository, c *cache.LocalCache) *TenantDomainService {
	return &TenantDomainService{
		db:         db,
		domainRepo: domainRepo,
		tenantRepo: tenantRepo,
		cache:      c,
	}
}

// QueryByDomain 根据域名查询对应租户信息（带缓存）
func (s *TenantDomainService) QueryByDomain(domain string) (*QueryDomainResult, error) {
	if domain == "" {
		return nil, errors.NewAuthError(errors.ErrInvalidParam, "域名不能为空")
	}

	cacheKey := tenantDomainCachePrefix + domain

	// 查缓存
	if val, ok := s.cache.Get(cacheKey); ok {
		if result, ok := val.(*QueryDomainResult); ok {
			return result, nil
		}
	}

	// 缓存 miss，查数据库
	td, err := s.domainRepo.FindByDomain(s.db, domain)
	if err != nil {
		return nil, err
	}

	var result *QueryDomainResult

	if td != nil {
		// 找到映射记录，查对应租户
		tenant, err := s.tenantRepo.FindByID(s.db, td.TenantID)
		if err != nil {
			// FindByID 返回 AuthError 时表示未找到，走 fallback
			result = s.fallbackDefault()
		} else {
			result = &QueryDomainResult{
				TenantCode: tenant.TenantCode,
				TenantName: tenant.Name,
				Matched:    true,
			}
		}
	} else {
		// 未找到映射记录，fallback 到 default 租户
		result = s.fallbackDefault()
	}

	// 写入缓存
	s.cache.Set(cacheKey, result, tenantDomainCacheTTL)
	return result, nil
}

// fallbackDefault 获取 default 租户信息，不存在时硬编码
func (s *TenantDomainService) fallbackDefault() *QueryDomainResult {
	tenant, err := s.tenantRepo.FindByCode(s.db, "default")
	if err != nil || tenant == nil {
		return &QueryDomainResult{
			TenantCode: "default",
			TenantName: "默认租户",
			Matched:    false,
		}
	}
	return &QueryDomainResult{
		TenantCode: tenant.TenantCode,
		TenantName: tenant.Name,
		Matched:    false,
	}
}

// Create 创建域名映射
func (s *TenantDomainService) Create(userID int64, req *CreateTenantDomainRequest) (*model.TenantDomain, error) {
	// 权限校验
	if err := s.assertSuperAdmin(userID); err != nil {
		return nil, err
	}

	// 域名格式校验
	if err := s.validateDomain(req.Domain); err != nil {
		return nil, err
	}

	// 租户存在性校验
	if _, err := s.tenantRepo.FindByID(s.db, req.TenantID); err != nil {
		return nil, errors.NewAuthError(errors.ErrEntityNotFound, "租户不存在")
	}

	// 唯一性校验
	existing, err := s.domainRepo.FindActiveByDomain(s.db, req.Domain)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "域名已存在")
	}

	td := &model.TenantDomain{
		Domain:    req.Domain,
		MatchType: "EXACT",
		TenantID:  req.TenantID,
		Remark:    req.Remark,
		Version:   1,
	}

	if err := s.domainRepo.Create(s.db, td); err != nil {
		return nil, err
	}

	// 清除缓存
	s.cache.Delete(tenantDomainCachePrefix + req.Domain)
	return td, nil
}

// Update 更新域名映射
func (s *TenantDomainService) Update(userID int64, id int64, req *UpdateTenantDomainRequest) (*model.TenantDomain, error) {
	// 权限校验
	if err := s.assertSuperAdmin(userID); err != nil {
		return nil, err
	}

	// 查找记录
	td, err := s.domainRepo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}
	if td == nil {
		return nil, errors.NewAuthError(errors.ErrEntityNotFound, "域名映射不存在")
	}

	// 乐观锁版本校验
	if td.Version != req.Version {
		return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
	}

	oldDomain := td.Domain

	// 按指针字段更新
	if req.Domain != nil {
		if err := s.validateDomain(*req.Domain); err != nil {
			return nil, err
		}
		// 唯一性校验（排除自身）
		if *req.Domain != td.Domain {
			existing, err := s.domainRepo.FindActiveByDomain(s.db, *req.Domain)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "域名已存在")
			}
		}
		td.Domain = *req.Domain
	}

	if req.TenantID != nil {
		if _, err := s.tenantRepo.FindByID(s.db, *req.TenantID); err != nil {
			return nil, errors.NewAuthError(errors.ErrEntityNotFound, "租户不存在")
		}
		td.TenantID = *req.TenantID
	}

	if req.Remark != nil {
		td.Remark = *req.Remark
	}

	oldVersion := td.Version
	td.Version = oldVersion + 1
	now := time.Now()
	td.UpdatedAt = &now

	if err := s.domainRepo.Update(s.db, td, oldVersion); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrDuplicateEntity, "数据已被其他操作修改，请刷新后重试")
		}
		return nil, err
	}

	// 清除新旧域名缓存
	s.cache.Delete(tenantDomainCachePrefix + oldDomain)
	if req.Domain != nil && *req.Domain != oldDomain {
		s.cache.Delete(tenantDomainCachePrefix + *req.Domain)
	}

	return td, nil
}

// Delete 删除域名映射
func (s *TenantDomainService) Delete(userID int64, id int64) error {
	// 权限校验
	if err := s.assertSuperAdmin(userID); err != nil {
		return err
	}

	// 查找记录（用于清缓存）
	td, err := s.domainRepo.FindByID(s.db, id)
	if err != nil {
		return err
	}
	if td == nil {
		return errors.NewAuthError(errors.ErrEntityNotFound, "域名映射不存在")
	}

	if err := s.domainRepo.SoftDelete(s.db, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewAuthError(errors.ErrEntityNotFound, "域名映射不存在")
		}
		return err
	}

	// 清除缓存
	s.cache.Delete(tenantDomainCachePrefix + td.Domain)
	return nil
}

// List 分页查询域名映射列表（含租户名称）
func (s *TenantDomainService) List(params repository.TenantDomainListParams) ([]TenantDomainListItem, int64, error) {
	list, total, err := s.domainRepo.List(s.db, params)
	if err != nil {
		return nil, 0, err
	}

	// 批量补充租户信息
	items := make([]TenantDomainListItem, 0, len(list))
	for _, td := range list {
		item := TenantDomainListItem{
			ID:        td.ID,
			Domain:    td.Domain,
			MatchType: td.MatchType,
			TenantID:  td.TenantID,
			Remark:    td.Remark,
			Version:   td.Version,
		}
		if td.CreatedAt != nil {
			item.CreatedAt = td.CreatedAt.Format("2006-01-02 15:04:05")
		}
		// 补充租户编码和名称
		tenant, err := s.tenantRepo.FindByID(s.db, td.TenantID)
		if err == nil {
			item.TenantCode = tenant.TenantCode
			item.TenantName = tenant.Name
		}
		items = append(items, item)
	}

	return items, total, nil
}

// assertSuperAdmin 校验用户是否拥有 SUPER_ADMIN 角色
func (s *TenantDomainService) assertSuperAdmin(userID int64) error {
	var count int64
	err := s.db.Table("admin_user_role").
		Joins("JOIN admin_role ON admin_role.id = admin_user_role.role_id").
		Where("admin_user_role.user_id = ? AND admin_role.role_type = ?", userID, "SUPER_ADMIN").
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.NewAuthError(errors.ErrPermissionDenied, "仅超级管理员可执行此操作")
	}
	return nil
}

// validateDomain 校验域名格式
func (s *TenantDomainService) validateDomain(domain string) error {
	if len(domain) > 255 {
		return errors.NewAuthError(errors.ErrInvalidParam, fmt.Sprintf("域名长度不能超过 255 字符，当前 %d", len(domain)))
	}

	lower := strings.ToLower(domain)
	for _, reserved := range reservedDomains {
		if lower == reserved {
			return errors.NewAuthError(errors.ErrInvalidParam, fmt.Sprintf("'%s' 为保留域名，不允许配置", domain))
		}
	}

	return nil
}
