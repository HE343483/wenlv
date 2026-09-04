<script setup lang="ts">
/**
 * RoutesPage.vue — 路线规划
 * 功能：精品路线 Tab + 自定义路线 Tab，左侧路线卡片/表单，右侧真实百度地图
 * 地图：百度地图 JS API v3.0，AK 从 .env 的 VITE_BAIDU_MAP_AK 读取
 *       - 精品路线：在地图上标注途经站点，连线展示
 *       - 自定义路线：支持“我的位置”定位 + 驾车路线规划
 */
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useLanguageStore } from '@/stores/language'
import { travelRoutes, scenicSpots, getScenicSpotById } from '@/data/chengdu'
import { loadBaiduMap } from '@/utils/baiduMap'
import AppIcon from '@/components/AppIcon.vue'
import HomeBanner from '@/components/HomeBanner.vue'
import type { TravelRoute, ScenicSpot } from '@/types'

const langStore = useLanguageStore()

/* ── 地图 AK ── */
const mapAK = (import.meta.env.VITE_BAIDU_MAP_AK as string | undefined) ?? ''

/* ── 页头统计 ── */
const routeStats = computed(() => {
  const themes = new Set(travelRoutes.map(r => r.themeZh))
  const coveredSpots = new Set(travelRoutes.flatMap(r => r.stops.map(s => s.spotId)))
  return [
    { value: String(travelRoutes.length), label: langStore.lang === 'zh' ? '精品路线' : 'Curated Routes' },
    { value: String(themes.size), label: langStore.lang === 'zh' ? '出行主题' : 'Themes' },
    { value: String(coveredSpots.size), label: langStore.lang === 'zh' ? '覆盖景点' : 'Covered Spots' },
    { value: String(scenicSpots.length), label: langStore.lang === 'zh' ? '可规划景点' : 'Plan-able Spots' },
  ]
})

/* ── Tab ── */
type RoutesTab = 'prebuilt' | 'custom'
const activeTab = ref<RoutesTab>('prebuilt')

/* ── 选中的精品路线 (默认第一条) ── */
const selectedRoute = ref<TravelRoute>(travelRoutes[0]!)

function pickRoute(route: TravelRoute) {
  selectedRoute.value = route
}

/* 当前路线途经景点（补齐 spot 详情，便于展示站名） */
const selectedStops = computed(() =>
  selectedRoute.value.stops
    .map(stop => ({
      ...stop,
      spot: getScenicSpotById(stop.spotId),
    }))
    .filter(s => !!s.spot)
)

/* 精品路线卡片精简信息 */
function routeStopsSummary(route: TravelRoute): string {
  return route.stops
    .map(stop => {
      const spot = getScenicSpotById(stop.spotId)
      if (!spot) return ''
      return langStore.lang === 'zh' ? spot.nameZh : spot.nameEn
    })
    .filter(Boolean)
    .join(' → ')
}

/* ============================================================
   百度地图
   ============================================================ */
type MapStatus = 'idle' | 'loading' | 'ready' | 'error'

/* 地图容器 */
const prebuiltMapEl = ref<HTMLDivElement | null>(null)
const customMapEl = ref<HTMLDivElement | null>(null)

/* 地图状态 */
const prebuiltMapStatus = ref<MapStatus>('idle')
const customMapStatus = ref<MapStatus>('idle')

/* 地图实例与图层 */
let prebuiltMap: any = null
let prebuiltOverlays: any[] = []
let customMap: any = null
let myLocationMarker: any = null
let routeSearch: any = null

/* 自定义路线规划结果 */
const routePlanned = ref(false)
const routeDistance = ref('')
const routeDuration = ref('')

/* ── 加载 BMap 脚本（幂等） ── */
async function ensureBMap(): Promise<boolean> {
  if (!mapAK) return false
  try {
    await loadBaiduMap(mapAK)
    return true
  } catch {
    return false
  }
}

/* ── 精品路线地图 ── */
async function initPrebuiltMap() {
  if (prebuiltMap || prebuiltMapStatus.value !== 'idle') return
  prebuiltMapStatus.value = 'loading'
  const ok = await ensureBMap()
  if (!ok) {
    prebuiltMapStatus.value = 'error'
    return
  }
  await nextTick()
  const el = prebuiltMapEl.value
  if (!el) return
  const BMap = (window as any).BMap
  prebuiltMap = new BMap.Map(el)
  prebuiltMap.enableScrollWheelZoom()
  prebuiltMap.addControl(new BMap.NavigationControl())
  prebuiltMap.addControl(new BMap.ScaleControl())
  renderPrebuiltRoute()
  prebuiltMapStatus.value = 'ready'
}

/* 渲染精品路线：标注站点 + 连线 */
function renderPrebuiltRoute() {
  if (!prebuiltMap) return
  const BMap = (window as any).BMap

  // 清除旧的覆盖物
  prebuiltOverlays.forEach(o => prebuiltMap.removeOverlay(o))
  prebuiltOverlays = []

  const points: any[] = []
  selectedStops.value.forEach(stop => {
    const spot = stop.spot!
    const pt = new BMap.Point(spot.coords.lng, spot.coords.lat)
    points.push(pt)

    const marker = new BMap.Marker(pt)
    const label = new BMap.Label(
      langStore.lang === 'zh' ? spot.nameZh : spot.nameEn,
      { position: pt, offset: new BMap.Size(18, -32) }
    )
    label.setStyle({
      color: '#0F0D0B',
      background: '#C9A96E',
      border: 'none',
      borderRadius: '4px',
      padding: '2px 8px',
      fontSize: '12px',
      letterSpacing: '0.08em',
      fontWeight: '600',
      whiteSpace: 'nowrap',
      boxShadow: '0 2px 8px rgba(0,0,0,0.35)',
    })
    marker.setLabel(label)

    marker.addEventListener('click', () => {
      const content =
        `<div style="font-family: 'Noto Serif SC', serif; min-width: 180px;">
          <h4 style="margin:0 0 6px; font-size:15px; color:#211D17;">${langStore.lang === 'zh' ? spot.nameZh : spot.nameEn}</h4>
          <p style="margin:0; font-size:12px; line-height:1.6; color:#6E6151;">${langStore.lang === 'zh' ? spot.shortDescZh : spot.shortDescEn}</p>
        </div>`
      const info = new BMap.InfoWindow(content)
      prebuiltMap.openInfoWindow(info, pt)
    })

    prebuiltMap.addOverlay(marker)
    prebuiltOverlays.push(marker)
  })

  // 站点连线
  if (points.length >= 2) {
    const line = new BMap.Polyline(points, {
      strokeColor: '#C9A96E',
      strokeWeight: 4,
      strokeOpacity: 0.85,
      strokeStyle: 'dashed',
    })
    prebuiltMap.addOverlay(line)
    prebuiltOverlays.push(line)
  }

  if (points.length > 0) {
    prebuiltMap.setViewport(points, { enableAnimation: true })
  }
}

/* ── 自定义路线地图 ── */
async function initCustomMap() {
  if (customMap || customMapStatus.value !== 'idle') return
  customMapStatus.value = 'loading'
  const ok = await ensureBMap()
  if (!ok) {
    customMapStatus.value = 'error'
    return
  }
  await nextTick()
  const el = customMapEl.value
  if (!el) return
  const BMap = (window as any).BMap
  customMap = new BMap.Map(el)
  customMap.enableScrollWheelZoom()
  customMap.addControl(new BMap.NavigationControl())
  customMap.addControl(new BMap.ScaleControl())
  // 默认以成都为中心
  customMap.centerAndZoom(new BMap.Point(104.065, 30.66), 11)
  customMapStatus.value = 'ready'

  // 如果之前点击了“我的位置”但地图未就绪，地图就绪后自动定位
  if (myLocation.value && locating.value) {
    locateMe()
  }
}

/* ── 自定义路线 ── */
const startSpotId = ref('')
const destSpotId = ref('')
const myLocation = ref(false)
const locating = ref(false)

/* 自定义路线绘制的覆盖物（路线折线 + 标记） */
let routeOverlays: any[] = []

const spotOptions = computed<ScenicSpot[]>(() => scenicSpots)

/* 定位：获取当前位置 */
function locateMe() {
  if (!customMap) return
  locating.value = true
  const BMap = (window as any).BMap
  const geo = new BMap.Geolocation()
  geo.getCurrentPosition((res: any) => {
    locating.value = false
    if (geo.getStatus() === (window as any).BMAP_STATUS_SUCCESS) {
      if (myLocationMarker && customMap) customMap.removeOverlay(myLocationMarker)
      const pt = new BMap.Point(res.point.lng, res.point.lat)
      myLocationMarker = new BMap.Marker(pt)
      const label = new BMap.Label(
        langStore.lang === 'zh' ? '我的位置' : 'My Location',
        { position: pt, offset: new BMap.Size(18, -32) }
      )
      label.setStyle({
        color: '#0F0D0B',
        background: '#7A8A7A',
        border: 'none',
        borderRadius: '4px',
        padding: '2px 8px',
        fontSize: '12px',
        letterSpacing: '0.08em',
        fontWeight: '600',
        whiteSpace: 'nowrap',
        boxShadow: '0 2px 8px rgba(0,0,0,0.35)',
      })
      myLocationMarker.setLabel(label)
      customMap.addOverlay(myLocationMarker)
      customMap.panTo(pt)
    } else {
      myLocation.value = false
      window.alert?.(langStore.t('routes.locationError'))
    }
  })
}

function toggleMyLocation() {
  myLocation.value = !myLocation.value
  if (!myLocation.value) {
    locating.value = false
    startSpotId.value = ''
    if (myLocationMarker && customMap) {
      customMap.removeOverlay(myLocationMarker)
      myLocationMarker = null
    }
    return
  }
  startSpotId.value = ''
  if (customMap) locateMe()
  else locating.value = true
}

/* 规划驾车路线 */
async function planRoute() {
  if (!customMap) {
    await initCustomMap()
  }
  if (!customMap) return
  if (!destSpotId.value) return

  const BMap = (window as any).BMap

  // 起点：我的位置 或 选择的起点景点
  let startPt: any = null
  let startName = ''
  if (myLocation.value) {
    if (!myLocationMarker) {
      window.alert?.(langStore.t('routes.noStart'))
      return
    }
    startPt = myLocationMarker.getPosition()
    startName = langStore.lang === 'zh' ? '我的位置' : 'My Location'
  } else if (startSpotId.value) {
    const start = getScenicSpotById(startSpotId.value)
    if (!start) {
      window.alert?.(langStore.t('routes.noStart'))
      return
    }
    startPt = new BMap.Point(start.coords.lng, start.coords.lat)
    startName = langStore.lang === 'zh' ? start.nameZh : start.nameEn
  } else {
    window.alert?.(langStore.t('routes.noStart'))
    return
  }

  const dest = getScenicSpotById(destSpotId.value)
  if (!dest) return
  const endPt = new BMap.Point(dest.coords.lng, dest.coords.lat)
  const endName = langStore.lang === 'zh' ? dest.nameZh : dest.nameEn

  // 清除上次规划结果
  if (routeSearch) {
    routeSearch.clearResults()
    routeSearch = null
  }
  clearCustomRouteOverlays()
  routePlanned.value = false

  // 使用 BMap.DrivingRoute 计算路线，关闭自动渲染，手动绘制品牌色路径
  const driving = new BMap.DrivingRoute(customMap, {
    onSearchComplete: (res: any) => {
      if (driving.getStatus() !== (window as any).BMAP_STATUS_SUCCESS) {
        window.alert?.(langStore.t('routes.planError'))
        return
      }
      const plan = res.getPlan(0)
      const route = plan.getRoute(0)
      const points = route.getPath()
      if (!points || points.length === 0) {
        window.alert?.(langStore.t('routes.planError'))
        return
      }

      // 金色路线折线
      const polyline = new BMap.Polyline(points, {
        strokeColor: '#C9A96E',
        strokeWeight: 5,
        strokeOpacity: 0.9,
      })
      customMap.addOverlay(polyline)
      routeOverlays.push(polyline)

      // 起点标记（黛绿）
      const startMarker = new BMap.Marker(points[0])
      const startLabel = new BMap.Label(startName, {
        position: points[0],
        offset: new BMap.Size(18, -30),
      })
      startLabel.setStyle({
        color: '#0F0D0B',
        background: '#7A8A7A',
        border: 'none',
        borderRadius: '4px',
        padding: '2px 8px',
        fontSize: '12px',
        letterSpacing: '0.08em',
        fontWeight: '600',
        whiteSpace: 'nowrap',
        boxShadow: '0 2px 8px rgba(0,0,0,0.35)',
      })
      startMarker.setLabel(startLabel)
      customMap.addOverlay(startMarker)
      routeOverlays.push(startMarker)

      // 终点标记（朱红）
      const endMarker = new BMap.Marker(points[points.length - 1])
      const endLabel = new BMap.Label(endName, {
        position: points[points.length - 1],
        offset: new BMap.Size(18, -30),
      })
      endLabel.setStyle({
        color: '#0F0D0B',
        background: '#A23B3B',
        border: 'none',
        borderRadius: '4px',
        padding: '2px 8px',
        fontSize: '12px',
        letterSpacing: '0.08em',
        fontWeight: '600',
        whiteSpace: 'nowrap',
        boxShadow: '0 2px 8px rgba(0,0,0,0.35)',
      })
      endMarker.setLabel(endLabel)
      customMap.addOverlay(endMarker)
      routeOverlays.push(endMarker)

      routeDistance.value = plan.getDistance(true)
      routeDuration.value = plan.getDuration(true)
      routePlanned.value = true

      // 视野自适应
      customMap.setViewport(points, { enableAnimation: true })
    },
  })
  driving.search(startPt, endPt)
  routeSearch = driving
}

/* 清除自定义路线覆盖物 */
function clearCustomRouteOverlays() {
  if (!customMap) return
  routeOverlays.forEach(o => customMap.removeOverlay(o))
  routeOverlays = []
}

/* ── Tab 切换时初始化对应地图 ── */
watch(activeTab, (tab) => {
  if (tab === 'prebuilt') initPrebuiltMap()
  else initCustomMap()
})

/* ── 语言切换：重绘标签 ── */
watch(() => langStore.lang, () => {
  renderPrebuiltRoute()
})

onMounted(() => {
  initPrebuiltMap()
})
</script>

<template>
  <div class="routes-page">
    <!-- ──── HERO ──── -->
    <HomeBanner
      :eyebrow="langStore.lang === 'zh' ? 'Tianfu Journey' : '天府之旅'"
      :title="langStore.t('routes.title')"
      :subtitle="langStore.t('routes.subtitle')"
      watermark="路"
    >
      <!-- 页头统计 -->
      <div class="routes-hero__stats">
        <div v-for="stat in routeStats" :key="stat.label" class="routes-hero__stat">
          <span class="routes-hero__stat-num">{{ stat.value }}</span>
          <span class="routes-hero__stat-label">{{ stat.label }}</span>
        </div>
      </div>
    </HomeBanner>

    <!-- ──── TAB 切换 ──── -->
    <section class="routes-tabs container">
      <div class="routes-tabs__bar">
        <button
          class="routes-tabs__btn"
          :class="{ 'routes-tabs__btn--active': activeTab === 'prebuilt' }"
          @click="activeTab = 'prebuilt'"
        >
          <span class="routes-tabs__icon"><AppIcon name="map" :size="18" /></span>
          {{ langStore.t('routes.tabPrebuilt') }}
        </button>
        <button
          class="routes-tabs__btn"
          :class="{ 'routes-tabs__btn--active': activeTab === 'custom' }"
          @click="activeTab = 'custom'"
        >
          <span class="routes-tabs__icon"><AppIcon name="edit" :size="18" /></span>
          {{ langStore.t('routes.tabCustom') }}
        </button>
      </div>
    </section>

    <!-- ════════════ 精品路线 ════════════ -->
    <section v-if="activeTab === 'prebuilt'" class="routes-prebuilt">
      <!-- 头部说明 -->
      <header class="routes-prebuilt__head container">
        <div>
          <h2 class="section-title">{{ langStore.t('routes.prebuiltTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('routes.prebuiltSubtitle') }}</p>
        </div>
      </header>

      <div class="routes-prebuilt__layout container">
        <!-- 左：路线卡片列表 -->
        <aside class="routes-prebuilt__list">
          <article
            v-for="route in travelRoutes"
            :key="route.id"
            class="route-card"
            :class="{ 'route-card--active': selectedRoute.id === route.id }"
            @click="pickRoute(route)"
          >
            <div class="route-card__head">
              <div>
                <h3 class="route-card__name">
                  {{ langStore.lang === 'zh' ? route.nameZh : route.nameEn }}
                </h3>
                <p class="route-card__meta">
                  {{ routeStopsSummary(route) }}
                </p>
              </div>
              <span class="route-card__chevron">›</span>
            </div>

            <div class="route-card__badges">
              <span class="route-card__badge route-card__badge--stops">
                <AppIcon name="pin" :size="14" />
                {{ route.stops.length }}
                {{ langStore.lang === 'zh' ? '站' : 'stops' }}
              </span>
              <span class="route-card__badge">
                <b>{{ langStore.t('routes.duration') }}</b>
                {{ langStore.lang === 'zh' ? route.durationZh : route.durationEn }}
              </span>
              <span class="route-card__badge">
                <b>{{ langStore.t('routes.theme') }}</b>
                {{ langStore.lang === 'zh' ? route.themeZh : route.themeEn }}
              </span>
              <span class="route-card__badge">
                <b>{{ langStore.t('routes.level') }}</b>
                {{ langStore.lang === 'zh' ? route.levelZh : route.levelEn }}
              </span>
            </div>
          </article>
        </aside>

        <!-- 右：地图 + 行程 -->
        <div class="routes-prebuilt__detail">
          <!-- 百度地图 -->
          <div class="route-map">
            <div ref="prebuiltMapEl" class="route-map__canvas" />
            <!-- 加载中 / 失败覆盖层 -->
            <div v-if="prebuiltMapStatus !== 'ready'" class="route-map__overlay">
              <template v-if="prebuiltMapStatus === 'loading'">
                <div class="route-map__spinner" aria-hidden="true" />
                <p class="route-map__hint">{{ langStore.t('routes.mapLoading') }}</p>
              </template>
              <template v-else>
                <span class="route-map__pin"><AppIcon name="pin" :size="40" /></span>
                <p class="route-map__hint">
                  {{ mapAK ? langStore.t('routes.mapError') : langStore.t('routes.mapNoKey') }}
                </p>
              </template>
            </div>
          </div>

          <!-- 行程安排 -->
          <div class="route-detail">
            <div class="route-detail__head">
              <h3 class="route-detail__title">
                {{ langStore.lang === 'zh' ? selectedRoute.nameZh : selectedRoute.nameEn }}
              </h3>
              <div class="route-detail__badges">
                <span class="route-detail__badge">
                  {{ langStore.lang === 'zh' ? selectedRoute.durationZh : selectedRoute.durationEn }}
                </span>
                <span class="route-detail__badge">
                  {{ langStore.lang === 'zh' ? selectedRoute.levelZh : selectedRoute.levelEn }}
                </span>
              </div>
            </div>

            <p class="route-detail__summary">
              {{ langStore.lang === 'zh' ? selectedRoute.summaryZh : selectedRoute.summaryEn }}
            </p>

            <!-- 行程步骤 -->
            <ol class="route-detail__stops">
              <li
                v-for="stop in selectedStops"
                :key="stop.spotId"
                class="route-stop"
              >
                <span class="route-stop__rail">
                  <span class="route-stop__dot" />
                </span>
                <div class="route-stop__body">
                  <span class="route-stop__note">
                    {{ langStore.lang === 'zh' ? stop.noteZh : stop.noteEn }}
                  </span>
                  <span class="route-stop__name">
                    {{ langStore.lang === 'zh' ? stop.spot!.nameZh : stop.spot!.nameEn }}
                  </span>
                </div>
              </li>
            </ol>

            <!-- 亮点 -->
            <div class="route-detail__highlights">
              <span
                v-for="hl in (langStore.lang === 'zh' ? selectedRoute.highlightsZh : selectedRoute.highlightsEn)"
                :key="hl"
                class="route-detail__highlight"
              >
                ✦ {{ hl }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ════════════ 自定义路线 ════════════ -->
    <section v-else class="routes-custom">
      <header class="routes-custom__head container">
        <div>
          <h2 class="section-title">{{ langStore.t('routes.customTitle') }}</h2>
          <p class="section-subtitle">{{ langStore.t('routes.customSubtitle') }}</p>
        </div>
      </header>

      <div class="routes-custom__layout container">
        <!-- 左：规划表单 -->
        <aside class="routes-custom__panel">
          <!-- 起点 -->
          <label class="route-field">
            <span class="route-field__key">
              <span class="route-field__dot route-field__dot--start" />
              {{ langStore.t('routes.start') }}
            </span>
            <button
              class="route-field__locate"
              :class="{ 'route-field__locate--on': myLocation }"
              :disabled="locating"
              @click="toggleMyLocation"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="12" cy="12" r="3"/>
                <path d="M12 2v3M12 19v3M2 12h3M19 12h3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M19.1 4.9L17 7M7 17l-2.1 2.1"/>
              </svg>
              {{ locating
                ? langStore.t('routes.locating')
                : (myLocation ? langStore.t('routes.locationSuccess') : langStore.t('routes.myLocation')) }}
            </button>
          </label>

          <div class="route-field">
            <select
              v-model="startSpotId"
              class="route-field__select"
              :disabled="myLocation"
            >
              <option value="" disabled>
                {{ myLocation
                  ? langStore.t('routes.locationSuccess')
                  : langStore.t('routes.startPlaceholder') }}
              </option>
              <option v-for="s in spotOptions" :key="s.id" :value="s.id">
                {{ langStore.lang === 'zh' ? s.nameZh : s.nameEn }}
              </option>
            </select>
          </div>

          <!-- 终点 -->
          <div class="route-field">
            <span class="route-field__key">
              <span class="route-field__dot route-field__dot--dest" />
              {{ langStore.t('routes.destination') }}
            </span>
            <select
              v-model="destSpotId"
              class="route-field__select"
            >
              <option value="" disabled>{{ langStore.t('routes.destinationPlaceholder') }}</option>
              <option v-for="s in spotOptions" :key="s.id" :value="s.id">
                {{ langStore.lang === 'zh' ? s.nameZh : s.nameEn }}
              </option>
            </select>
          </div>

          <!-- 规划按钮 -->
          <button class="route-field__plan" :disabled="!destSpotId" @click="planRoute">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M21 3L3 10.5l7.5 2.5L13 20.5 21 3z"/>
              <path d="M10.5 13l10-10"/>
            </svg>
            {{ langStore.t('routes.plan') }}
          </button>

          <!-- 规划结果 -->
          <div v-if="routePlanned" class="route-result">
            <span class="route-result__item">
              {{ langStore.t('routes.totalDistance') }}
              <b>{{ routeDistance }}</b>
            </span>
            <span class="route-result__item">
              {{ langStore.t('routes.totalDuration') }}
              <b>{{ routeDuration }}</b>
            </span>
          </div>

          <!-- API 对接说明（占位提示） -->
          <p class="route-field__hint">{{ langStore.t('routes.hint') }}</p>
        </aside>

        <!-- 右：地图 -->
        <div class="routes-custom__map">
          <div class="route-map">
            <div ref="customMapEl" class="route-map__canvas route-map__canvas--custom" />
            <!-- 加载中 / 失败覆盖层 -->
            <div v-if="customMapStatus !== 'ready'" class="route-map__overlay">
              <template v-if="customMapStatus === 'loading'">
                <div class="route-map__spinner" aria-hidden="true" />
                <p class="route-map__hint">{{ langStore.t('routes.mapLoading') }}</p>
              </template>
              <template v-else>
                <span class="route-map__pin"><AppIcon name="pin" :size="40" /></span>
                <p class="route-map__hint">
                  {{ mapAK ? langStore.t('routes.mapError') : langStore.t('routes.mapNoKey') }}
                </p>
              </template>
            </div>
          </div>
        </div>
    </div>
    </section>
  </div>
</template>

<style scoped>
/* ========================================
   页头统计（欢迎横幅内）
   ======================================== */
.routes-hero__stats {
  display: flex;
  gap: var(--space-1);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  padding: var(--space-2) var(--space-3);
}

.routes-hero__stat {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  transition: background var(--transition-fast);
}

.routes-hero__stat:hover {
  background: var(--color-surface-hover);
}

.routes-hero__stat + .routes-hero__stat {
  border-left: 1px solid var(--color-border);
  padding-left: var(--space-5);
}

.routes-hero__stat-num {
  font-family: var(--font-en-display);
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-gold);
  line-height: 1;
}

.routes-hero__stat-label {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
}

/* ========================================
   TABS
   ======================================== */
.routes-tabs {
  padding-bottom: var(--space-8);
}

.routes-tabs__bar {
  display: inline-flex;
  gap: var(--space-2);
  padding: var(--space-1);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
}

.routes-tabs__btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-muted);
  padding: var(--space-2) var(--space-5);
  border-radius: var(--radius-full);
  transition: all var(--transition-base);
  white-space: nowrap;
}

.routes-tabs__btn:hover {
  color: var(--color-gold);
}

.routes-tabs__btn--active {
  background: var(--color-gold);
  color: var(--color-text-inverse);
}

.routes-tabs__btn--active:hover {
  color: var(--color-text-inverse);
}

.routes-tabs__icon {
  font-size: var(--text-base);
  line-height: 1;
}

/* ========================================
   区块头部
   ======================================== */
.routes-prebuilt__head,
.routes-custom__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

/* ========================================
   双栏布局
   ======================================== */
.routes-prebuilt__layout,
.routes-custom__layout {
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: var(--space-6);
  align-items: start;
  padding-bottom: var(--space-16);
}

/* ========================================
   精品路线 — 卡片列表
   ======================================== */
.routes-prebuilt__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  max-height: 720px;
  overflow-y: auto;
  padding-right: var(--space-2);
}

.route-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  cursor: pointer;
  transition: all var(--transition-base);
  text-align: left;
}

.route-card:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px var(--color-gold-glow);
}

.route-card--active {
  border-color: var(--color-gold);
  box-shadow: inset 0 0 0 1px var(--color-gold), 0 6px 20px var(--color-gold-glow);
}

.route-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.route-card__name {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  margin-bottom: var(--space-2);
}

.route-card__meta {
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-wide);
  line-height: var(--leading-normal);
}

.route-card__chevron {
  font-size: var(--text-2xl);
  color: var(--color-text-muted);
  transition: color var(--transition-fast), transform var(--transition-fast);
  flex-shrink: 0;
  line-height: 1;
}

.route-card:hover .route-card__chevron,
.route-card--active .route-card__chevron {
  color: var(--color-gold);
}

.route-card--active .route-card__chevron {
  transform: translateX(3px);
}

.route-card__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.route-card__badge {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  padding: 3px 10px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  letter-spacing: var(--tracking-wide);
}

.route-card__badge--stops {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  color: var(--color-gold);
  border-color: color-mix(in srgb, var(--color-gold) 40%, transparent);
}

.route-card__badge b {
  font-weight: 500;
  color: var(--color-text-muted);
  margin-right: var(--space-1);
}

/* ========================================
   地图
   ======================================== */
.route-map {
  position: relative;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.route-map__canvas {
  width: 100%;
  aspect-ratio: 16 / 8;
  background: var(--color-bg-alt);
}

.route-map__canvas--custom {
  aspect-ratio: 16 / 8;
}

/* 地图覆盖层（加载中 / 未配置 AK / 加载失败） */
.route-map__overlay {
  position: absolute;
  inset: 0;
  z-index: 5;
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
}

.route-map__spinner {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 3px solid var(--color-border-light);
  border-top-color: var(--color-gold);
  animation: map-spin 0.9s linear infinite;
}

@keyframes map-spin {
  to { transform: rotate(360deg); }
}

.route-map__pin {
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

.route-map__hint {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  text-align: center;
  background: color-mix(in srgb, var(--color-bg) 62%, transparent);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  padding: var(--space-2) var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  max-width: 86%;
}

/* ========================================
   行程详情
   ======================================== */
.route-detail {
  margin-top: var(--space-6);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.route-detail__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.route-detail__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.route-detail__badges {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.route-detail__badge {
  font-size: var(--text-xs);
  padding: 3px 10px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-gold-dark);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-wide);
}

.route-detail__summary {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
}

/* 行程步骤 */
.route-detail__stops {
  display: flex;
  flex-direction: column;
}

.route-stop {
  display: flex;
  gap: var(--space-4);
}

.route-stop__rail {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  width: 12px;
}

.route-stop__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--color-gold);
  border: 2px solid color-mix(in srgb, var(--color-gold) 35%, transparent);
  box-shadow: 0 0 8px var(--color-gold-glow);
  margin-top: 6px;
}

.route-stop:not(:last-child) .route-stop__rail::after {
  content: '';
  flex: 1;
  width: 1px;
  background: linear-gradient(180deg, var(--color-gold-dark) 0%, transparent 100%);
  opacity: 0.4;
  min-height: 20px;
}

.route-stop__body {
  display: flex;
  align-items: baseline;
  gap: var(--space-4);
  padding: var(--space-3) 0;
  flex-wrap: wrap;
}

.route-stop__note {
  font-size: var(--text-xs);
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
  min-width: 96px;
}

.route-stop__name {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

/* 亮点 */
.route-detail__highlights {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  border-top: 1px solid var(--color-border);
  padding-top: var(--space-4);
}

.route-detail__highlight {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  padding: 3px 12px;
  border-radius: var(--radius-full);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  letter-spacing: var(--tracking-wide);
}

/* ========================================
   自定义路线 — 表单
   ======================================== */
.routes-custom__panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.route-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.route-field__key {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-family: var(--font-display);
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.route-field__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.route-field__dot--start {
  background: var(--color-sage);
  box-shadow: 0 0 8px color-mix(in srgb, var(--color-sage) 50%, transparent);
}

.route-field__dot--dest {
  background: var(--color-cinnabar);
  box-shadow: 0 0 8px color-mix(in srgb, var(--color-cinnabar) 50%, transparent);
}

.route-field__select {
  width: 100%;
  appearance: none;
  -webkit-appearance: none;
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  outline: none;
  cursor: pointer;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.route-field__select:focus {
  border-color: var(--color-gold);
  box-shadow: 0 0 0 2px var(--color-gold-glow);
}

.route-field__select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.route-field__select option {
  background: var(--color-surface);
  color: var(--color-text-primary);
}

.route-field__locate {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  border: 1px dashed var(--color-border-light);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
  align-self: flex-start;
}

.route-field__locate:hover {
  color: var(--color-gold);
  border-color: var(--color-gold-dark);
}

.route-field__locate--on {
  color: var(--color-gold);
  border-color: var(--color-gold);
  border-style: solid;
  background: color-mix(in srgb, var(--color-gold) 8%, transparent);
}

.route-field__plan {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-inverse);
  background: var(--color-gold);
  padding: var(--space-3);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.route-field__plan:hover:not(:disabled) {
  background: var(--color-gold-light);
  box-shadow: 0 4px 16px var(--color-gold-glow);
}

.route-field__plan:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* 规划结果 */
.route-result {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-4);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.route-result__item b {
  color: var(--color-gold);
  font-weight: 600;
  margin-left: var(--space-2);
}

.route-field__hint {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  text-align: center;
  line-height: var(--leading-relaxed);
}

/* ========================================
   Responsive
   ======================================== */
@media (max-width: 1024px) {
  .routes-prebuilt__layout,
  .routes-custom__layout {
    grid-template-columns: 1fr;
  }
  .routes-prebuilt__list {
    max-height: none;
    overflow: visible;
    display: grid;
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .routes-hero__stats {
    flex-wrap: wrap;
    justify-content: center;
    border-radius: var(--radius-lg);
    gap: 0;
  }
  .routes-hero__stat + .routes-hero__stat {
    border-left: none;
  }
  .routes-hero__stat {
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-2) var(--space-4);
  }
  .routes-tabs__bar {
    display: flex;
    width: 100%;
  }
  .routes-tabs__btn {
    flex: 1;
    justify-content: center;
  }
  .routes-prebuilt__list {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .routes-prebuilt__layout,
  .routes-custom__layout {
    padding-bottom: var(--space-10);
  }
  .route-map__canvas,
  .route-map__canvas--custom {
    aspect-ratio: 4 / 3;
  }
  .route-stop__body {
    flex-direction: column;
    gap: var(--space-1);
  }
  .route-stop__note {
    min-width: 0;
  }
}
</style>
