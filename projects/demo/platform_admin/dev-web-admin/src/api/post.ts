/**
 * 岗位管理 API（对接 go-admin）
 * 后端路由：/api/v1/post（CRUD）
 */
import request from '@/utils/request'
import { adaptPageResponse, type PageQuery, type PageResult } from './helpers'

export interface PostItem {
  postId: number
  postCode: string
  postName: string
  sort: number
  status: number
  createTime: string
}

export interface PostCreateParams {
  postCode: string
  postName: string
  sort: number
  status: number
}

/** 获取岗位分页列表 */
export async function listPosts(params: PageQuery): Promise<PageResult<PostItem>> {
  const res = await request.get('/post', { params })
  return adaptPageResponse<PostItem>(res)
}

/** 获取岗位详情 */
export function getPost(id: number) {
  return request.get<PostItem>(`/post/${id}`)
}

/** 创建岗位 */
export function createPost(data: PostCreateParams) {
  return request.post('/post', data)
}

/** 更新岗位 */
export function updatePost(id: number, data: Partial<PostCreateParams>) {
  return request.put(`/post/${id}`, data)
}

/** 删除岗位 */
export function deletePost(id: number) {
  return request.delete(`/post/${id}`)
}
