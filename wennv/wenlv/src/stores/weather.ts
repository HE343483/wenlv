/** 天气状态 Store — 管理与后端天气 API 的交互 */

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { WeatherData, WeatherState } from '@/types'
import { get } from '@/api/request'

export interface LocationSelection {
  /** 省份(展示) */
  province: string
  /** 市(展示) */
  cityName: string
  /** 区/县名(展示) */
  districtName: string
  /** 区/县 adcode — 天气查询主键 */
  districtAdcode: string
  /** 镇/街道名(展示,可选 — 天气仍按所属区县查询) */
  streetName?: string
}

/** 天气结果缓存 TTL(高德预报按天更新,30 分钟足够新鲜) */
const WEATHER_CACHE_TTL = 30 * 60 * 1000
const WEATHER_CACHE_KEY = 'weather-cache-v1'

/** 天气结果缓存:adcode → {data,ts}, sessionStorage 持久化避免刷新即重查 */
const weatherCache = new Map<string, { data: WeatherData; ts: number }>()

function loadCache() {
  try {
    const raw = sessionStorage.getItem(WEATHER_CACHE_KEY)
    if (raw) {
      for (const [k, v] of JSON.parse(raw) as Array<[string, { data: WeatherData; ts: number }]>) {
        weatherCache.set(k, v)
      }
    }
  } catch {
    /* 缓存损坏则忽略 */
  }
}

function persistCache() {
  try {
    sessionStorage.setItem(WEATHER_CACHE_KEY, JSON.stringify([...weatherCache]))
  } catch {
    /* 存储满则放弃持久化 */
  }
}

loadCache()

export const useWeatherStore = defineStore('weather', () => {
  const state = ref<WeatherState>('idle')
  const data = ref<WeatherData | null>(null)
  const error = ref<string | null>(null)

  /** 级联定位(四川省 → 市 → 区县 → 镇/街道),天气按 districtAdcode 查询 */
  const location = ref<LocationSelection>({
    province: '四川省',
    cityName: '成都市',
    districtName: '锦江区',
    districtAdcode: '510104',
  })

  /** 下拉面板开关 */
  const open = ref(false)

  /** 展示名:优先镇/街道,否则区县 */
  const displayName = computed(() =>
    location.value.streetName ?? location.value.districtName,
  )

  /** 平均温度 = ({{min_temp}} + {{max_temp}}) / 2 — 全局色调据此派生 */
  const avgTemp = computed<number | null>(() =>
    data.value ? (data.value.min_temp + data.value.max_temp) / 2 : null,
  )

  /* 竞态防护：递增序号，仅采用最后一次请求的响应 */
  let fetchSeq = 0

  /* 区域筛选防抖：快速切换时仅发起最后一次请求 */
  let districtFetchTimer: ReturnType<typeof setTimeout> | null = null

  /* IP 自动定位只尝试一次（失败不再重试，保持默认成都锦江区） */
  let ipLocateDone = false

  /** 按客户端 IP 自动定位城市（后端高德 v3/ip 代理）。
   *  成功则把默认位置改写为访客所在市（IP 定位精确到市级，展示与查询均用市）；
   *  境外/内网 IP 或接口失败时静默保持默认位置。 */
  async function locateByIP() {
    if (ipLocateDone) return
    ipLocateDone = true
    try {
      const loc = await get<{ province: string; city: string; adcode: string }>('/map/ip-locate')
      if (loc?.adcode) {
        const city = loc.city || loc.province || '成都市'
        location.value = {
          province: loc.province || city,
          cityName: city,
          districtName: city,
          districtAdcode: loc.adcode,
        }
      }
    } catch {
      /* 定位失败保持默认位置 */
    }
  }

  /** 区域选择(行政区划级联) — 防抖拉取,命中缓存则零请求 */
  function selectRegion(sel: {
    cityName: string
    districtName: string
    districtAdcode: string
    streetName?: string
  }) {
    location.value = { province: '四川省', ...sel }
    if (districtFetchTimer) clearTimeout(districtFetchTimer)
    districtFetchTimer = setTimeout(() => {
      fetchWeather()
    }, 300)
  }

  /* ---- 下拉面板 ---- */
  function openDropdown() {
    open.value = true
    // 上次拉取失败(如后端暂不可用)时,展开面板自动重试
    if (state.value === 'error') void fetchWeather()
  }
  function closeDropdown() { open.value = false }
  function toggleDropdown() { open.value = !open.value }

  /** 拉取天气;force=true 跳过缓存强制刷新(面板刷新按钮) */
  async function fetchWeather(force = false) {
    const seq = ++fetchSeq
    // 首次拉取前先按 IP 定位城市,拿到真实位置后再查天气
    await locateByIP()
    if (seq !== fetchSeq) return // 定位期间发生了新的请求,放弃本次
    const adcode = location.value.districtAdcode
    if (!adcode) return
    state.value = 'loading'
    error.value = null

    if (!force) {
      const cached = weatherCache.get(adcode)
      if (cached && Date.now() - cached.ts < WEATHER_CACHE_TTL) {
        data.value = cached.data
        state.value = 'success'
        return
      }
    }

    try {
      // 后端 /api/map/weather 返回高德 4 天预报,取今天(day 0)渲染首页
      // 同时带 adcode(精确)与 city(旧版后端兼容/兜底)
      const list = await get<WeatherItem[]>('/map/weather', {
        params: { adcode, city: displayName.value },
      })
      if (seq !== fetchSeq) return // 丢弃过期响应
      const today = list?.[0]
      if (!today) {
        throw new Error('未获取到天气数据')
      }
      const result: WeatherData = {
        min_temp: Number(today.night_temp) || 0,
        max_temp: Number(today.day_temp) || 0,
        weather_icon: mapWeatherIcon(today.day_weather),
        weather_desc: today.day_weather || '--',
        city_name: displayName.value,
        humidity: undefined,
        wind_speed: undefined,
        wind_power: [today.wind_direction, today.wind_power]
          .filter(Boolean)
          .join(' ') || undefined,
        updated_at: new Date().toISOString(),
      }
      weatherCache.set(adcode, { data: result, ts: Date.now() })
      persistCache()
      data.value = result
      state.value = 'success'
    } catch (err) {
      if (seq !== fetchSeq) return
      state.value = 'error'
      error.value = err instanceof Error ? err.message : '获取天气失败'
      data.value = null
    }
  }

  return {
    state, data, error,
    location, displayName,
    open, avgTemp,
    selectRegion,
    openDropdown, closeDropdown, toggleDropdown,
    fetchWeather,
  }
})

/** 后端 /api/map/weather 返回的单日预报结构 */
interface WeatherItem {
  date: string
  day_weather: string
  night_weather: string
  day_temp: number | string
  night_temp: number | string
  wind_direction: string
  wind_power: string
}

/** 天气描述 → AppIcon 图标名（与行程页 ResultView 的图标归类规则一致） */
function mapWeatherIcon(text: string): string {
  if (!text) return 'sunny'
  if (/(雪|sleet|snow|冰雹|hail)/i.test(text)) return 'rainy'
  if (/(雨|rain|雷|thunder|shower)/i.test(text)) return 'rainy'
  if (/(云|阴|cloud|overcast|雾|霾|fog|mist|haze|风|breeze|gale)/i.test(text)) return 'cloudy'
  return 'sunny'
}
