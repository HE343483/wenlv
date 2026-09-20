/**
 * AI 行程 .ics 日历导出工具(纯前端、零第三方依赖,符合 RFC 5545)。
 *
 * 事件规则:
 * - 每天每个景点一个 VEVENT;第 N 天的日期 = 行程起始日(start_date) + (N-1) 天,
 *   单日 day.date 可解析时优先采用单日日期;
 * - 行程计划本身无具体时段,景点按当天顺序均分到 上午 09:00-12:00 / 下午 14:00-18:00 两块;
 * - 某天完全无日期信息(start_date 与 day.date 均不可解析)时,该天整体生成一个
 *   全天事件(DTSTART;VALUE=DATE),并以"今天为第 1 天"兜底推算日期;
 * - 时刻统一按东八区(UTC+8,中国无夏令时)换算为 UTC 时间戳(DTSTART:...Z),
 *   无需内嵌 VTIMEZONE,任何时区的日历 App 打开都显示正确的北京时间。
 *
 * 对应界面文案为 i18n 的 ics.* 键组(src/i18n/locales/{zh,en,ja}.json)。
 */
import type { Language } from '@/types'
import type { Attraction, DayPlan, TripPlan } from '@/types/trip'
import { pickDesc } from '@/utils/storyI18n'

/** 历法日期(仅年月日,不绑定时刻与时区,避免 Date 本地时区干扰) */
interface CalendarDate {
  year: number
  month: number
  day: number
}

/** 单个景点在当天内的起止时刻(自 00:00 起的分钟数) */
interface DaySlot {
  startMin: number
  endMin: number
}

/** 上午/下午两个时段窗口(分钟数,自当天 00:00 起) */
const MORNING_START_MIN = 9 * 60
const MORNING_END_MIN = 12 * 60
const AFTERNOON_START_MIN = 14 * 60
const AFTERNOON_END_MIN = 18 * 60

/** 全天事件标题中的"第 N 天"按界面语言措辞 */
const DAY_LABELS: Record<Language, (dayNumber: number) => string> = {
  zh: (n) => `第 ${n} 天`,
  en: (n) => `Day ${n}`,
  ja: (n) => `${n} 日目`,
}

const textEncoder = new TextEncoder()

/** 字符串的 UTF-8 八位组长度(ICS 行折叠按字节而非字符计) */
function utf8Octets(value: string): number {
  return textEncoder.encode(value).length
}

/**
 * 按 RFC 5545 折叠超长物理行:单行不超过 75 八位组,
 * 续行以单个空格开头(该空格计入续行的 75 八位组限制)。
 */
function foldLine(line: string): string {
  const LIMIT = 75
  if (utf8Octets(line) <= LIMIT) return line
  const parts: string[] = []
  let current = ''
  let currentOctets = 0
  for (const ch of line) {
    const chOctets = utf8Octets(ch)
    // 首行可用满 75 八位组,续行需预留 1 八位组给行首空格
    const cap = parts.length === 0 ? LIMIT : LIMIT - 1
    if (currentOctets > 0 && currentOctets + chOctets > cap) {
      parts.push(current)
      current = ''
      currentOctets = 0
    }
    current += ch
    currentOctets += chOctets
  }
  if (current) parts.push(current)
  return parts.join('\r\n ')
}

/** TEXT 值转义(RFC 5545 §3.3.11):反斜杠/分号/逗号/换行 */
function escapeText(value: string): string {
  return value
    .replace(/\\/g, '\\\\')
    .replace(/;/g, '\\;')
    .replace(/,/g, '\\,')
    .replace(/\r?\n/g, '\\n')
}

/**
 * 解析日期字符串为历法日期,主格式 YYYY-MM-DD(表单提交格式),
 * 兼容 YYYY/M/D、YYYY.M.D;非法日期(如 2 月 30 日)返回 null。
 */
function parseCalendarDate(raw: string | undefined | null): CalendarDate | null {
  if (!raw) return null
  const matched = String(raw).trim().match(/^(\d{4})[-/.](\d{1,2})[-/.](\d{1,2})/)
  if (!matched) return null
  const year = Number(matched[1] ?? NaN)
  const month = Number(matched[2] ?? NaN)
  const day = Number(matched[3] ?? NaN)
  const probe = new Date(Date.UTC(year, month - 1, day))
  // 用 UTC 往返校验,拦截 2026-02-30 这类字面合法但日历不存在的日期
  if (
    !Number.isFinite(year) ||
    probe.getUTCFullYear() !== year ||
    probe.getUTCMonth() !== month - 1 ||
    probe.getUTCDate() !== day
  ) {
    return null
  }
  return { year, month, day }
}

/** 在历法日期上加 n 天(UTC 日期运算,无时区副作用) */
function addDays(date: CalendarDate, n: number): CalendarDate {
  const shifted = new Date(Date.UTC(date.year, date.month - 1, date.day + n))
  return {
    year: shifted.getUTCFullYear(),
    month: shifted.getUTCMonth() + 1,
    day: shifted.getUTCDate(),
  }
}

/** 历法日期 → ICS 全天日期值(YYYYMMDD) */
function toDateStamp(date: CalendarDate): string {
  const mm = String(date.month).padStart(2, '0')
  const dd = String(date.day).padStart(2, '0')
  return `${date.year}${mm}${dd}`
}

/** 把"东八区某天的 HH:mm"换算为 UTC 时间戳值(YYYYMMDDTHHMMSSZ,固定 -8 小时) */
function toUtcStamp(date: CalendarDate, hhmm: string): string {
  const [hhRaw, mmRaw] = hhmm.split(':')
  const utc = new Date(Date.UTC(date.year, date.month - 1, date.day, Number(hhRaw ?? 0) - 8, Number(mmRaw ?? 0)))
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${utc.getUTCFullYear()}${pad(utc.getUTCMonth() + 1)}${pad(utc.getUTCDate())}T${pad(utc.getUTCHours())}${pad(utc.getUTCMinutes())}00Z`
}

/** 当天分钟数 → UTC 时间戳值 */
function minutesToUtcStamp(date: CalendarDate, minutes: number): string {
  const hh = String(Math.floor(minutes / 60)).padStart(2, '0')
  const mm = String(minutes % 60).padStart(2, '0')
  return toUtcStamp(date, `${hh}:${mm}`)
}

/** 当前时刻的 DTSTAMP 值(UTC) */
function utcNowStamp(): string {
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${now.getUTCFullYear()}${pad(now.getUTCMonth() + 1)}${pad(now.getUTCDate())}T${pad(now.getUTCHours())}${pad(now.getUTCMinutes())}${pad(now.getUTCSeconds())}Z`
}

/** 本地"今天"的历法日期(无日期信息时以今天为第 1 天兜底) */
function localToday(): CalendarDate {
  const now = new Date()
  return { year: now.getFullYear(), month: now.getMonth() + 1, day: now.getDate() }
}

/**
 * 把当天景点按顺序均分到上午/下午两块:
 * 前 ceil(n/2) 个均分 09:00-12:00,其余均分 14:00-18:00;
 * n<=2 时恰好对应"上午一块、下午一块"。
 */
function buildDaySlots(count: number): DaySlot[] {
  if (count <= 0) return []
  const morningCount = Math.ceil(count / 2)
  const afternoonCount = count - morningCount
  const blockSlots = (from: number, to: number, n: number): DaySlot[] => {
    const result: DaySlot[] = []
    const span = to - from
    for (let i = 0; i < n; i += 1) {
      result.push({
        startMin: from + Math.round((span * i) / n),
        endMin: from + Math.round((span * (i + 1)) / n),
      })
    }
    return result
  }
  return [
    ...blockSlots(MORNING_START_MIN, MORNING_END_MIN, morningCount),
    ...blockSlots(AFTERNOON_START_MIN, AFTERNOON_END_MIN, afternoonCount),
  ]
}

/** FNV-1a 32 位散列:由"景点标识+日期"生成稳定唯一的 UID 前缀 */
function hash32(input: string): string {
  let hash = 0x811c9dc5
  for (let i = 0; i < input.length; i += 1) {
    hash ^= input.charCodeAt(i)
    hash = Math.imul(hash, 0x01000193)
  }
  return (hash >>> 0).toString(16).padStart(8, '0')
}

/** 读取景点上可能存在的多语简介扩展字段(desc_en/desc_ja,类型未声明但后端可能附带) */
function extraText(attraction: Attraction, key: string): string {
  const value = (attraction as unknown as Record<string, unknown>)[key]
  return typeof value === 'string' ? value : ''
}

/** 单个 VEVENT 的属性列表:键为完整属性名(可含 ;VALUE=DATE 参数),值为未经转义的原文 */
type EventProperty = [name: string, rawValue: string]

/**
 * 把行程导出为 .ics 日历文件。
 * @param plan 行程计划数据(含 start_date/days/attractions)
 * @param lang 当前界面语言,用于 DESCRIPTION 多语取值与全天事件标题措辞
 * @returns text/calendar Blob,可直接触发下载
 */
export function exportTripIcs(plan: TripPlan, lang: Language): Blob {
  const days: DayPlan[] = Array.isArray(plan.days) ? plan.days : []
  const planStart = parseCalendarDate(plan.start_date)
  const today = localToday()
  const stamp = utcNowStamp()
  const events: string[][] = []

  const pushEvent = (uid: string, props: EventProperty[]) => {
    const eventLines = ['BEGIN:VEVENT', `UID:${uid}`, `DTSTAMP:${stamp}`]
    props.forEach(([name, rawValue]) => eventLines.push(`${name}:${escapeText(rawValue)}`))
    eventLines.push('END:VEVENT')
    events.push(eventLines)
  }

  days.forEach((day, index) => {
    const dayNumber = index + 1
    const spots = (Array.isArray(day?.attractions) ? day.attractions : []).filter((spot) => Boolean(spot?.name))
    // 日期解析优先级:单日 date > 行程起始日 + (N-1) > 无日期(全天事件兜底)
    const dayDate = parseCalendarDate(day?.date) ?? (planStart ? addDays(planStart, index) : null)

    if (dayDate) {
      const slots = buildDaySlots(spots.length)
      spots.forEach((spot, slotIndex) => {
        const slot = slots[slotIndex]
        if (!slot) return
        const descriptionParts = [
          pickDesc({ desc: spot.description ?? '', desc_en: extraText(spot, 'desc_en'), desc_ja: extraText(spot, 'desc_ja') }, lang),
        ]
        if (spot.address) descriptionParts.push(`📍${spot.address}`)
        pushEvent(`${hash32(`spot|${dayNumber}|${slotIndex}|${spot.name}|${toDateStamp(dayDate)}`)}@wenlv-trip`, [
          ['DTSTART', minutesToUtcStamp(dayDate, slot.startMin)],
          ['DTEND', minutesToUtcStamp(dayDate, slot.endMin)],
          ['SUMMARY', spot.name],
          ['LOCATION', spot.name],
          ['DESCRIPTION', descriptionParts.filter(Boolean).join('\n')],
        ])
      })
      return
    }

    // 该天无任何日期信息:整体生成一个全天事件(今天为第 1 天兜底),汇总当天所有站点
    const fallbackDate = addDays(today, index)
    const overviewParts = [
      day?.description ?? '',
      ...spots.map((spot) => `· ${spot.name}`),
      day?.accommodation ? `🏨 ${day.accommodation}` : '',
    ]
    pushEvent(`${hash32(`allday|${dayNumber}|${plan.city}|${toDateStamp(fallbackDate)}`)}@wenlv-trip`, [
      ['DTSTART;VALUE=DATE', toDateStamp(fallbackDate)],
      ['DTEND;VALUE=DATE', toDateStamp(addDays(fallbackDate, 1))],
      ['SUMMARY', `${plan.city} · ${DAY_LABELS[lang](dayNumber)}`],
      ['LOCATION', plan.city],
      ['DESCRIPTION', overviewParts.filter(Boolean).join('\n')],
    ])
  })

  const calendarName = [plan.city, plan.start_date, plan.end_date].filter(Boolean).join(' ')
  const calendarLines = [
    'BEGIN:VCALENDAR',
    'VERSION:2.0',
    'PRODID:-//wenlv//AI Trip Calendar 1.0//ZH',
    'CALSCALE:GREGORIAN',
    'METHOD:PUBLISH',
    `X-WR-CALNAME:${escapeText(calendarName)}`,
    'X-WR-TIMEZONE:Asia/Shanghai',
    ...events.flat(),
    'END:VCALENDAR',
  ]

  // CRLF 换行 + 75 八位组行折叠,结尾以 CRLF 收束(RFC 5545 要求)
  const icsText = `${calendarLines.map(foldLine).join('\r\n')}\r\n`
  return new Blob([icsText], { type: 'text/calendar;charset=utf-8' })
}
