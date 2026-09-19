/**
 * curatedRoutes.ts — 精选主题路线静态数据
 * 路线页（/home/routes）展示用：卡片 + 详情弹窗，并可一键导流到 AI 行程（/trip）预填表单。
 * 说明：title/desc/tags 等文案放在 locales 的 routes.items.<id> 下保证三语切换；
 *      景点为专有名词，直接在此处提供 zh/en/ja 三语名称。
 *      interests 取值需与 LandingView.vue interestOptions 的 value 一致，
 *      以便「用 AI 生成同款行程」时能直接预填偏好。
 */

export type Locale = 'zh' | 'en' | 'ja'

export interface CuratedRouteStop {
  zh: string
  en: string
  ja: string
}

export interface CuratedRoute {
  /** 唯一 id，同时用于 locales 的 routes.items.<id> 取文案 */
  id: string
  /** 建议游玩天数（预填 AI 行程表单的城市天数） */
  days: number
  /** 偏好标签（与 AI 行程 interestOptions.value 对齐，用于预填 preferences） */
  interests: string[]
  /** 卡片/弹窗头图 */
  image: string
  /** 途经景点（顺序即游玩顺序） */
  stops: CuratedRouteStop[]
}

/** 按当前语言取景点名称 */
export function stopName(stop: CuratedRouteStop, lang: Locale): string {
  return stop[lang] || stop.zh
}

export const curatedRoutes: CuratedRoute[] = [
  {
    id: 'classic',
    days: 1,
    interests: ['历史文化', '美食'],
    image:
      'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=Chengdu%20Kuanzhai%20Alley%20historic%20street%20with%20traditional%20Sichuan%20architecture%2C%20red%20lanterns%2C%20golden%20sunset%20light%2C%20tourists%20strolling%2C%20photorealistic%20travel%20photography%2C%20warm%20tones&image_size=landscape_4_3',
    stops: [
      { zh: '武侯祠', en: 'Wuhou Shrine', ja: '武侯祠' },
      { zh: '锦里古街', en: 'Jinli Ancient Street', ja: '錦里古街' },
      { zh: '宽窄巷子', en: 'Kuanzhai Alley', ja: '寛窄巷子' },
      { zh: '人民公园', en: "People's Park", ja: '人民公園' },
      { zh: '春熙路·太古里', en: 'Chunxi Road & Taikoo Li', ja: '春熙路・太古里' },
    ],
  },
  {
    id: 'panda',
    days: 2,
    interests: ['自然风光', '休闲'],
    image:
      'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=Giant%20panda%20eating%20bamboo%20in%20lush%20green%20enclosure%2C%20Chengdu%20panda%20base%2C%20cute%20fluffy%20panda%20close%20up%2C%20soft%20morning%20light%2C%20photorealistic%20wildlife%20photography&image_size=landscape_4_3',
    stops: [
      { zh: '大熊猫繁育研究基地', en: 'Chengdu Panda Base', ja: 'パンダ繁殖研究基地' },
      { zh: '文殊院', en: 'Wenshu Monastery', ja: '文殊院' },
      { zh: '东郊记忆', en: 'Eastern Suburb Memory', ja: '東郊記憶' },
      { zh: '都江堰', en: 'Dujiangyan Irrigation System', ja: '都江堰' },
      { zh: '南桥夜景', en: 'Nanqiao Bridge Night View', ja: '南橋の夜景' },
    ],
  },
  {
    id: 'food',
    days: 1,
    interests: ['美食', '休闲'],
    image:
      'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=Sichuan%20street%20food%20night%20market%2C%20spicy%20hotpot%20skewers%20and%20mapo%20tofu%2C%20red%20lanterns%2C%20steaming%20wok%2C%20bustling%20Chengdu%20food%20street%20at%20night%2C%20photorealistic%20food%20photography%2C%20vibrant%20colors&image_size=landscape_4_3',
    stops: [
      { zh: '建设路小吃街', en: 'Jianshe Road Food Street', ja: '建設路グルメ街' },
      { zh: '玉林路小酒馆', en: 'Yulin Road & Little Bar', ja: '玉林路の酒場' },
      { zh: '望平街·香香巷', en: 'Wangping Street & Xiangxiang Lane', ja: '望平街・香香巷' },
      { zh: '九眼桥酒吧街', en: 'Jiuyanqiao Bar Street', ja: '九眼橋バーストリート' },
    ],
  },
  {
    id: 'culture',
    days: 3,
    interests: ['历史文化', '艺术'],
    image:
      'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=Ancient%20bronze%20mask%20of%20Sanxingdui%20museum%2C%20mysterious%20Shu%20civilization%20artifact%2C%20dramatic%20museum%20lighting%2C%20dark%20background%2C%20golden%20bronze%20texture%2C%20photorealistic%20close%20up&image_size=landscape_4_3',
    stops: [
      { zh: '三星堆博物馆', en: 'Sanxingdui Museum', ja: '三星堆博物館' },
      { zh: '杜甫草堂', en: 'Du Fu Thatched Cottage', ja: '杜甫草堂' },
      { zh: '四川博物院', en: 'Sichuan Museum', ja: '四川博物院' },
      { zh: '青城山', en: 'Mount Qingcheng', ja: '青城山' },
      { zh: '夜游锦江', en: 'Jinjiang River Night Cruise', ja: '錦江ナイトクルーズ' },
    ],
  },
]

/** 按 id 查找路线 */
export function findCuratedRoute(id: string | null | undefined): CuratedRoute | undefined {
  if (!id) return undefined
  return curatedRoutes.find(r => r.id === id)
}
