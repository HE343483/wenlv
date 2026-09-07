<script setup lang="ts">
/**
 * WeatherRowCell.vue — 表页面内嵌天气行（紧凑，无独立卡片/弹层）
 * 展示：色调色板 + 区名 + 温度区间
 * 只订阅 weather store，与下拉面板实时同步（切换区域后温度同步刷新）
 */
import { storeToRefs } from 'pinia'
import { useWeatherStore } from '@/stores/weather'
import { useLanguageStore } from '@/stores/language'

const weatherStore = useWeatherStore()
const langStore = useLanguageStore()
const { data, location, state } = storeToRefs(weatherStore)
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
</style>
