package service

import (
	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

// RecordShareService 记录共享业务服务
type RecordShareService struct {
	db   *gorm.DB
	repo repository.RecordShareRepository
}

// NewRecordShareService 创建 RecordShareService 实例
func NewRecordShareService(db *gorm.DB, repo repository.RecordShareRepository) *RecordShareService {
	return &RecordShareService{db: db, repo: repo}
}

// Create 创建共享规则
func (s *RecordShareService) Create(share *model.RecordShare) (*model.RecordShare, error) {
	if err := s.repo.Create(s.db, share); err != nil {
		return nil, err
	}
	return share, nil
}

// ListByRecord 查询记录的共享规则列表
func (s *RecordShareService) ListByRecord(tenantID int64, objectCode string, recordID int64) ([]model.RecordShare, error) {
	return s.repo.ListByRecord(s.db, tenantID, objectCode, recordID)
}

// Delete 删除共享规则
func (s *RecordShareService) Delete(id int64) error {
	return s.repo.Delete(s.db, id)
}
