/** 数智文旅 + 国际传播：巴蜀文化出海 — 类型定义 */

/** 成都行政区划 */
export interface District {
  id: string
  nameZh: string
  nameEn: string
  /** 区域介绍 */
  descriptionZh: string
  descriptionEn: string
  /** 代表色 (用于卡片/标识) */
  color: string
}

/** 景点 */
export interface ScenicSpot {
  id: string
  districtId: string
  nameZh: string
  nameEn: string
  /** 简短描述 */
  shortDescZh: string
  shortDescEn: string
  /** 详细描述 */
  descriptionZh: string
  descriptionEn: string
  /** 分类标签 */
  tags: string[]
  /** 图片占位URL */
  imageUrl: string
  /** 评级 1-5 */
  rating: number
}

/** 天气数据 — 温度采用区间格式 {{min_temp}}° ~ {{max_temp}}° */
export interface WeatherData {
  /** 最低温 (动态占位 {{min_temp}}) */
  min_temp: number
  /** 最高温 (动态占位 {{max_temp}}) */
  max_temp: number
  /** 天气图标 (emoji) */
  weather_icon: string
  /** 天气描述，随语言切换 */
  weather_desc: string
  /** 城市名 */
  city_name: string
  /** 相对湿度 % */
  humidity: number
  /** 风速 m/s */
  wind_speed: number
  /** 更新时间 ISO */
  updated_at: string
}

/** 天气组件状态 */
export type WeatherState = 'idle' | 'loading' | 'success' | 'error'

/** 语言 */
export type Language = 'zh' | 'en'
