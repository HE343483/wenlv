<script setup lang="ts">
/**
 * ScenicCard.vue — 景点卡片
 * 设计特征：深色底面 + 蜀锦金边框 + 图片占位
 * 包含：图片区、标题、描述、标签、评分
 * 状态：正常、加载中
 */
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useLanguageStore } from '@/stores/language'
import type { ScenicSpot } from '@/types'

const props = defineProps<{
  spot: ScenicSpot
  loading?: boolean
}>()

const langStore = useLanguageStore()
const { lang } = storeToRefs(langStore)

const name = computed(() => lang.value === 'zh' ? props.spot.nameZh : props.spot.nameEn)
const shortDesc = computed(() => lang.value === 'zh' ? props.spot.shortDescZh : props.spot.shortDescEn)

function ratingStars(rating: number): string {
  const full = Math.floor(rating)
  const half = rating % 1 >= 0.5 ? 1 : 0
  return '★'.repeat(full) + (half ? '☆' : '') + '☆'.repeat(5 - full - half)
}
</script>

<template>
  <article
    class="scenic-card"
    :class="{ 'scenic-card--loading': loading }"
  >
    <!-- 图片占位区 -->
    <div class="scenic-card__image">
      <div class="scenic-card__placeholder">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
          <rect x="3" y="3" width="18" height="18" rx="2"/>
          <circle cx="8.5" cy="8.5" r="1.5"/>
          <path d="M21 15l-5-5L5 21"/>
        </svg>
        <span class="scenic-card__placeholder-text">{{ spot.nameZh }}</span>
        <span class="scenic-card__api-badge">{{ langStore.t('scenic.dataFromApi') }}</span>
      </div>
    </div>

    <!-- 内容 -->
    <div class="scenic-card__body">
      <h3 class="scenic-card__title">{{ name }}</h3>
      <p class="scenic-card__desc">{{ shortDesc }}</p>

      <!-- 标签 -->
      <div class="scenic-card__tags" v-if="spot.tags.length">
        <span v-for="tag in spot.tags" :key="tag" class="scenic-card__tag">{{ tag }}</span>
      </div>

      <!-- 底部：评分 + 操作 -->
      <div class="scenic-card__footer">
        <div class="scenic-card__rating">
          <span class="scenic-card__stars">{{ ratingStars(spot.rating) }}</span>
          <span class="scenic-card__rating-num">{{ spot.rating }}</span>
        </div>
        <button class="scenic-card__action">
          {{ langStore.t('scenic.viewDetail') }}
          <span class="scenic-card__arrow">→</span>
        </button>
      </div>
    </div>
  </article>
</template>

<style scoped>
.scenic-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all var(--transition-base);
  display: flex;
  flex-direction: column;
}

.scenic-card:hover {
  border-color: var(--color-gold-dark);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px var(--color-gold-glow), var(--shadow-lg);
}

/* 图片占位 */
.scenic-card__image {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  background: var(--color-bg-alt);
}

.scenic-card__placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  color: var(--color-text-muted);
}

.scenic-card__placeholder-text {
  font-family: var(--font-display);
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  opacity: 0.6;
}

.scenic-card__api-badge {
  font-size: var(--text-xs);
  padding: 2px 8px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* 内容 */
.scenic-card__body {
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  flex: 1;
}

.scenic-card__title {
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
}

.scenic-card__desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: var(--leading-relaxed);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 标签 */
.scenic-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.scenic-card__tag {
  font-size: var(--text-xs);
  padding: 2px 10px;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

/* 底部 */
.scenic-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-border);
}

.scenic-card__rating {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.scenic-card__stars {
  color: var(--color-gold);
  font-size: var(--text-sm);
  letter-spacing: 2px;
}

.scenic-card__rating-num {
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
}

.scenic-card__action {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  gap: var(--space-1);
  transition: color var(--transition-fast);
}

.scenic-card__action:hover {
  color: var(--color-gold);
}

.scenic-card__arrow {
  transition: transform var(--transition-fast);
}

.scenic-card__action:hover .scenic-card__arrow {
  transform: translateX(3px);
}

/* Loading state */
.scenic-card--loading {
  pointer-events: none;
}

.scenic-card--loading .scenic-card__image {
  background: linear-gradient(110deg, var(--color-surface) 30%, var(--color-surface-hover) 50%, var(--color-surface) 70%);
  background-size: 200% 100%;
  animation: shimmer 1.5s ease-in-out infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* Mobile */
@media (max-width: 640px) {
  .scenic-card__body {
    padding: var(--space-4);
  }
  .scenic-card__title {
    font-size: var(--text-base);
  }
}
</style>
