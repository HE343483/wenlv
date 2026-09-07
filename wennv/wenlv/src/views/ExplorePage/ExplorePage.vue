<script setup lang="ts">
/**
 * ExplorePage.vue — 探索成都
 * 功能：搜索景点 → 区域筛选 → 多选标签筛选 → 景点网格
 * 筛选区：自定义下拉（区域）+ 多选标签下拉 + 活跃筛选 Chips
 */
import { ref, computed, watch } from 'vue'
import { useLanguageStore } from '@/stores/language'
import { scenicSpots, districts } from '@/data/chengdu'
import TagFilter from '@/components/TagFilter.vue'
import DistrictFilter from '@/components/DistrictFilter.vue'
import Carousel from '@/components/Carousel.vue'
import ScenicCard from '@/components/ScenicCard.vue'
import HomeBanner from '@/components/HomeBanner.vue'
import type { CarouselItem } from '@/components/Carousel.vue'

const langStore = useLanguageStore()

/* ── 搜索 ── */
const searchQuery = ref('')

/* ── 区域筛选 ── */
const selectedDistrict = ref('all')


/* ── 标签筛选（多选）── */
const allTags = computed(() => {
  const tagSet = new Set<string>()
  scenicSpots.forEach(s => s.tags.forEach(t => tagSet.add(t)))
  return Array.from(tagSet).sort()
})
const selectedTags = ref<string[]>([])

function onTagsUpdate(tags: string[]) {
  selectedTags.value = tags
}

/* ── 筛选结果 ── */
const filteredSpots = computed(() => {
  let result = scenicSpots

  // 区域筛选
  if (selectedDistrict.value !== 'all') {
    result = result.filter(s => s.districtId === selectedDistrict.value)
  }

  // 标签筛选
  if (selectedTags.value.length > 0) {
    result = result.filter(s => selectedTags.value.some(t => s.tags.includes(t)))
  }

  // 搜索
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    result = result.filter(s =>
      s.nameZh.toLowerCase().includes(q) ||
      s.nameEn.toLowerCase().includes(q) ||
      s.shortDescZh.toLowerCase().includes(q) ||
      s.shortDescEn.toLowerCase().includes(q) ||
      s.tags.some(t => t.toLowerCase().includes(q))
    )
  }

  return result
})

/* ── 已选区域名称 ── */
const selectedDistrictName = computed(() => {
  if (selectedDistrict.value === 'all') return langStore.t('district.allDistricts')
  const d = districts.find(d => d.id === selectedDistrict.value)
  if (!d) return ''
  return langStore.lang === 'zh' ? d.nameZh : d.nameEn
})

/* ── 选中区县的景点 → 轮播数据 ── */
const carouselItems = computed<CarouselItem[]>(() => {
  // 仅当选中了具体区县时使用轮播展示
  if (selectedDistrict.value === 'all') return []
  return filteredSpots.value.map(spot => ({
    id: spot.id,
    imageUrl: spot.imageUrl,
    titleZh: spot.nameZh,
    titleEn: spot.nameEn,
    subtitleZh: spot.shortDescZh,
    subtitleEn: spot.shortDescEn,
  }))
})

const showCarousel = computed(() => selectedDistrict.value !== 'all' && carouselItems.value.length > 0)

/* ── 活跃筛选 Chips 数据 ── */
const activeFilters = computed(() => {
  const chips: Array<{ kind: 'district' | 'tag'; label: string; value: string }> = []
  if (selectedDistrict.value !== 'all') {
    chips.push({ kind: 'district', label: selectedDistrictName.value, value: selectedDistrict.value })
  }
  selectedTags.value.forEach(t => {
    chips.push({ kind: 'tag', label: t, value: t })
  })
  return chips
})

const hasActiveFilters = computed(() => activeFilters.value.length > 0)

function removeFilter(chip: { kind: 'district' | 'tag'; value: string }) {
  if (chip.kind === 'district') {
    selectedDistrict.value = 'all'
  } else {
    selectedTags.value = selectedTags.value.filter(t => t !== chip.value)
  }
}

function clearAllFilters() {
  selectedDistrict.value = 'all'
  selectedTags.value = []
  searchQuery.value = ''
}

/* ── 清空搜索后自动聚焦搜索框 ── */
const searchInputRef = ref<HTMLInputElement | null>(null)
function clearSearch() {
  searchQuery.value = ''
  searchInputRef.value?.focus()
}

/* ── 网格入场动画：筛选结果变化时重置 key 触发 TransitionGroup ── */
const gridKey = ref(0)
watch(filteredSpots, () => {
  gridKey.value++
})

</script>

<template>
  <div class="explore-page">
    <!-- ──── HERO / BANNER ──── -->
    <HomeBanner
      :eyebrow="langStore.t('scenic.badge')"
      :title="langStore.t('scenic.heroTitle')"
      :subtitle="langStore.t('scenic.subtitle')"
      watermark="游"
    >
      <!-- 搜索栏 -->
      <div class="explore-search">
        <svg
          class="explore-search__icon"
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
        >
          <circle cx="11" cy="11" r="7"/>
          <path d="M21 21l-4.3-4.3"/>
        </svg>
        <input
          v-model="searchQuery"
          class="explore-search__input"
          type="text"
          :placeholder="langStore.t('scenic.searchPlaceholder')"
          :aria-label="langStore.t('scenic.searchPlaceholder')"
        />
        <button
          v-if="searchQuery"
          class="explore-search__clear"
          @click="clearSearch"
          :title="langStore.lang === 'zh' ? '清空搜索' : 'Clear search'"
          :aria-label="langStore.lang === 'zh' ? '清空搜索' : 'Clear search'"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
            <path d="M18 6L6 18M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </HomeBanner>

    <!-- ──── 筛选面板 ──── -->
    <section class="explore-filters">
      <div class="container">
        <div class="explore-filters__bar">
          <!-- 左：筛选控件 -->
          <div class="explore-filters__controls">
            <DistrictFilter v-model="selectedDistrict" />
            <TagFilter :tags="allTags" :selected="selectedTags" @update:selected="onTagsUpdate" @clear="clearAllFilters" />
          </div>

          <!-- 右：结果统计 -->
          <div class="explore-filters__stats">
            <Transition name="stat-pop" mode="out-in">
              <span :key="filteredSpots.length" class="explore-stats__count">
                <b class="explore-stats__num">{{ filteredSpots.length }}</b>
                <span class="explore-stats__unit">{{ langStore.lang === 'zh' ? '个景点' : 'spots' }}</span>
                <span v-if="selectedDistrict !== 'all'" class="explore-stats__district">
                  · {{ selectedDistrictName }}
                </span>
              </span>
            </Transition>
          </div>
        </div>

        <!-- 活跃筛选 Chips -->
        <Transition name="chips-slide">
          <div v-if="hasActiveFilters" class="explore-filters__chips">
            <span class="explore-filters__chips-label">
              {{ langStore.lang === 'zh' ? '筛选' : 'Filtered by' }}
            </span>
            <TransitionGroup name="chip" tag="div" class="explore-filters__chips-track">
              <button
                v-for="chip in activeFilters"
                :key="`${chip.kind}-${chip.value}`"
                class="filter-chip"
                :class="`filter-chip--${chip.kind}`"
                @click="removeFilter(chip)"
                :title="langStore.lang === 'zh' ? '移除筛选' : 'Remove filter'"
              >
                <span class="filter-chip__label">{{ chip.label }}</span>
                <svg class="filter-chip__x" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                  <path d="M18 6L6 18M6 6l12 12"/>
                </svg>
              </button>
            </TransitionGroup>
            <button class="explore-filters__clear-all" @click="clearAllFilters">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2m3 0v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/>
              </svg>
              {{ langStore.lang === 'zh' ? '清除全部' : 'Clear all' }}
            </button>
          </div>
        </Transition>
      </div>
    </section>

    <!-- ──── 景点展示 ──── -->
    <section class="explore-grid">
      <div class="container">
        <!-- 空状态 -->
        <div v-if="filteredSpots.length === 0" class="explore-empty">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" style="opacity: 0.3">
            <path d="M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0z"/>
            <path d="M9 9h.01M15 9h.01M7 15h10"/>
          </svg>
          <span>{{ langStore.lang === 'zh' ? '没有找到匹配的景点，试试其他关键词' : 'No matching attractions found. Try different keywords.' }}</span>
          <button v-if="hasActiveFilters || searchQuery" class="explore-empty__reset" @click="clearAllFilters">
            {{ langStore.lang === 'zh' ? '重置所有筛选' : 'Reset all filters' }}
          </button>
        </div>

        <!-- 选中具体区县 → 轮播展示 -->
        <div v-else-if="showCarousel" class="explore-carousel">
          <Carousel :key="`${selectedDistrict}-${langStore.lang}`" :items="carouselItems" />
        </div>

        <!-- 全部区域 → 网格 -->
        <TransitionGroup v-else :key="gridKey" name="grid" tag="div" class="explore-grid__grid">
          <ScenicCard
            v-for="(spot, i) in filteredSpots"
            :key="spot.id"
            :spot="spot"
            :style="{ '--i': Math.min(i, 8) }"
          />
        </TransitionGroup>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* ========================================
   搜索栏（欢迎横幅内）
   ======================================== */
.explore-search {
  position: relative;
  width: 100%;
  max-width: 520px;
}

.explore-search__icon {
  position: absolute;
  left: var(--space-4);
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-muted);
  pointer-events: none;
}

.explore-search__input {
  width: 100%;
  padding: var(--space-3) var(--space-12) var(--space-3) var(--space-12);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: var(--text-base);
  outline: none;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.explore-search__input:focus {
  border-color: var(--color-gold);
  box-shadow: 0 0 0 2px var(--color-gold-glow);
}

.explore-search__input::placeholder {
  color: var(--color-text-muted);
}

.explore-search__clear {
  position: absolute;
  right: var(--space-4);
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-full);
  color: var(--color-text-muted);
  transition: all var(--transition-fast);
}

.explore-search__clear:hover {
  color: var(--color-gold);
  background: var(--color-surface-hover);
}

/* ========================================
   筛选面板
   ======================================== */
.explore-filters {
  padding-bottom: var(--space-10);
}

.explore-filters__bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  flex-wrap: wrap;
  padding: var(--space-4);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  position: relative;
}

.explore-filters__controls {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

/* 结果统计 */
.explore-filters__stats {
  display: flex;
  align-items: center;
}

.explore-stats__count {
  display: inline-flex;
  align-items: baseline;
  gap: var(--space-1);
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  white-space: nowrap;
}

.explore-stats__num {
  font-family: var(--font-en-display);
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-gold);
  line-height: 1;
}

.explore-stats__unit {
  color: var(--color-text-secondary);
}

.explore-stats__district {
  color: var(--color-gold-dark);
  margin-left: var(--space-1);
}

.stat-pop-enter-active,
.stat-pop-leave-active {
  transition: all var(--transition-fast);
}

.stat-pop-enter-from {
  opacity: 0;
  transform: translateY(6px);
}

.stat-pop-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* ========================================
   活跃筛选 Chips
   ======================================== */
.explore-filters__chips {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px dashed var(--color-border);
}

.explore-filters__chips-label {
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
  margin-right: var(--space-1);
}

.explore-filters__chips-track {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  padding: 4px 10px;
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wide);
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
}

.filter-chip:hover {
  border-color: var(--color-cinnabar);
  color: var(--color-cinnabar);
}

.filter-chip--district {
  border-color: color-mix(in srgb, var(--color-gold) 45%, transparent);
  color: var(--color-gold-dark);
}

.filter-chip--district:hover {
  border-color: var(--color-gold);
  color: var(--color-gold);
}

.filter-chip--tag {
  border-color: color-mix(in srgb, var(--color-sage) 45%, transparent);
  color: var(--color-sage);
}

.filter-chip--tag:hover {
  border-color: var(--color-sage);
  color: var(--color-sage);
}

.filter-chip__x {
  opacity: 0.6;
  transition: opacity var(--transition-fast);
}

.filter-chip:hover .filter-chip__x {
  opacity: 1;
}

.explore-filters__clear-all {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-xs);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  padding: 4px 10px;
  border-radius: var(--radius-full);
  border: 1px solid transparent;
  transition: all var(--transition-fast);
}

.explore-filters__clear-all:hover {
  color: var(--color-cinnabar);
  border-color: var(--color-cinnabar-dim);
  background: var(--color-cinnabar-dim);
}

/* Chips 过渡 */
.chips-slide-enter-active,
.chips-slide-leave-active {
  transition: all var(--transition-base);
}

.chips-slide-enter-from,
.chips-slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.chips-slide-leave-active {
  position: absolute;
  width: 100%;
  left: 0;
}

.chip-enter-active {
  transition: all var(--transition-fast);
}

.chip-leave-active {
  transition: all 200ms cubic-bezier(0.4, 0, 0.2, 1);
  position: absolute;
}

.chip-enter-from,
.chip-leave-to {
  opacity: 0;
  transform: scale(0.7);
}

.chip-move {
  transition: transform var(--transition-base);
}

/* ========================================
   网格
   ======================================== */
.explore-grid {
  padding: 0 0 var(--space-16);
}

.explore-grid__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

/* 网格入场动画 */
.grid-enter-active {
  transition: all 400ms cubic-bezier(0.4, 0, 0.2, 1);
  transition-delay: calc(var(--i, 0) * 40ms);
}

.grid-enter-from {
  opacity: 0;
  transform: translateY(16px);
}

/* 空状态 */
.explore-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-16) 0;
  color: var(--color-text-muted);
  font-size: var(--text-base);
  text-align: center;
}

.explore-empty__reset {
  font-size: var(--text-sm);
  color: var(--color-gold);
  border: 1px solid var(--color-gold-dark);
  border-radius: var(--radius-full);
  padding: var(--space-2) var(--space-5);
  letter-spacing: var(--tracking-wide);
  transition: all var(--transition-fast);
}

.explore-empty__reset:hover {
  background: var(--color-gold);
  color: var(--color-text-inverse);
  border-color: var(--color-gold);
}

/* ========================================
   区县导览地图
   ======================================== */
.explore-map {
  padding: 0 0 var(--space-8);
}

.explore-map__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

.explore-map__reset {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.explore-map__reset:hover {
  color: var(--color-gold);
  border-color: var(--color-gold-dark);
  background: color-mix(in srgb, var(--color-gold) 8%, transparent);
}

/* ========================================
   轮播
   ======================================== */
.explore-carousel {
  max-width: 900px;
  margin: 0 auto;
}

/* ========================================
   Responsive
   ======================================== */
@media (max-width: 1024px) {
  .explore-grid__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .explore-filters__bar {
    align-items: stretch;
    flex-direction: column;
  }
  .explore-filters__controls {
    width: 100%;
  }
  .explore-filters__controls > * {
    flex: 1;
  }
  .explore-filters__stats {
    justify-content: flex-end;
    padding-top: var(--space-2);
    border-top: 1px solid var(--color-border);
  }
}

@media (max-width: 640px) {
  .explore-grid__grid {
    grid-template-columns: 1fr;
  }
}
</style>
