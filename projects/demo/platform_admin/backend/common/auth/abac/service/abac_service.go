// Package service 提供 ABAC 策略的业务逻辑：策略 CRUD、缓存管理、评估。
package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"go-admin/common/auth/abac/engine"
	"go-admin/common/auth/abac/model"
	"go-admin/common/auth/abac/repository"
	autherrors "go-admin/common/auth/errors"
	"go-admin/common/event"

	"github.com/go-admin-team/go-admin-core/storage"
	"gorm.io/gorm"
)

// AbacPolicyWithDetail 策略详情（含行/列权限 + 主体显示名）
type AbacPolicyWithDetail struct {
	model.AbacPolicy
	SubjectDisplayName string                  `json:"subject_display_name"`
	RowPolicies        []model.AbacRowPolicy   `json:"row_policies"`
	ColPolicies        []model.AbacColPolicy   `json:"col_policies"`
}

// CreateAbacPolicyRequest 创建策略请求
type CreateAbacPolicyRequest struct {
	Name         string                `json:"name"        binding:"required"`
	ResourceType string                `json:"resource_type" binding:"required"`
	SubjectType  string                `json:"subject_type"  binding:"required"`
	SubjectID    string                `json:"subject_id"    binding:"required"`
	Effect       string                `json:"effect"        binding:"required"`
	Priority     int                   `json:"priority"`
	Description  string                `json:"description"`
	RowPolicies  []RowPolicyInput      `json:"row_policies"`
	ColPolicies  []ColPolicyInput      `json:"col_policies"`
}

// UpdateAbacPolicyRequest 更新策略请求
type UpdateAbacPolicyRequest struct {
	Name        string            `json:"name"`
	Effect      string            `json:"effect"`
	Priority    int               `json:"priority"`
	Status      *int              `json:"status"`
	Description string            `json:"description"`
	Version     int               `json:"version"     binding:"required"`
	RowPolicies []RowPolicyInput  `json:"row_policies"`
	ColPolicies []ColPolicyInput  `json:"col_policies"`
}

// RowPolicyInput 行权限输入
type RowPolicyInput struct {
	Action        string           `json:"action" binding:"required"`
	ConditionExpr *engine.CondNode `json:"condition_expr"`
}

// ColPolicyInput 列权限输入
type ColPolicyInput struct {
	FieldName   string `json:"field_name"   binding:"required"`
	Effect      string `json:"effect"       binding:"required"`
	MaskType    string `json:"mask_type"`
	MaskPattern string `json:"mask_pattern"`
}

// ListAbacPoliciesRequest 列表查询请求
type ListAbacPoliciesRequest struct {
	TenantID     int64  `json:"-"`
	ResourceType string `form:"resource_type"`
	SubjectType  string `form:"subject_type"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

// AbacService 策略业务服务
type AbacService struct {
	db         *gorm.DB
	policyRepo repository.AbacPolicyRepository
	rowRepo    repository.AbacRowPolicyRepository
	colRepo    repository.AbacColPolicyRepository
	cache      storage.AdapterCache
}

// NewAbacService 创建 AbacService
func NewAbacService(
	db *gorm.DB,
	policyRepo repository.AbacPolicyRepository,
	rowRepo repository.AbacRowPolicyRepository,
	colRepo repository.AbacColPolicyRepository,
	cache storage.AdapterCache,
) *AbacService {
	svc := &AbacService{
		db:         db,
		policyRepo: policyRepo,
		rowRepo:    rowRepo,
		colRepo:    colRepo,
		cache:      cache,
	}
	// 订阅缓存失效事件：按 resource_type 范围清缓存
	// 由于 storage.AdapterCache 无 DeleteByPrefix，改用记录已知 key 前缀标记失效
	// 实际清缓存在 LoadPoliciesForCallback 的 "写入时附加失效标记" 机制中实现：
	// 每次 CRUD 后，用一个版本 key 来使同 resource_type 的缓存失效
	event.DefaultBus.Subscribe(event.EventAbacPolicyChanged, func(payload interface{}) {
		ev, ok := payload.(*event.AbacPolicyChangedEvent)
		if !ok {
			return
		}
		if cache == nil {
			return
		}
		// 递增版本号，使该 resource_type 下所有用户缓存失效
		versionKey := fmt.Sprintf("abac_ver:%d:%s", ev.TenantID, ev.ResourceType)
		_ = cache.Increase(versionKey)
	})
	return svc
}

// Create 创建策略（事务：主表 + 行/列权限一起写入）
func (s *AbacService) Create(tenantID int64, req *CreateAbacPolicyRequest) (*model.AbacPolicy, error) {
	// 名称唯一性校验（同租户下 deleted_at IS NULL）
	var count int64
	s.db.Model(&model.AbacPolicy{}).
		Where("tenant_id = ? AND name = ? AND deleted_at IS NULL", tenantID, req.Name).
		Count(&count)
	if count > 0 {
		return nil, autherrors.NewAuthError(autherrors.ErrDuplicateEntity, "策略名称在当前租户内已存在")
	}

	priority := req.Priority
	if priority == 0 {
		priority = 100
	}

	policy := &model.AbacPolicy{
		TenantID:     tenantID,
		Name:         req.Name,
		ResourceType: req.ResourceType,
		SubjectType:  req.SubjectType,
		SubjectID:    req.SubjectID,
		Effect:       req.Effect,
		Priority:     priority,
		Status:       1,
		Description:  req.Description,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.policyRepo.Create(tx, policy); err != nil {
			return err
		}
		if err := s.saveRowPolicies(tx, policy.ID, req.RowPolicies); err != nil {
			return err
		}
		return s.saveColPolicies(tx, policy.ID, req.ColPolicies)
	})
	if err != nil {
		return nil, err
	}

	// 事务成功后发布缓存失效事件
	event.DefaultBus.Publish(event.EventAbacPolicyChanged, &event.AbacPolicyChangedEvent{
		TenantID:     tenantID,
		ResourceType: req.ResourceType,
	})

	return policy, nil
}

// Update 更新策略（乐观锁）
func (s *AbacService) Update(tenantID, id int64, req *UpdateAbacPolicyRequest) error {
	policy, err := s.policyRepo.FindByID(s.db, id)
	if err != nil {
		return err
	}
	if policy.TenantID != tenantID && policy.TenantID != 0 {
		return autherrors.NewAuthError(autherrors.ErrPermissionDenied, "无权修改此策略")
	}

	policy.Version = req.Version
	if req.Name != "" {
		policy.Name = req.Name
	}
	if req.Effect != "" {
		policy.Effect = req.Effect
	}
	if req.Priority != 0 {
		policy.Priority = req.Priority
	}
	if req.Status != nil {
		policy.Status = *req.Status
	}
	if req.Description != "" {
		policy.Description = req.Description
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.policyRepo.Update(tx, policy); err != nil {
			return err
		}
		if req.RowPolicies != nil {
			if err := s.saveRowPolicies(tx, id, req.RowPolicies); err != nil {
				return err
			}
		}
		if req.ColPolicies != nil {
			if err := s.saveColPolicies(tx, id, req.ColPolicies); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	event.DefaultBus.Publish(event.EventAbacPolicyChanged, &event.AbacPolicyChangedEvent{
		TenantID:     tenantID,
		ResourceType: policy.ResourceType,
	})
	return nil
}

// Delete 删除策略（软删除）
func (s *AbacService) Delete(tenantID, id int64) error {
	policy, err := s.policyRepo.FindByID(s.db, id)
	if err != nil {
		return err
	}
	if policy.TenantID != tenantID && policy.TenantID != 0 {
		return autherrors.NewAuthError(autherrors.ErrPermissionDenied, "无权删除此策略")
	}
	resourceType := policy.ResourceType

	if err := s.policyRepo.Delete(s.db, id); err != nil {
		return err
	}

	event.DefaultBus.Publish(event.EventAbacPolicyChanged, &event.AbacPolicyChangedEvent{
		TenantID:     tenantID,
		ResourceType: resourceType,
	})
	return nil
}

// GetByID 查询策略详情（含行/列权限）
func (s *AbacService) GetByID(id int64) (*AbacPolicyWithDetail, error) {
	policy, err := s.policyRepo.FindByID(s.db, id)
	if err != nil {
		return nil, err
	}
	rows, err := s.rowRepo.FindByPolicy(s.db, id)
	if err != nil {
		return nil, err
	}
	cols, err := s.colRepo.FindByPolicy(s.db, id)
	if err != nil {
		return nil, err
	}
	return &AbacPolicyWithDetail{
		AbacPolicy:  *policy,
		RowPolicies: rows,
		ColPolicies: cols,
	}, nil
}

// List 分页查询策略列表
func (s *AbacService) List(req *ListAbacPoliciesRequest) ([]*AbacPolicyWithDetail, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	policies, total, err := s.policyRepo.ListByTenant(s.db, req.TenantID, req.ResourceType, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*AbacPolicyWithDetail, 0, len(policies))
	for _, p := range policies {
		detail := &AbacPolicyWithDetail{AbacPolicy: *p}
		detail.SubjectDisplayName = s.resolveSubjectName(p)
		result = append(result, detail)
	}
	return result, total, nil
}

// LoadPoliciesForCallback 供 GORM Callback 使用：加载策略并转换为引擎格式
// 优先从缓存读取，缓存不可用时直查 DB
// 缓存失效通过版本号机制：策略变更时递增 abac_ver:{tenant}:{resource_type}
func (s *AbacService) LoadPoliciesForCallback(tenantID int64, userID string, resourceType string) ([]engine.PolicyRowEntry, []engine.PolicyColEntry) {
	cacheKey := fmt.Sprintf("abac:%d:%s:%s", tenantID, userID, resourceType)
	versionKey := fmt.Sprintf("abac_ver:%d:%s", tenantID, resourceType)

	// 尝试缓存：读取数据 + 版本号，不一致则穿透
	if s.cache != nil {
		rawData, dataErr := s.cache.Get(cacheKey)
		rawVer, verErr := s.cache.Get(versionKey + ":snap:" + userID)
		curVer, _ := s.cache.Get(versionKey)
		if dataErr == nil && verErr == nil && rawVer == curVer {
			var cached struct {
				RowEntries []engine.PolicyRowEntry `json:"row"`
				ColEntries []engine.PolicyColEntry `json:"col"`
			}
			if err := json.Unmarshal([]byte(rawData), &cached); err == nil {
				return cached.RowEntries, cached.ColEntries
			}
		}
	}

	// 查 DB
	policies, err := s.policyRepo.FindByResourceType(s.db, tenantID, resourceType)
	if err != nil {
		log.Printf("[abac] LoadPolicies error: %v", err)
		return nil, nil
	}

	if len(policies) == 0 {
		return nil, nil
	}

	policyIDs := make([]int64, len(policies))
	for i, p := range policies {
		policyIDs[i] = p.ID
	}

	rows, _ := s.rowRepo.FindByPolicies(s.db, policyIDs)
	cols, _ := s.colRepo.FindByPolicies(s.db, policyIDs)

	// 建 map
	policyMap := make(map[int64]*model.AbacPolicy, len(policies))
	for _, p := range policies {
		policyMap[p.ID] = p
	}

	var rowEntries []engine.PolicyRowEntry
	var colEntries []engine.PolicyColEntry

	for _, row := range rows {
		p, ok := policyMap[row.PolicyID]
		if !ok {
			continue
		}
		var node *engine.CondNode
		if len(row.ConditionExpr) > 0 {
			node = &engine.CondNode{}
			if err := json.Unmarshal(row.ConditionExpr, node); err != nil {
				log.Printf("[abac] unmarshal condition_expr policyID=%d: %v", row.PolicyID, err)
				node = nil
			}
		}
		rowEntries = append(rowEntries, engine.PolicyRowEntry{
			Effect:        p.Effect,
			Action:        row.Action,
			ConditionNode: node,
		})
	}

	for _, col := range cols {
		colEntries = append(colEntries, engine.PolicyColEntry{
			FieldName:   col.FieldName,
			Effect:      col.Effect,
			MaskType:    col.MaskType,
			MaskPattern: col.MaskPattern,
		})
	}

	// 写缓存：序列化数据 + 快照当前版本号（TTL 10 分钟）
	if s.cache != nil {
		cached := struct {
			RowEntries []engine.PolicyRowEntry `json:"row"`
			ColEntries []engine.PolicyColEntry `json:"col"`
		}{RowEntries: rowEntries, ColEntries: colEntries}
		if raw, err := json.Marshal(cached); err == nil {
			ttlSeconds := 600 // 10 分钟
			_ = s.cache.Set(cacheKey, string(raw), ttlSeconds)
			// 快照当前版本号
			curVer, _ := s.cache.Get(versionKey)
			_ = s.cache.Set(versionKey+":snap:"+userID, curVer, ttlSeconds)
		}
	}

	return rowEntries, colEntries
}

// saveRowPolicies 保存行权限（先删后插）
func (s *AbacService) saveRowPolicies(tx *gorm.DB, policyID int64, inputs []RowPolicyInput) error {
	rows := make([]model.AbacRowPolicy, 0, len(inputs))
	for _, input := range inputs {
		var condJSON []byte
		if input.ConditionExpr != nil {
			var err error
			condJSON, err = json.Marshal(input.ConditionExpr)
			if err != nil {
				return fmt.Errorf("marshal condition_expr: %w", err)
			}
		}
		rows = append(rows, model.AbacRowPolicy{
			PolicyID:      policyID,
			Action:        input.Action,
			ConditionExpr: condJSON,
		})
	}
	return s.rowRepo.ReplaceByPolicy(tx, policyID, rows)
}

// saveColPolicies 保存列权限（先删后插）
func (s *AbacService) saveColPolicies(tx *gorm.DB, policyID int64, inputs []ColPolicyInput) error {
	cols := make([]model.AbacColPolicy, 0, len(inputs))
	for _, input := range inputs {
		cols = append(cols, model.AbacColPolicy{
			PolicyID:    policyID,
			FieldName:   input.FieldName,
			Effect:      input.Effect,
			MaskType:    input.MaskType,
			MaskPattern: input.MaskPattern,
		})
	}
	return s.colRepo.ReplaceByPolicy(tx, policyID, cols)
}

// resolveSubjectName 解析主体显示名（简单实现：直接返回 subject_id，可后续扩展查库）
func (s *AbacService) resolveSubjectName(p *model.AbacPolicy) string {
	switch p.SubjectType {
	case engine.SubjectUser:
		return "用户:" + p.SubjectID
	case engine.SubjectRole, engine.SubjectPermissionSet:
		return "角色:" + p.SubjectID
	case engine.SubjectDept:
		return "部门:" + p.SubjectID
	default:
		return p.SubjectID
	}
}

// userIDToString 辅助：int64 user_id → string
func userIDToString(id int64) string {
	return strconv.FormatInt(id, 10)
}
