/** 认证接口封装 */
import { get, post } from './request'
import { applyTokenData } from '@/utils/token'

export interface SessionUser {
  id: number
  username: string
  avatar_url?: string
}

export interface LoginResult {
  token_type: string
  user: SessionUser
  tokens: {
    access_token: string
    refresh_token: string
    expires_in?: number
  }
}

export function register(username: string, password: string) {
  return post<void>('/auth/register', { username, password })
}

export async function login(username: string, password: string): Promise<LoginResult> {
  const data = await post<LoginResult>('/auth/login', { username, password })
  applyTokenData(data.tokens)
  return data
}

export function logout(refresh_token?: string) {
  return post<void>('/auth/logout', { refresh_token })
}

export function me(): Promise<SessionUser> {
  return get<SessionUser>('/auth/me')
}