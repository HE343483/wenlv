/** 行政区划级联数据(市→区县→镇/街道)
 *  走后端 /api/map/districts 高德代理(后端自身缓存 24h),
 *  前端再以 sessionStorage 缓存 7 天——行政区划数据基本不变,避免重复请求
 */
import { get } from '@/api/request'

/** 行政区划节点 */
export interface RegionNode {
  /** 名称,如「成都市」「武侯区」「浆洗街街道」 */
  name: string
  /** 高德 adcode */
  adcode: string
  /** 级别:province / city / district / street */
  level: string
}

const CACHE_PREFIX = 'region-cache-v1:'
const CACHE_TTL = 7 * 24 * 60 * 60 * 1000

interface RawDistrict {
  name: string
  adcode: string
  level: string
  districts?: RawDistrict[]
}

/** 获取指定行政区划的下级列表(keywords 支持名称或 adcode) */
export async function fetchRegionChildren(key: string): Promise<RegionNode[]> {
  const cacheKey = CACHE_PREFIX + key
  try {
    const raw = sessionStorage.getItem(cacheKey)
    if (raw) {
      const parsed = JSON.parse(raw) as { ts: number; list: RegionNode[] }
      if (Date.now() - parsed.ts < CACHE_TTL && Array.isArray(parsed.list) && parsed.list.length > 0) {
        return parsed.list
      }
    }
  } catch {
    /* 缓存读取失败则直接请求 */
  }

  const resp = await get<{ status: string; districts?: RawDistrict[] }>('/map/districts', {
    params: { keywords: key, subdistrict: 1 },
  })
  // 高德返回 districts[0] 为关键词自身,其内部 districts 才是下级列表
  const children = resp?.districts?.[0]?.districts ?? []
  const nodes: RegionNode[] = children
    .filter(d => d && d.name && d.adcode)
    .map(d => ({ name: d.name, adcode: d.adcode, level: d.level ?? '' }))

  try {
    sessionStorage.setItem(cacheKey, JSON.stringify({ ts: Date.now(), list: nodes }))
  } catch {
    /* 存储满则放弃缓存 */
  }
  return nodes
}
