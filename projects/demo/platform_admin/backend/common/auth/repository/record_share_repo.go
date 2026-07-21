package repository

import (
	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// RecordShareRepository 记录共享数据访问接口
type RecordShareRepository interface {
	// Create 创建共享规则
	Create(db *gorm.DB, share *model.RecordShare) error
	// ListByRecord 查询记录的共享规则列表
	ListByRecord(db *gorm.DB, tenantID int64, objectCode string, recordID int64) ([]model.RecordShare, error)
	// Delete 根据ID删除共享规则
	Delete(db *gorm.DB, id int64) error
	// FindByCondition 查询命中共享规则的 record_id 列表
	FindByCondition(db *gorm.DB, tenantID int64, objectCode string, recordID int64, userID int64, roleIDs []int64, deptIDs []int64) ([]int64, error)
}

// recordShareRepo RecordShareRepository 默认实现
type recordShareRepo struct{}

// NewRecordShareRepository 创建 RecordShareRepository 实例
func NewRecordShareRepository() RecordShareRepository {
	return &recordShareRepo{}
}

// Create 创建共享规则
func (r *recordShareRepo) Create(db *gorm.DB, share *model.RecordShare) error {
	return db.Create(share).Error
}

// ListByRecord 查询记录的共享规则列表（按租户隔离，过滤已过期规则）
func (r *recordShareRepo) ListByRecord(db *gorm.DB, tenantID int64, objectCode string, recordID int64) ([]model.RecordShare, error) {
	var list []model.RecordShare
	err := db.Where("tenant_id = ? AND object_code = ? AND record_id = ?", tenantID, objectCode, recordID).
		Where("expire_at IS NULL OR expire_at > NOW()").
		Find(&list).Error
	return list, err
}

// Delete 根据ID删除共享规则
func (r *recordShareRepo) Delete(db *gorm.DB, id int64) error {
	result := db.Delete(&model.RecordShare{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindByCondition 查询命中共享规则的 record_id 列表
// 根据用户ID、角色ID列表、部门ID列表匹配 share_to_type 和 share_to_id，返回有效的 record_id
func (r *recordShareRepo) FindByCondition(db *gorm.DB, tenantID int64, objectCode string, recordID int64, userID int64, roleIDs []int64, deptIDs []int64) ([]int64, error) {
	var recordIDs []int64

	query := db.Model(&model.RecordShare{}).
		Select("DISTINCT record_id").
		Where("tenant_id = ? AND object_code = ?", tenantID, objectCode).
		Where("(expire_at IS NULL OR expire_at > NOW())")

	// 如果指定了 recordID，则只查该记录
	if recordID > 0 {
		query = query.Where("record_id = ?", recordID)
	}

	// 构建共享目标匹配条件
	conditions := db.Where("share_to_type = 'USER' AND share_to_id = ?", userID)
	if len(roleIDs) > 0 {
		conditions = conditions.Or("share_to_type = 'ROLE' AND share_to_id IN ?", roleIDs)
	}
	if len(deptIDs) > 0 {
		conditions = conditions.Or("share_to_type = 'DEPT' AND share_to_id IN ?", deptIDs)
	}

	query = query.Where(conditions)
	err := query.Pluck("record_id", &recordIDs).Error
	return recordIDs, err
}
