/**
 * 用户 API（H5 端）
 * 用户个人信息、注册
 */
import request from '@/utils/request'

// ==================== 类型定义 ====================

/** 房产绑定项 */
export interface PropertyBindingItem {
  id: number
  communityName: string
  buildingName: string
  unitName: string
  roomNumber: string
}

/** 用户个人信息 */
export interface H5ProfileVO {
  id: number
  name: string
  phone: string
  phoneMasked: string
  idCard?: string
  idCardMasked?: string
  gender: number
  email?: string
  avatar?: string
  status: string
  propertyBindings?: PropertyBindingItem[]
}

/** 注册参数 */
export interface H5RegisterDTO {
  phone: string
  password: string
  ownerName: string
  idCard?: string
  captcha: string
}

// ==================== API 方法 ====================

/** 获取当前用户个人信息 */
export function getProfile(): Promise<H5ProfileVO> {
  return request.get('/user/profile')
}

/** 用户注册 */
export function register(data: H5RegisterDTO): Promise<void> {
  return request.post('/user/register', data)
}
