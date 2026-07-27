/**
 * 审批流引擎 API
 */
import request from '@/utils/request'

// ==================== 类型定义 ====================

export interface FlowNodeConfig {
  node_order: number
  node_type: 'SINGLE' | 'AND_SIGN' | 'OR_SIGN'
  assignee_type: 'USER' | 'ROLE' | 'DEPT_HEAD'
  assignee_ids: string[]
  timeout_hours?: number
  timeout_action?: 'AUTO_APPROVE' | 'AUTO_REJECT' | 'ESCALATE'
  escalate_to?: string
}

export interface ApprovalFlow {
  id: string
  tenant_id: string
  flow_code: string
  flow_name: string
  flow_config: FlowNodeConfig[]
  description?: string
  created_at: string
  updated_at: string
}

export interface ApprovalInstance {
  id: string
  tenant_id: string
  flow_id: string
  flow_code: string
  flow_name?: string
  biz_type: string
  biz_id: string
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'CANCELLED'
  applicant_id: string
  applicant_name?: string
  cancel_by?: string
  cancel_reason?: string
  created_at: string
}

export interface ApprovalNode {
  id: string
  approval_id: string
  node_order: number
  node_type: string
  assignee_type: string
  assignee_user_ids: string[]
  status: string
  approve_comment?: string
  reject_reason?: string
  assignee_note?: string
}

export interface ApprovalDetail extends ApprovalInstance {
  nodes: ApprovalNode[]
}

export interface PageResult<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
}

// ==================== 审批流定义 ====================

export const getApprovalFlows = (params?: any) =>
  request.get('/api/v1/admin/approval-flows', { params })

export const getApprovalFlow = (id: string) =>
  request.get(`/api/v1/admin/approval-flows/${id}`)

export const createApprovalFlow = (data: any) =>
  request.post('/api/v1/admin/approval-flows', data)

export const updateApprovalFlow = (id: string, data: any) =>
  request.put(`/api/v1/admin/approval-flows/${id}`, data)

export const deleteApprovalFlow = (id: string) =>
  request.delete(`/api/v1/admin/approval-flows/${id}`)

// ==================== 审批实例 ====================

export const getApprovals = (params?: any) =>
  request.get('/api/v1/admin/approvals', { params })

export const getApproval = (id: string) =>
  request.get(`/api/v1/admin/approvals/${id}`)

export const createApproval = (data: any) =>
  request.post('/api/v1/admin/approvals', data)

export const approveApproval = (id: string, data?: { comment?: string }) =>
  request.post(`/api/v1/admin/approvals/${id}/approve`, data)

export const rejectApproval = (id: string, data: { reason: string }) =>
  request.post(`/api/v1/admin/approvals/${id}/reject`, data)

export const cancelApproval = (id: string, data?: { cancel_reason?: string }) =>
  request.post(`/api/v1/admin/approvals/${id}/cancel`, data)
