<script setup lang="ts">
/**
 * HomeLayout.vue — 登录后内部页面布局
 * 顶栏与公开首页 NavBar 保持一致的设计语言:
 *   左: 返回键 + 品牌(菱形徽标 + 中英文名)
 *   中: 快捷导航(首页/探索/美食/路线, 桌面端显示)
 *   右: 天气 + 语言切换(共用 LanguageSwitch) + 我的
 * 下: 子路由内容 + 底部导航
 */
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { useWeatherStore } from '@/stores/weather'
import WeatherTrigger from '@/components/WeatherTrigger.vue'
import WeatherPanel from '@/components/WeatherPanel.vue'
import LanguageSwitch from '@/components/LanguageSwitch.vue'
import PandaCursor from '@/components/PandaCursor.vue'

const route = useRoute()
const router = useRouter()
const langStore = useLanguageStore()
const weatherStore = useWeatherStore()
const { open } = storeToRefs(weatherStore)

function goBack() {
  router.push('/')
}

/* 中部快捷导航 — 与底部导航一致，桌面端在顶栏直达 */
const quickNavs = [
  { route: '/home/index', key: 'bottomNav.home' },
  { route: '/home/explore', key: 'bottomNav.explore' },
  { route: '/home/food', key: 'bottomNav.food' },
  { route: '/home/routes', key: 'bottomNav.routes' },
] as const

const isQuickActive = (target: string) => route.path === target

/* 天气下拉：触发点 ref + 面板 ref（面板 Teleport 到 body，用暴露的 getElement 判点击范围） */
const triggerRef = ref<InstanceType<typeof WeatherTrigger> | null>(null)
const panelRef = ref<InstanceType<typeof WeatherPanel> | null>(null)
</script>

<template>
  <div class="internal-layout">
    <PandaCursor />

    <!-- 顶栏 -->
    <header class="internal-topbar">
      <!-- ── 左：返回 + 品牌 ── -->
      <div class="internal-topbar__left">
        <button class="internal-topbar__back" :title="langStore.t('common.back')" @click="goBack">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M15 18l-6-6 6-6" />
          </svg>
        </button>

        <div class="internal-topbar__brand" @click="router.push('/home/index')">
          <svg class="internal-topbar__brand-mark" width="24" height="24" viewBox="0 0 26 26" fill="none">
            <rect x="6" y="6" width="14" height="14" transform="rotate(45 13 13)" stroke="currentColor" stroke-width="1.2" />
            <rect x="10.5" y="10.5" width="5" height="5" transform="rotate(45 13 13)" fill="currentColor" />
          </svg>
          <div class="internal-topbar__brand-text">
            <span class="internal-topbar__title">蜀韵·成都</span>
            <span class="internal-topbar__subtitle">Shu·Chengdu</span>
          </div>
        </div>
      </div>

      <!-- ── 中：快捷导航 ── -->
      <nav class="internal-topbar__nav" :aria-label="langStore.t('nav.ariaNav')">
        <RouterLink
          v-for="item in quickNavs"
          :key="item.route"
          :to="item.route"
          class="internal-topbar__nav-link"
          :class="{ 'internal-topbar__nav-link--active': isQuickActive(item.route) }"
        >
          {{ langStore.t(item.key) }}
        </RouterLink>
      </nav>

      <!-- ── 右：天气 + 语言 + 我的 ── -->
      <div class="internal-topbar__actions">
        <!-- 天气：锚点按钮（含温度区间） + 下拉面板 -->
        <WeatherTrigger ref="triggerRef" />
        <WeatherPanel v-if="open" ref="panelRef" :anchor="triggerRef" />

        <!-- 语言切换：与首页一致的分段式 中/EN -->
        <LanguageSwitch />

        <span class="internal-topbar__divider" aria-hidden="true" />

        <RouterLink to="/home/profile" class="internal-topbar__profile">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
            <circle cx="12" cy="8" r="3.5" />
            <path d="M5 20c1.6-3.6 4.1-5 7-5s5.4 1.4 7 5" />
          </svg>
          {{ langStore.t('bottomNav.mine') }}
        </RouterLink>
      </div>
    </header>

    <!-- 主内容区 (子路由) -->
    <main class="internal-main">
      <RouterView />
    </main>

  </div>
</template>

<style scoped>
.internal-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding-bottom: var(--bottom-nav-height);
}

/* ── 顶栏 ── */
.internal-topbar {
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
  transition: background var(--transition-base);
}

/* 底部金色渐变线 — 呼应公开首页 NavBar 的金线装饰 */
.internal-topbar::after {
  content: '';
  position: absolute;
  bottom: -1px;
  left: 50%;
  transform: translateX(-50%);
  width: 60%;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--color-gold), transparent);
  pointer-events: none;
}

/* ── 左：返回 + 品牌 ── */
.internal-topbar__left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.internal-topbar__back {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
  transition: all var(--transition-fast);
}

.internal-topbar__back:hover {
  border-color: var(--color-gold);
  color: var(--color-gold);
  box-shadow: 0 0 12px var(--color-gold-glow);
}

.internal-topbar__brand {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  cursor: pointer;
  user-select: none;
}

.internal-topbar__brand-mark {
  color: var(--color-gold);
  flex-shrink: 0;
}

.internal-topbar__brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.internal-topbar__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
}

.internal-topbar__subtitle {
  font-family: var(--font-en-display);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wider);
  text-transform: uppercase;
}

/* ── 中：快捷导航（绝对定位居中，与 NavBar 同款） ── */
.internal-topbar__nav {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  align-items: center;
  gap: var(--space-2);
  white-space: nowrap;
}

.internal-topbar__nav-link {
  position: relative;
  font-family: var(--font-display);
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  text-decoration: none;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.internal-topbar__nav-link:hover {
  color: var(--color-gold-dark);
  background: color-mix(in srgb, var(--color-gold) 8%, transparent);
}

/* 激活项：金色文字 + 下方短横线，与 NavBar 一致 */
.internal-topbar__nav-link--active {
  color: var(--color-gold-dark);
  font-weight: 600;
}

.internal-topbar__nav-link--active::after {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 2px;
  transform: translateX(-50%);
  width: 16px;
  height: 2px;
  border-radius: var(--radius-full);
  background: var(--color-gold);
}

/* ── 右：动作区 ── */
.internal-topbar__actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.internal-topbar__divider {
  width: 1px;
  height: 18px;
  background: var(--color-border);
}

.internal-topbar__profile {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  color: var(--color-text-secondary);
  text-decoration: none;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.internal-topbar__profile:hover,
.internal-topbar__profile.router-link-active {
  color: var(--color-gold-dark);
  border-color: var(--color-gold);
  background: color-mix(in srgb, var(--color-gold) 8%, transparent);
}

.internal-topbar__profile svg {
  flex-shrink: 0;
}

/* ── 主内容 ── */
.internal-main {
  flex: 1;
  padding: var(--space-6);
  padding-bottom: calc(var(--space-6) + var(--bottom-nav-height));
}

@media (max-width: 1024px) {
  /* 中部导航让位：底部导航已覆盖同等入口 */
  .internal-topbar__nav {
    display: none;
  }
}

@media (max-width: 768px) {
  .internal-topbar {
    padding: 0 var(--space-4);
  }

  .internal-topbar__divider,
  .internal-topbar__profile {
    display: none;
  }

  .internal-topbar__subtitle {
    display: none;
  }

  .internal-main {
    padding: var(--space-4);
    padding-bottom: calc(var(--space-4) + var(--bottom-nav-height));
  }
}
</style>
