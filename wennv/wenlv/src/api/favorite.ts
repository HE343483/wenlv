/** 收藏接口封装 */
import { get, post, del } from './request'

export interface FavoriteItem {
  id: number
  user_id: number
  target_type: string
  target_id: number
  created_at: string
}

/** 我的收藏列表；target_type 可选 scenic|food|route */
export function listFavorites(target_type?: string): Promise<FavoriteItem[]> {
  return get<FavoriteItem[]>('/favorites', { params: target_type ? { target_type } : {} })
}

/** 添加收藏 */
export function addFavorite(target_type: string, target_id: number) {
  return post<void>('/favorites', { target_type, target_id })
}

/** 取消收藏 */
export function removeFavorite(target_type: string, target_id: number) {
  return del<void>(`/favorites/${target_type}/${target_id}`)
}