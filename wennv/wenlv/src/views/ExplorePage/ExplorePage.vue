<script setup lang="ts">
/**
 * ExplorePage.vue — 探索成都
 * 功能：搜索景点 → 分类标签筛选 → 区域筛选 → 景点网格
 * 复用 ScenicCard 组件，展示所有景点数据
 */
import { ref, computed } from 'vue'
import { useLanguageStore } from '@/stores/language'
import { scenicSpots, districts } from '@/data/chengdu'
import DistrictSelector from '@/components/DistrictSelector.vue'
import ScenicCard from '@/components/ScenicCard.vue'
import type { ScenicSpot } from '@/types'

const langStore = useLanguageStore()

/* ── 搜索 ── */
const searchQuery = ref('')

/* ── 区域筛选 ── */
const selectedDistrict = ref('all')
function onDistrictSelect(id: string) {
  selectedDistrict.value = id
}

/* ── 标签筛选 ── */
const allTags = computed(() => {
  const tagSet = new Set<string>()
  scenicSpots.forEach(s => s.tags.forEach(t => tagSet.add(t)))
  return Array.from(tagSet).sort()
})
const selectedTags = ref<string[]>([])

function toggleTag(tag: string) {
  const idx = selectedTags.value.indexOf(tag)
  if (idx >= 0) {
    selectedTags.value.splice(idx, 1)
  } else {
    selectedTags.value.push(tag)
  }
}

function clearTags() {
  selectedTags.value = []
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
</script>

<template>
  <div class="explore-page">
    <!-- ──── HERO / BANNER ──── -->
    <section class="explore-hero">
      <div class="explore-hero__bg">
        <div class="explore-hero__gradient" />
      </div>
      <div class="explore-hero__content">
        <h1 class="explore-hero__title">{{ langStore.t('nav.explore') }}</h1>
        <p class="explore-hero__subtitle">{{ langStore.t('scenic.subtitle') }}</p>

        <!-- 搜索栏 -->
        <div class="explore-search">
          <svg class="explore-search__icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
          </svg>
          <input
            v-model="searchQuery"
            class="explore-search__input"
            :placeholder="langStore.lang === 'zh' ? '搜索景点名称、关键词...' : 'Search attractions, keywords...'"
            type="text"
          />
        </div>
      </div>
    </section>

    <!-- ──── 区域选择 ──── -->
    <section class="explore-district">
      <DistrictSelector @select="onDistrictSelect" />
    </section>

    <!-- ──── 标签筛选 ──── -->
    <section class="explore-tags">
      <div class="container">
        <div class="explore-tags__track">
          <button
            v-for="tag in allTags"
            :key="tag"
            class="explore-tag"
            :class="{ 'explore-tag--active': selectedTags.includes(tag) }"
            @click="toggleTag(tag)"
          >
            {{ tag }}
          </button>
          <button
            v-if="selectedTags.length > 0"
            class="explore-tag explore-tag--clear"
            @click="clearTags()"
          >
            {{ langStore.lang === 'zh' ? '清除筛选' : 'Clear filters' }}
          </button>
        </div>
      </div>
    </section>

    <!-- ──── 结果统计 ──── -->
    <section class="explore-stats">
      <div class="container">
        <div class="explore-stats__bar">
          <span class="explore-stats__count">
            {{ langStore.lang === 'zh'
              ? `找到 ${filteredSpots.length} 个景点`
              : `Found ${filteredSpots.length} attractions` }}
            <span v-if="selectedDistrict !== 'all'" class="explore-stats__district">
              · {{ selectedDistrictName }}
            </span>
          </span>
        </div>
      </div>
    </section>

    <!-- ──── 景点网格 ──── -->
    <section class="explore-grid">
      <div class="container">
        <!-- 空状态 -->
        <div v-if="filteredSpots.length === 0" class="explore-empty">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" style="opacity: 0.3">
            <path d="M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0z"/>
            <path d="M9 9h.01M15 9h.01M7 15h10"/>
          </svg>
          <span>{{ langStore.lang === 'zh' ? '没有找到匹配的景点，试试其他关键词' : 'No matching attractions found. Try different keywords.' }}</span>
        </div>

        <!-- 网格 -->
        <div v-else class="explore-grid__grid">
          <ScenicCard v-for="spot in filteredSpots" :key="spot.id" :spot="spot" />
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* ========================================
   HERO
   ======================================== */
.explore-hero {
  position: relative;
  padding: var(--space-12) 0 var(--space-8);
  overflow: hidden;
}

.explore-hero__bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.explore-hero__gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 80% 60% at 50% 40%, rgba(201, 169, 110, 0.06) 0%, transparent 70%),
    linear-gradient(180deg, rgba(15, 13, 11, 0.3) 0%, var(--color-bg) 100%);
}

.explore-hero__content {
  position: relative;
  z-index: 1;
  text-align: center;
  padding: var(--space-8) var(--space-8) var(--space-6);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}

.explore-hero__title {
  font-family: var(--font-display);
  font-size: var(--text-4xl);
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.1;
}

.explore-hero__subtitle {
  font-family: var(--font-body);
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  font-weight: 300;
  letter-spacing: var(--tracking-wide);
}

/* 搜索栏 */
.explore-search {
  position: relative;
  width: 100%;
  max-width: 520px;
  margin-top: var(--space-2);
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
  padding: var(--space-3) var(--space-4) var(--space-3) var(--space-12);
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

/* ========================================
   DISTRICT - 复用 DistrictSelector 的内边距重置
   ======================================== */
.explore-district {
  margin-top: -2rem;
}

/* ========================================
   TAGS
   ======================================== */
.explore-tags {
  padding: 0 0 var(--space-6);
}

.explore-tags__track {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  justify-content: center;
}

.explore-tag {
  font-size: var(--text-sm);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  background: var(--color-surface);
  transition: all var(--transition-fast);
  letter-spacing: var(--tracking-wide);
}

.explore-tag:hover {
  border-color: var(--color-gold-dark);
  color: var(--color-gold);
}

.explore-tag--active {
  background: var(--color-gold);
  color: var(--color-text-inverse);
  border-color: var(--color-gold);
  font-weight: 500;
}

.explore-tag--active:hover {
  background: var(--color-gold-light);
  border-color: var(--color-gold-light);
  color: var(--color-text-inverse);
}

.explore-tag--clear {
  border-color: var(--color-cinnabar-dim);
  color: var(--color-cinnabar);
  font-size: var(--text-xs);
}

.explore-tag--clear:hover {
  border-color: var(--color-cinnabar);
  color: var(--color-cinnabar);
}

/* ========================================
   STATS
   ======================================== */
.explore-stats {
  padding: 0 0 var(--space-6);
}

.explore-stats__bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border);
}

.explore-stats__count {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  letter-spacing: var(--tracking-wide);
}

.explore-stats__district {
  color: var(--color-gold);
}

/* ========================================
   GRID
   ======================================== */
.explore-grid {
  padding: 0 0 var(--space-16);
}

.explore-grid__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
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

/* ========================================
   RESPONSIVE
   ======================================== */
@media (max-width: 1024px) {
  .explore-grid__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .explore-hero__title {
    font-size: var(--text-3xl);
  }
  .explore-hero__content {
    padding: var(--space-6) var(--space-4);
  }
  .explore-tags__track {
    justify-content: flex-start;
    overflow-x: auto;
    flex-wrap: nowrap;
    padding-bottom: var(--space-2);
    -webkit-overflow-scrolling: touch;
  }
  .explore-tag {
    flex-shrink: 0;
  }
}

@media (max-width: 640px) {
  .explore-grid__grid {
    grid-template-columns: 1fr;
  }
}
</style>