<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  CULTURE_SCROLL_HOTSPOTS,
  CULTURE_SCROLL_SEGMENTS,
  type CultureScrollHotspot,
  type CultureScrollSegment,
} from '@/data/cultureScroll'
import { useLanguageStore } from '@/stores/language'
import ScrollArtwork from '@/components/culture-scroll/ScrollArtwork.vue'

const WORLD_VW = 6
const AXIS_LOCK_PX = 5
const WALK_EPSILON = 0.0008
const WALK_IDLE_MS = 120
/** 黑线简笔（透明底）叠在黄蓝山水插画之上（非实景照片） */
const SCENERY_SRC = '/images/culture-scroll/era-scenery-shanshui-v1.jpg'
const LINEART_SRC = '/images/culture-scroll/era-scroll-lineart-transparent-v1.png'
const LINEART_TILES = 4

const langStore = useLanguageStore()
const segments = CULTURE_SCROLL_SEGMENTS
const hotspots = CULTURE_SCROLL_HOTSPOTS

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
const openHotspotId = ref<string | null>(null)

let walkIdleTimer: ReturnType<typeof setTimeout> | null = null

const activeHotspot = computed(() =>
  openHotspotId.value
    ? segments.find((seg) => seg.id === openHotspotId.value) ?? null
    : null,
)

function localizedField(
  segment: CultureScrollSegment,
  key: 'era' | 'period' | 'desc',
) {
  const enKey = `${key}En` as const
  return langStore.lang === 'zh' ? segment[key] : segment[enKey]
}

function hotspotAriaLabel(hotspot: CultureScrollHotspot) {
  const segment = segments.find((seg) => seg.id === hotspot.segmentId)
  if (!segment) return hotspot.label ?? hotspot.segmentId
  return localizedField(segment, 'era')
}

function openHotspot(segmentId: string) {
  openHotspotId.value = segmentId
}

function closeHotspot() {
  openHotspotId.value = null
}

function onHotspotPointerDown(event: PointerEvent) {
  event.stopPropagation()
}

function onDialogKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && openHotspotId.value) {
    event.preventDefault()
    closeHotspot()
  }
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
  window.addEventListener('keydown', onDialogKeydown)

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
  window.removeEventListener('keydown', onDialogKeydown)
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
          <div class="culture-scroll__scenery-track" data-scroll-art="mid-scenery" aria-hidden="true">
            <img
              v-for="n in LINEART_TILES"
              :key="`scenery-${n}`"
              class="culture-scroll__scenery"
              :src="SCENERY_SRC"
              alt=""
              draggable="false"
            />
          </div>
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
          <button
            v-for="hotspot in hotspots"
            :key="hotspot.segmentId"
            type="button"
            class="culture-scroll__hotspot"
            :data-hotspot="hotspot.segmentId"
            :style="{ left: `${hotspot.xPercent}%` }"
            :aria-label="hotspotAriaLabel(hotspot)"
            aria-haspopup="dialog"
            :aria-expanded="openHotspotId === hotspot.segmentId"
            @pointerdown="onHotspotPointerDown"
            @click.stop="openHotspot(hotspot.segmentId)"
          />
        </div>
      </div>

      <p class="culture-scroll__colophon">
        {{ activeSegment?.era }} · {{ activeSegment?.eraEn }}
      </p>

      <div
        v-if="activeHotspot"
        class="culture-scroll__scrim"
        @click="closeHotspot"
        @pointerdown.stop
      >
        <div
          class="culture-scroll__dialog"
          role="dialog"
          aria-modal="true"
          aria-labelledby="culture-scroll-dialog-title"
          @click.stop
          @pointerdown.stop
        >
          <button
            type="button"
            class="culture-scroll__dialog-close"
            :aria-label="langStore.lang === 'zh' ? '关闭' : 'Close'"
            @click="closeHotspot"
          >
            ×
          </button>
          <h3 id="culture-scroll-dialog-title" class="culture-scroll__dialog-era">
            {{ localizedField(activeHotspot, 'era') }}
          </h3>
          <p class="culture-scroll__dialog-period">
            {{ localizedField(activeHotspot, 'period') }}
          </p>
          <p class="culture-scroll__dialog-desc">
            {{ localizedField(activeHotspot, 'desc') }}
          </p>
        </div>
      </div>
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

.culture-scroll__scenery-track,
.culture-scroll__lineart-track {
  display: flex;
  align-items: flex-end;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.culture-scroll__scenery-track {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.culture-scroll__lineart-track {
  position: absolute;
  inset: 0;
  z-index: 1;
}

.culture-scroll__scenery,
.culture-scroll__lineart {
  flex: 1 0 0;
  width: 0;
  height: 92%;
  object-fit: cover;
  object-position: center bottom;
  user-select: none;
  -webkit-user-drag: none;
}

.culture-scroll__scenery {
  opacity: 0.92;
  filter: saturate(1.05) contrast(1.02);
}

.culture-scroll__lineart {
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

.culture-scroll__hotspot {
  position: absolute;
  top: 48%;
  z-index: 3;
  width: 16px;
  height: 16px;
  padding: 0;
  border: 1.5px solid color-mix(in srgb, var(--color-gold) 80%, #fff);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-gold) 88%, #f4ead0);
  box-shadow:
    0 0 0 5px color-mix(in srgb, var(--color-gold) 22%, transparent),
    0 1px 4px color-mix(in srgb, var(--color-text-primary) 18%, transparent);
  transform: translate(-50%, -50%);
  pointer-events: auto;
  cursor: pointer;
}

.culture-scroll__hotspot:focus-visible {
  outline: 2px solid var(--color-gold-dark);
  outline-offset: 4px;
}

.culture-scroll:not(.culture-scroll--static) .culture-scroll__hotspot {
  animation: hotspot-pulse 2.4s ease-in-out infinite;
}

@keyframes hotspot-pulse {
  0%,
  100% {
    box-shadow:
      0 0 0 5px color-mix(in srgb, var(--color-gold) 22%, transparent),
      0 1px 4px color-mix(in srgb, var(--color-text-primary) 18%, transparent);
  }
  50% {
    box-shadow:
      0 0 0 9px color-mix(in srgb, var(--color-gold) 10%, transparent),
      0 1px 4px color-mix(in srgb, var(--color-text-primary) 18%, transparent);
  }
}

.culture-scroll__scrim {
  position: absolute;
  inset: 0;
  z-index: 6;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
  background: color-mix(in srgb, var(--color-text-primary) 28%, transparent);
}

.culture-scroll__dialog {
  position: relative;
  width: min(420px, 100%);
  padding: var(--space-6) var(--space-6) var(--space-5);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  box-shadow: var(--shadow-gold);
}

.culture-scroll__dialog-close {
  position: absolute;
  top: var(--space-3);
  right: var(--space-3);
  width: 32px;
  height: 32px;
  border: 0;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 1.35rem;
  line-height: 1;
  cursor: pointer;
}

.culture-scroll__dialog-era {
  margin: 0 0 var(--space-2);
  padding-right: var(--space-8);
  font-family: var(--font-display);
  font-size: var(--text-lg);
  font-weight: 700;
  letter-spacing: var(--tracking-wide);
  color: var(--color-text-primary);
}

.culture-scroll__dialog-period {
  margin: 0 0 var(--space-3);
  font-family: var(--font-en-body);
  font-size: var(--text-xs);
  letter-spacing: var(--tracking-wider);
  color: var(--color-gold);
}

.culture-scroll__dialog-desc {
  margin: 0;
  font-size: var(--text-sm);
  line-height: var(--leading-relaxed);
  color: var(--color-text-secondary);
}

.culture-scroll--static {
  height: auto;
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
