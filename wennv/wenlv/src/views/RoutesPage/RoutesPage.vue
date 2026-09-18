<script setup lang="ts">
/**
 * RoutesPage.vue — AI 智能行程入口页
 * 原「路线规划」（精品路线/自定义路线 + 百度地图）功能与 AI 行程重合，已移除。
 * 本页仅保留页头横幅与入口按钮，点击后跳转到 AI 行程规划页（/trip）。
 */
import { useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import HomeBanner from '@/components/HomeBanner.vue'
import AppIcon from '@/components/AppIcon.vue'

const router = useRouter()
const langStore = useLanguageStore()

/** 进入 AI 行程规划页 */
function goAiTrip() {
  router.push('/trip')
}
</script>

<template>
  <div class="ai-trip-entry">
    <!-- ──── HERO ──── -->
    <HomeBanner
      :eyebrow="langStore.lang === 'zh' ? 'Tianfu Journey' : '天府之旅'"
      :title="langStore.t('routes.title')"
      :subtitle="langStore.t('routes.subtitle')"
      watermark="行"
    />

    <!-- ──── AI 行程入口 ──── -->
    <section class="ai-trip-entry__body container">
      <div class="ai-trip-entry__card">
        <span class="ai-trip-entry__icon" aria-hidden="true">
          <AppIcon name="star" :size="40" />
        </span>
        <h2 class="ai-trip-entry__title">{{ langStore.t('routes.entryTitle') }}</h2>
        <p class="ai-trip-entry__desc">{{ langStore.t('routes.entryDesc') }}</p>
        <button class="ai-trip-entry__cta" @click="goAiTrip">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 3l1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9z" />
            <path d="M18.5 15.5l.7 1.8 1.8.7-1.8.7-.7 1.8-.7-1.8-1.8-.7 1.8-.7z" />
          </svg>
          {{ langStore.t('routes.cta') }}
          <span class="ai-trip-entry__cta-arrow">→</span>
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* ========================================
   AI 行程入口
   ======================================== */
.ai-trip-entry__body {
  padding-bottom: var(--space-16);
}

.ai-trip-entry__card {
  max-width: 640px;
  margin: 0 auto;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-12) var(--space-8);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  text-align: center;
}

.ai-trip-entry__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  border-radius: var(--radius-full);
  background: var(--color-gold-glow);
  color: var(--color-gold);
}

.ai-trip-entry__title {
  font-family: var(--font-display);
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.ai-trip-entry__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  max-width: 480px;
}

/* 主行动点：金色实心按钮 */
.ai-trip-entry__cta {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
  margin-top: var(--space-4);
  padding: var(--space-4) var(--space-8);
  border-radius: var(--radius-full);
  background: var(--color-gold);
  color: var(--color-text-inverse);
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-base);
}

.ai-trip-entry__cta:hover {
  background: var(--color-gold-light);
  box-shadow: 0 6px 24px var(--color-gold-glow);
  transform: translateY(-2px);
}

.ai-trip-entry__cta-arrow {
  transition: transform var(--transition-fast);
}

.ai-trip-entry__cta:hover .ai-trip-entry__cta-arrow {
  transform: translateX(4px);
}

@media (max-width: 640px) {
  .ai-trip-entry__card {
    padding: var(--space-8) var(--space-5);
  }
}
</style>
