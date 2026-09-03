<script setup lang="ts">
/**
 * RoutesPage.vue — 路线规划 (静态版式 Mockup)
 * 功能：精品路线 Tab + 自定义路线 Tab，左侧路线卡片/表单，右侧地图区
 * 说明：当前全部使用本地静态数据（travelRoutes / scenicSpots）做版式，
 *       后期接入后端 API 与百度地图 JS API 后替换数据源与渲染逻辑。
 */
import { ref, computed } from 'vue'
import { useLanguageStore } from '@/stores/language'
import { travelRoutes, scenicSpots, getScenicSpotById } from '@/data/chengdu'
import type { TravelRoute, ScenicSpot } from '@/types'

const langStore = useLanguageStore()

/* ── Tab ── */
type RoutesTab = 'prebuilt' | 'custom'
const activeTab = ref<RoutesTab>('prebuilt')

/* ── 选中的精品路线 (默认第一条) ── */
const selectedRoute = ref<TravelRoute>(travelRoutes[0]!)

function pickRoute(route: TravelRoute) {
  selectedRoute.value = route
}

/* 当前路线途经景点（补齐 spot 详情，便于展示站名） */
const selectedStops = computed(() =>
  selectedRoute.value.stops
    .map(stop => ({
      ...stop,
      spot: getScenicSpotById(stop.spotId),
    }))
    .filter(s => !!s.spot)
)

/* 精品路线卡片精简信息 */
function routeStopsSummary(route: TravelRoute): string {
  return route.stops
    .map(stop => {
      const spot = getScenicSpotById(stop.spotId)
      if (!spot) return ''
      return langStore.lang === 'zh' ? spot.nameZh : spot.nameEn
    })
    .filter(Boolean)
    .join(' → ')
}

/* ── 自定义路线 ── */
const startSpotId = ref('')
const destSpotId = ref('')
const myLocation = ref(false)

const spotOptions = computed<ScenicSpot[]>(() => scenicSpots)

function toggleMyLocation() {
  myLocation.value = !myLocation.value
  if (myLocation.value) {
    startSpotId.value = '' // 起点交给定位
  }
}

function planRoute() {
  // 后期：调用百度地图/后端路径规划 API
  if (myLocation.value || startSpotId.value) {
    // TODO: 接入 API 后替换为真实规划逻辑与状态提示
    console.log('planRoute →', { start: startSpotId.value, dest: destSpotId.value })
  }
}
</script>

<template>
  <div class="routes-page">
    <!-- ──── HERO ──── -->
    <section class="routes-hero">
      <div class="routes-hero__content">
        <p class="routes-hero__eyebrow">
          <span>◈</span>
          {{ langStore.lang === 'zh' ? 'Tianfu Journey' : '天府之旅' }}
          <span>◈</span>
        </p>
        <h1 class="routes-hero__title">{{ langStore.t('routes.title') }}</h1>
        <p class="routes-hero__subtitle">{{ langStore.t('routes.subtitle') }}</p>
      </div>
    </section>

    <!-- ──── TAB 切换 ──── -->
    <section class="routes-tabs container">
      <div class="routes-tabs__bar">
        <button
          class="routes-tabs__btn"
          :class="{ 'routes-tabs__btn--active': activeTab === 'prebuilt' }"
          @click="activeTab = 'prebuilt'"
        >
          <span class="routes-tabs__icon">🗺</span>
          {{ langStore.t('routes.tabPrebuilt') }}
        </button>
        <button
          class="routes-tabs__btn"
          :class="{ 'routes-tabs__btn--active': activeTab === 'custom' }"
          @click="activeTab = 'custom'"
        >
          <span class="routes-tabs__icon">✏️</span>
          {{ langStore.t('routes.tabCustom') }}
        </button>
      </div>
    </section>

    <!-- ════════════ 精品路线 ════════════ -->
    <section v-if="activeTab === 'prebuilt'" class="routes-prebuilt">
      <!-- 头部说明 -->
      <header class="routes-prebuilt__head container">
        <div>
          <h2 class="section-title">{{ langStore.t('routes.prebuiltTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('routes.prebuiltSubtitle') }}</p>
        </div>
      </header>

      <div class="routes-prebuilt__layout container">
        <!-- 左：路线卡片列表 -->
        <aside class="routes-prebuilt__list">
          <article
            v-for="route in travelRoutes"
            :key="route.id"
            class="route-card"
            :class="{ 'route-card--active': selectedRoute.id === route.id }"
            @click="pickRoute(route)"
          >
            <div class="route-card__head">
              <div>
                <h3 class="route-card__name">
                  {{ langStore.lang === 'zh' ? route.nameZh : route.nameEn }}
                </h3>
                <p class="route-card__meta">
                  {{ routeStopsSummary(route) }}
                </p>
              </div>
              <span class="route-card__chevron">›</span>
            </div>

            <div class="route-card__badges">
              <span class="route-card__badge">
                <b>{{ langStore.t('routes.duration') }}</b>
                {{ langStore.lang === 'zh' ? route.durationZh : route.durationEn }}
              </span>
              <span class="route-card__badge">
                <b>{{ langStore.t('routes.theme') }}</b>
                {{ langStore.lang === 'zh' ? route.themeZh : route.themeEn }}
              </span>
              <span class="route-card__badge">
                <b>{{ langStore.t('routes.level') }}</b>
                {{ langStore.lang === 'zh' ? route.levelZh : route.levelEn }}
              </span>
            </div>
          </article>
        </aside>

        <!-- 右：地图 + 行程 -->
        <div class="routes-prebuilt__detail">
          <!-- 地图占位（接入百度地图后替换为真实地图渲染） -->
          <div class="route-map">
            <div class="route-map__placeholder">
              <!-- 虚拟站点连线示意 (静态版式) -->
              <div class="route-map__stops">
                <div
                  v-for="(stop, i) in selectedStops"
                  :key="stop.spotId"
                  class="route-map__stop"
                  :style="{ '--i': i }"
                >
                  <span class="route-map__dot" />
                  <span class="route-map__label">
                    {{ langStore.lang === 'zh' ? stop.spot!.nameZh : stop.spot!.nameEn }}
                  </span>
                  <span class="route-map__note">
                    {{ langStore.lang === 'zh' ? stop.noteZh : stop.noteEn }}
                  </span>
                </div>
              </div>
              <!-- 连接线 -->
              <svg class="route-map__line" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
                <path
                  d="M 10 88 Q 30 60 50 50 T 90 14"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="0.4"
                  stroke-dasharray="2 1.5"
                  stroke-linecap="round"
                />
              </svg>
              <span class="route-map__pin">📍</span>
              <p class="route-map__hint">
                {{ langStore.t('routes.mapNoKey') }}
              </p>
            </div>
          </div>

          <!-- 行程安排 -->
          <div class="route-detail">
            <div class="route-detail__head">
              <h3 class="route-detail__title">
                {{ langStore.lang === 'zh' ? selectedRoute.nameZh : selectedRoute.nameEn }}
              </h3>
              <div class="route-detail__badges">
                <span class="route-detail__badge">
                  {{ langStore.lang === 'zh' ? selectedRoute.durationZh : selectedRoute.durationEn }}
                </span>
                <span class="route-detail__badge">
                  {{ langStore.lang === 'zh' ? selectedRoute.levelZh : selectedRoute.levelEn }}
                </span>
              </div>
            </div>

            <p class="route-detail__summary">
              {{ langStore.lang === 'zh' ? selectedRoute.summaryZh : selectedRoute.summaryEn }}
            </p>

            <!-- 行程步骤 -->
            <ol class="route-detail__stops">
              <li
                v-for="stop in selectedStops"
                :key="stop.spotId"
                class="route-stop"
              >
                <span class="route-stop__rail">
                  <span class="route-stop__dot" />
                </span>
                <div class="route-stop__body">
                  <span class="route-stop__note">
                    {{ langStore.lang === 'zh' ? stop.noteZh : stop.noteEn }}
                  </span>
                  <span class="route-stop__name">
                    {{ langStore.lang === 'zh' ? stop.spot!.nameZh : stop.spot!.nameEn }}
                  </span>
                </div>
              </li>
            </ol>

            <!-- 亮点 -->
            <div class="route-detail__highlights">
              <span
                v-for="hl in (langStore.lang === 'zh' ? selectedRoute.highlightsZh : selectedRoute.highlightsEn)"
                :key="hl"
                class="route-detail__highlight"
              >
                ✦ {{ hl }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ════════════ 自定义路线 ════════════ -->
    <section v-else class="routes-custom">
      <header class="routes-custom__head container">
        <div>
          <h2 class="section-title">{{ langStore.t('routes.customTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('routes.customSubtitle') }}</p>
        </div>
      </header>

      <div class="routes-custom__layout container">
        <!-- 左：规划表单 -->
        <aside class="routes-custom__panel">
          <!-- 起点 -->
          <label class="route-field">
            <span class="route-field__key">
              <span class="route-field__dot route-field__dot--start" />
              {{ langStore.t('routes.start') }}
            </span>
            <button
              class="route-field__locate"
              :class="{ 'route-field__locate--on': myLocation }"
              @click="toggleMyLocation"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="12" cy="12" r="3"/>
                <path d="M12 2v3M12 19v3M2 12h3M19 12h3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M19.1 4.9L17 7M7 17l-2.1 2.1"/>
              </svg>
              {{ myLocation ? langStore.t('routes.locationSuccess') : langStore.t('routes.myLocation') }}
            </button>
          </label>

          <div class="route-field">
            <select
              v-model="startSpotId"
              class="route-field__select"
              :disabled="myLocation"
            >
              <option value="" disabled>
                {{ myLocation
                  ? langStore.t('routes.locationSuccess')
                  : langStore.t('routes.startPlaceholder') }}
              </option>
              <option v-for="s in spotOptions" :key="s.id" :value="s.id">
                {{ langStore.lang === 'zh' ? s.nameZh : s.nameEn }}
              </option>
            </select>
          </div>

          <!-- 终点 -->
          <div class="route-field">
            <span class="route-field__key">
              <span class="route-field__dot route-field__dot--dest" />
              {{ langStore.t('routes.destination') }}
            </span>
            <select
              v-model="destSpotId"
              class="route-field__select"
            >
              <option value="" disabled>{{ langStore.t('routes.destinationPlaceholder') }}</option>
              <option v-for="s in spotOptions" :key="s.id" :value="s.id">
                {{ langStore.lang === 'zh' ? s.nameZh : s.nameEn }}
              </option>
            </select>
          </div>

          <!-- 规划按钮 -->
          <button class="route-field__plan" :disabled="!destSpotId" @click="planRoute">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M21 3L3 10.5l7.5 2.5L13 20.5 21 3z"/>
              <path d="M10.5 13l10-10"/>
            </svg>
            {{ langStore.t('routes.plan') }}
          </button>

          <!-- API 对接说明（占位提示） -->
          <p class="route-field__hint">{{ langStore.t('routes.hint') }}</p>
        </aside>

        <!-- 右：地图 -->
        <div class="routes-custom__map">
          <div class="route-map">
            <div class="route-map__placeholder route-map__placeholder--custom">
              <svg class="route-map__line" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
                <path
                  d="M 20 78 Q 40 55 60 45 T 82 20"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="0.4"
                  stroke-dasharray="2 1.5"
                  stroke-linecap="round"
                />
              </svg>
              <span class="route-map__pin">📍</span>
              <p class="route-map__hint">
                {{ langStore.t('routes.mapNoKey') }}
              </p>
            </div>
          </div>
        </div>
    </div>
    </section>
  </div>
</template>

<style scoped>
/* ========================================
   HERO
   ======================================== */
.routes-hero {
  position: relative;
  text-align: center;
  padding: var(--space-10) 0 var(--space-8);
}

.routes-hero__content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-6) var(--space-4);
}

.routes-hero__eyebrow {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.routes-hero__title {
  font-family: var(--font-display);
  font-size: var(--text-4xl);
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.1;
}

.routes-hero__subtitle {
  font-family: var(--font-body);
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  font-weight: 300;
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   TABS
   ======================================== */
.routes-tabs {
  padding-bottom: var(--space-8);
}

.routes-tabs__bar {
  display: inline-flex;
  gap: var(--space-2);
  padding: var(--space-1);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
}

.routes-tabs__btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-muted);
  padding: var(--space-2) var(--space-5);
  border-radius: var(--radius-full);
  transition: all var(--transition-base);
  white-space: nowrap;
}

.routes-tabs__btn:hover {
  color: var(--color-gold);
}

.routes-tabs__btn--active {
  background: var(--color-gold);
  color: var(--color-text-inverse);
}

.routes-tabs__btn--active:hover {
  color: var(--color-text-inverse);
}

.routes-tabs__icon {
  font-size: var(--text-base);
  line-height: 1;
}

/* ========================================
   区块头部
   ======================================== */
.routes-prebuilt__head,
.routes-custom__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

/* ========================================
   双栏布局
   ======================================== */
.routes-prebuilt__layout,
.routes-custom__layout {
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: var(--space-6);
  align-items: start;
  padding-bottom: var(--space-16);
}

/* ========================================
   精品路线 — 卡片列表
   ======================================== */
.routes-prebuilt__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  max-height: 720px;
  overflow-y: auto;
  padding-right: var(--space-2);
}

.route-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  cursor: pointer;
  transition: all var(--transition-base);
  text-align: left;
}

.route-card:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px var(--color-gold-glow);
}

.route-card--active {
  border-color: var(--color-gold);
  box-shadow: inset 0 0 0 1px var(--color-gold), 0 6px 20px var(--color-gold-glow);
}

.route-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.route-card__name {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  margin-bottom: var(--space-2);
}

.route-card__meta {
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-wide);
  line-height: var(--leading-normal);
}

.route-card__chevron {
  font-size: var(--text-2xl);
  color: var(--color-text-muted);
  transition: color var(--transition-fast), transform var(--transition-fast);
  flex-shrink: 0;
  line-height: 1;
}

.route-card:hover .route-card__chevron,
.route-card--active .route-card__chevron {
  color: var(--color-gold);
}

.route-card--active .route-card__chevron {
  transform: translateX(3px);
}

.route-card__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.route-card__badge {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  padding: 3px 10px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  letter-spacing: var(--tracking-wide);
}

.route-card__badge b {
  font-weight: 500;
  color: var(--color-text-muted);
  margin-right: var(--space-1);
}

/* ========================================
   地图占位
   ======================================== */
.route-map {
  position: relative;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.route-map__placeholder {
  position: relative;
  aspect-ratio: 16 / 8;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  background:
    radial-gradient(ellipse 55% 70% at 50% 45%, color-mix(in srgb, var(--color-gold) 7%, transparent) 0%, transparent 70%),
    repeating-linear-gradient(0deg, transparent 0 49px, color-mix(in srgb, var(--color-border) 50%, transparent) 49px 50px),
    repeating-linear-gradient(90deg, transparent 0 49px, color-mix(in srgb, var(--color-border) 50%, transparent) 49px 50px),
    var(--color-bg-alt);
}

.route-map__placeholder--custom {
  aspect-ratio: 16 / 8;
}

.route-map__pin {
  font-size: var(--text-4xl);
  filter: drop-shadow(0 0 16px var(--color-gold-glow));
  animation: pin-float 2.6s ease-in-out infinite;
  z-index: 2;
}

@keyframes pin-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

.route-map__hint {
  position: relative;
  z-index: 2;
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  text-align: center;
  background: color-mix(in srgb, var(--color-bg) 62%, transparent);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  padding: var(--space-2) var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  max-width: 86%;
}

.route-map__line {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  color: color-mix(in srgb, var(--color-gold) 55%, transparent);
  z-index: 1;
}

/* 虚拟站点 (静态示意，接入地图后删除) */
.route-map__stops {
  position: absolute;
  inset: 0;
  z-index: 2;
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: calc(var(--space-8) + 12px);
}

.route-map__stop {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.route-map__dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--color-gold);
  border: 3px solid color-mix(in srgb, var(--color-bg) 70%, transparent);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-gold) 45%, transparent), 0 0 14px var(--color-gold-glow);
  flex-shrink: 0;
}

.route-map__stop:nth-child(2n) .route-map__dot {
  background: var(--color-cinnabar);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-cinnabar) 40%, transparent);
}

.route-map__stop:nth-child(2n) {
  align-self: flex-end;
  flex-direction: row-reverse;
  text-align: right;
}

.route-map__label {
  font-family: var(--font-display);
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  background: color-mix(in srgb, var(--color-bg) 78%, transparent);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  border: 1px solid var(--color-border);
  padding: 2px 10px;
  border-radius: var(--radius-full);
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
}

.route-map__note {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
}

/* ========================================
   行程详情
   ======================================== */
.route-detail {
  margin-top: var(--space-6);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.route-detail__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.route-detail__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.route-detail__badges {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.route-detail__badge {
  font-size: var(--text-xs);
  padding: 3px 10px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-gold-dark);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-wide);
}

.route-detail__summary {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

/* 行程步骤 */
.route-detail__stops {
  display: flex;
  flex-direction: column;
}

.route-stop {
  display: flex;
  gap: var(--space-4);
}

.route-stop__rail {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  width: 12px;
}

.route-stop__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--color-gold);
  border: 2px solid color-mix(in srgb, var(--color-gold) 35%, transparent);
  box-shadow: 0 0 8px var(--color-gold-glow);
  margin-top: 6px;
}

.route-stop:not(:last-child) .route-stop__rail::after {
  content: '';
  flex: 1;
  width: 1px;
  background: linear-gradient(180deg, var(--color-gold-dark) 0%, transparent 100%);
  opacity: 0.4;
  min-height: 20px;
}

.route-stop__body {
  display: flex;
  align-items: baseline;
  gap: var(--space-4);
  padding: var(--space-3) 0;
  flex-wrap: wrap;
}

.route-stop__note {
  font-size: var(--text-xs);
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
  min-width: 96px;
}

.route-stop__name {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

/* 亮点 */
.route-detail__highlights {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  border-top: 1px solid var(--color-border);
  padding-top: var(--space-4);
}

.route-detail__highlight {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  padding: 3px 12px;
  border-radius: var(--radius-full);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   自定义路线 — 表单
   ======================================== */
.routes-custom__panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.route-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.route-field__key {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.route-field__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.route-field__dot--start {
  background: var(--color-sage);
  box-shadow: 0 0 8px color-mix(in srgb, var(--color-sage) 50%, transparent);
}

.route-field__dot--dest {
  background: var(--color-cinnabar);
  box-shadow: 0 0 8px color-mix(in srgb, var(--color-cinnabar) 50%, transparent);
}

.route-field__select {
  width: 100%;
  appearance: none;
  -webkit-appearance: none;
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  outline: none;
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.route-field__select:focus {
  border-color: var(--color-gold);
  box-shadow: 0 0 0 2px var(--color-gold-glow);
}

.route-field__select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.route-field__select option {
  background: var(--color-surface);
  color: var(--color-text-primary);
}

.route-field__locate {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  border: 1px dashed var(--color-border-light);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
  align-self: flex-start;
}

.route-field__locate:hover {
  color: var(--color-gold);
  border-color: var(--color-gold-dark);
}

.route-field__locate--on {
  color: var(--color-gold);
  border-color: var(--color-gold);
  border-style: solid;
  background: color-mix(in srgb, var(--color-gold) 8%, transparent);
}

.route-field__plan {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-inverse);
  background: var(--color-gold);
  padding: var(--space-3);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.route-field__plan:hover:not(:disabled) {
  background: var(--color-gold-light);
  box-shadow: 0 4px 16px var(--color-gold-glow);
}

.route-field__plan:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.route-field__hint {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  text-align: center;
  line-height: var(--leading-relaxed);
}

/* ========================================
   Responsive
   ======================================== */
@media (max-width: 1024px) {
  .routes-prebuilt__layout,
  .routes-custom__layout {
    grid-template-columns: 1fr;
  }
  .routes-prebuilt__list {
    max-height: none;
    overflow: visible;
    display: grid;
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .routes-hero__title {
    font-size: var(--text-3xl);
  }
  .routes-tabs__bar {
    display: flex;
    width: 100%;
  }
  .routes-tabs__btn {
    flex: 1;
    justify-content: center;
  }
  .routes-prebuilt__list {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .routes-prebuilt__layout,
  .routes-custom__layout {
    padding-bottom: var(--space-10);
  }
  .route-map__placeholder {
    aspect-ratio: 4 / 3;
  }
  .route-map__stops {
    display: none; /* 移动端隐藏虚拟站点连线 */
  }
  .route-stop__body {
    flex-direction: column;
    gap: var(--space-1);
  }
  .route-stop__note {
    min-width: 0;
  }
}
</style>
