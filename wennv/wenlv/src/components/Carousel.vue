<script setup lang="ts">
/**
 * Carousel.vue — 景点轮播图
 * 交互逻辑:
 *   默认 → 自动播放 (每4s翻页)
 *   @mouseenter → 暂停自动播放
 *   @mouseleave → 恢复自动播放
 *   @wheel      → 根据滚动方向手动翻页 + 重置定时器
 *
 * Props: items — 轮播项数组
 * Emits: change — 当前索引变化
 */
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useLanguageStore } from '@/stores/language'

export interface CarouselItem {
  id: string
  imageUrl: string
  titleZh: string
  titleEn: string
  subtitleZh?: string
  subtitleEn?: string
}

const props = defineProps<{
  items: CarouselItem[]
  interval?: number
}>()

defineEmits<{ change: [index: number] }>()

const langStore = useLanguageStore()

const current = ref(0)
const isPaused = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

const total = computed(() => props.items.length)

/* ---- 自动播放控制 ---- */
function startTimer() {
  stopTimer()
  if (total.value <= 1) return
  timer = setInterval(() => {
    if (!isPaused.value) {
      current.value = (current.value + 1) % total.value
    }
  }, props.interval ?? 4000)
}

function stopTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

function resetTimer() {
  stopTimer()
  startTimer()
}

/* ---- 手动翻页 ---- */
function next() {
  current.value = (current.value + 1) % total.value
  resetTimer()
}

function prev() {
  current.value = (current.value - 1 + total.value) % total.value
  resetTimer()
}

function goTo(n: number) {
  current.value = n
  resetTimer()
}

/* ---- 事件处理 ---- */
function onEnter() {
  isPaused.value = true
}

function onLeave() {
  isPaused.value = false
  resetTimer()
}

function onWheel(e: WheelEvent) {
  e.preventDefault()
  if (e.deltaY > 0) next()
  else prev()
}

onMounted(startTimer)
onUnmounted(stopTimer)
</script>

<template>
  <div
    class="carousel"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
    @wheel.prevent="onWheel"
  >
    <!-- 滑动容器 -->
    <div
      class="carousel__track"
      :style="{ transform: `translateX(-${current * 100}%)` }"
    >
      <div
        v-for="item in items"
        :key="item.id"
        class="carousel__slide"
      >
        <!-- 图片占位 -->
        <div class="carousel__image-placeholder">
          <div class="carousel__pattern" />
          <div class="carousel__overlay" />
          <div class="carousel__content">
            <h2 class="carousel__title">
              {{ langStore.lang === 'zh' ? item.titleZh : item.titleEn }}
            </h2>
            <p v-if="langStore.lang === 'zh' && item.subtitleZh" class="carousel__subtitle">
              {{ item.subtitleZh }}
            </p>
            <p v-else-if="item.subtitleEn" class="carousel__subtitle">
              {{ item.subtitleEn }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- 箭头 -->
    <button class="carousel__arrow carousel__arrow--left" @click="prev" :title="langStore.t('carousel.prev')">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M15 18l-6-6 6-6"/>
      </svg>
    </button>
    <button class="carousel__arrow carousel__arrow--right" @click="next" :title="langStore.t('carousel.next')">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M9 18l6-6-6-6"/>
      </svg>
    </button>

    <!-- 指示器 (蜀锦菱形) -->
    <div class="carousel__dots" role="tablist" aria-label="轮播指示">
      <button
        v-for="(item, i) in items"
        :key="item.id"
        class="carousel__dot"
        :class="{ 'carousel__dot--active': i === current }"
        @click="goTo(i)"
        role="tab"
        :aria-selected="i === current"
      >
        <span class="carousel__diamond">◈</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.carousel {
  position: relative;
  width: 100%;
  border-radius: var(--radius-lg);
  overflow: hidden;
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border);
  user-select: none;
}

.carousel__track {
  display: flex;
  transition: transform 500ms cubic-bezier(0.4, 0, 0.2, 1);
  will-change: transform;
}

.carousel__slide {
  min-width: 100%;
  position: relative;
}

.carousel__image-placeholder {
  position: relative;
  width: 100%;
  aspect-ratio: 21 / 9;
  min-height: 320px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.carousel__pattern {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 60% 50% at 30% 50%, rgba(201, 169, 110, 0.08) 0%, transparent 70%),
    radial-gradient(ellipse 50% 40% at 70% 60%, rgba(162, 59, 59, 0.04) 0%, transparent 60%);
}

.carousel__overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(15, 13, 11, 0.7) 0%, rgba(15, 13, 11, 0.3) 50%, rgba(15, 13, 11, 0.6) 100%);
}

.theme-light .carousel__overlay {
  background: linear-gradient(135deg, rgba(244, 239, 230, 0.7) 0%, rgba(244, 239, 230, 0.2) 50%, rgba(244, 239, 230, 0.5) 100%);
}

.carousel__content {
  position: relative;
  z-index: 1;
  text-align: center;
  padding: var(--space-8);
  max-width: 70%;
}

.carousel__title {
  font-family: var(--font-display);
  font-size: var(--text-4xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.2;
  margin-bottom: var(--space-4);
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
}

.carousel__subtitle {
  font-family: var(--font-body);
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  letter-spacing: var(--tracking-wider);
  max-width: 480px;
  margin: 0 auto;
}

/* 箭头 */
.carousel__arrow {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 44px;
  height: 44px;
  border-radius: var(--radius-full);
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(8px);
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity var(--transition-fast), background var(--transition-fast);
  z-index: 2;
}

.carousel:hover .carousel__arrow {
  opacity: 1;
}

.carousel__arrow:hover {
  background: var(--color-gold);
  color: var(--color-text-inverse);
}

.carousel__arrow--left { left: var(--space-4); }
.carousel__arrow--right { right: var(--space-4); }

/* 指示器 */
.carousel__dots {
  position: absolute;
  bottom: var(--space-6);
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: var(--space-3);
  z-index: 2;
}

.carousel__dot {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  font-size: var(--text-sm);
  line-height: 1;
  padding: var(--space-1);
  transition: color var(--transition-fast);
}

.carousel__diamond {
  display: block;
  transition: font-size var(--transition-base), color var(--transition-fast);
}

.carousel__dot:hover {
  color: var(--color-text-secondary);
}

.carousel__dot--active {
  color: var(--color-gold);
}

.carousel__dot--active .carousel__diamond {
  font-size: var(--text-base);
  text-shadow: 0 0 12px var(--color-gold-glow);
  animation: diamond-pulse 2s ease-in-out infinite;
}

@keyframes diamond-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}


@media (max-width: 768px) {
  .carousel__title {
    font-size: var(--text-2xl);
  }
  .carousel__image-placeholder {
    aspect-ratio: 16 / 9;
    min-height: 220px;
  }
  .carousel__arrow {
    display: none;
  }
  .carousel__content {
    max-width: 90%;
  }
}
</style>
