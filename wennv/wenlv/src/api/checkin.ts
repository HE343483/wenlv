/** 打卡接口封装 */
import { get, post, del } from './request'

export interface CheckInItem {
  id: number
  scenic_id: number
  photo_url?: string
  comment?: string
  visited_at: string
  created_at: string
}

/** 打卡 */
export function createCheckIn(payload: {
  scenic_id: number
  photo_url?: string
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