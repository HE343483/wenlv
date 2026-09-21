/** 用户 Store — 个人信息 / 收藏（景点·美食·路线）/ 打卡景点（含照片）- localStorage 持久化 */

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

/** 收藏目标类型（与后端 target_type 一致）：景点 / 美食 / 路线 */
export type FavoriteType = 'scenic' | 'food' | 'route'

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

  /** 是否已收藏（收藏ID形如 scenic-12 / food-3 / route-1） */
  function isFavorite(id: string): boolean {
    return favorites.value.includes(id)
  }

  /** 后端收藏ID（scenic-12 / food-3 / route-1）→ 类型与数字ID；非该格式返回 null */
  function parseFavoriteId(id: string): { type: FavoriteType; numericId: number } | null {
    const m = /^(scenic|food|route)-(\d+)$/.exec(id)
    return m ? { type: m[1] as FavoriteType, numericId: Number(m[2]) } : null
  }

  /** 登录后从后端同步全部收藏（景点/美食/路线），覆盖本地缓存 */
  async function syncFavorites() {
    if (!hasToken()) return
    try {
      const items = await listFavorites()
      favorites.value = items.map(i => `${i.target_type}-${i.target_id}`)
    } catch {
      /* 同步失败时保留本地缓存 */
    }
  }

  /** 切换收藏：乐观更新本地，再调后端接口；失败回滚 */
  async function toggleFavorite(id: string) {
    const exists = favorites.value.includes(id)
    const target = parseFavoriteId(id)
    if (!hasToken() || !target) {
      // 未登录或非后端数字ID：保持本地行为
      if (exists) favorites.value = favorites.value.filter(x => x !== id)
      else favorites.value.push(id)
      return
    }
    try {
      if (exists) {
        favorites.value = favorites.value.filter(x => x !== id)
        await apiRemoveFavorite(target.type, target.numericId)
      } else {
        favorites.value.push(id)
        await addFavorite(target.type, target.numericId)
      }
    } catch (err) {
      // 接口失败回滚本地状态
      if (exists) favorites.value.push(id)
      else favorites.value = favorites.value.filter(x => x !== id)
      throw err
    }
  }

  /** 取消收藏：乐观更新本地，再调后端接口；失败回滚 */
  async function removeFavorite(id: string) {
    const exists = favorites.value.includes(id)
    const target = parseFavoriteId(id)
    if (!hasToken() || !target) {
      favorites.value = favorites.value.filter(x => x !== id)
      return
    }
    try {
      favorites.value = favorites.value.filter(x => x !== id)
      await apiRemoveFavorite(target.type, target.numericId)
    } catch (err) {
      if (exists) favorites.value.push(id)
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
