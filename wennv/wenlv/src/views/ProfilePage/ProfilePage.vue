<script setup lang="ts">
/**
 * ProfilePage.vue — 「我的」个人中心
 * 仅保留：个人信息编辑（含头像上传）+ 已打卡景点（含现场照片上传）
 * 收藏已迁移至独立页 FavoritesPage
 */
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useLanguageStore } from '@/stores/language'
import { useUserStore } from '@/stores/user'
import { scenicSpots, getScenicSpotById } from '@/data/chengdu'
import type { ScenicSpot } from '@/types'

const router = useRouter()
const langStore = useLanguageStore()
const userStore = useUserStore()

const { lang } = storeToRefs(langStore)
const { profile, visits } = storeToRefs(userStore)

/* ── 头像 ── */
const avatarChar = computed(() => profile.value.nickname?.trim().charAt(0) || '蜀')
const avatarInputRef = ref<HTMLInputElement | null>(null)

function triggerAvatarPick() {
  avatarInputRef.value?.click()
}
function onAvatarFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/')) return
  readAsDataUrl(file).then(url => userStore.updateAvatar(url))
  ;(e.target as HTMLInputElement).value = ''
}

/* ── 个人信息编辑 ── */
const editForm = ref({
  nickname: profile.value.nickname,
  email: profile.value.email,
  phone: profile.value.phone,
})
watch(profile, v => {
  editForm.value.nickname = v.nickname
  editForm.value.email = v.email
  editForm.value.phone = v.phone
})
const saved = ref(false)
function saveProfile() {
  userStore.updateProfile({ ...editForm.value, avatar: profile.value.avatar })
  saved.value = true
  window.setTimeout(() => (saved.value = false), 1500)
}

/* ── 已打卡景点 ── */
const visitSpotId = ref('')
const visitedList = computed(() =>
  visits.value
    .map(record => ({ record, spot: getScenicSpotById(record.spotId) }))
    .filter((x): x is { record: typeof visits.value[number]; spot: ScenicSpot } => Boolean(x.spot)),
)

function addVisit() {
  if (!visitSpotId.value) return
  userStore.addVisit(visitSpotId.value)
  visitSpotId.value = ''
}

function spotName(spot: ScenicSpot): string {
  return lang.value === 'zh' ? spot.nameZh : spot.nameEn
}
function formatTime(iso: string): string {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
function goDetail(id: string) {
  router.push({ name: 'scenic-detail', params: { id } })
}

/* ── 照片上传 ── */
function readAsDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const raw = String(reader.result ?? '')
      compressDataUrl(raw).then(resolve).catch(() => resolve(raw))
    }
    reader.onerror = () => reject(new Error('read failed'))
    reader.readAsDataURL(file)
  })
}

function compressDataUrl(dataUrl: string, maxSide = 960, quality = 0.72): Promise<string> {
  return new Promise(resolve => {
    const img = new Image()
    img.onload = () => {
      let { width: w, height: h } = img
      if (w <= maxSide && h <= maxSide) { resolve(dataUrl); return }
      if (w > h) { h = Math.round((h * maxSide) / w); w = maxSide }
      else { w = Math.round((w * maxSide) / h); h = maxSide }
      const canvas = document.createElement('canvas')
      canvas.width = w; canvas.height = h
      const ctx = canvas.getContext('2d')
      if (!ctx) { resolve(dataUrl); return }
      ctx.drawImage(img, 0, 0, w, h)
      resolve(canvas.toDataURL('image/jpeg', quality))
    }
    img.onerror = () => resolve(dataUrl)
    img.src = dataUrl
  })
}

async function onPhotoFiles(e: Event, spotId: string) {
  const input = e.target as HTMLInputElement
  const files = input.files ? Array.from(input.files) : []
  input.value = ''
  if (!files.length) return
  const avail = 9 - (visits.value.find(v => v.spotId === spotId)?.photos.length ?? 0)
  if (avail <= 0) return
  const picked = files.filter(f => f.type.startsWith('image/')).slice(0, avail)
  const urls: string[] = []
  for (const f of picked) {
    try {
      const u = await readAsDataUrl(f)
      urls.push(u)
    } catch { /* ignore */ }
  }
  if (urls.length) userStore.addVisitPhotos(spotId, urls)
}
</script>

<template>
  <div class="profile">
    <!-- 头部：头像 + 身份 + 去收藏 -->
    <section class="profile__header">
      <button class="profile__avatar" :title="langStore.t('profile.avatarUpload')" @click="triggerAvatarPick">
        <img v-if="profile.avatar" :src="profile.avatar" alt="avatar" class="profile__avatar-img" />
        <span v-else class="profile__avatar-char">{{ avatarChar }}</span>
        <span class="profile__avatar-mask">{{ langStore.t('profile.avatarUpload') }}</span>
      </button>
      <input ref="avatarInputRef" type="file" accept="image/jpeg,image/png,image/webp,image/jpg" class="profile__file-hidden" @change="onAvatarFile" />
      <div class="profile__identity">
        <h1 class="profile__name">{{ profile.nickname || langStore.t('profile.title') }}</h1>
        <p class="profile__meta">{{ profile.email || langStore.t('profile.subtitle') }}</p>
        <p class="profile__avatar-hint">{{ langStore.t('profile.avatarDesc') }}</p>
      </div>
      <button class="profile__to-fav" @click="router.push('/home/favorites')">{{ langStore.t('mineMenu.favorites') }} →</button>
    </section>

    <div class="profile__grid">
      <!-- 个人信息 -->
      <section class="profile-card">
        <header class="profile-card__head">
          <h2 class="profile-card__title">{{ langStore.t('profile.infoTitle') }}</h2>
          <p class="profile-card__desc">{{ langStore.t('profile.infoDesc') }}</p>
        </header>
        <form class="profile-form" @submit.prevent="saveProfile">
          <label class="profile-form__field">
            <span class="profile-form__label">{{ langStore.t('profile.nickname') }}</span>
            <input v-model="editForm.nickname" class="profile-form__input" type="text" :placeholder="langStore.t('profile.nicknamePlaceholder')" />
          </label>
          <label class="profile-form__field">
            <span class="profile-form__label">{{ langStore.t('profile.email') }}</span>
            <input v-model="editForm.email" class="profile-form__input" type="email" :placeholder="langStore.t('profile.emailPlaceholder')" />
          </label>
          <label class="profile-form__field">
            <span class="profile-form__label">{{ langStore.t('profile.phone') }}</span>
            <input v-model="editForm.phone" class="profile-form__input" type="tel" :placeholder="langStore.t('profile.phonePlaceholder')" />
          </label>
          <button type="submit" class="profile-form__submit">
            {{ saved ? langStore.t('profile.saveSuccess') : langStore.t('profile.save') }}
          </button>
        </form>
      </section>

      <!-- 已打卡 + 照片上传 -->
      <section class="profile-card profile-card--wide">
        <header class="profile-card__head">
          <h2 class="profile-card__title">{{ langStore.t('profile.visitedTitle') }}</h2>
          <p class="profile-card__desc">{{ langStore.t('profile.visitedDesc') }} · {{ langStore.t('profile.uploadHint') }}</p>
        </header>

        <form class="profile-checkin" @submit.prevent="addVisit">
          <select v-model="visitSpotId" class="profile-form__input">
            <option value="" disabled>{{ langStore.t('profile.visitedSelectPlaceholder') }}</option>
            <option v-for="spot in scenicSpots" :key="spot.id" :value="spot.id">{{ spotName(spot) }}</option>
          </select>
          <button type="submit" class="profile-checkin__btn" :disabled="!visitSpotId">{{ langStore.t('profile.checkin') }}</button>
        </form>

        <div v-if="!visitedList.length" class="profile-empty">{{ langStore.t('profile.visitedEmpty') }}</div>

        <ul v-else class="profile-list profile-list--visits">
          <li v-for="{ record, spot } in visitedList" :key="record.spotId" class="profile-list__item profile-list__item--visit">
            <div class="profile-list__visit-head">
              <button class="profile-list__main" @click="goDetail(spot.id)">
                <span class="profile-list__name">{{ spotName(spot) }}</span>
                <span class="profile-list__time">{{ formatTime(record.visitedAt) }}</span>
              </button>
              <div class="profile-list__visit-actions">
                <label class="profile-list__upload-btn">
                  {{ record.photos.length >= 9 ? langStore.t('profile.photoLimitHit') : langStore.t('profile.uploadPhotos') }}
                  <input type="file" accept="image/jpeg,image/png,image/webp,image/jpg" multiple class="profile__file-hidden" :disabled="record.photos.length >= 9" @change="onPhotoFiles($event, record.spotId)" />
                </label>
                <button class="profile-list__remove" @click="userStore.removeVisit(record.spotId)">{{ langStore.t('profile.removeVisit') }}</button>
              </div>
            </div>
            <!-- 照片墙 -->
            <div v-if="record.photos.length" class="profile-photos">
              <div v-for="(src, idx) in record.photos" :key="idx" class="profile-photos__cell">
                <img :src="src" :alt="spotName(spot) + ' photo ' + (idx+1)" class="profile-photos__img" />
                <button class="profile-photos__del" title="删除" @click="userStore.removeVisitPhoto(record.spotId, idx)">×</button>
              </div>
              <span v-if="record.photos.length < 9" class="profile-photos__count">{{ record.photos.length }}/9</span>
            </div>
            <div v-else class="profile-photos__empty">{{ langStore.t('profile.uploadHint') }}</div>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>

<style scoped>
.profile { max-width: 1080px; margin: 0 auto; display: flex; flex-direction: column; gap: var(--space-8); }
.profile__header {
  display: flex; align-items: center; gap: var(--space-6);
  padding: var(--space-8); background: var(--color-surface);
  border: 1px solid var(--color-border); border-radius: var(--radius-xl);
}
.profile__avatar {
  position: relative; width: 72px; height: 72px; flex-shrink: 0; border-radius: var(--radius-full); overflow: hidden;
  display: flex; align-items: center; justify-content: center;
  background: color-mix(in srgb, var(--color-gold) 15%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-gold) 45%, transparent);
  color: var(--color-gold); font-family: var(--font-display); font-size: var(--text-3xl); font-weight: 700;
}
.profile__avatar-img { width: 100%; height: 100%; object-fit: cover; }
.profile__avatar-char { line-height: 1; }
.profile__avatar-mask {
  position: absolute; inset: 0; display: flex; align-items: center; justify-content: center;
  background: rgba(0,0,0,0.52); color: #fff; font-size: var(--text-xs); opacity: 0; transition: opacity var(--transition-fast); text-align: center; padding: 4px;
}
.profile__avatar:hover .profile__avatar-mask { opacity: 1; }
.profile__file-hidden { display: none; }
.profile__identity { display: flex; flex-direction: column; gap: 4px; flex: 1; min-width: 0; }
.profile__name { font-family: var(--font-display); font-size: var(--text-2xl); font-weight: 700; color: var(--color-text-primary); }
.profile__meta { font-size: var(--text-sm); color: var(--color-text-muted); }
.profile__avatar-hint { font-size: var(--text-xs); color: var(--color-text-muted); }
.profile__to-fav {
  flex-shrink: 0; padding: var(--space-2) var(--space-4); border-radius: var(--radius-full);
  border: 1px solid var(--color-border); font-size: var(--text-sm); color: var(--color-text-secondary); transition: all var(--transition-fast);
}
.profile__to-fav:hover { border-color: var(--color-gold); color: var(--color-gold); }

.profile__grid { display: grid; grid-template-columns: 380px 1fr; gap: var(--space-6); align-items: start; }
.profile-card { background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius-lg); padding: var(--space-6); display: flex; flex-direction: column; gap: var(--space-5); }
.profile-card--wide { min-width: 0; }
.profile-card__head { display: flex; flex-direction: column; gap: var(--space-1); padding-bottom: var(--space-4); border-bottom: 1px solid var(--color-border); }
.profile-card__title { font-family: var(--font-display); font-size: var(--text-lg); font-weight: 600; color: var(--color-gold); }
.profile-card__desc { font-size: var(--text-xs); color: var(--color-text-muted); }
.profile-form { display: flex; flex-direction: column; gap: var(--space-4); }
.profile-form__field { display: flex; flex-direction: column; gap: var(--space-2); }
.profile-form__label { font-size: var(--text-sm); color: var(--color-text-secondary); }
.profile-form__input { padding: var(--space-3) var(--space-4); background: var(--color-bg-alt); border: 1px solid var(--color-border); border-radius: var(--radius-md); color: var(--color-text-primary); font-family: var(--font-body); font-size: var(--text-base); outline: none; transition: border-color var(--transition-fast); }
.profile-form__input:focus { border-color: var(--color-gold); box-shadow: 0 0 0 2px var(--color-gold-glow); }
.profile-form__submit { padding: var(--space-3); background: var(--color-gold); color: var(--color-text-inverse); border-radius: var(--radius-md); font-weight: 500; margin-top: var(--space-2); transition: all var(--transition-fast); }
.profile-form__submit:hover { background: var(--color-gold-light); box-shadow: 0 0 20px var(--color-gold-glow); }

.profile-checkin { display: flex; gap: var(--space-3); }
.profile-checkin .profile-form__input { flex: 1; }
.profile-checkin__btn { padding: var(--space-3) var(--space-5); background: var(--color-gold); color: var(--color-text-inverse); border-radius: var(--radius-md); font-weight: 500; white-space: nowrap; transition: all var(--transition-fast); }
.profile-checkin__btn:disabled { opacity: 0.5; cursor: not-allowed; }
.profile-checkin__btn:not(:disabled):hover { background: var(--color-gold-light); box-shadow: 0 0 20px var(--color-gold-glow); }

.profile-empty { padding: var(--space-6); text-align: center; font-size: var(--text-sm); color: var(--color-text-muted); border: 1px dashed var(--color-border); border-radius: var(--radius-md); }
.profile-list { display: flex; flex-direction: column; gap: var(--space-3); list-style: none; margin: 0; padding: 0; }
.profile-list__item { display: flex; flex-direction: column; gap: var(--space-3); padding: var(--space-4); background: var(--color-bg-alt); border: 1px solid var(--color-border); border-radius: var(--radius-md); }
.profile-list__visit-head { display: flex; align-items: center; gap: var(--space-3); }
.profile-list__main { flex: 1; display: flex; flex-direction: column; align-items: flex-start; gap: 2px; text-align: left; }
.profile-list__name { font-family: var(--font-display); font-size: var(--text-base); font-weight: 600; color: var(--color-text-primary); }
.profile-list__time { font-size: var(--text-xs); color: var(--color-text-muted); }
.profile-list__visit-actions { display: flex; align-items: center; gap: var(--space-2); flex-shrink: 0; }
.profile-list__upload-btn {
  padding: var(--space-2) var(--space-3); border: 1px solid var(--color-gold-dark); color: var(--color-gold-dark);
  border-radius: var(--radius-sm); font-size: var(--text-xs); cursor: pointer; transition: all var(--transition-fast); background: color-mix(in srgb, var(--color-gold) 8%, transparent);
}
.profile-list__upload-btn:hover { background: color-mix(in srgb, var(--color-gold) 16%, transparent); }
.profile-list__remove { padding: var(--space-2) var(--space-3); border: 1px solid var(--color-border); border-radius: var(--radius-sm); font-size: var(--text-xs); color: var(--color-text-secondary); transition: all var(--transition-fast); }
.profile-list__remove:hover { border-color: var(--color-gold); color: var(--color-gold); }

.profile-photos { display: grid; grid-template-columns: repeat(3, 1fr); gap: var(--space-2); }
.profile-photos__cell { position: relative; aspect-ratio: 1; border-radius: var(--radius-md); overflow: hidden; border: 1px solid var(--color-border); background: var(--color-surface); }
.profile-photos__img { width: 100%; height: 100%; object-fit: cover; display: block; }
.profile-photos__del {
  position: absolute; top: 4px; right: 4px; width: 22px; height: 22px; border-radius: var(--radius-full);
  background: rgba(0,0,0,0.62); color: #fff; font-size: 14px; line-height: 1; display: flex; align-items: center; justify-content: center;
}
.profile-photos__count { grid-column: 1/-1; font-size: var(--text-xs); color: var(--color-text-muted); text-align: right; }
.profile-photos__empty { font-size: var(--text-xs); color: var(--color-text-muted); border: 1px dashed var(--color-border); border-radius: var(--radius-md); padding: var(--space-3); text-align: center; }

@media (max-width: 900px) { .profile__grid { grid-template-columns: 1fr; } .profile__header { padding: var(--space-6); } .profile-photos { grid-template-columns: repeat(3, 1fr); } }
@media (max-width: 640px) { .profile-photos { grid-template-columns: repeat(2, 1fr); } .profile-list__visit-head { flex-direction: column; align-items: stretch; } }
</style>
