<script setup lang="ts">
/** ExplorePage.vue — 探索成都 (占位) */
import { useLanguageStore } from '@/stores/language'
const langStore = useLanguageStore()
</script>

<template>
  <div class="placeholder-page">
    <div class="placeholder-page__card">
      <span class="placeholder-page__icon">🔍</span>
      <h1 class="placeholder-page__title">{{ langStore.t('nav.explore') }}</h1>
      <p class="placeholder-page__desc">{{ langStore.t('internal.exploreDesc') }}</p>
      <div class="placeholder-page__divider" />
      <p class="placeholder-page__hint">{{ langStore.t('common.loading') }}</p>
    </div>
  </div>
</template>

<style scoped>
.placeholder-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
}
.placeholder-page__card {
  text-align: center;
  padding: var(--space-12);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  max-width: 420px;
  width: 100%;
}
.placeholder-page__icon { font-size: 3rem; display: block; margin-bottom: var(--space-4); }
.placeholder-page__title {
  font-family: var(--font-display);
  font-size: var(--text-2xl);
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  margin-bottom: var(--space-3);
}
.placeholder-page__desc { color: var(--color-text-secondary); font-size: var(--text-sm); line-height: var(--leading-relaxed); }
.placeholder-page__divider { height: 1px; background: var(--color-border); margin: var(--space-6) auto; max-width: 60px; }
.placeholder-page__hint { color: var(--color-text-muted); font-size: var(--text-xs); letter-spacing: var(--tracking-wide); }
</style>
