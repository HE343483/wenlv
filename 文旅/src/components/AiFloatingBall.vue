<template>
  <div class="fab-root">
    <!-- 粒子 Canvas -->
    <canvas ref="canvasRef" class="p-canvas" v-show="showParticles" />

    <!-- 悬浮容器 -->
    <div
      ref="wrapperRef"
      class="fab-wrap"
      :class="state"
      :style="wrapperStyle"
    >
      <!-- 球体 -->
      <div
        ref="ballRef"
        class="fab-ball"
        @pointerdown="onDown"
        role="button"
        tabindex="0"
        :aria-label="isExpanded ? '关闭AI助手' : '打开AI助手'"
        :aria-expanded="isExpanded"
      >
        <svg viewBox="0 0 24 24" class="fab-icon" width="22" height="22" fill="none" stroke="#fff" stroke-width="2">
          <path d="M12 2l2 7h7l-5.5 4 2 7L12 16l-5.5 4 2-7L3 9h7z"/>
        </svg>
        <div class="fab-glow"></div>
      </div>

      <!-- 面板 -->
      <div v-if="isExpanded || panning" class="fab-panel" ref="panelRef">
        <div class="fp-header">
          <div class="fp-brand">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="#E85D3A" stroke-width="2">
              <path d="M12 2l2 7h7l-5.5 4 2 7L12 16l-5.5 4 2-7L3 9h7z"/>
            </svg>
            <span>AI 蓉城助手</span>
          </div>
          <button class="fp-close" @click="close" aria-label="关闭">✕</button>
        </div>

        <div class="fp-body" ref="bodyRef">
          <div class="fp-msg fp-reply">您好！我是蓉城文旅助手，请问有什么可以帮助您的？</div>
          <div class="fp-msg fp-user" v-for="m in messages" :key="m">{{ m }}</div>
        </div>

        <div class="fp-foot">
          <input v-model="inputText" class="fp-input" placeholder="输入问题..." @keyup.enter="send" />
          <button class="fp-send" @click="send">发送</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive, watch, onBeforeUnmount, nextTick } from 'vue'

// ===== 状态机 =====
const STATES = {
  IDLE: 'idle',
  DRAGGING: 'dragging',
  EXPANDING: 'expanding',
  EXPANDED: 'expanded',
  CLOSING: 'closing'
}
// 允许的状态转换
const ALLOWED = {
  [STATES.IDLE]:      [STATES.DRAGGING, STATES.EXPANDING],
  [STATES.DRAGGING]:  [STATES.IDLE],
  [STATES.EXPANDING]: [STATES.EXPANDED],
  [STATES.EXPANDED]:  [STATES.CLOSING],
  [STATES.CLOSING]:   [STATES.IDLE]
}

const state = ref(STATES.IDLE)
const isExpanded = computed(() => state.value === STATES.EXPANDED)
const panning = computed(() => state.value === STATES.EXPANDING || state.value === STATES.CLOSING)
const DRAG_THRESHOLD = 8

function setState(s) {
  if (ALLOWED[state.value]?.includes(s)) state.value = s
}

// ===== 位置 =====
const BALL_SIZE = 52
const posX = ref(window.innerWidth - BALL_SIZE - 24)
const posY = ref(window.innerHeight - BALL_SIZE - 24)
const dragOff = reactive({ x: 0, y: 0 })
let dragStart = { x: 0, y: 0, px: 0, py: 0 }
let hasMoved = false

const wrapperStyle = computed(() => ({
  left: posX.value + 'px',
  top: posY.value + 'px',
  transform: `translate(${dragOff.x}px, ${dragOff.y}px)`
}))

// ===== 指针事件 =====
function onDown(e) {
  if (state.value !== STATES.IDLE) return
  e.preventDefault()
  dragStart = { x: e.clientX, y: e.clientY, px: posX.value, py: posY.value }
  hasMoved = false
  dragOff.x = 0; dragOff.y = 0

  const onMove = (ev) => {
    const dx = ev.clientX - dragStart.x
    const dy = ev.clientY - dragStart.y
    if (!hasMoved && Math.hypot(dx, dy) > DRAG_THRESHOLD) {
      hasMoved = true
      setState(STATES.DRAGGING)
    }
    if (hasMoved) {
      dragOff.x = dx
      dragOff.y = dy
    }
  }

  const onUp = (ev) => {
    document.removeEventListener('pointermove', onMove)
    document.removeEventListener('pointerup', onUp)
    if (hasMoved) {
      // 拖拽结束 → 吸附边缘
      snapToEdge(dragStart.px + dragOff.x, dragStart.py + dragOff.y)
    } else {
      // 单击 → 展开
      setState(STATES.EXPANDING)
      spawnParticles(dragStart.x, dragStart.y, true)
      setTimeout(() => setState(STATES.EXPANDED), 400)
    }
  }

  document.addEventListener('pointermove', onMove, { passive: true })
  document.addEventListener('pointerup', onUp)
}

function snapToEdge(x, y) {
  const margin = 16
  const winW = window.innerWidth
  const winH = window.innerHeight
  // 水平：近哪边靠哪边
  const snapX = Math.min(Math.max(x, margin), winW - BALL_SIZE - margin)
  // 垂直：同上
  const snapY = Math.min(Math.max(y, margin), winH - BALL_SIZE - margin)

  posX.value = snapX
  posY.value = snapY
  dragOff.x = 0
  dragOff.y = 0
  setState(STATES.IDLE)
}

// ===== 关闭面板 =====
function close() {
  const rect = wrapperRef.value?.getBoundingClientRect()
  const cx = rect ? rect.left + rect.width / 2 : posX.value + BALL_SIZE / 2
  const cy = rect ? rect.top + rect.height / 2 : posY.value + BALL_SIZE / 2
  setState(STATES.CLOSING)
  spawnParticles(cx, cy, false)
  setTimeout(() => setState(STATES.IDLE), 300)
}

// ===== 聊天 =====
const inputText = ref('')
const messages = ref([])

function send() {
  const t = inputText.value.trim()
  if (!t) return
  messages.value.push(t)
  inputText.value = ''
  setTimeout(() => {
    messages.value.push('🤖 已收到您的问题，我将为您查询成都景区相关信息。')
  }, 600)
}

// ===== 粒子系统 =====
const canvasRef = ref(null)
const showParticles = ref(false)
let particleRaf = null

function spawnParticles(targetX, targetY, converging) {
  const canvas = canvasRef.value
  if (!canvas) return
  showParticles.value = true
  canvas.width = window.innerWidth
  canvas.height = window.innerHeight
  const ctx = canvas.getContext('2d')

  const COLORS = ['#E85D3A', '#D4A854', '#F5A623']
  const COUNT = 24
  const particles = []

  for (let i = 0; i < COUNT; i++) {
    let sx, sy, ex, ey
    if (converging) {
      // 从屏幕边缘随机位置出发 → 汇聚到目标点
      const side = Math.floor(Math.random() * 4)
      if (side === 0) { sx = Math.random() * canvas.width; sy = -20 - Math.random() * 40 }
      else if (side === 1) { sx = canvas.width + 20 + Math.random() * 40; sy = Math.random() * canvas.height }
      else if (side === 2) { sx = Math.random() * canvas.width; sy = canvas.height + 20 + Math.random() * 40 }
      else { sx = -20 - Math.random() * 40; sy = Math.random() * canvas.height }
      ex = targetX; ey = targetY
    } else {
      // 从目标点出发 → 向外扩散到屏幕边缘
      sx = targetX; sy = targetY
      const angle = Math.random() * Math.PI * 2
      const dist = 200 + Math.random() * 300
      ex = targetX + Math.cos(angle) * dist
      ey = targetY + Math.sin(angle) * dist
    }

    const cp1x = sx + (ex - sx) * 0.3 + (Math.random() - 0.5) * 80
    const cp1y = sy + (ey - sy) * 0.2 + (Math.random() - 0.5) * 60
    const cp2x = sx + (ex - sx) * 0.7 + (Math.random() - 0.5) * 60
    const cp2y = sy + (ey - sy) * 0.8 + (Math.random() - 0.5) * 40

    particles.push({
      sx, sy, cp1x, cp1y, cp2x, cp2y, ex, ey,
      radius: 2 + Math.random() * 3,
      color: COLORS[Math.floor(Math.random() * COLORS.length)],
      delay: Math.random() * 60,
      duration: 300 + Math.random() * 100,
      progress: 0
    })
  }

  let start = null
  function draw(ts) {
    if (!start) start = ts
    const elapsed = ts - start
    ctx.clearRect(0, 0, canvas.width, canvas.height)

    let allDone = true
    for (const p of particles) {
      const t = Math.min(1, Math.max(0, (elapsed - p.delay) / p.duration))
      if (t === 0) continue
      allDone = false
      const u = 1 - t
      const x = u*u*u*p.sx + 3*u*u*t*p.cp1x + 3*u*t*t*p.cp2x + t*t*t*p.ex
      const y = u*u*u*p.sy + 3*u*u*t*p.cp1y + 3*u*t*t*p.cp2y + t*t*t*p.ey
      const alpha = t < 0.15 ? t / 0.15 * 0.6 : t > 0.85 ? (1 - t) / 0.15 * 0.6 : 0.6

      ctx.beginPath()
      ctx.arc(x, y, p.radius, 0, Math.PI * 2)
      ctx.fillStyle = p.color
      ctx.globalAlpha = alpha
      ctx.fill()
    }

    if (!allDone) {
      particleRaf = requestAnimationFrame(draw)
    } else {
      showParticles.value = false
    }
  }

  if (particleRaf) cancelAnimationFrame(particleRaf)
  particleRaf = requestAnimationFrame(draw)
}

// ===== 全局 ESC 关闭 =====
function onKeydown(e) {
  if (e.key === 'Escape' && isExpanded.value) close()
}

// ===== 面板位置适配（自动避开边缘） =====
const wrapperRef = ref(null)
const ballRef = ref(null)
const panelRef = ref(null)
const bodyRef = ref(null)

// ===== 生命周期 =====
document.addEventListener('keydown', onKeydown)

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
  if (particleRaf) cancelAnimationFrame(particleRaf)
  showParticles.value = false
})
</script>

<style scoped>
/* ===== 根容器 ===== */
.fab-root {
  position: relative;
}

/* ===== 粒子 Canvas ===== */
.p-canvas {
  position: fixed;
  inset: 0;
  z-index: 9998;
  pointer-events: none;
}

/* ===== 悬浮容器 ===== */
.fab-wrap {
  position: fixed;
  z-index: 9999;
  /* 位置通过 left/top + transform 控制 */
}

/* ===== 球体 ===== */
.fab-ball {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: linear-gradient(135deg, #E85D3A, #D4A854);
  box-shadow: 0 4px 16px rgba(232, 93, 58, .25);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: grab;
  position: relative;
  user-select: none;
  -webkit-user-select: none;
  touch-action: none;
  transition: transform .25s cubic-bezier(0.34, 1.56, 0.64, 1),
              opacity .2s ease,
              box-shadow .3s ease;
  will-change: transform;
}

.fab-ball:hover {
  box-shadow: 0 6px 24px rgba(232, 93, 58, .35);
}

.fab-icon {
  position: relative;
  z-index: 2;
  filter: drop-shadow(0 1px 2px rgba(0,0,0,.1));
}

/* 呼吸光晕 */
.fab-glow {
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, rgba(255,255,255,.2), transparent 60%);
  pointer-events: none;
  animation: glow-pulse 2s ease-in-out infinite;
}

@keyframes glow-pulse {
  0%, 100% { opacity: .5; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.08); }
}

/* ===== 拖拽状态 ===== */
.dragging .fab-ball {
  cursor: grabbing;
  transform: scale(1.05);
  box-shadow: 0 8px 32px rgba(232, 93, 58, .3);
  transition: none;
}
.dragging .fab-glow { animation: none; opacity: .8; }

/* ===== 展开中 — 球体缩小隐去 ===== */
.expanding .fab-ball {
  transform: scale(0.3);
  opacity: 0;
  transition: transform .25s cubic-bezier(0.34, 1.56, 0.64, 1),
              opacity .2s ease;
}

/* ===== 收起中 ===== */
.closing .fab-ball {
  transform: scale(0.3);
  opacity: 0;
}

.closing.idle .fab-ball {
  transition: transform .3s cubic-bezier(0.34, 1.56, 0.64, 1),
              opacity .25s ease;
  transform: scale(1);
  opacity: 1;
}

/* ===== 面板 ===== */
.fab-panel {
  position: absolute;
  bottom: calc(100% + 12px);
  right: 0;
  width: 360px;
  max-height: 460px;
  background: rgba(255, 255, 255, .96);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(232, 93, 58, .08);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, .08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transform-origin: bottom right;
}

/* 展开动画 */
.expanding .fab-panel {
  animation: panel-in .35s cubic-bezier(0.16, 1, 0.3, 1) .08s both;
}

@keyframes panel-in {
  0% { opacity: 0; transform: scale(0.6) translateY(16px); }
  100% { opacity: 1; transform: scale(1) translateY(0); }
}

/* 收起动画 */
.closing .fab-panel {
  animation: panel-out .2s ease-in both;
}

@keyframes panel-out {
  0% { opacity: 1; transform: scale(1) translateY(0); }
  100% { opacity: 0; transform: scale(0.6) translateY(16px); }
}

/* ===== 面板内容 ===== */
.fp-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(232, 93, 58, .06);
  flex-shrink: 0;
}

.fp-brand {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: .8rem;
  font-weight: 600;
  color: #E85D3A;
}

.fp-close {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #9CA3AF;
  font-size: .8rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .25s;
}

.fp-close:hover {
  color: #E85D3A;
  background: rgba(232, 93, 58, .06);
}

/* 消息体 — 内容渐入 */
.fp-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.fp-body > * {
  animation: msg-in .3s ease both;
}

@keyframes msg-in {
  0% { opacity: 0; transform: translateY(6px); }
  100% { opacity: 1; transform: translateY(0); }
}

.fp-msg {
  max-width: 85%;
  padding: 8px 12px;
  border-radius: 10px;
  font-size: .74rem;
  line-height: 1.5;
}

.fp-reply {
  background: rgba(232, 93, 58, .06);
  color: #4B5563;
  align-self: flex-start;
}

.fp-user {
  background: rgba(14, 165, 160, .08);
  color: #4B5563;
  align-self: flex-end;
}

/* 底部输入 */
.fp-foot {
  display: flex;
  gap: 6px;
  padding: 10px 12px;
  border-top: 1px solid rgba(0, 0, 0, .04);
  flex-shrink: 0;
}

.fp-input {
  flex: 1;
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid rgba(0, 0, 0, .06);
  background: rgba(255, 255, 255, .8);
  color: #2D2D3A;
  font-size: .74rem;
  font-family: inherit;
  outline: none;
}

.fp-input:focus { border-color: rgba(232, 93, 58, .25); }

.fp-send {
  padding: 8px 14px;
  border-radius: 8px;
  border: none;
  background: linear-gradient(135deg, #E85D3A, #D4A854);
  color: #fff;
  font-size: .7rem;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
  transition: opacity .25s;
}

.fp-send:hover { opacity: .9; }

/* ===== 响应式 ===== */
@media (max-width: 640px) {
  .fab-ball { width: 48px; height: 48px; }
  .fab-panel { width: 85vw; right: -10px; max-height: 70vh; }
}

/* ===== 动画降级 ===== */
@media (prefers-reduced-motion: reduce) {
  .fab-glow { animation: none; }
  .expanding .fab-panel { animation: none; }
  .closing .fab-panel { animation: none; }
  .fp-body > * { animation: none; }
}
</style>
