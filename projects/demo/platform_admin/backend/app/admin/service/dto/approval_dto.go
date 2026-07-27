package dto

// InitApprovalReq 发起审批请求
type InitApprovalReq struct {
	FlowCode    string `json:"flow_code" binding:"required"` // 流程标识
	BizType     string `json:"biz_type" binding:"required"`  // 业务类型
	BizID       string `json:"biz_id" binding:"required"`    // 业务对象ID
	TenantID    int64  `json:"tenant_id"`                    // 租户ID（可从 context 覆盖）
	ApplicantID int64  `json:"applicant_id"`                 // 发起人用户ID（可从 context 覆盖）
}

// ApprovalNodeConfig 流程节点配置（对应 flow_config JSON 中的单节点）
type ApprovalNodeConfig struct {
	NodeOrder     int      `json:"node_order"`
	NodeType      string   `json:"node_type"`      // SINGLE/AND_SIGN/OR_SIGN
	AssigneeType  string   `json:"assignee_type"`  // USER/ROLE/DEPT_HEAD
	AssigneeIDs   []string `json:"assignee_ids"`   // 原始配置的 ID 列表
	TimeoutHours  int      `json:"timeout_hours"`  // 0=不超时
	TimeoutAction string   `json:"timeout_action"` // AUTO_APPROVE/AUTO_REJECT/ESCALATE
	EscalateTo    string   `json:"escalate_to"`    // 超时升级目标
}

// ApproveReq 审批通过请求
type ApproveReq struct {
	ApprovalID int64  `json:"approval_id"`
	Comment    string `json:"comment"` // 可选
}

// RejectReq 驳回请求
type RejectReq struct {
	ApprovalID int64  `json:"approval_id"`
	Reason     string `json:"reason" binding:"required"`
}

// CancelReq 撤销请求
type CancelReq struct {
	ApprovalID   int64  `json:"approval_id"`
	CancelReason string `json:"cancel_reason"`
}
