/** 用户 Store — 个人信息 / 收藏 / 打卡景点（含照片）- localStorage 持久化 */

import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

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

  function toggleFavorite(spotId: string) {
    const idx = favorites.value.indexOf(spotId)
    if (idx >= 0) favorites.value.splice(idx, 1)
    else favorites.value.push(spotId)
  }

  function removeFavorite(spotId: string) {
    favorites.value = favorites.value.filter(id => id !== spotId)
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

  return {
    profile,
    favorites,
    visits,
    updateProfile,
    updateAvatar,
    isFavorite,
    toggleFavorite,
    removeFavorite,
    isVisited,
    addVisit,
    removeVisit,
    addVisitPhotos,
    removeVisitPhoto,
  }
})
