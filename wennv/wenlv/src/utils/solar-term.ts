/**
 * solar-term.ts — 二十四节气工具(内容数据,非 UI 标签)
 * 1. 节气近似公历日期表:二十四节气每年公历日期相当固定(±1 天浮动,如冬至 12/21-12/22),
 *    这里取常见年份的众数日期作近似日历表,容差 ±1 天,足以支撑"当天处于哪个节气"的条幅展示;
 *    如需精确到当日,可后续接入天文历法库,数据结构无需变化。
 * 2. 节气三语名称 + 蜀俗条幅文案(每组 1-2 句:节气名 + 一句蜀地食俗)。
 *    这是内容数据而非 UI 标签,按约定内联为本文件 TS 常量,不放入 locales。
 */

/** 单个节气的三语名称、近似公历日期与三语条幅文案 */
export interface SolarTermInfo {
  /** 中文节气名(与后端 foods.solar_terms 打标值保持一致) */
  name: string
  nameEn: string
  nameJa: string
  /** 近似公历月(1-12) */
  month: number
  /** 近似公历日 */
  day: number
  /** 三语条幅文案(节气名 + 一句蜀俗) */
  banner: { zh: string; en: string; ja: string }
  /** 节气区间文本(交节日—下一交节日前一天,三语);仅 getCurrentSolarTerm 返回时填充 */
  rangeText?: { zh: string; en: string; ja: string }
}

/** 冬至(年末最后一个节气):单独提取,供 1 月初未到小寒时回落上一年的冬至 */
const DONGZHI_TERM: SolarTermInfo = {
  name: '冬至', nameEn: 'Winter Solstice (Dongzhi)', nameJa: '冬至',
  month: 12, day: 22,
  banner: {
    zh: '冬至时节，蜀人这天要喝一碗热腾腾的羊肉汤，浑身通泰过寒冬。',
    en: 'It is the season of Dongzhi (Winter Solstice) — Sichuanese mark the day with a steaming bowl of mutton soup to warm body and soul through the cold.',
    ja: '冬至の時節。四川では熱々の羊肉スープを飲んで、寒い冬を元気に乗り切ります。',
  },
}

/** 二十四节气表:按公历年内出现顺序排列(小寒 → 冬至) */
const SOLAR_TERMS: SolarTermInfo[] = [
  {
    name: '小寒', nameEn: 'Minor Cold (Xiaohan)', nameJa: '小寒',
    month: 1, day: 6,
    banner: {
      zh: '小寒时节，数九寒天，蜀地羊肉汤与火锅最抚人心，暖到骨子里。',
      en: "It is the season of Xiaohan (Minor Cold), the depths of winter — nothing comforts Sichuan hearts like mutton soup and hotpot, warming you to the bone.",
      ja: '小寒の時節。本格的な寒さの中、四川では羊肉スープや火鍋が心も体も骨の髄まで温めてくれます。',
    },
  },
  {
    name: '大寒', nameEn: 'Major Cold (Dahan)', nameJa: '大寒',
    month: 1, day: 20,
    banner: {
      zh: '大寒时节，岁末将至，蜀家灶头飘出腊味香，忙年备菜迎新春。',
      en: "It is the season of Dahan (Major Cold) — the year's end nears, and aromas of cured meats drift from Sichuan kitchens as families prepare for Spring Festival.",
      ja: '大寒の時節。歳末が近づき、四川の台所からは腊味の香りが漂い、春節の準備に追われます。',
    },
  },
  {
    name: '立春', nameEn: 'Start of Spring (Lichun)', nameJa: '立春',
    month: 2, day: 4,
    banner: {
      zh: '立春时节，蜀人咬春吃春卷，迎一年的头一口鲜。',
      en: "It is the season of Lichun (Start of Spring) — Sichuan folks 'bite the spring' with fresh spring rolls to welcome the year's first taste.",
      ja: '立春の時節。四川では春巻をかじって、一年の最初の春の味を楽しみます。',
    },
  },
  {
    name: '雨水', nameEn: 'Rain Water (Yushui)', nameJa: '雨水',
    month: 2, day: 19,
    banner: {
      zh: '雨水时节，恰逢元宵前后，蜀人吃一碗汤圆，甜糯迎春耕。',
      en: 'It is the season of Yushui (Rain Water), around the Lantern Festival — Sichuanese enjoy sweet glutinous rice balls to greet the farming season.',
      ja: '雨水の時節。元宵節の頃にあたり、四川では甘いゴマ団子のようなタンユェンを食べて春の農作業を迎えます。',
    },
  },
  {
    name: '惊蛰', nameEn: 'Awakening of Insects (Jingzhe)', nameJa: '啓蟄',
    month: 3, day: 6,
    banner: {
      zh: '惊蛰时节，春雷唤虫出，蜀人吃梨润燥，盼一年平安。',
      en: 'It is the season of Jingzhe (Awakening of Insects) — as spring thunder wakes the insects, Sichuanese eat pears to moisten the throat and wish for a safe year.',
      ja: '啓蟄の時節。春の雷が虫を起こすころ、四川では梨を食べて喉を潤し、一年の平安を願います。',
    },
  },
  {
    name: '春分', nameEn: 'Spring Equinox (Chunfen)', nameJa: '春分',
    month: 3, day: 21,
    banner: {
      zh: '春分时节，昼夜均分，蜀地踏青尝鲜，正是掐春芽的时节。',
      en: 'It is the season of Chunfen (Spring Equinox) — day and night stand equal; in Shu it is time for spring outings and tasting tender spring greens.',
      ja: '春分の時節。昼と夜が同じ長さになり、四川では春の新芽を味わいながらお花見散策を楽しみます。',
    },
  },
  {
    name: '清明', nameEn: 'Pure Brightness (Qingming)', nameJa: '清明',
    month: 4, day: 5,
    banner: {
      zh: '清明时节，蜀人踏青祭祖，青团艾粑飘香，追思也尝春。',
      en: 'It is the season of Qingming (Pure Brightness) — Sichuanese visit ancestors\u2019 tombs and savor qingtuan and mugwort cakes amid spring outings.',
      ja: '清明の時節。四川ではお墓参りをしながら、ヨモギの団子(青団)の香りを楽しみます。',
    },
  },
  {
    name: '谷雨', nameEn: 'Grain Rain (Guyu)', nameJa: '穀雨',
    month: 4, day: 20,
    banner: {
      zh: '谷雨时节，雨生百谷，蒙顶山上新茶正好，一盏春茶敬时节。',
      en: 'It is the season of Guyu (Grain Rain) — \u201crain feeds a hundred grains\u201d; fresh Mengding Mountain tea is at its very best this season.',
      ja: '穀雨の時節。雨が百穀を育むころ、蒙頂山の新茶がちょうど飲みごろです。',
    },
  },
  {
    name: '立夏', nameEn: 'Start of Summer (Lixia)', nameJa: '立夏',
    month: 5, day: 6,
    banner: {
      zh: '立夏时节，蜀地蝉鸣渐起，一碗凉面开胃迎夏。',
      en: 'It is the season of Lixia (Start of Summer) — cicadas begin to sing in Sichuan, and a bowl of chilled noodles opens the summer appetite.',
      ja: '立夏の時節。四川ではセミの声が聞こえ始め、冷やし麺で夏の訪れを味わいます。',
    },
  },
  {
    name: '小满', nameEn: 'Grain Buds (Xiaoman)', nameJa: '小満',
    month: 5, day: 21,
    banner: {
      zh: '小满时节，麦粒渐满，蜀乡新麦上桌，尝一口初夏的丰盈。',
      en: 'It is the season of Xiaoman (Grain Buds) — wheat kernels plump up, and new-wheat dishes bring the abundance of early summer to Shu tables.',
      ja: '小満の時節。麦の粒が満ちてくるころ、四川では新麦の味で初夏の実りを楽しみます。',
    },
  },
  {
    name: '芒种', nameEn: 'Grain in Ear (Mangzhong)', nameJa: '芒種',
    month: 6, day: 6,
    banner: {
      zh: '芒种时节，农忙插秧，端午粽香正浓，蜀人裹粽忙。',
      en: 'It is the season of Mangzhong (Grain in Ear) — amid the busy planting season, Dragon Boat Festival zongzi fill Sichuan kitchens with aroma.',
      ja: '芒種の時節。田植えで忙しい時期、四川では端午の粽(ちまき)の香りが漂います。',
    },
  },
  {
    name: '夏至', nameEn: 'Summer Solstice (Xiazhi)', nameJa: '夏至',
    month: 6, day: 21,
    banner: {
      zh: '夏至时节，日长至极，蜀人一碗凉面消暑，顺时而食。',
      en: 'It is the season of Xiazhi (Summer Solstice), the longest day of the year — Sichuanese cool off with a bowl of chilled noodles, eating with the seasons.',
      ja: '夏至の時節。一年で最も昼が長い日。四川では冷たい麺を食べて暑さをしのぎます。',
    },
  },
  {
    name: '小暑', nameEn: 'Minor Heat (Xiaoshu)', nameJa: '小暑',
    month: 7, day: 7,
    banner: {
      zh: '小暑时节，暑气渐盛，蜀地一碗冰粉，红糖醪糟透心凉。',
      en: 'It is the season of Xiaoshu (Minor Heat) — the heat builds up, and a bowl of Sichuan bingfen (ice jelly) with brown sugar and fermented rice brings instant cool.',
      ja: '小暑の時節。暑さが本格的に。四川では黒糖と酒麹のかかった氷粉(ビンフン)で涼を取ります。',
    },
  },
  {
    name: '大暑', nameEn: 'Major Heat (Dashu)', nameJa: '大暑',
    month: 7, day: 23,
    banner: {
      zh: '大暑时节，一年最热，蜀人偏要趁热打锅，火锅配冰粉越热越爽。',
      en: 'It is the season of Dashu (Major Heat), the hottest time of the year — Sichuanese fight fire with fire: bubbling hotpot paired with ice jelly.',
      ja: '大暑の時節。一年で最も暑い日。四川では火鍋と氷粉を合わせて、暑さを吹き飛ばします。',
    },
  },
  {
    name: '立秋', nameEn: 'Start of Autumn (Liqiu)', nameJa: '立秋',
    month: 8, day: 8,
    banner: {
      zh: '立秋时节，蜀人\u201c贴秋膘\u201d，炖一锅老鸭汤，滋润迎秋。',
      en: 'It is the season of Liqiu (Start of Autumn) — Sichuanese \u201cput on autumn fat\u201d with a slow-stewed old duck soup to greet the season.',
      ja: '立秋の時節。四川では鴨のスープをコトコト炊いて、秋バテに備えます。',
    },
  },
  {
    name: '处暑', nameEn: 'End of Heat (Chushu)', nameJa: '処暑',
    month: 8, day: 23,
    banner: {
      zh: '处暑时节，暑气渐消，蜀人炖盅银耳汤，润一润初秋的燥。',
      en: 'It is the season of Chushu (End of Heat) — the heat retreats, and Sichuanese simmer snow-fungus soup to ease the early-autumn dryness.',
      ja: '処暑の時節。暑さが和らぎ、四川では白キクラゲのスープで初秋の乾燥を潤します。',
    },
  },
  {
    name: '白露', nameEn: 'White Dew (Bailu)', nameJa: '白露',
    month: 9, day: 8,
    banner: {
      zh: '白露时节，露凝而白，蜀地桂花初绽，一盏桂花酒酿香满城。',
      en: 'It is the season of Bailu (White Dew) — dew turns white, osmanthus blossoms open across Shu, and sweet osmanthus rice wine perfumes the whole city.',
      ja: '白露の時節。露が白く凝るころ、四川ではキンモクセイの香りとともに桂花酒を楽しみます。',
    },
  },
  {
    name: '秋分', nameEn: 'Autumn Equinox (Qiufen)', nameJa: '秋分',
    month: 9, day: 23,
    banner: {
      zh: '秋分时节，昼夜再平分，恰逢中秋前后，蜀人分月饼话丰收。',
      en: 'It is the season of Qiufen (Autumn Equinox), around Mid-Autumn — Sichuanese share mooncakes and talk of the harvest.',
      ja: '秋分の時節。中秋節の頃にあたり、四川では月餅を分け合いながら豊作を語り合います。',
    },
  },
  {
    name: '寒露', nameEn: 'Cold Dew (Hanlu)', nameJa: '寒露',
    month: 10, day: 8,
    banner: {
      zh: '寒露时节，露寒将凝，蜀人登高赏菊，一盏菊花茶暖手暖心。',
      en: 'It is the season of Hanlu (Cold Dew) — dew grows cold; Sichuanese climb heights to admire chrysanthemums over a warm cup of chrysanthemum tea.',
      ja: '寒露の時節。露が冷たくなるころ、四川では菊の茶を飲みながら登高を楽しみます。',
    },
  },
  {
    name: '霜降', nameEn: "Frost's Descent (Shuangjiang)", nameJa: '霜降',
    month: 10, day: 23,
    banner: {
      zh: '霜降时节，蜀人讲究进补，\u201c补冬不如补霜降\u201d，煨汤炖肉正当时。',
      en: 'It is the season of Shuangjiang (Frost\u2019s Descent) — Sichuan folks start their tonics early: as the saying goes, \u201cnourishing now beats nourishing in winter,\u201d with stews on the stove.',
      ja: '霜降の時節。四川では「冬の補養より霜降の補養」といわれ、煮込み料理で栄養を付けます。',
    },
  },
  {
    name: '立冬', nameEn: 'Start of Winter (Lidong)', nameJa: '立冬',
    month: 11, day: 7,
    banner: {
      zh: '立冬时节，冬藏之始，蜀人羊肉汤锅渐旺，先暖为敬。',
      en: 'It is the season of Lidong (Start of Winter) — mutton soup pots begin bubbling across Sichuan; warmth comes first.',
      ja: '立冬の時節。冬を藏する始まり。四川では羊肉鍋が店に並び始め、あたたかさで冬を迎えます。',
    },
  },
  {
    name: '小雪', nameEn: 'Minor Snow (Xiaoxue)', nameJa: '小雪',
    month: 11, day: 22,
    banner: {
      zh: '小雪时节，蜀家挂腊味，香肠腊肉上房梁，腌出岁末的年味。',
      en: 'It is the season of Xiaoxue (Minor Snow) — Sichuan households hang sausages and cured pork up to air, salting in the flavor of the year\u2019s end.',
      ja: '小雪の時節。四川の家々では腸詰や腊肉(干し肉)を軒につるし、年末の味を漬け込みます。',
    },
  },
  {
    name: '大雪', nameEn: 'Major Snow (Daxue)', nameJa: '大雪',
    month: 12, day: 7,
    banner: {
      zh: '大雪时节，腊味正香，简阳羊肉汤头锅开涮，越冷越鲜。',
      en: 'It is the season of Daxue (Major Snow) — cured meats ripen, and the first steaming pots of Jianyang mutton soup are served; the colder it gets, the tastier.',
      ja: '大雪の時節。腊味の香りがたち、簡陽の羊肉スープの最初の鍋が提供され始めます。寒いほど美味しい。',
    },
  },
  DONGZHI_TERM,
]

/**
 * 取指定日期所处的节气(近似表,容差 ±1 天)。
 * 规则:取年内日期表中"最后一个已到"的节气;1 月 6 日(小寒)之前属上一年的冬至,
 * 三语名称与文案与冬至一致,直接回落冬至条目。
 * 同步返回 rangeText:本节气管辖区间(交节日—下一交节日前一天)的三语文本,
 * 供前端展示"白露 · 9月7日—9月22日"式期间标注,避免"今日X"的歧义。
 */
export function getCurrentSolarTerm(date: Date = new Date()): SolarTermInfo {
  const md = (date.getMonth() + 1) * 100 + date.getDate()
  let found: SolarTermInfo | undefined
  let next: SolarTermInfo | undefined
  for (let i = SOLAR_TERMS.length - 1; i >= 0; i--) {
    const t = SOLAR_TERMS[i]
    if (t && md >= t.month * 100 + t.day) {
      found = t
      next = SOLAR_TERMS[i + 1] ?? SOLAR_TERMS[0]
      break
    }
  }
  if (!found || !next) {
    // 年初未到小寒 → 上一年的冬至,区间终点为年内小寒交节日前一天
    found = DONGZHI_TERM
    next = SOLAR_TERMS[0] ?? DONGZHI_TERM
  }

  // 跨年区间(冬至→次年小寒):终点落到次年;起点若在 1 月则属于上一年交节
  const crossesYear = next.month < found.month
  const start = new Date(date.getFullYear() + (crossesYear && date.getMonth() === 0 ? -1 : 0), found.month - 1, found.day)
  const end = new Date(date.getFullYear() + (crossesYear ? 1 : 0), next.month - 1, next.day)
  end.setDate(end.getDate() - 1)

  const mEn = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const rangeText = {
    zh: `${start.getMonth() + 1}月${start.getDate()}日—${end.getMonth() + 1}月${end.getDate()}日`,
    ja: `${start.getMonth() + 1}月${start.getDate()}日〜${end.getMonth() + 1}月${end.getDate()}日`,
    en: `${mEn[start.getMonth()]} ${start.getDate()} – ${mEn[end.getMonth()]} ${end.getDate()}`,
  }
  return { ...found, rangeText }
}
