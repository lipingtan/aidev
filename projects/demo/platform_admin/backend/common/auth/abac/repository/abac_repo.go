package repository

import (
	"go-admin/common/auth/abac/model"
	autherrors "go-admin/common/auth/errors"

	"gorm.io/gorm"
)

// ---- AbacPolicy Repository ----

// AbacPolicyRepository 策略数据访问接口
type AbacPolicyRepository interface {
	Create(db *gorm.DB, policy *model.AbacPolicy) error
	Update(db *gorm.DB, policy *model.AbacPolicy) error // 乐观锁
	Delete(db *gorm.DB, id int64) error
	FindByID(db *gorm.DB, id int64) (*model.AbacPolicy, error)
	ListByTenant(db *gorm.DB, tenantID int64, resourceType string, page, pageSize int) ([]*model.AbacPolicy, int64, error)
	FindByResourceType(db *gorm.DB, tenantID int64, resourceType string) ([]*model.AbacPolicy, error)
}

type abacPolicyRepo struct{}

// NewAbacPolicyRepository 创建实例
func NewAbacPolicyRepository() AbacPolicyRepository {
	return &abacPolicyRepo{}
}

func (r *abacPolicyRepo) Create(db *gorm.DB, policy *model.AbacPolicy) error {
	return db.Create(policy).Error
}

// Update 乐观锁更新：WHERE id = ? AND version = ?，更新后 version+1
func (r *abacPolicyRepo) Update(db *gorm.DB, policy *model.AbacPolicy) error {
	oldVersion := policy.Version
	result := db.Model(policy).
		Where("id = ? AND version = ? AND deleted_at IS NULL", policy.ID, oldVersion).
		Updates(map[string]interface{}{
			"name":          policy.Name,
			"resource_type": policy.ResourceType,
			"subject_type":  policy.SubjectType,
			"subject_id":    policy.SubjectID,
			"effect":        policy.Effect,
			"priority":      policy.Priority,
			"status":        policy.Status,
			"description":   policy.Description,
			"version":       oldVersion + 1,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return autherrors.NewAuthError(autherrors.ErrOptimisticLock, "策略已被他人修改，请刷新后重试")
	}
	policy.Version = oldVersion + 1
	return nil
}

func (r *abacPolicyRepo) Delete(db *gorm.DB, id int64) error {
	result := db.Where("id = ? AND deleted_at IS NULL", id).Delete(&model.AbacPolicy{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return autherrors.NewAuthError(autherrors.ErrEntityNotFound, "策略不存在")
	}
	return nil
}

func (r *abacPolicyRepo) FindByID(db *gorm.DB, id int64) (*model.AbacPolicy, error) {
	var policy model.AbacPolicy
	err := db.Where("id = ? AND deleted_at IS NULL", id).First(&policy).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, autherrors.NewAuthError(autherrors.ErrEntityNotFound, "策略不存在")
		}
		return nil, err
	}
	return &policy, nil
}

// ListByTenant 分页查询：tenant_id 匹配或 tenant_id=0（平台级策略对所有租户可见）
func (r *abacPolicyRepo) ListByTenant(db *gorm.DB, tenantID int64, resourceType string, page, pageSize int) ([]*model.AbacPolicy, int64, error) {
	q := db.Where("(tenant_id = ? OR tenant_id = 0) AND deleted_at IS NULL", tenantID)
	if resourceType != "" {
		q = q.Where("resource_type = ?", resourceType)
	}
	var total int64
	if err := q.Model(&model.AbacPolicy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.AbacPolicy
	offset := (page - 1) * pageSize
	err := q.Order("priority ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// FindByResourceType 查找指定租户+资源类型的所有启用策略（含平台级 tenant_id=0）
func (r *abacPolicyRepo) FindByResourceType(db *gorm.DB, tenantID int64, resourceType string) ([]*model.AbacPolicy, error) {
	var list []*model.AbacPolicy
	err := db.Where("(tenant_id = ? OR tenant_id = 0) AND resource_type = ? AND status = 1 AND deleted_at IS NULL",
		tenantID, resourceType).
		Order("priority ASC").
		Find(&list).Error
	return list, err
}

// ---- AbacRowPolicy Repository ----

// AbacRowPolicyRepository 行权限数据访问接口
type AbacRowPolicyRepository interface {
	ReplaceByPolicy(db *gorm.DB, policyID int64, rows []model.AbacRowPolicy) error
	FindByPolicy(db *gorm.DB, policyID int64) ([]model.AbacRowPolicy, error)
	FindByPolicies(db *gorm.DB, policyIDs []int64) ([]model.AbacRowPolicy, error)
}

type abacRowPolicyRepo struct{}

// NewAbacRowPolicyRepository 创建实例
func NewAbacRowPolicyRepository() AbacRowPolicyRepository {
	return &abacRowPolicyRepo{}
}

func (r *abacRowPolicyRepo) ReplaceByPolicy(db *gorm.DB, policyID int64, rows []model.AbacRowPolicy) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("policy_id = ?", policyID).Delete(&model.AbacRowPolicy{}).Error; err != nil {
			return err
		}
		for i := range rows {
			rows[i].PolicyID = policyID
			if err := tx.Create(&rows[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *abacRowPolicyRepo) FindByPolicy(db *gorm.DB, policyID int64) ([]model.AbacRowPolicy, error) {
	var list []model.AbacRowPolicy
	err := db.Where("policy_id = ?", policyID).Find(&list).Error
	return list, err
}

func (r *abacRowPolicyRepo) FindByPolicies(db *gorm.DB, policyIDs []int64) ([]model.AbacRowPolicy, error) {
	if len(policyIDs) == 0 {
		return nil, nil
	}
	var list []model.AbacRowPolicy
	err := db.Where("policy_id IN ?", policyIDs).Find(&list).Error
	return list, err
}

// ---- AbacColPolicy Repository ----

// AbacColPolicyRepository 列权限数据访问接口
type AbacColPolicyRepository interface {
	ReplaceByPolicy(db *gorm.DB, policyID int64, cols []model.AbacColPolicy) error
	FindByPolicy(db *gorm.DB, policyID int64) ([]model.AbacColPolicy, error)
	FindByPolicies(db *gorm.DB, policyIDs []int64) ([]model.AbacColPolicy, error)
}

type abacColPolicyRepo struct{}

// NewAbacColPolicyRepository 创建实例
func NewAbacColPolicyRepository() AbacColPolicyRepository {
	return &abacColPolicyRepo{}
}

func (r *abacColPolicyRepo) ReplaceByPolicy(db *gorm.DB, policyID int64, cols []model.AbacColPolicy) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("policy_id = ?", policyID).Delete(&model.AbacColPolicy{}).Error; err != nil {
			return err
		}
		for i := range cols {
			cols[i].PolicyID = policyID
			if err := tx.Create(&cols[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *abacColPolicyRepo) FindByPolicy(db *gorm.DB, policyID int64) ([]model.AbacColPolicy, error) {
	var list []model.AbacColPolicy
	err := db.Where("policy_id = ?", policyID).Find(&list).Error
	return list, err
}

func (r *abacColPolicyRepo) FindByPolicies(db *gorm.DB, policyIDs []int64) ([]model.AbacColPolicy, error) {
	if len(policyIDs) == 0 {
		return nil, nil
	}
	var list []model.AbacColPolicy
	err := db.Where("policy_id IN ?", policyIDs).Find(&list).Error
	return list, err
}
