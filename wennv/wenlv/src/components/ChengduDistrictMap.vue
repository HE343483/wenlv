<script setup lang="ts">
/**
 * ChengduDistrictMap.vue — 成都区县交互地图
 * 加载百度地图，获取各区县行政边界，绘制可点击多边形
 * 点击区县 → emit('select', districtId)
 * 地图颜色随主题自动切换（暗色/亮色）
 */
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useLanguageStore } from '@/stores/language'
import { useThemeStore } from '@/stores/theme'
import { loadBaiduMap } from '@/utils/baiduMap'
import { darkMapStyle, lightMapStyle } from '@/utils/mapStyle'
import { districts, scenicSpots } from '@/data/chengdu'
import AppIcon from '@/components/AppIcon.vue'

const props = defineProps<{
  selected?: string
}>()

const emit = defineEmits<{
  select: [districtId: string]
}>()

const langStore = useLanguageStore()
const themeStore = useThemeStore()

const mapAK = (import.meta.env.VITE_BAIDU_MAP_AK as string | undefined) ?? ''

/* 地图状态 */
type MapStatus = 'idle' | 'loading' | 'ready' | 'error'
const mapStatus = ref<MapStatus>('idle')

/* 地图容器 */
const mapEl = ref<HTMLDivElement | null>(null)

/* 地图实例 */
let map: any = null
let districtPolygons: Array<{ id: string; polygon: any; label: any }> = []
let currentHover: string | null = null
let currentSelected: string | null = null
let isMapReady = false

/* 应用地图样式（按主题） */
function applyMapStyle() {
  if (!map) return
  const styleJson = themeStore.theme === 'dark' ? darkMapStyle : lightMapStyle
  try {
    map.setMapStyleV2({ styleJson })
  } catch {
    // 兼容旧版 API
  }
}

/* 初始化地图 */
async function initMap() {
  if (map || mapStatus.value !== 'idle') return
  if (!mapAK) {
    mapStatus.value = 'error'
    return
  }
  mapStatus.value = 'loading'
  try {
    await loadBaiduMap(mapAK)
    await new Promise(resolve => setTimeout(resolve, 100))
    const el = mapEl.value
    if (!el) return
    const BMap = (window as any).BMap
    map = new BMap.Map(el)
    map.enableScrollWheelZoom()
    map.addControl(new BMap.NavigationControl({
      type: BMAP_NAVIGATION_CONTROL_ZOOM,
      anchor: BMAP_ANCHOR_TOP_LEFT,
    }))
    // 成都中心点（zoom 10 显示整个成都辖区）
    map.centerAndZoom(new BMap.Point(104.065, 30.66), 10)
    applyMapStyle()
    isMapReady = true
    mapStatus.value = 'ready'

    // 等地图加载完再获取区县边界
    setTimeout(() => loadDistrictBoundaries(), 300)
  } catch {
    mapStatus.value = 'error'
  }
}

/* 获取区县边界并绘制多边形 */
function loadDistrictBoundaries() {
  if (!isMapReady) return
  const BMap = (window as any).BMap
  const boundary = new BMap.Boundary()

  let completed = 0
  const total = districts.length

  districts.forEach((d) => {
    boundary.get(d.nameZh, (result: any) => {
      completed++
      if (!isMapReady || !map) return

      if (!result || !result.boundaries || result.boundaries.length === 0) {
        // 边界获取失败，用该区景点中心位置画一个圆作为 fallback
        drawFallback(d)
        if (completed >= total) afterAllBoundaries()
        return
      }

      // 一个区可能有多个边界块（如都江堰市有飞地）
      result.boundaries.forEach((boundaryStr: string) => {
        const points = boundaryStr.split(';').map((p: string) => {
          const parts = p.split(',')
          const lng = parseFloat(parts[0] ?? '')
          const lat = parseFloat(parts[1] ?? '')
          return new BMap.Point(lng, lat)
        })
        if (points.length < 3) return

        const polygon = new BMap.Polygon(points, getPolygonStyle(d.color, false))
        polygon.addEventListener('click', () => onDistrictClick(d.id))
        polygon.addEventListener('mouseover', () => onDistrictHover(d.id))
        polygon.addEventListener('mouseout', () => onDistrictHover(null))
        map.addOverlay(polygon)

        // 区名标签
        const center = getPolygonCenter(points)
        const label = new BMap.Label(
          langStore.lang === 'zh' ? d.nameZh : d.nameEn,
          { position: center, offset: new BMap.Size(0, 0) }
        )
        label.setStyle({
          color: d.color,
          background: 'none',
          border: 'none',
          fontSize: '13px',
          fontWeight: '700',
          fontFamily: "'Noto Serif SC', serif",
          letterSpacing: '0.08em',
          textShadow: themeStore.theme === 'dark'
            ? '0 0 8px rgba(0,0,0,0.8), 0 0 16px rgba(0,0,0,0.6)'
            : '0 0 8px rgba(255,255,255,0.8), 0 0 16px rgba(255,255,255,0.6)',
          cursor: 'pointer',
          whiteSpace: 'nowrap',
        })
        label.addEventListener('click', () => onDistrictClick(d.id))
        map.addOverlay(label)

        districtPolygons.push({ id: d.id, polygon, label })
      })

      if (completed >= total) afterAllBoundaries()
    })
  })
}

/* 边界获取失败时的 fallback（在景点坐标中心画圆） */
function drawFallback(district: typeof districts[number]) {
  const BMap = (window as any).BMap
  // 取该区第一个景点的坐标作为圆心
  const spot = scenicSpots.find(s => s.districtId === district.id)
  if (!spot) return

  const center = new BMap.Point(spot.coords.lng, spot.coords.lat)
  const circle = new BMap.Circle(center, 6000, {
    fillColor: district.color,
    fillOpacity: 0.12,
    strokeColor: district.color,
    strokeWeight: 2,
    strokeOpacity: 0.5,
  })
  circle.addEventListener('click', () => onDistrictClick(district.id))
  circle.addEventListener('mouseover', () => onDistrictHover(district.id))
  circle.addEventListener('mouseout', () => onDistrictHover(null))
  map.addOverlay(circle)
  districtPolygons.push({ id: district.id, polygon: circle, label: null })
}

/* 所有边界加载完毕后，高亮当前选中的区 */
function afterAllBoundaries() {
  if (props.selected && props.selected !== 'all') {
    highlightDistrict(props.selected)
  }
}

/* 计算多边形中心点（用于标签定位） */
function getPolygonCenter(points: any[]): any {
  const BMap = (window as any).BMap
  let lng = 0, lat = 0, count = 0
  points.forEach(p => { lng += p.lng; lat += p.lat; count++ })
  return new BMap.Point(lng / count, lat / count)
}

/* 获取多边形样式 */
function getPolygonStyle(color: string, active: boolean) {
  return {
    fillColor: color,
    fillOpacity: active ? 0.35 : 0.12,
    strokeColor: active ? '#C9A96E' : color,
    strokeWeight: active ? 3 : 2,
    strokeOpacity: active ? 0.9 : 0.6,
    strokeStyle: active ? 'solid' : 'solid' as const,
  }
}

/* 点击区县 */
function onDistrictClick(id: string) {
  currentSelected = id
  highlightDistrict(id)
  emit('select', id)
}

/* 高亮某个区县 */
function highlightDistrict(id: string) {
  districtPolygons.forEach(({ polygon, id: pid }) => {
    const d = districts.find(dd => dd.id === pid)
    if (!d) return
    const active = pid === id
    polygon.setOptions(getPolygonStyle(d.color, active))
  })
}

/* hover 效果 */
function onDistrictHover(id: string | null) {
  if (id === currentSelected) return
  if (currentHover === id) return
  currentHover = id

  districtPolygons.forEach(({ polygon, id: pid }) => {
    const d = districts.find(dd => dd.id === pid)
    if (!d) return
    const isActive = pid === currentSelected
    const isHover = pid === id
    if (isActive) return
    polygon.setOptions(getPolygonStyle(d.color, isHover || isActive))
  })
}

/* 监听主题变化 → 切换地图样式 + 标签文字阴影 */
watch(() => themeStore.theme, () => {
  applyMapStyle()
  districtPolygons.forEach(({ label }) => {
    if (!label) return
    label.setStyle({
      textShadow: themeStore.theme === 'dark'
        ? '0 0 8px rgba(0,0,0,0.8), 0 0 16px rgba(0,0,0,0.6)'
        : '0 0 8px rgba(255,255,255,0.8), 0 0 16px rgba(255,255,255,0.6)',
    })
  })
})

/* 监听语言切换 → 更新标签文字 */
watch(() => langStore.lang, () => {
  districtPolygons.forEach(({ id, label }) => {
    if (!label) return
    const d = districts.find(dd => dd.id === id)
    if (!d) return
    label.setContent(langStore.lang === 'zh' ? d.nameZh : d.nameEn)
  })
})

/* 监听 selected prop 变化 → 高亮选中区 */
watch(() => props.selected, (val) => {
  if (val && val !== 'all') {
    highlightDistrict(val)
  } else {
    // 重置所有高亮
    districtPolygons.forEach(({ polygon, id: pid }) => {
      const d = districts.find(dd => dd.id === pid)
      if (!d) return
      polygon.setOptions(getPolygonStyle(d.color, false))
    })
    currentSelected = null
  }
})

onMounted(() => {
  initMap()
})

onBeforeUnmount(() => {
  isMapReady = false
  if (map) {
    try { map.destroy() } catch {}
    map = null
  }
  districtPolygons = []
})
</script>

<template>
  <div class="district-map">
    <!-- 地图容器 -->
    <div ref="mapEl" class="district-map__canvas" :class="{ 'district-map__canvas--ready': mapStatus === 'ready' }" />

    <!-- 加载/错误覆盖层 -->
    <div v-if="mapStatus !== 'ready'" class="district-map__overlay">
      <template v-if="mapStatus === 'loading'">
        <div class="district-map__spinner" aria-hidden="true" />
        <p class="district-map__hint">{{ langStore.t('routes.mapLoading') }}</p>
      </template>
      <template v-else>
        <span class="district-map__pin"><AppIcon name="pin" :size="40" /></span>
        <p class="district-map__hint">
          {{ mapAK ? langStore.t('routes.mapError') : langStore.t('routes.mapNoKey') }}
        </p>
      </template>
    </div>
  </div>
</template>

<style scoped>
.district-map {
  position: relative;
  width: 100%;
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--color-border);
  background: var(--color-bg-alt);
  min-height: 420px;
}

.district-map__canvas {
  width: 100%;
  aspect-ratio: 16 / 7;
  min-height: 420px;
  opacity: 0;
  transition: opacity var(--transition-base);
}

.district-map__canvas--ready {
  opacity: 1;
}

.district-map__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  background:
    radial-gradient(ellipse 55% 70% at 50% 45%, color-mix(in srgb, var(--color-gold) 7%, transparent) 0%, transparent 70%),
    repeating-linear-gradient(0deg, transparent 0 49px, color-mix(in srgb, var(--color-border) 50%, transparent) 49px 50px),
    repeating-linear-gradient(90deg, transparent 0 49px, color-mix(in srgb, var(--color-border) 50%, transparent) 49px 50px),
    var(--color-bg-alt);
  z-index: 1;
}

.district-map__pin {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-gold);
  filter: drop-shadow(0 0 16px var(--color-gold-glow));
  animation: pin-float 2.6s ease-in-out infinite;
}

@keyframes pin-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

.district-map__spinner {
  width: 36px;
  height: 36px;
  border: 2px solid var(--color-border);
  border-top-color: var(--color-gold);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.district-map__hint {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-widest);
  background: color-mix(in srgb, var(--color-bg) 62%, transparent);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  padding: var(--space-2) var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  max-width: 86%;
  text-align: center;
}

@media (max-width: 768px) {
  .district-map {
    min-height: 320px;
  }
  .district-map__canvas {
    aspect-ratio: 4 / 3;
    min-height: 320px;
  }
}
</style>