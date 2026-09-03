<script setup lang="ts">
/**
 * HomeDashboard.vue — 内部首页仪表盘
 * 内容: 欢迎横幅 + 天气 + 推荐景点 + 文旅热点 + 成都冷知识
 * 说明: 与探索页(ExplorePage)去重 — 不再展示区域筛选/全部景点网格，
 *       改为首页专属的精选推荐、热点新闻与趣味冷知识。
 */
import { useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import WeatherRowCell from '@/components/WeatherRowCell.vue'
import { getRecommendedSpots, newsItems, funFacts } from '@/data/chengdu'

const router = useRouter()
const langStore = useLanguageStore()

const recommended = getRecommendedSpots()

function padRank(n: number): string {
  return String(n + 1).padStart(2, '0')
}

function goExplore() {
  router.push('/home/explore')
}
</script>

<template>
  <div class="dashboard">
    <!-- ──── 欢迎横幅 ──── -->
    <section class="welcome">
      <div class="welcome__shu" aria-hidden="true">蜀</div>
      <div class="welcome__content">
        <p class="welcome__eyebrow">
          <span>◈</span>
          {{ langStore.lang === 'zh' ? '四川省 · 成都' : 'Chengdu · Sichuan' }}
          <span>◈</span>
        </p>
        <h1 class="welcome__title">{{ langStore.t('home.welcome') }}</h1>
        <p class="welcome__subtitle">{{ langStore.t('home.welcomeSubtitle') }}</p>
      </div>
    </section>

    <!-- ──── 天气 ──── -->
    <div class="dashboard__weather">
      <WeatherRowCell />
    </div>

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
          :key="spot.id"
          class="recommend-card"
          tabindex="0"
          @click="goExplore"
          @keyup.enter="goExplore"
        >
          <div class="recommend-card__top">
            <span class="recommend-card__rank">{{ padRank(index) }}</span>
            <span class="recommend-card__rating">{{ spot.rating }} ★</span>
          </div>
          <div class="recommend-card__placeholder" aria-hidden="true">
            <span class="recommend-card__shu">{{ spot.nameZh.charAt(0) }}</span>
          </div>
          <div class="recommend-card__body">
            <h3 class="recommend-card__title">
              {{ langStore.lang === 'zh' ? spot.nameZh : spot.nameEn }}
            </h3>
            <p class="recommend-card__desc">
              {{ langStore.lang === 'zh' ? spot.shortDescZh : spot.shortDescEn }}
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
          v-for="item in newsItems"
          :key="item.id"
          class="news__item"
        >
          <span v-if="item.hot" class="news__hot">
            {{ langStore.t('home.newsHot') }}
          </span>
          <div class="news__body">
            <h3 class="news__title">{{ langStore.lang === 'zh' ? item.titleZh : item.titleEn }}</h3>
            <p class="news__summary">{{ langStore.lang === 'zh' ? item.summaryZh : item.summaryEn }}</p>
            <div class="news__meta">
              <span class="news__source">{{ langStore.lang === 'zh' ? item.sourceZh : item.sourceEn }}</span>
              <span class="news__dot">·</span>
              <span class="news__time">{{ langStore.lang === 'zh' ? item.timeZh : item.timeEn }}</span>
            </div>
          </div>
        </li>
      </ul>
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
          <span class="fact-card__icon" aria-hidden="true">{{ fact.icon }}</span>
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
   欢迎横幅
   ======================================== */
.welcome {
  position: relative;
  overflow: hidden;
  text-align: center;
  padding: var(--space-16) var(--space-6) var(--space-12);
  border-bottom: 1px solid var(--color-border);
}

.welcome__shu {
  position: absolute;
  font-family: var(--font-display);
  font-size: clamp(160px, 28vw, 340px);
  font-weight: 900;
  line-height: 1;
  color: transparent;
  -webkit-text-stroke: 1px rgba(201, 169, 110, 0.12);
  user-select: none;
  pointer-events: none;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.welcome__content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}

.welcome__eyebrow {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.welcome__title {
  font-family: var(--font-display);
  font-size: var(--text-4xl);
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.1;
}

.welcome__subtitle {
  font-family: var(--font-body);
  font-size: var(--text-base);
  font-weight: 300;
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
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
  font-size: var(--text-3xl);
  line-height: 1;
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
  .welcome {
    padding: var(--space-10) var(--space-4) var(--space-8);
  }
  .welcome__title {
    font-size: var(--text-3xl);
  }
  .recommend__grid,
  .facts__grid {
    grid-template-columns: 1fr;
  }
  .dashboard__weather {
    max-width: 100%;
  }
}
</style>
