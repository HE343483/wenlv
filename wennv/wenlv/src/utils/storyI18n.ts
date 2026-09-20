/** 多语种故事展示工具:景点/美食/路线共用的介绍与文化注解取值纯函数。 */
import type { Language } from '@/types'

/** 把 unknown 字段安全读成字符串(非字符串一律视为空) */
function asText(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

/**
 * pickDesc 按语言取介绍正文:
 * zh 返回 desc;en 返回 desc_en || desc;ja 返回 desc_ja || desc。
 * 外语故事未生成(空串)时自动回落中文 desc。
 */
export function pickDesc(item: unknown, lang: Language): string {
  const rec = (item ?? {}) as Record<string, unknown>
  const zh = asText(rec.desc)
  if (lang === 'zh') return zh
  if (lang === 'en') return asText(rec.desc_en) || zh
  return asText(rec.desc_ja) || zh
}

/**
 * pickName 按语言取名称:
 * zh 返回 name_zh;en 返回 name_en || name_zh;ja 返回 name_ja || name_en || name_zh。
 * 对应语言名称未生成时逐级回落,保证总有值可展示。
 */
export function pickName(item: unknown, lang: Language): string {
  const rec = (item ?? {}) as Record<string, unknown>
  const zh = asText(rec.name_zh)
  if (lang === 'zh') return zh
  if (lang === 'en') return asText(rec.name_en) || zh
  return asText(rec.name_ja) || asText(rec.name_en) || zh
}

/**
 * pickCultureNote 按语言取文化注解(按 \n 拆分为数组,每条一句)。
 * 中文语言返回空数组(注解仅面向外国游客展示);对应语言未生成时也返回空数组。
 */
export function pickCultureNote(item: unknown, lang: Language): string[] {
  if (lang === 'zh') return []
  const rec = (item ?? {}) as Record<string, unknown>
  const raw = lang === 'en' ? asText(rec.culture_note_en) : asText(rec.culture_note_ja)
  if (!raw) return []
  return raw
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
}
