/** 每日蜀签接口封装(公开接口,无需登录) */
import { get } from './request'

/** 蜀签条目:诗句引用/四川方言/蜀文化冷知识,三语 */
export interface CultureDailyItem {
  id: number
  content_zh: string
  /** 英文内容(LLM 生成,可能为空,展示回落中文) */
  content_en?: string
  /** 日文内容(LLM 生成,可能为空,展示回落中文) */
  content_ja?: string
  /** 分类:诗句 / 方言 / 冷知识 */
  category: string
  /** 关联景点 ID(可空,有值时展示"去看看"跳景点详情) */
  related_spot_id?: number | null
}

/**
 * 取今日蜀签:后端返回第 (dayOfYear + offset) % 总数 条。
 * 无数据时 data 为 null(前端整体隐藏卡片)。
 */
export function getCultureDaily(offset = 0): Promise<CultureDailyItem | null> {
  return get<CultureDailyItem | null>('/culture/daily', { params: { offset } })
}

/** 应季美食条目:节气蜀俗条幅小卡所需精简字段 */
export interface SolarFoodItem {
  id: number
  name_zh: string
  /** 英/日文名称(LLM 生成,可能为空,展示回落中文) */
  name_en?: string
  name_ja?: string
  desc?: string
  /** 图片 URL 列表(逗号分隔,取第一张展示) */
  images?: string
  district?: string
}

/**
 * 按节气查应季美食:term 传中文节气名(如"冬至")。
 * 节气名非法/无匹配时后端返回空数组(前端据此隐藏条幅)。
 */
export function getSolarFoods(term: string): Promise<SolarFoodItem[]> {
  return get<SolarFoodItem[]>('/culture/solar-food', { params: { term } })
}
