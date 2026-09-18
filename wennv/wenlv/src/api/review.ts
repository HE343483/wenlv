/** 评价接口封装 */
import { get, post, del } from './request'
import type { PageResult, QueryParams } from './content'

export interface ReviewItem {
  id: number
  user_id: number
  target_type: string
  target_id: number
  content: string
  rating: number
  created_at: string
}

/** 目标评价列表 */
export function listReviews(params: {
  target_type: string
  target_id: number
  page?: number
  page_size?: number
}): Promise<PageResult<ReviewItem>> {
  return get<PageResult<ReviewItem>>('/reviews', { params })
}

/** 目标平均评分摘要 */
export function reviewSummary(
  target_type: string,
  target_id: number,
): Promise<{ target_type: string; target_id: number; avg_rating: number }> {
  return get<{ target_type: string; target_id: number; avg_rating: number }>(
    `/reviews/summary/${target_type}/${target_id}`,
  )
}

/** 发表评价 */
export function createReview(payload: {
  target_type: string
  target_id: number
  content: string
  rating: number
}) {
  return post<void>('/reviews', payload)
}

/** 我的评价 */
export function myReviews(): Promise<ReviewItem[]> {
  return get<ReviewItem[]>('/reviews/mine')
}

/** 删除评价 */
export function removeReview(id: number) {
  return del<void>(`/reviews/${id}`)
}