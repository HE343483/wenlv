<script setup lang="ts">
/**
 * WeatherPanel.vue — 天气下拉面板 (非模态浮动层，无卡片/无跳转)
 * 锚定于 WeatherTrigger 正下方，Teleport 至 body
 * 内容：温度区间 + 区域筛选(市→区县→镇/街道 三级级联,覆盖四川全省)
 * 面板内操作直接写入 weather store，全局（按钮/表行）实时同步
 */
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useWeatherStore } from '@/stores/weather'
import { useLanguageStore } from '@/stores/language'
import { fetchRegionChildren, type RegionNode } from '@/api/region'
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

/** 区域级联数据源 — 市 / 区县 / 镇街道 三级,逐级懒加载 */
const cityList = ref<RegionNode[]>([])
const districtList = ref<RegionNode[]>([])
const streetList = ref<RegionNode[]>([])
const cascadeLoading = ref(0)
const cascadeError = ref<string | null>(null)
const pickedCity = ref<RegionNode | null>(null)
const pickedDistrict = ref<RegionNode | null>(null)

/** 区域下拉（自定义级联单选）— 选中区县即拉取天气,街道仅细化展示 */
const districtOpen = ref(false)
const districtTriggerRef = ref<HTMLButtonElement | null>(null)
const districtListRef = ref<HTMLDivElement | null>(null)

/** 触发按钮显示的当前区域名(街道优先) */
const currentDistrictName = computed(() => weatherStore.displayName)

/** 首次展开级联时加载市级列表(sessionStorage 缓存命中则零请求) */
async function ensureCityList() {
  if (cityList.value.length > 0 || cascadeLoading.value > 0) return
  cascadeLoading.value++
  cascadeError.value = null
  try {
    cityList.value = await fetchRegionChildren('四川省')
  } catch {
    cascadeError.value = '加载行政区划失败'
  } finally {
    cascadeLoading.value--
  }
}

/** 选中市 → 加载区县列 */
async function onPickCity(node: RegionNode) {
  pickedCity.value = node
  pickedDistrict.value = null
  districtList.value = []
  streetList.value = []
  if (node.level === 'district') {
    // 省直辖县级行政单位兜底:无下级时直接作为区县选择
    onPickDistrict(node)
    return
  }
  cascadeLoading.value++
  try {
    districtList.value = await fetchRegionChildren(node.adcode)
  } catch {
    cascadeError.value = '加载区县失败'
  } finally {
    cascadeLoading.value--
  }
}

/** 选中区县 → 加载街道列 + 按区县 adcode 查天气 */
async function onPickDistrict(node: RegionNode) {
  pickedDistrict.value = node
  streetList.value = []
  weatherStore.selectRegion({
    cityName: pickedCity.value?.name ?? weatherStore.location.cityName,
    districtName: node.name,
    districtAdcode: node.adcode,
  })
  cascadeLoading.value++
  try {
    streetList.value = await fetchRegionChildren(node.adcode)
  } catch {
    streetList.value = [] // 街道列表加载失败不影响天气
  } finally {
    cascadeLoading.value--
  }
}

/** 选中街道 — 展示细化,天气仍按区县 adcode(命中缓存,零请求);选完即收起下拉 */
function onPickStreet(node: RegionNode) {
  if (!pickedDistrict.value) return
  weatherStore.selectRegion({
    cityName: pickedCity.value?.name ?? weatherStore.location.cityName,
    districtName: pickedDistrict.value.name,
    districtAdcode: pickedDistrict.value.adcode,
    streetName: node.name,
  })
  districtOpen.value = false
}

function toggleDistrict() {
  districtOpen.value = !districtOpen.value
  if (districtOpen.value) void ensureCityList()
}

/** 点击下拉外部时收起 */
function onDistrictOutsideClick(e: MouseEvent) {
  if (!districtOpen.value) return
  const t = e.target as HTMLElement
  if (districtTriggerRef.value?.contains(t) || districtListRef.value?.contains(t)) return
  districtOpen.value = false
}

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
  weatherStore.fetchWeather(true)
}

function close() {
  weatherStore.closeDropdown()
}

/** Esc 关闭（优先收起地区下拉，再关闭面板） */
function onKeydown(e: KeyboardEvent) {
  if (e.key !== 'Escape') return
  if (districtOpen.value) {
    districtOpen.value = false
  } else {
    close()
  }
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
  document.addEventListener('mousedown', onDistrictOutsideClick)
})

onBeforeUnmount(() => {
  props.anchor?.getElement?.()?.removeEventListener('transitionend', onAnchorTransition)
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', updatePos)
  document.removeEventListener('mousedown', onDistrictOutsideClick)
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
          <div class="weather-panel__location">
            <button
              ref="districtTriggerRef"
              class="weather-panel__location-trigger"
              @click="toggleDistrict"
              :aria-expanded="districtOpen"
              aria-haspopup="listbox"
              :aria-label="langStore.t('weather.region')"
            >
              <svg class="weather-panel__location-pin" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 2C8 2 5 5 5 9c0 5 7 13 7 13s7-8 7-13c0-4-3-7-7-7z"/>
                <circle cx="12" cy="9" r="3"/>
              </svg>
              <span class="weather-panel__location-label">{{ currentDistrictName }}</span>
              <svg
                class="weather-panel__location-chevron"
                :class="{ 'weather-panel__location-chevron--flip': districtOpen }"
                width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"
              >
                <path d="M6 9l6 6 6-6"/>
              </svg>
            </button>

            <Transition name="district-drop">
              <div v-if="districtOpen" ref="districtListRef" class="weather-panel__location-dropdown" role="listbox">
                <div v-if="cascadeError" class="weather-panel__cascade-error">{{ cascadeError }}</div>
                <div v-else class="weather-panel__cascade">
                  <div class="weather-panel__cascade-col">
                    <div class="weather-panel__cascade-title">市</div>
                    <div class="weather-panel__cascade-scroll">
                      <div
                        v-for="c in cityList"
                        :key="c.adcode"
                        class="weather-panel__location-option"
                        :class="{ 'weather-panel__location-option--active': pickedCity?.adcode === c.adcode }"
                        role="option"
                        :aria-selected="pickedCity?.adcode === c.adcode"
                        @click="onPickCity(c)"
                      >
                        <span class="weather-panel__location-option-label">{{ c.name }}</span>
                      </div>
                      <div v-if="cascadeLoading > 0 && cityList.length === 0" class="weather-panel__cascade-hint">加载中…</div>
                    </div>
                  </div>
                  <div class="weather-panel__cascade-col">
                    <div class="weather-panel__cascade-title">区/县</div>
                    <div class="weather-panel__cascade-scroll">
                      <div
                        v-for="d in districtList"
                        :key="d.adcode"
                        class="weather-panel__location-option"
                        :class="{ 'weather-panel__location-option--active': pickedDistrict?.adcode === d.adcode }"
                        role="option"
                        :aria-selected="pickedDistrict?.adcode === d.adcode"
                        @click="onPickDistrict(d)"
                      >
                        <span class="weather-panel__location-option-label">{{ d.name }}</span>
                      </div>
                      <div v-if="!pickedCity" class="weather-panel__cascade-hint">请先选择市</div>
                      <div v-else-if="cascadeLoading > 0 && districtList.length === 0" class="weather-panel__cascade-hint">加载中…</div>
                    </div>
                  </div>
                  <div class="weather-panel__cascade-col">
                    <div class="weather-panel__cascade-title">镇/街道</div>
                    <div class="weather-panel__cascade-scroll">
                      <div
                        v-for="s in streetList"
                        :key="s.adcode + s.name"
                        class="weather-panel__location-option"
                        :class="{ 'weather-panel__location-option--active': weatherStore.location.streetName === s.name }"
                        role="option"
                        :aria-selected="weatherStore.location.streetName === s.name"
                        @click="onPickStreet(s)"
                      >
                        <span class="weather-panel__location-option-label">{{ s.name }}</span>
                      </div>
                      <div v-if="!pickedDistrict" class="weather-panel__cascade-hint">--</div>
                    </div>
                  </div>
                </div>
              </div>
            </Transition>
          </div>
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
              <span class="weather-panel__value">{{ data.humidity != null ? `${data.humidity}%` : '--' }}</span>
            </div>
            <div class="weather-panel__detail">
              <span class="weather-panel__label">{{ langStore.t('weather.wind') }}</span>
              <span class="weather-panel__value">{{
                data.wind_power
                  ? `${data.wind_power}级`
                  : data.wind_speed != null
                    ? `${data.wind_speed.toFixed(2)} m/s`
                    : '--'
              }}</span>
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

.weather-panel__location {
  position: relative;
  min-width: 0;
}

/* 触发按钮 — 柔和朱砂笺标签（平和） */
.weather-panel__location-trigger {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
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

.weather-panel__location-trigger:hover,
.weather-panel__location-trigger:focus-visible {
  border-color: color-mix(in srgb, var(--color-cinnabar) 55%, transparent);
  background: color-mix(in srgb, var(--color-cinnabar) 16%, var(--color-surface));
  color: var(--color-cinnabar);
  outline: none;
}

.weather-panel__location-pin {
  flex-shrink: 0;
  opacity: 0.8;
}

.weather-panel__location-label {
  flex: 1;
  text-align: left;
}

.weather-panel__location-chevron {
  color: var(--color-text-muted);
  flex-shrink: 0;
  pointer-events: none;
  transition: transform 0.3s ease;
}

.weather-panel__location-chevron--flip {
  transform: rotate(180deg);
}

/* 下拉列表 */
.weather-panel__location-dropdown {
  position: absolute;
  top: calc(100% + var(--space-2));
  left: 0;
  z-index: 20;
  min-width: 260px;
  padding: var(--space-1);
  background: var(--color-surface-elevated);
  border: 1px solid color-mix(in srgb, var(--color-gold) 25%, transparent);
  border-radius: var(--radius-md);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.45);
}

/* 三列级联 */
.weather-panel__cascade {
  display: flex;
  gap: 2px;
}

.weather-panel__cascade-col {
  flex: 1;
  min-width: 0;
}

.weather-panel__cascade-title {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wider);
  padding: var(--space-1) var(--space-2);
  border-bottom: 1px solid color-mix(in srgb, var(--color-gold) 15%, transparent);
  margin-bottom: var(--space-1);
}

.weather-panel__cascade-scroll {
  max-height: 220px;
  overflow-y: auto;
}

.weather-panel__cascade-hint {
  padding: var(--space-2);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  text-align: center;
}

.weather-panel__cascade-error {
  padding: var(--space-3);
  font-size: var(--text-xs);
  color: var(--color-cinnabar);
  text-align: center;
}

.weather-panel__location-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background 0.2s ease;
  user-select: none;
}

.weather-panel__location-option:hover {
  background: var(--color-surface-hover);
}

.weather-panel__location-option--active {
  color: var(--color-gold);
}

.weather-panel__location-option-label {
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-primary);
}

.weather-panel__location-option--active .weather-panel__location-option-label {
  color: var(--color-gold);
}

.weather-panel__location-check {
  color: var(--color-gold);
  flex-shrink: 0;
}

/* 下拉过渡 */
.district-drop-enter-active {
  transition: all 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.district-drop-leave-active {
  transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1);
}

.district-drop-enter-from {
  opacity: 0;
  transform: translateY(-6px) scale(0.97);
}

.district-drop-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
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
