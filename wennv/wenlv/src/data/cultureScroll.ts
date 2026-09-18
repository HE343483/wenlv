export interface CultureScrollSegment {
  id: string
  era: string
  eraEn: string
  period: string
  periodEn: string
  desc: string
  descEn: string
  progressStart: number
  progressEnd: number
}

export interface CultureScrollHotspot {
  segmentId: string
  xPercent: number
  label?: string
}

const TIMELINE_FIELDS = [
  {
    id: 'ancient-shu',
    era: '古蜀时期',
    eraEn: 'Ancient Shu',
    period: '约公元前1600年 — 公元前316年',
    periodEn: 'c. 1600 BC — 316 BC',
    desc: '三星堆与金沙遗址代表了古蜀文明的辉煌成就，青铜神树、黄金面具、太阳神鸟等文物震惊世界。',
    descEn:
      'Sanxingdui and Jinsha sites represent the brilliance of ancient Shu — bronze trees, gold masks, and the Sun Bird.',
  },
  {
    id: 'qin',
    era: '秦并巴蜀',
    eraEn: 'Qin Annexation',
    period: '公元前316年 — 公元221年',
    periodEn: '316 BC — 221 AD',
    desc: '秦灭巴蜀后，李冰父子修建都江堰水利工程，使成都平原成为"天府之国"。',
    descEn:
      'After Qin conquered Shu, Li Bing and his son built the Dujiangyan irrigation system, turning Chengdu into the "Land of Abundance."',
  },
  {
    id: 'shu-han',
    era: '蜀汉风云',
    eraEn: 'Shu Han',
    period: '公元221年 — 263年',
    periodEn: '221 — 263 AD',
    desc: '刘备在成都称帝建立蜀汉政权，诸葛亮六出祁山，三国文化自此深深烙印在成都的血脉中。',
    descEn:
      "Liu Bei founded Shu Han in Chengdu. Zhuge Liang's northern campaigns etched Three Kingdoms legacy into the city's soul.",
  },
  {
    id: 'tang-song',
    era: '唐宋锦绣',
    eraEn: 'Tang & Song',
    period: '公元618年 — 1279年',
    periodEn: '618 — 1279 AD',
    desc: '"锦官城"之名响彻天下，杜甫、陆游等诗人留居成都，留下无数传世诗篇。蜀锦、蜀绣、雕版印刷空前繁荣。',
    descEn:
      'Chengdu flourished as the "City of Brocade." Poets Du Fu and Lu You lived here, leaving timeless verses.',
  },
  {
    id: 'ming-qing',
    era: '明清延续',
    eraEn: 'Ming & Qing',
    period: '公元1368年 — 1911年',
    periodEn: '1368 — 1911 AD',
    desc: '宽窄巷子、锦里等明清古街格局形成，川剧、川菜、茶馆文化日趋成熟，成都慢生活文化源远流长。',
    descEn:
      "Kuanzhai Alley and Jinli took shape. Sichuan opera, cuisine, and teahouse culture matured into the city's signature slow pace.",
  },
  {
    id: 'modern',
    era: '现代成都',
    eraEn: 'Modern Chengdu',
    period: '1911年 — 至今',
    periodEn: '1911 — Present',
    desc: '千年古都焕发新生，以公园城市理念建设践行新发展理念，天府新区、大运会场馆见证城市蝶变。',
    descEn:
      'The ancient capital reinvents itself as a "Park City." Tianfu New Area and global events showcase Chengdu\'s transformation.',
  },
] as const

const SEGMENT_COUNT = TIMELINE_FIELDS.length

export const CULTURE_SCROLL_SEGMENTS: CultureScrollSegment[] = TIMELINE_FIELDS.map(
  (item, index) => ({
    ...item,
    progressStart: index / SEGMENT_COUNT,
    progressEnd: (index + 1) / SEGMENT_COUNT,
  }),
)

/** Mid-layer hit points, `left` as % of the 600vw / 7200-unit scroll. */
export const CULTURE_SCROLL_HOTSPOTS: CultureScrollHotspot[] = [
  { segmentId: 'ancient-shu', xPercent: 12.6, label: '神树' },
  { segmentId: 'qin', xPercent: 23.3, label: '都江堰' },
  { segmentId: 'shu-han', xPercent: 41.5, label: '城阙' },
  { segmentId: 'tang-song', xPercent: 60.1, label: '锦官城' },
  { segmentId: 'ming-qing', xPercent: 73.3, label: '街巷' },
  { segmentId: 'modern', xPercent: 90, label: '天府' },
]
