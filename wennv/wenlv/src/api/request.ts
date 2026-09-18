/** 统一的 axios 实例与响应处理
 *  - 请求拦截：注入 Authorization Bearer access_token
 *  - 响应拦截：解包 {code,message,data}，code!==0 抛错；HTTP 401 走刷新重放
 */
import axios, {
  type AxiosRequestConfig,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from 'axios'
import { getAccessToken, setTokens, clearTokens, refreshTokens } from '@/utils/token'

// 后端统一响应结构
export interface ApiResult<T = unknown> {
  code: number
  message: string
  data: T
}

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 20000,
})

// 请求拦截：附加 Bearer token
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getAccessToken()
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`)
  }
  return config
})

// 响应拦截：解包并处理 401 刷新
let refreshing = false
let pendingQueue: Array<() => void> = []

function flushQueue() {
  pendingQueue.forEach(cb => cb())
  pendingQueue = []
}

api.interceptors.response.use(
  (response: AxiosResponse) => {
    const body = response.data as ApiResult
    if (body && typeof body.code === 'number' && body.code !== 0) {
      const err = new Error(body.message || '请求失败') as Error & { code?: number }
      err.code = body.code
      throw err
    }
    return response.data
  },
  async (error) => {
    const status = error?.response?.status
    const original = error?.config as (AxiosRequestConfig & { _retried?: boolean }) | undefined

    if (status === 401 && original && !original._retried && !isAuthUrl(original.url)) {
      if (refreshing) {
        // 已有刷新在进行，排队等待后重放
        return new Promise(resolve => {
          pendingQueue.push(() => resolve(api(original)))
        })
      }

      original._retried = true
      refreshing = true
      try {
        const ok = await refreshTokens()
        flushQueue()
        if (ok) return api(original)
      } catch {
        pendingQueue = []
      } finally {
        refreshing = false
      }
    }

    throw error
  },
)

function isAuthUrl(url?: string): boolean {
  if (!url) return false
  return url.includes('/auth/login') || url.includes('/auth/register') || url.includes('/auth/refresh')
}

/** 发送 GET 请求，返回 data 字段 */
export function get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return api.get<T, ApiResult<T>>(url, config).then(r => r.data)
}

/** 发送 POST 请求，返回 data 字段 */
export function post<T = unknown>(url: string, body?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return api.post<T, ApiResult<T>>(url, body, config).then(r => r.data)
}

/** 发送 PUT 请求，返回 data 字段 */
export function put<T = unknown>(url: string, body?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return api.put<T, ApiResult<T>>(url, body, config).then(r => r.data)
}

/** 发送 DELETE 请求，返回 data 字段 */
export function del<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return api.delete<T, ApiResult<T>>(url, config).then(r => r.data)
}

export default api