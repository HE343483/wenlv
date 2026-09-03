/** 天气状态 Store — 管理与后端天气 API 的交互 */

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { WeatherData, WeatherState } from '@/types'
import { useLanguageStore } from './language'
import { getDistrictById } from '@/data/chengdu'

export interface LocationSelection {
  province: string
  city: string
  /** 区级唯一标识 — 查询天气的主键 */
  districtId: string
  /** 区名 — 仅用于展示 */
  districtName: string
}

export const useWeatherStore = defineStore('weather', () => {
  const state = ref<WeatherState>('idle')
  const data = ref<WeatherData | null>(null)
  const error = ref<string | null>(null)

  /** 级联定位（省/市仅展示，查询以 districtId 为准） */
  const location = ref<LocationSelection>({
    province: '四川省',
    city: '成都市',
    districtId: 'jinjiang',
    districtName: '锦江区',
  })

  /** 下拉面板开关 */
  const open = ref(false)

  /** 平均温度 = ({{min_temp}} + {{max_temp}}) / 2 — 全局色调据此派生 */
  const avgTemp = computed<number | null>(() =>
    data.value ? (data.value.min_temp + data.value.max_temp) / 2 : null,
  )

  /* 竞态防护：递增序号，仅采用最后一次请求的响应 */
  let fetchSeq = 0

  /* 区域筛选防抖：快速切换时仅发起最后一次请求 */
  let districtFetchTimer: ReturnType<typeof setTimeout> | null = null

  /** 区域筛选 — 传区级唯一标识（防抖拉取，定位名即时更新） */
  function selectDistrict(districtId: string) {
    location.value = { ...location.value, districtId }
    if (districtFetchTimer) clearTimeout(districtFetchTimer)
    districtFetchTimer = setTimeout(() => {
      fetchWeather(districtId)
    }, 300)
  }

  /* ---- 下拉面板 ---- */
  function openDropdown() { open.value = true }
  function closeDropdown() { open.value = false }
  function toggleDropdown() { open.value = !open.value }

  async function fetchWeather(districtId = location.value.districtId) {
    const seq = ++fetchSeq
    state.value = 'loading'
    error.value = null

    try {
      const response = await mockWeatherApi(districtId)
      if (seq !== fetchSeq) return // 丢弃过期响应
      data.value = response
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
    location,
    open, avgTemp,
    selectDistrict,
    openDropdown, closeDropdown, toggleDropdown,
    fetchWeather,
  }
})

/** 模拟天气 API — 开发占位，对接后删除 */
async function mockWeatherApi(districtId: string): Promise<WeatherData> {
  await new Promise(resolve => setTimeout(resolve, 600))
  const langStore = useLanguageStore()
  const district = getDistrictById(districtId)

  const tempMap: Record<string, number> = {
    jinjiang: 28, wuhou: 27, qingyang: 28, jinniu: 27,
    chenghua: 27, gaoxin: 28, longquanyi: 26, tianfu: 27,
    wenjiang: 26, shuangliu: 27, pidu: 26, xindu: 27,
    dujiangyan: 24, qingbaijiang: 26, dayi: 23,
  }
  const base = tempMap[districtId] ?? 27
  const min = base - 4
  const max = base + 3

  return {
    min_temp: min,
    max_temp: max,
    weather_icon: 'sunny',
    weather_desc: langStore.lang === 'zh' ? '晴' : 'Sunny',
    city_name: district?.nameZh ?? districtId,
    humidity: 50 + Math.floor(Math.random() * 15),
    wind_speed: Math.round((2 + Math.random() * 3) * 100) / 100,
    updated_at: new Date().toISOString(),
  }
}
