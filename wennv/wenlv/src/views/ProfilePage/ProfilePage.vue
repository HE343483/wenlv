<script setup lang="ts">
/**
 * ProfilePage.vue — 「我的」页面
 * 三大功能：查看收藏 / 更改个人信息 / 上传已打卡景点
 */
import { computed, ref } from 'vue'
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
const { profile, favorites, visits } = storeToRefs(userStore)

/* ── 头像首字母 ── */
const avatarChar = computed(() => profile.value.nickname?.trim().charAt(0) || '蜀')

/* ── 个人信息编辑 ── */
const editForm = ref({
  nickname: profile.value.nickname,
  email: profile.value.email,
  phone: profile.value.phone,
})
const saved = ref(false)

function saveProfile() {
  userStore.updateProfile({ ...editForm.value })
  saved.value = true
  window.setTimeout(() => (saved.value = false), 1500)
}

/* ── 我的收藏 ── */
const favoriteSpots = computed<ScenicSpot[]>(() =>
  favorites.value
    .map(id => getScenicSpotById(id))
    .filter((s): s is ScenicSpot => Boolean(s)),
)

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
</script>

<template>
  <div class="profile">
    <!-- ── 头部：头像 + 身份信息 ── -->
    <section class="profile__header">
      <div class="profile__avatar">{{ avatarChar }}</div>
      <div class="profile__identity">
        <h1 class="profile__name">{{ profile.nickname || langStore.t('profile.title') }}</h1>
        <p class="profile__meta">{{ profile.email || langStore.t('profile.subtitle') }}</p>
      </div>
    </section>

    <div class="profile__grid">
      <!-- ── 更改个人信息 ── -->
      <section class="profile-card">
        <header class="profile-card__head">
          <h2 class="profile-card__title">{{ langStore.t('profile.infoTitle') }}</h2>
          <p class="profile-card__desc">{{ langStore.t('profile.infoDesc') }}</p>
        </header>

        <form class="profile-form" @submit.prevent="saveProfile">
          <label class="profile-form__field">
            <span class="profile-form__label">{{ langStore.t('profile.nickname') }}</span>
            <input
              v-model="editForm.nickname"
              class="profile-form__input"
              type="text"
              :placeholder="langStore.t('profile.nicknamePlaceholder')"
            />
          </label>

          <label class="profile-form__field">
            <span class="profile-form__label">{{ langStore.t('profile.email') }}</span>
            <input
              v-model="editForm.email"
              class="profile-form__input"
              type="email"
              :placeholder="langStore.t('profile.emailPlaceholder')"
            />
          </label>

          <label class="profile-form__field">
            <span class="profile-form__label">{{ langStore.t('profile.phone') }}</span>
            <input
              v-model="editForm.phone"
              class="profile-form__input"
              type="tel"
              :placeholder="langStore.t('profile.phonePlaceholder')"
            />
          </label>

          <button type="submit" class="profile-form__submit">
            {{ saved ? langStore.t('profile.saveSuccess') : langStore.t('profile.save') }}
          </button>
        </form>
      </section>

      <!-- ── 查看收藏 ── -->
      <section class="profile-card">
        <header class="profile-card__head">
          <h2 class="profile-card__title">{{ langStore.t('profile.favoritesTitle') }}</h2>
          <p class="profile-card__desc">{{ langStore.t('profile.favoritesDesc') }}</p>
        </header>

        <div v-if="!favoriteSpots.length" class="profile-empty">
          {{ langStore.t('profile.favoritesEmpty') }}
        </div>

        <ul v-else class="profile-list">
          <li v-for="spot in favoriteSpots" :key="spot.id" class="profile-list__item">
            <button class="profile-list__main" @click="goDetail(spot.id)">
              <span class="profile-list__name">{{ spotName(spot) }}</span>
              <span class="profile-list__hint">{{ langStore.t('profile.viewDetail') }} →</span>
            </button>
            <button class="profile-list__remove" @click="userStore.removeFavorite(spot.id)">
              {{ langStore.t('profile.remove') }}
            </button>
          </li>
        </ul>
      </section>

      <!-- ── 上传已打卡景点 ── -->
      <section class="profile-card">
        <header class="profile-card__head">
          <h2 class="profile-card__title">{{ langStore.t('profile.visitedTitle') }}</h2>
          <p class="profile-card__desc">{{ langStore.t('profile.visitedDesc') }}</p>
        </header>

        <form class="profile-checkin" @submit.prevent="addVisit">
          <select v-model="visitSpotId" class="profile-form__input">
            <option value="" disabled>{{ langStore.t('profile.visitedSelectPlaceholder') }}</option>
            <option v-for="spot in scenicSpots" :key="spot.id" :value="spot.id">
              {{ spotName(spot) }}
            </option>
          </select>
          <button type="submit" class="profile-checkin__btn" :disabled="!visitSpotId">
            {{ langStore.t('profile.checkin') }}
          </button>
        </form>

        <div v-if="!visitedList.length" class="profile-empty">
          {{ langStore.t('profile.visitedEmpty') }}
        </div>

        <ul v-else class="profile-list">
          <li v-for="{ record, spot } in visitedList" :key="record.spotId" class="profile-list__item">
            <button class="profile-list__main" @click="goDetail(spot.id)">
              <span class="profile-list__name">{{ spotName(spot) }}</span>
              <span class="profile-list__time">{{ formatTime(record.visitedAt) }}</span>
            </button>
            <button class="profile-list__remove" @click="userStore.removeVisit(record.spotId)">
              {{ langStore.t('profile.removeVisit') }}
            </button>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>

<style scoped>
.profile {
  max-width: 1080px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
}

/* 头部 */
.profile__header {
  display: flex;
  align-items: center;
  gap: var(--space-6);
  padding: var(--space-8);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
}

.profile__avatar {
  width: 72px;
  height: 72px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-gold) 15%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-gold) 45%, transparent);
  color: var(--color-gold);
  font-family: var(--font-display);
  font-size: var(--text-3xl);
  font-weight: 700;
}

.profile__identity {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.profile__name {
  font-family: var(--font-display);
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.profile__meta {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* 卡片网格 */
.profile__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-6);
  align-items: start;
}

.profile-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.profile-card__head {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border);
}

.profile-card__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-gold);
  letter-spacing: var(--tracking-wide);
}

.profile-card__desc {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* 表单 */
.profile-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.profile-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.profile-form__label {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
}

.profile-form__input {
  padding: var(--space-3) var(--space-4);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: var(--text-base);
  transition: border-color var(--transition-fast);
  outline: none;
}

.profile-form__input:focus {
  border-color: var(--color-gold);
  box-shadow: 0 0 0 2px var(--color-gold-glow);
}

.profile-form__input::placeholder {
  color: var(--color-text-muted);
}

.profile-form__submit {
  padding: var(--space-3);
  background: var(--color-gold);
  color: var(--color-text-inverse);
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
  margin-top: var(--space-2);
}

.profile-form__submit:hover {
  background: var(--color-gold-light);
  box-shadow: 0 0 20px var(--color-gold-glow);
}

/* 打卡 */
.profile-checkin {
  display: flex;
  gap: var(--space-3);
}

.profile-checkin .profile-form__input {
  flex: 1;
}

.profile-checkin__btn {
  padding: var(--space-3) var(--space-5);
  background: var(--color-gold);
  color: var(--color-text-inverse);
  border-radius: var(--radius-md);
  font-size: var(--text-base);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.profile-checkin__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.profile-checkin__btn:not(:disabled):hover {
  background: var(--color-gold-light);
  box-shadow: 0 0 20px var(--color-gold-glow);
}

/* 空状态 */
.profile-empty {
  padding: var(--space-6);
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-md);
}

/* 列表 */
.profile-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  list-style: none;
  margin: 0;
  padding: 0;
}

.profile-list__item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: border-color var(--transition-fast);
}

.profile-list__item:hover {
  border-color: var(--color-gold-dark);
}

.profile-list__main {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  text-align: left;
}

.profile-list__name {
  font-family: var(--font-display);
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.profile-list__hint {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
}

.profile-list__time {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
}

.profile-list__remove {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.profile-list__remove:hover {
  border-color: var(--color-gold);
  color: var(--color-gold);
}

/* Responsive */
@media (max-width: 768px) {
  .profile__grid {
    grid-template-columns: 1fr;
  }
  .profile__header {
    padding: var(--space-6);
  }
}
</style>
