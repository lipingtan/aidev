package dto

import (
	"go-admin/common/dto"
)

// ApprovalFlowGetPageReq 审批流定义分页查询请求
type ApprovalFlowGetPageReq struct {
	dto.Pagination `search:"-"`
	FlowCode       string `form:"flow_code" search:"type:contains;column:flow_code;table:admin_approval_flow" comment:"流程标识"`
	FlowName       string `form:"flow_name" search:"type:contains;column:flow_name;table:admin_approval_flow" comment:"流程名称"`
}

func (m *ApprovalFlowGetPageReq) GetNeedSearch() interface{} {
	return *m
}

// ApprovalFlowGetReq 审批流定义详情请求
type ApprovalFlowGetReq struct {
	Id string `uri:"id"`
}

func (s *ApprovalFlowGetReq) GetId() interface{} {
	return s.Id
}

// FlowNodeConfig 流程节点配置（flow_config JSON 中的单个节点）
type FlowNodeConfig struct {
	NodeOrder     int      `json:"node_order"`
	NodeType      string   `json:"node_type"`      // SINGLE|AND_SIGN|OR_SIGN
	AssigneeType  string   `json:"assignee_type"`  // USER|ROLE|DEPT_HEAD
	AssigneeIDs   []string `json:"assignee_ids"`   // 审批人/角色 ID 列表
	TimeoutHours  int      `json:"timeout_hours"`  // 0=不超时
	TimeoutAction string   `json:"timeout_action"` // AUTO_APPROVE|AUTO_REJECT|ESCALATE
	EscalateTo    string   `json:"escalate_to"`    // 超时升级目标
}

// ApprovalFlowInsertReq 创建审批流请求
type ApprovalFlowInsertReq struct {
	FlowCode    string           `json:"flow_code" binding:"required"`
	FlowName    string           `json:"flow_name" binding:"required"`
	FlowConfig  []FlowNodeConfig `json:"flow_config" binding:"required"`
	Description string           `json:"description"`
	CreateBy    int              `json:"-"`
}

// ApprovalFlowUpdateReq 更新审批流请求
type ApprovalFlowUpdateReq struct {
	Id          string           `uri:"id" binding:"required"`
	FlowName    string           `json:"flow_name"`
	FlowConfig  []FlowNodeConfig `json:"flow_config"`
	Description string           `json:"description"`
	UpdateBy    int              `json:"-"`
}

func (s *ApprovalFlowUpdateReq) GetId() interface{} {
	return s.Id
}

// ApprovalFlowDeleteReq 删除审批流请求
type ApprovalFlowDeleteReq struct {
	Id string `uri:"id"`
}

func (s *ApprovalFlowDeleteReq) GetId() interface{} {
	return s.Id
}
