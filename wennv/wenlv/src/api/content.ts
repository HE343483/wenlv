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
  // ===== 详情页扩展字段(高德门店采集 + LLM 参考值) =====
  /** 评分(高德事实或 LLM 参考值) */
  rating?: string
  /** 风味标签(LLM 参考值) */
  flavor?: string
  /** 辣度(LLM 参考值) */
  spice_level?: string
  /** 人均消费(如 人均 ¥71) */
  avg_price?: string
  /** 招牌推荐(LLM 参考值) */
  signature?: string
  /** 推荐场景(LLM 参考值) */
  recommend_scene?: string
  /** 风味故事段落 JSON 串 */
  story_sections?: string
  /** 相册图片 URL 列表(逗号分隔) */
  gallery_images?: string
  /** 高德门店名(事实) */
  poi_name?: string
  /** 门店地址(事实) */
  address?: string
  /** 门店纬度(事实) */
  lat?: number
  /** 门店经度(事实) */
  lng?: number
  /** 参考值字段名(逗号分隔,前端加"参考值"标注) */
  estimated_fields?: string
  /** 数据来源(amap / amap+llm / llm / wiki) */
  data_source?: string
  /** 数据采集时间 */
  data_updated_at?: string
}

export function listFoods(params?: QueryParams): Promise<PageResult<FoodItem>> {
  return get<PageResult<FoodItem>>('/food', { params })
}
export function getFood(id: number): Promise<FoodItem> {
  return get<FoodItem>(`/food/${id}`)
}

/** 美食名片(美食页六大风味卡片,配图存于数据库) */
export interface FoodCardItem {
  id: number
  /** 卡片标识,与前端 i18n 键 food.card.{key} 对应 */
  card_key: string
  name_zh: string
  name_en?: string
  /** 配图 OSS 地址 */
  image?: string
  sort?: number
}

export function listFoodCards(): Promise<FoodCardItem[]> {
  return get<FoodCardItem[]>('/food-cards')
}

/** 美食大类(川菜/名小吃/夜宵)：详情页介绍"这一类"而非某道菜 */
export interface CategorySection {
  title: string
  text: string
  image?: string
}

export interface FoodCategoryItem {
  id: number
  /** 类别键(cuisine/snacks/nightfood) */
  key: string
  name_zh: string
  name_en?: string
  /** 类别概述(LLM 依据维基素材改写,参考值) */
  intro?: string
  /** 类别图文段落 JSON 串 */
  sections?: string
  /** 类别图集 URL 列表(逗号分隔) */
  gallery_images?: string
  /** 素材来源链接(维基条目) */
  source_url?: string
  /** 参考值字段名(逗号分隔,前端加"参考值"标注) */
  estimated_fields?: string
  /** 数据来源(wiki+llm+oss) */
  data_source?: string
  /** 数据采集时间 */
  data_updated_at?: string
}

export function getFoodCategory(key: string): Promise<FoodCategoryItem> {
  return get<FoodCategoryItem>(`/food-category/${key}`)
}

/** 路线站点(Stops JSON 解析后的结构,爬虫从维基百科采集) */
export interface RouteStop {
  name_zh: string
  name_en?: string
  name_ja?: string
  /** 站点简介(维基百科导言,无词条时为运营文案) */
  desc?: string
  /** 站点配图 OSS 地址 */
  image?: string
}

/** 精选路线(数据由爬虫采集,图片存 OSS) */
export interface RouteItem {
  id: number
  /** 路线标识(classic/panda/food/culture) */
  route_key: string
  title_zh: string
  title_en?: string
  title_ja?: string
  theme?: string
  description?: string
  /** 建议游玩天数 */
  days?: number
  /** 偏好标签(逗号分隔,与 AI 行程兴趣项一致) */
  interests?: string
  /** 封面图 OSS 地址 */
  cover_image?: string
  /** 途经站点 JSON 串 */
  stops?: string
  sort?: number
}

/** 解析站点 JSON 串(容错:非法 JSON 返回空数组) */
export function parseRouteStops(stops?: string): RouteStop[] {
  if (!stops) return []
  try {
    const arr = JSON.parse(stops)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

export function listRoutes(params?: QueryParams): Promise<PageResult<RouteItem>> {
  return get<PageResult<RouteItem>>('/routes', { params })
}
export function getRoute(id: number): Promise<RouteItem> {
  return get<PageResult<RouteItem>>('/routes').then(r => r.items.find(it => it.id === id)!)
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