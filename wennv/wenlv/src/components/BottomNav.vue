<script setup lang="ts">
/**
 * BottomNav.vue — 底部导航栏
 * 4个标签: 首页 / 探索 / 文化 / 路线
 * 使用 router-link 实现路由切换
 */
import { useLanguageStore } from '@/stores/language'

const langStore = useLanguageStore()

const tabs = [
  { route: '/home/index', icon: 'home', key: 'bottomNav.home' },
  { route: '/home/explore', icon: 'explore', key: 'bottomNav.explore' },
  { route: '/home/culture', icon: 'culture', key: 'bottomNav.culture' },
  { route: '/home/routes', icon: 'routes', key: 'bottomNav.routes' },
] as const

function getIcon(name: string): string {
  const icons: Record<string, string> = {
    home:    '🏠',
    explore: '🔍',
    culture: '🏛',
    routes:  '🗺',
  }
  return icons[name] || '●'
}
</script>

<template>
  <nav class="bottom-nav">
    <div class="bottom-nav__inner">
      <router-link
        v-for="tab in tabs"
        :key="tab.route"
        :to="tab.route"
        class="bottom-nav__item"
        active-class="bottom-nav__item--active"
      >
        <span class="bottom-nav__icon">{{ getIcon(tab.icon) }}</span>
        <span class="bottom-nav__label">{{ langStore.t(tab.key) }}</span>
      </router-link>
    </div>
  </nav>
</template>

<style scoped>
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: var(--bottom-nav-height);
  background: var(--color-bottom-nav);
  border-top: 1px solid var(--color-border);
  z-index: 1000;
  transition: background var(--transition-base), border-color var(--transition-base);
}

.bottom-nav__inner {
  display: flex;
  align-items: center;
  justify-content: space-around;
  height: 100%;
  max-width: 600px;
  margin: 0 auto;
  padding: 0 var(--space-4);
}

.bottom-nav__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: var(--space-2) var(--space-4);
  color: var(--color-text-muted);
  text-decoration: none;
  transition: color var(--transition-fast);
  border-radius: var(--radius-md);
  position: relative;
}

.bottom-nav__item:hover {
  color: var(--color-text-secondary);
}

.bottom-nav__item--active {
  color: var(--color-gold);
}

.bottom-nav__item--active::after {
  content: '';
  position: absolute;
  top: -1px;
  left: 50%;
  transform: translateX(-50%);
  width: 24px;
  height: 2px;
  background: var(--color-gold);
  border-radius: 1px;
}

.bottom-nav__icon {
  font-size: var(--text-xl);
  line-height: 1;
}

.bottom-nav__label {
  font-size: 11px;
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
}
</style>
