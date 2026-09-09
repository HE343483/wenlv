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
  </div>
</template>

<style scoped>
.lang-switch {
  display: inline-flex;
  align-items: center;
  padding: 2px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface) 60%, transparent);
}

.lang-switch__btn {
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-full);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.lang-switch__btn:hover {
  color: var(--color-gold-dark);
}

.lang-switch__btn--active {
  background: var(--color-gold);
  color: var(--color-bg);
}
</style>
