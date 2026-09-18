<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import {
  createRippleState,
  disturb,
  stepRipple,
  DEFAULT_RADIUS,
  DEFAULT_STRENGTH,
  type RippleState,
} from '@/utils/waterRipple'

const MAX_EDGE = 896
const DPR_CAP = 1.25
const RESIZE_DEBOUNCE_MS = 150
const AUTO_MIN_MS = 2500
const AUTO_MAX_MS = 4000
const AUTO_INSET = 0.15
const AUTO_SPAN = 0.7
const AUTO_STRENGTH = DEFAULT_STRENGTH * 0.75
/** 连续 N 帧无涟漪能量后停止模拟循环,避免空转占用主线程 */
const IDLE_STOP_FRAMES = 2

const props = withDefaults(defineProps<{ src?: string }>(), {
  src: '/images/home/hero-chengdu.jpg',
})

const rootRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const active = ref(false)
const follow = ref(false)

let disposed = false
let image: HTMLImageElement | null = null
let state: RippleState | null = null
let texture: ImageData | null = null
let output: ImageData | null = null
let ctx: CanvasRenderingContext2D | null = null
let rafId = 0
let autoTimer = 0
let resizeTimer = 0
let intersecting = true
let idleFrames = 0
let pendingPointer: { x: number; y: number } | null = null
let io: IntersectionObserver | null = null
let ro: ResizeObserver | null = null

function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

function hasFinePointer(): boolean {
  return window.matchMedia('(pointer: fine)').matches
}

function canAnimate(): boolean {
  return !disposed && active.value && intersecting && !document.hidden
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.decoding = 'async'
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error('hero image failed'))
    img.src = src
  })
}

function computeBufferSize(el: HTMLElement): { width: number; height: number } {
  const rect = el.getBoundingClientRect()
  const cssW = Math.max(1, Math.round(rect.width) || el.clientWidth || 1)
  const cssH = Math.max(1, Math.round(rect.height) || el.clientHeight || 1)
  const dpr = Math.min(window.devicePixelRatio || 1, DPR_CAP)
  let width = Math.round(cssW * dpr)
  let height = Math.round(cssH * dpr)
  const longest = Math.max(width, height)
  if (longest > MAX_EDGE) {
    const scale = MAX_EDGE / longest
    width = Math.max(1, Math.round(width * scale))
    height = Math.max(1, Math.round(height * scale))
  }
  return { width, height }
}

function drawCover(
  context: CanvasRenderingContext2D,
  img: HTMLImageElement,
  destW: number,
  destH: number,
): void {
  const iw = img.naturalWidth || img.width
  const ih = img.naturalHeight || img.height
  if (!iw || !ih) return
  const scale = Math.max(destW / iw, destH / ih)
  const sw = destW / scale
  const sh = destH / scale
  const sx = (iw - sw) / 2
  const sy = (ih - sh) / 2
  context.drawImage(img, sx, sy, sw, sh, 0, 0, destW, destH)
}

function rebuildBuffers(): void {
  const root = rootRef.value
  const canvas = canvasRef.value
  if (!root || !canvas || !image) return
  const { width, height } = computeBufferSize(root)
  if (state && canvas.width === width && canvas.height === height && texture && output) return

  canvas.width = width
  canvas.height = height
  const nextCtx = canvas.getContext('2d', { willReadFrequently: true })
  if (!nextCtx) return

  nextCtx.clearRect(0, 0, width, height)
  drawCover(nextCtx, image, width, height)
  texture = nextCtx.getImageData(0, 0, width, height)
  output = nextCtx.createImageData(width, height)
  output.data.set(texture.data)
  state = createRippleState(width, height)
  ctx = nextCtx
  ctx.putImageData(texture, 0, 0)
}

function stopLoop(): void {
  if (rafId) {
    cancelAnimationFrame(rafId)
    rafId = 0
  }
}

function tick(): void {
  rafId = 0
  if (!canAnimate() || !state || !texture || !output || !ctx) return
  if (pendingPointer) {
    disturb(state, pendingPointer.x, pendingPointer.y, DEFAULT_RADIUS, DEFAULT_STRENGTH)
    pendingPointer = null
  }
  const energetic = stepRipple(state, texture.data, output.data)
  ctx.putImageData(output, 0, 0)
  if (energetic) {
    idleFrames = 0
  } else {
    idleFrames++
    if (idleFrames >= IDLE_STOP_FRAMES) return // 涟漪已耗尽,停帧等待下一次扰动
  }
  rafId = requestAnimationFrame(tick)
}

function startLoop(): void {
  if (!canAnimate() || rafId) return
  rafId = requestAnimationFrame(tick)
}

function randomAutoDelay(): number {
  return AUTO_MIN_MS + Math.random() * (AUTO_MAX_MS - AUTO_MIN_MS)
}

function dropAutoRipple(): void {
  if (!canAnimate() || !state) return
  const x = state.width * (AUTO_INSET + Math.random() * AUTO_SPAN)
  const y = state.height * (AUTO_INSET + Math.random() * AUTO_SPAN)
  disturb(state, x, y, DEFAULT_RADIUS, AUTO_STRENGTH)
  idleFrames = 0
  startLoop()
}

function stopAuto(): void {
  if (autoTimer) {
    clearTimeout(autoTimer)
    autoTimer = 0
  }
}

function scheduleAuto(): void {
  stopAuto()
  if (!canAnimate()) return
  autoTimer = window.setTimeout(() => {
    autoTimer = 0
    dropAutoRipple()
    scheduleAuto()
  }, randomAutoDelay())
}

function syncPlayback(): void {
  if (canAnimate()) {
    startLoop()
    scheduleAuto()
  } else {
    stopLoop()
    stopAuto()
  }
}

function onPointerMove(event: PointerEvent): void {
  if (event.pointerType === 'touch' || !state) return
  queueRippleAtClient(event.clientX, event.clientY)
}

/** 外部（如 CTA 悬停）按视口坐标注入涟漪 */
function rippleAtClient(clientX: number, clientY: number): void {
  if (!active.value || !state) return
  queueRippleAtClient(clientX, clientY)
}

function queueRippleAtClient(clientX: number, clientY: number): void {
  if (!state) return
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  if (rect.width === 0 || rect.height === 0) return
  pendingPointer = {
    x: ((clientX - rect.left) / rect.width) * state.width,
    y: ((clientY - rect.top) / rect.height) * state.height,
  }
  idleFrames = 0
  startLoop() // 停帧状态下收到新扰动,恢复模拟循环
}

function bindPointer(): void {
  const root = rootRef.value
  if (!root || !follow.value) return
  root.addEventListener('pointermove', onPointerMove, { passive: true })
}

function unbindPointer(): void {
  rootRef.value?.removeEventListener('pointermove', onPointerMove)
}

function onVisibilityChange(): void {
  syncPlayback()
}

function bindObservers(): void {
  const root = rootRef.value
  if (!root) return

  if (typeof IntersectionObserver !== 'undefined') {
    io = new IntersectionObserver(
      (entries) => {
        intersecting = entries.some((entry) => entry.isIntersecting)
        syncPlayback()
      },
      { rootMargin: '0px' },
    )
    io.observe(root)
  }

  if (typeof ResizeObserver !== 'undefined') {
    ro = new ResizeObserver(() => {
      if (resizeTimer) clearTimeout(resizeTimer)
      resizeTimer = window.setTimeout(() => {
        resizeTimer = 0
        if (disposed || !active.value) return
        rebuildBuffers()
      }, RESIZE_DEBOUNCE_MS)
    })
    ro.observe(root)
  }

  document.addEventListener('visibilitychange', onVisibilityChange)
}

function unbindObservers(): void {
  io?.disconnect()
  io = null
  ro?.disconnect()
  ro = null
  document.removeEventListener('visibilitychange', onVisibilityChange)
  if (resizeTimer) {
    clearTimeout(resizeTimer)
    resizeTimer = 0
  }
}

function teardown(): void {
  stopLoop()
  stopAuto()
  unbindPointer()
  unbindObservers()
  pendingPointer = null
  idleFrames = 0
  state = null
  texture = null
  output = null
  ctx = null
  image = null
}

async function start(): Promise<void> {
  try {
    const loaded = await loadImage(props.src)
    if (disposed) return
    image = loaded
    rebuildBuffers()
    if (!state || !ctx) return
    active.value = true
    follow.value = hasFinePointer()
    bindPointer()
    bindObservers()
    syncPlayback()
  } catch {
    if (!disposed) active.value = false
  }
}

onMounted(() => {
  if (prefersReducedMotion()) return
  void start()
})

onBeforeUnmount(() => {
  disposed = true
  active.value = false
  follow.value = false
  teardown()
})

defineExpose({ rippleAtClient })
</script>

<template>
  <div
    ref="rootRef"
    class="hero-ripple"
    :data-active="active ? 'true' : undefined"
    :data-follow="follow ? 'true' : undefined"
    aria-hidden="true"
  >
    <canvas ref="canvasRef" class="hero-ripple__canvas" aria-hidden="true" />
  </div>
</template>

<style scoped>
.hero-ripple {
  position: absolute;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  transform: scale(1.02);
}
.hero-ripple[data-follow='true'] {
  pointer-events: auto;
}
.hero-ripple__canvas {
  width: 100%;
  height: 100%;
  display: block;
  pointer-events: none;
}
.hero-ripple[data-follow='true'] .hero-ripple__canvas {
  pointer-events: auto;
}
.hero-ripple:not([data-active]) .hero-ripple__canvas {
  opacity: 0;
}
</style>
