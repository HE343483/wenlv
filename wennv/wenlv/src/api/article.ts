/** 游记接口封装 */
import { get, post, put, del } from './request'
import type { PageResult } from './content'

export interface ArticleItem {
  id: number
  user_id: number
  title: string
  content?: string
  cover_url?: string
  created_at: string
  updated_at: string
}

/** 游记列表 */
export function listArticles(params?: { page?: number; page_size?: number }): Promise<PageResult<ArticleItem>> {
  return get<PageResult<ArticleItem>>('/articles', { params })
}

/** 游记详情 */
export function getArticle(id: number): Promise<ArticleItem> {
  return get<ArticleItem>(`/articles/${id}`)
}

/** 发布游记 */
export function createArticle(payload: { title: string; content: string; cover_url?: string }) {
  return post<void>('/articles', payload)
}

/** 更新游记 */
export function updateArticle(id: number, payload: { title: string; content: string; cover_url?: string }) {
  return put<void>(`/articles/${id}`, payload)
}

/** 删除游记 */
export function removeArticle(id: number) {
  return del<void>(`/articles/${id}`)
}