import axios from 'axios'
import type {
  BackendRuntimeSettings,
  RuntimeSettings,
  TripFormData,
  TripHistoryItem,
  TripPlanResponse,
  TripTaskEvent,
} from '@/types/trip'
import { i18n } from '@/i18n'

const ENV_API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''
const ENV_AMAP_WEB_JS_KEY = import.meta.env.VITE_AMAP_WEB_JS_KEY ?? ''
const RUNTIME_API_BASE_STORAGE_KEY = 'tripstar.runtime.api_base_url'
const RUNTIME_AMAP_WEB_JS_KEY_STORAGE_KEY = 'tripstar.runtime.amap_web_js_key'
const RUNTIME_GOOGLE_MAPS_API_KEY_STORAGE_KEY = 'tripstar.runtime.google_maps_api_key'
const DEFAULT_RUNTIME_BACKEND_SETTINGS: BackendRuntimeSettings = {
  vite_amap_web_key: '',
  vite_amap_web_js_key: '',
  google_maps_api_key: '',
  google_maps_proxy: '',
  xhs_cookie: '',
  douyin_cookie: '',
  openai_api_key: '',
  openai_base_url: '',
  openai_model: '',
}

export const RUNTIME_SETTINGS_UPDATED_EVENT = 'tripstar:runtime-settings-updated'
const t = i18n.global.t

const normalizeBaseUrl = (value: string | null | undefined): string => {
  const text = String(value ?? '').trim()
  return text.replace(/\/+$/, '')
}

const normalizeText = (value: unknown): string => String(value ?? '').trim()

const resolveDefaultApiBaseUrl = (): string => {
  const fromEnv = normalizeBaseUrl(ENV_API_BASE_URL)
  if (fromEnv) return fromEnv
  // 同源部署（Docker / 云端）：API 与前端在同一 origin 下
  if (typeof window !== 'undefined' && window.location) {
    return normalizeBaseUrl(window.location.origin) || ''
  }
  // 仅本地开发 fallback
  return 'http://localhost:8000'
}

const DEFAULT_API_BASE_URL = resolveDefaultApiBaseUrl()
const DEFAULT_AMAP_WEB_JS_KEY = normalizeText(ENV_AMAP_WEB_JS_KEY)

interface SubmitTripPlanResponse {
  task_id: string
  plan_id: string
  status: 'processing'
  ws_url: string
  message: string
}

interface GenerateTripPlanOptions {
  onTaskCreated?: (task: SubmitTripPlanResponse) => void
  onTaskEvent?: (event: TripTaskEvent) => void
}

interface RuntimeSettingsApiResponse {
  success: boolean
  message?: string
  data?: Partial<BackendRuntimeSettings>
}

interface TripHistoryResponse {
  items?: TripHistoryItem[]
}

export const getRuntimeApiBaseUrl = (): string => {
  if (typeof window === 'undefined') {
    return DEFAULT_API_BASE_URL
  }
  const saved = normalizeBaseUrl(window.localStorage.getItem(RUNTIME_API_BASE_STORAGE_KEY))
  return saved || DEFAULT_API_BASE_URL
}

export const setRuntimeApiBaseUrl = (value: string): string => {
  const normalized = normalizeBaseUrl(value) || DEFAULT_API_BASE_URL
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(RUNTIME_API_BASE_STORAGE_KEY, normalized)
  }
  return normalized
}

export const getRuntimeMapJsKey = (): string => {
  if (typeof window === 'undefined') {
    return DEFAULT_AMAP_WEB_JS_KEY
  }
  const saved = normalizeText(window.localStorage.getItem(RUNTIME_AMAP_WEB_JS_KEY_STORAGE_KEY))
  return saved || DEFAULT_AMAP_WEB_JS_KEY
}

export const setRuntimeMapJsKey = (value: string): string => {
  const normalized = normalizeText(value)
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(RUNTIME_AMAP_WEB_JS_KEY_STORAGE_KEY, normalized)
  }
  return normalized
}

export const getRuntimeGoogleMapsApiKey = (): string => {
  if (typeof window === 'undefined') return ''
  return normalizeText(window.localStorage.getItem(RUNTIME_GOOGLE_MAPS_API_KEY_STORAGE_KEY))
}

export const setRuntimeGoogleMapsApiKey = (value: string): string => {
  const normalized = normalizeText(value)
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(RUNTIME_GOOGLE_MAPS_API_KEY_STORAGE_KEY, normalized)
  }
  return normalized
}

const getWsBaseUrl = (): string => getRuntimeApiBaseUrl().replace(/^http/i, 'ws').replace(/\/+$/, '')

// ========== 用户记忆：user_id 与偏好记忆开关 ==========
const USER_ID_STORAGE_KEY = 'tripstar.user_id'
const MEMORY_ENABLED_STORAGE_KEY = 'tripstar.memory_enabled'

/** 登录用户的 user_id 前缀:优先使用登录用户名,保证跨浏览器/清缓存后偏好仍关联同一账号 */
const AUTH_USER_STORAGE_KEY = 'shuyun-chengdu-user'

function getLoggedInUsername(): string {
  try {
    const raw = window.localStorage.getItem(AUTH_USER_STORAGE_KEY)
    if (!raw) return ''
    const parsed = JSON.parse(raw) as { profile?: { nickname?: string } }
    return parsed.profile?.nickname?.trim() || ''
  } catch {
    return ''
  }
}

/**
 * 获取或创建 user_id:
 * - 已登录:使用 `user-{用户名}`,偏好记忆跟随账号,清缓存/换设备后重新登录即可找回
 * - 未登录:使用本地持久化的匿名 UUID
 */
export const getOrCreateUserId = (): string => {
  if (typeof window === 'undefined') return ''
  const username = getLoggedInUsername()
  if (username) return `user-${username}`
  let uid = window.localStorage.getItem(USER_ID_STORAGE_KEY)
  if (!uid) {
    const cryptoObj = window.crypto as Crypto | undefined
    uid =
      cryptoObj && typeof cryptoObj.randomUUID === 'function'
        ? cryptoObj.randomUUID()
        : `${Date.now()}-${Math.random().toString(36).slice(2)}`
    window.localStorage.setItem(USER_ID_STORAGE_KEY, uid)
  }
  return uid
}

/** 用户偏好记忆开关(默认关闭,需用户主动开启) */
export const isMemoryEnabled = (): boolean => {
  if (typeof window === 'undefined') return false
  return window.localStorage.getItem(MEMORY_ENABLED_STORAGE_KEY) === 'true'
}

export const setMemoryEnabled = (enabled: boolean): void => {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(MEMORY_ENABLED_STORAGE_KEY, enabled ? 'true' : 'false')
}

const normalizeBackendRuntimeSettings = (
  data?: Partial<BackendRuntimeSettings>
): BackendRuntimeSettings => ({
  vite_amap_web_key: normalizeText(data?.vite_amap_web_key ?? DEFAULT_RUNTIME_BACKEND_SETTINGS.vite_amap_web_key),
  vite_amap_web_js_key: normalizeText(
    data?.vite_amap_web_js_key ?? DEFAULT_RUNTIME_BACKEND_SETTINGS.vite_amap_web_js_key
  ),
  google_maps_api_key: normalizeText(
    data?.google_maps_api_key ?? DEFAULT_RUNTIME_BACKEND_SETTINGS.google_maps_api_key
  ),
  google_maps_proxy: normalizeText(
    data?.google_maps_proxy ?? DEFAULT_RUNTIME_BACKEND_SETTINGS.google_maps_proxy
  ),
  xhs_cookie: normalizeText(data?.xhs_cookie ?? DEFAULT_RUNTIME_BACKEND_SETTINGS.xhs_cookie),
  douyin_cookie: normalizeText(
    data?.douyin_cookie ?? DEFAULT_RUNTIME_BACKEND_SETTINGS.douyin_cookie
  ),
  openai_api_key: normalizeText(data?.openai_api_key ?? DEFAULT_RUNTIME_BACKEND_SETTINGS.openai_api_key),
  openai_base_url:
    normalizeText(data?.openai_base_url ?? DEFAULT_RUNTIME_BACKEND_SETTINGS.openai_base_url) ||
    DEFAULT_RUNTIME_BACKEND_SETTINGS.openai_base_url,
  openai_model:
    normalizeText(data?.openai_model ?? DEFAULT_RUNTIME_BACKEND_SETTINGS.openai_model) ||
    DEFAULT_RUNTIME_BACKEND_SETTINGS.openai_model,
})

const emitRuntimeSettingsUpdated = () => {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent(RUNTIME_SETTINGS_UPDATED_EVENT))
}

const apiClient = axios.create({
  timeout: 0, // 无超时限制，等待后端返回结果
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    config.baseURL = getRuntimeApiBaseUrl()
    if (import.meta.env.DEV) {
      console.log('发送请求:', config.method?.toUpperCase(), config.url)
    }
    return config
  },
  (error) => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => {
    if (import.meta.env.DEV) {
      console.log('收到响应:', response.status, response.config.url)
    }
    return response
  },
  (error) => {
    console.error('响应错误:', error.response?.status, error.message)
    return Promise.reject(error)
  }
)

export async function getBackendRuntimeSettings(): Promise<BackendRuntimeSettings> {
  try {
    const response = await apiClient.get<RuntimeSettingsApiResponse>('/api/settings')
    return normalizeBackendRuntimeSettings(response.data?.data)
  } catch (error: any) {
    console.error('读取运行时配置失败:', error)
    throw new Error(error.response?.data?.detail || error.message || '读取配置失败')
  }
}

export async function updateBackendRuntimeSettings(
  updates: Partial<BackendRuntimeSettings>
): Promise<BackendRuntimeSettings> {
  try {
    const response = await apiClient.put<RuntimeSettingsApiResponse>('/api/settings', updates)
    return normalizeBackendRuntimeSettings(response.data?.data)
  } catch (error: any) {
    console.error('保存运行时配置失败:', error)
    throw new Error(error.response?.data?.detail || error.message || '保存配置失败')
  }
}

export async function getRuntimeSettings(): Promise<RuntimeSettings> {
  const backend = await getBackendRuntimeSettings()
  const apiBaseUrl = getRuntimeApiBaseUrl()
  const mapJsKey = getRuntimeMapJsKey() || backend.vite_amap_web_js_key

  // 同步 Google Maps API Key 到 localStorage 供前端地图组件读取
  if (backend.google_maps_api_key) {
    setRuntimeGoogleMapsApiKey(backend.google_maps_api_key)
  }

  return {
    api_base_url: apiBaseUrl,
    ...backend,
    vite_amap_web_js_key: mapJsKey,
  }
}

export async function saveRuntimeSettings(settings: RuntimeSettings): Promise<RuntimeSettings> {
  const previousApiBaseUrl = getRuntimeApiBaseUrl()
  const targetApiBaseUrl = normalizeBaseUrl(settings.api_base_url) || previousApiBaseUrl
  const updates: Partial<BackendRuntimeSettings> = {
    vite_amap_web_key: settings.vite_amap_web_key,
    vite_amap_web_js_key: settings.vite_amap_web_js_key,
    google_maps_api_key: settings.google_maps_api_key,
    google_maps_proxy: settings.google_maps_proxy,
    xhs_cookie: settings.xhs_cookie,
    douyin_cookie: settings.douyin_cookie,
    openai_api_key: settings.openai_api_key,
    openai_base_url: settings.openai_base_url,
    openai_model: settings.openai_model,
  }
  setRuntimeApiBaseUrl(targetApiBaseUrl)

  let backend: BackendRuntimeSettings
  try {
    backend = await updateBackendRuntimeSettings(updates)
  } catch (error) {
    setRuntimeApiBaseUrl(previousApiBaseUrl)
    throw error
  }

  const apiBaseUrl = setRuntimeApiBaseUrl(targetApiBaseUrl)
  const mapJsKey = setRuntimeMapJsKey(settings.vite_amap_web_js_key || backend.vite_amap_web_js_key)
  setRuntimeGoogleMapsApiKey(settings.google_maps_api_key || backend.google_maps_api_key)

  emitRuntimeSettingsUpdated()

  return {
    api_base_url: apiBaseUrl,
    ...backend,
    vite_amap_web_js_key: mapJsKey || backend.vite_amap_web_js_key,
  }
}

/**
 * 提交旅行规划任务（立即返回 task_id）
 */
export async function submitTripPlan(formData: TripFormData): Promise<SubmitTripPlanResponse> {
  try {
    const payload = { ...formData, user_id: getOrCreateUserId(), memory_enabled: isMemoryEnabled() }
    const response = await apiClient.post('/api/trip/plan', payload)
    return response.data
  } catch (error: any) {
    console.error('提交旅行计划失败:', error)
    throw new Error(error.response?.data?.detail || error.message || t('api.submitTripPlanFailed'))
  }
}

/**
 * 轮询任务状态
 */
export async function pollTaskStatus(taskId: string): Promise<any> {
  try {
    // 携带当前语言,后端对缺失 overall_suggestions 的历史任务快照按语言生成兜底文案
    const lang = i18n.global.locale.value
    const response = await apiClient.get(`/api/trip/status/${taskId}`, { params: { lang } })
    return response.data
  } catch (error: any) {
    console.error('查询任务状态失败:', error)
    throw new Error(error.response?.data?.detail || error.message || t('api.queryTaskStatusFailed'))
  }
}

export async function getTripHistory(limit = 8): Promise<TripHistoryItem[]> {
  try {
    const response = await apiClient.get<TripHistoryResponse>('/api/trip/history', {
      params: { limit },
    })
    return Array.isArray(response.data?.items) ? response.data.items : []
  } catch (error: any) {
    console.error('查询历史计划失败:', error)
    throw new Error(error.response?.data?.detail || error.message || t('api.queryTaskStatusFailed'))
  }
}

/**
 * 回看历史计划:从后端 trip_plans 落库记录读取完整行程(与 pollTaskStatus 响应结构兼容)
 */
export async function getTripPlan(planId: string): Promise<any> {
  try {
    // 携带当前语言,后端对缺失 overall_suggestions 的历史计划按语言生成兜底文案
    const lang = i18n.global.locale.value
    const response = await apiClient.get(`/api/trip/plans/${planId}`, { params: { lang } })
    return response.data
  } catch (error: any) {
    console.error('读取历史计划详情失败:', error)
    throw new Error(error.response?.data?.detail || error.message || t('api.queryTaskStatusFailed'))
  }
}

export interface StoryCardResponse {
  success: boolean
  language: 'zh' | 'en' | 'ja'
  fallback: boolean
  title?: string
  body?: string
}

export type StoryCardLanguage = 'zh' | 'en' | 'ja'

/**
 * 生成旅行故事卡文案(AI 生成失败时后端返回 fallback:true,前端需用本地模板兜底)
 */
export async function generateStoryCard(
  planId: string,
  language: StoryCardLanguage
): Promise<StoryCardResponse> {
  try {
    const response = await apiClient.post<StoryCardResponse>('/api/trip/story-card', {
      plan_id: planId,
      language,
    })
    return response.data
  } catch (error: any) {
    console.error('生成旅行故事卡失败:', error)
    throw new Error(error.response?.data?.detail || error.message || t('api.generateStoryCardFailed'))
  }
}

/**
 * 删除落库的历史计划
 */
export async function deleteTripPlan(planId: string): Promise<void> {
  try {
    await apiClient.delete(`/api/trip/history/${planId}`)
  } catch (error: any) {
    console.error('删除历史计划失败:', error)
    throw new Error(error.response?.data?.detail || error.message || t('api.deleteTripPlanFailed'))
  }
}

const resolveTaskWsUrl = (wsUrl: string): string =>
  wsUrl.startsWith('ws://') || wsUrl.startsWith('wss://') ? wsUrl : `${getWsBaseUrl()}${wsUrl}`

function subscribeTripTask(
  wsUrl: string,
  options?: GenerateTripPlanOptions
): Promise<TripPlanResponse> {
  return new Promise((resolve, reject) => {
    let settled = false
    const socket = new WebSocket(wsUrl)

    const safeResolve = (value: TripPlanResponse) => {
      if (settled) return
      settled = true
      socket.close()
      resolve(value)
    }

    const safeReject = (error: unknown) => {
      if (settled) return
      settled = true
      socket.close()
      reject(error)
    }

    socket.onmessage = (ev) => {
      try {
        const event = JSON.parse(ev.data) as TripTaskEvent
        options?.onTaskEvent?.(event)

        if (event.status === 'completed') {
          if (!event.result) {
            safeReject(new Error(t('api.generateTripPlanFailed')))
            return
          }
          safeResolve(event.result)
          return
        }

        if (event.status === 'failed') {
          safeReject(new Error(event.error || event.message || t('api.generateTripPlanFailed')))
        }
      } catch (err) {
        safeReject(err)
      }
    }

    socket.onerror = () => {
      safeReject(new Error(t('api.generateTripPlanFailed')))
    }

    socket.onclose = () => {
      if (!settled) {
        safeReject(new Error(t('api.generateTripPlanFailed')))
      }
    }
  })
}

/**
 * 生成旅行计划
 */
export async function generateTripPlan(
  formData: TripFormData,
  options?: GenerateTripPlanOptions
): Promise<TripPlanResponse> {
  const task = await submitTripPlan(formData)
  options?.onTaskCreated?.(task)
  return subscribeTripTask(resolveTaskWsUrl(task.ws_url), options)
}

/**
 * 重新挂接一个已提交的任务。
 * 页面刷新、WebSocket 断线或代理空闲超时后，后端任务可能仍在继续执行。
 */
export async function resumeTripPlan(
  taskId: string,
  options?: GenerateTripPlanOptions
): Promise<TripPlanResponse> {
  const status = await pollTaskStatus(taskId)

  if (status?.status === 'completed' && status.result) {
    return status.result as TripPlanResponse
  }
  if (status?.status === 'failed') {
    throw new Error(status.error || t('api.generateTripPlanFailed'))
  }

  return subscribeTripTask(resolveTaskWsUrl(`/api/trip/ws/${taskId}`), options)
}

/**
 * 健康检查
 */
export async function healthCheck(): Promise<any> {
  try {
    const response = await apiClient.get('/health')
    return response.data
  } catch (error: any) {
    console.error('健康检查失败:', error)
    throw new Error(error.message || t('api.healthCheckFailed'))
  }
}

// ========== 用户偏好记忆管理 ==========
export interface UserMemoryItem {
  memory_id: string
  content: string
  source: string
  weight: number
  create_time: number
  last_access_time: number
}

export async function listUserMemory(userId: string): Promise<UserMemoryItem[]> {
  const response = await apiClient.get('/api/memory/list', { params: { user_id: userId } })
  return response.data?.data ?? []
}

export async function addExplicitMemory(userId: string, content: string) {
  const response = await apiClient.post('/api/memory/add-explicit', null, {
    params: { user_id: userId, content },
  })
  return response.data
}

export async function clearUserMemory(userId: string) {
  const response = await apiClient.delete('/api/memory/clear', { params: { user_id: userId } })
  return response.data
}

export async function deleteMemoryItem(userId: string, memoryId: string) {
  const response = await apiClient.delete('/api/memory/item', {
    params: { user_id: userId, memory_id: memoryId },
  })
  return response.data
}

export default apiClient

