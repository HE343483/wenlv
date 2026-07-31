<script setup lang="ts">
/**
 * WeatherTrigger.vue — 天气锚点按钮 (NavBar 固定区)
 * 收缩版：图标 + 温度区间（{{min_temp}}°~{{max_temp}}°）
 * 点击展开：追加区名并弹出下拉面板；再次点击收缩回去
 * 暴露 getAnchorRect / getElement 供面板定位与外部点击判定
 */
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useWeatherStore } from '@/stores/weather'

const weatherStore = useWeatherStore()
const { data, open, state, location } = storeToRefs(weatherStore)

const btnRef = ref<HTMLButtonElement | null>(null)

function onClick() {
  weatherStore.toggleDropdown()
  // 首次展开且无数据时自动拉取
  if (state.value === 'idle') {
    weatherStore.fetchWeather()
  }
}

/** 供 WeatherPanel 计算锚点坐标 */
function getAnchorRect(): DOMRect {
  return btnRef.value?.getBoundingClientRect() ?? new DOMRect()
}

/** 供外部点击判定面板点击范围 */
function getElement(): HTMLElement | null {
  return btnRef.value
}

defineExpose({ getAnchorRect, getElement })
</script>

<template>
  <button
    ref="btnRef"
    class="weather-trigger"
    :class="{ 'weather-trigger--open': open }"
    @click="onClick"
    :aria-expanded="open"
    aria-haspopup="dialog"
    :title="open ? '收起天气' : '展开查看天气'"
  >
    <svg class="weather-trigger__icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
      <circle cx="12" cy="12" r="5"/>
      <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2"/>
    </svg>
    <!-- 展开态：追加区名（带过渡动画） -->
    <Transition name="weather-trigger-text">
      <span v-if="open && data" class="weather-trigger__text">
        <span class="weather-trigger__city">{{ location.districtName }}</span>
      </span>
    </Transition>
    <!-- 温度区间：收缩版也显示（图片 + 温度） -->
    <span v-if="data" class="weather-trigger__range">{{ data.min_temp }}°~{{ data.max_temp }}°</span>
    <svg
      class="weather-trigger__chevron"
      :class="{ 'weather-trigger__chevron--flip': open }"
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="1.5"
    >
      <path d="M6 9l6 6 6-6"/>
    </svg>
  </button>
</template>

<style scoped>
.weather-trigger {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  height: 36px;
  padding: 0 var(--space-3);
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  background: transparent;
  transition: all 0.3s ease;
  white-space: nowrap;
}

/* 色调作用：边框 + 图标 + 温度文字 */
.weather-trigger__icon {
  color: var(--temp-tone);
}

/* 展开态文本（区名）— 平滑宽度展开（grid 0fr → 1fr） */
.weather-trigger__text {
  display: grid;
  grid-template-columns: 1fr;
  align-items: center;
  transition: grid-template-columns 0.35s cubic-bezier(0.22, 1, 0.36, 1),
              opacity 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}

.weather-trigger__text > * {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
}

.weather-trigger-text-enter-from,
.weather-trigger-text-leave-to {
  grid-template-columns: 0fr;
  opacity: 0;
}

.weather-trigger__range {
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  font-weight: 500;
  letter-spacing: 0.02em;
  color: var(--temp-tone-text);
}

.weather-trigger__city {
  font-family: var(--font-display);
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
}

/* 指示箭头：展开时翻转 */
.weather-trigger__chevron {
  color: var(--color-text-muted);
  transition: transform 0.3s ease;
  flex-shrink: 0;
}

.weather-trigger__chevron--flip {
  transform: rotate(180deg);
  color: var(--temp-tone);
}

.weather-trigger:hover {
  border-color: var(--temp-tone);
  box-shadow: 0 0 14px var(--temp-tone-soft);
}

.weather-trigger--open {
  border-color: var(--temp-tone);
  background: var(--temp-tone-soft);
}

@media (prefers-reduced-motion: reduce) {
  .weather-trigger,
  .weather-trigger__chevron,
  .weather-trigger__text,
  .weather-trigger-text-enter-active,
  .weather-trigger-text-leave-active {
    transition: none;
  }
}
</style>
