<script setup lang="ts">
/**
 * FoodDetail.vue — 美食详情页
 * 数字 id(如 /food/3)走 /api/food/:id 真实数据渲染;
 * 非数字 id(美食大类 cuisine/snacks/nightfood)走 /api/food-category/:key 渲染类别页:
 * 类别概述 + 图文段落 + 类别图集 + 「本类美食」(跳转真实菜品);
 * 类别接口失败/无数据时不渲染类别专属区块,先尝试按"类别键→代表菜名"解析成菜品 id 跳转
 * (如 /food/hotpot → /food/1),解析不到则回退 i18n 文案 + 本类美食列表的最小降级视图,保证不白屏。
 */
import { computed, ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { getFood, listFoods, getFoodCategory } from '@/api/content'
import type { FoodItem, FoodCategoryItem } from '@/api/content'
import { parseSections, estimatedSet, splitList, displayFact } from '@/utils/scenicDetail'
import { pickDesc, pickCultureNote, pickName } from '@/utils/storyI18n'
import { isSpeaking as voiceSpeaking, speak, stopSpeaking, hasVoiceFor } from '@/utils/speech'
import dictZh from '@/locales/zh'
import dictEn from '@/locales/en'
import dictJa from '@/locales/ja'
import { getRuntimeMapJsKey } from '@/api/trip'
import AMapLoader from '@amap/amap-jsapi-loader'
import AppIcon from '@/components/AppIcon.vue'

const route = useRoute()
const router = useRouter()
const langStore = useLanguageStore()

/* 路由参数：数字 id → 数据库美食；分类 key(hotpot 等) → i18n 降级 */
const foodId = computed(() => String(route.params.id ?? ''))
const numericId = computed(() => {
  const n = Number(foodId.value)
  return Number.isFinite(n) && n > 0 ? n : 0
})
const isFallback = computed(() => numericId.value === 0)

/* 美食名片分类 ↔ 菜品标签关键词映射(按标签"包含"匹配,见 matchCategoryDishes) */
const CATEGORY_TAGS: Record<string, string[]> = {
  hotpot: ['火锅'],
  chuanchuan: ['串串'],
  cuisine: ['川菜'],
  snacks: ['小吃'],
  tea: ['茶'],
  nightfood: ['夜宵'],
}

/**
 * 类别键 → 代表菜名(name_zh)。
 * 这些键在 food_categories 表没有记录(类别接口 404),但美食页确有对应菜品,
 * 直接访问 /food/hotpot 时应作为菜品详情页处理,按代表菜名解析出真实菜品 id。
 * 代表菜名与 FoodPage.vue 的 foodCards.nameZh 保持一致。
 */
const CATEGORY_DISH_NAMES: Record<string, string> = {
  hotpot: '火锅',
  chuanchuan: '串串香',
  tea: '盖碗茶',
}

/** 按标签关键词从美食列表中筛出该分类的菜品,保留接口返回的原始顺序 */
function matchCategoryDishes(items: FoodItem[], keywords: string[]): FoodItem[] {
  if (!keywords.length) return []
  return items.filter((item) =>
    splitList(item.tags).some((tag) => keywords.some((kw) => tag.includes(kw))),
  )
}

function goBack() {
  router.back()
}

/* ════════════════════════════════════════════════
 *  详情数据：数字 id 走 /api/food/:id；分类 key 走 i18n 降级
 *  展示层纯函数(parseSections/splitList/estimatedSet/displayFact)复用 @/utils/scenicDetail
 *  ════════════════════════════════════════════════ */
/* ── 详情数据：数字 id 走接口，分类 key 走 i18n 降级 ── */
const food = ref<FoodItem | null>(null)
const imgFailed = ref(false)
const related = ref<FoodItem[]>([])

const noData = computed(() => langStore.t('foodDetail.noData'))
const estimated = computed(() => estimatedSet(food.value?.estimated_fields))
/** 仅在字段有值且被登记为参考值时才标注，避免给空值加徽标 */
function showEst(field: string, value?: string): boolean {
  return !!value && estimated.value.has(field)
}

/* 分类 key 降级分支：直接渲染 i18n 文案(food.card.*) */
const fallbackNameZh = computed(() => langStore.t(`food.card.${foodId.value}.name`))
const fallbackNameEn = computed(() => langStore.t(`food.card.${foodId.value}.en`))

/* 分类页「本类美食」：按标签关键词从 /api/food 列表筛出的真实菜品 */
const categoryDishes = ref<FoodItem[]>([])

/* 类别页数据：/api/food-category/:key(加载失败保持 null,回退 i18n 文案,不白屏) */
const category = ref<FoodCategoryItem | null>(null)
const categoryEstimated = computed(() => estimatedSet(category.value?.estimated_fields))
/** 类别概述：接口无数据时回退 i18n 分类简介 */
const categoryIntro = computed(() => {
  const intro = (category.value?.intro || '').trim()
  return intro || langStore.t(`food.card.${foodId.value}.desc`)
})
/** 概述由 LLM 改写(参考值)时加徽标；接口未返回 intro 时不加,避免给空值标注 */
const categoryIntroEstimated = computed(
  () => !!category.value?.intro && categoryEstimated.value.has('intro'),
)
const categorySections = computed(() => parseSections(category.value?.sections))
const categoryGallery = computed(() => splitList(category.value?.gallery_images))
const categorySourceUrl = computed(() => (category.value?.source_url || '').trim())

/**
 * 是否为"真"类别页：非数字 key 且类别接口确实返回了数据。
 * 无数据时不渲染类别专属区块(类别故事/美味瞬间/来源),避免出现"空的类别骨架 + 暂无数据"。
 */
const isCategoryPage = computed(() => isFallback.value && !!category.value)

const displayName = computed(() => {
  if (food.value) {
    return pickName(food.value, langStore.lang)
  }
  if (isFallback.value) {
    return langStore.lang === 'en' ? fallbackNameEn.value : fallbackNameZh.value
  }
  return noData.value
})

const subTitle = computed(() => {
  if (food.value) {
    return langStore.lang === 'zh' ? (food.value.name_en || '') : food.value.name_zh
  }
  if (isFallback.value) {
    return langStore.lang === 'zh' ? fallbackNameEn.value : fallbackNameZh.value
  }
  return ''
})

const heroTags = computed(() => splitList(food.value?.tags))
const heroImage = computed(() => (food.value?.images && !imgFailed.value) ? food.value.images : '')

const displayDesc = computed(() => {
  if (food.value) return displayFact(pickDesc(food.value, langStore.lang), noData.value)
  if (isFallback.value) return categoryIntro.value
  return noData.value
})

/* 文化注解:仅非中文语言且后端已生成时展示(类别页无此字段,自动为空) */
const cultureNotes = computed(() => pickCultureNote(food.value, langStore.lang))

/* ── 多语语音导览:朗读当前语言的美食故事;分类页无 desc,朗读类别概述 ── */
const speakText = computed(() => {
  if (food.value) return pickDesc(food.value, langStore.lang).trim()
  if (isFallback.value) return categoryIntro.value.trim()
  return ''
})
const voiceAvailable = computed(() => !!speakText.value && hasVoiceFor(langStore.lang))
const voiceTitle = computed(() => {
  if (voiceSpeaking.value) return langStore.t('voice.stop')
  return voiceAvailable.value ? langStore.t('voice.play') : langStore.t('voice.unsupported')
})

function toggleVoice() {
  if (voiceSpeaking.value) {
    stopSpeaking()
    return
  }
  if (speakText.value) speak(speakText.value, langStore.lang)
}

/* 切换语言时停止朗读,避免跨语言续播 */
watch(() => langStore.lang, () => stopSpeaking())

/* ── 四川话一分钟:结构化方言数据(数组取值走语言字典,langStore.t 仅支持字符串) ── */
interface DialectWord { word: string; pinyin: string; meaning: string; scene: string }
const dialectDicts = { zh: dictZh, en: dictEn, ja: dictJa } as const
const dialectWords = computed<DialectWord[]>(
  () => (dialectDicts[langStore.lang] as unknown as { dialect: { words: DialectWord[] } }).dialect.words
)

/* 事实 / 参考值字段：空值统一显示"暂无数据" */
const ratingText = computed(() => displayFact(food.value?.rating, noData.value))
const flavorText = computed(() => displayFact(food.value?.flavor, noData.value))
const spiceText = computed(() => displayFact(food.value?.spice_level, noData.value))
const priceText = computed(() => displayFact(food.value?.avg_price, noData.value))
const signatureText = computed(() => displayFact(food.value?.signature, noData.value))
const sceneText = computed(() => displayFact(food.value?.recommend_scene, noData.value))

/* 风味故事 / 美味瞬间 */
const sections = computed(() => parseSections(food.value?.story_sections))
const galleryImages = computed(() => splitList(food.value?.gallery_images))

/* ── 国际点菜卡:给外国游客看的菜名/直译/食材/辣度(三语均展示) ── */
/**
 * 辣度文本 → 辣椒可视化(0-5 个 🌶)。
 * 常见值映射:不辣/无辣 0、微辣 1、中辣 2、重辣/特辣 4、变态辣/魔鬼辣 5;
 * 未知文本原样展示并给 1 个辣椒;空值显示占位文案。
 */
function spiceChilies(level: string | undefined, placeholder: string): { count: number; text: string } {
  const v = (level || '').trim()
  if (!v) return { count: 0, text: placeholder }
  if (/(不辣|无辣|免辣|清淡)/.test(v)) return { count: 0, text: v }
  if (v.includes('微辣')) return { count: 1, text: v }
  if (v.includes('中辣')) return { count: 2, text: v }
  if (/(变态辣|魔鬼辣|爆辣)/.test(v)) return { count: 5, text: v }
  if (/(重辣|特辣|超辣|狠辣)/.test(v)) return { count: 4, text: v }
  return { count: 1, text: v }
}
const menuSpice = computed(() => spiceChilies(food.value?.spice_level, noData.value))
const hasIngredients = computed(() => !!(food.value?.ingredients_zh || food.value?.ingredients_en))

/* ── 寻味地图：容器常驻可见,占位块绝对定位覆盖；Key/坐标缺失或加载失败时保持占位 ── */
const mapKey = getRuntimeMapJsKey() || import.meta.env.VITE_AMAP_WEB_JS_KEY || ''
const mapReady = ref(false)
let amapInstance: { destroy: () => void } | null = null

async function initMap() {
  const item = food.value
  if (!mapKey || !item?.lng || !item?.lat) return
  try {
    const AMap = await AMapLoader.load({ key: mapKey, version: '2.0', plugins: ['AMap.Marker'] })
    const map = new AMap.Map('food-map', {
      zoom: 15,
      center: [item.lng, item.lat],
      viewMode: '3D',
    })
    new AMap.Marker({ position: [item.lng, item.lat], title: displayName.value, map })
    amapInstance = map
    mapReady.value = true
    await nextTick()
    if (typeof map.resize === 'function') map.resize()
  } catch (err) {
    mapReady.value = false
    console.warn('高德地图初始化失败', err)
  }
}

onBeforeUnmount(() => {
  stopSpeaking()
  if (amapInstance) amapInstance.destroy()
})

/* ── 相关推荐：按 tags 交集数量排序(排除自身),取前 3 条 ── */
async function loadRelated() {
  const self = food.value
  if (!self || !numericId.value) return
  const selfTags = new Set(splitList(self.tags))
  try {
    const res = await listFoods({ page_size: 50 })
    related.value = (res.items || [])
      .filter((it) => it.id !== numericId.value)
      .map((it) => ({ item: it, score: splitList(it.tags).filter((t) => selfTags.has(t)).length }))
      .filter((row) => row.score > 0)
      .sort((a, b) => b.score - a.score)
      .slice(0, 3)
      .map((row) => row.item)
  } catch {
    related.value = []
  }
}

function relatedName(item: FoodItem): string {
  return pickName(item, langStore.lang)
}

/* 简介截断到 40 字以内 */
function relatedDesc(item: FoodItem): string {
  const d = (item.desc || '').trim()
  return d.length > 40 ? d.slice(0, 40) + '…' : d
}

function goRelated(item: FoodItem) {
  router.push({ name: 'food-detail', params: { id: String(item.id) } })
}

/* ── 分类页：仅拉取美食列表并按标签筛选,不请求详情 ── */
async function loadCategoryDishes() {
  const keywords = CATEGORY_TAGS[foodId.value] || []
  if (!keywords.length) return
  try {
    const res = await listFoods({ page_size: 50 })
    categoryDishes.value = matchCategoryDishes(res.items || [], keywords)
  } catch {
    categoryDishes.value = []
  }
}

/* ── 类别页：拉类别介绍;失败保持 null,页面回退 i18n 文案 ── */
async function loadCategory() {
  try {
    category.value = await getFoodCategory(foodId.value)
  } catch {
    category.value = null
  }
}

/**
 * 类别接口无数据时,把该 key 当作菜品键处理(火锅/串串香/盖碗茶)：
 * 用 listFoods 列表按代表菜名 name_zh 精确匹配出真实菜品 id,再 replace 到菜品详情页
 * (replace 避免污染历史栈)。不写死数字 id。
 * @returns 是否已跳转;未收录该键 / 未匹配到 / 接口失败 → false,由调用方走最小降级视图
 */
async function goDishByCategoryKey(key: string): Promise<boolean> {
  const nameZh = CATEGORY_DISH_NAMES[key]
  if (!nameZh) {
    console.warn(`[FoodDetail] 类别接口无数据且「${key}」不在菜品键映射中，回退最小展示`)
    return false
  }
  try {
    const res = await listFoods({ page_size: 50 })
    const found = (res.items || []).find((item) => item.name_zh === nameZh)
    if (found) {
      await router.replace({ name: 'food-detail', params: { id: String(found.id) } })
      return true
    }
    console.warn(`[FoodDetail] 未匹配到菜品「${nameZh}」(/food/${key})，回退最小展示`)
  } catch (err) {
    console.warn(`[FoodDetail] 解析「${key}」对应菜品失败，回退最小展示`, err)
  }
  return false
}

/* ── 载入：数字 id 拉详情 + 地图 + 相关推荐；类别 key 拉类别介绍 + 本类美食 ── */
async function loadDetail() {
  if (amapInstance) {
    amapInstance.destroy()
    amapInstance = null
  }
  food.value = null
  related.value = []
  category.value = null
  categoryDishes.value = []
  imgFailed.value = false
  mapReady.value = false
  if (!numericId.value) {
    // 已知菜品键(火锅/串串香/盖碗茶)：先解析菜品数字 id 跳转，
    // 避免对类别接口发一次必然 404 的请求（那会让控制台出现失败请求）
    if (CATEGORY_DISH_NAMES[foodId.value] && (await goDishByCategoryKey(foodId.value))) return
    await loadCategory()
    // 类别接口失败/无数据：可能是不在类别表里的菜品键,解析出菜品数字 id 后跳转
    if (!category.value && (await goDishByCategoryKey(foodId.value))) return
    await loadCategoryDishes()
    return
  }
  try {
    food.value = await getFood(numericId.value)
  } catch {
    /* 加载失败保持占位展示 */
  }
  await initMap()
  await loadRelated()
}

onMounted(loadDetail)
/* 同路由记录参数变化(点击相关推荐/切换分类)时组件被复用，需手动重载 */
watch(foodId, () => { loadDetail() })

/* ════════════════════════════════════════════════
 *  说明：数据驱动的载入逻辑见上方 loadDetail / initMap / loadRelated
 *  ════════════════════════════════════════════════ */


/* ════════════════════════════════════════════════
 *  接口字段与数据来源标注
 *  ════════════════════════════════════════════════
 *  详见 @/api/content 的 FoodItem 类型与 @/utils/scenicDetail 纯函数
 *  ════════════════════════════════════════════════ */

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
      <!-- ──── HERO — 封面 + 名称浮层 ──── -->
      <section class="detail-hero">
        <div class="detail-hero__media">
          <!-- 接口返回封面图,加载失败回退占位 -->
          <img
            v-if="heroImage"
            class="detail-hero__photo"
            :src="heroImage"
            :alt="displayName"
            referrerpolicy="no-referrer"
            @error="imgFailed = true"
          />
          <div v-else class="detail-hero__placeholder">
            <div class="detail-hero__shu" aria-hidden="true">味</div>
            <span v-if="!isFallback" class="detail-hero__api-badge">{{ langStore.t('foodDetail.imagePlaceholder') }}</span>
          </div>

          <!-- 名称浮层 -->
          <div class="detail-hero__overlay">
            <h1 class="detail-hero__title">{{ displayName }}</h1>
            <p class="detail-hero__en-title">{{ subTitle }}</p>
            <div v-if="heroTags.length" class="detail-hero__tags">
              <span
                v-for="(t, i) in heroTags"
                :key="t"
                class="detail-hero__tag"
                :class="{ 'detail-hero__tag--accent': i === 0 }"
              >
                {{ t }}
              </span>
            </div>
          </div>
        </div>
      </section>

      <!-- ──── 概要：评分 + 简介（分类页仅保留简介）──── -->
      <section class="detail-section container">
        <div class="detail-summary" :class="{ 'detail-summary--solo': isFallback }">
          <!-- 左：评分面板 -->
          <aside v-if="!isFallback" class="detail-summary__aside">
            <div class="detail-score">
              <span class="detail-score__num">{{ ratingText }}</span>
              <div class="detail-score__meta">
                <span class="detail-score__stars">
                  <AppIcon v-for="i in 5" :key="i" name="star" :size="16" />
                </span>
                <span class="detail-score__label">
                  {{ langStore.t('foodDetail.rating') }}
                  <em v-if="showEst('rating', food?.rating)" class="detail-est">{{ langStore.t('foodDetail.estimated') }}</em>
                </span>
              </div>
            </div>
            <div class="detail-aside__row">
              <span class="detail-aside__key">{{ langStore.t('foodDetail.flavor') }}</span>
              <span class="detail-aside__value">
                {{ flavorText }}
                <em v-if="showEst('flavor', food?.flavor)" class="detail-est">{{ langStore.t('foodDetail.estimated') }}</em>
              </span>
            </div>
            <div class="detail-aside__row">
              <span class="detail-aside__key">{{ langStore.t('foodDetail.spiceLevel') }}</span>
              <span class="detail-aside__value">
                {{ spiceText }}
                <em v-if="showEst('spice_level', food?.spice_level)" class="detail-est">{{ langStore.t('foodDetail.estimated') }}</em>
              </span>
            </div>
            <div class="detail-aside__row">
              <span class="detail-aside__key">{{ langStore.t('foodDetail.avgPrice') }}</span>
              <span class="detail-aside__value">
                {{ priceText }}
                <em v-if="showEst('avg_price', food?.avg_price)" class="detail-est">{{ langStore.t('foodDetail.estimated') }}</em>
              </span>
            </div>
          </aside>

          <!-- 右：简介（类别页为类别概述） -->
          <div class="detail-summary__body">
            <div class="detail-summary__eyebrow">
              <span>◈</span>
              {{ langStore.t(isFallback ? 'foodDetail.categoryOverview' : 'foodDetail.overview') }}
              <span>◈</span>
              <!-- 语音导览:朗读当前语言的美食故事 -->
              <button
                type="button"
                class="voice-btn"
                :class="{ 'voice-btn--active': voiceSpeaking }"
                :disabled="!voiceAvailable && !voiceSpeaking"
                :title="voiceTitle"
                @click="toggleVoice"
              >
                <span class="voice-btn__icon" aria-hidden="true">{{ voiceSpeaking ? '⏹' : '🔊' }}</span>
                {{ voiceSpeaking ? langStore.t('voice.stop') : langStore.t('voice.play') }}
              </button>
            </div>
            <p class="detail-summary__text">
              {{ displayDesc }}
              <em v-if="isFallback && categoryIntroEstimated" class="detail-est">{{ langStore.t('foodDetail.estimated') }}</em>
            </p>

            <!-- 文化注解:外语模式下展示 LLM 生成的当地饮食文化背景 -->
            <div v-if="cultureNotes.length" class="culture-note">
              <div class="culture-note__head">
                <span class="culture-note__icon" aria-hidden="true">📖</span>
                <div class="culture-note__titles">
                  <h3 class="culture-note__title">{{ langStore.t('cultureNote.title') }}</h3>
                  <p class="culture-note__hint">{{ langStore.t('cultureNote.hint') }}</p>
                </div>
              </div>
              <ul class="culture-note__list">
                <li v-for="(note, i) in cultureNotes" :key="i" class="culture-note__item">
                  <span class="culture-note__dot" aria-hidden="true" />
                  <span>{{ note }}</span>
                </li>
              </ul>
            </div>

            <!-- 四川话一分钟:方言彩蛋,所有语言均展示 -->
            <div class="dialect-card">
              <div class="dialect-card__head">
                <span class="dialect-card__icon" aria-hidden="true">🌶</span>
                <div class="dialect-card__titles">
                  <h3 class="dialect-card__title">{{ langStore.t('dialect.title') }}</h3>
                  <p class="dialect-card__hint">{{ langStore.t('dialect.hint') }}</p>
                </div>
              </div>
              <ul class="dialect-card__list">
                <li v-for="w in dialectWords" :key="w.word" class="dialect-card__item">
                  <span class="dialect-card__word">{{ w.word }}</span>
                  <span class="dialect-card__pinyin">{{ w.pinyin }}</span>
                  <span class="dialect-card__meaning">{{ w.meaning }}</span>
                  <span class="dialect-card__scene">{{ w.scene }}</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      <!-- ──── 国际点菜卡（仅数字 id 详情页,三语均展示）──── -->
      <section v-if="food" class="detail-section container">
        <div class="menu-card">
          <header class="menu-card__head">
            <span class="menu-card__icon" aria-hidden="true">🧾</span>
            <h3 class="menu-card__title">{{ langStore.t('menuCard.title') }}</h3>
          </header>

          <div class="menu-card__names">
            <span class="menu-card__name-zh">{{ food.name_zh }}</span>
            <span v-if="food.name_en" class="menu-card__name-en">{{ food.name_en }}</span>
          </div>

          <div v-if="food.name_literal_en" class="menu-card__literal">
            <span class="menu-card__literal-text">
              {{ langStore.t('menuCard.literal') }}: {{ food.name_literal_en }}
            </span>
            <span class="menu-card__literal-tip">{{ langStore.t('menuCard.literalTip') }}</span>
          </div>

          <div class="menu-card__rows">
            <div class="menu-card__row">
              <span class="menu-card__key">{{ langStore.t('menuCard.ingredients') }}</span>
              <span v-if="hasIngredients" class="menu-card__val">
                <span v-if="food.ingredients_zh" class="menu-card__ing">
                  <i class="menu-card__ing-label">{{ langStore.t('menuCard.ingredientsZh') }}</i>
                  {{ food.ingredients_zh }}
                </span>
                <span v-if="food.ingredients_en" class="menu-card__ing">
                  <i class="menu-card__ing-label">{{ langStore.t('menuCard.ingredientsEn') }}</i>
                  {{ food.ingredients_en }}
                </span>
              </span>
              <span v-else class="menu-card__val menu-card__val--empty">{{ noData }}</span>
            </div>
            <div class="menu-card__row">
              <span class="menu-card__key">{{ langStore.t('menuCard.spice') }}</span>
              <span class="menu-card__val">
                <span class="menu-card__chilies" aria-hidden="true">
                  <span
                    v-for="i in 5"
                    :key="i"
                    class="menu-card__chili"
                    :class="{ 'menu-card__chili--dim': i > menuSpice.count }"
                  >🌶</span>
                </span>
                <span class="menu-card__spice-text">{{ menuSpice.text }}</span>
              </span>
            </div>
          </div>
        </div>
      </section>

      <!-- ──── 类别故事（类别页：接口返回的图文段落,左右交替）──── -->
      <section v-if="isCategoryPage" class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.categoryStories') }}</h2>
        </header>

        <div v-if="categorySections.length" class="detail-content">
          <div
            v-for="(sec, i) in categorySections"
            :key="sec.title"
            class="detail-content__row"
            :class="{ 'detail-content__row--reverse': i % 2 === 1 }"
          >
            <div class="detail-content__figure">
              <img
                v-if="sec.image"
                class="detail-content__img detail-content__img--photo"
                :src="sec.image"
                :alt="sec.title"
                referrerpolicy="no-referrer"
              />
              <div v-else class="detail-content__img detail-content__img--empty">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
                  <rect x="3" y="3" width="18" height="18" rx="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <path d="M21 15l-5-5L5 21"/>
                </svg>
              </div>
            </div>
            <div class="detail-content__text">
              <h3 class="detail-content__caption">{{ sec.title }}</h3>
              <p class="detail-content__para">{{ sec.text }}</p>
            </div>
          </div>
        </div>
        <p v-else class="detail-empty">{{ noData }}</p>
      </section>

      <!-- ──── 美味瞬间（类别页：类别图集）──── -->
      <section v-if="isCategoryPage" class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.galleryTitle') }}</h2>
        </header>

        <div v-if="categoryGallery.length" class="detail-gallery">
          <div v-for="(img, i) in categoryGallery" :key="img" class="detail-gallery__item">
            <img class="detail-gallery__photo" :src="img" :alt="`${displayName} ${i + 1}`" referrerpolicy="no-referrer" />
          </div>
        </div>
        <p v-else class="detail-empty">{{ noData }}</p>
      </section>

      <!-- ──── 本类美食（分类页：按标签筛出的真实菜品）──── -->
      <section v-if="isFallback" class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.categoryDishes') }}</h2>
        </header>

        <div v-if="categoryDishes.length" class="detail-dishes">
          <article
            v-for="item in categoryDishes"
            :key="item.id"
            class="detail-dishes__card"
            tabindex="0"
            @click="goRelated(item)"
            @keyup.enter="goRelated(item)"
          >
            <div class="detail-dishes__media">
              <img
                v-if="item.images"
                class="detail-dishes__photo"
                :src="item.images"
                :alt="relatedName(item)"
                loading="lazy"
                referrerpolicy="no-referrer"
              />
              <span v-else class="detail-dishes__placeholder" aria-hidden="true">
                <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
                  <rect x="3" y="3" width="18" height="18" rx="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <path d="M21 15l-5-5L5 21"/>
                </svg>
              </span>
            </div>
            <div class="detail-dishes__body">
              <h3 class="detail-dishes__name">{{ relatedName(item) }}</h3>
              <p class="detail-dishes__desc">{{ relatedDesc(item) }}</p>
            </div>
          </article>
        </div>
        <p v-else class="detail-empty">{{ langStore.t('foodDetail.categoryEmpty') }}</p>
      </section>

      <!-- ──── 类别页页脚：数据来源 + 内容素材来源(维基百科) ──── -->
      <section v-if="isCategoryPage" class="detail-section container">
        <p class="detail-source">
          {{ langStore.t('foodDetail.dataSource') }}
          <template v-if="categoryEstimated.size">· {{ langStore.t('foodDetail.estimatedTip') }}</template>
        </p>
        <p class="detail-source">
          {{ langStore.t('foodDetail.contentSource') }}
          <a
            v-if="categorySourceUrl"
            class="detail-source__link"
            :href="categorySourceUrl"
            target="_blank"
            rel="noopener noreferrer"
          >{{ categorySourceUrl }}</a>
        </p>
      </section>

      <!-- ──── 实用信息（仅数字 id 详情页）──── -->
      <section v-if="!isFallback" class="detail-section container">
        <div class="detail-info">
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="star" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('foodDetail.signature') }}</h3>
            <p class="detail-info__value">
              {{ signatureText }}
              <em v-if="showEst('signature', food?.signature)" class="detail-est">{{ langStore.t('foodDetail.estimated') }}</em>
            </p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="fire" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('foodDetail.spiceLevel') }}</h3>
            <p class="detail-info__value">
              {{ spiceText }}
              <em v-if="showEst('spice_level', food?.spice_level)" class="detail-est">{{ langStore.t('foodDetail.estimated') }}</em>
            </p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="ticket" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('foodDetail.avgPrice') }}</h3>
            <p class="detail-info__value">
              {{ priceText }}
              <em v-if="showEst('avg_price', food?.avg_price)" class="detail-est">{{ langStore.t('foodDetail.estimated') }}</em>
            </p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon"><AppIcon name="clock" :size="24" /></span>
            <h3 class="detail-info__title">{{ langStore.t('foodDetail.recommendScene') }}</h3>
            <p class="detail-info__value">
              {{ sceneText }}
              <em v-if="showEst('recommend_scene', food?.recommend_scene)" class="detail-est">{{ langStore.t('foodDetail.estimated') }}</em>
            </p>
          </div>
        </div>
      </section>

      <!-- ──── 风味故事（图文段落,左右交替）（仅数字 id 详情页）──── -->
      <section v-if="!isFallback" class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.detailTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('foodDetail.detailSubtitle') }}</p>
        </header>

        <div v-if="sections.length" class="detail-content">
          <div
            v-for="(sec, i) in sections"
            :key="sec.title"
            class="detail-content__row"
            :class="{ 'detail-content__row--reverse': i % 2 === 1 }"
          >
            <div class="detail-content__figure">
              <img
                v-if="sec.image"
                class="detail-content__img detail-content__img--photo"
                :src="sec.image"
                :alt="sec.title"
                referrerpolicy="no-referrer"
              />
              <div v-else class="detail-content__img detail-content__img--empty">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
                  <rect x="3" y="3" width="18" height="18" rx="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <path d="M21 15l-5-5L5 21"/>
                </svg>
              </div>
            </div>
            <div class="detail-content__text">
              <h3 class="detail-content__caption">{{ sec.title }}</h3>
              <p class="detail-content__para">{{ sec.text }}</p>
            </div>
          </div>
        </div>
        <p v-else class="detail-empty">{{ noData }}</p>
      </section>

      <!-- ──── 美味瞬间（仅数字 id 详情页）──── -->
      <section v-if="!isFallback" class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.galleryTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('foodDetail.gallerySubtitle') }}</p>
        </header>

        <div v-if="galleryImages.length" class="detail-gallery">
          <div v-for="(img, i) in galleryImages" :key="img" class="detail-gallery__item">
            <img class="detail-gallery__photo" :src="img" :alt="`${displayName} ${i + 1}`" referrerpolicy="no-referrer" />
          </div>
        </div>
        <p v-else class="detail-empty">{{ noData }}</p>
      </section>

      <!-- ──── 寻味地图（容器常驻可见,占位块绝对定位覆盖）（仅数字 id 详情页）──── -->
      <section v-if="!isFallback" class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.mapTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('foodDetail.mapSubtitle') }}</p>
        </header>

        <div class="detail-map">
          <div id="food-map" class="detail-map__canvas" />
          <div v-if="!mapReady" class="detail-map__placeholder">
            <span class="detail-map__pin"><AppIcon name="pin" :size="40" /></span>
            <span class="detail-map__hint">{{ noData }}</span>
          </div>
        </div>
      </section>

      <!-- ──── 相关推荐 + 数据来源（仅数字 id 详情页）──── -->
      <section v-if="!isFallback" class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('foodDetail.aroundTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('foodDetail.aroundSubtitle') }}</p>
        </header>

        <div v-if="related.length" class="detail-around">
          <div
            v-for="item in related"
            :key="item.id"
            class="detail-around__card"
            @click="goRelated(item)"
          >
            <div class="detail-around__body">
              <h3 class="detail-around__name">{{ relatedName(item) }}</h3>
              <p class="detail-around__desc">{{ relatedDesc(item) }}</p>
            </div>
          </div>
        </div>
        <p v-else class="detail-empty">{{ langStore.t('foodDetail.relatedEmpty') }}</p>

        <p class="detail-source">
          {{ langStore.t('foodDetail.dataSource') }}
          <template v-if="estimated.size">· {{ langStore.t('foodDetail.estimatedTip') }}</template>
        </p>
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

.detail-hero__photo {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.detail-content__img--photo {
  display: block;
  object-fit: cover;
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

/* 分类页无评分面板,简介独占整行 */
.detail-summary--solo {
  grid-template-columns: 1fr;
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
   文化注解卡(外语模式:琥珀色渐变)
   ======================================== */
.culture-note {
  padding: var(--space-5) var(--space-6);
  border-radius: var(--radius-lg);
  border: 1px solid rgba(201, 138, 43, 0.35);
  background: linear-gradient(135deg, rgba(243, 224, 178, 0.6) 0%, rgba(252, 246, 230, 0.9) 60%, rgba(250, 236, 205, 0.55) 100%);
}

.culture-note__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.culture-note__icon {
  display: flex;
  align-items: center;
  font-size: var(--text-xl);
}

.culture-note__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  color: #8a5a12;
  letter-spacing: var(--tracking-wide);
}

.culture-note__hint {
  margin-top: 2px;
  font-size: var(--text-xs);
  color: rgba(138, 90, 18, 0.72);
  letter-spacing: var(--tracking-wide);
}

.culture-note__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.culture-note__item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
  font-size: var(--text-sm);
  line-height: var(--leading-relaxed);
  color: #6b4f26;
}

.culture-note__dot {
  flex-shrink: 0;
  width: 6px;
  height: 6px;
  margin-top: 8px;
  border-radius: var(--radius-full);
  background: #c98a2b;
}

/* ── 语音导览按钮(故事标题旁) ── */
.voice-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: var(--space-3);
  padding: 3px 12px;
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wide);
  color: var(--color-cinnabar);
  border: 1px solid color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-cinnabar) 6%, transparent);
  transition: color var(--transition-fast), background var(--transition-fast),
    border-color var(--transition-fast), transform var(--transition-fast);
}

.voice-btn:hover:not(:disabled) {
  background: color-mix(in srgb, var(--color-cinnabar) 14%, transparent);
  border-color: var(--color-cinnabar);
  transform: translateY(-1px);
}

.voice-btn--active {
  color: var(--color-cinnabar);
  border-color: var(--color-cinnabar);
  background: color-mix(in srgb, var(--color-cinnabar) 14%, transparent);
}

.voice-btn:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.voice-btn__icon {
  font-size: var(--text-sm);
  line-height: 1;
}

/* ========================================
   四川话一分钟卡(方言彩蛋:辣椒红渐变,与琥珀色 Culture Note 区分)
   ======================================== */
.dialect-card {
  padding: var(--space-5) var(--space-6);
  border-radius: var(--radius-lg);
  border: 1px solid rgba(190, 47, 30, 0.32);
  background: linear-gradient(135deg, rgba(248, 218, 208, 0.6) 0%, rgba(253, 244, 239, 0.9) 60%, rgba(250, 226, 216, 0.55) 100%);
}

.dialect-card__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.dialect-card__icon {
  display: flex;
  align-items: center;
  font-size: var(--text-xl);
}

.dialect-card__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  color: #a02c1d;
  letter-spacing: var(--tracking-wide);
}

.dialect-card__hint {
  margin-top: 2px;
  font-size: var(--text-xs);
  color: rgba(160, 44, 29, 0.72);
  letter-spacing: var(--tracking-wide);
}

.dialect-card__list {
  display: flex;
  flex-direction: column;
}

.dialect-card__item {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: var(--space-1) var(--space-3);
  padding: var(--space-2) 0;
  border-bottom: 1px dashed rgba(190, 47, 30, 0.22);
}

.dialect-card__item:last-child {
  border-bottom: none;
}

.dialect-card__word {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 800;
  color: var(--color-cinnabar);
  letter-spacing: var(--tracking-wide);
  line-height: 1.2;
}

.dialect-card__pinyin {
  font-family: var(--font-en-body);
  font-style: italic;
  font-size: var(--text-xs);
  color: rgba(160, 44, 29, 0.78);
  letter-spacing: var(--tracking-wide);
}

.dialect-card__meaning {
  font-size: var(--text-sm);
  color: #6f2a20;
}

.dialect-card__scene {
  margin-left: auto;
  font-size: var(--text-xs);
  color: rgba(111, 42, 32, 0.6);
  text-align: right;
}

@media (max-width: 640px) {
  .dialect-card {
    padding: var(--space-4);
  }
  .dialect-card__scene {
    margin-left: 0;
    width: 100%;
    text-align: left;
  }
}

/* ========================================
   国际点菜卡(给外国游客点菜用,三语均展示)
   ======================================== */
.menu-card {
  max-width: 720px;
  margin: 0 auto;
  padding: var(--space-6) var(--space-8);
  background:
    linear-gradient(135deg, rgba(243, 224, 178, 0.35) 0%, rgba(252, 246, 230, 0.65) 100%),
    var(--color-surface);
  border: 1px solid rgba(201, 138, 43, 0.4);
  border-radius: var(--radius-lg);
  box-shadow: 0 6px 24px rgba(201, 138, 43, 0.12);
}

.menu-card__head {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding-bottom: var(--space-3);
  border-bottom: 1px dashed rgba(201, 138, 43, 0.45);
}

.menu-card__icon {
  display: flex;
  align-items: center;
  font-size: var(--text-lg);
}

.menu-card__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  color: #8a5a12;
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.menu-card__names {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: var(--space-3) var(--space-5);
  padding: var(--space-5) 0 var(--space-2);
}

.menu-card__name-zh {
  font-family: var(--font-display);
  font-size: clamp(var(--text-3xl), 5vw, var(--text-5xl));
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.15;
}

.menu-card__name-en {
  font-family: var(--font-en-display);
  font-style: italic;
  font-size: var(--text-base);
  color: var(--color-cinnabar);
  letter-spacing: var(--tracking-wider);
}

.menu-card__literal {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2) var(--space-3);
  margin-bottom: var(--space-4);
}

.menu-card__literal-text {
  font-family: var(--font-en-body);
  font-size: var(--text-sm);
  font-style: italic;
  color: #8a5a12;
  letter-spacing: var(--tracking-wide);
}

.menu-card__literal-tip {
  font-size: var(--text-xs);
  padding: 2px 10px;
  border-radius: var(--radius-full);
  border: 1px solid rgba(201, 138, 43, 0.5);
  background: rgba(255, 255, 255, 0.55);
  color: #8a5a12;
  letter-spacing: var(--tracking-wide);
}

.menu-card__rows {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.menu-card__row {
  display: grid;
  grid-template-columns: 120px 1fr;
  align-items: start;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-4);
  background: rgba(255, 255, 255, 0.5);
  border: 1px solid rgba(201, 138, 43, 0.25);
  border-radius: var(--radius-md);
}

.menu-card__key {
  font-size: var(--text-xs);
  color: #8a5a12;
  letter-spacing: var(--tracking-wide);
  padding-top: 2px;
}

.menu-card__val {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

.menu-card__val--empty {
  color: var(--color-text-muted);
}

.menu-card__ing {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
}

.menu-card__ing-label {
  flex-shrink: 0;
  font-style: normal;
  font-size: var(--text-xs);
  color: rgba(138, 90, 18, 0.72);
}

.menu-card__chilies {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: var(--text-base);
  line-height: 1;
}

.menu-card__chili--dim {
  opacity: 0.18;
  filter: grayscale(1);
}

.menu-card__spice-text {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
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

/* 左右交替：奇数行把配图放到右侧 */
.detail-content__row--reverse .detail-content__figure {
  order: 2;
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
  white-space: pre-line;
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

.detail-gallery__photo {
  width: 100%;
  height: 100%;
  aspect-ratio: 4 / 3;
  object-fit: cover;
  display: block;
}

/* ========================================
   地图占位
   ======================================== */
.detail-map {
  position: relative;
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

/* 地图容器常驻可见(高德需要非零尺寸),占位块绝对定位覆盖其上 */
.detail-map__canvas {
  width: 100%;
  aspect-ratio: 16 / 6;
}

.detail-map__placeholder {
  position: absolute;
  inset: 0;
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
   本类美食(分类页网格)
   ======================================== */
.detail-dishes {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.detail-dishes__card {
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  outline: none;
  transition: all var(--transition-base);
}

.detail-dishes__card:hover,
.detail-dishes__card:focus-visible {
  border-color: var(--color-cinnabar);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px var(--color-cinnabar-dim), var(--shadow-lg);
}

.detail-dishes__media {
  position: relative;
  aspect-ratio: 4 / 3;
  overflow: hidden;
  background:
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-cinnabar) 8%, transparent) 0%, transparent 70%),
    var(--color-bg-alt);
}

.detail-dishes__photo {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.35s ease;
}

.detail-dishes__card:hover .detail-dishes__photo {
  transform: scale(1.05);
}

.detail-dishes__placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  opacity: 0.7;
}

.detail-dishes__body {
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.detail-dishes__name {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.detail-dishes__desc {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  line-height: var(--leading-relaxed);
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
  cursor: pointer;
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
   参考值徽标 / 空数据 / 数据来源
   ======================================== */
.detail-est {
  display: inline-block;
  margin-left: var(--space-1);
  padding: 1px 6px;
  font-size: var(--text-xs);
  font-style: normal;
  color: var(--color-cinnabar);
  border: 1px solid color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
  border-radius: var(--radius-full);
  vertical-align: middle;
}

.detail-empty {
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.detail-source {
  margin-top: var(--space-8);
  text-align: center;
  font-size: var(--text-xs);
  color: var(--color-text-muted);
}

/* 类别页页脚第二行(内容素材来源)紧贴上一行 */
.detail-source + .detail-source {
  margin-top: var(--space-2);
}

.detail-source__link {
  margin-left: var(--space-1);
  color: var(--color-cinnabar);
  text-decoration: underline;
  word-break: break-all;
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
  .detail-dishes {
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
  .detail-around,
  .detail-dishes {
    grid-template-columns: 1fr;
  }
  .detail-hero__overlay {
    padding: var(--space-6) var(--space-4);
  }
  .menu-card {
    padding: var(--space-5) var(--space-4);
  }
  .menu-card__row {
    grid-template-columns: 1fr;
    gap: var(--space-2);
  }
}
</style>
