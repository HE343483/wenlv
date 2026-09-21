<script setup lang="ts">
/**
 * RoutesPage.vue — 精选路线页
 * 数据来自后端 /api/routes(爬虫采集维基百科简介与配图,图片存 OSS),
 * 接口失败时回退本地静态数据(curatedRoutes)。
 * 布局:图片主导的编辑式图文行(非卡片网格);点击整行打开详情弹窗,
 * 弹窗内站点时间线带配图与简介,并可一键「用 AI 生成同款行程」跳 /trip 预填表单。
 */
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { useUserStore } from '@/stores/user'
import { listRoutes, parseRouteStops, type RouteItem, type RouteStop } from '@/api/content'
import { pickDesc } from '@/utils/storyI18n'
import { curatedRoutes, type Locale } from '@/data/curatedRoutes'
import dictZh from '@/locales/zh'
import dictEn from '@/locales/en'
import dictJa from '@/locales/ja'
import HomeBanner from '@/components/HomeBanner.vue'

const route = useRoute()
const router = useRouter()
const langStore = useLanguageStore()
const userStore = useUserStore()

const locale = computed<Locale>(() => (langStore.lang === 'en' ? 'en' : langStore.lang === 'ja' ? 'ja' : 'zh'))

/** 语言字典(读取标签数组等非字符串值) */
const dicts = { zh: dictZh, en: dictEn, ja: dictJa } as const

/** 按点号路径读字典值,取不到返回 undefined */
function dictValue(key: string): unknown {
  let result: unknown = dicts[locale.value]
  for (const part of key.split('.')) {
    if (result && typeof result === 'object' && part in (result as Record<string, unknown>)) {
      result = (result as Record<string, unknown>)[part]
    } else {
      return undefined
    }
  }
  return result
}

/** ── 页面统一视图模型(DB 与静态兜底都归一化到这里) ── */
interface RouteStopView {
  name: string
  desc: string
  /** 站点简介多语种故事(LLM 生成,可能未生成,展示时经 pickDesc 回落 desc) */
  desc_en?: string
  desc_ja?: string
  image: string
}
interface RouteView {
  key: string
  /** 后端路线数字ID(静态兜底数据无此字段,此时不显示收藏按钮) */
  id?: number
  image: string
  title: string
  desc: string
  /** 路线简介多语种故事(LLM 生成,可能未生成,展示时经 pickDesc 回落 desc) */
  desc_en?: string
  desc_ja?: string
  days: number
  tags: string[]
  /** 路线主题:玩法路线为中文分类词,culture 为蜀文化叙事路线(前端按此加角标) */
  theme?: string
  stops: RouteStopView[]
}

/** 站点三语名称(DB 站点与静态兜底站点两种结构) */
function stopName(stop: RouteStop | { zh: string; en: string; ja: string }): string {
  const l = locale.value
  if ('name_zh' in stop) {
    return (l === 'en' ? stop.name_en || stop.name_zh : l === 'ja' ? stop.name_ja || stop.name_zh : stop.name_zh) || ''
  }
  return stop[l]
}

/** DB 路线 → 视图模型 */
function fromDB(item: RouteItem): RouteView {
  const title = locale.value === 'en' ? item.title_en || item.title_zh : locale.value === 'ja' ? item.title_ja || item.title_zh : item.title_zh
  const interestMap = (dictValue('routes.interestMap') ?? {}) as Record<string, string>
  const tags = (item.interests || '')
    .split(',')
    .map(s => s.trim())
    .filter(Boolean)
    .map(v => interestMap[v] ?? v)
  return {
    key: item.route_key,
    id: item.id,
    image: item.cover_image || '',
    title,
    desc: item.description || '',
    desc_en: item.description_en || '',
    desc_ja: item.description_ja || '',
    days: item.days || 1,
    tags,
    theme: (item.theme || '').trim(),
    stops: parseRouteStops(item.stops).map(s => ({
      name: stopName(s),
      desc: s.desc || '',
      desc_en: s.desc_en || '',
      desc_ja: s.desc_ja || '',
      image: s.image || '',
    })),
  }
}

/** 静态兜底路线 → 视图模型(接口不可用时展示) */
function fromStatic(key: string): RouteView {
  const r = curatedRoutes.find(x => x.id === key)
  if (!r) {
    return { key, image: '', title: key, desc: '', days: 1, tags: [], stops: [] }
  }
  const itemText = (dictValue(`routes.items.${r.id}`) ?? {}) as { title?: string; desc?: string; tags?: string[] }
  return {
    key: r.id,
    image: r.image,
    title: itemText.title ?? '',
    desc: itemText.desc ?? '',
    days: r.days,
    tags: itemText.tags ?? [],
    stops: r.stops.map(s => ({ name: stopName(s), desc: '', image: '' })),
  }
}

/** 展示列表:优先 DB,失败/为空回退静态 */
const routes = ref<RouteView[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await listRoutes({ page_size: 20 })
    const items = (res.items || []).slice().sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0))
    if (items.length > 0) {
      routes.value = items.map(fromDB)
    }
  } catch {
    // 接口不可用时静默回退静态数据
  } finally {
    if (routes.value.length === 0) {
      routes.value = ['classic', 'panda', 'food', 'culture'].map(fromStatic)
    }
    loading.value = false
    // 等待 Vue 把路线数据渲染成真实 DOM 后再启动进场动画,
    // 否则 querySelectorAll('.route-row') 查不到元素,行会永远停留在 opacity:0
    await nextTick()
    setupReveal()
    openRouteFromQuery()
  }
})

/** ── 滚动进入视口时的浮现动画 ── */
const rootRef = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null

function setupReveal() {
  if (!rootRef.value) return
  observer?.disconnect() // 筛选切换重跑时先释放旧观察者
  const rows = Array.from(rootRef.value.querySelectorAll('.route-row'))
  // 浏览器不支持 IntersectionObserver 时直接全部显示
  if (typeof IntersectionObserver === 'undefined') {
    rows.forEach(el => el.classList.add('is-visible'))
    return
  }
  observer = new IntersectionObserver(
    entries => {
      for (const e of entries) {
        if (e.isIntersecting) {
          ;(e.target as HTMLElement).classList.add('is-visible')
          observer?.unobserve(e.target)
        }
      }
    },
    { threshold: 0.15 },
  )
  rows.forEach(el => observer?.observe(el))
  // 兜底:动画因任何原因未触发时强制显示,避免列表空白
  window.setTimeout(() => {
    rows.forEach(el => el.classList.add('is-visible'))
  }, 1500)
}

onBeforeUnmount(() => observer?.disconnect())

/** ── 主题筛选:全部 / 玩法主题(theme 中文分类词) / 文化主题(theme=culture) ── */
const activeFilter = ref<string>('all')

/** theme 值 → 展示标签(culture 走 themeCultural 文案,其余查 themeMap,缺失原样显示) */
function themeLabel(theme: string): string {
  if (theme === 'culture') return langStore.t('routes.themeCultural')
  const map = (dictValue('routes.themeMap') ?? {}) as Record<string, string>
  return map[theme] ?? theme
}

/** 筛选选项:全部 + 按出现顺序的玩法主题 + 文化主题(固定在末位) */
const filters = computed(() => {
  const list: Array<{ value: string; label: string }> = [{ value: 'all', label: langStore.t('routes.filterAll') }]
  const seen = new Set<string>()
  for (const r of routes.value) {
    const th = r.theme || ''
    if (!th || th === 'culture' || seen.has(th)) continue
    seen.add(th)
    list.push({ value: th, label: themeLabel(th) })
  }
  if (routes.value.some(r => r.theme === 'culture')) {
    list.push({ value: 'culture', label: langStore.t('routes.themeCultural') })
  }
  return list
})

const filteredRoutes = computed(() => {
  if (activeFilter.value === 'all') return routes.value
  return routes.value.filter(r => r.theme === activeFilter.value)
})

// 筛选切换后行会重新挂载,需重新触发进场浮现,否则新行停留在 opacity:0
watch(filteredRoutes, async () => {
  await nextTick()
  setupReveal()
})

/** ── 详情弹窗 ── */
const activeRoute = ref<RouteView | null>(null)

function openDetail(r: RouteView) {
  activeRoute.value = r
}

function closeDetail() {
  activeRoute.value = null
  // 清除 URL 上的定位参数,保证再次从收藏页进入同一路线时能重新打开弹窗
  if (route.query.route) router.replace({ name: 'internal-routes' })
}

/** 从收藏页跳转过来时(?route={route_key})自动打开对应路线详情弹窗 */
function openRouteFromQuery() {
  const wantKey = String(route.query.route ?? '')
  if (!wantKey) return
  const hit = routes.value.find(r => r.key === wantKey)
  if (hit) activeRoute.value = hit
}

/** ── 路线收藏(后端 target_type=route;静态兜底路线无数字ID,不显示收藏按钮) ── */
function isRouteFav(r: RouteView): boolean {
  return r.id != null && userStore.isFavorite(`route-${r.id}`)
}

function toggleRouteFav(r: RouteView) {
  if (r.id == null) return
  userStore.toggleFavorite(`route-${r.id}`).catch((err: unknown) => {
    // 接口失败已回滚本地状态，此处提示用户
    alert(err instanceof Error ? err.message : '操作失败')
  })
}

/** 进入 AI 行程页并预填该路线 */
function goAiTrip(r?: RouteView) {
  closeDetail()
  router.push(r ? { path: '/trip', query: { prefill: r.key } } : '/trip')
}

/** 站点序号 01/02… */
function pad(n: number): string {
  return String(n + 1).padStart(2, '0')
}
</script>

<template>
  <div ref="rootRef" class="routes-page">
    <!-- ──── HERO ──── -->
    <HomeBanner
      :eyebrow="langStore.lang === 'zh' ? 'Tianfu Journey' : '天府之旅'"
      :title="langStore.t('routes.title')"
      :subtitle="langStore.t('routes.subtitle')"
      watermark="路"
    />

    <!-- ──── 主题筛选:全部 / 玩法主题 / 文化主题 ──── -->
    <nav v-if="filters.length > 1" class="routes-page__filters container" aria-label="route themes">
      <button
        v-for="f in filters"
        :key="f.value"
        type="button"
        class="routes-page__filter"
        :class="{ 'routes-page__filter--active': activeFilter === f.value }"
        :aria-pressed="activeFilter === f.value"
        @click="activeFilter = f.value"
      >
        {{ f.label }}
      </button>
    </nav>

    <!-- ──── 路线编辑式图文行 ──── -->
    <section class="routes-page__list container">
      <article
        v-for="(r, i) in filteredRoutes"
        :key="r.key"
        class="route-row"
        :class="{ 'route-row--flip': i % 2 === 1 }"
        role="button"
        tabindex="0"
        @click="openDetail(r)"
        @keydown.enter.prevent="openDetail(r)"
      >
        <div class="route-row__media">
          <img v-if="r.image" :src="r.image" :alt="r.title" loading="lazy" />
          <span v-if="r.theme === 'culture'" class="route-row__culture-badge">
            <span aria-hidden="true">🧭</span>
            {{ langStore.t('routes.themeCultural') }}
          </span>
          <span class="route-row__days">
            <b>{{ r.days }}</b>
            {{ langStore.t('routes.daysUnit') }}
          </span>
        </div>

        <div class="route-row__body">
          <span v-if="r.tags.length" class="route-row__theme">{{ r.tags.join(' · ') }}</span>
          <h3 class="route-row__title">{{ r.title }}</h3>
          <p class="route-row__desc">{{ pickDesc(r, langStore.lang) }}</p>

          <div class="route-row__stops">
            <span v-for="(s, j) in r.stops.slice(0, 5)" :key="j" class="route-row__stop">
              <i>{{ pad(j) }}</i>
              {{ s.name }}
            </span>
          </div>

          <div class="route-row__actions">
            <button
              v-if="r.id != null"
              type="button"
              class="route-row__fav"
              :class="{ 'route-row__fav--active': isRouteFav(r) }"
              :title="isRouteFav(r) ? langStore.t('scenic.favorited') : langStore.t('scenic.favorite')"
              :aria-label="isRouteFav(r) ? langStore.t('scenic.favorited') : langStore.t('scenic.favorite')"
              :aria-pressed="isRouteFav(r)"
              @click.stop="toggleRouteFav(r)"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round">
                <path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z" :fill="isRouteFav(r) ? 'currentColor' : 'none'"/>
              </svg>
              {{ isRouteFav(r) ? langStore.t('scenic.favorited') : langStore.t('scenic.favorite') }}
            </button>
            <button type="button" class="route-row__detail" @click.stop="openDetail(r)">
              {{ langStore.t('routes.viewDetail') }}
            </button>
            <button type="button" class="route-row__ai" @click.stop="goAiTrip(r)">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 3l1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9z" />
              </svg>
              {{ langStore.t('routes.generateAi') }}
            </button>
          </div>
        </div>
      </article>
    </section>

    <!-- ──── AI 定制入口 ──── -->
    <section class="routes-page__custom container">
      <div class="ai-entry">
        <div class="ai-entry__text">
          <h2 class="ai-entry__title">{{ langStore.t('routes.customTitle') }}</h2>
          <p class="ai-entry__desc">{{ langStore.t('routes.customDesc') }}</p>
        </div>
        <button class="ai-entry__cta" @click="goAiTrip()">
          {{ langStore.t('routes.customCta') }}
          <span class="ai-entry__cta-arrow">→</span>
        </button>
      </div>
    </section>

    <!-- ──── 路线详情弹窗 ──── -->
    <Teleport to="body">
      <Transition name="route-modal">
        <div v-if="activeRoute" class="route-modal" @click.self="closeDetail">
          <div class="route-modal__panel" role="dialog" aria-modal="true">
            <button class="route-modal__close" aria-label="close" @click="closeDetail">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                <path d="M6 6l12 12M18 6L6 18" />
              </svg>
            </button>

            <div class="route-modal__media">
              <img v-if="activeRoute.image" :src="activeRoute.image" :alt="activeRoute.title" />
              <div class="route-modal__media-text">
                <span class="route-modal__days">{{ activeRoute.days }} {{ langStore.t('routes.daysUnit') }}</span>
                <h3 class="route-modal__title">{{ activeRoute.title }}</h3>
              </div>
            </div>

            <div class="route-modal__body">
              <div class="route-modal__tags">
                <span v-for="tag in activeRoute.tags" :key="tag" class="route-modal__tag">{{ tag }}</span>
              </div>
              <p class="route-modal__desc">{{ pickDesc(activeRoute, langStore.lang) }}</p>

              <h4 class="route-modal__stops-title">{{ langStore.t('routes.stopsLabel') }}</h4>
              <ol class="route-modal__stops">
                <li v-for="(s, i) in activeRoute.stops" :key="i" class="route-modal__stop">
                  <span class="route-modal__stop-index">{{ pad(i) }}</span>
                  <div class="route-modal__stop-main">
                    <span class="route-modal__stop-name">{{ s.name }}</span>
                    <p v-if="pickDesc(s, langStore.lang)" class="route-modal__stop-desc">{{ pickDesc(s, langStore.lang) }}</p>
                  </div>
                  <img v-if="s.image" class="route-modal__stop-img" :src="s.image" :alt="s.name" loading="lazy" />
                </li>
              </ol>

              <button class="route-modal__cta" @click="goAiTrip(activeRoute)">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 3l1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9z" />
                  <path d="M18.5 15.5l.7 1.8 1.8.7-1.8.7-.7 1.8-.7-1.8-1.8-.7 1.8-.7z" />
                </svg>
                {{ langStore.t('routes.generateAi') }}
                <span class="route-modal__cta-arrow">→</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
/* ========================================
   主题筛选条:编辑式 pill 按钮
   ======================================== */
.routes-page__filters {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  padding-top: var(--space-8);
  padding-bottom: var(--space-8);
}

.routes-page__filter {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text-secondary);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.routes-page__filter:hover {
  border-color: var(--color-gold);
  color: var(--color-gold-dark);
}

.routes-page__filter--active {
  border-color: var(--color-gold);
  background: color-mix(in srgb, var(--color-gold) 12%, transparent);
  color: var(--color-gold-dark);
}

.route-row__culture-badge {
  position: absolute;
  top: var(--space-4);
  right: var(--space-4);
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-gold) 88%, black);
  backdrop-filter: blur(6px);
  color: var(--color-text-inverse);
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wide);
  z-index: 1;
}

/* ========================================
   路线编辑式图文行:大图 + 文字交替,图片主导
   ======================================== */
.routes-page__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-12);
  padding-bottom: var(--space-12);
}

.route-row {
  display: grid;
  grid-template-columns: minmax(0, 7fr) minmax(0, 5fr);
  gap: var(--space-8);
  align-items: center;
  cursor: pointer;
  /* 进场浮现 */
  opacity: 0;
  transform: translateY(28px);
  transition: opacity 0.7s ease, transform 0.7s ease;
}

.route-row.is-visible {
  opacity: 1;
  transform: none;
}

/* 偶数行图文对调 */
.route-row--flip .route-row__media {
  order: 2;
}

.route-row--flip .route-row__body {
  order: 1;
}

.route-row__media {
  position: relative;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  border-radius: var(--radius-lg);
}

.route-row__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 1.2s ease;
}

.route-row:hover .route-row__media img {
  transform: scale(1.04);
}

.route-row__days {
  position: absolute;
  top: var(--space-4);
  left: var(--space-4);
  display: inline-flex;
  align-items: baseline;
  gap: 2px;
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-full);
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(6px);
  color: #fff;
  font-size: var(--text-xs);
}

.route-row__days b {
  font-family: var(--font-en-display);
  font-size: var(--text-base);
}

.route-row__body {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  min-width: 0;
}

.route-row__theme {
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-wider);
}

.route-row__title {
  font-family: var(--font-display);
  font-size: var(--text-3xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: var(--leading-tight);
}

.route-row__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 站点编号预览 */
.route-row__stops {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.route-row__stop {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.route-row__stop i {
  font-style: normal;
  font-family: var(--font-en-display);
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
}

.route-row__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  margin-top: var(--space-2);
}

/* 路线收藏:与查看详情同款的胶囊按钮,收藏后变实心描边 */
.route-row__fav {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-5);
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.route-row__fav:hover {
  border-color: var(--color-gold);
  color: var(--color-gold-dark);
}

.route-row__fav--active {
  border-color: color-mix(in srgb, var(--color-gold) 55%, transparent);
  color: var(--color-gold-dark);
}

.route-row__detail {
  padding: var(--space-2) var(--space-5);
  border-radius: var(--radius-full);
  border: 1px solid color-mix(in srgb, var(--color-gold) 55%, transparent);
  color: var(--color-gold-dark);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.route-row__detail:hover {
  background: color-mix(in srgb, var(--color-gold) 10%, transparent);
  border-color: var(--color-gold);
}

.route-row__ai {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-5);
  border-radius: var(--radius-full);
  background: var(--color-gold);
  color: var(--color-text-inverse);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.route-row__ai:hover {
  background: var(--color-gold-light);
  box-shadow: 0 4px 16px var(--color-gold-glow);
}

/* ========================================
   AI 定制入口:上下细线分隔的编辑式横带(无卡片感)
   ======================================== */
.routes-page__custom {
  padding-bottom: var(--space-16);
}

.ai-entry {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-8);
  padding: var(--space-8) 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.ai-entry__text {
  flex: 1;
  min-width: 0;
}

.ai-entry__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.ai-entry__desc {
  margin-top: var(--space-2);
  max-width: 46em;
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

.ai-entry__cta {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
  padding: var(--space-3) var(--space-6);
  border-radius: var(--radius-full);
  border: 1px solid var(--color-gold);
  background: transparent;
  color: var(--color-gold-dark);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-base);
  white-space: nowrap;
}

.ai-entry__cta:hover {
  background: var(--color-gold);
  color: var(--color-text-inverse);
  box-shadow: 0 6px 24px var(--color-gold-glow);
  transform: translateY(-2px);
}

.ai-entry__cta-arrow {
  transition: transform var(--transition-fast);
}

.ai-entry__cta:hover .ai-entry__cta-arrow {
  transform: translateX(4px);
}

/* ========================================
   路线详情弹窗
   ======================================== */
.route-modal {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
  background: rgba(0, 0, 0, 0.6);
}

.route-modal__panel {
  position: relative;
  width: min(720px, 100%);
  max-height: 88vh;
  overflow-y: auto;
  background: var(--color-surface-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
}

.route-modal__close {
  position: absolute;
  top: var(--space-4);
  right: var(--space-4);
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  transition: background var(--transition-fast);
}

.route-modal__close:hover {
  background: rgba(0, 0, 0, 0.7);
}

.route-modal__media {
  position: relative;
  aspect-ratio: 21 / 9;
  overflow: hidden;
}

.route-modal__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

/* 头图压暗渐变 + 标题叠加 */
.route-modal__media::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, transparent 30%, rgba(0, 0, 0, 0.72));
}

.route-modal__media-text {
  position: absolute;
  left: var(--space-6);
  right: var(--space-6);
  bottom: var(--space-5);
  z-index: 1;
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.route-modal__days {
  flex-shrink: 0;
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-full);
  background: rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(6px);
  color: #fff;
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wide);
}

.route-modal__title {
  font-family: var(--font-display);
  font-size: var(--text-2xl);
  font-weight: 700;
  color: #fff;
  letter-spacing: var(--tracking-wide);
}

.route-modal__body {
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.route-modal__tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.route-modal__tag {
  padding: 2px var(--space-3);
  border-radius: var(--radius-full);
  border: 1px solid color-mix(in srgb, var(--color-gold) 40%, transparent);
  color: var(--color-gold-dark);
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wide);
}

.route-modal__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

.route-modal__stops-title {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.route-modal__stops {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
}

.route-modal__stop {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: var(--space-4);
  padding: var(--space-4) 0;
}

/* 站点连线 */
.route-modal__stop:not(:last-child)::before {
  content: '';
  position: absolute;
  left: 17px;
  top: 54px;
  bottom: -4px;
  width: 1px;
  background: color-mix(in srgb, var(--color-gold) 35%, transparent);
}

.route-modal__stop-index {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 34px;
  height: 34px;
  border-radius: var(--radius-full);
  border: 1px solid color-mix(in srgb, var(--color-gold) 45%, transparent);
  color: var(--color-gold-dark);
  font-family: var(--font-en-display);
  font-size: var(--text-xs);
  background: var(--color-gold-glow);
}

.route-modal__stop-main {
  flex: 1;
  min-width: 0;
}

.route-modal__stop-name {
  display: block;
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.route-modal__stop-desc {
  margin-top: var(--space-1);
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.route-modal__stop-img {
  flex-shrink: 0;
  width: 88px;
  height: 64px;
  object-fit: cover;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
}

.route-modal__cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  align-self: center;
  gap: var(--space-3);
  margin-top: var(--space-2);
  padding: var(--space-3) var(--space-8);
  border-radius: var(--radius-full);
  background: var(--color-gold);
  color: var(--color-text-inverse);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-base);
}

.route-modal__cta:hover {
  background: var(--color-gold-light);
  box-shadow: 0 6px 24px var(--color-gold-glow);
  transform: translateY(-2px);
}

.route-modal__cta-arrow {
  transition: transform var(--transition-fast);
}

.route-modal__cta:hover .route-modal__cta-arrow {
  transform: translateX(4px);
}

/* 弹窗过渡 */
.route-modal-enter-active,
.route-modal-leave-active {
  transition: opacity var(--transition-base);
}

.route-modal-enter-active .route-modal__panel,
.route-modal-leave-active .route-modal__panel {
  transition: transform var(--transition-base), opacity var(--transition-base);
}

.route-modal-enter-from,
.route-modal-leave-to {
  opacity: 0;
}

.route-modal-enter-from .route-modal__panel,
.route-modal-leave-to .route-modal__panel {
  transform: translateY(16px) scale(0.98);
  opacity: 0;
}

/* 响应式 */
@media (max-width: 860px) {
  .route-row {
    grid-template-columns: 1fr;
    gap: var(--space-5);
  }
  /* 移动端恢复图上文下 */
  .route-row--flip .route-row__media {
    order: 0;
  }
  .route-row--flip .route-row__body {
    order: 0;
  }
  .route-row__title {
    font-size: var(--text-2xl);
  }
  .ai-entry {
    flex-direction: column;
    text-align: center;
  }
  .route-modal {
    padding: var(--space-3);
  }
}
</style>
