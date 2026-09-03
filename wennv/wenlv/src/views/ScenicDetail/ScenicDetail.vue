<script setup lang="ts">
/**
 * ScenicDetail.vue — 景点详情页
 * 说明：当前为纯布局骨架（图片占位 / 文案占位），
 *       后期由后端接口按路由参数 :id 拉取景点详情数据填充。
 */
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { useThemeStore } from '@/stores/theme'

const route = useRoute()
const router = useRouter()
const langStore = useLanguageStore()
const themeStore = useThemeStore()

/* 路由参数：景点ID（后端数据接入后据此拉取详情） */
const scenicId = computed(() => String(route.params.id ?? ''))

function goBack() {
  router.back()
}

/* 占位标签 — 后期替换为接口返回的 tags */
const placeholderTags = ['文化', '地标', '美食']
const tagIndex = computed(() =>
  Math.abs([...scenicId.value].reduce((acc, ch) => acc + ch.charCodeAt(0), 0)) % placeholderTags.length
)

/* 占位图片区标题 — 后期替换为接口返回的名称 */
const placeholderName = computed(() =>
  langStore.lang === 'zh' ? `景点 · ${scenicId.value}` : `Scenic · ${scenicId.value}`
)
</script>

<template>
  <div class="detail-page">
    <!-- ──── 顶栏 ──── -->
    <header class="detail-topbar">
      <button class="detail-topbar__back" @click="goBack">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M15 18l-6-6 6-6"/>
        </svg>
      </button>

      <div class="detail-topbar__brand">
        <span class="detail-topbar__title">蜀韵·成都</span>
        <span class="detail-topbar__crumb">{{ langStore.t('scenicDetail.breadcrumb') }}</span>
      </div>

      <div class="detail-topbar__actions">
        <button
          class="detail-topbar__icon-btn"
          @click="themeStore.toggle()"
          :title="themeStore.theme === 'dark' ? '切换浅色' : '切换深色'"
        >
          <svg v-if="themeStore.theme === 'dark'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2"/>
          </svg>
        </button>
        <button class="detail-topbar__lang-btn" @click="langStore.toggle()">
          {{ langStore.t('nav.langSwitch') }}
        </button>
      </div>
    </header>

    <main class="detail-main">
      <!-- ──── HERO — 图片区（占位）──── -->
      <section class="detail-hero">
        <div class="detail-hero__media">
          <!-- 图片占位：后期替换为接口返回的景区封面图 -->
          <div class="detail-hero__placeholder">
            <div class="detail-hero__shu" aria-hidden="true">景</div>
            <span class="detail-hero__api-badge">{{ langStore.t('scenicDetail.imagePlaceholder') }}</span>
          </div>

          <!-- 名称浮层 -->
          <div class="detail-hero__overlay">
            <h1 class="detail-hero__title">{{ placeholderName }}</h1>
            <p class="detail-hero__en-title">{{ langStore.t('scenicDetail.loading') }}</p>
            <div class="detail-hero__tags">
              <span
                v-for="(t, i) in (langStore.lang === 'zh'
                  ? placeholderTags
                  : ['Culture', 'Landmark', 'Food'])"
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
              <span class="detail-score__num">—</span>
              <div class="detail-score__meta">
                <span class="detail-score__stars">★★★★★</span>
                <span class="detail-score__label">{{ langStore.t('scenic.rating') }}</span>
              </div>
            </div>
            <div class="detail-aside__row">
              <span class="detail-aside__key">{{ langStore.t('scenicDetail.district') }}</span>
              <span class="detail-aside__value">—</span>
            </div>
            <div class="detail-aside__row">
              <span class="detail-aside__key">{{ langStore.t('scenicDetail.visits') }}</span>
              <span class="detail-aside__value">—</span>
            </div>
            <div class="detail-aside__row">
              <span class="detail-aside__key">{{ langStore.t('scenicDetail.recommendTime') }}</span>
              <span class="detail-aside__value">—</span>
            </div>
          </aside>

          <!-- 右：简介 -->
          <div class="detail-summary__body">
            <div class="detail-summary__eyebrow">
              <span>◈</span>
              {{ langStore.t('scenicDetail.overview') }}
              <span>◈</span>
            </div>
            <p class="detail-summary__text">
              {{ langStore.t('scenicDetail.overviewPlaceholder') }}
            </p>
          </div>
        </div>
      </section>

      <!-- ──── 实用信息 ──── -->
      <section class="detail-section container">
        <div class="detail-info">
          <div class="detail-info__card">
            <span class="detail-info__icon">🕘</span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.openHours') }}</h3>
            <p class="detail-info__value">{{ langStore.t('common.loading') }}</p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon">🎫</span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.ticket') }}</h3>
            <p class="detail-info__value">{{ langStore.t('common.loading') }}</p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon">🚌</span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.transport') }}</h3>
            <p class="detail-info__value">{{ langStore.t('common.loading') }}</p>
          </div>
          <div class="detail-info__card">
            <span class="detail-info__icon">📍</span>
            <h3 class="detail-info__title">{{ langStore.t('scenicDetail.address') }}</h3>
            <p class="detail-info__value">{{ langStore.t('common.loading') }}</p>
          </div>
        </div>
      </section>

      <!-- ──── 图文详情 ──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('scenicDetail.detailTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('scenicDetail.detailSubtitle') }}</p>
        </header>

        <div class="detail-content">
          <!-- 长图占位块 — 后期替换为后端富文本/段落+配图 -->
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
              <h3 class="detail-content__caption">{{ langStore.t('common.loading') }}</h3>
              <p class="detail-content__para">{{ langStore.t('scenicDetail.paraPlaceholder') }}</p>
              <p class="detail-content__para">{{ langStore.t('scenicDetail.paraPlaceholder') }}</p>
            </div>
          </div>

          <div class="detail-content__row detail-content__row--reverse">
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
              <h3 class="detail-content__caption">{{ langStore.t('common.loading') }}</h3>
              <p class="detail-content__para">{{ langStore.t('scenicDetail.paraPlaceholder') }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- ──── 相册 ──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('scenicDetail.galleryTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('scenicDetail.gallerySubtitle') }}</p>
        </header>

        <div class="detail-gallery">
          <div v-for="n in 6" :key="n" class="detail-gallery__item">
            <div class="detail-gallery__img">
              <span class="detail-gallery__num">{{ String(n).padStart(2, '0') }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- ──── 地图（占位）──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('scenicDetail.mapTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('scenicDetail.mapSubtitle') }}</p>
        </header>

        <div class="detail-map">
          <div class="detail-map__placeholder">
            <span class="detail-map__pin">📍</span>
            <span class="detail-map__hint">{{ langStore.t('scenicDetail.mapPlaceholder') }}</span>
          </div>
        </div>
      </section>

      <!-- ──── 周边推荐 ──── -->
      <section class="detail-section container">
        <header class="detail-block-head">
          <h2 class="section-title">{{ langStore.t('scenicDetail.aroundTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('scenicDetail.aroundSubtitle') }}</p>
        </header>

        <div class="detail-around">
          <div v-for="n in 3" :key="n" class="detail-around__card">
            <div class="detail-around__img">
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
                <rect x="3" y="3" width="18" height="18" rx="2"/>
                <circle cx="8.5" cy="8.5" r="1.5"/>
                <path d="M21 15l-5-5L5 21"/>
              </svg>
            </div>
            <div class="detail-around__body">
              <h3 class="detail-around__name">{{ langStore.t('common.loading') }}</h3>
              <p class="detail-around__desc">{{ langStore.t('scenicDetail.nearbyPlaceholder') }}</p>
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

.detail-topbar__icon-btn {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
}

.detail-topbar__icon-btn:hover {
  border-color: var(--color-gold);
  color: var(--color-gold);
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
    radial-gradient(ellipse 70% 60% at 50% 35%, color-mix(in srgb, var(--color-gold) 10%, transparent) 0%, transparent 70%),
    linear-gradient(180deg, var(--color-surface-elevated), var(--color-bg-alt));
  border-bottom: 1px solid var(--color-border);
}

.detail-hero__shu {
  font-family: var(--font-display);
  font-size: clamp(120px, 22vw, 240px);
  font-weight: 900;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1px color-mix(in srgb, var(--color-gold) 40%, transparent);
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
  color: var(--color-gold);
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
  border-color: color-mix(in srgb, var(--color-gold) 55%, transparent);
  color: var(--color-gold);
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
  color: var(--color-gold);
  font-size: var(--text-sm);
  letter-spacing: 2px;
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
  color: var(--color-gold-dark);
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
  border-color: var(--color-gold-dark);
  transform: translateY(-2px);
}

.detail-info__icon {
  font-size: var(--text-2xl);
  line-height: 1;
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
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-gold) 8%, transparent) 0%, transparent 70%),
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
  border-color: var(--color-gold-dark);
  transform: translateY(-2px);
}

.detail-gallery__img {
  aspect-ratio: 4 / 3;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-gold) 8%, transparent) 0%, transparent 70%),
    var(--color-bg-alt);
}

.detail-gallery__num {
  font-family: var(--font-en-display);
  font-size: var(--text-3xl);
  font-weight: 700;
  color: transparent;
  -webkit-text-stroke: 1px color-mix(in srgb, var(--color-gold) 45%, transparent);
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
    radial-gradient(ellipse 50% 70% at 50% 50%, color-mix(in srgb, var(--color-gold) 6%, transparent) 0%, transparent 70%),
    repeating-linear-gradient(0deg, transparent 0 39px, color-mix(in srgb, var(--color-border) 45%, transparent) 39px 40px),
    repeating-linear-gradient(90deg, transparent 0 39px, color-mix(in srgb, var(--color-border) 45%, transparent) 39px 40px),
    var(--color-bg-alt);
}

.detail-map__pin {
  font-size: var(--text-4xl);
  filter: drop-shadow(0 0 16px var(--color-gold-glow));
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
   周边推荐
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
  border-color: var(--color-gold-dark);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px var(--color-gold-glow), var(--shadow-lg);
}

.detail-around__img {
  aspect-ratio: 16 / 9;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  background:
    radial-gradient(ellipse 60% 60% at 50% 40%, color-mix(in srgb, var(--color-gold) 8%, transparent) 0%, transparent 70%),
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
