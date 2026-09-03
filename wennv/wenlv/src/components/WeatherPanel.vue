<script setup lang="ts">
/**
 * WeatherPanel.vue — 天气下拉面板 (非模态浮动层，无卡片/无跳转)
 * 锚定于 WeatherTrigger 正下方，Teleport 至 body
 * 内容：温度区间 + 区域筛选（仅“区”级）
 * 面板内操作直接写入 weather store，全局（按钮/表行）实时同步
 */
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useWeatherStore } from '@/stores/weather'
import { useLanguageStore } from '@/stores/language'
import { getDistrictOptions } from '@/data/chengdu'
import AppIcon from '@/components/AppIcon.vue'

const props = defineProps<{
  /** WeatherTrigger 暴露的定位/取值接口 */
  anchor: { getAnchorRect: () => DOMRect; getElement?: () => HTMLElement | null } | null
}>()

const weatherStore = useWeatherStore()
const langStore = useLanguageStore()

const { data, state, error } = storeToRefs(weatherStore)

const panelRef = ref<HTMLDivElement | null>(null)
const pos = ref({ top: 0, left: 0 })
/** 挂载后置 true，触发进入过渡（配合外部 v-if 卸载动画） */
const visible = ref(false)
const PANEL_W = 280
const GAP = 8

/** 区域筛选数据源 — 仅“区”级行政区划 */
const districtOptions = getDistrictOptions()

/** 区域下拉：切换即按 districtId 重新拉取天气 */
const districtIdModel = computed({
  get: () => weatherStore.location.districtId,
  set: (id: string) => weatherStore.selectDistrict(id),
})

/** 定位：触发点下方 8px，空间不足时翻转 */
function updatePos() {
  const rect = props.anchor?.getAnchorRect()
  if (!rect) return
  const panelH = panelRef.value?.offsetHeight ?? 220
  const left = Math.min(Math.max(rect.left, 8), window.innerWidth - PANEL_W - 8)
  let top = rect.bottom + GAP
  if (top + panelH > window.innerHeight - 8) {
    top = Math.max(8, rect.top - GAP - panelH)
  }
  pos.value = { top, left }
}

function refresh() {
  weatherStore.fetchWeather()
}

function close() {
  weatherStore.closeDropdown()
}

/** Esc 关闭 */
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') close()
}

/** 供外部（NavBar）判定面板点击范围 */
function getElement(): HTMLElement | null {
  return panelRef.value
}

defineExpose({ getElement })

/** 按钮展开过渡结束后重新定位，保持面板与按钮左缘对齐 */
function onAnchorTransition(e: TransitionEvent) {
  if (e.propertyName === 'grid-template-columns') {
    updatePos()
  }
}

onMounted(async () => {
  await nextTick()
  updatePos()
  visible.value = true
  await nextTick()
  updatePos() // 面板实际高度就位后重定位，避免底部溢出
  props.anchor?.getElement?.()?.addEventListener('transitionend', onAnchorTransition)
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('resize', updatePos)
})

onBeforeUnmount(() => {
  props.anchor?.getElement?.()?.removeEventListener('transitionend', onAnchorTransition)
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', updatePos)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="weather-panel">
      <div
        v-if="visible"
        ref="panelRef"
        class="weather-panel"
        :style="{ top: pos.top + 'px', left: pos.left + 'px', width: PANEL_W + 'px' }"
        role="dialog"
        aria-label="天气信息"
      >
        <!-- 头部：区域就地选择 + 刷新 -->
        <div class="weather-panel__header">
          <span class="weather-panel__location-wrap">
            <select
              v-model="districtIdModel"
              class="weather-panel__location-select"
              :aria-label="langStore.t('weather.region')"
            >
              <option v-for="d in districtOptions" :key="d.id" :value="d.id">{{ d.name }}</option>
            </select>
            <svg class="weather-panel__location-chevron" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M6 9l6 6 6-6"/>
            </svg>
          </span>
          <button class="weather-panel__refresh" @click="refresh" :title="langStore.t('weather.retry')">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M1 4v6h6M23 20v-6h-6"/>
              <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"/>
            </svg>
          </button>
        </div>

        <!-- 加载中 -->
        <div v-if="state === 'loading'" class="weather-panel__loading">
          <span class="weather-panel__spinner" />
          <span>{{ langStore.t('weather.refreshing') }}</span>
        </div>

        <!-- 错误（内联，带重试） -->
        <div v-else-if="state === 'error'" class="weather-panel__error">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="10"/><path d="M12 8v4M12 16h.01"/>
          </svg>
          <span>{{ error || langStore.t('weather.error') }}</span>
          <button class="weather-panel__retry" @click="refresh">{{ langStore.t('weather.retry') }}</button>
        </div>

        <!-- 成功：温度区间 -->
        <template v-else-if="data">
          <div class="weather-panel__current">
            <span class="weather-panel__icon">
              <AppIcon :name="data.weather_icon" :size="28" />
            </span>
            <span class="weather-panel__desc">{{ data.weather_desc }}</span>
          </div>

          <div class="weather-panel__range">
            {{ data.min_temp }}° ~ {{ data.max_temp }}°
          </div>

          <div class="weather-panel__details">
            <div class="weather-panel__detail">
              <span class="weather-panel__label">{{ langStore.t('weather.humidity') }}</span>
              <span class="weather-panel__value">{{ data.humidity }}%</span>
            </div>
            <div class="weather-panel__detail">
              <span class="weather-panel__label">{{ langStore.t('weather.wind') }}</span>
              <span class="weather-panel__value">{{ data.wind_speed.toFixed(2) }} m/s</span>
            </div>
          </div>
        </template>

      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* 蜀锦笺 — 实心材质 + 金线描边 + 极淡织纹底 */
.weather-panel {
  position: fixed;
  z-index: 2000;
  background: var(--color-surface);
  border: 1px solid color-mix(in srgb, var(--color-gold) 32%, transparent);
  border-radius: var(--radius-md);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.45);
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

/* 蜀锦织纹底（极淡，压住内容之下） */
.weather-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  pointer-events: none;
  background-image: repeating-conic-gradient(
    transparent 0deg 44deg,
    color-mix(in srgb, var(--color-gold) 5%, transparent) 45deg 46deg,
    transparent 46deg 89deg,
    color-mix(in srgb, var(--color-gold) 5%, transparent) 90deg 91deg,
    transparent 91deg
  );
  background-size: 24px 24px;
  opacity: 0.35;
}

/* 头部 */
.weather-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.weather-panel__location-wrap {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}

/* 区域选择 — 柔和朱砂笺标签（平和） */
.weather-panel__location-select {
  appearance: none;
  -webkit-appearance: none;
  border: 1px solid color-mix(in srgb, var(--color-cinnabar) 28%, transparent);
  background: color-mix(in srgb, var(--color-cinnabar) 10%, var(--color-surface));
  color: color-mix(in srgb, var(--color-cinnabar) 52%, var(--color-text-primary));
  font-family: var(--font-display);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  padding: 4px 10px;
  border-radius: var(--radius-md);
  white-space: nowrap;
  cursor: pointer;
  transition: border-color 0.3s ease, background 0.3s ease, color 0.3s ease;
}

.weather-panel__location-select:hover,
.weather-panel__location-select:focus {
  border-color: color-mix(in srgb, var(--color-cinnabar) 55%, transparent);
  background: color-mix(in srgb, var(--color-cinnabar) 16%, var(--color-surface));
  color: var(--color-cinnabar);
  outline: none;
}

.weather-panel__location-chevron {
  color: var(--color-text-muted);
  flex-shrink: 0;
  pointer-events: none;
}

.weather-panel__refresh {
  color: var(--color-text-muted);
  transition: color 0.3s ease, transform 0.3s ease;
  flex-shrink: 0;
}

.weather-panel__refresh:hover {
  color: var(--temp-tone);
  transform: rotate(180deg);
}

/* 加载 */
.weather-panel__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-6) 0;
  color: var(--color-text-muted);
  font-size: var(--text-sm);
}

.weather-panel__spinner {
  width: 18px;
  height: 18px;
  border: 2px solid var(--color-border);
  border-top-color: var(--temp-tone);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* 错误 */
.weather-panel__error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) 0;
  color: var(--color-cinnabar);
  font-size: var(--text-sm);
  text-align: center;
}

.weather-panel__retry {
  padding: var(--space-2) var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wide);
  transition: all 0.3s ease;
}

.weather-panel__retry:hover {
  border-color: var(--temp-tone);
  color: var(--temp-tone);
}

/* 成功内容 */
.weather-panel__current {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.weather-panel__icon {
  display: flex;
  align-items: center;
  color: var(--temp-tone);
}

.weather-panel__desc {
  font-family: var(--font-display);
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
}

/* 温度区间 — 衬线数字 + 色调强调 + 金线 */
.weather-panel__range {
  font-family: var(--font-en-display);
  font-size: var(--text-3xl);
  font-weight: 400;
  color: var(--temp-tone-text);
  letter-spacing: 0.04em;
  line-height: 1.2;
  padding-bottom: var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--color-gold) 25%, transparent);
}

/* 附加信息 */
.weather-panel__details {
  display: flex;
  gap: var(--space-6);
}

.weather-panel__detail {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.weather-panel__label {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: var(--tracking-wider);
}

.weather-panel__value {
  font-family: var(--font-en-body);
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  font-weight: 500;
}

/* 过渡动画 — 进入缓落、退出快收 */
.weather-panel-enter-active {
  transition: opacity 0.32s cubic-bezier(0.22, 1, 0.36, 1),
              transform 0.32s cubic-bezier(0.22, 1, 0.36, 1);
}

.weather-panel-leave-active {
  transition: opacity 0.2s ease-in, transform 0.2s ease-in;
}

.weather-panel-enter-from,
.weather-panel-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.98);
}

@media (prefers-reduced-motion: reduce) {
  .weather-panel-enter-active,
  .weather-panel-leave-active {
    transition: none;
  }
  .weather-panel__refresh,
  .weather-panel__spinner {
    transition: none;
    animation: none;
  }
}
</style>
