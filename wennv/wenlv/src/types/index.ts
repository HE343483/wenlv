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
  /** 地图坐标 (百度地图 BD-09 近似) */
  coords: {
    lng: number
    lat: number
  }
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

/** 新闻热点 */
export interface NewsItem {
  id: string
  /** 新闻标题 */
  titleZh: string
  titleEn: string
  /** 来源 / 频道标签 */
  sourceZh: string
  sourceEn: string
  /** 发布时间 (显示用) */
  timeZh: string
  timeEn: string
  /** 摘要 */
  summaryZh: string
  summaryEn: string
  /** 是否置顶/热门 */
  hot?: boolean
}

/** 趣闻故事 */
export interface FunFact {
  id: string
  titleZh: string
  titleEn: string
  contentZh: string
  contentEn: string
  /** emoji 图标 */
  icon: string
}

/** 路线中的站点 */
export interface RouteStop {
  /** 对应景点ID */
  spotId: string
  /** 站点备注 (如 "上午" / "下午") */
  noteZh: string
  noteEn: string
}

/** 精品旅游路线 */
export interface TravelRoute {
  id: string
  nameZh: string
  nameEn: string
  /** 行程时长 */
  durationZh: string
  durationEn: string
  /** 主题 */
  themeZh: string
  themeEn: string
  /** 难度/强度 */
  levelZh: string
  levelEn: string
  summaryZh: string
  summaryEn: string
  /** 路线站点 */
  stops: RouteStop[]
  /** 行程亮点 */
  highlightsZh: string[]
  highlightsEn: string[]
}
