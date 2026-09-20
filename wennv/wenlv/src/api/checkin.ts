/** 打卡接口封装（数据永久落库 MySQL check_ins 表，护照集章凭证） */
import { get, post, del } from './request'

export interface CheckInItem {
  id: number
  scenic_id: number
  photo_url?: string
  /** 打卡照片 URL 列表（后端由 JSON 列解析为数组） */
  photos: string[]
  comment?: string
  visited_at: string
  created_at: string
}

/** 打卡（photos 必传至少 1 张，均为 OSS 直传后的公开 URL） */
export function createCheckIn(payload: {
  scenic_id: number
  photos: string[]
  comment?: string
  visited_at?: string
}) {
  return post<void>('/check-ins', payload)
}

/** 我的打卡列表 */
export function listCheckIns(): Promise<CheckInItem[]> {
  return get<CheckInItem[]>('/check-ins')
}

/** 删除打卡 */
export function removeCheckIn(id: number) {
  return del<void>(`/check-ins/${id}`)
}
