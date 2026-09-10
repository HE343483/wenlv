<script setup lang="ts">
/**
 * BottomNav.vue — 底部导航栏
 * 6个标签: 首页 / 探索 / 美食 / 路线 / 收藏 / 我的
 * 收藏为独立页，个人资料与打卡照片在「我的」内完成
 */
import { useLanguageStore } from '@/stores/language'

const langStore = useLanguageStore()

const tabs = [
  { route: '/home/index', icon: 'home', key: 'bottomNav.home' },
  { route: '/home/explore', icon: 'explore', key: 'bottomNav.explore' },
  { route: '/home/food', icon: 'food', key: 'bottomNav.food' },
  { route: '/home/routes', icon: 'routes', key: 'bottomNav.routes' },
  { route: '/home/favorites', icon: 'favorites', key: 'bottomNav.favorites' },
  { route: '/home/profile', icon: 'mine', key: 'bottomNav.mine' },
] as const
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
        <span class="bottom-nav__icon">
          <!-- 首页 — 传统中式屋顶房屋 -->
          <svg v-if="tab.icon === 'home'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 9.5 L12 3 L21 9.5"/>
            <path d="M3 9.5 Q12 7 21 9.5"/>
            <rect x="5" y="9.5" width="14" height="12.5" rx="0.5"/>
            <rect x="10" y="14" width="4" height="8" rx="0.5"/>
          </svg>
          <!-- 探索 — 指南针 -->
          <svg v-else-if="tab.icon === 'explore'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 2 L16 12 L12 22 L8 12 Z"/>
            <line x1="12" y1="2" x2="12" y2="4"/>
            <line x1="12" y1="20" x2="12" y2="22"/>
            <line x1="2" y1="12" x2="4" y2="12"/>
            <line x1="20" y1="12" x2="22" y2="12"/>
          </svg>
          <!-- 美食 — 碗筷 -->
          <svg v-else-if="tab.icon === 'food'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 3 L4 9"/>
            <path d="M8 3 L8 9"/>
            <path d="M3 9 Q3 13 6 13 L6 21"/>
            <path d="M9 9 Q9 13 6 13"/>
            <path d="M15 3 C15 7 18 7 18 10 L18 21"/>
            <path d="M18 3 C18 6 21 6 21 9 L21 21"/>
          </svg>
          <!-- 路线 — 路线/地图 -->
          <svg v-else-if="tab.icon === 'routes'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="5" cy="5" r="2.5"/>
            <circle cx="19" cy="19" r="2.5"/>
            <path d="M7 7 L17 17"/>
            <path d="M7 7 L9 7 L9 11 L15 17 L17 17"/>
            <circle cx="12" cy="12" r="1.5" fill="currentColor"/>
          </svg>
          <!-- 收藏 — 心形 -->
          <svg v-else-if="tab.icon === 'favorites'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z"/>
          </svg>
          <!-- 我的 — 人形 -->
          <svg v-else-if="tab.icon === 'mine'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="8" r="4"/>
            <path d="M5 20c1.6-3.6 4.1-5 7-5s5.4 1.4 7 5"/>
          </svg>
        </span>
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
