<script setup lang="ts">
/**
 * FavoritesPage.vue — 收藏独立页
 * 网格卡片展示已收藏景点，支持取消收藏与跳详情
 */
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useLanguageStore } from '@/stores/language'
import { useUserStore } from '@/stores/user'
import { getScenicSpotById } from '@/data/chengdu'
import AppIcon from '@/components/AppIcon.vue'
import HomeBanner from '@/components/HomeBanner.vue'
import type { ScenicSpot } from '@/types'

const router = useRouter()
const langStore = useLanguageStore()
const userStore = useUserStore()
const { lang } = storeToRefs(langStore)
const { favorites } = storeToRefs(userStore)

const items = computed<ScenicSpot[]>(() =>
  favorites.value.map(id => getScenicSpotById(id)).filter((s): s is ScenicSpot => Boolean(s)),
)

function spotName(s: ScenicSpot): string {
  return lang.value === 'zh' ? s.nameZh : s.nameEn
}
function spotDesc(s: ScenicSpot): string {
  return lang.value === 'zh' ? s.shortDescZh : s.shortDescEn
}
function goDetail(id: string) {
  router.push({ name: 'scenic-detail', params: { id } })
}
function remove(id: string) {
  userStore.removeFavorite(id)
}
</script>

<template>
  <div class="favorites-page">
    <HomeBanner
      :eyebrow="langStore.lang === 'zh' ? '蜀韵·成都' : 'Shu·Chengdu'"
      :title="langStore.t('favorites.title')"
      :subtitle="langStore.t('favorites.subtitle')"
      watermark="藏"
    />

    <div v-if="items.length" class="favorites-page__count">
      {{ items.length }} {{ langStore.t('favorites.countLabel') }}
    </div>

    <!-- 空状态 -->
    <div v-if="!items.length" class="favorites-empty">
      <div class="favorites-empty__icon" aria-hidden="true">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2">
          <path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z" />
        </svg>
      </div>
      <h2 class="favorites-empty__title">{{ langStore.t('favorites.empty') }}</h2>
      <p class="favorites-empty__hint">{{ langStore.t('favorites.emptyHint') }}</p>
      <button class="favorites-empty__btn" @click="router.push('/home/explore')">{{ langStore.t('favorites.goExplore') }}</button>
    </div>

    <!-- 网格 -->
    <div v-else class="favorites-grid">
      <article v-for="spot in items" :key="spot.id" class="fav-card" tabindex="0" @click="goDetail(spot.id)" @keyup.enter="goDetail(spot.id)">
        <div class="fav-card__media" aria-hidden="true">
          <span class="fav-card__char">{{ spot.nameZh.charAt(0) }}</span>
          <span class="fav-card__rating"><AppIcon name="star" :size="11" filled /> {{ spot.rating }}</span>
          <button class="fav-card__fav-btn" :title="langStore.t('favorites.unfavorite')" @click.stop="remove(spot.id)">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.5"><path d="M12 21C12 21 3 15.5 3 9.5C3 6.5 5 4.5 8 4.5C10 4.5 11.5 5.8 12 7C12.5 5.8 14 4.5 16 4.5C19 4.5 21 6.5 21 9.5C21 15.5 12 21 12 21Z" /></svg>
          </button>
        </div>
        <div class="fav-card__body">
          <h3 class="fav-card__title">{{ spotName(spot) }}</h3>
          <p class="fav-card__desc">{{ spotDesc(spot) }}</p>
          <div class="fav-card__tags">
            <span v-for="tag in spot.tags.slice(0, 3)" :key="tag" class="fav-card__tag">{{ tag }}</span>
          </div>
          <div class="fav-card__footer">
            <button class="fav-card__link" @click.stop="goDetail(spot.id)">{{ langStore.t('favorites.viewDetail') }} →</button>
          </div>
        </div>
      </article>
    </div>
  </div>
</template>

<style scoped>
.favorites-page {
  max-width: var(--max-width);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}
.favorites-page__count {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}
.favorites-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-12) var(--space-6);
  background: var(--color-surface);
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-xl);
  text-align: center;
}
.favorites-empty__icon { color: var(--color-gold); opacity: 0.7; }
.favorites-empty__title { font-family: var(--font-display); font-size: var(--text-xl); color: var(--color-text-primary); }
.favorites-empty__hint { font-size: var(--text-sm); color: var(--color-text-muted); }
.favorites-empty__btn {
  margin-top: var(--space-2);
  padding: var(--space-3) var(--space-6);
  background: var(--color-gold);
  color: var(--color-text-inverse);
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
  transition: all var(--transition-fast);
}
.favorites-empty__btn:hover { background: var(--color-gold-light); box-shadow: var(--shadow-gold); }

.favorites-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}
.fav-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  cursor: pointer;
  transition: all var(--transition-base);
  outline: none;
}
.fav-card:hover, .fav-card:focus-visible { border-color: var(--color-gold-dark); transform: translateY(-3px); box-shadow: 0 8px 24px var(--color-gold-glow), var(--shadow-lg); }
.fav-card__media {
  position: relative;
  aspect-ratio: 16 / 10;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(ellipse 60% 60% at 50% 30%, color-mix(in srgb, var(--color-gold) 10%, transparent), transparent 70%), var(--color-bg-alt);
}
.fav-card__char { font-family: var(--font-display); font-size: 72px; font-weight: 900; color: transparent; -webkit-text-stroke: 1px color-mix(in srgb, var(--color-gold) 40%, transparent); user-select: none; }
.fav-card__rating {
  position: absolute; left: var(--space-3); top: var(--space-3);
  font-size: var(--text-xs); color: var(--color-gold);
  background: color-mix(in srgb, var(--color-bg) 70%, transparent);
  backdrop-filter: blur(6px); border: 1px solid color-mix(in srgb, var(--color-gold) 35%, transparent);
  border-radius: var(--radius-full); padding: 2px 10px;
  display: inline-flex; align-items: center; gap: 4px;
}
.fav-card__fav-btn {
  position: absolute; right: var(--space-3); top: var(--space-3);
  width: 34px; height: 34px; display: flex; align-items: center; justify-content: center;
  border-radius: var(--radius-full); color: var(--color-gold);
  background: color-mix(in srgb, var(--color-bg) 70%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-gold) 40%, transparent);
  backdrop-filter: blur(6px); transition: all var(--transition-fast);
}
.fav-card__fav-btn:hover { background: color-mix(in srgb, var(--color-gold) 12%, transparent); }
.fav-card__body { padding: var(--space-5); display: flex; flex-direction: column; gap: var(--space-3); flex: 1; }
.fav-card__title { font-family: var(--font-display); font-size: var(--text-lg); font-weight: 600; color: var(--color-text-primary); }
.fav-card__desc { font-size: var(--text-sm); color: var(--color-text-secondary); line-height: var(--leading-relaxed); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.fav-card__tags { display: flex; flex-wrap: wrap; gap: var(--space-2); }
.fav-card__tag { font-size: var(--text-xs); padding: 2px 10px; border: 1px solid var(--color-border); border-radius: var(--radius-full); color: var(--color-text-muted); }
.fav-card__footer { margin-top: auto; padding-top: var(--space-3); border-top: 1px solid var(--color-border); }
.fav-card__link { font-size: var(--text-sm); color: var(--color-text-muted); transition: color var(--transition-fast); }
.fav-card__link:hover { color: var(--color-gold); }

@media (max-width: 1024px) { .favorites-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 640px) { .favorites-grid { grid-template-columns: 1fr; } }
</style>
