<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  CULTURE_SCROLL_DECORS,
  CULTURE_SCROLL_HOTSPOTS,
  CULTURE_SCROLL_SEGMENTS,
  type CultureScrollDecor,
  type CultureScrollSegment,
} from '@/data/cultureScroll'
import { useLanguageStore } from '@/stores/language'
import ScrollArtwork from '@/components/culture-scroll/ScrollArtwork.vue'

const WORLD_VW = 6
const AXIS_LOCK_PX = 5
const WALK_EPSILON = 0.0008
const WALK_IDLE_MS = 120
/** 黑线简笔（透明底）叠在 mid 层时间轴之上 */
const LINEART_SRC = '/images/culture-scroll/era-scroll-lineart-transparent-v1.png'
const LINEART_TILES = 4

const langStore = useLanguageStore()
const segments = CULTURE_SCROLL_SEGMENTS
const hotspots = CULTURE_SCROLL_HOTSPOTS
const decors = CULTURE_SCROLL_DECORS
/** Progress distance within which a decor floats in */
const DECOR_ACTIVE_RADIUS = 0.085

const railRef = ref<HTMLElement | null>(null)
const stageRef = ref<HTMLElement | null>(null)
const worldRef = ref<HTMLElement | null>(null)
const progress = ref(0)
const reducedMotion = ref(
  typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches,
)
const viewportWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 0)
const viewportHeight = ref(typeof window !== 'undefined' ? window.innerHeight : 0)
const dragging = ref(false)
const prevProgress = ref(0)
const walking = ref(false)
const walkDir = ref<'left' | 'right'>('right')

let walkIdleTimer: ReturnType<typeof setTimeout> | null = null

function localizedField(
  segment: CultureScrollSegment,
  key: 'era' | 'period' | 'desc',
) {
  const enKey = `${key}En` as const
  return langStore.lang === 'zh' ? segment[key] : segment[enKey]
}

function segmentForHotspot(segmentId: string) {
  return segments.find((seg) => seg.id === segmentId)
}

function isActiveMarker(segmentId: string) {
  return activeSegment.value?.id === segmentId
}

/** Map decor x% on the 600vw mid layer to scrub progress when it sits near viewport center. */
function decorProgressCenter(xPercent: number) {
  return clamp((xPercent / 100) * WORLD_VW - 0.5, 0, WORLD_VW - 1) / (WORLD_VW - 1)
}

function isDecorActive(decor: CultureScrollDecor) {
  return Math.abs(progress.value - decorProgressCenter(decor.xPercent)) <= DECOR_ACTIVE_RADIUS
}

const worldWidth = computed(() => WORLD_VW * viewportWidth.value)
const maxTravel = computed(() => Math.max(0, worldWidth.value - viewportWidth.value))
const railHeightCss = computed(() => `${maxTravel.value + viewportHeight.value}px`)

const worldStyle = computed(() =>
  reducedMotion.value
    ? undefined
    : { width: `${WORLD_VW * 100}vw` },
)

const activeSegment = computed(() => {
  const p = progress.value
  const last = segments[segments.length - 1]
  return segments.find((seg, i) => {
    if (i === segments.length - 1) return p >= seg.progressStart
    return p >= seg.progressStart && p < seg.progressEnd
  }) ?? last
})

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function prefersReducedMotion() {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

function layerStyle(factor: number) {
  if (reducedMotion.value) return undefined
  const x = -Math.min(progress.value * maxTravel.value * factor, maxTravel.value)
  return { transform: `translate3d(${x}px, 0, 0)` }
}

function syncWalkerFromProgress() {
  if (reducedMotion.value) {
    walking.value = false
    prevProgress.value = progress.value
    return
  }

  const delta = progress.value - prevProgress.value
  if (Math.abs(delta) > WALK_EPSILON) {
    walking.value = true
    walkDir.value = delta > 0 ? 'right' : 'left'
  }
  prevProgress.value = progress.value

  if (walkIdleTimer) clearTimeout(walkIdleTimer)
  walkIdleTimer = setTimeout(() => {
    walking.value = false
  }, WALK_IDLE_MS)
}

function updateProgress() {
  if (reducedMotion.value) {
    const scroller = worldRef.value
    if (!scroller) return
    const max = scroller.scrollWidth - scroller.clientWidth
    progress.value = max <= 0 ? 0 : clamp(scroller.scrollLeft / max, 0, 1)
    syncWalkerFromProgress()
    return
  }

  const rail = railRef.value
  if (!rail) return
  const railTop = window.scrollY + rail.getBoundingClientRect().top
  const denom = rail.offsetHeight - window.innerHeight
  progress.value = denom <= 0 ? 0 : clamp((window.scrollY - railTop) / denom, 0, 1)
  syncWalkerFromProgress()
}

function measure() {
  viewportWidth.value = window.innerWidth
  viewportHeight.value = window.innerHeight
  updateProgress()
}

type DragAxis = 'x' | 'y' | null

const drag = {
  active: false,
  pointerId: -1,
  lastX: 0,
  lastY: 0,
  axis: null as DragAxis,
}

function applyHorizontalDrag(dx: number) {
  const rail = railRef.value
  if (!rail) return
  const denom = Math.max(1, rail.offsetHeight - viewportHeight.value)
  const travel = Math.max(1, maxTravel.value)
  const next = clamp(progress.value - dx / travel, 0, 1)
  const railTop = window.scrollY + rail.getBoundingClientRect().top
  window.scrollTo(0, railTop + next * denom)
}

function onPointerDown(event: PointerEvent) {
  if (reducedMotion.value) return
  if (event.button !== 0) return
  drag.active = true
  drag.pointerId = event.pointerId
  drag.lastX = event.clientX
  drag.lastY = event.clientY
  drag.axis = null
  dragging.value = true
  stageRef.value?.setPointerCapture(event.pointerId)
}

function onPointerMove(event: PointerEvent) {
  if (!drag.active || event.pointerId !== drag.pointerId) return
  const dx = event.clientX - drag.lastX
  const dy = event.clientY - drag.lastY

  if (!drag.axis) {
    if (Math.abs(dx) < AXIS_LOCK_PX && Math.abs(dy) < AXIS_LOCK_PX) return
    drag.axis = Math.abs(dx) > Math.abs(dy) ? 'x' : 'y'
  }

  if (drag.axis !== 'x') return
  event.preventDefault()
  drag.lastX = event.clientX
  drag.lastY = event.clientY
  applyHorizontalDrag(dx)
}

function endDrag(event: PointerEvent) {
  if (!drag.active || event.pointerId !== drag.pointerId) return
  drag.active = false
  drag.axis = null
  dragging.value = false
  if (stageRef.value?.hasPointerCapture(event.pointerId)) {
    stageRef.value.releasePointerCapture(event.pointerId)
  }
}

function onWindowScroll() {
  if (reducedMotion.value) return
  updateProgress()
}

function onStageScroll() {
  updateProgress()
}

function syncScrollListeners() {
  window.removeEventListener('scroll', onWindowScroll)
  window.removeEventListener('scroll', updateProgress)
  if (reducedMotion.value) {
    window.addEventListener('scroll', onWindowScroll, { passive: true })
  } else {
    window.addEventListener('scroll', updateProgress, { passive: true })
  }
}

function onMotionChange() {
  reducedMotion.value = prefersReducedMotion()
  syncScrollListeners()
  measure()
}

let motionQuery: MediaQueryList | null = null
let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  reducedMotion.value = prefersReducedMotion()
  motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  motionQuery.addEventListener('change', onMotionChange)

  measure()
  syncScrollListeners()
  window.addEventListener('resize', measure)

  const stage = stageRef.value
  const world = worldRef.value
  if (world) {
    world.addEventListener('scroll', onStageScroll, { passive: true })
  }
  if (stage) {
    stage.addEventListener('pointermove', onPointerMove, { passive: false })
  }

  if (typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => measure())
    if (railRef.value) resizeObserver.observe(railRef.value)
    if (stage) resizeObserver.observe(stage)
    if (world) resizeObserver.observe(world)
  }
})

onBeforeUnmount(() => {
  if (walkIdleTimer) clearTimeout(walkIdleTimer)
  motionQuery?.removeEventListener('change', onMotionChange)
  window.removeEventListener('scroll', onWindowScroll)
  window.removeEventListener('scroll', updateProgress)
  window.removeEventListener('resize', measure)
  const stage = stageRef.value
  worldRef.value?.removeEventListener('scroll', onStageScroll)
  stage?.removeEventListener('pointermove', onPointerMove)
  resizeObserver?.disconnect()
})
</script>

<template>
  <section
    ref="railRef"
    class="culture-scroll"
    :class="{
      'culture-scroll--static': reducedMotion,
      'culture-scroll--dragging': dragging,
    }"
    :style="reducedMotion ? undefined : { height: railHeightCss }"
  >
    <div
      ref="stageRef"
      class="culture-scroll__stage"
      @pointerdown="onPointerDown"
      @pointerup="endDrag"
      @pointercancel="endDrag"
    >
      <div ref="worldRef" class="culture-scroll__world" :style="worldStyle">
        <div class="culture-scroll__layer culture-scroll__layer--far" :style="layerStyle(0.35)">
          <ScrollArtwork layer="far" />
        </div>

        <div class="culture-scroll__layer culture-scroll__layer--mid" :style="layerStyle(1)">
          <div class="culture-scroll__lineart-track" data-scroll-art="mid-lineart" aria-hidden="true">
            <img
              v-for="n in LINEART_TILES"
              :key="`line-${n}`"
              class="culture-scroll__lineart"
              :src="LINEART_SRC"
              alt=""
              draggable="false"
            />
          </div>
          <div class="culture-scroll__timeline" aria-hidden="false">
            <div class="culture-scroll__timeline-line" aria-hidden="true" />
            <figure
              v-for="decor in decors"
              :key="decor.id"
              class="culture-scroll__decor"
              :class="{
                'culture-scroll__decor--above': decor.side === 'above',
                'culture-scroll__decor--below': decor.side === 'below',
                'culture-scroll__decor--sm': decor.size === 'sm',
                'culture-scroll__decor--md': decor.size === 'md',
                'culture-scroll__decor--active': isDecorActive(decor),
              }"
              :data-decor="decor.id"
              :style="{ left: `${decor.xPercent}%` }"
            >
              <img :src="decor.imageUrl" alt="" draggable="false" />
            </figure>
            <article
              v-for="(hotspot, index) in hotspots"
              :key="hotspot.segmentId"
              class="culture-scroll__marker"
              :class="{
                'culture-scroll__marker--above': index % 2 === 0,
                'culture-scroll__marker--below': index % 2 === 1,
                'culture-scroll__marker--active': isActiveMarker(hotspot.segmentId),
              }"
              :data-marker="hotspot.segmentId"
              :data-active="isActiveMarker(hotspot.segmentId) ? 'true' : 'false'"
              :style="{ left: `${hotspot.xPercent}%` }"
            >
              <span class="culture-scroll__dot" aria-hidden="true" />
              <p class="culture-scroll__axis-time">
                {{ localizedField(segmentForHotspot(hotspot.segmentId)!, 'period') }}
              </p>
              <div class="culture-scroll__era-card">
                <h3 class="culture-scroll__era-card-title">
                  {{ localizedField(segmentForHotspot(hotspot.segmentId)!, 'era') }}
                </h3>
                <p class="culture-scroll__era-card-period">
                  {{ localizedField(segmentForHotspot(hotspot.segmentId)!, 'period') }}
                </p>
                <p class="culture-scroll__era-card-desc">
                  {{ localizedField(segmentForHotspot(hotspot.segmentId)!, 'desc') }}
                </p>
              </div>
              <div class="culture-scroll__era-media">
                <img
                  :src="segmentForHotspot(hotspot.segmentId)!.imageUrl"
                  alt=""
                  draggable="false"
                />
              </div>
            </article>
          </div>
        </div>
      </div>

      <p class="culture-scroll__colophon">
        {{ activeSegment?.era }} · {{ activeSegment?.eraEn }}
      </p>
    </div>
  </section>
</template>

<style scoped>
.culture-scroll {
  position: relative;
  background: var(--color-bg-alt);
}

.culture-scroll__stage {
  position: sticky;
  top: 0;
  height: 100vh;
  overflow: hidden;
  touch-action: pan-y;
  cursor: grab;
  background:
    radial-gradient(ellipse 80% 50% at 50% 0%, #e4eee8 0%, transparent 58%),
    linear-gradient(180deg, #e8efe8 0%, var(--color-bg-alt) 42%, #d9cbb0 100%);
}

.culture-scroll--dragging .culture-scroll__stage {
  cursor: grabbing;
}

.culture-scroll__world {
  position: relative;
  height: 100%;
}

.culture-scroll__layer {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  width: 600vw;
  will-change: transform;
  pointer-events: none;
}

.culture-scroll__layer--far {
  color: color-mix(in srgb, var(--color-sage) 62%, #b7c4b4);
  opacity: 0.52;
}

.culture-scroll__layer--mid {
  z-index: 1;
  color: var(--color-text-primary);
}

.culture-scroll__lineart-track {
  display: flex;
  align-items: flex-end;
  width: 100%;
  height: 100%;
  pointer-events: none;
  position: absolute;
  inset: 0;
  z-index: 1;
}

.culture-scroll__lineart {
  flex: 1 0 0;
  width: 0;
  height: 92%;
  object-fit: cover;
  object-position: center bottom;
  user-select: none;
  -webkit-user-drag: none;
  mix-blend-mode: normal;
  /* 整体略下移，贴近地平线 */
  transform: translateY(6%);
  object-position: center 62%;
}

.culture-scroll__layer--near {
  z-index: 2;
  color: color-mix(in srgb, var(--color-text-primary) 90%, #3f3428);
}

.culture-scroll__walker {
  position: absolute;
  left: 42%;
  bottom: 18%;
  z-index: 3;
  color: var(--color-text-primary);
  filter: drop-shadow(0 2px 0 color-mix(in srgb, var(--color-bg) 70%, transparent));
  pointer-events: none;
}

.culture-scroll__walker-figure {
  transform-origin: center bottom;
}

.culture-scroll__walker[data-dir='left'] .culture-scroll__walker-figure {
  transform: scaleX(-1);
}

.culture-scroll__walker[data-walking='true'][data-dir='right'] .culture-scroll__walker-figure {
  animation: walk-bob-right 0.45s ease-in-out infinite;
}

.culture-scroll__walker[data-walking='true'][data-dir='left'] .culture-scroll__walker-figure {
  animation: walk-bob-left 0.45s ease-in-out infinite;
}

.culture-scroll__walker[data-walking='true'] .culture-scroll__leg--l {
  animation: leg-swing-l 0.45s ease-in-out infinite;
  transform-origin: 20px 38px;
}

.culture-scroll__walker[data-walking='true'] .culture-scroll__leg--r {
  animation: leg-swing-r 0.45s ease-in-out infinite;
  transform-origin: 20px 38px;
}

@keyframes walk-bob-right {
  0%,
  100% {
    transform: scaleX(1) translateY(0);
  }
  50% {
    transform: scaleX(1) translateY(-2px);
  }
}

@keyframes walk-bob-left {
  0%,
  100% {
    transform: scaleX(-1) translateY(0);
  }
  50% {
    transform: scaleX(-1) translateY(-2px);
  }
}

@keyframes leg-swing-l {
  0%,
  100% {
    transform: rotate(8deg);
  }
  50% {
    transform: rotate(-14deg);
  }
}

@keyframes leg-swing-r {
  0%,
  100% {
    transform: rotate(-8deg);
  }
  50% {
    transform: rotate(14deg);
  }
}

.culture-scroll__walker svg {
  fill: none;
  stroke: currentColor;
  stroke-width: 2.2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.culture-scroll__leg {
  vector-effect: non-scaling-stroke;
}

.culture-scroll__colophon {
  position: absolute;
  left: var(--space-6);
  bottom: var(--space-5);
  z-index: 4;
  font-family: var(--font-display);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  color: color-mix(in srgb, var(--color-text-primary) 72%, transparent);
  pointer-events: none;
}

.culture-scroll__timeline {
  position: absolute;
  top: 56%;
  left: 0;
  width: 100%;
  height: 0;
  z-index: 4;
  pointer-events: none;
}

.culture-scroll__timeline-line {
  position: absolute;
  left: 4%;
  right: 4%;
  top: 0;
  height: 1px;
  background: color-mix(in srgb, var(--color-text-primary) 42%, transparent);
}

.culture-scroll__decor {
  position: absolute;
  top: 0;
  z-index: 2;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--color-border) 70%, transparent);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-surface) 55%, transparent);
  box-shadow: 0 1px 0 color-mix(in srgb, var(--color-text-primary) 5%, transparent);
  opacity: 0;
  pointer-events: none;
  will-change: opacity, transform;
  transition:
    opacity 0.65s cubic-bezier(0.22, 1, 0.36, 1),
    transform 0.65s cubic-bezier(0.22, 1, 0.36, 1);
}

.culture-scroll__decor--sm {
  width: min(168px, 14vw);
  height: 112px;
}

.culture-scroll__decor--md {
  width: min(210px, 17vw);
  height: 140px;
}

.culture-scroll__decor--above {
  transform: translate(-50%, 28px) rotate(-2.5deg);
  bottom: calc(100% + 56px);
  top: auto;
}

.culture-scroll__decor--below {
  transform: translate(-50%, 28px) rotate(2.5deg);
  top: 52px;
}

.culture-scroll__decor img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  user-select: none;
  -webkit-user-drag: none;
  filter: saturate(0.92) contrast(1.02);
}

.culture-scroll__decor--active.culture-scroll__decor--above {
  opacity: 0.92;
  transform: translate(-50%, 0) rotate(-2.5deg);
}

.culture-scroll__decor--active.culture-scroll__decor--below {
  opacity: 0.92;
  transform: translate(-50%, 0) rotate(2.5deg);
}

.culture-scroll__marker {
  position: absolute;
  top: 0;
  z-index: 3;
  width: min(460px, 36vw);
  transform: translate(-50%, 0);
}

.culture-scroll__dot {
  position: absolute;
  left: 50%;
  top: 0;
  width: 10px;
  height: 10px;
  border-radius: var(--radius-full);
  border: 1.5px solid color-mix(in srgb, var(--color-gold) 75%, #fff);
  background: color-mix(in srgb, var(--color-gold) 85%, #f4ead0);
  transform: translate(-50%, -50%);
}

.culture-scroll__marker--active .culture-scroll__dot {
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--color-gold) 22%, transparent);
}

.culture-scroll__axis-time {
  position: absolute;
  left: 50%;
  top: 10px;
  margin: 0;
  transform: translateX(-50%);
  white-space: nowrap;
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wider);
  color: color-mix(in srgb, var(--color-text-primary) 70%, transparent);
}

.culture-scroll__marker--above .culture-scroll__axis-time {
  top: auto;
  bottom: 14px;
}

.culture-scroll__era-card {
  position: absolute;
  left: 50%;
  width: 100%;
  padding: var(--space-4) var(--space-5);
  border: 1px solid color-mix(in srgb, var(--color-border) 80%, transparent);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-surface) 88%, transparent);
  box-shadow: 0 1px 0 color-mix(in srgb, var(--color-text-primary) 6%, transparent);
  opacity: 0;
  transform: translateX(-50%) translateY(28px);
  transition:
    opacity 0.65s cubic-bezier(0.22, 1, 0.36, 1),
    transform 0.65s cubic-bezier(0.22, 1, 0.36, 1),
    box-shadow 0.45s ease;
  will-change: opacity, transform;
}

.culture-scroll__marker--above .culture-scroll__era-card {
  bottom: calc(100% + 28px);
}

.culture-scroll__marker--below .culture-scroll__era-card {
  top: 36px;
}

.culture-scroll__era-media {
  position: absolute;
  left: 50%;
  width: 100%;
  height: 240px;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--color-border) 80%, transparent);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-surface) 70%, transparent);
  opacity: 0;
  transform: translateX(-50%) translateY(28px);
  transition:
    opacity 0.65s cubic-bezier(0.22, 1, 0.36, 1) 0.1s,
    transform 0.65s cubic-bezier(0.22, 1, 0.36, 1) 0.1s;
  pointer-events: none;
  will-change: opacity, transform;
}

.culture-scroll__era-media img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  user-select: none;
  -webkit-user-drag: none;
}

/* text above → media below axis */
.culture-scroll__marker--above .culture-scroll__era-media {
  top: 48px;
}

/* text below → media above axis */
.culture-scroll__marker--below .culture-scroll__era-media {
  bottom: calc(100% + 28px);
}

/* 滚到当前区段：文案卡与图片卡上浮显现 */
.culture-scroll__marker--active .culture-scroll__era-card,
.culture-scroll__marker--active .culture-scroll__era-media {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
}

.culture-scroll__marker--active .culture-scroll__era-card {
  box-shadow:
    0 1px 0 color-mix(in srgb, var(--color-text-primary) 6%, transparent),
    0 12px 28px color-mix(in srgb, var(--color-text-primary) 10%, transparent);
}

.culture-scroll__era-card-title {
  margin: 0 0 var(--space-1);
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-primary);
}

.culture-scroll__era-card-period {
  margin: 0 0 var(--space-2);
  font-family: var(--font-en-body);
  font-size: var(--text-sm);
  color: var(--color-gold);
}

.culture-scroll__era-card-desc {
  margin: 0;
  font-size: var(--text-base);
  line-height: var(--leading-relaxed);
  color: var(--color-text-secondary);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 10;
  overflow: hidden;
}

@media (max-width: 720px) {
  .culture-scroll__marker {
    width: min(320px, 56vw);
  }
  .culture-scroll__era-card-desc {
    -webkit-line-clamp: 7;
  }
  .culture-scroll__era-media {
    height: 180px;
  }
  .culture-scroll__decor--sm {
    width: min(120px, 28vw);
    height: 84px;
  }
  .culture-scroll__decor--md {
    width: min(148px, 34vw);
    height: 100px;
  }
}

.culture-scroll--static {
  height: auto;
}

.culture-scroll--static .culture-scroll__era-card,
.culture-scroll--static .culture-scroll__era-media {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
  transition: none;
  box-shadow: 0 1px 0 color-mix(in srgb, var(--color-text-primary) 6%, transparent);
}

.culture-scroll--static .culture-scroll__decor--above {
  opacity: 0.88;
  transform: translate(-50%, 0) rotate(-2.5deg);
  transition: none;
}

.culture-scroll--static .culture-scroll__decor--below {
  opacity: 0.88;
  transform: translate(-50%, 0) rotate(2.5deg);
  transition: none;
}

@media (prefers-reduced-motion: reduce) {
  .culture-scroll__era-card,
  .culture-scroll__era-media {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
    transition: none;
  }

  .culture-scroll__decor--above {
    opacity: 0.88;
    transform: translate(-50%, 0) rotate(-2.5deg);
    transition: none;
  }

  .culture-scroll__decor--below {
    opacity: 0.88;
    transform: translate(-50%, 0) rotate(2.5deg);
    transition: none;
  }
}

.culture-scroll--static .culture-scroll__stage {
  position: relative;
  height: min(70vh, 640px);
  overflow: hidden;
  touch-action: pan-x;
  cursor: default;
}

.culture-scroll--static .culture-scroll__world {
  height: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior-x: contain;
}

.culture-scroll--static .culture-scroll__layer {
  will-change: auto;
}

.culture-scroll--static .culture-scroll__walker {
  position: sticky;
  left: 42%;
}
</style>
