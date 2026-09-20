/**
 * passport.ts — 数字足迹护照（localStorage 集章）
 * 浏览景点详情即"盖章"，数据存 localStorage：免登录、免后端；
 * 换设备/清缓存会丢（MVP 取舍，账号体系预留到后续）。
 */

const STORAGE_KEY = 'wenlv.passport.stamps'

/** 单枚印章：景点三语名 + 盖章时间戳 */
export interface PassportStamp {
  id: number
  name_zh: string
  name_en: string
  name_ja: string
  ts: number
}

/** addStamp 入参（与 ScenicItem 的相关字段对齐） */
export interface StampSource {
  id: number
  name_zh: string
  name_en?: string
  name_ja?: string
}

/** 读取并逐字段校验本地印章列表（JSON 解析失败/结构异常一律按空处理） */
function readStamps(): PassportStamp[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter((it): it is Record<string, unknown> => !!it && typeof it === 'object')
      .map((it) => ({
        id: Number(it.id),
        name_zh: typeof it.name_zh === 'string' ? it.name_zh : '',
        name_en: typeof it.name_en === 'string' ? it.name_en : '',
        name_ja: typeof it.name_ja === 'string' ? it.name_ja : '',
        ts: Number(it.ts) || 0,
      }))
      .filter((it) => Number.isFinite(it.id) && it.id > 0)
  } catch {
    return []
  }
}

function writeStamps(list: PassportStamp[]): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(list))
  } catch {
    /* 隐私模式/容量不足时静默失败，不影响主流程 */
  }
}

/** 盖章：按 id 去重，返回盖章后的完整列表 */
export function addStamp(item: StampSource): PassportStamp[] {
  const list = readStamps()
  if (list.some((s) => s.id === item.id)) return list
  list.push({
    id: item.id,
    name_zh: item.name_zh,
    name_en: item.name_en || '',
    name_ja: item.name_ja || '',
    ts: Date.now(),
  })
  writeStamps(list)
  return list
}

/** 全部印章（按盖章时间升序） */
export function getStamps(): PassportStamp[] {
  return readStamps().sort((a, b) => a.ts - b.ts)
}

/** 已集章数量 */
export function stampCount(): number {
  return readStamps().length
}

/** 是否已盖过某景点 */
export function hasStamp(id: number): boolean {
  return readStamps().some((s) => s.id === id)
}
