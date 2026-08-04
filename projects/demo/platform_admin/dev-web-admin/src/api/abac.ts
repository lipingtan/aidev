/**
 * ABAC 策略引擎 API
 * 后端路由：/api/v1/admin/abac/
 */
import request from '@/utils/request'

// ---- 条件表达式树类型 ----

/** 条件值（左值或右值） */
export interface CondValue {
  source: 'resource' | 'subject' | 'const'
  attr?: string       // source=resource/subject 时的逻辑属性名
  value?: unknown     // source=const 时的字面量
}

/** 条件节点（递归树） */
export interface CondNode {
  type: 'expr' | 'group'
  // expr 专用
  left?: CondValue
  op?: string
  right?: CondValue
  // group 专用
  operator?: 'AND' | 'OR'
  children?: CondNode[]
}

// ---- 策略类型 ----

export interface AbacRowPolicyItem {
  id: string
  policy_id: string
  action: string           // read|create|update|delete
  condition_expr: CondNode | null
  created_at: string
}

export interface AbacColPolicyItem {
  id: string
  policy_id: string
  field_name: string
  effect: string           // SHOW|HIDE|MASK
  mask_type?: string       // phone|email|id_card|custom
  mask_pattern?: string
  created_at: string
}

/** 策略主表 */
export interface AbacPolicyItem {
  id: string
  tenant_id: string
  name: string
  resource_type: string
  subject_type: string     // ROLE|PERMISSION_SET|USER|DEPT
  subject_id: string
  subject_display_name?: string
  effect: string           // ALLOW|DENY
  priority: number
  status: number
  version: number
  description?: string
  created_at: string
  updated_at: string
  row_policies?: AbacRowPolicyItem[]
  col_policies?: AbacColPolicyItem[]
}

// ---- 请求参数 ----

export interface ListAbacPoliciesParams {
  resource_type?: string
  subject_type?: string
  page?: number
  page_size?: number
}

export interface RowPolicyInput {
  action: string
  condition_expr?: CondNode | null
}

export interface ColPolicyInput {
  field_name: string
  effect: string
  mask_type?: string
  mask_pattern?: string
}

export interface CreateAbacPolicyParams {
  name: string
  resource_type: string
  subject_type: string
  subject_id: string
  effect: string
  priority?: number
  description?: string
  row_policies?: RowPolicyInput[]
  col_policies?: ColPolicyInput[]
}

export interface UpdateAbacPolicyParams {
  name?: string
  effect?: string
  priority?: number
  status?: number
  description?: string
  version: number           // 乐观锁，必填
  row_policies?: RowPolicyInput[]
  col_policies?: ColPolicyInput[]
}

// ---- 资源属性 ----

export interface ResourceAttrVO {
  attr_name: string
  display: string
  data_type: string
}

export interface ResourceVO {
  type: string
  display_name: string
  attributes: ResourceAttrVO[]
}

export interface SubjectAttrVO {
  attr_name: string
  display: string
  data_type: string
}

export interface ListResourcesResponse {
  resources: ResourceVO[]
  subject_attrs: SubjectAttrVO[]
}

// ---- 评估 API ----

export interface EvaluateRequest {
  resource_type: string
  action: string
  subject: Record<string, unknown>
}

export interface EvaluateResponse {
  allowed: boolean
  row_condition?: CondNode
  col_effects?: Record<string, { effect: string; mask_type?: string; mask_pattern?: string }>
}

// ---- API 函数 ----

/** 分页查询策略列表 */
export function listAbacPolicies(params: ListAbacPoliciesParams) {
  return request.get<any, { list: AbacPolicyItem[]; total: number }>(
    '/api/v1/admin/abac/policies',
    { params }
  )
}

/** 查询策略详情 */
export function getAbacPolicy(id: string) {
  return request.get<any, AbacPolicyItem>(`/api/v1/admin/abac/policies/${id}`)
}

/** 创建策略 */
export function createAbacPolicy(data: CreateAbacPolicyParams) {
  return request.post<any, AbacPolicyItem>('/api/v1/admin/abac/policies', data)
}

/** 更新策略 */
export function updateAbacPolicy(id: string, data: UpdateAbacPolicyParams) {
  return request.put(`/api/v1/admin/abac/policies/${id}`, data)
}

/** 删除策略 */
export function deleteAbacPolicy(id: string) {
  return request.delete(`/api/v1/admin/abac/policies/${id}`)
}

/** 查询已注册资源及主体属性 */
export function listAbacResources() {
  return request.get<any, ListResourcesResponse>('/api/v1/admin/abac/resources')
}

/** 策略评估 */
export function evaluateAbacPolicy(data: EvaluateRequest) {
  return request.post<any, EvaluateResponse>('/api/v1/admin/abac/evaluate', data)
}
