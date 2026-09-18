/** 图墙照片接口封装 */
import { get, post } from './request'
import type { PageResult } from './content'

export interface PhotoItem {
  id: number
  user_id: number
  target_type: string
  target_id: number
  url: string
  created_at: string
}

/** 目标的照片 */
export function listPhotosByTarget(target_type: string, target_id: number): Promise<PhotoItem[]> {
  return get<PhotoItem[]>('/photos', { params: { target_type, target_id } })
}

/** 全量照片墙 */
export function photoWall(params?: { page?: number; page_size?: number }): Promise<PageResult<PhotoItem>> {
  return get<PageResult<PhotoItem>>('/photos/wall', { params })
}

/** 上传照片记录 */
export function addPhoto(payload: { target_type: string; target_id: number; url: string }) {
  return post<void>('/photos', payload)
}