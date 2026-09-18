/** 内容(景点/美食/路线)公共查询接口封装 */
import { get } from './request'

export interface PageResult<T = unknown> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface QueryParams {
  district?: string
  tag?: string
  keyword?: string
  page?: number
  page_size?: number
}

/** 景点 */
export interface ScenicItem {
  id: number
  name_zh: string
  name_en?: string
  district: string
  tags?: string
  score?: number
  lat?: number
  lng?: number
  desc?: string
  images?: string
}

export function listScenics(params?: QueryParams): Promise<PageResult<ScenicItem>> {
  return get<PageResult<ScenicItem>>('/scenic', { params })
}
export function getScenic(id: number): Promise<ScenicItem> {
  return get<ScenicItem>(`/scenic/${id}`)
}

/** 美食 */
export interface FoodItem {
  id: number
  name_zh: string
  name_en?: string
  district: string
  tags?: string
  desc?: string
  images?: string
}

export function listFoods(params?: QueryParams): Promise<PageResult<FoodItem>> {
  return get<PageResult<FoodItem>>('/food', { params })
}
export function getFood(id: number): Promise<FoodItem> {
  return get<FoodItem>(`/food/${id}`)
}

/** 路线 */
export interface RouteItem {
  id: number
  title_zh: string
  title_en?: string
  theme?: string
  stops?: string
  duration?: string
  difficulty?: string
}

export function listRoutes(params?: QueryParams): Promise<PageResult<RouteItem>> {
  return get<PageResult<RouteItem>>('/routes', { params })
}
export function getRoute(id: number): Promise<RouteItem> {
  return get<RouteItem>(`/routes/${id}`)
}

/** 文旅热点(后端定时抓取官方文旅新闻源) */
export interface HotspotItem {
  id: number
  title_zh: string
  title_en?: string
  title_ja?: string
  summary_zh: string
  summary_en?: string
  summary_ja?: string
  source_zh: string
  source_en?: string
  url: string
  hot: boolean
  published_at: string
}

export function listHotspots(page = 1, pageSize = 6): Promise<PageResult<HotspotItem>> {
  return get<PageResult<HotspotItem>>('/news/hotspots', { params: { page, page_size: pageSize } })
}