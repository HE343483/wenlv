/** Token 工具 — access_token / refresh_token 的存取与刷新
 *  - 存取 Access/Refresh token（localStorage 持久化）
 *  - 401 时调用 /auth/refresh 换新token，成功后覆盖存储
 */
import { post } from '@/api/request'

const ACCESS_KEY = 'shuyun-wenlv-access-token'
const REFRESH_KEY = 'shuyun-wenlv-refresh-token'

export function getAccessToken(): string {
  return localStorage.getItem(ACCESS_KEY) || ''
}

export function getRefreshToken(): string {
  return localStorage.getItem(REFRESH_KEY) || ''
}

export function setTokens(access: string, refresh: string) {
  localStorage.setItem(ACCESS_KEY, access)
  localStorage.setItem(REFRESH_KEY, refresh)
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_KEY)
  localStorage.removeItem(REFRESH_KEY)
}

export function hasToken(): boolean {
  return !!getAccessToken()
}

/** 刷新逻辑：login 的 tokens 结构 vs refresh 的顶层结构在此统一处理 */
export function applyTokenData(tokenData: {
  access_token: string
  refresh_token: string
  expires_in?: number
}) {
  setTokens(tokenData.access_token, tokenData.refresh_token)
}

interface RefreshResp {
  access_token: string
  refresh_token: string
  expires_in?: number
}

/** 用 refresh_token 换新 token；失败返回 false */
export async function refreshTokens(): Promise<boolean> {
  const refresh = getRefreshToken()
  if (!refresh) {
    clearTokens()
    return false
  }
  try {
    const data = await post<RefreshResp>('/auth/refresh', { refresh_token: refresh })
    if (data && data.access_token) {
      setTokens(data.access_token, data.refresh_token)
      return true
    }
    return false
  } catch {
    return false
  }
}