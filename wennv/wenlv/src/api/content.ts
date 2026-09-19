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
  // ===== 详情页扩展字段(高德 POI 采集 + LLM 参考值) =====
  /** 详细地址(高德) */
  address?: string
  /** 咨询电话(高德) */
  tel?: string
  /** 开放时间(高德,部分为 LLM 参考值) */
  open_hours?: string
  /** 门票价格(LLM 参考值) */
  ticket_price?: string
  /** 建议游玩时长(LLM 参考值) */
  recommend_hours?: string
  /** 年接待游客(LLM 参考值) */
  yearly_visitors?: string
  /** 相册图片 URL 列表(逗号分隔) */
  gallery_images?: string
  /** 图文详情段落 JSON 串 */
  detail_sections?: string
  /** 参考值字段名(逗号分隔,前端加"参考值"标注) */
  estimated_fields?: string
  /** 数据来源(amap / amap+llm / wiki / llm / manual) */
  data_source?: string
  /** 数据采集时间 */
  data_updated_at?: string
}

export function listScenics(params?: QueryParams): Promise<PageResult<ScenicItem>> {
  return get<PageResult<ScenicItem>>('/scenic', { params })
}
export function getScenic(id: number): Promise<ScenicItem> {
  return get<ScenicItem>(`/scenic/${id}`)
}

/** 图文详情的一段(标题 + 正文 + 配图) */
export interface ScenicDetailSection {
  title: string
  text: string
  image?: string
}

export interface ScenicAroundItem {
  id: string
  name: string
  type?: string
  address?: string
  distance?: number
  lat?: number
  lng?: number
}

export interface ScenicTransitStop {
  name: string
  type?: string
  distance?: number
}

/** 周边推荐(高德实时查询,后端 Redis 缓存 24h) */
export function getScenicAround(id: number, limit = 6): Promise<ScenicAroundItem[]> {
  return get<ScenicAroundItem[]>(`/scenic/${id}/around`, { params: { limit } })
}

/** 邻近交通站点(地铁站/公交站) */
export function getScenicTransport(id: number): Promise<ScenicTransitStop[]> {
  return get<ScenicTransitStop[]>(`/scenic/${id}/transport`)
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