<script setup lang="ts">
/**
 * ScenicGrid.vue — 景点卡片网格
 * 响应式网格布局 (3列 → 2列 → 1列)
 * 根据选择的区域动态过滤景点
 * 包含加载状态和空状态处理
 */
import { computed, ref, watch } from 'vue'
import { useLanguageStore } from '@/stores/language'
import { getScenicSpotsByDistrict } from '@/data/chengdu'
import ScenicCard from './ScenicCard.vue'
import type { ScenicSpot } from '@/types'

const props = defineProps<{
  districtId: string
}>()

const langStore = useLanguageStore()
const loading = ref(false)
const spots = ref<ScenicSpot[]>([])

// 模拟API加载
async function loadSpots(districtId: string) {
  loading.value = true
  // 模拟网络延迟 — 对接后端API时替换为真实请求
  // const response = await fetch(`/api/scenic-spots?district=${districtId}`)
  // const data = await response.json()
  await new Promise(resolve => setTimeout(resolve, 400))
  spots.value = getScenicSpotsByDistrict(districtId)
  loading.value = false
}

watch(() => props.districtId, loadSpots, { immediate: true })
</script>

<template>
  <section class="scenic-grid">
    <!-- 标题 -->
    <div class="scenic-grid__header">
      <h2 class="section-title">{{ langStore.t('scenic.title') }}</h2>
      <p class="section-subtitle">{{ langStore.t('scenic.subtitle') }}</p>
    </div>

    <!-- 加载态 -->
    <div v-if="loading" class="scenic-grid__loading">
      <div class="scenic-grid__shimmer">
        <ScenicCard v-for="n in 6" :key="n" :spot="{ id: 'loading', districtId: '', nameZh: '', nameEn: '', shortDescZh: '', shortDescEn: '', descriptionZh: '', descriptionEn: '', tags: [], imageUrl: '', rating: 0, coords: { lng: 0, lat: 0 } }" :loading="true" />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else-if="!spots.length" class="scenic-grid__empty">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" style="opacity: 0.3">
        <path d="M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0z"/>
        <path d="M9 9h.01M15 9h.01M7 15h10"/>
      </svg>
      <span>{{ langStore.t('scenic.empty') }}</span>
    </div>

    <!-- 景点网格 -->
    <div v-else class="scenic-grid__grid">
      <ScenicCard v-for="spot in spots" :key="spot.id" :spot="spot" />
    </div>
  </section>
</template>

<style scoped>
.scenic-grid {
  padding: var(--space-16) 0;
}

.scenic-grid__header {
  margin-bottom: var(--space-10);
}

/* 骨架屏网格 */
.scenic-grid__shimmer,
.scenic-grid__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

/* 空状态 */
.scenic-grid__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-16) 0;
  color: var(--color-text-muted);
  font-size: var(--text-base);
}

/* 加载状态 */
.scenic-grid__loading {
  min-height: 400px;
}

@media (max-width: 1024px) {
  .scenic-grid__shimmer,
  .scenic-grid__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .scenic-grid__shimmer,
  .scenic-grid__grid {
    grid-template-columns: 1fr;
  }
  .scenic-grid {
    padding: var(--space-10) 0;
  }
}
</style>
