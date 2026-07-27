package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
	authModel "go-admin/common/auth/model"
	"go-admin/common/event"
)

// ApprovalService 审批实例 Service
type ApprovalService struct {
	service.Service
}

// InitApproval 发起审批
// 单一事务内完成：创建实例 + 创建所有节点 + 可选更新 subscription_status
func (e *ApprovalService) InitApproval(req *dto.InitApprovalReq) (approvalID int64, err error) {
	db := e.Orm

	// 1. 查找流程定义：优先租户自定义，无则全局
	var flow models.AdminApprovalFlow
	err = db.Where("tenant_id = ? AND flow_code = ? AND deleted_at IS NULL", req.TenantID, req.FlowCode).
		First(&flow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 降级查全局
		err = db.Where("tenant_id = 0 AND flow_code = ? AND deleted_at IS NULL", req.FlowCode).
			First(&flow).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("流程定义不存在: flow_code=%s", req.FlowCode)
		}
		return 0, fmt.Errorf("查询流程定义失败: %w", err)
	}

	// 2. 解析 flow_config 节点列表
	var nodeCfgs []dto.ApprovalNodeConfig
	if err = json.Unmarshal(flow.FlowConfig, &nodeCfgs); err != nil {
		return 0, fmt.Errorf("解析 flow_config 失败: %w", err)
	}
	if len(nodeCfgs) == 0 {
		return 0, errors.New("流程定义节点为空")
	}

	// 3. flow_snapshot = flow_config 深拷贝（复制原始 JSON bytes）
	snapshotBytes := make([]byte, len(flow.FlowConfig))
	copy(snapshotBytes, flow.FlowConfig)

	// 4. 开启事务
	tx := db.Begin()
	if tx.Error != nil {
		return 0, fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// 4a. 创建审批实例
	approval := models.AdminApproval{
		TenantID:     req.TenantID,
		FlowID:       flow.ID,
		FlowCode:     flow.FlowCode,
		FlowSnapshot: snapshotBytes,
		BizType:      req.BizType,
		BizID:        req.BizID,
		Status:       "PENDING",
		ApplicantID:  req.ApplicantID,
	}
	if err = tx.Create(&approval).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("创建审批实例失败: %w", err)
	}

	// 4b. 逐节点创建
	for i, cfg := range nodeCfgs {
		// 解析审批人用户 ID 列表
		userIDs, note, resolveErr := e.resolveAssigneeUserIDs(tx, cfg, req.TenantID)
		if resolveErr != nil {
			tx.Rollback()
			return 0, fmt.Errorf("解析节点 %d 审批人失败: %w", cfg.NodeOrder, resolveErr)
		}

		assigneeIDsJSON, _ := json.Marshal(cfg.AssigneeIDs)
		assigneeUserIDsJSON, _ := json.Marshal(userIDs)

		status := "WAITING"
		var timeoutAt *time.Time
		if i == 0 {
			status = "PENDING"
			if cfg.TimeoutHours > 0 {
				t := time.Now().Add(time.Duration(cfg.TimeoutHours) * time.Hour)
				timeoutAt = &t
			}
		}

		node := models.AdminApprovalNode{
			TenantID:        req.TenantID,
			ApprovalID:      approval.ID,
			NodeOrder:       cfg.NodeOrder,
			NodeType:        cfg.NodeType,
			AssigneeType:    cfg.AssigneeType,
			AssigneeIDs:     assigneeIDsJSON,
			AssigneeUserIDs: assigneeUserIDsJSON,
			Status:          status,
			AssigneeNote:    note,
			TimeoutHours:    cfg.TimeoutHours,
			TimeoutAction:   cfg.TimeoutAction,
			EscalateTo:      cfg.EscalateTo,
			TimeoutAt:       timeoutAt,
		}
		if err = tx.Create(&node).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("创建审批节点 %d 失败: %w", cfg.NodeOrder, err)
		}
	}

	// 4c. biz_type=tenant_app_subscription 时更新 subscription_status
	if req.BizType == "tenant_app_subscription" {
		if err = tx.Model(&authModel.TenantApp{}).
			Where("app_code = ? AND tenant_id = ?", req.BizID, req.TenantID).
			Update("subscription_status", "pending_approval").Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("更新 subscription_status 失败: %w", err)
		}
	}

	// 5. 提交事务
	if err = tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("提交事务失败: %w", err)
	}

	return approval.ID, nil
}

// resolveAssigneeUserIDs 解析审批人配置 → 实际用户 ID 列表（字符串）
// 返回 userIDs（字符串列表）、assigneeNote（降级说明）、error
func (e *ApprovalService) resolveAssigneeUserIDs(
	tx *gorm.DB,
	cfg dto.ApprovalNodeConfig,
	tenantID int64,
) (userIDs []string, note string, err error) {

	switch cfg.AssigneeType {
	case "USER":
		// 直接使用配置中的用户 ID 列表
		userIDs = cfg.AssigneeIDs

	case "ROLE":
		// 查 sys_user WHERE role_id IN (assignee_ids) AND tenant_id = ?
		// assignee_ids 存的是角色 ID 字符串，转换为 int
		roleIntIDs := make([]int, 0, len(cfg.AssigneeIDs))
		for _, rid := range cfg.AssigneeIDs {
			var roleID int
			if _, scanErr := fmt.Sscanf(rid, "%d", &roleID); scanErr == nil {
				roleIntIDs = append(roleIntIDs, roleID)
			}
		}
		if len(roleIntIDs) == 0 {
			return nil, "", errors.New("ROLE 类型审批人配置中角色 ID 为空")
		}

		var users []models.SysUser
		if err = tx.Select("user_id").
			Where("role_id IN ? AND tenant_id = ?", roleIntIDs, int(tenantID)).
			Find(&users).Error; err != nil {
			return nil, "", fmt.Errorf("查询角色用户失败: %w", err)
		}
		for _, u := range users {
			userIDs = append(userIDs, fmt.Sprintf("%d", u.UserId))
		}
		if len(userIDs) == 0 {
			// 降级 SUPER_ADMIN
			userIDs = []string{"1"}
			note = fmt.Sprintf("角色 %v 下无用户，降级为 SUPER_ADMIN", cfg.AssigneeIDs)
		}

	case "DEPT_HEAD":
		// 1. 查申请人的部门
		// assignee_ids 此时为空或包含申请人标记，实际通过 tenantID 上下文解析
		// 这里 assignee_ids 中存的是申请人 user_id（在 DEPT_HEAD 场景下由调用方填入）
		if len(cfg.AssigneeIDs) == 0 {
			return nil, "", errors.New("DEPT_HEAD 类型 assignee_ids 需要包含申请人 user_id")
		}

		var applicantIDInt int
		if _, scanErr := fmt.Sscanf(cfg.AssigneeIDs[0], "%d", &applicantIDInt); scanErr != nil {
			return nil, "", fmt.Errorf("DEPT_HEAD 类型申请人 ID 解析失败: %w", scanErr)
		}

		// 2. 查申请人的 dept_id
		var applicant models.SysUser
		if err = tx.Select("user_id, dept_id").
			Where("user_id = ?", applicantIDInt).
			First(&applicant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				userIDs = []string{"1"}
				note = fmt.Sprintf("申请人 %d 不存在，降级为 SUPER_ADMIN", applicantIDInt)
				return userIDs, note, nil
			}
			return nil, "", fmt.Errorf("查询申请人失败: %w", err)
		}

		// 3. 查部门负责人姓名
		var dept models.SysDept
		if err = tx.Select("dept_id, leader").
			Where("dept_id = ?", applicant.DeptId).
			First(&dept).Error; err != nil || dept.Leader == "" {
			// 部门不存在或 leader 为空，降级 SUPER_ADMIN
			userIDs = []string{"1"}
			note = fmt.Sprintf("部门 %d 负责人为空或部门不存在，降级为 SUPER_ADMIN", applicant.DeptId)
			return userIDs, note, nil
		}

		// 4. 通过 leader 姓名查找用户 ID
		var leaderUser models.SysUser
		if err = tx.Select("user_id").
			Where("nick_name = ?", dept.Leader).
			First(&leaderUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				userIDs = []string{"1"}
				note = fmt.Sprintf("部门 %d 负责人「%s」在 sys_user 中未找到，降级为 SUPER_ADMIN", applicant.DeptId, dept.Leader)
				return userIDs, note, nil
			}
			return nil, "", fmt.Errorf("查询部门负责人用户失败: %w", err)
		}

		userIDs = []string{fmt.Sprintf("%d", leaderUser.UserId)}

	default:
		return nil, "", fmt.Errorf("不支持的 assignee_type: %s", cfg.AssigneeType)
	}

	return userIDs, note, nil
}

// ─────────────────────────────────────────────
//  Approve / Reject / Cancel
// ─────────────────────────────────────────────

// Approve 审批通过
func (e *ApprovalService) Approve(approvalID int64, currentUserID int64, comment string) error {
	db := e.Orm

	// 1. 加载审批实例
	var approval models.AdminApproval
	if err := db.First(&approval, approvalID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("审批实例不存在: id=%d", approvalID)
		}
		return fmt.Errorf("查询审批实例失败: %w", err)
	}
	if approval.Status != "PENDING" {
		return fmt.Errorf("审批实例状态为 %s，无法操作", approval.Status)
	}

	// 2. 找到当前 PENDING 节点
	var node models.AdminApprovalNode
	if err := db.Where("approval_id = ? AND status = 'PENDING'", approvalID).
		Order("node_order ASC").First(&node).Error; err != nil {
		return fmt.Errorf("未找到待审批节点: %w", err)
	}

	// 3. 验证当前用户在 assignee_user_ids 中
	if err := e.checkAssignee(node.AssigneeUserIDs, currentUserID); err != nil {
		return err
	}

	// 4. 根据节点类型执行投票逻辑
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var publishEvent string
	var publishPayload interface{}

	switch node.NodeType {
	case "SINGLE":
		// 直接通过
		if err := e.approveNode(tx, &node, comment); err != nil {
			tx.Rollback()
			return err
		}
		// 激活下一节点或完成
		done, err := e.activateNextOrComplete(tx, &approval, node.NodeOrder)
		if err != nil {
			tx.Rollback()
			return err
		}
		if done {
			publishEvent = "approval.completed"
			publishPayload = buildCompletedEvent(&approval)
		}

	case "AND_SIGN":
		// 插入投票记录（UNIQUE KEY 防重复）
		vote := models.AdminApprovalVote{
			TenantID: node.TenantID,
			NodeID:   node.ID,
			UserID:   currentUserID,
			Action:   "APPROVE",
			Comment:  comment,
		}
		if err := tx.Create(&vote).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("投票记录创建失败（可能重复投票）: %w", err)
		}
		// 统计 APPROVE 票数
		var approveCount int64
		tx.Model(&models.AdminApprovalVote{}).
			Where("node_id = ? AND action = 'APPROVE'", node.ID).
			Count(&approveCount)
		// 解析 assignee_user_ids 总人数
		total, err := e.countAssignees(node.AssigneeUserIDs)
		if err != nil {
			tx.Rollback()
			return err
		}
		if approveCount >= int64(total) {
			// 全部通过
			if err := e.approveNode(tx, &node, comment); err != nil {
				tx.Rollback()
				return err
			}
			done, err := e.activateNextOrComplete(tx, &approval, node.NodeOrder)
			if err != nil {
				tx.Rollback()
				return err
			}
			if done {
				publishEvent = "approval.completed"
				publishPayload = buildCompletedEvent(&approval)
			}
		}
		// 未达全部：仅插入 vote，等待

	case "OR_SIGN":
		// 插入投票记录
		vote := models.AdminApprovalVote{
			TenantID: node.TenantID,
			NodeID:   node.ID,
			UserID:   currentUserID,
			Action:   "APPROVE",
			Comment:  comment,
		}
		if err := tx.Create(&vote).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("投票记录创建失败（可能重复投票）: %w", err)
		}
		// OR_SIGN：任意一人通过即可，节点 APPROVED，其余未投票人 SKIPPED
		if err := e.approveNode(tx, &node, comment); err != nil {
			tx.Rollback()
			return err
		}
		// 将其余未投票人对应的待处理状态标记完成（vote 表不插入，节点状态已 APPROVED）
		// 其余节点（同一 approval 下其余 PENDING/WAITING 节点不跳过，只跳过本节点其他审批人，
		// 本节点已 APPROVED，激活下一节点）
		done, err := e.activateNextOrComplete(tx, &approval, node.NodeOrder)
		if err != nil {
			tx.Rollback()
			return err
		}
		if done {
			publishEvent = "approval.completed"
			publishPayload = buildCompletedEvent(&approval)
		}

	default:
		tx.Rollback()
		return fmt.Errorf("不支持的节点类型: %s", node.NodeType)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 事务 Commit 成功后异步发布事件
	if publishEvent != "" {
		go func() {
			defer func() { recover() }() //nolint:errcheck
			e.publishEvent(publishEvent, publishPayload)
		}()
	}
	return nil
}

// Reject 驳回
func (e *ApprovalService) Reject(approvalID int64, currentUserID int64, reason string) error {
	db := e.Orm

	// 1. 加载审批实例
	var approval models.AdminApproval
	if err := db.First(&approval, approvalID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("审批实例不存在: id=%d", approvalID)
		}
		return fmt.Errorf("查询审批实例失败: %w", err)
	}
	if approval.Status != "PENDING" {
		return fmt.Errorf("审批实例状态为 %s，无法驳回", approval.Status)
	}

	// 2. 找到当前 PENDING 节点
	var node models.AdminApprovalNode
	if err := db.Where("approval_id = ? AND status = 'PENDING'", approvalID).
		Order("node_order ASC").First(&node).Error; err != nil {
		return fmt.Errorf("未找到待审批节点: %w", err)
	}

	// 3. 验证当前用户在 assignee_user_ids 中
	if err := e.checkAssignee(node.AssigneeUserIDs, currentUserID); err != nil {
		return err
	}

	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	switch node.NodeType {
	case "SINGLE":
		// 直接驳回
		if err := e.rejectNodeAndInstance(tx, &node, &approval, reason); err != nil {
			tx.Rollback()
			return err
		}

	case "AND_SIGN":
		// 插入 REJECT 投票
		vote := models.AdminApprovalVote{
			TenantID: node.TenantID,
			NodeID:   node.ID,
			UserID:   currentUserID,
			Action:   "REJECT",
			Comment:  reason,
		}
		if err := tx.Create(&vote).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("投票记录创建失败（可能重复投票）: %w", err)
		}
		// AND_SIGN 任一驳回立即终止
		if err := e.rejectNodeAndInstance(tx, &node, &approval, reason); err != nil {
			tx.Rollback()
			return err
		}

	case "OR_SIGN":
		// 插入 REJECT 投票
		vote := models.AdminApprovalVote{
			TenantID: node.TenantID,
			NodeID:   node.ID,
			UserID:   currentUserID,
			Action:   "REJECT",
			Comment:  reason,
		}
		if err := tx.Create(&vote).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("投票记录创建失败（可能重复投票）: %w", err)
		}
		// 统计 REJECT 票数
		var rejectCount int64
		tx.Model(&models.AdminApprovalVote{}).
			Where("node_id = ? AND action = 'REJECT'", node.ID).
			Count(&rejectCount)
		total, err := e.countAssignees(node.AssigneeUserIDs)
		if err != nil {
			tx.Rollback()
			return err
		}
		if rejectCount >= int64(total) {
			// 全部驳回 → 节点+实例 REJECTED
			if err := e.rejectNodeAndInstance(tx, &node, &approval, reason); err != nil {
				tx.Rollback()
				return err
			}
		}
		// 未全部驳回：仅插入 vote，等待

	default:
		tx.Rollback()
		return fmt.Errorf("不支持的节点类型: %s", node.NodeType)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 事务 Commit 成功后异步发布事件
	go func() {
		defer func() { recover() }() //nolint:errcheck
		e.publishEvent("approval.rejected", buildRejectedEvent(&approval))
	}()

	return nil
}

// Cancel 撤销（发起人撤回或管理员强制终止）
func (e *ApprovalService) Cancel(approvalID int64, currentUserID int64, cancelReason string, isSuperAdmin bool) error {
	db := e.Orm

	// 1. 加载审批实例
	var approval models.AdminApproval
	if err := db.First(&approval, approvalID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("审批实例不存在: id=%d", approvalID)
		}
		return fmt.Errorf("查询审批实例失败: %w", err)
	}

	// 2. 只有 PENDING 实例可以撤销
	if approval.Status != "PENDING" {
		return fmt.Errorf("审批实例状态为 %s，无法撤销", approval.Status)
	}

	// 3. 权限校验
	if currentUserID == approval.ApplicantID {
		// 发起人撤回：无需 cancel_reason，直接允许
	} else if isSuperAdmin {
		// SUPER_ADMIN 强制终止他人的审批：cancel_reason 必填
		if cancelReason == "" {
			return errors.New("管理员强制终止必须填写撤销原因")
		}
	} else {
		// 既不是发起人也不是管理员
		return fmt.Errorf("无权撤销此审批，只有发起人可以撤回")
	}

	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 4. 实例置为 CANCELLED，记录 cancel_by 和 cancel_reason
	cancelBy := currentUserID
	if err := tx.Model(&approval).Updates(map[string]interface{}{
		"status":        "CANCELLED",
		"cancel_by":     cancelBy,
		"cancel_reason": cancelReason,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新审批实例状态失败: %w", err)
	}
	// 注意：已审批节点保留原状态，不回滚，只将 PENDING/WAITING 节点置为 SKIPPED
	if err := tx.Model(&models.AdminApprovalNode{}).
		Where("approval_id = ? AND status IN ('PENDING','WAITING')", approvalID).
		Update("status", "SKIPPED").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新节点状态失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 事务 Commit 成功后异步发布事件
	go func() {
		defer func() { recover() }() //nolint:errcheck
		e.publishEvent("approval.cancelled", buildCancelledEvent(&approval))
	}()

	return nil
}

// ─────────────────────────────────────────────
//  内部辅助方法
// ─────────────────────────────────────────────

// checkAssignee 验证 currentUserID 是否在 assignee_user_ids JSON 字符串数组中
// assignee_user_ids 存储格式：["123","456"]
func (e *ApprovalService) checkAssignee(assigneeUserIDsJSON []byte, currentUserID int64) error {
	var ids []string
	if err := json.Unmarshal(assigneeUserIDsJSON, &ids); err != nil {
		return fmt.Errorf("解析 assignee_user_ids 失败: %w", err)
	}
	userIDStr := fmt.Sprintf("%d", currentUserID)
	for _, id := range ids {
		if id == userIDStr {
			return nil
		}
	}
	return fmt.Errorf("无审批权限（403）: 用户 %d 不在审批人列表中", currentUserID)
}

// countAssignees 解析 assignee_user_ids 返回审批人总数
func (e *ApprovalService) countAssignees(assigneeUserIDsJSON []byte) (int, error) {
	var ids []string
	if err := json.Unmarshal(assigneeUserIDsJSON, &ids); err != nil {
		return 0, fmt.Errorf("解析 assignee_user_ids 失败: %w", err)
	}
	return len(ids), nil
}

// approveNode 将节点置为 APPROVED
func (e *ApprovalService) approveNode(tx *gorm.DB, node *models.AdminApprovalNode, comment string) error {
	return tx.Model(node).Updates(map[string]interface{}{
		"status":          "APPROVED",
		"approve_comment": comment,
	}).Error
}

// activateNextOrComplete 激活下一节点；无下一节点则将实例置为 APPROVED。
// 返回 done=true 表示流程已完成（实例变为 APPROVED）。
func (e *ApprovalService) activateNextOrComplete(
	tx *gorm.DB,
	approval *models.AdminApproval,
	currentNodeOrder int,
) (done bool, err error) {
	var nextNode models.AdminApprovalNode
	err = tx.Where("approval_id = ? AND node_order > ? AND status = 'WAITING'", approval.ID, currentNodeOrder).
		Order("node_order ASC").First(&nextNode).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 无下一节点，实例 APPROVED
		if err = tx.Model(approval).Update("status", "APPROVED").Error; err != nil {
			return false, fmt.Errorf("更新审批实例为 APPROVED 失败: %w", err)
		}
		approval.Status = "APPROVED"
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("查询下一节点失败: %w", err)
	}

	// 激活下一节点：WAITING → PENDING，计算 timeout_at，notified_at=NULL
	updates := map[string]interface{}{
		"status":      "PENDING",
		"notified_at": nil,
	}
	if nextNode.TimeoutHours > 0 {
		t := time.Now().Add(time.Duration(nextNode.TimeoutHours) * time.Hour)
		updates["timeout_at"] = t
	}
	if err = tx.Model(&nextNode).Updates(updates).Error; err != nil {
		return false, fmt.Errorf("激活下一节点失败: %w", err)
	}
	return false, nil
}

// rejectNodeAndInstance 驳回节点、跳过其余 PENDING/WAITING 节点、将实例置为 REJECTED
func (e *ApprovalService) rejectNodeAndInstance(
	tx *gorm.DB,
	node *models.AdminApprovalNode,
	approval *models.AdminApproval,
	reason string,
) error {
	// 节点 REJECTED
	if err := tx.Model(node).Updates(map[string]interface{}{
		"status":        "REJECTED",
		"reject_reason": reason,
	}).Error; err != nil {
		return fmt.Errorf("更新节点状态为 REJECTED 失败: %w", err)
	}
	// 其余 PENDING/WAITING 节点 SKIPPED
	if err := tx.Model(&models.AdminApprovalNode{}).
		Where("approval_id = ? AND status IN ('PENDING','WAITING') AND id != ?", approval.ID, node.ID).
		Update("status", "SKIPPED").Error; err != nil {
		return fmt.Errorf("更新其余节点为 SKIPPED 失败: %w", err)
	}
	// 实例 REJECTED
	if err := tx.Model(approval).Update("status", "REJECTED").Error; err != nil {
		return fmt.Errorf("更新审批实例为 REJECTED 失败: %w", err)
	}
	approval.Status = "REJECTED"
	return nil
}

// publishEvent 通过全局 DefaultBus 发布事件（调用方已在 goroutine 中）
func (e *ApprovalService) publishEvent(eventName string, payload interface{}) {
	event.DefaultBus.Publish(eventName, payload)
}

// ─────────────────────────────────────────────
//  事件 payload 构造辅助
// ─────────────────────────────────────────────

func buildCompletedEvent(a *models.AdminApproval) event.ApprovalCompletedEvent {
	return event.ApprovalCompletedEvent{
		ApprovalID: a.ID,
		BizType:    a.BizType,
		BizID:      a.BizID,
		TenantID:   a.TenantID,
	}
}

func buildRejectedEvent(a *models.AdminApproval) event.ApprovalRejectedEvent {
	return event.ApprovalRejectedEvent{
		ApprovalID: a.ID,
		BizType:    a.BizType,
		BizID:      a.BizID,
		TenantID:   a.TenantID,
	}
}

func buildCancelledEvent(a *models.AdminApproval) event.ApprovalCancelledEvent {
	return event.ApprovalCancelledEvent{
		ApprovalID: a.ID,
		BizType:    a.BizType,
		BizID:      a.BizID,
		TenantID:   a.TenantID,
	}
}
