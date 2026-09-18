export interface CultureScrollSegment {
  id: string
  era: string
  eraEn: string
  period: string
  periodEn: string
  desc: string
  descEn: string
  imageUrl: string
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
    desc: '三星堆与金沙遗址代表了古蜀文明的辉煌成就。青铜神树、黄金面具、太阳神鸟等文物震惊世界，揭示出一个自成体系、想象瑰丽的早期王国：祭祀礼仪、冶铸技艺与天文信仰交织，使成都平原在中华文明版图上留下不可替代的一页。',
    descEn:
      'Sanxingdui and Jinsha reveal the brilliance of ancient Shu. Bronze sacred trees, gold masks, and the Sun Bird astonished the world, pointing to a self-contained early kingdom where ritual, metallurgy, and celestial belief intertwined — an irreplaceable chapter on the Chengdu Plain.',
    imageUrl: '/images/home/1.jpg',
  },
  {
    id: 'qin',
    era: '秦并巴蜀',
    eraEn: 'Qin Annexation',
    period: '公元前316年 — 公元221年',
    periodEn: '316 BC — 221 AD',
    desc: '秦灭巴蜀后，李冰父子主持修建都江堰：分水鱼嘴、飞沙堰与宝瓶口联动，无坝引水、岁修不辍，使岷江之利润泽平原。自此成都平原旱涝保收，号称“天府之国”，也为后来的城市繁荣与农耕文明奠定了千年水利根基。',
    descEn:
      'After Qin annexed Shu, Li Bing and his son built Dujiangyan — the fish-mouth diversion, Feishayan spillway, and Bottle-Neck Channel working as one. Dam-free irrigation made the Min River water the plain, turning Chengdu into the “Land of Abundance” and laying a hydraulic foundation for centuries of harvest and city life.',
    imageUrl: '/images/home/carousel-xiling.jpg',
  },
  {
    id: 'shu-han',
    era: '蜀汉风云',
    eraEn: 'Shu Han',
    period: '公元221年 — 263年',
    periodEn: '221 — 263 AD',
    desc: '刘备在成都称帝，建立蜀汉，以武侯治蜀、联吴抗魏书写三国格局。武侯祠、惠陵与城中街巷传说，使忠义、智谋与家国情怀沉淀为地方记忆。诸葛亮六出祁山等故事经戏曲、话本与影视不断重述，三国文化自此深深烙印在成都的血脉中。',
    descEn:
      'Liu Bei founded Shu Han in Chengdu; Zhuge Liang’s governance and northern campaigns shaped the Three Kingdoms era. The Temple of Marquis Wu, imperial tombs, and city legends keep loyalty and strategy alive in local memory — a legacy continually retold in opera, fiction, and film, etched into Chengdu’s identity.',
    imageUrl: '/images/home/hero-chengdu.jpg',
  },
  {
    id: 'tang-song',
    era: '唐宋锦绣',
    eraEn: 'Tang & Song',
    period: '公元618年 — 1279年',
    periodEn: '618 — 1279 AD',
    desc: '“锦官城”之名响彻天下：蜀锦、蜀绣与雕版印刷空前繁荣，商贸与工艺使成都成为西南都会。杜甫草堂、陆游足迹与无数诗篇，把烟火街市、江畔梅花写进中国文学史；书院、茶坊与花市兴起，城市生活细密而雅致，奠定了成都人文气质的底色。',
    descEn:
      'Chengdu flourished as the “City of Brocade.” Shu brocade, embroidery, and woodblock printing thrived; trade made it a Southwest hub. Du Fu’s Thatched Cottage, Lu You’s sojourns, and countless verses placed its streets and river plum blossoms in literary history — academies, teahouses, and flower markets shaping a refined urban culture that still defines the city.',
    imageUrl: '/images/home/154b682806f848688077dbabd5bcac3f_720.jpg',
  },
  {
    id: 'ming-qing',
    era: '明清延续',
    eraEn: 'Ming & Qing',
    period: '公元1368年 — 1911年',
    periodEn: '1368 — 1911 AD',
    desc: '战乱与重建之后，宽窄巷子、锦里等街巷格局渐次成型，少城肌理与市井烟火并存。川剧变脸吐火、川菜麻辣百味、盖碗茶馆闲坐清谈，共同养成成都特有的慢生活与公共交往方式；庙会、灯会与行会活动让城市文化在日常中绵延不绝。',
    descEn:
      'After upheaval and rebuilding, Kuanzhai Alley, Jinli, and the old Manchu city grain took shape beside lively markets. Sichuan opera’s face-changing, the hundred flavors of Sichuan cuisine, and covered-bowl teahouses forged Chengdu’s slow pace and public sociability — temple fairs, lantern festivals, and guild life keeping culture woven into everyday streets.',
    imageUrl: '/images/culture-scroll/era-scenery-shanshui-v1.jpg',
  },
  {
    id: 'modern',
    era: '现代成都',
    eraEn: 'Modern Chengdu',
    period: '1911年 — 至今',
    periodEn: '1911 — Present',
    desc: '从近代开埠到当代建设，千年古都持续焕新：公园城市理念推动绿道、湿地与街区更新，天府新区与科学城拓展发展空间，大运会等国际盛会让世界看见开放的成都。科技、文创与烟火市井并存，一座向未来生长、又守住闲适气质的超大城市正在展开新篇。',
    descEn:
      'From early modern openings to today’s building boom, the ancient capital keeps renewing itself. The Park City vision weaves greenways and neighborhoods; Tianfu New Area and science hubs expand the map; global events like the Universiade put Chengdu on the world stage. Tech, culture, and street life coexist — a megacity growing toward the future while keeping its easygoing soul.',
    imageUrl: '/images/home/97178dc100d4868a7d4cb804e37ef902_720.jpg',
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

export interface CultureScrollDecor {
  id: string
  imageUrl: string
  /** Horizontal position as % of mid-layer width */
  xPercent: number
  side: 'above' | 'below'
  size: 'sm' | 'md'
}

/** Decorative photos between era nodes, alternating above/below the axis. */
export const CULTURE_SCROLL_DECORS: CultureScrollDecor[] = [
  {
    id: 'decor-bamboo',
    imageUrl: '/images/home/carousel-panda.jpg',
    xPercent: 7.5,
    side: 'below',
    size: 'sm',
  },
  {
    id: 'decor-mist',
    imageUrl: '/images/home/carousel-dujiangyan.jpg',
    xPercent: 18.2,
    side: 'above',
    size: 'md',
  },
  {
    id: 'decor-valley',
    imageUrl: '/images/home/carousel-xiling.jpg',
    xPercent: 32.4,
    side: 'below',
    size: 'sm',
  },
  {
    id: 'decor-ink',
    imageUrl: '/images/culture-scroll/era-modern-shanshui-v1.png',
    xPercent: 51.2,
    side: 'above',
    size: 'md',
  },
  {
    id: 'decor-night',
    imageUrl: '/images/home/0.jpg',
    xPercent: 66.8,
    side: 'below',
    size: 'md',
  },
  {
    id: 'decor-tower',
    imageUrl: '/images/home/carousel-kuanzhai.jpg',
    xPercent: 81.5,
    side: 'above',
    size: 'sm',
  },
  {
    id: 'decor-glow',
    imageUrl: '/images/home/carousel-jinli.jpg',
    xPercent: 95.2,
    side: 'below',
    size: 'sm',
  },
]
