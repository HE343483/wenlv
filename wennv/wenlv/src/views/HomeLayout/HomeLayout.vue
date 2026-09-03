<script setup lang="ts">
/**
 * HomeLayout.vue — 登录后内部页面布局
 * 包含: 简易顶栏 + 子路由内容 + 底部导航
 */
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { useLanguageStore } from '@/stores/language'
import { useThemeStore } from '@/stores/theme'
import { useWeatherStore } from '@/stores/weather'
import BottomNav from '@/components/BottomNav.vue'
import WeatherTrigger from '@/components/WeatherTrigger.vue'
import WeatherPanel from '@/components/WeatherPanel.vue'

const router = useRouter()
const langStore = useLanguageStore()
const themeStore = useThemeStore()
const weatherStore = useWeatherStore()
const { open } = storeToRefs(weatherStore)
const { theme } = storeToRefs(themeStore)

function goBack() {
  router.push('/')
}

/* 天气下拉：触发点 ref + 面板 ref（面板 Teleport 到 body，用暴露的 getElement 判点击范围） */
const triggerRef = ref<InstanceType<typeof WeatherTrigger> | null>(null)
const panelRef = ref<InstanceType<typeof WeatherPanel> | null>(null)
</script>

<template>
  <div class="internal-layout">
    <!-- 顶栏 -->
    <header class="internal-topbar">
      <button class="internal-topbar__back" @click="goBack">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M15 18l-6-6 6-6"/>
        </svg>
      </button>

      <div class="internal-topbar__brand">
        <span class="internal-topbar__title">蜀韵·成都</span>
      </div>

      <div class="internal-topbar__actions">
        <!-- 天气：锚点按钮（含温度区间） + 下拉面板 -->
        <WeatherTrigger ref="triggerRef" />
        <WeatherPanel v-if="open" ref="panelRef" :anchor="triggerRef" />

        <!-- 亮/暗色调切换 -->
        <button
          class="internal-topbar__icon-btn"
          @click="themeStore.toggle()"
          :title="theme === 'dark' ? langStore.t('theme.switchToLight') : langStore.t('theme.switchToDark')"
          :aria-label="theme === 'dark' ? langStore.t('theme.switchToLight') : langStore.t('theme.switchToDark')"
        >
          <!-- 暗色 → 点按切换亮色 -->
          <svg v-if="theme === 'dark'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="5"/>
            <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2"/>
          </svg>
          <!-- 亮色 → 点按切换暗色 -->
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
        </button>
        <!-- 语言切换 -->
        <button class="internal-topbar__lang-btn" @click="langStore.toggle()">
          {{ langStore.t('nav.langSwitch') }}
        </button>
      </div>
    </header>

    <!-- 主内容区 (子路由) -->
    <main class="internal-main">
      <RouterView />
    </main>

    <!-- 底部导航 -->
    <BottomNav />
  </div>
</template>

<style scoped>
.internal-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding-bottom: var(--bottom-nav-height);
}

/* 顶栏 */
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

.internal-topbar__back {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
}

.internal-topbar__back:hover {
  background: var(--color-surface-hover);
  color: var(--color-gold);
}

.internal-topbar__brand {
  display: flex;
  align-items: center;
}

.internal-topbar__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
}

.internal-topbar__actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.internal-topbar__icon-btn {
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

.internal-topbar__icon-btn:hover {
  border-color: var(--color-gold);
  color: var(--color-gold);
}

.internal-topbar__lang-btn {
  font-family: var(--font-display);
  font-size: var(--text-sm);
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.internal-topbar__lang-btn:hover {
  border-color: var(--color-gold);
  color: var(--color-gold);
}

/* 主内容 */
.internal-main {
  flex: 1;
  padding: var(--space-6);
  padding-bottom: calc(var(--space-6) + var(--bottom-nav-height));
}

@media (max-width: 768px) {
  .internal-main {
    padding: var(--space-4);
    padding-bottom: calc(var(--space-4) + var(--bottom-nav-height));
  }
}
</style>
