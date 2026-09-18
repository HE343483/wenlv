<script setup lang="ts">
/**
 * LanguageSwitch.vue — 语言切换分段器（中 | EN）
 * 首页 NavBar 与内部 HomeLayout 顶栏共用，保证全局语言按钮一致
 */
import { useLanguageStore } from '@/stores/language'

const langStore = useLanguageStore()
</script>

<template>
  <div class="lang-switch" role="group" :aria-label="langStore.t('nav.ariaLang')">
    <button
      type="button"
      class="lang-switch__btn"
      :class="{ 'lang-switch__btn--active': langStore.lang === 'zh' }"
      @click="langStore.setLang('zh')"
    >
      中
    </button>
    <button
      type="button"
      class="lang-switch__btn"
      :class="{ 'lang-switch__btn--active': langStore.lang === 'en' }"
      @click="langStore.setLang('en')"
    >
      EN
    </button>
    <button
      type="button"
      class="lang-switch__btn"
      :class="{ 'lang-switch__btn--active': langStore.lang === 'ja' }"
      @click="langStore.setLang('ja')"
    >
      日
    </button>
  </div>
</template>

<style scoped>
.lang-switch {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  padding: 2px;
  border: 1px solid color-mix(in srgb, var(--color-border) 55%, transparent);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface) 72%, transparent);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  box-shadow: var(--shadow-sm);
}

.lang-switch__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2.4rem;
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-secondary);
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-full);
  letter-spacing: var(--tracking-wide);
  transition: color var(--transition-fast), background var(--transition-fast), box-shadow var(--transition-fast);
}

.lang-switch__btn:hover {
  color: var(--color-gold-dark);
  background: color-mix(in srgb, var(--color-surface) 55%, transparent);
}

.lang-switch__btn--active {
  background: color-mix(in srgb, var(--color-surface) 92%, transparent);
  color: var(--color-text-primary);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--color-border) 70%, transparent);
}
</style>
