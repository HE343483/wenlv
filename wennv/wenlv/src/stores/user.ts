/** 用户 Store — 个人信息 / 收藏 / 打卡景点（含照片）- localStorage 持久化 */

import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { listFavorites, addFavorite, removeFavorite as apiRemoveFavorite } from '@/api/favorite'
import { hasToken } from '@/utils/token'

export interface UserProfile {
  nickname: string
  email: string
  phone: string
  avatar: string
}

/** 已打卡景点记录 */
export interface VisitRecord {
  spotId: string
  /** 打卡时间 ISO */
  visitedAt: string
  /** 现场照片 dataURL 列表，最多 9 张 */
  photos: string[]
}

interface PersistedState {
  profile: UserProfile
  favorites: string[]
  visits: VisitRecord[]
}

const STORAGE_KEY = 'shuyun-chengdu-user'

const DEFAULT_STATE: PersistedState = {
  profile: { nickname: '', email: '', phone: '', avatar: '' },
  favorites: [],
  visits: [],
}

function loadState(): PersistedState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw) as Partial<PersistedState>
      const rawVisits = (parsed.visits ?? []) as VisitRecord[]
      return {
        profile: { ...DEFAULT_STATE.profile, ...(parsed.profile ?? {}) } as UserProfile,
        favorites: parsed.favorites ?? [],
        visits: rawVisits.map(v => ({
          spotId: String((v as VisitRecord).spotId ?? ''),
          visitedAt: String((v as VisitRecord).visitedAt ?? new Date().toISOString()),
          photos: Array.isArray((v as VisitRecord).photos) ? ((v as VisitRecord).photos as string[]) : [],
        })),
      }
    }
  } catch {
    /* 忽略解析失败，回退默认值 */
  }
  return { ...DEFAULT_STATE, profile: { ...DEFAULT_STATE.profile }, favorites: [], visits: [] }
}

function persist(state: PersistedState) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  } catch {
    /* 忽略写入失败（如隐私模式） */
  }
}

export const useUserStore = defineStore('user', () => {
  const initial = loadState()

  const profile = ref<UserProfile>(initial.profile)
  const favorites = ref<string[]>(initial.favorites)
  const visits = ref<VisitRecord[]>(initial.visits)

  watch(
    [profile, favorites, visits],
    () => {
      persist({
        profile: profile.value,
        favorites: favorites.value,
        visits: visits.value,
      })
    },
    { deep: true },
  )

  function updateProfile(next: Partial<UserProfile> & { nickname: string; email: string; phone: string }) {
    profile.value = { ...profile.value, ...next }
  }

  function updateAvatar(dataUrl: string) {
    profile.value.avatar = dataUrl
  }

  function isFavorite(spotId: string): boolean {
    return favorites.value.includes(spotId)
  }

  /** 后端景点ID（scenic-{数字}）→ 数字ID；非该格式返回 null */
  function scenicNumericId(spotId: string): number | null {
    const m = /^scenic-(\d+)$/.exec(spotId)
    return m ? Number(m[1]) : null
  }

  /** 登录后从后端同步收藏列表（target_type=scenic），覆盖本地缓存 */
  async function syncFavorites() {
    if (!hasToken()) return
    try {
      const items = await listFavorites('scenic')
      favorites.value = items.map(i => `scenic-${i.target_id}`)
    } catch {
      /* 同步失败时保留本地缓存 */
    }
  }

  /** 切换收藏：乐观更新本地，再调后端接口；失败回滚 */
  async function toggleFavorite(spotId: string) {
    const exists = favorites.value.includes(spotId)
    const nid = scenicNumericId(spotId)
    if (!hasToken() || nid === null) {
      // 未登录或非后端景点ID：保持本地行为
      if (exists) favorites.value = favorites.value.filter(id => id !== spotId)
      else favorites.value.push(spotId)
      return
    }
    try {
      if (exists) {
        favorites.value = favorites.value.filter(id => id !== spotId)
        await apiRemoveFavorite('scenic', nid)
      } else {
        favorites.value.push(spotId)
        await addFavorite('scenic', nid)
      }
    } catch (err) {
      // 接口失败回滚本地状态
      if (exists) favorites.value.push(spotId)
      else favorites.value = favorites.value.filter(id => id !== spotId)
      throw err
    }
  }

  /** 取消收藏：乐观更新本地，再调后端接口；失败回滚 */
  async function removeFavorite(spotId: string) {
    const exists = favorites.value.includes(spotId)
    const nid = scenicNumericId(spotId)
    if (!hasToken() || nid === null) {
      favorites.value = favorites.value.filter(id => id !== spotId)
      return
    }
    try {
      favorites.value = favorites.value.filter(id => id !== spotId)
      await apiRemoveFavorite('scenic', nid)
    } catch (err) {
      if (exists) favorites.value.push(spotId)
      throw err
    }
  }

  function isVisited(spotId: string): boolean {
    return visits.value.some(v => v.spotId === spotId)
  }

  function addVisit(spotId: string) {
    if (isVisited(spotId)) return
    visits.value.push({ spotId, visitedAt: new Date().toISOString(), photos: [] })
  }

  function removeVisit(spotId: string) {
    visits.value = visits.value.filter(v => v.spotId !== spotId)
  }

  function addVisitPhotos(spotId: string, dataUrls: string[]) {
    const rec = visits.value.find(v => v.spotId === spotId)
    if (!rec) return
    const remain = 9 - rec.photos.length
    if (remain <= 0) return
    rec.photos.push(...dataUrls.slice(0, remain))
  }

  function removeVisitPhoto(spotId: string, index: number) {
    const rec = visits.value.find(v => v.spotId === spotId)
    if (!rec) return
    rec.photos.splice(index, 1)
  }

  /** 登出时重置所有用户态（清除本地持久化） */
  function resetAll() {
    profile.value = { ...DEFAULT_STATE.profile }
    favorites.value = []
    visits.value = []
  }

  return {
    profile,
    favorites,
    visits,
    updateProfile,
    updateAvatar,
    isFavorite,
    syncFavorites,
    toggleFavorite,
    removeFavorite,
    isVisited,
    addVisit,
    removeVisit,
    addVisitPhotos,
    removeVisitPhoto,
    resetAll,
  }
})
