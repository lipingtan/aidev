// Package jobs 超时扫描 Cron Job — 审批节点超时处理
// 每 5 分钟扫描一次 admin_approval_node，处理 AUTO_APPROVE / AUTO_REJECT / ESCALATE
package jobs

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/go-admin-team/go-admin-core/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go-admin/app/admin/models"
	"go-admin/common/event"
)

// RunApprovalTimeoutScan 超时扫描入口，每批 LIMIT 100，SELECT FOR UPDATE 防多 pod 重复执行
func RunApprovalTimeoutScan(db *gorm.DB) {
	now := time.Now()
	log.Infof("[ApprovalTimeout] 开始扫描，时间: %s", now.Format("2006-01-02 15:04:05"))

	// SELECT ... FOR UPDATE WHERE status='PENDING' AND timeout_at IS NOT NULL AND timeout_at <= NOW() LIMIT 100
	var nodes []models.AdminApprovalNode
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("status = 'PENDING' AND timeout_at IS NOT NULL AND timeout_at <= ?", now).
		Order("timeout_at ASC").
		Limit(100).
		Find(&nodes).Error
	if err != nil {
		log.Errorf("[ApprovalTimeout] 查询超时节点失败: %v", err)
		return
	}

	if len(nodes) == 0 {
		log.Infof("[ApprovalTimeout] 无超时节点")
		return
	}

	log.Infof("[ApprovalTimeout] 发现 %d 条超时节点，开始处理", len(nodes))

	for i := range nodes {
		node := &nodes[i]
		if err := handleTimeoutNode(db, node); err != nil {
			log.Errorf("[ApprovalTimeout] 节点 %d 处理失败: %v", node.ID, err)
		}
	}
}

// handleTimeoutNode 根据 timeout_action 分派处理
func handleTimeoutNode(db *gorm.DB, node *models.AdminApprovalNode) error {
	log.Infof("[ApprovalTimeout] 处理节点 id=%d approvalID=%d action=%s", node.ID, node.ApprovalID, node.TimeoutAction)

	switch node.TimeoutAction {
	case "AUTO_APPROVE":
		return timeoutAutoApprove(db, node, "")
	case "AUTO_REJECT":
		return timeoutAutoReject(db, node)
	case "ESCALATE":
		return timeoutEscalate(db, node)
	default:
		// timeout_action 未配置或不识别：降级 AUTO_APPROVE
		log.Warnf("[ApprovalTimeout] 节点 %d timeout_action=%q 不识别，降级 AUTO_APPROVE", node.ID, node.TimeoutAction)
		return timeoutAutoApprove(db, node, "timeout_action 未配置，降级自动通过")
	}
}

// timeoutAutoApprove 超时自动通过：直接更新 DB 状态，不走 ApprovalService（避免用户校验）
func timeoutAutoApprove(db *gorm.DB, node *models.AdminApprovalNode, extraNote string) error {
	// 加载审批实例
	var approval models.AdminApproval
	if err := db.First(&approval, node.ApprovalID).Error; err != nil {
		return fmt.Errorf("加载审批实例失败: %w", err)
	}
	if approval.Status != "PENDING" {
		log.Warnf("[ApprovalTimeout] 节点 %d 对应实例 %d 状态为 %s，跳过", node.ID, approval.ID, approval.Status)
		return nil
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

	// 节点置为 APPROVED
	note := "超时自动通过"
	if extraNote != "" {
		note = extraNote
	}
	if err := tx.Model(node).Updates(map[string]interface{}{
		"status":          "APPROVED",
		"approve_comment": note,
		"assignee_note":   note,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新节点状态失败: %w", err)
	}

	// 激活下一节点或完成实例
	done, err := activateNextOrCompleteTimeout(tx, &approval, node.NodeOrder)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 事务提交后异步发布事件
	if done {
		go func() {
			defer func() { recover() }() //nolint:errcheck
			event.DefaultBus.Publish("approval.completed", event.ApprovalCompletedEvent{
				ApprovalID: approval.ID,
				BizType:    approval.BizType,
				BizID:      approval.BizID,
				TenantID:   approval.TenantID,
			})
		}()
	}
	return nil
}

// timeoutAutoReject 超时自动驳回
func timeoutAutoReject(db *gorm.DB, node *models.AdminApprovalNode) error {
	var approval models.AdminApproval
	if err := db.First(&approval, node.ApprovalID).Error; err != nil {
		return fmt.Errorf("加载审批实例失败: %w", err)
	}
	if approval.Status != "PENDING" {
		log.Warnf("[ApprovalTimeout] 节点 %d 对应实例 %d 状态为 %s，跳过", node.ID, approval.ID, approval.Status)
		return nil
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

	reason := "超时自动驳回"
	// 节点 REJECTED
	if err := tx.Model(node).Updates(map[string]interface{}{
		"status":        "REJECTED",
		"reject_reason": reason,
		"assignee_note": reason,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新节点状态失败: %w", err)
	}
	// 其余 PENDING/WAITING 节点 SKIPPED
	if err := tx.Model(&models.AdminApprovalNode{}).
		Where("approval_id = ? AND status IN ('PENDING','WAITING') AND id != ?", approval.ID, node.ID).
		Update("status", "SKIPPED").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新其余节点失败: %w", err)
	}
	// 实例 REJECTED
	if err := tx.Model(&approval).Update("status", "REJECTED").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新实例状态失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 事务提交后异步发布事件
	go func() {
		defer func() { recover() }() //nolint:errcheck
		event.DefaultBus.Publish("approval.rejected", event.ApprovalRejectedEvent{
			ApprovalID: approval.ID,
			BizType:    approval.BizType,
			BizID:      approval.BizID,
			TenantID:   approval.TenantID,
		})
	}()
	return nil
}

// timeoutEscalate 超时升级转派
// 有效：原地更新 assignee_user_ids + 重置 timeout_at，状态改回 PENDING
// 无效：降级 AUTO_APPROVE，记录 assignee_note
func timeoutEscalate(db *gorm.DB, node *models.AdminApprovalNode) error {
	if node.EscalateTo == "" {
		// escalate_to 未配置，降级
		note := "escalate_to 未配置，降级自动通过"
		log.Warnf("[ApprovalTimeout] 节点 %d %s", node.ID, note)
		return timeoutAutoApprove(db, node, note)
	}

	// 验证 escalate_to 用户/角色是否有效，并展开为用户 ID 列表
	userIDs, valid := resolveEscalateTo(db, node.EscalateTo, node.TenantID)
	if !valid || len(userIDs) == 0 {
		note := fmt.Sprintf("escalate_to=%s 无效，降级自动通过", node.EscalateTo)
		log.Warnf("[ApprovalTimeout] 节点 %d %s", node.ID, note)
		return timeoutAutoApprove(db, node, note)
	}

	// 原地转派
	assigneeUserIDsJSON, _ := json.Marshal(userIDs)
	var newTimeoutAt *time.Time
	if node.TimeoutHours > 0 {
		t := time.Now().Add(time.Duration(node.TimeoutHours) * time.Hour)
		newTimeoutAt = &t
	}
	note := fmt.Sprintf("超时转派至 escalate_to=%s", node.EscalateTo)

	if err := db.Model(node).Updates(map[string]interface{}{
		"assignee_user_ids": assigneeUserIDsJSON,
		"timeout_at":        newTimeoutAt,
		"status":            "PENDING",
		"assignee_note":     note,
	}).Error; err != nil {
		return fmt.Errorf("原地转派更新失败: %w", err)
	}

	log.Infof("[ApprovalTimeout] 节点 %d 转派成功，新审批人: %v，新 timeout_at: %v", node.ID, userIDs, newTimeoutAt)
	return nil
}

// resolveEscalateTo 验证并展开 escalate_to 为用户 ID 字符串列表
// escalate_to 可以是用户 ID（纯数字）或角色 ID（"ROLE:xxx" 格式）
// 返回 (userIDs, 是否有效)
func resolveEscalateTo(db *gorm.DB, escalateTo string, tenantID int64) ([]string, bool) {
	// 优先尝试按用户 ID 查询
	var userID int
	if _, err := fmt.Sscanf(escalateTo, "%d", &userID); err == nil && userID > 0 {
		var user models.SysUser
		if err := db.Select("user_id").Where("user_id = ? AND status = '0'", userID).First(&user).Error; err == nil {
			return []string{fmt.Sprintf("%d", user.UserId)}, true
		}
		// 用户不存在或已禁用
		return nil, false
	}

	// 尝试按角色解析（格式 "ROLE:角色ID"）
	var roleID int
	if n, _ := fmt.Sscanf(escalateTo, "ROLE:%d", &roleID); n == 1 && roleID > 0 {
		var users []models.SysUser
		if err := db.Select("user_id").
			Where("role_id = ? AND tenant_id = ? AND status = '0'", roleID, int(tenantID)).
			Find(&users).Error; err != nil || len(users) == 0 {
			return nil, false
		}
		ids := make([]string, 0, len(users))
		for _, u := range users {
			ids = append(ids, fmt.Sprintf("%d", u.UserId))
		}
		return ids, true
	}

	return nil, false
}

// activateNextOrCompleteTimeout 激活下一节点或将实例置为 APPROVED（超时场景内部使用）
func activateNextOrCompleteTimeout(tx *gorm.DB, approval *models.AdminApproval, currentNodeOrder int) (done bool, err error) {
	var nextNode models.AdminApprovalNode
	err = tx.Where("approval_id = ? AND node_order > ? AND status = 'WAITING'", approval.ID, currentNodeOrder).
		Order("node_order ASC").First(&nextNode).Error

	if err != nil {
		if err.Error() == "record not found" {
			// 无下一节点，实例 APPROVED
			if err = tx.Model(approval).Update("status", "APPROVED").Error; err != nil {
				return false, fmt.Errorf("更新实例为 APPROVED 失败: %w", err)
			}
			approval.Status = "APPROVED"
			return true, nil
		}
		return false, fmt.Errorf("查询下一节点失败: %w", err)
	}

	// 激活下一节点
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
