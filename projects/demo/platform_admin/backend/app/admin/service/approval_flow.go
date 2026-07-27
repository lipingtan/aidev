package service

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
)

// ApprovalFlowService 审批流定义 CRUD 服务
type ApprovalFlowService struct {
	service.Service
}

// GetPage 分页查询审批流定义（租户隔离）
func (e *ApprovalFlowService) GetPage(req *dto.ApprovalFlowGetPageReq, tenantID int64, isSuperAdmin bool, list *[]models.AdminApprovalFlow, count *int64) error {
	db := e.Orm.Model(&models.AdminApprovalFlow{})

	// 租户隔离：SUPER_ADMIN 可查所有（含 tenant_id=0），TENANT_ADMIN 只查本租户
	if !isSuperAdmin {
		db = db.Where("tenant_id = ?", tenantID)
	}

	if req.FlowCode != "" {
		db = db.Where("flow_code LIKE ?", "%"+req.FlowCode+"%")
	}
	if req.FlowName != "" {
		db = db.Where("flow_name LIKE ?", "%"+req.FlowName+"%")
	}

	if err := db.Count(count).Error; err != nil {
		e.Log.Errorf("count approval flow error: %s", err)
		return err
	}

	pageSize := req.GetPageSize()
	pageIndex := req.GetPageIndex()
	offset := (pageIndex - 1) * pageSize

	if err := db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(list).Error; err != nil {
		e.Log.Errorf("find approval flow error: %s", err)
		return err
	}
	return nil
}

// Get 查询单条审批流定义
func (e *ApprovalFlowService) Get(req *dto.ApprovalFlowGetReq, tenantID int64, isSuperAdmin bool, model *models.AdminApprovalFlow) error {
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return errors.New("无效的 ID")
	}

	db := e.Orm.Where("id = ?", id)
	if !isSuperAdmin {
		db = db.Where("tenant_id = ?", tenantID)
	}

	if err := db.First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("记录不存在或无权访问")
		}
		e.Log.Errorf("get approval flow error: %s", err)
		return err
	}
	return nil
}

// Insert 创建审批流定义
func (e *ApprovalFlowService) Insert(req *dto.ApprovalFlowInsertReq, tenantID int64, isSuperAdmin bool) error {
	// TENANT_ADMIN 不能创建 tenant_id=0 的全局流程
	if !isSuperAdmin && tenantID == 0 {
		return errors.New("无权创建全局流程")
	}

	// 唯一性校验：同 tenant_id + 同 flow_code 且未软删除
	var count int64
	if err := e.Orm.Model(&models.AdminApprovalFlow{}).
		Where("tenant_id = ? AND flow_code = ? AND deleted_at IS NULL", tenantID, req.FlowCode).
		Count(&count).Error; err != nil {
		e.Log.Errorf("check flow_code unique error: %s", err)
		return err
	}
	if count > 0 {
		return errors.New("flow_code 在当前租户下已存在")
	}

	// 序列化 flow_config
	configJSON, err := json.Marshal(req.FlowConfig)
	if err != nil {
		return errors.New("flow_config 序列化失败: " + err.Error())
	}

	data := &models.AdminApprovalFlow{
		TenantID:    tenantID,
		FlowCode:    req.FlowCode,
		FlowName:    req.FlowName,
		FlowConfig:  configJSON,
		Description: req.Description,
		CreateBy:    req.CreateBy,
		UpdateBy:    req.CreateBy,
	}

	if err := e.Orm.Create(data).Error; err != nil {
		e.Log.Errorf("create approval flow error: %s", err)
		return err
	}
	return nil
}

// Update 更新审批流定义
func (e *ApprovalFlowService) Update(req *dto.ApprovalFlowUpdateReq, tenantID int64, isSuperAdmin bool) error {
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return errors.New("无效的 ID")
	}

	var data models.AdminApprovalFlow
	db := e.Orm.Where("id = ?", id)
	if !isSuperAdmin {
		db = db.Where("tenant_id = ?", tenantID)
	}
	if err := db.First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("记录不存在或无权操作")
		}
		return err
	}

	// TENANT_ADMIN 不能修改 tenant_id=0 的全局流程
	if !isSuperAdmin && data.TenantID == 0 {
		return errors.New("无权修改全局流程")
	}

	updates := map[string]interface{}{
		"update_by": req.UpdateBy,
	}
	if req.FlowName != "" {
		updates["flow_name"] = req.FlowName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if len(req.FlowConfig) > 0 {
		configJSON, err := json.Marshal(req.FlowConfig)
		if err != nil {
			return errors.New("flow_config 序列化失败: " + err.Error())
		}
		updates["flow_config"] = configJSON
	}

	if err := e.Orm.Model(&data).Updates(updates).Error; err != nil {
		e.Log.Errorf("update approval flow error: %s", err)
		return err
	}
	return nil
}

// Delete 删除审批流定义（软删除，有 PENDING 实例时拒绝）
func (e *ApprovalFlowService) Delete(req *dto.ApprovalFlowDeleteReq, tenantID int64, isSuperAdmin bool) error {
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return errors.New("无效的 ID")
	}

	var data models.AdminApprovalFlow
	db := e.Orm.Where("id = ?", id)
	if !isSuperAdmin {
		db = db.Where("tenant_id = ?", tenantID)
	}
	if err := db.First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("记录不存在或无权操作")
		}
		return err
	}

	// TENANT_ADMIN 不能删除 tenant_id=0 的全局流程
	if !isSuperAdmin && data.TenantID == 0 {
		return errors.New("无权删除全局流程")
	}

	// 删除前校验：有 PENDING 状态的审批实例时拒绝删除
	var pendingCount int64
	if err := e.Orm.Model(&models.AdminApproval{}).
		Where("flow_id = ? AND status = 'PENDING'", id).
		Count(&pendingCount).Error; err != nil {
		e.Log.Errorf("check pending approval error: %s", err)
		return err
	}
	if pendingCount > 0 {
		return errors.New("存在进行中的审批实例，无法删除流程定义")
	}

	if err := e.Orm.Delete(&data).Error; err != nil {
		e.Log.Errorf("delete approval flow error: %s", err)
		return err
	}
	return nil
}
