<script setup lang="ts">
/**
 * HomeDashboard.vue — 内部首页仪表盘
 * 内容: 欢迎横幅 + 天气 + 推荐景点 + 文旅热点 + 成都冷知识
 * 说明: 与探索页(ExplorePage)去重 — 不再展示区域筛选/全部景点网格，
 *       改为首页专属的精选推荐、热点新闻与趣味冷知识。
 */
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import AppIcon from '@/components/AppIcon.vue'
import HomeBanner from '@/components/HomeBanner.vue'
import { getRecommendedSpots, newsItems, funFacts } from '@/data/chengdu'
import { listScenics, listHotspots, type ScenicItem, type HotspotItem } from '@/api/content'
import { getCultureDaily, getSolarFoods, type CultureDailyItem, type SolarFoodItem } from '@/api/culture'
import { pickName } from '@/utils/storyI18n'
import { getCurrentSolarTerm } from '@/utils/solar-term'
import { fetchStamps } from '@/utils/passport'
import { hasToken } from '@/utils/token'

const router = useRouter()
const langStore = useLanguageStore()

/** 数字足迹护照:进入本页时从后端拉取的集章数(徽标展示,登录用户按账号落库) */
const stampCountValue = ref(0)

onMounted(async () => {
  if (!hasToken()) return
  try {
    stampCountValue.value = (await fetchStamps()).length
  } catch {
    /* 拉取失败徽标保持 0,不影响主流程 */
  }
  // 今日蜀签:失败/无数据时卡片整体隐藏
  loadDaily(0)
  // 节气蜀俗:失败/无应季美食时条幅整体隐藏
  loadSolarFoods()
})

function goPassport() {
  router.push('/passport')
}

/** ============================================================
 *  今日蜀签:每天一条蜀文化日签(诗句/方言/冷知识,三语)
 *  后端按 (dayOfYear + offset) % 总数 轮换;接口失败/无数据时整体隐藏
 *  ============================================================ */
const daily = ref<CultureDailyItem | null>(null)
const dailyOffset = ref(0)
const dailyLoading = ref(false)

/** 数据库分类值(中文) → locales 键,未识别的分类回落显示原值 */
const DAILY_CATEGORY_I18N: Record<string, string> = {
  诗句: 'daily.categoryPoem',
  方言: 'daily.categoryDialect',
  冷知识: 'daily.categoryFact',
}

/** 拉取蜀签(失败静默,保持卡片隐藏/内容不变) */
async function loadDaily(offset: number) {
  if (dailyLoading.value) return
  dailyLoading.value = true
  try {
    const item = await getCultureDaily(offset)
    if (item && item.id) {
      daily.value = item
      dailyOffset.value = offset
    }
  } catch {
    /* 静默:无数据/接口异常时隐藏卡片,不留空块 */
  } finally {
    dailyLoading.value = false
  }
}

/** 换一条:偏移 +1 顺延取下一条 */
function swapDaily() {
  loadDaily(dailyOffset.value + 1)
}

/** 按当前语言取内容(en/ja 未生成时回落中文) */
const dailyContent = computed(() => {
  const item = daily.value
  if (!item) return ''
  if (langStore.lang === 'en') return item.content_en || item.content_zh
  if (langStore.lang === 'ja') return item.content_ja || item.content_zh
  return item.content_zh
})

/** 分类标签文案(三语映射,未识别回落中文原值) */
const dailyCategoryLabel = computed(() => {
  const key = DAILY_CATEGORY_I18N[daily.value?.category ?? '']
  return key ? langStore.t(key) : (daily.value?.category ?? '')
})

/** 去看看:关联景点详情页 */
function goDailySpot() {
  const id = daily.value?.related_spot_id
  if (id) router.push(`/scenic/${id}`)
}

/** ============================================================
 *  节气蜀俗:按当前节气(近似日历表,容差 ±1 天)展示三语条幅与应季美食。
 *  条幅文案/节气三语名内联在 src/utils/solar-term.ts(内容数据非 UI 标签);
 *  后端无匹配节气美食或接口失败时整体隐藏,不留空块。
 *  ============================================================ */
const solarTerm = getCurrentSolarTerm()
const solarFoods = ref<SolarFoodItem[]>([])
const solarLoaded = ref(false)

/** 当前节气名(三语) */
const solarTermLabel = computed(() => {
  const name =
    langStore.lang === 'en' ? solarTerm.nameEn : langStore.lang === 'ja' ? solarTerm.nameJa : solarTerm.name
  // 附节气管辖区间(交节日—下一交节日前一天),明确"当前处于该节气期间"而非"今日交节"
  const range = solarTerm.rangeText
  return range ? `${name} · ${range[langStore.lang === 'en' ? 'en' : langStore.lang === 'ja' ? 'ja' : 'zh']}` : name
})

/** 条幅文案(三语) */
const solarBanner = computed(() => {
  if (langStore.lang === 'en') return solarTerm.banner.en
  if (langStore.lang === 'ja') return solarTerm.banner.ja
  return solarTerm.banner.zh
})

/** 条幅整体可见:接口请求完成即显示(节气文案 24 组内联常显,保证功能全年可见);应季美食小区块由长度单独控制 */
const solarVisible = computed(() => solarLoaded.value)

/** 拉取当前节气的应季美食(失败静默,条幅整体隐藏) */
async function loadSolarFoods() {
  try {
    const list = await getSolarFoods(solarTerm.name)
    if (Array.isArray(list)) solarFoods.value = list.slice(0, 3)
  } catch {
    /* 静默:接口失败时整体隐藏条幅 */
  } finally {
    solarLoaded.value = true
  }
}

/** 美食名按当前语言展示(en/ja 未生成时回落中文) */
function solarFoodName(item: SolarFoodItem) {
  if (langStore.lang === 'en') return item.name_en || item.name_zh
  if (langStore.lang === 'ja') return item.name_ja || item.name_zh
  return item.name_zh
}

/** 取第一张图片 */
function solarFoodImage(item: SolarFoodItem) {
  return (item.images || '').split(',').map(s => s.trim()).find(Boolean) || ''
}

/** 点击跳美食详情页 */
function goSolarFood(id: number) {
  router.push(`/food/${id}`)
}

/** 统一的推荐景点展示结构(接口数据优先,静态数据兜底) */
interface RecSpot {
  key: string
  nameZh: string
  nameEn: string
  descZh: string
  descEn: string
  tags: string[]
  rating: number
  image: string
}

const RECOMMEND_COUNT = 6

function fromRemoteScenic(item: ScenicItem): RecSpot {
  return {
    key: `db-${item.id}`,
    nameZh: item.name_zh,
    nameEn: pickName(item, langStore.lang),
    descZh: item.desc || '',
    descEn: item.desc || '',
    tags: (item.tags || '').split(',').map(t => t.trim()).filter(Boolean),
    rating: item.score || 0,
    image: item.images || '',
  }
}

function fromStatic(): RecSpot[] {
  return getRecommendedSpots().map(s => ({
    key: s.id,
    nameZh: s.nameZh,
    nameEn: s.nameEn,
    descZh: s.shortDescZh,
    descEn: s.shortDescEn,
    tags: s.tags,
    rating: s.rating,
    image: '',
  }))
}

const remoteSpots = ref<RecSpot[]>([])
const recommended = computed<RecSpot[]>(() =>
  remoteSpots.value.length >= RECOMMEND_COUNT ? remoteSpots.value : fromStatic(),
)

/** 接口失败时静默回退静态数据 */
onMounted(async () => {
  try {
    const page = await listScenics({ page: 1, page_size: 100 })
    const spots = [...page.items]
      .sort((a, b) => (b.score || 0) - (a.score || 0))
      .slice(0, RECOMMEND_COUNT)
      .map(fromRemoteScenic)
    if (spots.length >= RECOMMEND_COUNT) remoteSpots.value = spots
  } catch {
    /* 保持静态兜底 */
  }
})

/** 图片加载失败时移除,回退首字水印占位 */
function onImgError(spot: RecSpot) {
  spot.image = ''
}

/** ============================================================
 *  文旅热点:接口数据优先,静态数据兜底
 *  ============================================================ */
/** 统一的热点展示结构(远端与静态数据映射到同一形状) */
interface NewsDisplay {
  key: string
  titleZh: string
  titleEn: string
  titleJa: string
  summaryZh: string
  summaryEn: string
  summaryJa: string
  sourceZh: string
  sourceEn: string
  timeZh: string
  timeEn: string
  timeJa: string
  hot: boolean
  url: string
}

/** 远端热点(接口失败时保持空,展示静态兜底) */
const NEWS_PAGE_SIZE = 6
const remoteNews = ref<NewsDisplay[]>([])
const remoteNewsTotal = ref(0)
const newsPage = ref(1)
const newsLoading = ref(false)
const newsTotalPages = computed(() =>
  Math.max(1, Math.ceil(remoteNewsTotal.value / NEWS_PAGE_SIZE)),
)

const news = computed<NewsDisplay[]>(() =>
  remoteNews.value.length ? remoteNews.value : newsItems.map(n => ({
    key: n.id,
    titleZh: n.titleZh, titleEn: n.titleEn, titleJa: '',
    summaryZh: n.summaryZh, summaryEn: n.summaryEn, summaryJa: '',
    sourceZh: n.sourceZh, sourceEn: n.sourceEn,
    timeZh: n.timeZh, timeEn: n.timeEn, timeJa: n.timeZh,
    hot: n.hot ?? false,
    url: '',
  })),
)

/** 相对时间文案(中/英/日) */
function relativeTime(published: Date): { zh: string; en: string; ja: string } {
  const diff = Date.now() - published.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 60) return { zh: '刚刚', en: 'just now', ja: 'たった今' }
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return { zh: `${hours}小时前`, en: `${hours}h ago`, ja: `${hours}時間前` }
  const days = Math.floor(hours / 24)
  if (days < 30) return { zh: `${days}天前`, en: `${days} days ago`, ja: `${days}日前` }
  const months = Math.floor(days / 30)
  return { zh: `${months}个月前`, en: `${months} months ago`, ja: `${months}ヶ月前` }
}

function fromRemote(item: HotspotItem): NewsDisplay {
  const t = relativeTime(new Date(item.published_at))
  return {
    key: `hot-${item.id}`,
    titleZh: item.title_zh, titleEn: item.title_en || '', titleJa: item.title_ja || '',
    summaryZh: item.summary_zh || item.title_zh, summaryEn: item.summary_en || '', summaryJa: item.summary_ja || '',
    sourceZh: item.source_zh, sourceEn: item.source_en || item.source_zh,
    timeZh: t.zh, timeEn: t.en, timeJa: t.ja,
    hot: item.hot,
    url: item.url,
  }
}

/** 按当前语言取标题/摘要/来源/时间(缺失字段逐级回退中文) */
function newsTitle(item: NewsDisplay): string {
  if (langStore.lang === 'en') return item.titleEn || item.titleZh
  if (langStore.lang === 'ja') return item.titleJa || item.titleZh
  return item.titleZh
}
function newsSummary(item: NewsDisplay): string {
  if (langStore.lang === 'en') return item.summaryEn || item.summaryZh
  if (langStore.lang === 'ja') return item.summaryJa || item.summaryZh
  return item.summaryZh
}
function newsSource(item: NewsDisplay): string {
  return langStore.lang === 'en' ? (item.sourceEn || item.sourceZh) : item.sourceZh
}
function newsTime(item: NewsDisplay): string {
  if (langStore.lang === 'en') return item.timeEn
  if (langStore.lang === 'ja') return item.timeJa
  return item.timeZh
}

/** 点击热点条目:有原文链接时新窗口打开 */
function openNews(item: NewsDisplay) {
  if (item.url) window.open(item.url, '_blank', 'noopener,noreferrer')
}

/** 加载指定页的热点数据(失败静默保持静态兜底) */
async function loadNews(page: number) {
  if (newsLoading.value) return
  newsLoading.value = true
  try {
    const res = await listHotspots(page, NEWS_PAGE_SIZE)
    if (res.items.length) {
      remoteNews.value = res.items.map(fromRemote)
      remoteNewsTotal.value = res.total
      newsPage.value = res.page
    }
  } catch {
    /* 保持静态兜底 */
  } finally {
    newsLoading.value = false
  }
}

/** 热点翻页 */
function changeNewsPage(page: number) {
  const target = Math.min(Math.max(1, page), newsTotalPages.value)
  if (target === newsPage.value) return
  loadNews(target)
}

function padRank(n: number): string {
  return String(n + 1).padStart(2, '0')
}

function goExplore() {
  router.push('/home/explore')
}

/** 接口失败时静默回退静态数据 */
onMounted(async () => {
  try {
    const page = await listScenics({ page: 1, page_size: 100 })
    const spots = [...page.items]
      .sort((a, b) => (b.score || 0) - (a.score || 0))
      .slice(0, RECOMMEND_COUNT)
      .map(fromRemoteScenic)
    if (spots.length >= RECOMMEND_COUNT) remoteSpots.value = spots
  } catch {
    /* 保持静态兜底 */
  }
  // 文旅热点:失败静默保持静态兜底
  loadNews(1)
})
</script>

<template>
  <div class="dashboard">
    <!-- ──── 欢迎横幅 ──── -->
    <HomeBanner
      :eyebrow="langStore.lang === 'zh' ? '四川省 · 成都' : langStore.lang === 'ja' ? '四川省 · 成都' : 'Chengdu · Sichuan'"
      :title="langStore.t('home.welcome')"
      :subtitle="langStore.t('home.welcomeSubtitle')"
      watermark="蜀"
    />

    <!-- ──── 今日蜀签(接口失败/无数据时整体隐藏) ──── -->
    <section v-if="daily" class="daily" aria-label="今日蜀签">
      <div class="daily-card">
        <span class="daily-card__deco" aria-hidden="true">🍵</span>
        <span class="daily-card__watermark" aria-hidden="true">蜀</span>
        <header class="daily-card__head">
          <h2 class="daily-card__title">{{ langStore.t('daily.title') }}</h2>
          <span class="daily-card__category">{{ dailyCategoryLabel }}</span>
        </header>
        <p class="daily-card__content">{{ dailyContent }}</p>
        <footer class="daily-card__foot">
          <button v-if="daily.related_spot_id" type="button" class="daily-card__spot" @click="goDailySpot">
            {{ langStore.t('daily.viewSpot') }}
            <span class="daily-card__arrow" aria-hidden="true">→</span>
          </button>
          <button type="button" class="daily-card__swap" :disabled="dailyLoading" @click="swapDaily">
            {{ langStore.t('daily.swap') }}
          </button>
        </footer>
      </div>
    </section>

    <!-- ──── 节气蜀俗(无应季美食/接口失败时整体隐藏) ──── -->
    <section v-if="solarVisible" class="solar" aria-label="节气蜀俗">
      <div class="solar-card">
        <span class="solar-card__watermark" aria-hidden="true">{{ solarTerm.name }}</span>
        <header class="solar-card__head">
          <h2 class="solar-card__title">{{ langStore.t('solarTerm.title') }}</h2>
          <span class="solar-card__term">{{ solarTermLabel }}</span>
        </header>
        <p class="solar-card__content">{{ solarBanner }}</p>
        <div v-if="solarFoods.length > 0" class="solar-card__foods">
          <span class="solar-card__foods-label">{{ langStore.t('solarTerm.seasonalFood') }}</span>
          <div class="solar-card__list">
            <button
              v-for="f in solarFoods"
              :key="f.id"
              type="button"
              class="solar-food"
              @click="goSolarFood(f.id)"
            >
              <img
                v-if="solarFoodImage(f)"
                class="solar-food__img"
                :src="solarFoodImage(f)"
                :alt="solarFoodName(f)"
                loading="lazy"
              />
              <span class="solar-food__name">{{ solarFoodName(f) }}</span>
              <span class="solar-food__arrow" aria-hidden="true">→</span>
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- ──── 数字足迹护照入口 ──── -->
    <section class="passport-entry">
      <button type="button" class="passport-entry__card" @click="goPassport">
        <span class="passport-entry__body">
          <span class="passport-entry__title">{{ langStore.t('passport.title') }}</span>
          <span class="passport-entry__desc">{{ langStore.t('passport.hero') }}</span>
        </span>
        <span class="passport-entry__badge">
          {{ langStore.t('passport.cardBadge', { count: stampCountValue }) }}
        </span>
        <span class="passport-entry__cta">
          {{ langStore.t('passport.collect') }}
          <span class="passport-entry__arrow" aria-hidden="true">→</span>
        </span>
      </button>
    </section>

    <!-- ──── 推荐景点 ──── -->
    <section class="recommend">
      <header class="section-head">
        <div>
          <h2 class="section-title">{{ langStore.t('home.recommendedTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('home.recommendedSubtitle') }}</p>
        </div>
        <button class="section-head__more" @click="goExplore">
          {{ langStore.t('home.recommendedSeeAll') }}
          <span class="section-head__arrow">→</span>
        </button>
      </header>

      <div class="recommend__grid">
        <article
          v-for="(spot, index) in recommended"
          :key="spot.key"
          class="recommend-card"
          tabindex="0"
          @click="goExplore"
          @keyup.enter="goExplore"
        >
          <div class="recommend-card__top">
            <span class="recommend-card__rank">{{ padRank(index) }}</span>
            <span class="recommend-card__rating">
              <AppIcon name="star" :size="12" filled />
              {{ spot.rating }}
            </span>
          </div>
          <!-- 真实图片:接口返回有效 URL 时展示,加载失败回退首字水印占位 -->
          <img
            v-if="spot.image"
            class="recommend-card__photo"
            :src="spot.image"
            :alt="langStore.lang === 'zh' || langStore.lang === 'ja' ? spot.nameZh : spot.nameEn"
            loading="lazy"
            decoding="async"
            referrerpolicy="no-referrer"
            @error="onImgError(spot)"
          />
          <div v-else class="recommend-card__placeholder" aria-hidden="true">
            <span class="recommend-card__shu">{{ spot.nameZh.charAt(0) }}</span>
          </div>
          <div class="recommend-card__body">
            <h3 class="recommend-card__title">
              {{ langStore.lang === 'zh' || langStore.lang === 'ja' ? spot.nameZh : spot.nameEn }}
            </h3>
            <p class="recommend-card__desc">
              {{ langStore.lang === 'zh' || langStore.lang === 'ja' ? spot.descZh : spot.descEn }}
            </p>
            <div class="recommend-card__tags">
              <span v-for="tag in spot.tags.slice(0, 3)" :key="tag" class="recommend-card__tag">{{ tag }}</span>
            </div>
          </div>
        </article>
      </div>
    </section>

    <!-- ──── 文旅热点 ──── -->
    <section class="news">
      <header class="section-head">
        <div>
          <h2 class="section-title">{{ langStore.t('home.newsTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('home.newsSubtitle') }}</p>
        </div>
      </header>

      <ul class="news__list">
        <li
          v-for="item in news"
          :key="item.key"
          class="news__item"
          :class="{ 'news__item--link': item.url }"
          @click="openNews(item)"
        >
          <span v-if="item.hot" class="news__hot">
            {{ langStore.t('home.newsHot') }}
          </span>
          <div class="news__body">
            <h3 class="news__title">{{ newsTitle(item) }}</h3>
            <p class="news__summary">{{ newsSummary(item) }}</p>
            <div class="news__meta">
              <span class="news__source">{{ newsSource(item) }}</span>
              <span class="news__dot">·</span>
              <span class="news__time">{{ newsTime(item) }}</span>
            </div>
          </div>
        </li>
      </ul>

      <!-- 热点分页(仅远端数据超过一页时展示) -->
      <div v-if="remoteNewsTotal > NEWS_PAGE_SIZE" class="news__pager">
        <button
          class="news__pager-btn"
          :disabled="newsPage <= 1 || newsLoading"
          aria-label="上一页"
          @click="changeNewsPage(newsPage - 1)"
        >‹</button>
        <span class="news__pager-info">{{ newsPage }} / {{ newsTotalPages }}</span>
        <button
          class="news__pager-btn"
          :disabled="newsPage >= newsTotalPages || newsLoading"
          aria-label="下一页"
          @click="changeNewsPage(newsPage + 1)"
        >›</button>
      </div>
    </section>

    <!-- ──── 成都冷知识 ──── -->
    <section class="facts">
      <header class="section-head">
        <div>
          <h2 class="section-title">{{ langStore.t('home.factsTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('home.factsSubtitle') }}</p>
        </div>
      </header>

      <div class="facts__grid">
        <article v-for="fact in funFacts" :key="fact.id" class="fact-card">
          <span class="fact-card__icon" aria-hidden="true">
            <AppIcon :name="fact.icon" :size="28" />
          </span>
          <h3 class="fact-card__title">{{ langStore.lang === 'zh' ? fact.titleZh : fact.titleEn }}</h3>
          <p class="fact-card__content">{{ langStore.lang === 'zh' ? fact.contentZh : fact.contentEn }}</p>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: var(--max-width);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-12);
}

/* ========================================
   天气
   ======================================== */
.dashboard__weather {
  max-width: 480px;
  margin: 0 auto;
}

/* ========================================
   区块标题
   ======================================== */
.section-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-8);
}

.section-head__more {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  padding-bottom: 2px;
  border-bottom: 1px solid transparent;
  transition: color var(--transition-fast), border-color var(--transition-fast);
  flex-shrink: 0;
}

.section-head__more:hover {
  color: var(--color-gold);
  border-bottom-color: var(--color-gold-dark);
}

.section-head__arrow {
  transition: transform var(--transition-fast);
}

.section-head__more:hover .section-head__arrow {
  transform: translateX(4px);
}

/* ========================================
   今日蜀签(宣纸底 + 竹青色系)
   ======================================== */
.daily-card {
  position: relative;
  overflow: hidden;
  padding: var(--space-6) var(--space-8);
  background:
    radial-gradient(ellipse 70% 120% at 100% 0%, rgba(45, 106, 79, 0.08) 0%, transparent 60%),
    linear-gradient(150deg, #faf6ec 0%, #f1f5ee 100%);
  border: 1px solid color-mix(in srgb, #2d6a4f 28%, transparent);
  border-radius: var(--radius-lg);
}

.daily-card__watermark {
  position: absolute;
  right: calc(var(--space-8) + 40px);
  bottom: -18px;
  font-family: var(--font-display);
  font-size: 96px;
  font-weight: 900;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1px color-mix(in srgb, #2d6a4f 18%, transparent);
  user-select: none;
  pointer-events: none;
}

.daily-card__deco {
  position: absolute;
  top: var(--space-4);
  right: var(--space-5);
  font-size: var(--text-2xl);
  line-height: 1;
  user-select: none;
}

.daily-card__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.daily-card__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: #2d6a4f;
  letter-spacing: var(--tracking-wide);
}

.daily-card__category {
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-cinnabar);
  padding: 2px 10px;
  border: 1px solid color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-cinnabar) 5%, transparent);
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
}

.daily-card__content {
  position: relative;
  max-width: 42em;
  font-family: var(--font-display);
  font-size: var(--text-base);
  color: #33463c;
  line-height: var(--leading-relaxed);
  letter-spacing: var(--tracking-wide);
}

.daily-card__foot {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-5);
  margin-top: var(--space-4);
}

.daily-card__spot {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-sm);
  font-weight: 600;
  color: #2d6a4f;
  letter-spacing: var(--tracking-wide);
  border-bottom: 1px solid transparent;
  transition: border-color var(--transition-fast);
}

.daily-card__spot:hover {
  border-bottom-color: #2d6a4f;
}

.daily-card__arrow {
  transition: transform var(--transition-fast);
}

.daily-card__spot:hover .daily-card__arrow {
  transform: translateX(4px);
}

.daily-card__swap {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  padding: 4px 14px;
  border: 1px solid color-mix(in srgb, #2d6a4f 35%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, #ffffff 55%, transparent);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.daily-card__swap:hover:not(:disabled) {
  color: #2d6a4f;
  border-color: #2d6a4f;
}

.daily-card__swap:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ========================================
   节气蜀俗(宣纸底 + 竹青色系,与今日蜀签协调)
   ======================================== */
.solar-card {
  position: relative;
  overflow: hidden;
  padding: var(--space-6) var(--space-8);
  background:
    radial-gradient(ellipse 70% 120% at 0% 0%, rgba(45, 106, 79, 0.08) 0%, transparent 60%),
    linear-gradient(150deg, #f3f7ef 0%, #faf6ec 100%);
  border: 1px solid color-mix(in srgb, #2d6a4f 28%, transparent);
  border-radius: var(--radius-lg);
}

.solar-card__watermark {
  position: absolute;
  right: var(--space-6);
  bottom: -12px;
  font-family: var(--font-display);
  font-size: 84px;
  font-weight: 900;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1px color-mix(in srgb, #2d6a4f 16%, transparent);
  user-select: none;
  pointer-events: none;
}

.solar-card__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.solar-card__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: #2d6a4f;
  letter-spacing: var(--tracking-wide);
}

.solar-card__term {
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-cinnabar);
  padding: 2px 10px;
  border: 1px solid color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-cinnabar) 5%, transparent);
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
}

.solar-card__content {
  position: relative;
  max-width: 42em;
  font-family: var(--font-display);
  font-size: var(--text-base);
  color: #33463c;
  line-height: var(--leading-relaxed);
  letter-spacing: var(--tracking-wide);
}

.solar-card__foods {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px dashed color-mix(in srgb, #2d6a4f 25%, transparent);
}

.solar-card__foods-label {
  flex-shrink: 0;
  font-size: var(--text-sm);
  font-weight: 600;
  color: #2d6a4f;
  letter-spacing: var(--tracking-wide);
}

.solar-card__list {
  display: flex;
  gap: var(--space-3);
  overflow-x: auto;
  scrollbar-width: none;
}

.solar-card__list::-webkit-scrollbar {
  display: none;
}

.solar-food {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
  padding: var(--space-1) var(--space-4) var(--space-1) var(--space-1);
  border: 1px solid color-mix(in srgb, #2d6a4f 20%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, #ffffff 60%, transparent);
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.solar-food:hover {
  border-color: #2d6a4f;
  transform: translateY(-2px);
}

.solar-food__img {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-full);
  object-fit: cover;
  background: #eee7d9;
}

.solar-food__name {
  font-size: var(--text-sm);
  font-weight: 600;
  color: #33463c;
  letter-spacing: var(--tracking-wide);
}

.solar-food__arrow {
  font-size: var(--text-xs);
  color: #2d6a4f;
  transition: transform var(--transition-fast);
}

.solar-food:hover .solar-food__arrow {
  transform: translateX(3px);
}

/* ========================================
   数字足迹护照入口
   ======================================== */
.passport-entry__card {
  display: flex;
  align-items: center;
  gap: var(--space-5);
  width: 100%;
  padding: var(--space-6) var(--space-8);
  text-align: left;
  background:
    radial-gradient(ellipse 60% 90% at 8% 50%, rgba(45, 106, 79, 0.10) 0%, transparent 70%),
    var(--color-surface);
  border: 1px solid color-mix(in srgb, #2d6a4f 35%, transparent);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: border-color var(--transition-base), transform var(--transition-base),
    box-shadow var(--transition-base);
}

.passport-entry__card:hover {
  border-color: #2d6a4f;
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(45, 106, 79, 0.18), var(--shadow-lg);
}

.passport-entry__icon {
  display: flex;
  align-items: center;
  font-size: var(--text-3xl);
  line-height: 1;
  flex-shrink: 0;
}

.passport-entry__body {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
  flex: 1;
}

.passport-entry__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 700;
  color: #2d6a4f;
  letter-spacing: var(--tracking-wide);
}

.passport-entry__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.passport-entry__badge {
  flex-shrink: 0;
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-cinnabar);
  padding: 4px 12px;
  border: 1px solid color-mix(in srgb, var(--color-cinnabar) 45%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-cinnabar) 6%, transparent);
  white-space: nowrap;
}

.passport-entry__cta {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  flex-shrink: 0;
  font-size: var(--text-sm);
  color: #2d6a4f;
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
}

.passport-entry__arrow {
  transition: transform var(--transition-fast);
}

.passport-entry__card:hover .passport-entry__arrow {
  transform: translateX(4px);
}

/* ========================================
   推荐景点
   ======================================== */
.recommend__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.recommend-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: all var(--transition-base);
  cursor: pointer;
  outline: none;
}

.recommend-card:hover,
.recommend-card:focus-visible {
  border-color: var(--color-gold-dark);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px var(--color-gold-glow), var(--shadow-lg);
}

.recommend-card__top {
  position: absolute;
  z-index: 2;
  top: var(--space-3);
  left: var(--space-3);
  right: var(--space-3);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.recommend-card__rank {
  font-family: var(--font-en-display);
  font-size: var(--text-sm);
  font-weight: 700;
  color: var(--color-gold);
  letter-spacing: var(--tracking-widest);
  background: color-mix(in srgb, var(--color-bg) 70%, transparent);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  border: 1px solid color-mix(in srgb, var(--color-gold) 35%, transparent);
  border-radius: var(--radius-sm);
  padding: 2px 8px;
}

.recommend-card__rating {
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  color: var(--color-gold);
  background: color-mix(in srgb, var(--color-bg) 70%, transparent);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  border: 1px solid color-mix(in srgb, var(--color-gold) 35%, transparent);
  border-radius: var(--radius-full);
  padding: 2px 10px;
}

.recommend-card__photo {
  display: block;
  width: 100%;
  aspect-ratio: 16 / 9;
  object-fit: cover;
  background: var(--color-bg-alt);
}

.recommend-card__placeholder {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(ellipse 60% 60% at 50% 30%, color-mix(in srgb, var(--color-gold) 10%, transparent) 0%, transparent 70%),
    var(--color-bg-alt);
}

.recommend-card__shu {
  font-family: var(--font-display);
  font-size: clamp(56px, 8vw, 88px);
  font-weight: 900;
  color: transparent;
  -webkit-text-stroke: 1px color-mix(in srgb, var(--color-gold) 45%, transparent);
  user-select: none;
}

.recommend-card__body {
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  flex: 1;
}

.recommend-card__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.recommend-card__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.recommend-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: auto;
  padding-top: var(--space-2);
}

.recommend-card__tag {
  font-size: var(--text-xs);
  padding: 2px 10px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   文旅热点
   ======================================== */
.news__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.news__item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-4);
  padding: var(--space-5);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: border-color var(--transition-base), transform var(--transition-base);
}

.news__item:hover {
  border-color: var(--color-gold-dark);
  transform: translateX(4px);
}

/* 有原文链接的条目可点击跳转 */
.news__item--link {
  cursor: pointer;
}

/* ========================================
   热点分页
   ======================================== */
.news__pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  margin-top: var(--space-5);
}

.news__pager-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  font-size: var(--text-lg);
  color: var(--color-text-secondary);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.news__pager-btn:hover:not(:disabled) {
  color: var(--color-gold);
  border-color: var(--color-gold-dark);
}

.news__pager-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.news__pager-info {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  min-width: 48px;
  text-align: center;
}

.news__hot {
  flex-shrink: 0;
  font-family: var(--font-display);
  font-size: var(--text-xs);
  font-weight: 700;
  color: var(--color-cinnabar);
  border: 1px solid var(--color-cinnabar);
  border-radius: var(--radius-sm);
  padding: 2px 8px;
  letter-spacing: var(--tracking-widest);
  margin-top: 2px;
}

.news__body {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.news__title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: var(--leading-normal);
}

.news__summary {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

.news__meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.news__source {
  color: var(--color-gold-dark);
}

.news__dot {
  opacity: 0.6;
}

/* ========================================
   成都冷知识
   ======================================== */
.facts__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.fact-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  transition: all var(--transition-base);
}

.fact-card:hover {
  border-color: var(--color-gold-dark);
  box-shadow: 0 8px 24px var(--color-gold-glow);
}

.fact-card__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-gold);
}

.fact-card__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.fact-card__content {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

/* ========================================
   Responsive
   ======================================== */
@media (max-width: 1024px) {
  .recommend__grid,
  .facts__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .dashboard {
    gap: var(--space-10);
  }
  .recommend__grid,
  .facts__grid {
    grid-template-columns: 1fr;
  }
  .dashboard__weather {
    max-width: 100%;
  }
  .passport-entry__card {
    padding: var(--space-5);
    gap: var(--space-3);
  }
  .passport-entry__desc,
  .passport-entry__cta {
    display: none;
  }
}
</style>
