<script setup lang="ts">
/**
 * WeatherRowCell.vue — 表页面内嵌天气行（紧凑，无独立卡片/弹层）
 * 展示：色调色板 + 区名 + 温度区间 + 亮/暗色调切换按钮
 * 只订阅 weather store / theme store，与下拉面板实时同步（切换区域后温度同步刷新）
 */
import { storeToRefs } from 'pinia'
import { useWeatherStore } from '@/stores/weather'
import { useLanguageStore } from '@/stores/language'
import { useThemeStore } from '@/stores/theme'

const weatherStore = useWeatherStore()
const langStore = useLanguageStore()
const themeStore = useThemeStore()
const { data, location, state } = storeToRefs(weatherStore)
const { theme } = storeToRefs(themeStore)

function toggleTheme() {
  themeStore.toggle()
}
</script>

<template>
  <div class="weather-row-cell">
    <span class="weather-row-cell__swatch" style="background: var(--temp-tone)" />

    <div class="weather-row-cell__info">
      <span class="weather-row-cell__district">{{ location.districtName }}</span>
      <span class="weather-row-cell__temp">
        <template v-if="data">{{ data.min_temp }}° ~ {{ data.max_temp }}°</template>
        <template v-else>--</template>
        <span v-if="state === 'loading'" class="weather-row-cell__loading">
          {{ langStore.t('weather.refreshing') }}
        </span>
      </span>
    </div>

    <!-- 亮/暗色调切换 -->
    <button
      class="weather-row-cell__theme-btn"
      @click="toggleTheme"
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
  </div>
</template>

<style scoped>
.weather-row-cell {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: border-color 0.3s ease;
}

.weather-row-cell:hover {
  border-color: var(--temp-tone);
}

/* 色调色板 */
.weather-row-cell__swatch {
  width: 14px;
  height: 14px;
  border-radius: var(--radius-full);
  background: var(--temp-tone);
  box-shadow: 0 0 10px var(--temp-tone-soft);
  flex-shrink: 0;
}

.weather-row-cell__info {
  display: flex;
  flex-direction: column;
  line-height: 1.3;
}

.weather-row-cell__district {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
}

.weather-row-cell__temp {
  font-family: var(--font-en-body);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--temp-tone-text);
  white-space: nowrap;
}

.weather-row-cell__loading {
  margin-left: var(--space-2);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
}

/* 亮/暗色调切换按钮 */
.weather-row-cell__theme-btn {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  background: transparent;
  margin-left: auto;
  flex-shrink: 0;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.weather-row-cell__theme-btn:hover {
  border-color: var(--color-gold);
  color: var(--color-gold);
  box-shadow: 0 0 12px var(--color-gold-glow);
}
</style>
