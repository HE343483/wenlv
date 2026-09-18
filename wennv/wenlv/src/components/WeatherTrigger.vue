<script setup lang="ts">
/**
 * WeatherTrigger.vue — 天气图标 + 右侧滑出简讯
 * 默认仅图标；点击后在右侧滑出「区名 最低°–最高°」
 * 暴露 getAnchorRect / getElement 供外部点击判定
 */
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useWeatherStore } from '@/stores/weather'
import AppIcon from '@/components/AppIcon.vue'

const weatherStore = useWeatherStore()
const { data, open, state } = storeToRefs(weatherStore)

const btnRef = ref<HTMLButtonElement | null>(null)

const iconName = computed(() => data.value?.weather_icon || 'sunny')

const summaryText = computed(() => {
  if (!data.value) return state.value === 'loading' ? '获取中…' : '天气'
  const place = weatherStore.displayName || '成都'
  return `${place} ${data.value.min_temp}°C–${data.value.max_temp}°C`
})

function onClick() {
  weatherStore.toggleDropdown()
  if (state.value === 'idle') {
    weatherStore.fetchWeather()
  }
}

function getAnchorRect(): DOMRect {
  return btnRef.value?.getBoundingClientRect() ?? new DOMRect()
}

function getElement(): HTMLElement | null {
  return btnRef.value
}

defineExpose({ getAnchorRect, getElement })
</script>

<template>
  <button
    ref="btnRef"
    type="button"
    class="weather-trigger"
    :class="{ 'weather-trigger--open': open }"
    @click="onClick"
    :aria-expanded="open"
    :aria-label="open ? summaryText : '展开天气简讯'"
    :title="summaryText"
  >
    <AppIcon class="weather-trigger__icon" :name="iconName" :size="22" />

    <span class="weather-trigger__slide" :aria-hidden="!open">
      <span class="weather-trigger__slide-inner">
        <span class="weather-trigger__summary">{{ summaryText }}</span>
      </span>
    </span>
  </button>
</template>

<style scoped>
.weather-trigger {
  display: inline-flex;
  align-items: center;
  gap: 0;
  height: 36px;
  flex-shrink: 0;
  padding: 0;
  border: none;
  border-radius: var(--radius-full);
  background: transparent;
  color: var(--temp-tone);
  cursor: pointer;
  transition: color 0.25s ease;
}

.weather-trigger__icon {
  display: block;
  flex-shrink: 0;
  color: inherit;
}

/* 右侧滑出槽：用 grid 0fr→1fr 做宽度动画，不挤坏布局 */
.weather-trigger__slide {
  display: grid;
  grid-template-columns: 0fr;
  transition: grid-template-columns 0.35s cubic-bezier(0.22, 1, 0.36, 1);
  min-width: 0;
}

.weather-trigger--open .weather-trigger__slide {
  grid-template-columns: 1fr;
}

.weather-trigger__slide-inner {
  overflow: hidden;
  min-width: 0;
}

.weather-trigger__summary {
  display: block;
  padding-left: 8px;
  white-space: nowrap;
  font-family: var(--font-body);
  font-size: var(--text-sm);
  font-weight: 500;
  letter-spacing: 0.02em;
  color: var(--color-text-secondary);
  font-variant-numeric: tabular-nums;
  opacity: 0;
  transform: translateX(-6px);
  transition:
    opacity 0.28s cubic-bezier(0.22, 1, 0.36, 1) 0.04s,
    transform 0.28s cubic-bezier(0.22, 1, 0.36, 1) 0.04s,
    color 0.25s ease;
}

.weather-trigger--open .weather-trigger__summary {
  opacity: 1;
  transform: translateX(0);
}

.weather-trigger:hover {
  color: var(--temp-tone-text);
}

.weather-trigger:hover .weather-trigger__summary,
.weather-trigger--open .weather-trigger__summary {
  color: var(--color-text-primary);
}

.weather-trigger:focus-visible {
  outline: 2px solid var(--temp-tone);
  outline-offset: 2px;
}

@media (prefers-reduced-motion: reduce) {
  .weather-trigger__slide,
  .weather-trigger__summary {
    transition: none;
  }
}
</style>
