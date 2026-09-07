<script setup lang="ts">
/**
 * HomeBanner.vue — 统一欢迎横幅
 * 适用：首页、探索、美食、路线等 `/home` 子页面
 * 巨型水印文字由粒子沿轮廓勾成：鼠标靠近时粒子被「吸附」聚拢，
 * 快速划过则如利刃「一刀切」开粒子，切口泛起金光后缓慢愈合复位。
 */
import { onMounted, onBeforeUnmount, ref, watch, nextTick } from 'vue'

const props = defineProps<{
  eyebrow: string
  title: string
  subtitle?: string
  enTitle?: string
  watermark: string
}>()

const bannerRef = ref<HTMLElement | null>(null)
const watermarkRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)

interface Particle {
  homeX: number
  homeY: number
  x: number
  y: number
  vx: number
  vy: number
  size: number
  alpha: number
  color: string
  cutScale: number
  heat: number
}

let particles: Particle[] = []
let ctx: CanvasRenderingContext2D | null = null
let rafId = 0
let cssW = 0
let cssH = 0
let dpr = 1
let goldColor = '#5DA4B1'

/* 随水印字号自适应的互动参数 */
let radius = 140
let inner = 28
let cutWidth = 48

const ready = ref(false)
const mouse = { x: -9999, y: -9999 }
const segFrom = { x: -9999, y: -9999 }

/* 物理参数 */
const SPRING = 0.02 // 回弹刚度偏弱 → 粒子松散、切口愈合缓慢
const FRICTION = 0.86
const ATTRACT = 0.95 // 鼠标靠近时的吸附力
const INNER_PUSH = 1.15 // 距光标过近时轻微外推，避免糊成一团
const CUT_SPEED = 6 // 单帧位移超过该值判定为「挥切」而非悬停吸附
const CUT_FORCE = 5.4 // 切割冲量
const CUT_DRAG = 0.35 // 切割时沿挥动方向的拖拽分量
const MAX_PARTICLES = 4200

function collect(
  data: Uint8ClampedArray,
  w: number,
  h: number,
  step: number,
): Particle[] {
  const next: Particle[] = []
  for (let y = 0; y < h; y += step) {
    for (let x = 0; x < w; x += step) {
      const alpha = data[(y * w + x) * 4 + 3]
      if (alpha > 60) {
        next.push({
          homeX: x,
          homeY: y,
          // 初始散落，入场时汇聚勾勒出文字
          x: Math.random() * cssW,
          y: Math.random() * cssH,
          vx: 0,
          vy: 0,
          size: step * 0.34 + Math.random() * step * 0.3,
          alpha: 0.3 + Math.random() * 0.5,
          color: Math.random() < 0.12 ? goldColor : 'rgb(62, 125, 138)',
          cutScale: 0.7 + Math.random() * 0.6,
          heat: 0,
        })
      }
    }
  }
  return next
}

function buildParticles() {
  const el = watermarkRef.value
  const canvas = canvasRef.value
  if (!el || !canvas) return

  const text = props.watermark
  const cs = window.getComputedStyle(el)
  const fontSize = parseFloat(cs.fontSize) || 320
  const fontWeight = cs.fontWeight || '900'
  const fontFamily = cs.fontFamily
  const letterSpacing = parseFloat(cs.letterSpacing) || 0

  cssW = el.clientWidth
  cssH = el.clientHeight
  if (!text || cssW <= 0 || cssH <= 0) {
    particles = []
    return
  }

  dpr = Math.min(window.devicePixelRatio || 1, 2)
  canvas.width = Math.max(1, Math.round(cssW * dpr))
  canvas.height = Math.max(1, Math.round(cssH * dpr))
  canvas.style.width = `${cssW}px`
  canvas.style.height = `${cssH}px`

  /* 离屏描边采样：勾出巨型字的轮廓线，粒子沿轮廓生成 */
  const off = document.createElement('canvas')
  off.width = Math.max(1, Math.round(cssW))
  off.height = Math.max(1, Math.round(cssH))
  const offCtx = off.getContext('2d', { willReadFrequently: true })
  if (!offCtx) return

  const offCtxAny = offCtx as CanvasRenderingContext2D & { letterSpacing?: string }
  offCtxAny.font = `${fontWeight} ${fontSize}px ${fontFamily}`
  offCtxAny.letterSpacing = `${letterSpacing}px`
  offCtx.strokeStyle = '#000'
  offCtx.lineWidth = Math.max(2.5, Math.min(5, fontSize / 90))
  offCtx.textAlign = 'center'
  offCtx.textBaseline = 'middle'
  offCtx.strokeText(text, off.width / 2, off.height / 2)

  const data = offCtx.getImageData(0, 0, off.width, off.height).data

  const rootStyle = getComputedStyle(document.documentElement)
  goldColor = rootStyle.getPropertyValue('--color-gold').trim() || '#5DA4B1'

  radius = Math.max(90, Math.min(180, fontSize * 0.36))
  inner = radius * 0.2
  cutWidth = Math.max(26, Math.min(60, fontSize * 0.12))

  let step = Math.max(3, Math.round(fontSize / 110))
  let next = collect(data, off.width, off.height, step)
  if (next.length > MAX_PARTICLES) {
    step = Math.max(2, Math.round(step * Math.sqrt(next.length / MAX_PARTICLES)))
    next = collect(data, off.width, off.height, step)
  }
  particles = next
  ready.value = true
}

function onPointerMove(e: PointerEvent) {
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  const nx = e.clientX - rect.left
  const ny = e.clientY - rect.top
  // 首次进入时同步线段起点，避免从远处拉出一条假切口
  if (segFrom.x < -999) {
    segFrom.x = nx
    segFrom.y = ny
  }
  mouse.x = nx
  mouse.y = ny
}

function onPointerLeave() {
  mouse.x = -9999
  mouse.y = -9999
  segFrom.x = -9999
  segFrom.y = -9999
}

function loop() {
  if (ctx) {
    /* 本帧光标扫过的线段与速度 */
    const fromX = segFrom.x
    const fromY = segFrom.y
    const dx = mouse.x - fromX
    const dy = mouse.y - fromY
    const segLen = Math.sqrt(dx * dx + dy * dy)
    const slicing = segLen > CUT_SPEED
    const speedFactor = Math.min(segLen / 22, 1.8)
    segFrom.x = mouse.x
    segFrom.y = mouse.y

    const ux = slicing ? dx / segLen : 0
    const uy = slicing ? dy / segLen : 0

    for (const p of particles) {
      // 松散回弹：缓慢归位
      p.vx += (p.homeX - p.x) * SPRING
      p.vy += (p.homeY - p.y) * SPRING

      if (slicing) {
        /* 一刀切：轨迹窄带内的粒子沿法线被劈向两侧，切口干净利落 */
        const rx = p.x - fromX
        const ry = p.y - fromY
        const t = Math.min(Math.max(rx * ux + ry * uy, 0), segLen)
        const ox = rx - ux * t
        const oy = ry - uy * t
        const od = Math.sqrt(ox * ox + oy * oy)
        if (od < cutWidth) {
          const fall = 1 - od / cutWidth
          const f = CUT_FORCE * fall * speedFactor * p.cutScale
          const side = ux * ry - uy * rx >= 0 ? 1 : -1
          p.vx += -uy * side * f + ux * f * CUT_DRAG
          p.vy += ux * side * f + uy * f * CUT_DRAG
          p.heat = Math.min(1, p.heat + fall * 0.85)
        }
      } else {
        /* 吸附：光标靠近时粒子被柔和吸引，悬成松散一团 */
        const mdx = mouse.x - p.x
        const mdy = mouse.y - p.y
        const d = Math.sqrt(mdx * mdx + mdy * mdy)
        if (d < radius && d > 0.001) {
          const inv = 1 / d
          const f = ATTRACT * (1 - d / radius)
          p.vx += mdx * inv * f
          p.vy += mdy * inv * f
          if (d < inner) {
            const g = INNER_PUSH * (1 - d / inner)
            p.vx -= mdx * inv * g
            p.vy -= mdy * inv * g
          }
        }
      }

      p.vx *= FRICTION
      p.vy *= FRICTION
      p.x += p.vx
      p.y += p.vy
      p.heat *= 0.93
    }

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, cssW, cssH)
    for (const p of particles) {
      if (p.heat > 0.05) {
        // 切口处短暂泛起金色高光，如刀锋余温
        ctx.globalAlpha = Math.min(1, p.alpha + p.heat * 0.5)
        ctx.fillStyle = goldColor
        ctx.beginPath()
        ctx.arc(p.x, p.y, p.size * (1 + p.heat * 0.4), 0, Math.PI * 2)
        ctx.fill()
        ctx.globalAlpha = p.heat * 0.16
        ctx.beginPath()
        ctx.arc(p.x, p.y, p.size * 2.6, 0, Math.PI * 2)
        ctx.fill()
      } else {
        ctx.globalAlpha = p.alpha
        ctx.fillStyle = p.color
        ctx.beginPath()
        ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2)
        ctx.fill()
      }
    }
    ctx.globalAlpha = 1
  }
  rafId = requestAnimationFrame(loop)
}

onMounted(() => {
  ctx = canvasRef.value?.getContext('2d') ?? null
  buildParticles()
  rafId = requestAnimationFrame(loop)

  /* 显示字体晚于首次采样时，待字体就绪后重建粒子 */
  if (document.fonts && document.fonts.status !== 'loaded') {
    void document.fonts.ready.then(() => buildParticles())
  }

  window.addEventListener('resize', buildParticles)
  bannerRef.value?.addEventListener('pointermove', onPointerMove)
  bannerRef.value?.addEventListener('pointerleave', onPointerLeave)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(rafId)
  window.removeEventListener('resize', buildParticles)
  bannerRef.value?.removeEventListener('pointermove', onPointerMove)
  bannerRef.value?.removeEventListener('pointerleave', onPointerLeave)
})

/* 语言切换会传入新的 watermark，重新采样生成粒子 */
watch(
  () => [props.watermark, props.title],
  () => {
    void nextTick(buildParticles)
  },
)
</script>

<template>
  <section ref="bannerRef" class="home-banner shu-pattern">
    <div class="home-banner__bg" aria-hidden="true">
      <div class="home-banner__gradient" />
    </div>
    <span ref="watermarkRef" class="home-banner__watermark" aria-hidden="true"><span
      class="home-banner__watermark-text"
      :class="{ 'home-banner__watermark-text--ghost': ready }"
    >{{ watermark }}</span><canvas ref="canvasRef" class="home-banner__particles" aria-hidden="true"></canvas></span>
    <div class="home-banner__content">
      <p class="home-banner__eyebrow">
        <span>◈</span>
        {{ eyebrow }}
        <span>◈</span>
      </p>
      <h1 class="home-banner__title">
        {{ title }}
      </h1>
      <p v-if="enTitle" class="home-banner__en-title">{{ enTitle }}</p>
      <p v-if="subtitle" class="home-banner__subtitle">{{ subtitle }}</p>
      <div v-if="$slots.default" class="home-banner__extra">
        <slot />
      </div>
    </div>
  </section>
</template>

<style scoped>
.home-banner {
  position: relative;
  overflow: hidden;
  text-align: center;
  padding: var(--space-16) 0 var(--space-10);
  min-height: 480px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.home-banner__bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.home-banner__gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 72% 58% at 50% 38%, var(--color-gold-glow) 0%, transparent 68%),
    radial-gradient(ellipse 44% 36% at 18% 62%, var(--color-cinnabar-dim) 0%, transparent 60%);
}

.home-banner__watermark {
  position: absolute;
  font-family: var(--font-display);
  font-size: clamp(180px, 32vw, 420px);
  font-weight: 900;
  line-height: 1;
  user-select: none;
  pointer-events: none;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 0;
}

/* 内层文字层：粒子就绪后淡出，canvas 粒子层保持常驻 */
.home-banner__watermark-text {
  color: transparent;
  -webkit-text-stroke: 1.4px var(--banner-watermark-stroke, rgba(62, 125, 138, 0.34));
  text-shadow: 0 0 18px var(--banner-watermark-glow, rgba(62, 125, 138, 0.12));
  opacity: 0.95;
  transition: opacity 0.9s ease 0.15s;
}

.home-banner__watermark-text--ghost {
  opacity: 0;
}

.home-banner__content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-6) var(--space-4);
}

.home-banner__eyebrow {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--text-xs);
  color: var(--color-gold-dark);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.home-banner__title {
  position: relative;
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3.5rem);
  font-weight: 900;
  color: var(--color-text-primary);
  letter-spacing: var(--tracking-wide);
  line-height: 1.08;
}

.home-banner__particles {
  position: absolute;
  left: 0;
  top: 0;
  display: block;
  pointer-events: none;
}

.home-banner__en-title {
  font-family: var(--font-en-display);
  font-style: italic;
  font-size: var(--text-lg);
  color: var(--color-gold);
  letter-spacing: var(--tracking-wider);
  font-weight: 400;
  margin-top: calc(-1 * var(--space-2));
}

.home-banner__subtitle {
  font-family: var(--font-body);
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  font-weight: 300;
  letter-spacing: var(--tracking-wide);
  max-width: 42rem;
}

/* 额外内容（如探索页搜索栏、路线页统计） */
.home-banner__extra {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
  width: 100%;
  margin-top: var(--space-2);
}

@media (max-width: 640px) {
  .home-banner {
    min-height: 300px;
    padding: var(--space-10) 0 var(--space-8);
  }

  .home-banner__content {
    padding: var(--space-4);
  }

  .home-banner__subtitle {
    font-size: var(--text-sm);
  }
}
</style>
