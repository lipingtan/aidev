/**
 * 字典管理 API（对接 go-admin）
 * 后端路由：/api/v1/dict/type（字典类型 CRUD）
 *          /api/v1/dict/data（字典数据 CRUD）
 *          /api/v1/dict/type-option-select（字典类型全量）
 *          /api/v1/dict-data/option-select（字典数据全量）
 */
import request from '@/utils/request'
import { adaptPageResponse, type PageQuery, type PageResult } from './helpers'

/** 字典类型 */
export interface DictTypeItem {
  dictId: number
  dictName: string
  dictType: string
  status: number
  remark: string
  createTime: string
}

export interface DictTypeCreateParams {
  dictName: string
  dictType: string
  status: number
  remark?: string
}

/** 字典数据 */
export interface DictDataItem {
  dictCode: number
  dictSort: number
  dictLabel: string
  dictValue: string
  dictType: string
  status: number
  remark: string
  createTime: string
}

export interface DictDataCreateParams {
  dictSort: number
  dictLabel: string
  dictValue: string
  dictType: string
  status: number
  remark?: string
}

// ===== 字典类型 =====

/** 获取字典类型分页列表 */
export async function listDictTypes(params: PageQuery): Promise<PageResult<DictTypeItem>> {
  const res = await request.get('/dict/type', { params })
  return adaptPageResponse<DictTypeItem>(res)
}

/** 获取字典类型详情 */
export function getDictType(id: number) {
  return request.get(`/dict/type/${id}`)
}

/** 获取字典类型全量列表（下拉选择用） */
export function listAllDictTypes() {
  return request.get('/dict/type-option-select')
}

/** 创建字典类型 */
export function createDictType(data: DictTypeCreateParams) {
  return request.post('/dict/type', data)
}

/** 更新字典类型 */
export function updateDictType(id: number, data: Partial<DictTypeCreateParams>) {
  return request.put(`/dict/type/${id}`, data)
}

/** 删除字典类型 */
export function deleteDictType(id: number) {
  return request.delete('/dict/type', { data: { ids: [id] } })
}

// ===== 字典数据 =====

/** 获取字典数据分页列表 */
export async function listDictData(params: PageQuery & { dictType: string }): Promise<PageResult<DictDataItem>> {
  const res = await request.get('/dict/data', { params })
  return adaptPageResponse<DictDataItem>(res)
}

/** 获取字典数据详情 */
export function getDictData(dictCode: number) {
  return request.get(`/dict/data/${dictCode}`)
}

/** 获取字典数据全量列表（按类型） */
export function listAllDictData() {
  return request.get('/dict-data/option-select')
}

/** 创建字典数据 */
export function createDictData(data: DictDataCreateParams) {
  return request.post('/dict/data', data)
}

/** 更新字典数据 */
export function updateDictData(dictCode: number, data: Partial<DictDataCreateParams>) {
  return request.put(`/dict/data/${dictCode}`, data)
}

/** 删除字典数据 */
export function deleteDictData(dictCode: number) {
  return request.delete('/dict/data', { data: { ids: [dictCode] } })
}
