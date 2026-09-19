/** 景点详情页展示层纯函数:解析后端字段并对空值做统一降级。 */
import type { ScenicDetailSection } from '@/api/content'

/** splitList 把逗号分隔串拆成数组,自动去空。 */
export function splitList(value?: string): string[] {
  return (value || '')
    .split(',')
    .map((v) => v.trim())
    .filter(Boolean)
}

/** parseSections 解析后端 detail_sections JSON;格式异常返回空数组,页面自动隐藏该区块。 */
export function parseSections(value?: string): ScenicDetailSection[] {
  if (!value) return []
  let raw: unknown
  try {
    raw = JSON.parse(value)
  } catch {
    return []
  }
  if (!Array.isArray(raw)) return []
  const out: ScenicDetailSection[] = []
  for (const item of raw as Array<Record<string, unknown>>) {
    if (!item || typeof item.title !== 'string' || typeof item.text !== 'string') continue
    out.push({
      title: item.title,
      text: item.text,
      image: typeof item.image === 'string' ? item.image : '',
    })
  }
  return out
}

/** estimatedSet 参考值字段集合,用于给 LLM 生成的数据加"参考值"标注。 */
export function estimatedSet(value?: string): Set<string> {
  return new Set(splitList(value))
}

/** displayFact 展示值:空值统一显示占位符,避免页面出现加载文案。 */
export function displayFact(value: string | undefined, placeholder: string): string {
  const v = (value || '').trim()
  return v || placeholder
}
