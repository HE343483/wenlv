<script setup lang="ts">
/**
 * FoodDetail.vue — 美食详情页
 * 说明：当前为纯布局骨架，基础名称/描述复用本地 i18n（food.card.*）渲染版式；
 *       后期由后端接口按路由参数 :id 拉取美食详情数据填充（见下方「后端接口接入预留区」）。
 */
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import AppIcon from '@/components/AppIcon.vue'

const route = useRoute()
const router = useRouter()
const langStore = useLanguageStore()

/* 路由参数：美食分类ID（hotpot / chuanchuan / cuisine / snacks / tea / nightfood） */
const foodId = computed(() => String(route.params.id ?? ''))

function goBack() {
  router.back()
}

/* 占位名称 — 复用 i18n 本地文案，后期替换为接口返回的名称 */
const placeholderName = computed(() => {
  if (langStore.lang === 'zh') {
    return langStore.t(`food.card.${foodId.value}.name`)
  }
  return langStore.t(`food.card.${foodId.value}.en`)
})

const placeholderEnTitle = computed(() => langStore.t(`food.card.${foodId.value}.en`))

const placeholderDesc = computed(() => langStore.t(`food.card.${foodId.value}.desc`))

/* 占位标签 — 后期替换为接口返回的 tags */
const placeholderTags = ['麻辣', '鲜香', '经典']
const enTags = ['Numbing', 'Fragrant', 'Classic']
const tagIndex = computed(() =>
  Math.abs([...foodId.value].reduce((acc, ch) => acc + ch.charCodeAt(0), 0)) % placeholderTags.length
)

/* ════════════════════════════════════════════════
 *  本地 Mock 详情数据（中英文双语）
 *  说明：接入后端接口前，用本地文案完整渲染版式；
 *        接入后端后，用 fetchFoodDetail 返回的数据替换 detail 即可。
 *  ════════════════════════════════════════════════ */
interface FoodLocalData {
  rating: string
  flavor: string
  spice: string
  price: string
  signature: string
  scene: string
  storyTitle: string
  paras: string[]
  related: Array<{ name: string; desc: string }>
}

const localDetails: Record<string, { zh: FoodLocalData; en: FoodLocalData }> = {
  hotpot: {
    zh: {
      rating: '4.9',
      flavor: '麻辣鲜香',
      spice: '重辣',
      price: '人均 ¥80',
      signature: '牛油锅底 · 毛肚 · 鸭肠',
      scene: '朋友聚餐 / 深夜食堂',
      storyTitle: '一锅红汤，煮尽江湖',
      paras: [
        '成都火锅的精髓在那口牛油锅底——辣椒、花椒与多种香料在滚油中交融，麻辣鲜香层层递进。',
        '毛肚七上八下、鸭肠三提三放，讲究的是火候与节奏。围炉而坐，热气蒸腾中尽是人间烟火。',
      ],
      related: [
        { name: '串串香', desc: '一根竹签串起百味，边走边吃更随性。' },
        { name: '夜宵', desc: '火锅之后的小龙虾配啤酒，巴适得板。' },
        { name: '盖碗茶', desc: '麻辣过后，一盏清茶解腻回甘。' },
      ],
    },
    en: {
      rating: '4.9',
      flavor: 'Numbing & spicy',
      spice: 'Extra spicy',
      price: '¥80 / person',
      signature: 'Beef-tallow base · Beef tripe · Duck intestine',
      scene: 'Friends gathering / Late-night',
      storyTitle: 'A Pot of Red Broth, a World of Flavor',
      paras: [
        'The soul of Chengdu hotpot lies in its beef-tallow base, where chilis, Sichuan peppercorns and spices bloom into layers of numbing heat.',
        'Beef tripe is dipped seven up and eight down, duck intestine three lifts — it is all about timing. Around the steaming pot, every gathering turns into the warmth of everyday life.',
      ],
      related: [
        { name: 'Chuan Chuan', desc: 'A hundred flavors on skewers, best enjoyed on the go.' },
        { name: 'Night Food', desc: 'Crayfish and beer after hotpot — life is good.' },
        { name: 'Gaiwan Tea', desc: 'A cup of tea to settle the spice and refresh the palate.' },
      ],
    },
  },

  chuanchuan: {
    zh: {
      rating: '4.7',
      flavor: '百味百串',
      spice: '中辣',
      price: '人均 ¥50',
      signature: '红油锅底 · 牛肉串 · 郡肝',
      scene: '一人食 / 三五好友',
      storyTitle: '一根竹签，串起百味',
      paras: [
        '串串香把成都的随性发挥到极致——想吃什么拿什么，一根竹签串起牛肉、郡肝与素菜，浸入红油翻滚。',
        '按签计费、按签结账，吃多少拿多少，是成都人最接地气的聚餐方式。',
      ],
      related: [
        { name: '火锅', desc: '同是红汤，围炉而坐更有仪式感。' },
        { name: '名小吃', desc: '串串之外，蛋烘糕、糖油果子也值得一试。' },
        { name: '夜宵', desc: '串串配冰粉，深夜的成都格外动人。' },
      ],
    },
    en: {
      rating: '4.7',
      flavor: 'Hundred flavors on sticks',
      spice: 'Medium spicy',
      price: '¥50 / person',
      signature: 'Red-oil base · Beef skewers · Gizzard',
      scene: 'Solo dining / Small group',
      storyTitle: 'A Skewer of a Hundred Flavors',
      paras: [
        'Chuan Chuan takes Chengdu\'s laid-back spirit to the extreme — pick whatever you like, skewered beef, gizzard and greens, and dunk them into bubbling red oil.',
        'Billed by the skewer, you take only what you eat. It is the most down-to-earth way Chengdu people gather for a meal.',
      ],
      related: [
        { name: 'Hotpot', desc: 'The same red broth, but a more ceremonial sit-down gathering.' },
        { name: 'Snacks', desc: 'Beyond skewers, try the egg-baked cakes and sugar-glazed fruit.' },
        { name: 'Night Food', desc: 'Skewers with iced jelly — Chengdu nights are especially lovely.' },
      ],
    },
  },

  cuisine: {
    zh: {
      rating: '4.8',
      flavor: '一菜一格，百菜百味',
      spice: '中辣',
      price: '人均 ¥70',
      signature: '麻婆豆腐 · 回锅肉 · 宫保鸡丁',
      scene: '家宴 / 正餐',
      storyTitle: '一菜一格，百菜百味',
      paras: [
        '川菜素有「一菜一格，百菜百味」之说，麻婆豆腐的麻、回锅肉的香、宫保鸡丁的酸甜，各成风味。',
        '郫县豆瓣是川菜之魂，经过发酵的红亮豆瓣为每一道菜打上巴蜀的底色。',
      ],
      related: [
        { name: '盖碗茶', desc: '正餐之后，一盏茶解腻助消化。' },
        { name: '名小吃', desc: '正餐之外的街头小食，同样精彩。' },
        { name: '火锅', desc: '想更热闹一点，就转战一锅红汤。' },
      ],
    },
    en: {
      rating: '4.8',
      flavor: 'One dish, one style',
      spice: 'Medium spicy',
      price: '¥70 / person',
      signature: 'Mapo tofu · Twice-cooked pork · Kung Pao chicken',
      scene: 'Family dinner / Formal meal',
      storyTitle: 'One Dish, One Style',
      paras: [
        'Sichuan cuisine is known for "one dish, one style; a hundred dishes, a hundred flavors" — the numbing mapo tofu, the fragrant twice-cooked pork and the sweet-sour Kung Pao chicken each stand on their own.',
        'Pixian douban (fermented broad-bean paste) is the soul of Sichuan cooking, lending every plate its signature red base.',
      ],
      related: [
        { name: 'Gaiwan Tea', desc: 'A cup of tea after the meal to cut the richness.' },
        { name: 'Snacks', desc: 'Street snacks beyond the main course are equally delightful.' },
        { name: 'Hotpot', desc: 'For something livelier, move on to a pot of red broth.' },
      ],
    },
  },

  snacks: {
    zh: {
      rating: '4.6',
      flavor: '甜咸交融',
      spice: '不辣',
      price: '人均 ¥30',
      signature: '蛋烘糕 · 糖油果子 · 冰粉',
      scene: '街头逛吃 / 下午茶',
      storyTitle: '街头烟火，甜咸之间',
      paras: [
        '成都的名小吃藏在街巷之间，蛋烘糕的软糯、糖油果子的焦香、冰粉的清甜，撑起了成都人的下午茶。',
        '甜咸交织的口味，正像这座城市——包容而鲜活。',
      ],
      related: [
        { name: '盖碗茶', desc: '小吃配茶，是老成都的经典组合。' },
        { name: '川菜', desc: '逛累了，再坐下吃一顿正餐。' },
        { name: '夜宵', desc: '白天小吃，夜晚烧烤，各有各的香。' },
      ],
    },
    en: {
      rating: '4.6',
      flavor: 'Sweet meets savory',
      spice: 'Not spicy',
      price: '¥30 / person',
      signature: 'Egg-baked cake · Sugar-glazed fruit · Iced jelly',
      scene: 'Street food crawl / Afternoon tea',
      storyTitle: 'Street Flavor, Between Sweet and Savory',
      paras: [
        'Chengdu\'s famous snacks hide in its alleys — the soft egg-baked cake, the caramelized sugar-glazed fruit and the refreshing iced jelly make up the city\'s afternoon tea.',
        'The mix of sweet and savory mirrors the city itself: inclusive and full of life.',
      ],
      related: [
        { name: 'Gaiwan Tea', desc: 'Snacks with tea — a classic old-Chengdu pairing.' },
        { name: 'Sichuan Cuisine', desc: 'When you tire of walking, sit down for a proper meal.' },
        { name: 'Night Food', desc: 'Street snacks by day, barbecue by night.' },
      ],
    },
  },

  tea: {
    zh: {
      rating: '4.5',
      flavor: '清雅回甘',
      spice: '不辣',
      price: '人均 ¥40',
      signature: '盖碗茶 · 茉莉花茶 · 竹叶青',
      scene: '茶馆消磨 / 会友',
      storyTitle: '一盏盖碗，半日闲',
      paras: [
        '盖碗茶是成都慢生活的注脚，三件套的盖、碗、托，一冲一泡之间，尽显从容。',
        '茶馆里的川剧变脸与评书，让一盏茶的时间也变得热闹。',
      ],
      related: [
        { name: '名小吃', desc: '茶点配小吃，闲坐一整个下午。' },
        { name: '川菜', desc: '品茶之后，再来一桌地道的川味。' },
        { name: '火锅', desc: '清茶过后，也能奔赴一场红汤。' },
      ],
    },
    en: {
      rating: '4.5',
      flavor: 'Elegant, with a sweet finish',
      spice: 'Not spicy',
      price: '¥40 / person',
      signature: 'Gaiwan tea · Jasmine tea · Zhuyeqing',
      scene: 'Teahouse leisure / Meeting friends',
      storyTitle: 'A Cup of Gaiwan, an Idle Afternoon',
      paras: [
        'Gaiwan tea is a footnote to Chengdu\'s slow-paced life. The three-piece set of lid, bowl and saucer brings composure to every pour.',
        'With Sichuan opera face-changing and storytelling, even a single cup of tea becomes lively.',
      ],
      related: [
        { name: 'Snacks', desc: 'Tea snacks and light bites — idle away the whole afternoon.' },
        { name: 'Sichuan Cuisine', desc: 'After tea, a table of authentic Sichuan flavor.' },
        { name: 'Hotpot', desc: 'After the clear tea, you can still rush into a pot of red broth.' },
      ],
    },
  },

  nightfood: {
    zh: {
      rating: '4.8',
      flavor: '烟火气十足',
      spice: '中辣',
      price: '人均 ¥60',
      signature: '小龙虾 · 烧烤 · 冷啖杯',
      scene: '深夜食堂 / 宵夜',
      storyTitle: '夜幕之下，越夜越香',
      paras: [
        '入夜后的成都，九眼桥灯火通明，小龙虾、烧烤与冷啖杯是夜宵桌上的主角。',
        '配上一瓶冰啤酒，谈天说地，是成都人一天的完美收尾。',
      ],
      related: [
        { name: '火锅', desc: '宵夜之前，一锅红汤更尽兴。' },
        { name: '串串香', desc: '夜市的串串，也是宵夜的绝配。' },
        { name: '盖碗茶', desc: '热闹过后，一盏茶慢慢收尾。' },
      ],
    },
    en: {
      rating: '4.8',
      flavor: 'Full of street-side warmth',
      spice: 'Medium spicy',
      price: '¥60 / person',
      signature: 'Crayfish · Barbecue · Cold dishes',
      scene: 'Late-night dining / Supper',
      storyTitle: 'The Later the Night, the Better It Smells',
      paras: [
        'After dark, Jiuyan Bridge lights up and crayfish, barbecue and cold dishes take center stage on the supper table.',
        'Washed down with an iced beer, it is the perfect ending to a Chengdu day.',
      ],
      related: [
        { name: 'Hotpot', desc: 'Before the late-night bite, a pot of red broth makes it livelier.' },
        { name: 'Chuan Chuan', desc: 'Night-market skewers also pair perfectly with supper.' },
        { name: 'Gaiwan Tea', desc: 'After the buzz, wind down with a cup of tea.' },
      ],
    },
  },
}

/* 当前详情数据：按语言返回本地 Mock；接入后端后替换为接口返回 */
const detail = computed<FoodLocalData>(() => {
  const entry = localDetails[foodId.value] ?? localDetails.hotpot
  if (!entry) {
    return langStore.lang === 'zh' ? localDetails.hotpot!.zh : localDetails.hotpot!.en
  }
  return langStore.lang === 'zh' ? entry.zh : entry.en
})

/* ════════════════════════════════════════════════
 *  后端接口接入预留区
 *  ════════════════════════════════════════════════
 *  建议接口：GET /api/food/:id
 *  返回字段参考下方 FoodDetailData 接口定义。
 *  接入后端后：将下方注释的 ref / fetchFoodDetail / onMounted 打开，
 *  并用 foodDetail.value 的数据替换模板中的占位内容即可。
 *  ════════════════════════════════════════════════ */
// interface FoodDetailData {
//   id: string
//   nameZh: string
//   nameEn: string
//   coverImage: string
//   tags: string[]
//   rating: number
//   avgPrice: string
//   spiceLevel: string
//   recommendScene: string
//   signature: string
//   descriptionZh: string
//   descriptionEn: string
//   gallery: string[]
//   location: { lng: number; lat: number }
//   related: Array<{ id: string; nameZh: string; nameEn: string; descZh: string; descEn: string }>
// }

// import { onMounted, ref } from 'vue'
// const foodDetail = ref<FoodDetailData | null>(null)
// const loading = ref(false)
// const error = ref('')
// async function fetchFoodDetail(id: string): Promise<FoodDetailData> {
//   const res = await fetch(`/api/food/${id}`)
//   if (!res.ok) throw new Error(`HTTP ${res.status}`)
//   return res.json()
// }
// onMounted(async () => {
//   loading.value = true
//   try {
//     foodDetail.value = await fetchFoodDetail(foodId.value)
//   } catch (e) {
//     error.value = e instanceof Error ? e.message : String(e)
//   } finally {
//     loading.value = false
//   }
// })
</script>

<template>
  <div class="detail-page">
    <!-- ──── 顶栏 ──── -->
    <header class="detail-topbar">
      <button class="detail-topbar__back" @click="goBack" :title="langStore.t('common.back')">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M15 18l-6-6 6-6"/>
        </svg>
      </button>

      <div class="detail-topbar__brand">
        <span class="detail-topbar__title">蜀韵·成都</span>
        <span class="detail-topbar__crumb">{{ langStore.t('foodDetail.breadcrumb') }}</span>
      </div>

      <div class="detail-topbar__actions">
        <button class="detail-topbar__lang-btn" @click="langStore.toggle()">
          {{ langStore.t('nav.langSwitch') }}
        </button>
      </div>
    </header>

    <main class="detail-main">
      <!-- ──── HERO — 图片区（占位）──── -->
      <section class="detail-hero">
        <div class="detail-hero__media">
          <!-- 图片占位：后期替换为接口返回的美食封面图 -->
          <div class="detail-hero__placeholder">
            <div class="detail-hero__shu" aria-hidden="true">味</div>
            <span class="detail-hero__api-badge">{{ langStore.t('foodDetail.imagePlaceholder') }}</span>
          </div>

          <!-- 名称浮层 -->
          <div class="detail-hero__overlay">
            <h1 class="detail-hero__title">{{ placeholderName }}</h1>
            <p class="detail-hero__en-title">{{ placeholderEnTitle }}</p>
            <div class="detail-hero__tags">
              <span
                v-for="(t, i) in (langStore.lang === 'zh' ? placeholderTags : enTags)"
                :key="t"
                class="detail-hero__tag"
                :class="{ 'detail-hero__tag--accent': i === tagIndex }"
              >
                {{ t }}
              </span>
            </div>
          </div>
        </div>
      </section>

      <!-- ──── 概要：评分 + 简介 ──── -->
      <section class="detail-section container">
        <div class="detail-summary">
          <!-- 左：评分面板 -->
          <aside class="detail-summary__aside">
            <div class="detail-score">
              <span class="detail-score__num">{{ detail.rating }}</span>
              <div class="detail-score__meta">
                <span class="detail-score__stars">
                  <AppIcon v-for="i in 5" :key="i" name="star" :size="16" />
                </span>
                <span class="detail-score__label">{{ langStore.t('foodDetail.rating') }}</span>
              </div>
            </div>
            <div class="detail-aside__row">
              <span class="detail-aside__key">{{ langStore.t('foodDetail.flavor') }}</span>
              <span class="detail-aside__value">{{ detail.flavor }}</span>
            </div>
            <div class="detail-aside__row">
              <span class="detail-aside__key">{{ langStore.t('foodDetail.spiceLevel') }}</span>
              <span class="detail-aside__value">{{ detail.spice }}</span>
            </div>
            <div class="detail-aside__row">
              <span class="detail-aside__key">{{ langStore.t('foodDetail.avgPrice') }}</span>
              <span class="detail-aside__value">{{ detail.price }}</span>
            </div>
          </aside>

          <!-- 右：简介 -->
          <div class="detail-summary__body">
            <div class="detail-summary__eyebrow">
              <span>◈</span>
              {{ langStore.t('foodDetail.overview') }}
              <span>◈</span>
            </div>
            <p class="detail-summary__text">{{ placeholderDesc }}</p>
          </div>
        </div>
      </section>

      <!-- ──── 实用信息 ──── -->
      <section class="detail-section container">
        <div class="detail-info">
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="star" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('foodDetail.signature') }}</h3>
            <p class="detail-info__value">{{ detail.signature }}</p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="fire" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('foodDetail.spiceLevel') }}</h3>
            <p class="detail-info__value">{{ detail.spice }}</p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="ticket" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('foodDetail.avgPrice') }}</h3>
            <p class="detail-info__value">{{ detail.price }}</p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="clock" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('foodDetail.recommendScene') }}</h3>
            <p class="detail-info__value">{{ detail.scene }}</p>
          </div>
        </div>
      </section>

      <!-- ──── 风味故事（图文详情）──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.detailTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('foodDetail.detailSubtitle') }}</p>
        </header>

        <div class="detail-content">
          <!-- 配图占位块 — 后期替换为后端富文本/段落+配图 -->
          <div class="detail-content__row">
            <div class="detail-content__figure">
              <div class="detail-content__img detail-content__img--empty">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
                  <rect x="3" y="3" width="18" height="18" rx="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <path d="M21 15l-5-5L5 21"/>
                </svg>
              </div>
            </div>
            <div class="detail-content__text">
              <h3 class="detail-content__caption">{{ detail.storyTitle }}</h3>
              <p v-for="p in detail.paras" :key="p" class="detail-content__para">{{ p }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- ──── 美味相册 ──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.galleryTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('foodDetail.gallerySubtitle') }}</p>
        </header>

        <div class="detail-gallery">
          <div v-for="n in 6" :key="n" class="detail-gallery__item">
            <div class="detail-gallery__img">
              <span class="detail-gallery__num">{{ String(n).padStart(2, '0') }}</span>
            </div>
          </div>
        </div>
      </section>


      <!-- ──── 寻味地图（占位）──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.mapTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('foodDetail.mapSubtitle') }}</p>
        </header>

        <div class="detail-map">
          <div class="detail-map__placeholder">
            <span class="detail-map__pin"><AppIcon name="pin" :size="40" /></span>
            <span class="detail-map__hint">{{ langStore.t('foodDetail.mapPlaceholder') }}</span>
          </div>
        </div>
      </section>

      <!-- ──── 相关推荐 ──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.aroundTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('foodDetail.aroundSubtitle') }}</p>
        </header>

        <div class="detail-around">
          <div v-for="r in detail.related" :key="r.name" class="detail-around__card">
            <div class="detail-around__img">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
                <rect x="3" y="3" width="18" height="18" rx="2"/>
                <circle cx="8.5" cy="8.5" r="1.5"/>
                <path d="M21 15l-5-5L5 21"/>
              </svg>
            </div>
            <div class="detail-around__body">
              <h3 class="detail-around__name">{{ r.name }}</h3>
              <p class="detail-around__desc">{{ r.desc }}</p>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
}

/* ========================================
   顶栏
   ======================================== */
.detail-topbar {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--nav-height);
  padding: 0 var(--space-6);
  background: var(--color-nav-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--color-border);
}

.detail-topbar__back {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
}

.detail-topbar__back:hover {
  background: var(--color-surface-hover);
  color: var(--color-gold);
}

.detail-topbar__brand {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.detail-topbar__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
}

.detail-topbar__crumb {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-widest);
  padding-left: var(--space-3);
  border-left: 1px solid var(--color-border-light);
}

.detail-topbar__actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.detail-topbar__lang-btn {
  font-family: var(--font-display);
  font-size: var(--text-sm);
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.detail-topbar__lang-btn:hover {
  border-color: var(--color-gold);
  color: var(--color-gold);
}

/* ========================================
   主体
   ======================================== */
.detail-main {
  flex: 1;
}

.detail-section {
  padding: var(--space-12) 0;
}

/* ========================================
   HERO — 图片占位
   ======================================== */
.detail-hero__media {
  position: relative;
  width: 100%;
  height: clamp(320px, 46vh, 480px);
  overflow: hidden;
}

.detail-hero__placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  background:
    radial-gradient(ellipse 70% 60% at 50% 35%, color-mix(in srgb, var(--color-cinnabar) 10%, transparent) 0%, transparent 70%),
    linear-gradient(180deg, var(--color-surface-elevated), var(--color-bg-alt));
  border-bottom: 1px solid var(--color-border);
}

.detail-hero__shu {
  font-family: var(--font-display);
  font-size: clamp(120px, 22vw, 240px);
  font-weight: 900;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1px color-mix(in srgb, var(--color-cinnabar) 40%, transparent);
  user-select: none;
  pointer-events: none;
}

.detail-hero__api-badge {
  font-size: var(--text-xs);
  padding: 4px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-widest);
  background: color-mix(in srgb, var(--color-bg) 60%, transparent);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
}

.detail-hero__overlay {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: var(--space-8) var(--space-6);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  text-align: center;
  background: linear-gradient(180deg, transparent 0%, color-mix(in srgb, var(--color-bg) 75%, transparent) 100%);
}

.detail-hero__title {
  font-family: var(--font-display);
  font-size: clamp(var(--text-3xl), 5vw, var(--text-5xl));
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.15;
  text-shadow: 0 2px 16px color-mix(in srgb, var(--color-bg) 60%, transparent);
}

.detail-hero__en-title {
  font-family: var(--font-en-display);
  font-style: italic;
  font-size: var(--text-base);
  color: var(--color-cinnabar);
  letter-spacing: var(--tracking-wider);
}

.detail-hero__tags {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: var(--space-2);
  margin-top: var(--space-1);
}

.detail-hero__tag {
  font-size: var(--text-xs);
  padding: 4px 14px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border-light);
  color: var(--color-text-secondary);
  background: color-mix(in srgb, var(--color-bg) 45%, transparent);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  letter-spacing: var(--tracking-wide);
}

.detail-hero__tag--accent {
  border-color: color-mix(in srgb, var(--color-cinnabar) 55%, transparent);
  color: var(--color-cinnabar);
}

/* ========================================
   概要 — 评分 + 简介
   ======================================== */
.detail-summary {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: var(--space-10);
  align-items: start;
}

.detail-summary__aside {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.detail-score {
  display: flex;
  align-items: center;
  gap: var(--space-5);
  padding: var(--space-6);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.detail-score__num {
  font-family: var(--font-en-display);
  font-size: var(--text-4xl);
  font-weight: 700;
  color: var(--color-gold);
  line-height: 1;
}

.detail-score__meta {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.detail-score__stars {
  display: flex;
  align-items: center;
  gap: 2px;
  color: var(--color-gold);
  font-size: var(--text-sm);
  opacity: 0.55;
}

.detail-score__label {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.detail-aside__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.detail-aside__key {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.detail-aside__value {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
}

.detail-summary__body {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.detail-summary__eyebrow {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--text-xs);
  color: var(--color-cinnabar);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.detail-summary__text {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  letter-spacing: var(--tracking-wide);
  font-weight: 400;
}

/* ========================================
   实用信息
   ======================================== */
.detail-info {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
}

.detail-info__card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-2);
  padding: var(--space-6);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: border-color var(--transition-base), transform var(--transition-base);
}

.detail-info__card:hover {
  border-color: var(--color-cinnabar);
  transform: translateY(-2px);
}

.detail-info__icon {
  display: flex;
  align-items: center;
  color: var(--color-cinnabar);
  margin-bottom: var(--space-1);
}

.detail-info__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.detail-info__value {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   区块标题
   ======================================== */
.detail-block-head {
  text-align: center;
  margin-bottom: var(--space-8);
}

/* ========================================
   图文详情
   ======================================== */
.detail-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-12);
}

.detail-content__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-10);
  align-items: center;
}

.detail-content__figure {
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.detail-content__img {
  aspect-ratio: 4 / 3;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  background:
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-cinnabar) 8%, transparent) 0%, transparent 70%),
    var(--color-bg-alt);
}

.detail-content__img--empty {
  opacity: 0.7;
}

.detail-content__text {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.detail-content__caption {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.detail-content__para {
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

/* ========================================
   相册
   ======================================== */
.detail-gallery {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-4);
}

.detail-gallery__item {
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--color-border);
  transition: border-color var(--transition-base), transform var(--transition-base);
}

.detail-gallery__item:hover {
  border-color: var(--color-cinnabar);
  transform: translateY(-2px);
}

.detail-gallery__img {
  aspect-ratio: 4 / 3;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-cinnabar) 8%, transparent) 0%, transparent 70%),
    var(--color-bg-alt);
}

.detail-gallery__num {
  font-family: var(--font-en-display);
  font-size: var(--text-3xl);
  font-weight: 700;
  color: transparent;
  -webkit-text-stroke: 1px color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
  user-select: none;
}

/* ========================================
   地图占位
   ======================================== */
.detail-map {
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.detail-map__placeholder {
  aspect-ratio: 16 / 6;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  background:
    radial-gradient(ellipse 50% 70% at 50% 50%, color-mix(in srgb, var(--color-cinnabar) 6%, transparent) 0%, transparent 70%),
    repeating-linear-gradient(0deg, transparent 0 39px, color-mix(in srgb, var(--color-border) 45%, transparent) 39px 40px),
    repeating-linear-gradient(90deg, transparent 0 39px, color-mix(in srgb, var(--color-border) 45%, transparent) 39px 40px),
    var(--color-bg-alt);
}

.detail-map__pin {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-cinnabar);
  filter: drop-shadow(0 0 16px var(--color-cinnabar-dim));
  animation: pin-float 2.6s ease-in-out infinite;
}

@keyframes pin-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

.detail-map__hint {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-widest);
  background: color-mix(in srgb, var(--color-bg) 60%, transparent);
  padding: 4px 12px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
}

/* ========================================
   相关推荐
   ======================================== */
.detail-around {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.detail-around__card {
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all var(--transition-base);
}

.detail-around__card:hover {
  border-color: var(--color-cinnabar);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px var(--color-cinnabar-dim), var(--shadow-lg);
}

.detail-around__img {
  aspect-ratio: 16 / 9;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  background:
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-cinnabar) 8%, transparent) 0%, transparent 70%),
    var(--color-bg-alt);
  opacity: 0.8;
}

.detail-around__body {
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.detail-around__name {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.detail-around__desc {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  line-height: var(--leading-relaxed);
}

/* ========================================
   Responsive
   ======================================== */
@media (max-width: 1024px) {
  .detail-info {
    grid-template-columns: repeat(2, 1fr);
  }
  .detail-gallery {
    grid-template-columns: repeat(2, 1fr);
  }
  .detail-around {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .detail-topbar__crumb {
    display: none;
  }
  .detail-topbar {
    padding: 0 var(--space-4);
  }
  .detail-section {
    padding: var(--space-8) 0;
  }
  .detail-summary {
    grid-template-columns: 1fr;
    gap: var(--space-6);
  }
  .detail-summary__text {
    font-size: var(--text-lg);
  }
  .detail-content__row {
    grid-template-columns: 1fr;
    gap: var(--space-6);
  }
}

@media (max-width: 640px) {
  .detail-info,
  .detail-gallery,
  .detail-around {
    grid-template-columns: 1fr;
  }
  .detail-hero__overlay {
    padding: var(--space-6) var(--space-4);
  }
}
</style>
